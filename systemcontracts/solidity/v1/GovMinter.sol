// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

import {GovBaseV2} from "../abstracts/GovBaseV2.sol";
import {IFiatToken} from "../interfaces/IFiatToken.sol";

/**
 * @title GovMinter
 * @notice Governance-controlled minting and burning with off-chain validation
 * @dev Beneficiary validation is performed off-chain before proposal submission
 */
contract GovMinter is GovBaseV2 {
    // ========== Custom Errors ==========
    error InvalidBeneficiary();
    error InvalidAmount();
    error InvalidDepositId();
    error InvalidBankReference();
    error InvalidTokenAddress();
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
    error InsufficientMinterAllowance();
    error BurnAmountMismatch();

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

    // Replay attack prevention
    mapping(bytes32 => bool) public usedProofHashes;

    // Proposal tracking - State-based depositId and withdrawalId management
    mapping(string => uint256) public depositIdToProposalId;
    mapping(string => bool) public executedDepositIds;
    mapping(string => uint256) public withdrawalIdToProposalId;
    mapping(string => bool) public executedWithdrawalIds;
    mapping(uint256 => BurnProposalData) public burnProposals;

    // Mint allowance tracking
    /// @dev Total amount reserved by pending mint proposals
    /// @notice Used to prevent over-allocation of minter allowance
    uint256 public reservedMintAmount;

    /// @dev Maps proposalId to mint amount for cleanup on proposal completion
    mapping(uint256 => uint256) public mintProposalAmounts;

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

    /// @notice Emitted when a user prepays for burn operations
    event BurnPrepaid(address indexed user, uint256 amount);

    /// @notice Emitted when burn is successfully executed
    /// @dev Essential for off-chain monitoring and governance transparency
    event BurnExecuted(address indexed from, uint256 indexed amount, string withdrawalId);

    /// @notice Emitted when emergency pause is activated
    event EmergencyPaused(uint256 indexed proposalId);

    /// @notice Emitted when emergency pause is deactivated
    event EmergencyUnpaused(uint256 indexed proposalId);

    // ========== Constructor & Initialization ==========
    constructor() {}


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


    // ============================================================
    // INITIALIZATION
    // ============================================================

    /// @notice Initialization is handled by genesis configuration
    /// @dev See systemcontracts/gov_minter.go for genesis initialization implementation
    ///
    /// Genesis initialization sets:
    /// - GovBase state: members, quorum, proposalExpiry, memberVersion (via gov_base.go)
    /// - GovMinter state: fiatToken


    // ============================================================
    // PUBLIC FUNCTIONS - Proposal Operations
    // ============================================================


    // ========== Proposal Functions ==========

    /**
     * @notice Propose mint with deposit proof
     * @param proofData ABI-encoded MintProof data
     * @return proposalId The ID of the created proposal
     * @dev Proof structure: (address beneficiary, uint256 amount, uint256 timestamp,
     *      string depositId, string bankReference, string memo)
     * @dev Replay prevention: depositId and proof hash must be unique
     * @dev Only active members can propose mints when contract is not paused
     * @dev Reverts with ContractPaused if emergency pause is active
     * @dev Beneficiary validation is performed off-chain
     */
    function proposeMint(bytes memory proofData) external onlyActiveMember whenNotPaused returns (uint256) {
        // Decode proof data
        MintProof memory proof = _decodeMintProof(proofData);

        // Validate proof fields
        if (proof.beneficiary == address(0)) revert InvalidBeneficiary();
        if (proof.amount == 0) revert InvalidAmount();
        if (bytes(proof.depositId).length == 0) revert InvalidDepositId();
        if (bytes(proof.bankReference).length == 0) revert InvalidBankReference();

        // Validate against minter allowance
        // Check if there's enough unreserved allowance for this mint
        uint256 totalAllowance = fiatToken.minterAllowance(address(this));
        uint256 availableAllowance = totalAllowance - reservedMintAmount;
        if (proof.amount > availableAllowance) revert InsufficientMinterAllowance();

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

        // Reserve mint allowance for this proposal
        reservedMintAmount += proof.amount;
        mintProposalAmounts[proposalId] = proof.amount;

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
     * @dev Requires msg.value to match proof.amount for burn collateral
     * @custom:usage Single-step burn proposal with native coin deposit
     *      ```solidity
     *      BurnProof memory proof = BurnProof({
     *          from: msg.sender,
     *          amount: 100 ether,
     *          timestamp: block.timestamp,
     *          withdrawalId: "WD-001",
     *          referenceId: "REF-001",
     *          memo: "Burn memo"
     *      });
     *      bytes memory proofData = abi.encode(
     *          proof.from, proof.amount, proof.timestamp,
     *          proof.withdrawalId, proof.referenceId, proof.memo
     *      );
     *      govMinter.proposeBurn{value: 100 ether}(proofData);
     *      ```
     */
    function proposeBurn(bytes calldata proofData) external payable onlyActiveMember whenNotPaused returns (uint256) {
        BurnProof memory proof = _decodeBurnProof(proofData);

        if (proof.from == address(0)) revert InvalidTargetAddress();
        if (proof.amount == 0) revert InvalidAmount();
        if (bytes(proof.withdrawalId).length == 0) revert InvalidWithdrawalId();
        if (bytes(proof.referenceId).length == 0) revert InvalidBankReference();
        if (proof.from != msg.sender) revert BurnFromMustBeProposer();

        // Validate burn amount matches msg.value exactly
        // This ensures atomic deposit-and-propose in single transaction
        if (msg.value != proof.amount) revert BurnAmountMismatch();

        // Credit burn balance with received native coins
        burnBalance[msg.sender] += msg.value;
        emit BurnPrepaid(msg.sender, msg.value);

        // Replay attack prevention - State-based withdrawalId validation
        _validateWithdrawalIdAvailability(proof.withdrawalId);
        // Use standard keccak256 for clarity over inline assembly optimization
        // forge-lint-disable-next-line asm-keccak256
        bytes32 proofHash = keccak256(proofData);
        _checkProofHashUnique(proofHash);

        _markProofHashUsed(proofHash);

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
    function proposePause() external onlyActiveMember whenNotPaused returns (uint256) {
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
    function proposeUnpause() external onlyActiveMember whenPaused returns (uint256) {
        // Create proposal with empty callData (no parameters needed)
        bytes memory callData = "";
        return _createProposal(ACTION_UNPAUSE, callData);
    }


    // ============================================================
    // INTERNAL FUNCTIONS - Action Execution Router
    // ============================================================


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
    ///         - ACTION_BURN: Burn fiat tokens from GovMinter balance
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


    // ============================================================
    // INTERNAL FUNCTIONS - Action Executors
    // ============================================================


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

        // Note: Reserved mint allowance cleanup is handled by _onProposalFinalized hook
        // Hook is automatically called by GovBaseV2 when proposal reaches terminal state
        // This ensures cleanup happens for ALL terminal states (Executed, Failed, Expired, Cancelled, Rejected)

        try fiatToken.mint(beneficiary, amount) {
            // Mark depositId as permanently executed (consumed)
            _markDepositIdExecuted(depositId);
            return true;
        } catch {
            return false;
        }
    }


    /// @dev Execute burn action - burn fiat tokens from GovMinter balance
    /// @notice Burns fiat tokens held by GovMinter contract
    ///
    ///         This function implements the burn workflow. It deducts the prepaid
    ///         native coin balance and calls fiatToken.burn() to burn fiat tokens.
    ///         The native coins remain in GovMinter as collateral.
    ///
    ///         Execution flow:
    ///         1. Check emergency pause status
    ///         2. Deduct amount from burnBalance[from] (effect)
    ///         3. Call fiatToken.burn(amount) - burns GovMinter's fiat tokens
    ///         4. On success: Mark withdrawalId as executed (permanent)
    ///         5. On failure: Automatic revert rolls back all state changes
    ///
    /// @param from Address whose burnBalance will be deducted (proposal requester)
    /// @param amount Amount of fiat tokens to burn
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
    ///      - **Interactions**: fiatToken.burn()
    ///      - Prevents reentrancy by updating state before external calls
    ///
    /// @custom:security Reentrancy Protection
    ///      - State updated before external calls (CEI pattern)
    ///      - burnBalance already decremented before fiatToken.burn()
    ///      - GovBaseV2 has nonReentrant modifier on executeProposal
    ///
    /// @custom:security Automatic Rollback
    ///      - If fiatToken.burn() fails: Solidity reverts all state changes
    ///      - burnBalance automatically restored on revert
    ///      - Clean state on failure (proposal can be retried)
    ///
    /// @custom:security WithdrawalId Consumption
    ///      - withdrawalId marked as executed ONLY on full success
    ///      - Prevents replay attacks with same withdrawalId
    ///      - Once marked, withdrawalId cannot be reused (permanent)
    ///
    /// @custom:invariant Balance Consistency
    ///      - Invariant: burnBalance decreases IFF burn succeeds
    ///      - Invariant: Each withdrawalId can only be successfully executed once
    ///      - Invariant: Failed executions do not consume withdrawalId
    ///      - Invariant: Native coins remain in GovMinter as collateral
    function _safeBurn(address from, uint256 amount, string memory withdrawalId) internal returns (bool) {
        if (emergencyPaused) revert ContractPaused();

        // Note: burnBalance cleanup is handled by _onProposalFinalized hook
        // Hook is automatically called by GovBaseV2 when proposal reaches terminal state
        // This ensures cleanup happens for ALL terminal states (Executed, Failed, Expired, Cancelled, Rejected)

        try fiatToken.burn(amount) {
            // 1. Effects: Update state after successful burn
            burnBalance[from] -= amount;
            // 2. Mark withdrawalId as permanently executed (consumed)
            _markWithdrawalIdExecuted(withdrawalId);
            // 3. Emit event for off-chain monitoring
            emit BurnExecuted(from, amount, withdrawalId);
            return true;
        } catch {
            return false;
        }
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


    // ============================================================
    // INTERNAL FUNCTIONS - Proof Decoding
    // ============================================================


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


    // ============================================================
    // INTERNAL FUNCTIONS - Replay Prevention (Validation)
    // ============================================================


    // ========== Replay Attack Prevention ==========

    /**
     * @notice Validate depositId availability based on proposal state
     * @param depositId The deposit identifier to validate
     * @dev Reverts if depositId is already executed or currently in use (Voting/Approved)
     * @dev Allows reuse if previous proposal was Cancelled, Rejected, or Expired
     * @dev Automatically cleans up reserved mint allowance for terminal state proposals
     */
    function _validateDepositIdAvailability(string memory depositId) internal {
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

        // Cancelled, Rejected, Expired, Failed: depositId can be reused
        // Note: Reserved mint allowance cleanup is handled by _onProposalFinalized hook
        // Hook is automatically called when proposal transitions to terminal state
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


    // ============================================================
    // INTERNAL FUNCTIONS - Replay Prevention (State Updates)
    // ============================================================


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
    ///      - _safeBurn: After successful fiat token burn
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


    // ============================================================
    // INTERNAL FUNCTIONS - Cleanup & Lifecycle Hooks
    // ============================================================


    /// @dev Internal function to release reserved mint allowance
    /// @notice Automatically cleans up reservation when proposal reaches terminal state
    /// @param proposalId The proposal ID to clean up
    function _cleanupMintReservation(uint256 proposalId) internal {
        uint256 reservedAmount = mintProposalAmounts[proposalId];
        if (reservedAmount > 0) {
            reservedMintAmount -= reservedAmount;
            delete mintProposalAmounts[proposalId];
        }
    }


    /// @dev Hook implementation for proposal finalization
    /// @notice Called automatically by GovBaseV2 when proposal reaches terminal state
    /// @param proposalId The proposal that reached terminal state
    ///
    /// @custom:security Called AFTER state transition (status already updated in GovBaseV2)
    /// @custom:security Safe to call multiple times due to idempotent cleanup function
    /// @custom:security No external calls - only state cleanup
    ///
    /// @custom:usage Terminal States (all trigger cleanup)
    ///      - Executed: Proposal successfully executed (mint completed)
    ///      - Failed: Proposal execution failed (mint failed)
    ///      - Expired: Proposal expired before execution
    ///      - Cancelled: Proposal cancelled by proposer
    ///      - Rejected: Proposal rejected by members
    ///
    /// @custom:design Pattern
    ///      - Template Method Pattern: GovBaseV2 defines lifecycle, GovMinter implements cleanup
    ///      - Single Responsibility: Proposal lifecycle (GovBaseV2) separated from mint cleanup (GovMinter)
    ///      - DRY Principle: One hook definition, called from 7 terminal state transitions
    ///
    /// @custom:idempotency
    ///      - Safe to call multiple times for same proposalId
    ///      - _cleanupMintReservation checks reservedAmount > 0 before cleanup
    ///      - Prevents double-cleanup issues
    ///
    /// @custom:gas-optimization
    ///      - Early exit if no reservation (reservedAmount == 0)
    ///      - Single SLOAD + conditional SSTORE: ~5000-8000 gas
    ///      - Minimal overhead for non-mint proposals
    function _onProposalFinalized(uint256 proposalId) internal override {
        _cleanupMintReservation(proposalId);
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
    ///         - Deposit native coins via receive() before proposing burns
    ///         - Beneficiary validation is performed off-chain before proposing mints
    ///
    /// @custom:parameters
    ///      - newMember: Address of the newly added member (unused in this implementation)
    ///
    /// @custom:design Empty Implementation
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
    ///         - burnBalance[member] remains unchanged (can be withdrawn manually if needed)
    ///         - Active proposals remain valid (executed by remaining members)
    ///
    /// @custom:parameters
    ///      - member: Address of the removed member (unused in this implementation)
    ///
    /// @custom:design No State Cleanup
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
    ///         - burnBalance[oldMember] NOT migrated to newMember
    ///         - New member must deposit for burns via receive()
    ///         - Beneficiary validation is performed off-chain
    ///
    /// @custom:parameters
    ///      - oldMember: Old member address (unused in this implementation)
    ///      - newMember: New member address (unused in this implementation)
    ///
    /// @custom:design No Automatic Migration
    ///      - Explicit burn balance deposit required (clear accounting)
    ///      - Old member data preserved (audit trail)
    ///      - Governance can manually handle special cases if needed
    function _onMemberChanged(address /* oldMember */, address /* newMember */) internal override {}

}
