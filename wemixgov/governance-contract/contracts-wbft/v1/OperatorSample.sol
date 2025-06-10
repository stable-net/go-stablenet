// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import { IFeeRecipient, IERC165 } from "./IFeeRecipient.sol";
import "./IMultiSigWallet.sol";

contract OperatorSample is IMultiSigWallet, IFeeRecipient {
    /* =========== STATE VARIABLES ===========*/

    // Constants
    uint256 public constant MAX_OWNER_COUNT = 50; // number of max owner
    address public constant GOV_STAKING = address(0x1001);
    uint8 private constant NOT_ENTERED = 1;
    uint8 private constant ENTERED = 2;
    bytes4 public constant GOV_CLAIM_SELECTOR = bytes4(keccak256("claim(address,bool)"));
    bytes4 public constant CLAIM_WITHOUT_RESTAKE_SELECTOR = bytes4(keccak256("claimWithoutRestake(address)"));
    bytes4 public constant WITHDRAW_SELECTOR = bytes4(keccak256("withdraw(uint256)"));
    bytes4 public constant REGISTER_STAKER_SELECTOR = bytes4(keccak256("registerStaker(uint256,address,address,uint256,bytes)"));
    bytes4 public constant STAKE_SELECTOR = bytes4(keccak256("stake(uint256)"));

    uint8 private __status;
    bool private __receivingRewardStat;

    address[] public owners; // owners address set
    address[] public fundManagers;
    Transaction[] public transactions; // transaction struct set

    uint256 public quorum; // minimum required confirmation number
    uint256 public unstakedAmount;
    uint256 public rewardAmount;
    uint256 public feeAmount;

    // Mappings (each takes a full slot)
    mapping(bytes32 => uint256) public proposalHashToTxId; // mapping from hash(proposer,blockNumber) =>  transaction id
    mapping(address => bool) public isOwner; // mapping from owner => bool
    mapping(uint256 => mapping(address => bool)) public isConfirmed; // mapping from transaction id => owner => bool
    mapping(address => bool) public isFundManager;

    /* =========== EVENTS  ===========*/

    event ReceivedUnstaked(address indexed from, uint256 amount);
    event ReceivedReward(address indexed from, uint256 amount);
    event SentUnstakedAmount(address indexed to, uint256 amount);
    event SentRewardAmount(address indexed to, uint256 amount);

    event AddFundManager(address indexed newFundManager);
    event RemoveFundManager(address indexed fundManager);

    /* =========== MODIFIERS  ===========*/

    modifier onlyOwner() {
        require(isOwner[msg.sender], "Operator: only owner can access");
        _;
    }

    modifier isOneOfOwner(address _owner) {
        require(isOwner[_owner], "Operator: only owner can access");
        _;
    }

    modifier isNotOneOfOwner(address _owner) {
        require(!isOwner[_owner], "Operator: the address is already an owner");
        _;
    }

    modifier onlyWalletOrSingleOwner() {
        // if contract has single owner, multiSig wallet is not activated
        // owner can execute this function right away
        if (owners.length == 1) {
            require(msg.sender == owners[0], "Operator: only owner can access");
        } else {
            require(msg.sender == address(this), "Operator: only wallet can access");
        }
        _;
    }

    modifier notNull(address _address) {
        require(_address != address(0), "Operator: owner cannot be zero address");
        _;
    }

    modifier isTransactionExist(uint256 _transactionId) {
        require(_transactionId <= transactions.length, "Operator: transaction does not exist");
        _;
    }

    modifier notExecuted(uint256 _transactionId) {
        require(!transactions[_transactionId].executed, "Operator: transaction is already executed");
        _;
    }

    modifier notConfirmed(uint256 _transactionId) {
        require(!isConfirmed[_transactionId][msg.sender], "Operator: transaction is already confirmed");
        _;
    }

    modifier validQuorumRequirement(uint256 _ownerCount, uint256 _quorum) {
        require(_quorum <= _ownerCount && _quorum != 0, "Operator: invalid quorum requirement");
        _;
    }

    modifier onlyFundManager() {
        require(isFundManager[msg.sender], "Operator: only fundManager can access");
        _;
    }

    modifier nonReentrant() {
        require(__status != ENTERED, "Operator: reentrant call");
        __status = ENTERED;
        _;
        __status = NOT_ENTERED;
    }

    constructor(address[] memory _owners, address[] memory _fundManagers, uint256 _quorum) {
        require(_owners.length > 0, "Operator: owners length must be at least 1");

        if (_owners.length > 1) {
            require(_quorum >= 1 && _quorum <= _owners.length, "Operator: number of confirmations does not satisfy quorum");
            quorum = _quorum;
        } else {
            // if owners.length is 1, set quorum as 1
            quorum = 1;
        }

        for (uint256 i = 0; i < _owners.length; i++) {
            address owner = _owners[i];

            require(owner != address(0), "Operator: owner cannot be zero address");
            require(!isOwner[owner], "Operator: owner address is duplicated");

            isOwner[owner] = true;
            owners.push(owner);
        }

        for (uint256 i = 0; i < _fundManagers.length; i++) {
            address fm = _fundManagers[i];

            require(fm != address(0), "Operator: fundManager cannot be zero address");
            require(!isFundManager[fm], "Operator: fundManager address is duplicated");

            isFundManager[fm] = true;
            fundManagers.push(fm);
        }

        __status = NOT_ENTERED;
        // fill index 0 with empty Transaction
        transactions.push(Transaction({ to: address(0), value: 0, data: new bytes(0), executed: false, currentNumberOfConfirmations: 0 }));
    }

    // @notice Function is to receive reward, unstaked, or ether that eoa sent to the contract.
    // @dev When _receivingRewardStat is true, means the received ether is reward value so increase the _rewardAmount.
    // Else, increase the _unstakedAmount
    receive() external payable {
        if (__receivingRewardStat) {
            rewardAmount += msg.value;
            emit ReceivedReward(msg.sender, msg.value);
        } else {
            unstakedAmount += msg.value;
            emit ReceivedUnstaked(msg.sender, msg.value);
        }
        emit Deposit(msg.sender, msg.value, address(this).balance);
    }

    /* ========== EXTERNAL FUNCTION ========== */

    // implements IERC165
    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return interfaceId == type(IFeeRecipient).interfaceId || interfaceId == type(IERC165).interfaceId;
    }

    // implements IFeeRecipient
    function receiveFee(uint256 _amount) external payable {
        require(_amount == msg.value, "Operator: fee amount and value sent mismatched");
        feeAmount += _amount;
        emit ReceivedFee(msg.sender, _amount);
    }

    // implements IFeeRecipient
    // @notice Function is for sending _to a fee amount
    // Only fundManger can access this function.
    function withdrawFee(address _to, uint256 _amount) external onlyFundManager notNull(_to) nonReentrant {
        require(_amount <= feeAmount && _amount > 0, "Operator: invalid amount");
        feeAmount -= _amount;
        (bool success, ) = payable(_to).call{ value: _amount }("");
        require(success, "Operator: failed to send fee amount");
        emit SentFeeAmount(_to, _amount);
    }

    // @notice Function is for sending _to a reward amount
    // Only fundManger can access this function.
    function withdrawReward(address _to, uint256 _amount) external onlyFundManager notNull(_to) nonReentrant {
        require(_amount <= rewardAmount && _amount > 0, "Operator: invalid amount");
        rewardAmount -= _amount;
        (bool success, ) = payable(_to).call{ value: _amount }("");
        require(success, "Operator: failed to send reward amount");
        emit SentRewardAmount(_to, _amount);
    }

    // @notice Function is for sending _to a unstaked amount
    // Note that fundManager has no right to manage unstaked amount. It is only managed by owners.
    function withdrawUnstaked(address _to, uint256 _amount) external onlyWalletOrSingleOwner notNull(_to) nonReentrant {
        require(_amount <= unstakedAmount, "Operator: withdraw amount is greater than unstaked amount");
        unstakedAmount -= _amount;
        (bool success, ) = payable(_to).call{ value: _amount }("");
        require(success, "Operator: failed to send unstaked amount");
        emit SentUnstakedAmount(_to, _amount);
    }

    // @notice Function calls the registerStaker method of the GovStaking contract with Ether value
    // You cannot directly submit a transaction to GovStaking's registerStaker method since it's blocked for tracking amounts.
    // Instead, you must submit a transaction that calls this wrapper function, or just call this function directly if there's only one owner.
    function registerStaker(
        uint256 _amount,
        address _staker,
        address _feeRecipient,
        uint256 _feeRate,
        bytes calldata _blsPK
    ) external onlyWalletOrSingleOwner nonReentrant {
        require(_amount <= unstakedAmount, "Operator : amount is greater than unstaked amount");
        unstakedAmount -= _amount;
        bytes memory data = abi.encodeWithSignature(
            "registerStaker(uint256,address,address,uint256,bytes)",
            _amount,
            _staker,
            _feeRecipient,
            _feeRate,
            _blsPK
        );
        (bool success, bytes memory returnData) = GOV_STAKING.call{ value: _amount }(data);
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: registerStaker tx failed"));
        }
        emit SentUnstakedAmount(GOV_STAKING, _amount);
    }

    // @notice Function calls the stake method of the GovStaking contract with Ether value
    // You cannot directly submit a transaction to GovStaking's stake method since it's blocked for tracking amounts.
    // Instead, you must submit a transaction that calls this wrapper function, or just call this function directly if there's only one owner.
    function stake(uint256 _amount) external onlyWalletOrSingleOwner nonReentrant {
        require(_amount <= unstakedAmount, "Operator: amount is greater than unstaked amount");
        unstakedAmount -= _amount;
        bytes memory data = abi.encodeWithSignature("stake(uint256)", _amount);
        (bool success, bytes memory returnData) = GOV_STAKING.call{ value: _amount }(data);
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: stake tx failed"));
        }
        emit SentUnstakedAmount(GOV_STAKING, _amount);
    }

    // @notice Function calls the unstake method of the GovStaking
    // You may submit a transaction calling GovStaking's unstake method,
    // Or if it's single user, just call this function right away.
    function unstake(uint256 _amount) external onlyWalletOrSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature("unstake(uint256)", _amount);
        (bool success, bytes memory returnData) = GOV_STAKING.call(data);
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: unstake tx failed"));
        }
    }

    // @notice Function calls the withdraw method of the GovStaking
    // Owner can directly call this method without multiSig signing.
    function withdraw(uint256 _withdrawalCount) external onlyOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature("withdraw(uint256)", _withdrawalCount);
        (bool success, bytes memory returnData) = GOV_STAKING.call(data);
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: withdraw tx failed"));
        }
    }

    // @notice Function calls the claim method of the GovStaking contract with restake option as false
    // Owner can directly call this method without multiSig signing.
    function claimWithoutRestake(address _staker) external onlyOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature("claim(address,bool)", _staker, false);
        __receivingRewardStat = true;
        (bool success, bytes memory returnData) = GOV_STAKING.call(data);
        __receivingRewardStat = false;
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: claim tx failed"));
        }
    }

    function claimWithRestake(address _staker) external onlyWalletOrSingleOwner nonReentrant {
        bytes memory data = abi.encodeWithSignature("claim(address,bool)", _staker, true);
        (bool success, bytes memory returnData) = GOV_STAKING.call(data);
        if (!success) {
            revert(_getRevertMsg(returnData, "Operator: claim tx failed"));
        }
    }

    // @notice Function to add owner
    // Added owner will be automatically added to claimer
    function addOwner(address _newOwner, bool _increaseQuorum) external onlyWalletOrSingleOwner notNull(_newOwner) isNotOneOfOwner(_newOwner) {
        require(owners.length + 1 <= MAX_OWNER_COUNT, "Operator: max owner count exceeded");
        isOwner[_newOwner] = true;
        owners.push(_newOwner);
        if (_increaseQuorum) {
            require(quorum + 1 <= owners.length, "Operator: quorum cannot be more than owner count");
            quorum += 1;
            emit ChangeQuorum(quorum);
        }
        emit AddOwner(_newOwner);
    }

    // @notice Function to add remove owner
    // Removed owner will be automatically removed from claimer
    function removeOwner(address _owner, bool _reduceQuorum) external onlyWalletOrSingleOwner isOneOfOwner(_owner) {
        require(owners.length > 1, "Operator: cannot remove single owner");
        isOwner[_owner] = false;

        for (uint256 i = 0; i < owners.length; ) {
            if (owners[i] == _owner) {
                owners[i] = owners[owners.length - 1];
                owners.pop();
                break;
            }
            unchecked {
                i++;
            }
        }
        if (_reduceQuorum) {
            require(quorum > 1, "Operator: quorum cannot be less than 1");
            quorum -= 1;
            emit ChangeQuorum(quorum);
        } else if (quorum > owners.length) {
            quorum = owners.length;
            emit ChangeQuorum(owners.length);
        }

        emit RemoveOwner(_owner);
    }

    // @notice Function to replace owner
    // Claimer will also be replaced
    function replaceOwner(address _owner, address _newOwner) external onlyWalletOrSingleOwner isOneOfOwner(_owner) isNotOneOfOwner(_newOwner) {
        for (uint i = 0; i < owners.length; ) {
            if (owners[i] == _owner) {
                owners[i] = _newOwner;
                break;
            }
            unchecked {
                i++;
            }
        }

        isOwner[_owner] = false;
        isOwner[_newOwner] = true;

        emit RemoveOwner(_owner);
        emit AddOwner(_newOwner);
    }

    // @notice Function to change quorum
    function changeQuorum(uint256 _quorum) external onlyWalletOrSingleOwner validQuorumRequirement(owners.length, _quorum) {
        quorum = _quorum;
        emit ChangeQuorum(_quorum);
    }

    // @notice Function to add fundManager
    function addFundManager(address _newFundManager) external onlyWalletOrSingleOwner notNull(_newFundManager) {
        require(!isFundManager[_newFundManager], "Operator: already registered fundManager");
        isFundManager[_newFundManager] = true;
        fundManagers.push(_newFundManager);
        emit AddFundManager(_newFundManager);
    }

    // @notice Function to remove fundManager
    function removeFundManager(address _fundManager) external onlyWalletOrSingleOwner notNull(_fundManager) {
        require(isFundManager[_fundManager], "Operator: fundManager is not registered");
        isFundManager[_fundManager] = false;

        for (uint256 i = 0; i < fundManagers.length; ) {
            if (fundManagers[i] == _fundManager) {
                fundManagers[i] = fundManagers[fundManagers.length - 1];
                fundManagers.pop();
                break;
            }
            unchecked {
                i++;
            }
        }
        emit RemoveFundManager(_fundManager);
    }

    /* ========== PUBLIC FUNCTION ========== */

    // @notice Function to submit a transaction for multiSig signing.
    // Only owner can access.
    function submitTransaction(address _to, uint256 _value, bytes memory _data) public onlyOwner {
        bytes32 proposalHash = keccak256(abi.encodePacked(msg.sender, block.number));
        require(proposalHashToTxId[proposalHash] == 0, "Operator: duplicate proposal in same block");

        // transaction id starts with 1
        uint256 transactionId = transactions.length;
        proposalHashToTxId[proposalHash] = transactionId;
        bytes4 selector = bytes4(_data);
        if (_to == GOV_STAKING) {
            // if the transaction destination is govStaking, registerStaker and stake function is not allowed
            require(selector != REGISTER_STAKER_SELECTOR && selector != STAKE_SELECTOR, "Operator: use proper stake functions");
            require(selector != GOV_CLAIM_SELECTOR && selector != WITHDRAW_SELECTOR, "Operator: use proper claim/withdraw functions");
        }

        if (_to == address(this)) {
            // claimWithoutRestake and withdraw function don't need multiSig. Owner cal call this function directly.
            // to prevent confusion, we block this function here.
            require(selector != CLAIM_WITHOUT_RESTAKE_SELECTOR && selector != WITHDRAW_SELECTOR, "Operator: use proper claim/withdraw functions");
        }

        if (_value > 0) {
            // The only way to withdraw ETH will be through the proper withdrawal functions, which will be executed with zero value in the transaction itself.
            revert("Operator: use proper withdraw functions to transfer value from the contract");
        }
        transactions.push(Transaction({ to: _to, value: _value, data: _data, executed: false, currentNumberOfConfirmations: 0 }));
        emit SubmitTransaction(msg.sender, transactionId, _to, _value, _data);
    }

    function executeTransaction(uint256 _transactionId) public payable onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];
        require(
            transaction.currentNumberOfConfirmations >= quorum,
            "Operator: current number of confirmations must be greater than or equal to quorum"
        );
        transaction.executed = true;
        (bool success, bytes memory returnData) = transaction.to.call{ value: 0 }(transaction.data);
        if (!success) {
            revert(_getRevertMsg(returnData, string(abi.encodePacked("Operator: transaction failed"))));
        }
        emit ExecuteTransaction(msg.sender, _transactionId);
    }

    // @notice Function to confirm a transaction submitted.
    // Only owner can access.
    function confirmTransaction(
        uint256 _transactionId
    ) public onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) notConfirmed(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];
        transaction.currentNumberOfConfirmations += 1;
        isConfirmed[_transactionId][msg.sender] = true;

        emit ConfirmTransaction(msg.sender, _transactionId);
    }

    // @notice Function to revoke confirmation.
    // Only owner who confirmed the transaction can access.
    function revokeConfirmation(uint256 _transactionId) public onlyOwner isTransactionExist(_transactionId) notExecuted(_transactionId) {
        Transaction storage transaction = transactions[_transactionId];

        require(isConfirmed[_transactionId][msg.sender], "Operator: transaction is not confirmed");

        transaction.currentNumberOfConfirmations -= 1;
        isConfirmed[_transactionId][msg.sender] = false;

        emit RevokeConfirmation(msg.sender, _transactionId);
    }

    // @notice Get Owners.
    function getOwners() public view returns (address[] memory) {
        return owners;
    }

    // @notice Get Owners Count.
    function getOwnerCount() public view returns (uint256) {
        return owners.length;
    }

    // @notice Get FundManagers.
    function getFundManagers() public view returns (address[] memory) {
        return fundManagers;
    }

    // @notice Get Transaction.
    function getTransaction(
        uint256 _transactionId
    ) public view returns (address to, uint256 value, bytes memory data, bool executed, uint256 currentNumberOfConfirmations) {
        Transaction storage transaction = transactions[_transactionId];

        return (transaction.to, transaction.value, transaction.data, transaction.executed, transaction.currentNumberOfConfirmations);
    }

    // @notice Get Transaction Count.
    function getTransactionCount() public view returns (uint256) {
        return transactions.length;
    }

    /* ========== PRIVATE FUNCTION ========== */

    // Helper function to extract revert reason from call returnData
    function _getRevertMsg(bytes memory _returnData, string memory _defaultMsg) private pure returns (string memory) {
        // If return data is at least 4 bytes (error selector) + some data
        if (_returnData.length > 4) {
            // Try to decode the standard error message
            bytes4 errorSelector;
            assembly {
                errorSelector := mload(add(_returnData, 0x20))
            }

            // Check if this is a standard error (Error(string))
            if (errorSelector == 0x08c379a0) {
                // Standard revert/require reason - decode manually to avoid slice error
                bytes memory slicedData = new bytes(_returnData.length - 4);
                for (uint i = 4; i < _returnData.length; i++) {
                    slicedData[i - 4] = _returnData[i];
                }
                string memory reason = abi.decode(slicedData, (string));
                return string(abi.encodePacked(_defaultMsg, ": ", reason));
            }
        }

        // Default message if no specific reason found
        return _defaultMsg;
    }
}
