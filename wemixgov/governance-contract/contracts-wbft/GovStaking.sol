// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "./GovConst.sol";

contract GovStaking {
    using EnumerableSet for EnumerableSet.AddressSet;
    using Address for address payable;

    struct Staker {
        address operator;
        address rewardee;
        uint256 staking;
        uint256 delegated;
    }

    struct WithdrawalCredential {
        address requester;
        uint256 amount;
        uint256 requestTime;
        uint256 withdrawableTime;
        WithdrawalStatus status;
    }

    enum WithdrawalStatus {
        None,
        Requested,
        Withdrawn
    }
    event StakerRegistered(address indexed staker, address operator, address rewardee, uint256 staking);
    event Staked(address indexed staker, uint256 amount);
    event Unstaked(address indexed staker, uint256 amount);
    event StakerRemoved(address indexed staker);
    event Delegated(address indexed delegator, address indexed staker, uint256 amount);
    event Undelegated(address indexed delegator, address indexed staker, uint256 amount);
    event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding);
    event Withdrawn(uint256 indexed credentialID, address requester, uint256 amount);

    GovConst public constant GOV_CONST = GovConst(address(0x1000));
    
    uint256 public totalStaking; // 0x0

    // Staker
    EnumerableSet.AddressSet private __stakerSet; // 0x1, 0x2
    mapping(address => Staker) public stakerInfo; // 0x3
    mapping(address => address) public stakerByOperator; // 0x4
    mapping(address => address) public stakerByRewardee; // 0x5

    // Delegate
    mapping(address => mapping(address => uint256)) public delegateTo; // 0x5

    // Withdrawal Credential
    uint256 public credentialCount; // 0x6
    mapping(uint256 => WithdrawalCredential) public credentials; // 0x7


    modifier checkAmount(uint256 _amount) {
        require(msg.value == _amount, "amount and msg.value mismatch");
        _;
    }

    function isStaker(address _staker) public view returns (bool) {
        return __stakerSet.contains(_staker);
    }

    function isOperatorOrRewardee(address _addr) public view returns (bool) {
        return stakerByOperator[_addr] != address(0) || stakerByRewardee[_addr] != address(0);
    }

    function stakerLength() external view returns (uint256) {
        return __stakerSet.length();
    }

    function stakers() external view returns (address[] memory) {
        return __stakerSet.values();
    }

    function registerStaker(uint256 _amount, address _staker, address _rewardee) external payable checkAmount(_amount) {
        require(_amount >= GOV_CONST.MINIMUM_STAKING() && _amount <= GOV_CONST.MAXIMUM_STAKING(), "out of bounds");
        require(msg.sender != _staker && msg.sender != _rewardee, "operator cannot be staker or rewardee");
        require(_staker != address(0) && _rewardee != address(0), "zero address");
        require(_staker != _rewardee, "staker cannot be rewardee");
        require(!isOperatorOrRewardee(msg.sender), "operator is already registered");
        require(!isOperatorOrRewardee(_staker), "staker is already registered");
        require(!isOperatorOrRewardee(_rewardee), "rewardee is already registered");

        require(__stakerSet.add(_staker), "staker exists");
        stakerInfo[_staker] = Staker({ operator: msg.sender, rewardee: _rewardee, staking: _amount, delegated: 0 });

        stakerByOperator[msg.sender] = _staker;
        stakerByRewardee[_rewardee] = _staker;

        totalStaking += _amount;

        emit StakerRegistered(_staker, msg.sender, _rewardee, _amount);
    }

    function stake(uint256 _amount) external payable checkAmount(_amount) {
        address _staker = stakerByOperator[msg.sender];
        _addStaking(stakerByOperator[msg.sender], _amount, false);

        emit Staked(_staker, _amount);
    }

    function unstake(uint256 _amount) external {
        address _staker = stakerByOperator[msg.sender];
        require(_staker != address(0), "unregistered staker");
        require(_amount > 0, "amount is zero");

        Staker storage _stakerInfo = stakerInfo[_staker];
        uint256 _stakerStaking = _stakerInfo.staking - _stakerInfo.delegated;

        require(_stakerStaking >= _amount, "insufficient balance");
        if (_stakerStaking - _amount < GOV_CONST.MINIMUM_STAKING()) {
            require(_stakerStaking == _amount, "amount must equal balance to remove staker");

            __stakerSet.remove(_staker);
            delete stakerByOperator[msg.sender];
            delete stakerByRewardee[_stakerInfo.rewardee];
            delete stakerInfo[_staker];

            emit StakerRemoved(_staker);
        } else {
            _stakerInfo.staking -= _amount;
        }

        totalStaking -= _amount;
        _newCredential(_amount, GOV_CONST.UNBONDING_PERIOD_STAKER());

        emit Unstaked(_staker, _amount);
    }

    function delegate(address _staker, uint256 _amount) external payable checkAmount(_amount) {
        require(!isStaker(msg.sender), "staker cannot delegate");
        require(!isOperatorOrRewardee(msg.sender), "operator(rewardee) cannot delegate");

        _addStaking(_staker, _amount, true);
        delegateTo[msg.sender][_staker] += _amount;

        emit Delegated(msg.sender, _staker, _amount);
    }

    function undelegate(address _staker, uint256 _amount) external {
        require(delegateTo[msg.sender][_staker] >= _amount, "insufficient balance");

        if (isStaker(_staker)) {
            Staker storage _stakerInfo = stakerInfo[_staker];
            _stakerInfo.delegated -= _amount;
            _stakerInfo.staking -= _amount;

            _newCredential(_amount, GOV_CONST.UNBONDING_PERIOD_DELEGATOR());
        } else {
            payable(msg.sender).sendValue(_amount);
        }

        delegateTo[msg.sender][_staker] -= _amount;
        totalStaking -= _amount;

        emit Undelegated(msg.sender, _staker, _amount);
    }

    function withdraw(uint256 _cid) external {
        WithdrawalCredential storage _credential = credentials[_cid];
        require(_credential.status == WithdrawalStatus.Requested, "invalid credential");
        require(_credential.requester == msg.sender, "msg.sender is not requester");
        require(block.timestamp >= _credential.withdrawableTime, "not yet time to withdraw");

        payable(_credential.requester).sendValue(_credential.amount);
        _credential.status = WithdrawalStatus.Withdrawn;

        emit Withdrawn(_cid, msg.sender, _credential.amount);
    }

    function _addStaking(address _staker, uint256 _amount, bool _delegated) private {
        require(isStaker(_staker), "unregistered staker");

        Staker storage _stakerInfo = stakerInfo[_staker];
        require(_stakerInfo.staking + _amount <= GOV_CONST.MAXIMUM_STAKING(), "exceeded the maximum");

        totalStaking += _amount;
        _stakerInfo.staking += _amount;
        if (_delegated) {
            _stakerInfo.delegated += _amount;
        }
    }

    function _newCredential(uint256 _amount, uint256 _unbondingPeriod) private {
        credentials[++credentialCount] = WithdrawalCredential({
            requester: msg.sender,
            amount: _amount,
            requestTime: block.timestamp,
            withdrawableTime: block.timestamp + _unbondingPeriod,
            status: WithdrawalStatus.Requested
        });

        emit NewCredential(credentialCount, msg.sender, _amount, block.timestamp, _unbondingPeriod);
    }
}
