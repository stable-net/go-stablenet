// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

import "./GovConst.sol";
import {GovRewardeeImp} from "./GovRewardeeImp.sol";
import {GovRewardee} from "./GovRewardee.sol";

contract GovStaking {
    using EnumerableSet for EnumerableSet.AddressSet;

    struct Staker {
        // configuration
        address operator;
        address rewardee;
        address feeRecipient;
        uint256 feeRate;
        bytes blsPubKey;

        // mutable
        uint256 totalStaked; // updated when stake/unstake/delegate/undelegate
        uint256 accRewardPerStaking; // updated when stake/unstake/delegate/undelegate/claim
        uint256 accFeePerStaking; // updated when stake/unstake/delegate/undelegate/claim
        uint256 lastRewardBalance; // updated when stake/unstake/delegate/undelegate/claim
    }

    struct UserInfo {
        uint256 stakingAmount;
        uint256 pendingReward;
        uint256 pendingFee;
        uint256 rewardPerStaking;
        uint256 feePerStaking;
    }

    struct WithdrawalCredential {
        address requester;
        uint256 amount;
        uint256 requestTime;
        uint256 withdrawableTime;
        WithdrawalStatus status;
    }

    struct ChangeFeeRequest {
        uint256 newFeeRate;
        uint256 requestTime;
    }

    enum WithdrawalStatus {
        None,
        Requested,
        Withdrawn
    }

    event StakerRegistered(address indexed staker, address operator, address rewardee, address feeRecipient, uint256 feeRate, uint256 staking, bytes blsPK);
    event Staked(address indexed staker, uint256 amount);
    event Unstaked(address indexed staker, uint256 amount);
    event StakerRemoved(address indexed staker);
    event Delegated(address indexed delegator, address indexed staker, uint256 amount);
    event Undelegated(address indexed delegator, address indexed staker, uint256 amount);
    event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding);
    event Withdrawn(uint256 indexed credentialID, address requester, uint256 amount);
    event RewardInfoUpdated(address indexed staker, uint256 totalStaked, uint256 balance, uint256 accBalance, uint256 accRewardPerStaking, uint256 accFeePerStaking);
    event UserRewardUpdated(address indexed staker, address indexed user, uint256 stakingAmount, uint256 pendingReward, uint256 accRewardPerStaking, uint256 accFeePerStaking);
    event Claimed(address indexed staker, address indexed rewardee, uint256 amount, bool restake);
    event FeeRecipientChanged(address indexed staker, address oldRecipient, address newRecipient);
    event FeeRateChangeRequested(address indexed staker, uint256 oldFeeRate, uint256 newFeeRate);

    GovConst public constant GOV_CONST = GovConst(address(0x1000));

    // this includes danglingDelegated
    uint256 public totalStaking; // 0x0

    // Staker
    // Staker state definition
    //  0. unregistered: stakerInfo[staker].operator = 0, __stakerSet.contains(staker) = false
    //  1. active: stakerInfo[staker].operator != 0, __stakerSet.contains(staker) = true
    //  2. inactive: stakerInfo[staker].operator != 0, __stakerSet.contains(staker) = false
    EnumerableSet.AddressSet private __stakerSet; // 0x1, 0x2
    mapping(address => Staker) public stakerInfo; // 0x3
    mapping(address => address) public stakerByOperator; // 0x4
    mapping(address => address) public stakerByRewardee; // 0x5

    // Withdrawal Credential
    uint256 public credentialCount; // 0x6
    mapping(uint256 => WithdrawalCredential) public credentials; // 0x7

    // pending request
    mapping(address => ChangeFeeRequest) public pendingRequest; // 0x8

    // User Reward Info
    mapping(address => mapping(address => UserInfo)) public userRewardInfo; // 0x9

    // danglingDelegated is the delegated balance for the inactive stakers
    // contract's balance = totalStaked + danglingDelegated + unbonding
    uint256 public danglingDelegated; // 0xa
    bool public afterStabilization; // 0xb

    modifier checkAmount(uint256 _amount) {
        require(msg.value == _amount, "amount and msg.value mismatch");
        _;
    }

    modifier checkStaker(address _staker) {
        require(_staker != address(0), "unregistered staker");
        require(isStaker(_staker), "unregistered staker");
        _;
    }

    // Rewardee sends coin to this contract, so receive() is required
    receive() external payable {
        require(isStaker(stakerByRewardee[msg.sender]), "only an active rewardee can send coin");
    }

    function isStaker(address _staker) public view returns (bool) {
        return __stakerSet.contains(_staker);
    }

    function isOperator(address _addr) public view returns (bool) {
        return stakerByOperator[_addr] != address(0);
    }

    function stakerLength() external view returns (uint256) {
        return __stakerSet.length();
    }

    function stakers() external view returns (address[] memory) {
        return __stakerSet.values();
    }

    function registerStaker(uint256 _amount, address _staker, address _feeRecipient, uint256 _feeRate, bytes calldata _blsPK) external payable checkAmount(_amount) {
        require(_amount >= GOV_CONST.MINIMUM_STAKING() && _amount <= GOV_CONST.MAXIMUM_STAKING(), "out of bounds");
        require(msg.sender != _staker, "operator cannot be staker");
        require(_staker != address(0), "zero address");
        require(!isOperator(msg.sender), "operator is already registered");
        require(!isStaker(_staker), "staker is already registered");
        require(_feeRecipient != address(0), "fee recipient is zero address");
        require(_feeRate <= GOV_CONST.FEE_PRECISION(), "fee rate exceeds precision");
        require(_blsPK.length == GOV_CONST.BLS_PUBLIC_KEY_LENGTH(), "invalid bls public key");

        GovRewardee _rewardee = new GovRewardee();
        stakerInfo[_staker].operator = msg.sender;
        stakerInfo[_staker].rewardee = address(_rewardee);
        stakerInfo[_staker].feeRecipient = _feeRecipient;
        stakerInfo[_staker].feeRate = _feeRate;
        stakerInfo[_staker].blsPubKey = _blsPK;

        stakerByOperator[msg.sender] = _staker;
        stakerByRewardee[address(_rewardee)] = _staker;

        __stakerSet.add(_staker);

        _updateRewardInfo(_staker, msg.sender);

        _addStaking(_staker, msg.sender, _amount);

        if (__stakerSet.length() >= GOV_CONST.MIN_STAKERS()) {
            afterStabilization = true;
        }

        emit StakerRegistered(
            _staker,
            msg.sender,
            address(_rewardee),
            _feeRecipient,
            _feeRate,
            _amount,
            _blsPK
        );
    }

    function changeFeeRecipient(address _newRecipient) external checkStaker(stakerByOperator[msg.sender]) {
        require(_newRecipient != address(0), "zero address");
        address _staker = stakerByOperator[msg.sender];
        address oldRecipient = stakerInfo[_staker].feeRecipient;
        stakerInfo[_staker].feeRecipient = _newRecipient;

        emit FeeRecipientChanged(_staker, oldRecipient, _newRecipient);
    }

    function requestChangeFee(uint256 _feeRate) external checkStaker(stakerByOperator[msg.sender]) {
        require(_feeRate <= GOV_CONST.FEE_PRECISION(), "fee rate exceeds precision");
        address _staker = stakerByOperator[msg.sender];
        require(pendingRequest[_staker].requestTime == 0, "request already is on going");

        uint256 oldFeeRate = stakerInfo[_staker].feeRate;
        if (getDelegatedAmount(_staker) > 0) {
            pendingRequest[_staker] = ChangeFeeRequest({ newFeeRate: _feeRate, requestTime: block.timestamp });
        }
        else {
            // if no delegator exists, change fee immediately
            stakerInfo[_staker].feeRate = _feeRate;
        }

        emit FeeRateChangeRequested(_staker, oldFeeRate, _feeRate);
    }

    function executeChangeFee(address _staker) external {
        require(pendingRequest[_staker].requestTime > 0, "no request exists");
        require(block.timestamp - pendingRequest[_staker].requestTime >= GOV_CONST.CHANGE_FEE_DELAY(),
            "the request cannot be executed before delay time");

        // don't update user info passing zero address
        _updateRewardInfo(_staker, address(0));
    }

    function stake(uint256 _amount) external payable checkAmount(_amount) {
        address _staker = stakerByOperator[msg.sender];
        require(_staker != address(0), "unregistered staker");

        // update stake info
        _updateRewardInfo(_staker, msg.sender);

        if (!isStaker(_staker)) {
            // reactivation case: if the staker is not active, then reactivate it

            require(_amount >= GOV_CONST.MINIMUM_STAKING(), "amount is less than minimum staking");

            __stakerSet.add(_staker);
            danglingDelegated -= stakerInfo[_staker].totalStaked;
        }

        _addStaking(_staker, msg.sender, _amount);

        emit Staked(_staker, _amount);
    }

    function unstake(uint256 _amount) external checkStaker(stakerByOperator[msg.sender]) {
        require(_amount > 0, "amount is zero");

        address _staker = stakerByOperator[msg.sender];

        // update stake info
        _updateRewardInfo(_staker, msg.sender);

        _subStaking(_staker, msg.sender, _amount);

        UserInfo storage _userInfo = userRewardInfo[_staker][msg.sender];
        if (_userInfo.stakingAmount < GOV_CONST.MINIMUM_STAKING()) {
            require(_userInfo.stakingAmount == 0, "amount must equal balance to deactivate staker");

            __stakerSet.remove(_staker);

            danglingDelegated += stakerInfo[_staker].totalStaked;

            emit StakerRemoved(_staker);
        }

        _newCredential(_amount, GOV_CONST.UNBONDING_PERIOD_STAKER());

        emit Unstaked(_staker, _amount);
    }

    function delegate(address _staker, uint256 _amount) external payable checkAmount(_amount) checkStaker(_staker) {
        require(msg.sender != _staker, "staker cannot delegate to self");
        require(msg.sender != stakerInfo[_staker].operator, "operator cannot delegate to self");

        _updateRewardInfo(_staker, msg.sender);

        _addStaking(_staker, msg.sender, _amount);

        emit Delegated(msg.sender, _staker, _amount);
    }

    function undelegate(address _staker, uint256 _amount) external {
        // we don't checkStaker; one can undelegate even if the staker is not active

        require(msg.sender != _staker, "staker cannot undelegate to self");
        require(msg.sender != stakerInfo[_staker].operator, "operator cannot undelegate to self");

        // update stake info
        _updateRewardInfo(_staker, msg.sender);

        _subStaking(_staker, msg.sender, _amount);

        if (isStaker(_staker)) {
            _newCredential(_amount, GOV_CONST.UNBONDING_PERIOD_DELEGATOR());
        } else {
            danglingDelegated -= _amount;

            (bool success, ) = payable(msg.sender).call{value: _amount}("");
            require(success, "failed to send undelegating amount");
        }

        emit Undelegated(msg.sender, _staker, _amount);
    }

    function claim(address _staker, bool _restake) external {
        require(userRewardInfo[_staker][msg.sender].stakingAmount > 0 ||
                userRewardInfo[_staker][msg.sender].pendingReward > 0, "no reward to claim");
        require(_staker != address(0), "unregistered staker");
        Staker storage _stakerInfo = stakerInfo[_staker];
        UserInfo storage _userInfo = userRewardInfo[_staker][msg.sender];

        // update stake info
        _updateRewardInfo(_staker, msg.sender);

        uint256 _reward = _userInfo.pendingReward;
        uint256 _fee = 0;
        if (msg.sender != _stakerInfo.operator) {
            // staker himself(operator) does not pay fee
            _fee = _userInfo.pendingFee;
            _reward = _reward - _fee;
        }
        _userInfo.pendingReward = 0;
        _userInfo.pendingFee = 0;

        if (_restake) {
            require(isStaker(_staker), "unregistered staker");
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendRewardTo(payable(address(this)), _reward);

            _addStaking(_staker, msg.sender, _reward);
        } else {
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendRewardTo(payable(msg.sender), _reward);
        }

        if (_fee > 0) {
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendRewardTo(payable(_stakerInfo.feeRecipient), _fee);
        }

        _stakerInfo.lastRewardBalance = _stakerInfo.rewardee.balance;

        emit Claimed(_staker, msg.sender, _reward, _restake);
    }

    function withdraw(uint256 _cid) external {
        WithdrawalCredential storage _credential = credentials[_cid];
        require(_credential.status == WithdrawalStatus.Requested, "invalid credential");
        require(_credential.requester == msg.sender, "msg.sender is not requester");
        require(block.timestamp >= _credential.withdrawableTime, "not yet time to withdraw");

        _credential.status = WithdrawalStatus.Withdrawn; // modify status first to prevent reentrancy
        (bool success, ) = payable(_credential.requester).call{value: _credential.amount}("");
        require(success, "failed to send withdrawal amount");

        emit Withdrawn(_cid, msg.sender, _credential.amount);
    }

    function _updateRewardInfo(address _staker, address _user) internal {
        Staker storage _stakerInfo = stakerInfo[_staker];
        if (_stakerInfo.totalStaked > 0) {
            uint256 _accBalance = _stakerInfo.rewardee.balance - _stakerInfo.lastRewardBalance;
            uint256 _rewardPerStaking = _accBalance * GOV_CONST.REWARD_PRECISION() / _stakerInfo.totalStaked;
            _stakerInfo.accRewardPerStaking += _rewardPerStaking;
            _stakerInfo.accFeePerStaking += _rewardPerStaking * _stakerInfo.feeRate / GOV_CONST.FEE_PRECISION();
            _stakerInfo.lastRewardBalance = _stakerInfo.rewardee.balance;

            emit RewardInfoUpdated(
                _staker,
                _stakerInfo.totalStaked,
                _stakerInfo.rewardee.balance,
                _accBalance,
                _stakerInfo.accRewardPerStaking,
                _stakerInfo.accFeePerStaking
            );

            if (_user != address(0)) {
                UserInfo storage _userInfo = userRewardInfo[_staker][_user];
                _userInfo.pendingReward += _userInfo.stakingAmount * (_stakerInfo.accRewardPerStaking - _userInfo.rewardPerStaking) / GOV_CONST.REWARD_PRECISION();
                _userInfo.pendingFee += _userInfo.stakingAmount * (_stakerInfo.accFeePerStaking - _userInfo.feePerStaking) / GOV_CONST.REWARD_PRECISION();
                _userInfo.rewardPerStaking = _stakerInfo.accRewardPerStaking;
                _userInfo.feePerStaking = _stakerInfo.accFeePerStaking;

                emit UserRewardUpdated(_staker, _user, _userInfo.stakingAmount, _userInfo.pendingReward, _userInfo.rewardPerStaking, _userInfo.feePerStaking);
            }

            // if any expired request exists, then execute it
            if (pendingRequest[_staker].requestTime > 0 && block.timestamp - pendingRequest[_staker].requestTime >= GOV_CONST.CHANGE_FEE_DELAY()) {
                stakerInfo[_staker].feeRate = pendingRequest[_staker].newFeeRate;
                delete pendingRequest[_staker];
            }
        }
    }

    function _addStaking(address _staker, address _user, uint256 _amount) private {
        Staker storage _stakerInfo = stakerInfo[_staker];
        require(_stakerInfo.totalStaked + _amount <= GOV_CONST.MAXIMUM_STAKING(), "exceeded the maximum");
        UserInfo storage _userInfo = userRewardInfo[_staker][_user];

        _stakerInfo.totalStaked += _amount;
        _userInfo.stakingAmount += _amount;
        totalStaking += _amount;
    }

    function _subStaking(address _staker, address _user, uint256 _amount) private {
        Staker storage _stakerInfo = stakerInfo[_staker];
        UserInfo storage _userInfo = userRewardInfo[_staker][_user];
        require(_userInfo.stakingAmount >= _amount, "insufficient balance");

        totalStaking -= _amount;
        _stakerInfo.totalStaked -= _amount;
        _userInfo.stakingAmount -= _amount;
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

    function getStakerAmount(address _staker) external view returns (uint256) {
        return userRewardInfo[_staker][stakerInfo[_staker].operator].stakingAmount;
    }

    function getDelegatedAmount(address _staker) public view returns (uint256) {
        return stakerInfo[_staker].totalStaked - userRewardInfo[_staker][stakerInfo[_staker].operator].stakingAmount;
    }
}
