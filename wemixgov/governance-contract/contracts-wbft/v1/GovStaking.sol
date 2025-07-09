// SPDX-License-Identifier: MIT

pragma solidity ^0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

import "./GovConfig.sol";
import { GovRewardeeImp } from "./GovRewardeeImp.sol";
import { GovRewardee } from "./GovRewardee.sol";
import { IGovCouncil } from "./IGovCouncil.sol";

contract GovStaking {
    using EnumerableSet for EnumerableSet.AddressSet;

    struct RegisterStakerParams {
        uint256 amount;
        address staker;
        address feeRecipient;
        uint256 feeRate;
        bytes blsPK;
    }

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

    struct UserCredentialInfo {
        uint256 credentialIndex;
        uint256 withdrawalIndex;
        // credentialCount = credentialIndex - withdrawalIndex
    }

    struct WithdrawalCredential {
        uint256 amount;
        uint256 requestTime;
        uint256 withdrawableTime;
    }

    struct ChangingFeeRequest {
        uint256 newFeeRate;
        uint256 requestTime;
    }

    event StakerRegistered(
        address indexed staker,
        address operator,
        address rewardee,
        address feeRecipient,
        uint256 feeRate,
        uint256 staking,
        bytes blsPK
    );
    event OperatorTransferred(address indexed staker, address indexed newOperator);
    event Staked(address indexed staker, uint256 amount);
    event Unstaked(address indexed staker, uint256 amount);
    event StakerRemoved(address indexed staker);
    event Delegated(address indexed delegator, address indexed staker, uint256 amount);
    event Undelegated(address indexed delegator, address indexed staker, uint256 amount);
    event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding);
    event Withdrawn(address indexed requester, uint256 withdrawalIndex, uint256 amount);
    event RewardInfoUpdated(
        address indexed staker,
        uint256 totalStaked,
        uint256 balance,
        uint256 accBalance,
        uint256 accRewardPerStaking,
        uint256 accFeePerStaking
    );
    event UserRewardUpdated(
        address indexed staker,
        address indexed user,
        uint256 stakingAmount,
        uint256 pendingReward,
        uint256 pendingFee,
        uint256 accRewardPerStaking,
        uint256 accFeePerStaking
    );
    event Claimed(address indexed staker, address indexed rewardee, uint256 amount, bool restake);
    event ChangingFeeRateRequested(address indexed staker, uint256 oldFeeRate, uint256 newFeeRate);
    event ChangingFeeRateExecuted(address indexed staker, uint256 newFeeRate);
    event FeeRecipientChanged(address indexed staker, address oldRecipient, address newRecipient);
    event FeeRateChangeRequested(address indexed staker, uint256 oldFeeRate, uint256 newFeeRate);

    uint256 public constant BLS_PUBLIC_KEY_LENGTH = 48;
    uint256 public constant BLS_SIGNATURE_LENGTH = 96;
    uint256 public constant REWARD_PRECISION = 1e27;

    address public blsPoP; /// 0x0; Precompiled contract address for BLS PoP verification
    address public govConfig; // 0x1; assigned by consensus engine
    address public govRewardeeImp; // 0x2; assigned by consensus engine

    // this includes danglingDelegated
    uint256 public totalStaking; // 0x3

    // Staker
    // Staker state definition
    //  0. unregistered: stakerInfo[staker].operator = 0, __stakerSet.contains(staker) = false
    //  1. active: stakerInfo[staker].operator != 0, __stakerSet.contains(staker) = true
    //  2. inactive: stakerInfo[staker].operator != 0, __stakerSet.contains(staker) = false
    EnumerableSet.AddressSet private __stakerSet; // 0x4, 0x5
    mapping(address => Staker) public stakerInfo; // 0x6
    mapping(address => address) public stakerByOperator; // 0x7
    mapping(address => address) public stakerByRewardee; // 0x8
    mapping(bytes => address) public stakerByBLSPublicKey; // 0x9

    // Withdrawal Credential: credentials[user][credentialIndex]
    mapping(address => mapping(uint256 => WithdrawalCredential)) public credentials; // 0xa
    mapping(address => UserCredentialInfo) public userCredential; // 0xb

    // pending request
    mapping(address => ChangingFeeRequest) public changingFeeRequests; // 0xc

    // User Reward Info
    mapping(address => mapping(address => UserInfo)) public userRewardInfo; // 0xd

    // danglingDelegated is the delegated balance for the inactive stakers
    // contract's balance = totalStaked + danglingDelegated + unbonding
    uint256 public danglingDelegated; // 0xe

    //***********************************************************************
    //* Caution for Upgrading
    //* - If you add new state variables, please add them after this comment
    //* - Never modify existing state variables
    //***********************************************************************

    // state definition
    // - UNREGISTERED: stakerInfo[staker].operator = 0
    // - REGISTERED: stakerInfo[staker].operator != 0
    //   - ACTIVE: __stakerSet.contains(staker) = true
    //   - INACTIVE: __stakerSet.contains(staker) = false
    //
    //
    // staker operations and states in which they are permitted to execute
    // - registerStaker: UNREGISTERED
    // - changeFeeRecipient: REGISTERED
    // - requestChangingFee: ACTIVE
    // - stake: REGISTERED
    // - unstake: ACTIVE
    // - delegate: ACTIVE
    // - undelegate: REGISTERED
    // - claim: REGISTERED
    // - withdraw: REGISTERED

    modifier isNotRegistered(address _staker) {
        require(stakerInfo[_staker].operator == address(0), "already registered staker");
        _;
    }

    modifier isRegistered(address _staker) {
        require(stakerInfo[_staker].operator != address(0), "unregistered staker");
        _;
    }

    modifier isActive(address _staker) {
        require(stakerInfo[_staker].operator != address(0), "unregistered staker");
        require(__stakerSet.contains(_staker), "staker is inactive");
        _;
    }

    modifier isInactive(address _staker) {
        require(stakerInfo[_staker].operator != address(0), "unregistered staker");
        require(!__stakerSet.contains(_staker), "staker is active");
        _;
    }

    modifier checkAmount(uint256 _amount) {
        require(msg.value == _amount, "amount and msg.value mismatch");
        _;
    }

    modifier inspectWithCouncil(bytes4 selector, bytes memory arguments) {
        address _govCouncil = GovConfig(govConfig).govCouncil();
        if (_govCouncil != address(0)) {
            // if the GovCouncil is set, then check an operation with governance council
            require(IGovCouncil(_govCouncil).inspectOperation(selector, msg.sender, arguments), "operation not permitted by council");
        }
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

    function registerStaker(
        uint256 _amount,
        address _staker,
        address _feeRecipient,
        uint256 _feeRate,
        bytes calldata _blsPK,
        bytes calldata _blsSig
    )
        external
        payable
        isNotRegistered(_staker)
        checkAmount(_amount)
        inspectWithCouncil(GovStaking.registerStaker.selector, abi.encode(_amount, _staker, _feeRecipient, _feeRate, _blsPK))
    {
        require(_amount >= GovConfig(govConfig).minimumStaking() && _amount <= GovConfig(govConfig).maximumStaking(), "out of bounds");
        require(msg.sender != _staker, "operator cannot be staker");
        require(_staker != address(0), "zero address");
        require(!isOperator(msg.sender), "operator is already registered");
        require(_feeRecipient != address(0), "fee recipient is zero address");
        require(_feeRate <= GovConfig(govConfig).feePrecision(), "fee rate exceeds precision");
        _checkBLSPublicKey(_blsPK, _blsSig);

        RegisterStakerParams memory _params = RegisterStakerParams({
            amount: _amount,
            staker: _staker,
            feeRecipient: _feeRecipient,
            feeRate: _feeRate,
            blsPK: _blsPK
        });
        _registerStaker(_params);
    }

    function _registerStaker(RegisterStakerParams memory _params) internal {
        GovRewardee _rewardee = new GovRewardee(govRewardeeImp);
        GovRewardeeImp(payable(address(_rewardee))).initialize(address(this));

        stakerInfo[_params.staker].operator = msg.sender;
        stakerInfo[_params.staker].rewardee = address(_rewardee);
        stakerInfo[_params.staker].feeRecipient = _params.feeRecipient;
        stakerInfo[_params.staker].feeRate = _params.feeRate;
        stakerInfo[_params.staker].blsPubKey = _params.blsPK;

        stakerByOperator[msg.sender] = _params.staker;
        stakerByRewardee[address(_rewardee)] = _params.staker;
        stakerByBLSPublicKey[_params.blsPK] = _params.staker;

        __stakerSet.add(_params.staker);

        _addStaking(_params.staker, _params.staker, _params.amount);

        emit StakerRegistered(_params.staker, msg.sender, address(_rewardee), _params.feeRecipient, _params.feeRate, _params.amount, _params.blsPK);
    }

    function _checkBLSPublicKey(bytes calldata _blsPK, bytes calldata _blsSig) internal view {
        require(_blsPK.length == BLS_PUBLIC_KEY_LENGTH, "invalid bls public key length");
        require(_blsSig.length == BLS_SIGNATURE_LENGTH, "invalid bls signature length");
        require(stakerByBLSPublicKey[_blsPK] == address(0), "already registered bls public key");

        (bool _success, bytes memory _result) = blsPoP.staticcall(abi.encodePacked(_blsPK, _blsSig));
        require(_success, "failed to verify bls pop");
        require(abi.decode(_result, (bool)), "invalid bls public key");
    }

    function transferOperatorShip(
        address _newOperator
    ) external isRegistered(stakerByOperator[msg.sender]) inspectWithCouncil(GovStaking.transferOperatorShip.selector, abi.encode(_newOperator)) {
        require(_newOperator != address(0), "zero address");
        require(stakerByOperator[_newOperator] == address(0), "new operator is already registered as other operator");
        address _staker = stakerByOperator[msg.sender];
        require(_newOperator != _staker, "cannot transfer to self");

        // update operator
        stakerInfo[_staker].operator = _newOperator;
        stakerByOperator[_newOperator] = _staker;
        delete stakerByOperator[msg.sender];

        emit OperatorTransferred(_staker, _newOperator);
    }

    function changeFeeRecipient(
        address _newRecipient
    ) external isRegistered(stakerByOperator[msg.sender]) inspectWithCouncil(GovStaking.changeFeeRecipient.selector, abi.encode(_newRecipient)) {
        require(_newRecipient != address(0), "zero address");
        address _staker = stakerByOperator[msg.sender];
        address oldRecipient = stakerInfo[_staker].feeRecipient;
        stakerInfo[_staker].feeRecipient = _newRecipient;

        emit FeeRecipientChanged(_staker, oldRecipient, _newRecipient);
    }

    function requestChangingFee(
        uint256 _feeRate
    ) external isActive(stakerByOperator[msg.sender]) inspectWithCouncil(GovStaking.requestChangingFee.selector, abi.encode(_feeRate)) {
        require(_feeRate <= GovConfig(govConfig).feePrecision(), "fee rate exceeds precision");
        address _staker = stakerByOperator[msg.sender];
        require(changingFeeRequests[_staker].requestTime == 0, "request already is on going");

        uint256 oldFeeRate = stakerInfo[_staker].feeRate;
        if (getDelegatedAmount(_staker) > 0) {
            changingFeeRequests[_staker] = ChangingFeeRequest({ newFeeRate: _feeRate, requestTime: block.timestamp });
        } else {
            // if no delegator exists, change fee immediately
            stakerInfo[_staker].feeRate = _feeRate;
        }

        emit ChangingFeeRateRequested(_staker, oldFeeRate, _feeRate);
    }

    function executeChangingFee(address _staker) external inspectWithCouncil(GovStaking.executeChangingFee.selector, abi.encode(_staker)) {
        require(changingFeeRequests[_staker].requestTime > 0, "no request exists");
        require(
            block.timestamp - changingFeeRequests[_staker].requestTime >= GovConfig(govConfig).changeFeeDelay(),
            "the request cannot be executed before delay time"
        );

        // don't update user info passing zero address
        _updateRewardInfo(_staker, address(0));
    }

    function stake(
        uint256 _amount
    )
        external
        payable
        isRegistered(stakerByOperator[msg.sender])
        checkAmount(_amount)
        inspectWithCouncil(GovStaking.stake.selector, abi.encode(_amount))
    {
        address _staker = stakerByOperator[msg.sender];

        // update stake info
        _updateRewardInfo(_staker, _staker);

        if (!isStaker(_staker)) {
            // reactivation case: if the staker is not active, then reactivate it

            require(_amount >= GovConfig(govConfig).minimumStaking(), "amount is less than minimum staking");

            __stakerSet.add(_staker);
            danglingDelegated -= stakerInfo[_staker].totalStaked;
        }

        _addStaking(_staker, _staker, _amount);

        emit Staked(_staker, _amount);
    }

    function unstake(
        uint256 _amount
    ) external isActive(stakerByOperator[msg.sender]) inspectWithCouncil(GovStaking.unstake.selector, abi.encode(_amount)) {
        require(_amount > 0, "amount is zero");

        address _staker = stakerByOperator[msg.sender];

        // update stake info
        _updateRewardInfo(_staker, _staker);

        _subStaking(_staker, _staker, _amount);

        UserInfo storage _userInfo = userRewardInfo[_staker][_staker];
        if (_userInfo.stakingAmount < GovConfig(govConfig).minimumStaking()) {
            require(_userInfo.stakingAmount == 0, "amount must equal balance to deactivate staker");

            __stakerSet.remove(_staker);

            danglingDelegated += stakerInfo[_staker].totalStaked;

            emit StakerRemoved(_staker);
        }

        _newCredential(_amount, GovConfig(govConfig).unbondingPeriodStaker());

        emit Unstaked(_staker, _amount);
    }

    function delegate(
        address _staker,
        uint256 _amount
    ) external payable isActive(_staker) checkAmount(_amount) inspectWithCouncil(GovStaking.delegate.selector, abi.encode(_staker, _amount)) {
        require(msg.sender != _staker, "staker cannot delegate to self");
        require(msg.sender != stakerInfo[_staker].operator, "operator cannot delegate to self");

        _updateRewardInfo(_staker, msg.sender);

        _addStaking(_staker, msg.sender, _amount);

        emit Delegated(msg.sender, _staker, _amount);
    }

    function undelegate(
        address _staker,
        uint256 _amount
    ) external isRegistered(_staker) inspectWithCouncil(GovStaking.undelegate.selector, abi.encode(_staker, _amount)) {
        require(msg.sender != _staker, "staker cannot undelegate to self");
        require(msg.sender != stakerInfo[_staker].operator, "operator cannot undelegate to self");

        // update stake info
        _updateRewardInfo(_staker, msg.sender);

        _subStaking(_staker, msg.sender, _amount);

        if (isStaker(_staker)) {
            _newCredential(_amount, GovConfig(govConfig).unbondingPeriodDelegator());
        } else {
            danglingDelegated -= _amount;

            (bool success, ) = payable(msg.sender).call{ value: _amount }("");
            require(success, "failed to send undelegating amount");
        }

        emit Undelegated(msg.sender, _staker, _amount);
    }

    function claim(
        address _staker,
        bool _restake
    ) external isRegistered(_staker) inspectWithCouncil(GovStaking.claim.selector, abi.encode(_staker, _restake)) {
        address _user = isOperator(msg.sender) ? _staker : msg.sender;
        require(userRewardInfo[_staker][_user].stakingAmount > 0 || userRewardInfo[_staker][_user].pendingReward > 0, "no reward to claim");
        Staker storage _stakerInfo = stakerInfo[_staker];
        UserInfo storage _userInfo = userRewardInfo[_staker][_user];
        // update stake info
        _updateRewardInfo(_staker, _user);

        uint256 _reward = _userInfo.pendingReward;
        uint256 _fee = 0;
        if (_user != _staker) {
            // staker himself does not pay fee
            _fee = _userInfo.pendingFee;
            _reward = _reward - _fee;
        }
        _userInfo.pendingReward = 0;
        _userInfo.pendingFee = 0;

        if (_restake) {
            require(isStaker(_staker), "staker is inactive");
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendRewardTo(payable(address(this)), _reward);

            _addStaking(_staker, _user, _reward);
        } else {
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendRewardTo(payable(msg.sender), _reward);
        }

        if (_fee > 0) {
            GovRewardeeImp(payable(_stakerInfo.rewardee)).sendFeeTo(payable(_stakerInfo.feeRecipient), _fee);
        }

        _stakerInfo.lastRewardBalance = _stakerInfo.rewardee.balance;
        emit Claimed(_staker, msg.sender, _reward, _restake);
    }

    function withdraw(uint256 _withdrawalCount) external inspectWithCouncil(GovStaking.withdraw.selector, abi.encode(_withdrawalCount)) {
        UserCredentialInfo storage _userCredential = userCredential[msg.sender];
        require(_userCredential.credentialIndex > _userCredential.withdrawalIndex, "no credential to withdraw");

        uint256 _lastIndex = _userCredential.credentialIndex;
        if (_withdrawalCount > 0) {
            _lastIndex = _userCredential.withdrawalIndex + _withdrawalCount;
            require(_lastIndex <= _userCredential.credentialIndex, "out of max user credential index");
            require(
                credentials[msg.sender][_userCredential.withdrawalIndex + _withdrawalCount - 1].withdrawableTime <= block.timestamp,
                "withdrawal time not reached"
            );
        }
        for (uint256 i = _userCredential.withdrawalIndex; i < _lastIndex; i++) {
            WithdrawalCredential storage _credential = credentials[msg.sender][i];
            if (_withdrawalCount == 0 && block.timestamp < _credential.withdrawableTime) {
                break;
            }
            _userCredential.withdrawalIndex++;
            uint256 _withdrawalIndex = _userCredential.withdrawalIndex;
            uint256 _amount = _credential.amount;
            delete credentials[msg.sender][i];

            (bool success, ) = payable(msg.sender).call{ value: _amount }("");
            require(success, "failed to send withdrawal amount");

            emit Withdrawn(msg.sender, _withdrawalIndex, _amount);
        }
    }

    function _updateRewardInfo(address _staker, address _user) private {
        Staker storage _stakerInfo = stakerInfo[_staker];

        if (_stakerInfo.totalStaked > 0) {
            uint256 _accBalance = _stakerInfo.rewardee.balance - _stakerInfo.lastRewardBalance;
            uint256 _rewardPerStaking = (_accBalance * REWARD_PRECISION) / _stakerInfo.totalStaked;
            _stakerInfo.accRewardPerStaking += _rewardPerStaking;
            _stakerInfo.accFeePerStaking += (_rewardPerStaking * _stakerInfo.feeRate) / GovConfig(govConfig).feePrecision();
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
                _userInfo.pendingReward +=
                    (_userInfo.stakingAmount * (_stakerInfo.accRewardPerStaking - _userInfo.rewardPerStaking)) /
                    REWARD_PRECISION;
                _userInfo.pendingFee += (_userInfo.stakingAmount * (_stakerInfo.accFeePerStaking - _userInfo.feePerStaking)) / REWARD_PRECISION;
                _userInfo.rewardPerStaking = _stakerInfo.accRewardPerStaking;
                _userInfo.feePerStaking = _stakerInfo.accFeePerStaking;

                emit UserRewardUpdated(
                    _staker,
                    _user,
                    _userInfo.stakingAmount,
                    _userInfo.pendingReward,
                    _userInfo.pendingFee,
                    _userInfo.rewardPerStaking,
                    _userInfo.feePerStaking
                );
            }
        }

        // if any expired request exists, then execute it
        if (
            changingFeeRequests[_staker].requestTime > 0 &&
            block.timestamp - changingFeeRequests[_staker].requestTime >= GovConfig(govConfig).changeFeeDelay()
        ) {
            stakerInfo[_staker].feeRate = changingFeeRequests[_staker].newFeeRate;
            delete changingFeeRequests[_staker];

            emit ChangingFeeRateExecuted(_staker, stakerInfo[_staker].feeRate);
        }
    }

    function _addStaking(address _staker, address _user, uint256 _amount) private {
        Staker storage _stakerInfo = stakerInfo[_staker];
        require(_stakerInfo.totalStaked + _amount <= GovConfig(govConfig).maximumStaking(), "exceeded the maximum");
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
        UserCredentialInfo storage _userCredential = userCredential[msg.sender];
        credentials[msg.sender][_userCredential.credentialIndex] = WithdrawalCredential({
            amount: _amount,
            requestTime: block.timestamp,
            withdrawableTime: block.timestamp + _unbondingPeriod
        });

        _userCredential.credentialIndex++;
        emit NewCredential(_userCredential.credentialIndex, msg.sender, _amount, block.timestamp, _unbondingPeriod);
    }

    function getStakerAmount(address _staker) external view returns (uint256) {
        return userRewardInfo[_staker][_staker].stakingAmount;
    }

    function getDelegatedAmount(address _staker) public view returns (uint256) {
        return stakerInfo[_staker].totalStaked - userRewardInfo[_staker][_staker].stakingAmount;
    }
}
