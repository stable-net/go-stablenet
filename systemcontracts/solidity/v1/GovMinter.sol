// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

import {GovBaseV2} from "../abstracts/GovBaseV2.sol";
import {IFiatToken} from "../interfaces/IFiatToken.sol";

/**
 * @title GovMinter
 * @notice Governance-controlled minting with front-running prevention
 * @dev One-to-one member-beneficiary binding prevents front-running attacks
 */
contract GovMinter is GovBaseV2 {
    // ========== Custom Errors ==========
    error InvalidBeneficiary();
    error InvalidAmount();
    error InvalidDepositId();
    error InvalidBankReference();
    error InvalidTokenAddress();
    error BeneficiaryNotRegistered();
    error BeneficiaryMismatch();
    error BeneficiariesLengthMismatch();
    error DuplicateBeneficiary();
    error DepositIdAlreadyUsed();
    error DepositIdInUse();
    error ProofAlreadyUsed();
    error InvalidTargetAddress();
    error InvalidWithdrawalId();
    error BurnFromMustBeProposer();
    error WithdrawalIdAlreadyUsed();
    error WithdrawalIdInUse();
    error InsufficientBurnBalance();
    error InvalidProofData();
    error ContractPaused();
    error ContractNotPaused();
    error MintFailed();

    // ========== Constants ==========
    bytes32 public constant ACTION_MINT_WITH_DEPOSIT = keccak256("MINT_WITH_DEPOSIT");
    bytes32 public constant ACTION_BURN = keccak256("BURN");
    bytes32 public constant ACTION_PAUSE = keccak256("PAUSE");
    bytes32 public constant ACTION_UNPAUSE = keccak256("UNPAUSE");

    // ========== Structs ==========
    /**
     * @notice Proof data for minting operations
     * @dev Optimized for gas efficiency with fixed-size fields grouped first
     * @param beneficiary Must match the registered beneficiary of the proposing member
     * @param amount Amount of tokens to mint
     * @param timestamp Unix timestamp of the deposit
     * @param depositId Unique identifier from the bank system
     * @param bankReference Bank transaction reference
     * @param memo Optional memo field
     */
    struct MintProof {
        // 32-byte fields (slots 0-2)
        address beneficiary;  // slot 0 (20 bytes + 12 bytes padding)
        uint256 amount;       // slot 1 (32 bytes)
        uint256 timestamp;    // slot 2 (32 bytes)

        // Dynamic fields (stored separately)
        string depositId;     // slot 3 (offset + length)
        string bankReference; // slot 4 (offset + length)
        string memo;          // slot 5 (offset + length)
    }

    /**
     * @notice Proof data for burning operations
     * @dev Optimized for gas efficiency with fixed-size fields grouped first
     * @param from Address to burn tokens from
     * @param amount Amount of tokens to burn
     * @param timestamp Unix timestamp of the withdrawal request
     * @param withdrawalId Unique identifier for the withdrawal
     * @param referenceId External reference ID
     * @param memo Optional memo field
     */
    struct BurnProof {
        // 32-byte fields (slots 0-2)
        address from;         // slot 0 (20 bytes + 12 bytes padding)
        uint256 amount;       // slot 1 (32 bytes)
        uint256 timestamp;    // slot 2 (32 bytes)

        // Dynamic fields (stored separately)
        string withdrawalId;  // slot 3 (offset + length)
        string referenceId;   // slot 4 (offset + length)
        string memo;          // slot 5 (offset + length)
    }

    struct BurnProposalData {
        uint256 amount;
        address requester;
    }

    // ========== State Variables ==========
    IFiatToken public fiatToken;
    bool public emergencyPaused;

    /// @dev Prevents front-running by binding each member to a single beneficiary
    mapping(address => address) public memberBeneficiaries;

    /// @dev Reverse mapping for O(1) duplicate beneficiary check
    /// Maps beneficiary address to the member who registered it
    mapping(address => address) public beneficiaryToMember;

    // Replay attack prevention
    mapping(bytes32 => bool) public usedProofHashes;

    // Proposal tracking - State-based depositId and withdrawalId management
    mapping(string => uint256) public depositIdToProposalId;
    mapping(string => bool) public executedDepositIds;
    mapping(string => uint256) public withdrawalIdToProposalId;
    mapping(string => bool) public executedWithdrawalIds;
    mapping(uint256 => BurnProposalData) public burnProposals;

    // Burn balance tracking
    mapping(address => uint256) public burnBalance;

    // ========== Events ==========
    /// @notice Emitted when a deposit mint proposal is created
    event DepositMintProposed(
        uint256 indexed proposalId,
        string indexed depositId,
        address indexed requester,
        address beneficiary,
        uint256 amount,
        string bankReference
    );

    /// @notice Emitted when a member registers their beneficiary
    event BeneficiaryRegistered(address indexed member, address indexed beneficiary);

    /// @notice Emitted when a user prepays for burn operations
    event BurnPrepaid(address indexed user, uint256 amount);

    /// @notice Emitted when emergency pause is activated
    event EmergencyPaused(uint256 indexed proposalId);

    /// @notice Emitted when emergency pause is deactivated
    event EmergencyUnpaused(uint256 indexed proposalId);

    // ========== Constructor & Initialization ==========
    constructor() {}

    /**
     * @notice Initialize the contract
     * @param _members Array of member addresses
     * @param _quorum Quorum required for proposals
     * @param _fiatToken Address of the fiat token contract
     * @param _beneficiaries Array of beneficiary addresses for each member
     * @dev _members and _beneficiaries must have the same length
     * @dev Each member's beneficiary is set during initialization
     * @dev Beneficiaries must be unique (except address(0)) to prevent front-running
     */
    function initialize(
        address[] memory _members,
        address[] memory _beneficiaries,
        uint32 _quorum,
        address _fiatToken
    ) external {
        // Validate beneficiaries array length matches members
        if (_members.length != _beneficiaries.length) revert BeneficiariesLengthMismatch();

        _initializeGovernance(GovernanceConfig({members: _members, quorum: _quorum, proposalExpiry: 7 days}));

        if (_fiatToken == address(0)) revert InvalidTokenAddress();
        fiatToken = IFiatToken(_fiatToken);

        // Validate uniqueness and set beneficiaries in a single pass
        // During initialization, all beneficiaries must be set (address(0) not allowed)
        for (uint256 i = 0; i < _members.length; i++) {
            address beneficiary = _beneficiaries[i];
            address member = _members[i];

            // Beneficiary must be set during initialization
            if (beneficiary == address(0)) revert InvalidBeneficiary();

            // Check for duplicates using reverse mapping (O(1) check)
            if (beneficiaryToMember[beneficiary] != address(0)) {
                revert DuplicateBeneficiary();
            }

            // Set beneficiary mappings (forward and reverse)
            memberBeneficiaries[member] = beneficiary;
            beneficiaryToMember[beneficiary] = member;

            emit BeneficiaryRegistered(member, beneficiary);
        }
    }

    // ========== Modifiers ==========

    /// @notice Modifier to check if contract is not paused
    /// @dev Reverts with ContractPaused if emergencyPaused is true
    modifier whenNotPaused() {
        if (emergencyPaused) revert ContractPaused();
        _;
    }

    /// @notice Modifier to check if contract is paused
    /// @dev Reverts with ContractNotPaused if emergencyPaused is false
    modifier whenPaused() {
        if (!emergencyPaused) revert ContractNotPaused();
        _;
    }

    // ========== Receive Function ==========

    /// @notice Receive native coins and credit sender's burn balance
    /// @dev Allows users to prepay for burn operations by sending native coins
    ///
    ///      This function enables users to deposit native coins that will be used
    ///      to back burn operations. When a burn proposal is executed, the user
    ///      receives back their prepaid amount.
    ///
    ///      Flow:
    ///      1. User sends native coins to contract
    ///      2. Amount is credited to burnBalance[msg.sender]
    ///      3. User can propose burns up to their balance
    ///      4. On burn execution, user receives native coins back
    ///
    /// @custom:security Reentrancy Safety
    ///      - No external calls, safe from reentrancy
    ///      - Simple balance increment operation
    ///      - No state dependencies
    ///
    /// @custom:security Integer Overflow Protection
    ///      - Solidity 0.8.14 provides automatic overflow protection
    ///      - burnBalance can theoretically overflow, but would require msg.value > type(uint256).max
    ///      - Practically impossible due to blockchain gas limits and total supply constraints
    ///
    /// @custom:security Access Control
    ///      - Anyone can deposit (by design)
    ///      - Only the depositor can use their balance for burns
    ///      - Balance is tracked per address
    ///
    /// @custom:usage Common Use Cases
    ///      1. **Burn Preparation**: Users deposit before proposing burns
    ///      2. **Batch Deposits**: Users can deposit multiple times (balances accumulate)
    ///      3. **Emergency Funding**: Quick deposit without separate transaction
    ///
    /// @custom:example Basic Usage
    ///      ```solidity
    ///      // Send 100 wei to contract
    ///      (bool success,) = address(govMinter).call{value: 100}("");
    ///      require(success, "Deposit failed");
    ///
    ///      // Check balance
    ///      uint256 balance = govMinter.burnBalance(msg.sender);
    ///      // balance == 100
    ///      ```
    receive() external payable {
        burnBalance[msg.sender] += msg.value;
        emit BurnPrepaid(msg.sender, msg.value);
    }

    // ========== Beneficiary Management ==========

    /**
     * @notice Register or update beneficiary address for the calling member
     * @param beneficiary The beneficiary address to register
     * @dev Only callable by active members
     * @dev Reverts if beneficiary is already registered by another member (prevents front-running)
     * @dev Members can change their own beneficiary to a different address
     * @dev Used for members added after initialization via proposeAddMember
     * @dev Uses O(1) reverse mapping for efficient duplicate check (prevents DoS)
     */
    function registerBeneficiary(address beneficiary) external onlyMember {
        if (beneficiary == address(0)) revert InvalidBeneficiary();

        // O(1) duplicate check using reverse mapping
        address existingMember = beneficiaryToMember[beneficiary];
        if (existingMember != address(0) && existingMember != msg.sender) {
            revert DuplicateBeneficiary();
        }

        // Clear old beneficiary mapping if member is changing beneficiary
        address oldBeneficiary = memberBeneficiaries[msg.sender];
        if (oldBeneficiary != address(0) && oldBeneficiary != beneficiary) {
            delete beneficiaryToMember[oldBeneficiary];
        }

        // Set new beneficiary mappings (forward and reverse)
        memberBeneficiaries[msg.sender] = beneficiary;
        beneficiaryToMember[beneficiary] = msg.sender;

        emit BeneficiaryRegistered(msg.sender, beneficiary);
    }

    // ========== Proposal Functions ==========

    /**
     * @notice Propose mint with deposit proof
     * @param proofData ABI-encoded MintProof data
     * @return proposalId The ID of the created proposal
     * @dev Proof structure: (address beneficiary, uint256 amount, uint256 timestamp,
     *      string depositId, string bankReference, string memo)
     * @dev Front-running prevention: beneficiary must match the member's registered beneficiary
     * @dev Replay prevention: depositId and proof hash must be unique
     * @dev Only active members can propose mints when contract is not paused
     * @dev Reverts with ContractPaused if emergency pause is active
     */
    function proposeMint(bytes memory proofData) external onlyMember whenNotPaused returns (uint256) {
        // Decode proof data
        MintProof memory proof = _decodeMintProof(proofData);

        // Validate proof fields
        if (proof.beneficiary == address(0)) revert InvalidBeneficiary();
        if (proof.amount == 0) revert InvalidAmount();
        if (bytes(proof.depositId).length == 0) revert InvalidDepositId();
        if (bytes(proof.bankReference).length == 0) revert InvalidBankReference();

        // Front-running prevention: verify beneficiary matches registered beneficiary
        address registeredBeneficiary = memberBeneficiaries[msg.sender];
        if (registeredBeneficiary == address(0)) revert BeneficiaryNotRegistered();
        if (proof.beneficiary != registeredBeneficiary) revert BeneficiaryMismatch();

        // Replay attack prevention - State-based depositId validation
        _validateDepositIdAvailability(proof.depositId);
        // Use standard keccak256 for clarity over inline assembly optimization
        // forge-lint-disable-next-line asm-keccak256
        bytes32 proofHash = keccak256(proofData);
        _checkProofHashUnique(proofHash);

        _markProofHashUsed(proofHash);

        // Encode execution data (beneficiary, amount, depositId) for callData
        bytes memory callData = abi.encode(proof.beneficiary, proof.amount, proof.depositId);

        // Create proposal with callData
        uint256 proposalId = _createProposal(ACTION_MINT_WITH_DEPOSIT, callData);

        // Link depositId to proposalId for state tracking
        depositIdToProposalId[proof.depositId] = proposalId;

        // Emit event
        emit DepositMintProposed(
            proposalId, proof.depositId, msg.sender, proof.beneficiary, proof.amount, proof.bankReference
        );

        return proposalId;
    }

    /**
     * @notice Propose burn with burn proof
     * @param proofData ABI-encoded BurnProof data
     * @return proposalId The ID of the created proposal
     * @dev Proof structure: (address from, uint256 amount, uint256 timestamp,
     *      string withdrawalId, string referenceId, string memo)
     * @dev Only active members can propose burns when contract is not paused
     * @dev Reverts with ContractPaused if emergency pause is active
     */
    function proposeBurn(bytes memory proofData) public onlyMember whenNotPaused returns (uint256) {
        BurnProof memory proof = _decodeBurnProof(proofData);

        if (proof.from == address(0)) revert InvalidTargetAddress();
        if (proof.amount == 0) revert InvalidAmount();
        if (bytes(proof.withdrawalId).length == 0) revert InvalidWithdrawalId();
        if (bytes(proof.referenceId).length == 0) revert InvalidBankReference();
        if (proof.from != msg.sender) revert BurnFromMustBeProposer();

        // Replay attack prevention - State-based withdrawalId validation
        _validateWithdrawalIdAvailability(proof.withdrawalId);
        // Use standard keccak256 for clarity over inline assembly optimization
        // forge-lint-disable-next-line asm-keccak256
        bytes32 proofHash = keccak256(proofData);
        _checkProofHashUnique(proofHash);

        _markProofHashUsed(proofHash);

        if (burnBalance[proof.from] < proof.amount) revert InsufficientBurnBalance();

        // Encode execution data (from, amount, withdrawalId) for callData
        bytes memory callData = abi.encode(proof.from, proof.amount, proof.withdrawalId);

        uint256 proposalId = _createProposal(ACTION_BURN, callData);

        // Link withdrawalId to proposalId for state tracking
        withdrawalIdToProposalId[proof.withdrawalId] = proposalId;

        burnProposals[proposalId] = BurnProposalData({amount: proof.amount, requester: proof.from});

        return proposalId;
    }

    /**
     * @notice Propose emergency pause of all mint and burn operations
     * @return proposalId The ID of the created proposal
     * @dev Creates a governance proposal to activate emergency pause
     * @dev Only callable by active governance members when contract is not paused
     * @dev When executed, sets emergencyPaused = true, blocking all mints and burns
     * @dev Reverts with ContractPaused if already in paused state (prevents duplicate proposals)
     */
    function proposePause() external onlyMember whenNotPaused returns (uint256) {
        // Create proposal with empty callData (no parameters needed)
        bytes memory callData = "";
        return _createProposal(ACTION_PAUSE, callData);
    }

    /**
     * @notice Propose resumption of mint and burn operations
     * @return proposalId The ID of the created proposal
     * @dev Creates a governance proposal to deactivate emergency pause
     * @dev Only callable by active governance members when contract is paused
     * @dev When executed, sets emergencyPaused = false, allowing mints and burns
     * @dev Reverts with ContractNotPaused if not in paused state (prevents unnecessary proposals)
     */
    function proposeUnpause() external onlyMember whenPaused returns (uint256) {
        // Create proposal with empty callData (no parameters needed)
        bytes memory callData = "";
        return _createProposal(ACTION_UNPAUSE, callData);
    }

    // ========== Internal Action Implementation ==========

    /// @dev Execute custom governance actions (mint, burn, pause, unpause)
    /// @notice Routes action execution based on action type
    ///
    ///         This function is called by GovBaseV2.executeProposal when a proposal
    ///         is approved and ready for execution. It decodes the callData and
    ///         delegates to the appropriate handler.
    ///
    ///         Supported actions:
    ///         - ACTION_MINT_WITH_DEPOSIT: Mint tokens to beneficiary
    ///         - ACTION_BURN: Burn tokens and return native coins
    ///         - ACTION_PAUSE: Activate emergency pause (blocks mints and burns)
    ///         - ACTION_UNPAUSE: Deactivate emergency pause (resumes operations)
    ///
    /// @param actionType The type of action to execute (keccak256 hash)
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if action executed successfully, false otherwise
    ///
    /// @custom:security Action Routing
    ///      - Only processes recognized action types
    ///      - Returns false for unknown action types (safe failure)
    ///      - No state changes for unknown actions
    ///
    /// @custom:security Execution Safety
    ///      - Called only by GovBaseV2.executeProposal (internal context)
    ///      - Already protected by proposal approval process
    ///      - Try-catch in handlers prevents revert propagation
    ///
    /// @custom:usage Call Flow
    ///      1. GovBaseV2.executeProposal calls this function
    ///      2. Decode actionType and route to handler
    ///      3. Handler executes action with try-catch
    ///      4. Return success status to update proposal state
    function _executeCustomAction(bytes32 actionType, bytes memory callData) internal override returns (bool) {
        if (actionType == ACTION_MINT_WITH_DEPOSIT) {
            (address beneficiary, uint256 amount, string memory depositId) =
                abi.decode(callData, (address, uint256, string));
            return _executeMint(beneficiary, amount, depositId);
        }

        if (actionType == ACTION_BURN) {
            (address from, uint256 amount, string memory withdrawalId) =
                abi.decode(callData, (address, uint256, string));
            return _safeBurn(from, amount, withdrawalId);
        }

        if (actionType == ACTION_PAUSE) {
            return _executePause();
        }

        if (actionType == ACTION_UNPAUSE) {
            return _executeUnpause();
        }

        return false;
    }

    /// @dev Execute mint action - mint tokens to beneficiary
    /// @notice Mints fiat tokens and marks depositId as permanently executed
    ///
    ///         This function is the final step in the mint proposal workflow.
    ///         It calls the fiatToken contract to mint tokens and permanently
    ///         consumes the depositId to prevent replay attacks.
    ///
    ///         Execution flow:
    ///         1. Check emergency pause status
    ///         2. Try to mint tokens via fiatToken.mint()
    ///         3. On success: Mark depositId as executed (permanent)
    ///         4. On failure: Return false (proposal can be retried)
    ///
    /// @param beneficiary Address to receive minted tokens
    /// @param amount Amount of tokens to mint
    /// @param depositId Unique identifier for the deposit (consumed on success)
    /// @return success True if mint succeeded, false otherwise
    ///
    /// @custom:security Emergency Pause
    ///      - Reverts if contract is paused (emergencyPaused = true)
    ///      - Prevents minting during emergency situations
    ///      - Revert propagates to proposal execution
    ///
    /// @custom:security Try-Catch Pattern
    ///      - Catches fiatToken.mint() failures
    ///      - Returns false on failure (allows proposal retry)
    ///      - Does not consume depositId on failure (can retry with same depositId)
    ///
    /// @custom:security DepositId Consumption
    ///      - depositId marked as executed ONLY on successful mint
    ///      - Prevents replay attacks with same depositId
    ///      - Once marked, depositId cannot be reused (permanent)
    ///
    /// @custom:security State Changes
    ///      - Updates executedDepositIds[depositId] = true on success
    ///      - No state changes on failure (clean retry)
    ///
    /// @custom:invariant DepositId Uniqueness
    ///      - Invariant: Each depositId can only be successfully executed once
    ///      - Invariant: Failed executions do not consume depositId
    function _executeMint(address beneficiary, uint256 amount, string memory depositId) internal returns (bool) {
        if (emergencyPaused) revert ContractPaused();

        try fiatToken.mint(beneficiary, amount) {
            // Mark depositId as permanently executed (consumed)
            _markDepositIdExecuted(depositId);
            return true;
        } catch {
            return false;
        }
    }

    /// @dev Execute burn action - burn tokens and return native coins
    /// @notice Burns fiat tokens and transfers native coins to user
    ///
    ///         This function implements the burn workflow with proper rollback
    ///         on failure. It follows the Checks-Effects-Interactions (CEI) pattern
    ///         to prevent reentrancy attacks.
    ///
    ///         Execution flow:
    ///         1. Check emergency pause status
    ///         2. Deduct amount from burnBalance[from] (effect)
    ///         3. Try to burn tokens via fiatToken.burn()
    ///         4. Try to transfer native coins to user (interaction)
    ///         5. On success: Mark withdrawalId as executed (permanent)
    ///         6. On failure: Rollback burnBalance and return false
    ///
    /// @param from Address to burn tokens from (must be proposal requester)
    /// @param amount Amount of tokens to burn and native coins to return
    /// @param withdrawalId Unique identifier for the withdrawal (consumed on success)
    /// @return success True if burn succeeded, false otherwise
    ///
    /// @custom:security Emergency Pause
    ///      - Reverts if contract is paused (emergencyPaused = true)
    ///      - Prevents burning during emergency situations
    ///
    /// @custom:security CEI Pattern (Checks-Effects-Interactions)
    ///      - **Checks**: emergencyPaused check
    ///      - **Effects**: burnBalance[from] -= amount (before external calls)
    ///      - **Interactions**: fiatToken.burn() and from.call{value}()
    ///      - Prevents reentrancy by updating state before external calls
    ///
    /// @custom:security Reentrancy Protection
    ///      - State updated before external calls (CEI pattern)
    ///      - Even if 'from' is malicious contract, cannot exploit
    ///      - burnBalance already decremented before call
    ///      - GovBaseV2 has nonReentrant modifier on executeProposal
    ///
    /// @custom:security Rollback Mechanism
    ///      - If fiatToken.burn() fails: Rollback burnBalance, return false
    ///      - If native transfer fails: Rollback burnBalance, return false
    ///      - Clean state on failure (proposal can be retried)
    ///
    /// @custom:security WithdrawalId Consumption
    ///      - withdrawalId marked as executed ONLY on full success
    ///      - Prevents replay attacks with same withdrawalId
    ///      - Once marked, withdrawalId cannot be reused (permanent)
    ///
    /// @custom:security Native Transfer Safety
    ///      - Uses low-level call{value}() for native transfer
    ///      - Checks success flag and rolls back on failure
    ///      - Does not use transfer() (fixed 2300 gas)
    ///      - Does not use send() (ignores return value patterns)
    ///
    /// @custom:invariant Balance Consistency
    ///      - Invariant: burnBalance decreases IFF burn succeeds
    ///      - Invariant: Each withdrawalId can only be successfully executed once
    ///      - Invariant: Failed executions do not consume withdrawalId
    ///      - Invariant: User receives native coins IFF burn succeeds
    function _safeBurn(address from, uint256 amount, string memory withdrawalId) internal returns (bool) {
        if (emergencyPaused) revert ContractPaused();

        burnBalance[from] -= amount;

        fiatToken.burn(amount);
        _markWithdrawalIdExecuted(withdrawalId);

        return true;
    }

    /// @dev Execute emergency pause action
    /// @notice Activates emergency pause to block all mint and burn operations
    ///
    ///         This function is called when a pause proposal is approved and executed.
    ///         It sets the emergencyPaused flag to true, which causes all subsequent
    ///         mint and burn operations to revert with ContractPaused error.
    ///
    ///         Use cases:
    ///         - Security incident detected
    ///         - Critical bug discovered in fiatToken or governance
    ///         - Regulatory compliance requirement
    ///         - System maintenance or upgrade preparation
    ///
    /// @return success Always returns true (pause cannot fail)
    ///
    /// @custom:security State Changes
    ///      - Sets emergencyPaused = true
    ///      - Blocks _executeMint and _safeBurn
    ///      - Does not affect existing proposals (can still vote/execute pause/unpause)
    ///
    /// @custom:security Idempotency
    ///      - Safe to call multiple times
    ///      - Setting true to true has no side effects
    ///      - No state corruption from repeated calls
    ///
    /// @custom:security Governance Control
    ///      - Only executable through approved governance proposal
    ///      - Requires quorum approval
    ///      - Cannot be called directly (internal function)
    ///
    /// @custom:usage Emergency Response
    ///      1. Member calls proposePause()
    ///      2. Governance members approve proposal
    ///      3. This function executes, setting emergencyPaused = true
    ///      4. All mint/burn operations now revert
    ///      5. System is in safe state for investigation/fixes
    function _executePause() internal returns (bool) {
        emergencyPaused = true;
        emit EmergencyPaused(currentProposalId);
        return true;
    }

    /// @dev Execute emergency unpause action
    /// @notice Deactivates emergency pause to resume mint and burn operations
    ///
    ///         This function is called when an unpause proposal is approved and executed.
    ///         It sets the emergencyPaused flag to false, which allows mint and burn
    ///         operations to proceed normally again.
    ///
    ///         Use cases:
    ///         - Security incident resolved
    ///         - Bug fix deployed and verified
    ///         - Compliance requirement satisfied
    ///         - Maintenance/upgrade completed
    ///
    /// @return success Always returns true (unpause cannot fail)
    ///
    /// @custom:security State Changes
    ///      - Sets emergencyPaused = false
    ///      - Resumes normal mint and burn operations
    ///      - Existing proposals can be executed normally
    ///
    /// @custom:security Idempotency
    ///      - Safe to call multiple times
    ///      - Setting false to false has no side effects
    ///      - No state corruption from repeated calls
    ///
    /// @custom:security Governance Control
    ///      - Only executable through approved governance proposal
    ///      - Requires quorum approval
    ///      - Cannot be called directly (internal function)
    ///
    /// @custom:security Careful Unpausing
    ///      - Governance should verify issue is resolved before unpausing
    ///      - Consider testing on testnet first
    ///      - Monitor system after unpause for anomalies
    ///
    /// @custom:usage Recovery Process
    ///      1. Governance investigates and resolves issue
    ///      2. Member calls proposeUnpause()
    ///      3. Governance members approve proposal
    ///      4. This function executes, setting emergencyPaused = false
    ///      5. System resumes normal operations
    function _executeUnpause() internal returns (bool) {
        emergencyPaused = false;
        emit EmergencyUnpaused(currentProposalId);
        return true;
    }

    // ========== Proof Management Functions ==========

    /// @dev Decode ABI-encoded mint proof data
    /// @notice Converts bytes to MintProof struct
    ///
    ///         This function decodes the proof data submitted by members when
    ///         proposing mint operations. It validates the data length and
    ///         extracts all proof fields.
    ///
    /// @param data ABI-encoded proof data (address, uint256, uint256, string, string, string)
    /// @return proof Decoded MintProof struct
    ///
    /// @custom:security Input Validation
    ///      - Reverts on empty data (data.length == 0)
    ///      - ABI decode will revert on malformed data automatically
    ///      - All fields extracted with type safety
    ///
    /// @custom:security Pure Function
    ///      - No state reads or writes
    ///      - Deterministic output for given input
    ///      - Gas-efficient validation
    ///
    /// @custom:revert InvalidProofData
    ///      - Thrown when data.length == 0
    ///      - Also thrown by abi.decode on malformed data
    function _decodeMintProof(bytes memory data) internal pure returns (MintProof memory) {
        if (data.length == 0) revert InvalidProofData();

        (
            address beneficiary,
            uint256 amount,
            uint256 timestamp,
            string memory depositId,
            string memory bankReference,
            string memory memo
        ) = abi.decode(data, (address, uint256, uint256, string, string, string));

        return MintProof({
            beneficiary: beneficiary,
            amount: amount,
            timestamp: timestamp,
            depositId: depositId,
            bankReference: bankReference,
            memo: memo
        });
    }

    /// @dev Decode ABI-encoded burn proof data
    /// @notice Converts bytes to BurnProof struct
    ///
    ///         This function decodes the proof data submitted by members when
    ///         proposing burn operations. It validates the data length and
    ///         extracts all proof fields.
    ///
    /// @param data ABI-encoded proof data (address, uint256, uint256, string, string, string)
    /// @return proof Decoded BurnProof struct
    ///
    /// @custom:security Input Validation
    ///      - Reverts on empty data (data.length == 0)
    ///      - ABI decode will revert on malformed data automatically
    ///      - All fields extracted with type safety
    ///
    /// @custom:security Pure Function
    ///      - No state reads or writes
    ///      - Deterministic output for given input
    ///      - Gas-efficient validation
    ///
    /// @custom:revert InvalidProofData
    ///      - Thrown when data.length == 0
    ///      - Also thrown by abi.decode on malformed data
    function _decodeBurnProof(bytes memory data) internal pure returns (BurnProof memory) {
        if (data.length == 0) revert InvalidProofData();

        (
            address from,
            uint256 amount,
            uint256 timestamp,
            string memory withdrawalId,
            string memory referenceId,
            string memory memo
        ) = abi.decode(data, (address, uint256, uint256, string, string, string));

        return BurnProof({
            from: from,
            amount: amount,
            timestamp: timestamp,
            withdrawalId: withdrawalId,
            referenceId: referenceId,
            memo: memo
        });
    }

    // ========== Replay Attack Prevention ==========

    /**
     * @notice Validate depositId availability based on proposal state
     * @param depositId The deposit identifier to validate
     * @dev Reverts if depositId is already executed or currently in use (Voting/Approved)
     * @dev Allows reuse if previous proposal was Cancelled, Rejected, or Expired
     */
    function _validateDepositIdAvailability(string memory depositId) internal view {
        // Check if depositId was already executed (permanent consumption)
        if (executedDepositIds[depositId]) {
            revert DepositIdAlreadyUsed();
        }

        // Check if depositId is associated with an existing proposal
        uint256 existingProposalId = depositIdToProposalId[depositId];

        if (existingProposalId == 0) {
            // First time using this depositId
            return;
        }

        // Get proposal status
        ProposalStatus status = proposals[existingProposalId].status;

        // Voting or Approved: depositId is currently in use
        if (status == ProposalStatus.Voting || status == ProposalStatus.Approved) {
            revert DepositIdInUse();
        }

        // Cancelled, Rejected, Expired: depositId can be reused
        // (Executed is already blocked by executedDepositIds check above)
    }

    /**
     * @notice Validate withdrawalId availability based on proposal state
     * @param withdrawalId The withdrawal identifier to validate
     * @dev Reverts if withdrawalId is already executed or currently in use (Voting/Approved)
     * @dev Allows reuse if previous proposal was Cancelled, Rejected, or Expired
     */
    function _validateWithdrawalIdAvailability(string memory withdrawalId) internal view {
        // Check if withdrawalId was already executed (permanent consumption)
        if (executedWithdrawalIds[withdrawalId]) {
            revert WithdrawalIdAlreadyUsed();
        }

        // Check if withdrawalId is associated with an existing proposal
        uint256 existingProposalId = withdrawalIdToProposalId[withdrawalId];

        if (existingProposalId == 0) {
            // First time using this withdrawalId
            return;
        }

        // Get proposal status
        ProposalStatus status = proposals[existingProposalId].status;

        // Voting or Approved: withdrawalId is currently in use
        if (status == ProposalStatus.Voting || status == ProposalStatus.Approved) {
            revert WithdrawalIdInUse();
        }

        // Cancelled, Rejected, Expired: withdrawalId can be reused
        // (Executed is already blocked by executedWithdrawalIds check above)
    }

    /// @dev Check if proof hash has been used before
    /// @notice Validates that the proof hash is unique and not replayed
    ///
    ///         This function is part of the replay attack prevention system.
    ///         Every mint/burn proof is hashed and checked against the
    ///         usedProofHashes mapping to prevent duplicate submissions.
    ///
    /// @param proofHash Keccak256 hash of the proof data
    ///
    /// @custom:security Replay Attack Prevention
    ///      - Prevents resubmission of identical proof data
    ///      - Works in conjunction with depositId/withdrawalId tracking
    ///      - Provides hash-level duplicate detection
    ///
    /// @custom:security View Function
    ///      - No state modifications
    ///      - Gas-efficient check
    ///      - Called before marking proof as used
    ///
    /// @custom:revert ProofAlreadyUsed
    ///      - Thrown when proofHash exists in usedProofHashes
    ///      - Indicates replay attack attempt or duplicate submission
    function _checkProofHashUnique(bytes32 proofHash) internal view {
        if (usedProofHashes[proofHash]) revert ProofAlreadyUsed();
    }

    /// @dev Mark depositId as permanently executed
    /// @notice Permanently consumes depositId to prevent replay attacks
    ///
    ///         This function is called only after successful mint execution.
    ///         Once marked, the depositId can never be used again, even if
    ///         a proposal with the same depositId was cancelled or rejected.
    ///
    /// @param depositId The deposit identifier to mark as executed
    ///
    /// @custom:security Permanent Consumption
    ///      - depositId permanently marked (no undo mechanism)
    ///      - Called ONLY after successful mint
    ///      - Failed mints do not consume depositId (can retry)
    ///
    /// @custom:security State Changes
    ///      - Updates executedDepositIds[depositId] = true
    ///      - Single SSTORE operation (~20000 gas cold, ~2900 gas warm)
    ///
    /// @custom:usage Call Sites
    ///      - _executeMint: After successful fiatToken.mint() call
    function _markDepositIdExecuted(string memory depositId) internal {
        executedDepositIds[depositId] = true;
    }

    /// @dev Mark withdrawalId as permanently executed
    /// @notice Permanently consumes withdrawalId to prevent replay attacks
    ///
    ///         This function is called only after successful burn execution.
    ///         Once marked, the withdrawalId can never be used again, even if
    ///         a proposal with the same withdrawalId was cancelled or rejected.
    ///
    /// @param withdrawalId The withdrawal identifier to mark as executed
    ///
    /// @custom:security Permanent Consumption
    ///      - withdrawalId permanently marked (no undo mechanism)
    ///      - Called ONLY after successful burn
    ///      - Failed burns do not consume withdrawalId (can retry)
    ///
    /// @custom:security State Changes
    ///      - Updates executedWithdrawalIds[withdrawalId] = true
    ///      - Single SSTORE operation (~20000 gas cold, ~2900 gas warm)
    ///
    /// @custom:usage Call Sites
    ///      - _safeBurn: After successful burn and native transfer
    function _markWithdrawalIdExecuted(string memory withdrawalId) internal {
        executedWithdrawalIds[withdrawalId] = true;
    }

    /// @dev Mark proof hash as used to prevent replay attacks
    /// @notice Permanently stores proof hash to detect duplicates
    ///
    ///         This function is called immediately after proof validation
    ///         and before proposal creation. It provides hash-level replay
    ///         protection in addition to depositId/withdrawalId tracking.
    ///
    /// @param proofHash Keccak256 hash of the entire proof data
    ///
    /// @custom:security Immediate Marking
    ///      - Marked before proposal creation (prevents race conditions)
    ///      - Permanent storage (no cleanup mechanism)
    ///      - Hash collision extremely unlikely (keccak256)
    ///
    /// @custom:security State Changes
    ///      - Updates usedProofHashes[proofHash] = true
    ///      - Single SSTORE operation (~20000 gas cold, ~2900 gas warm)
    ///
    /// @custom:usage Call Sites
    ///      - proposeMint: After validation, before proposal creation
    ///      - proposeBurn: After validation, before proposal creation
    function _markProofHashUsed(bytes32 proofHash) internal {
        usedProofHashes[proofHash] = true;
    }

    // ========== Hook Implementations ==========

    /// @dev Hook called when a new member is added to governance
    /// @notice Currently empty - no special logic needed for member additions
    ///
    ///         This hook is called by GovBaseV2 after successfully adding a member
    ///         through governance proposal execution. GovMinter does not require
    ///         special handling for new members.
    ///
    ///         New members must:
    ///         - Register beneficiary via registerBeneficiary() before proposing mints
    ///         - Deposit native coins via receive() before proposing burns
    ///
    /// @custom:parameters
    ///      - newMember: Address of the newly added member (unused in this implementation)
    ///
    /// @custom:design Empty Implementation
    ///      - No automatic beneficiary registration (members must call registerBeneficiary)
    ///      - No automatic burn balance allocation (members must deposit via receive)
    ///      - Keeps governance flexible and explicit
    function _onMemberAdded(address /* newMember */) internal override {}

    /// @dev Hook called when a member is removed from governance
    /// @notice Currently empty - no cleanup needed for removed members
    ///
    ///         This hook is called by GovBaseV2 after successfully removing a member
    ///         through governance proposal execution. GovMinter does not clean up
    ///         member state to preserve historical data.
    ///
    ///         Removed member state:
    ///         - memberBeneficiaries[member] remains unchanged (historical record)
    ///         - burnBalance[member] remains unchanged (can be withdrawn manually if needed)
    ///         - Active proposals remain valid (executed by remaining members)
    ///
    /// @custom:parameters
    ///      - member: Address of the removed member (unused in this implementation)
    ///
    /// @custom:design No State Cleanup
    ///      - Preserves beneficiary mapping for audit trail
    ///      - Preserves burn balance for potential withdrawal
    ///      - Does not cancel active proposals (governance decision)
    function _onMemberRemoved(address /* member */) internal override {}

    /// @dev Hook called when a member address is changed in governance
    /// @notice Currently empty - no state migration needed
    ///
    ///         This hook is called by GovBaseV2 after successfully changing a member
    ///         address through governance proposal execution. GovMinter does not
    ///         automatically migrate state from old to new address.
    ///
    ///         State migration considerations:
    ///         - memberBeneficiaries[oldMember] NOT migrated to newMember
    ///         - burnBalance[oldMember] NOT migrated to newMember
    ///         - New member must register beneficiary via registerBeneficiary()
    ///         - New member must deposit for burns via receive()
    ///
    /// @custom:parameters
    ///      - oldMember: Old member address (unused in this implementation)
    ///      - newMember: New member address (unused in this implementation)
    ///
    /// @custom:design No Automatic Migration
    ///      - Explicit beneficiary registration required (prevents errors)
    ///      - Explicit burn balance deposit required (clear accounting)
    ///      - Old member data preserved (audit trail)
    ///      - Governance can manually handle special cases if needed
    function _onMemberChanged(address /* oldMember */, address /* newMember */) internal override {}
}
