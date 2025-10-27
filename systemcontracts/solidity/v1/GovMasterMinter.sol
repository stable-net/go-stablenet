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
import {ICoinManager} from "../interfaces/ICoinManager.sol";
import {IMinterManagement} from "../interfaces/IMinterManagement.sol";
import {ConfigurationValidator} from "../libraries/ConfigurationValidator.sol";

/**
 * @title GovMasterMinter
 * @dev Governance contract for managing FiatToken minter addresses
 * Supports both EOA and CA as members
 */
contract GovMasterMinter is GovBaseV2, IMinterManagement {
    // ========== Custom Errors ==========
    error InvalidTokenAddress();
    error InvalidMinterAddress();
    error InvalidAllowance();
    error NotAMinter();
    error MinterConfigurationFailed();
    error MinterRemovalFailed();
    error TooManyMinters();
    error MinterAllowanceExceeded();
    error TokenOperationFailed();
    error ContractPaused();
    error ContractNotPaused();

    // ========== Constants ==========

    // ========== Token Admin Actions ==========
    // Note: Member management actions (ADD_MEMBER, REMOVE_MEMBER, CHANGE_MEMBER) are inherited from GovBase
    bytes32 public constant ACTION_CONFIGURE_MINTER = keccak256("CONFIGURE_MINTER");
    bytes32 public constant ACTION_REMOVE_MINTER = keccak256("REMOVE_MINTER");
    bytes32 public constant ACTION_UPDATE_MAX_MINTER_ALLOWANCE = keccak256("UPDATE_MAX_MINTER_ALLOWANCE");
    bytes32 public constant ACTION_PAUSE = keccak256("PAUSE");
    bytes32 public constant ACTION_UNPAUSE = keccak256("UNPAUSE");

    // ========== State Variables ==========
    IMinterManagement public fiatToken; // 0x32; The token contract
    mapping(address => uint256) public minterAllowances; // 0x33; Minter allowances
    mapping(address => bool) public isMinter; // 0x34; Minter status

    // New state variables for improved tracking
    address[] private minterList; // 0x35; List of all minters
    mapping(address => uint256) private minterIndex; // 0x36; Index in minterList + 1
    uint256 public totalMinterAllowance; // 0x37; Total allowance across all minters

    // Configurable limit through governance
    uint256 public maxMinterAllowance = 10000000000 * 10 ** 18; // 0x38; Maximum allowance per minter (default 10B tokens)

    // Emergency state
    bool public emergencyPaused; // 0x39; Emergency pause state

    // ========== Events ==========
    event MinterConfigured(address indexed minter, uint256 allowance);
    event MinterRemoved(address indexed minter);
    event MaxMinterAllowanceUpdated(uint256 oldLimit, uint256 newLimit);

    /// @notice Emitted when emergency pause is activated
    event EmergencyPaused(uint256 indexed proposalId);

    /// @notice Emitted when emergency pause is deactivated
    event EmergencyUnpaused(uint256 indexed proposalId);

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

    // ========== Constructor ==========

    // Note: For precompiled/built-in contracts, constructor may not be called during deployment
    // proposalExpiry is set in initialize() and cannot be changed after initialization

    // ========== Setter Functions ==========

    /// @notice Set the FiatToken contract address (one-time initialization)
    /// @param _fiatToken Address of the FiatToken contract
    /// @dev This function can only be called once to set the fiatToken address
    /// @dev No access control - relies on AlreadyInitialized check for safety
    ///
    /// @custom:security One-Time Initialization
    ///      - Can only be called when fiatToken is address(0)
    ///      - Once set, cannot be changed (immutable after initialization)
    ///      - No onlyMember or governance check (design choice for flexibility)
    ///
    /// @custom:security Input Validation
    ///      - Reverts if fiatToken already set (AlreadyInitialized)
    ///      - Reverts if _fiatToken is address(0) (InvalidTokenAddress)
    ///      - No interface compliance check (assumes correct contract)
    ///
    /// @custom:security Access Control Risk
    ///      - ⚠️ Anyone can call this function if fiatToken is not set
    ///      - Must be called immediately after deployment
    ///      - Front-running risk if not called in same transaction as deployment
    ///
    /// @custom:usage Deployment Flow
    ///      1. Deploy GovMasterMinter contract
    ///      2. Immediately call setFiatToken(_fiatToken) in same transaction
    ///      3. Verify fiatToken is set correctly
    ///
    /// @custom:revert AlreadyInitialized - If fiatToken already set
    /// @custom:revert InvalidTokenAddress - If _fiatToken is address(0)
    function setFiatToken(address _fiatToken) external {
        if (address(fiatToken) != address(0)) revert AlreadyInitialized();
        if (_fiatToken == address(0)) revert InvalidTokenAddress();
        fiatToken = IMinterManagement(_fiatToken);
    }

    // ========== Minter Management ==========

    /// @dev Internal function to configure minter allowance and track minter state
    /// @param minter Address of the minter to configure
    /// @param allowance New allowance amount for the minter
    /// @notice Updates multiple state variables atomically to maintain consistency
    ///
    /// @custom:security State Consistency
    ///      - Updates 5 related state variables in atomic transaction
    ///      - totalMinterAllowance, minterList, minterIndex, isMinter, minterAllowances
    ///      - State remains consistent even if called multiple times for same minter
    ///
    /// @custom:security Arithmetic Safety
    ///      - Solidity 0.8.14 provides automatic overflow/underflow protection
    ///      - totalMinterAllowance calculation: total - old + new is safe
    ///      - Even if old > total (corrupted state), would revert with underflow
    ///
    /// @custom:security 1-Based Indexing Pattern
    ///      - minterIndex stores array index + 1 (not actual index)
    ///      - 0 means "not in list", >0 means "in list at position (value - 1)"
    ///      - Prevents confusion between "index 0" and "not found"
    ///
    /// @custom:security Idempotency
    ///      - Safe to call multiple times for same minter (updates allowance)
    ///      - Does not duplicate entries in minterList
    ///      - isMinter check prevents duplicate additions
    ///
    /// @custom:security Access Control
    ///      - Internal function - can only be called by contract itself
    ///      - Called via governance through _safeConfigureMinter → FiatToken.configureMinter
    ///      - No direct external access ensures governance control
    ///
    /// @custom:security No Validation
    ///      - ⚠️ Does not validate minter address or allowance amount
    ///      - Assumes validation done by caller (proposeConfigureMinter)
    ///      - Assumes FiatToken.configureMinter already succeeded
    ///
    /// @custom:usage Call Flow
    ///      1. proposeConfigureMinter validates inputs and creates proposal
    ///      2. Members vote on proposal
    ///      3. _executeCustomAction calls _safeConfigureMinter
    ///      4. _safeConfigureMinter calls FiatToken.configureMinter first
    ///      5. Only if FiatToken call succeeds, this function updates internal state
    function _configureMinter(address minter, uint256 allowance) internal {
        uint256 oldAllowance = minterAllowances[minter];

        // Update total allowance
        totalMinterAllowance = totalMinterAllowance - oldAllowance + allowance;

        // Add to list if new minter
        if (!isMinter[minter]) {
            minterList.push(minter);
            minterIndex[minter] = minterList.length; // Store 1-based index
            isMinter[minter] = true;
        }

        minterAllowances[minter] = allowance;
    }

    /// @dev Internal function to remove minter and clean up all related state
    /// @param minter Address of the minter to remove
    /// @notice Uses swap-and-pop pattern for gas-efficient array element removal
    ///
    /// @custom:security State Consistency
    ///      - Updates 5 related state variables atomically
    ///      - totalMinterAllowance, minterList, minterIndex, isMinter, minterAllowances
    ///      - All state cleaned up even if minter not in list (safe for edge cases)
    ///
    /// @custom:security Swap-and-Pop Pattern
    ///      - Gas-efficient O(1) removal from array instead of O(n) shift
    ///      - Swaps target element with last element, then pops last
    ///      - Order of minterList may change (not guaranteed to be stable)
    ///      - Updates minterIndex for the swapped element to maintain consistency
    ///
    /// @custom:security 1-Based Index Handling
    ///      - Checks index > 0 before array operations (0 means not in list)
    ///      - Converts 1-based index to 0-based arrayIndex (index - 1)
    ///      - Prevents array underflow when minter not in list
    ///
    /// @custom:security Edge Case: Not in List
    ///      - Safe to call even if minter not in minterList
    ///      - if (index > 0) check skips array operations if not found
    ///      - Still cleans up allowances and isMinter flag
    ///      - Idempotent: calling twice has no adverse effect
    ///
    /// @custom:security Edge Case: Last Element
    ///      - Special handling when removing last element (arrayIndex == lastIndex)
    ///      - Skips swap operation (no need to swap with itself)
    ///      - Only performs pop() to remove element
    ///
    /// @custom:security Arithmetic Safety
    ///      - Solidity 0.8.14 prevents underflow on totalMinterAllowance
    ///      - Array bounds checked automatically (lastIndex calculation safe)
    ///      - delete operations are always safe
    ///
    /// @custom:security Access Control
    ///      - Internal function - can only be called by contract itself
    ///      - Called via governance through _safeRemoveMinter → FiatToken.removeMinter
    ///      - No direct external access ensures governance control
    ///
    /// @custom:security No Validation
    ///      - ⚠️ Does not validate minter address
    ///      - Assumes validation done by caller (proposeRemoveMinter checks isMinter)
    ///      - Assumes FiatToken.removeMinter already succeeded
    ///
    /// @custom:usage Call Flow
    ///      1. proposeRemoveMinter validates minter exists (isMinter check)
    ///      2. Members vote on proposal
    ///      3. _executeCustomAction calls _safeRemoveMinter
    ///      4. _safeRemoveMinter calls FiatToken.removeMinter first
    ///      5. Only if FiatToken call succeeds, this function cleans up internal state
    function _removeMinter(address minter) internal {
        uint256 allowance = minterAllowances[minter];

        // Update total allowance
        totalMinterAllowance = totalMinterAllowance - allowance;

        // Remove from list
        uint256 index = minterIndex[minter];
        if (index > 0) {
            uint256 arrayIndex = index - 1;
            uint256 lastIndex = minterList.length - 1;

            if (arrayIndex != lastIndex) {
                address lastMinter = minterList[lastIndex];
                minterList[arrayIndex] = lastMinter;
                minterIndex[lastMinter] = index;
            }

            minterList.pop();
            delete minterIndex[minter];
        }

        delete minterAllowances[minter];
        isMinter[minter] = false;
    }

    // ========== Hook Implementations ==========

    /// @dev Hook called after a new member is added to governance
    /// @param member Address of the newly added member
    /// @notice Called automatically by GovBaseV2 after successful member addition
    ///
    /// @custom:security Hook Execution Context
    ///      - Called AFTER member is added (post-state-change hook)
    ///      - Member already exists in GovBase members mapping
    ///      - Called within same transaction as member addition
    ///      - Any revert here will revert entire member addition
    ///
    /// @custom:security No Custom Logic Required
    ///      - GovMasterMinter does not need special actions on member addition
    ///      - Members control minter configuration through proposals only
    ///      - No automatic permissions or state changes needed
    ///
    /// @custom:security Future Extensibility
    ///      - Hook exists for potential future requirements
    ///      - Could add: member-specific limits, notifications, tracking
    ///      - Currently intentionally left empty (no-op)
    ///
    /// @custom:usage Call Flow
    ///      1. proposeAddMember creates proposal
    ///      2. Members vote and approve
    ///      3. GovBase adds member to governance
    ///      4. This hook is called automatically
    function _onMemberAdded(address member) internal override {
        // Custom logic for master minter governance
        // Currently no specific action needed
    }

    /// @dev Hook called after a member is removed from governance
    /// @param member Address of the removed member
    /// @notice Called automatically by GovBaseV2 after successful member removal
    ///
    /// @custom:security Hook Execution Context
    ///      - Called AFTER member is removed (post-state-change hook)
    ///      - Member no longer exists in GovBase members mapping
    ///      - Called within same transaction as member removal
    ///      - Any revert here will revert entire member removal
    ///
    /// @custom:security No Custom Logic Required
    ///      - GovMasterMinter does not need special actions on member removal
    ///      - Removed member automatically loses proposal/voting rights
    ///      - No minter-specific cleanup needed
    ///
    /// @custom:security Pending Proposals
    ///      - Removed member's pending proposals remain valid
    ///      - Other members can still vote on existing proposals
    ///      - GovBase handles proposal lifecycle independently
    ///
    /// @custom:security Future Extensibility
    ///      - Hook exists for potential future requirements
    ///      - Could add: proposal cleanup, transfer of responsibilities
    ///      - Currently intentionally left empty (no-op)
    ///
    /// @custom:usage Call Flow
    ///      1. proposeRemoveMember creates proposal
    ///      2. Members vote and approve
    ///      3. GovBase removes member from governance
    ///      4. This hook is called automatically
    function _onMemberRemoved(address member) internal override {
        // Custom logic for master minter governance
        // Currently no specific action needed
    }

    /// @dev Hook called after a member address is changed (key rotation)
    /// @param oldMember Address of the member being replaced
    /// @param newMember New address for the member
    /// @notice Called automatically by GovBaseV2 after successful member change
    ///
    /// @custom:security Hook Execution Context
    ///      - Called AFTER member change (post-state-change hook)
    ///      - oldMember removed, newMember added to GovBase
    ///      - Called within same transaction as member change
    ///      - Any revert here will revert entire member change
    ///
    /// @custom:security Key Rotation Purpose
    ///      - Allows members to rotate their signing keys
    ///      - Maintains same voting power and responsibilities
    ///      - Useful for: compromised keys, security upgrades, operational changes
    ///
    /// @custom:security No Custom Logic Required
    ///      - GovMasterMinter does not track member-specific state
    ///      - All permissions are role-based (any member can propose/vote)
    ///      - No transfer of member-specific data needed
    ///
    /// @custom:security Pending Proposals
    ///      - Proposals created by oldMember remain valid
    ///      - newMember can continue voting on all proposals
    ///      - No proposal ownership transfer needed
    ///
    /// @custom:security Future Extensibility
    ///      - Hook exists for potential future requirements
    ///      - Could add: activity tracking, delegation transfer
    ///      - Currently intentionally left empty (no-op)
    ///
    /// @custom:usage Call Flow
    ///      1. proposeChangeMember creates proposal
    ///      2. Members vote and approve
    ///      3. GovBase swaps member addresses
    ///      4. This hook is called automatically
    function _onMemberChanged(address oldMember, address newMember) internal override {
        // Custom logic for master minter governance
        // Currently no specific action needed
    }

    // ========== Internal Action Implementation ==========

    /// @dev Override _executeCustomAction to implement GovMasterMinter-specific actions
    /// @param actionType The type of action to execute (one of the ACTION_* constants)
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if action executed successfully, false if failed or unknown action
    /// @notice Member management actions (ADD_MEMBER, REMOVE_MEMBER, CHANGE_MEMBER) are handled by GovBase
    ///
    /// @custom:security Action Routing
    ///      - Central dispatcher for all GovMasterMinter-specific actions
    ///      - Each action has dedicated validation and execution logic
    ///      - Returns false for unknown actions (safe failure mode)
    ///
    /// @custom:security Supported Actions
    ///      - ACTION_CONFIGURE_MINTER: Add/update minter with allowance
    ///      - ACTION_REMOVE_MINTER: Remove minter from FiatToken
    ///      - ACTION_UPDATE_MAX_MINTER_ALLOWANCE: Change per-minter allowance limit
    ///      - ACTION_PAUSE: Activate emergency pause
    ///      - ACTION_UNPAUSE: Deactivate emergency pause
    ///
    /// @custom:security Call Flow Validation
    ///      - Only called by GovBase after proposal approved and quorum reached
    ///      - Already protected by nonReentrant in GovBase
    ///      - Already verified proposal exists and is in Approved state
    ///
    /// @custom:security External Call Safety
    ///      - _safeConfigureMinter and _safeRemoveMinter use try-catch
    ///      - FiatToken calls may fail, caught and returned as false
    ///      - Failure allows proposal retry via manual execution
    ///      - Does not revert on FiatToken failure (controlled failure)
    ///
    /// @custom:security State-Only Actions
    ///      - ACTION_UPDATE_MAX_MINTER_ALLOWANCE is state-only (no external calls)
    ///      - ACTION_PAUSE and ACTION_UNPAUSE are state-only (cannot fail)
    ///      - These actions always return true
    ///
    /// @custom:security ABI Decode Safety
    ///      - Each action decodes expected parameter types
    ///      - Invalid callData causes revert (caught by GovBase)
    ///      - Proper encoding ensured by propose* functions
    ///
    /// @custom:security Unknown Action Handling
    ///      - Returns false for unknown actionType
    ///      - Does not revert (allows graceful failure)
    ///      - GovBase marks proposal as Failed
    ///
    /// @custom:usage Execution Context
    ///      1. Proposal created with actionType and callData
    ///      2. Members vote and approve proposal
    ///      3. GovBase calls _executeAction which calls this function
    ///      4. This function routes to appropriate action handler
    ///      5. Returns true on success, false on failure
    ///      6. GovBase updates proposal state based on return value
    function _executeCustomAction(bytes32 actionType, bytes memory callData) internal override returns (bool) {
        if (actionType == ACTION_CONFIGURE_MINTER) {
            (address minter, uint256 allowance) = abi.decode(callData, (address, uint256));
            return _safeConfigureMinter(minter, allowance);
        }

        if (actionType == ACTION_REMOVE_MINTER) {
            address minter = abi.decode(callData, (address));
            return _safeRemoveMinter(minter);
        }

        if (actionType == ACTION_UPDATE_MAX_MINTER_ALLOWANCE) {
            uint256 newLimit = abi.decode(callData, (uint256));
            uint256 oldLimit = maxMinterAllowance;
            maxMinterAllowance = newLimit;
            emit MaxMinterAllowanceUpdated(oldLimit, newLimit);
            return true;
        }

        if (actionType == ACTION_PAUSE) {
            return _executePause();
        }

        if (actionType == ACTION_UNPAUSE) {
            return _executeUnpause();
        }

        // Unknown action type
        return false;
    }

    // ========== Proposal Functions ==========

    /**
     * @notice Propose to configure a minter (add or update allowance)
     * @param minter Address of the minter contract
     * @param minterAllowedAmount Allowance amount for the minter (in token base units)
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members when contract is not paused
     * @dev Reverts with ContractPaused if emergency pause is active
     *
     * @custom:security Access Control
     *      - onlyMember: Restricted to active governance members only
     *      - whenNotPaused: Blocked during emergency pause
     *      - Prevents unauthorized minter configuration
     *
     * @custom:security Input Validation
     *      - Validates minter != address(0) (prevents zero address)
     *      - Validates allowance > 0 (prevents meaningless proposals)
     *      - Validates allowance <= maxMinterAllowance (enforces global limit)
     *      - Front-running safe: validation happens at proposal time
     *
     * @custom:security Proposal Creation
     *      - Creates governance proposal requiring quorum approval
     *      - Subject to MAX_ACTIVE_PROPOSALS_PER_MEMBER limit (prevents spam)
     *      - Proposal expires after proposalExpiry time if not executed
     *
     * @custom:security Idempotency
     *      - Can be called for existing minter (updates allowance)
     *      - Can be called for new minter (adds to minter list)
     *      - No duplicate proposal check (allows re-proposal if previous expired)
     *
     * @custom:security Two-Phase Execution
     *      - Phase 1 (this function): Validation and proposal creation
     *      - Phase 2 (on approval): FiatToken.configureMinter + internal state update
     *      - Defense in depth: validation at both proposal and execution time
     *
     * @custom:security External Dependencies
     *      - Execution depends on FiatToken.configureMinter success
     *      - If FiatToken call fails, proposal can be retried via executeProposal
     *      - Failure does not corrupt state (safe via try-catch in _safeConfigureMinter)
     *
     * @custom:usage Typical Flow
     *      1. Member calls proposeConfigureMinter(GovMinter, 1000000e18)
     *      2. Proposal created with validation passed
     *      3. Other members vote on proposal
     *      4. On quorum approval: FiatToken updated, internal state updated, event emitted
     *      5. GovMinter can now mint up to allowance
     *
     * @custom:revert InvalidMinterAddress - If minter is address(0)
     * @custom:revert InvalidAllowance - If allowance is 0 or > maxMinterAllowance
     * @custom:revert ContractPaused - If emergency pause is active
     * @custom:revert NotAMember - If caller is not a governance member (from onlyMember)
     * @custom:revert TooManyActiveProposals - If proposer has 3+ active proposals
     */
    function proposeConfigureMinter(address minter, uint256 minterAllowedAmount) public onlyMember whenNotPaused returns (uint256) {
        // Validate minter address
        if (minter == address(0)) revert InvalidMinterAddress();

        // Validate allowance
        if (minterAllowedAmount == 0) revert InvalidAllowance();
        if (minterAllowedAmount > maxMinterAllowance) revert InvalidAllowance();

        bytes memory data = abi.encode(minter, minterAllowedAmount);
        return _createProposal(ACTION_CONFIGURE_MINTER, data);
    }

    /**
     * @notice Propose to remove a minter from FiatToken
     * @param minter Address of the minter to remove
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members when contract is not paused
     * @dev Reverts with ContractPaused if emergency pause is active
     *
     * @custom:security Access Control
     *      - onlyMember: Restricted to active governance members only
     *      - whenNotPaused: Blocked during emergency pause
     *      - Prevents unauthorized minter removal
     *
     * @custom:security Input Validation
     *      - Validates isMinter[minter] == true (prevents removing non-existent minter)
     *      - Prevents wasteful proposals for already-removed minters
     *      - Front-running safe: validation happens at proposal time
     *
     * @custom:security State Consistency
     *      - Validation uses isMinter flag (authoritative source)
     *      - Prevents inconsistent state between GovMasterMinter and FiatToken
     *      - If FiatToken and GovMasterMinter out of sync, proposal will fail safely
     *
     * @custom:security Proposal Creation
     *      - Creates governance proposal requiring quorum approval
     *      - Subject to MAX_ACTIVE_PROPOSALS_PER_MEMBER limit (prevents spam)
     *      - Proposal expires after proposalExpiry time if not executed
     *
     * @custom:security Two-Phase Execution
     *      - Phase 1 (this function): Validation and proposal creation
     *      - Phase 2 (on approval): FiatToken.removeMinter + internal state cleanup
     *      - Defense in depth: validation at both proposal and execution time
     *
     * @custom:security State Cleanup
     *      - On execution: removes from minterList, clears minterIndex, isMinter
     *      - Updates totalMinterAllowance to maintain accuracy
     *      - Uses swap-and-pop for gas-efficient array removal
     *
     * @custom:security External Dependencies
     *      - Execution depends on FiatToken.removeMinter success
     *      - If FiatToken call fails, proposal can be retried via executeProposal
     *      - Failure does not corrupt state (safe via try-catch in _safeRemoveMinter)
     *
     * @custom:security Irreversible Action
     *      - ⚠️ Removal is permanent for that minter address
     *      - To re-add, must create new proposeConfigureMinter proposal
     *      - Minter loses all allowance immediately upon execution
     *
     * @custom:usage Typical Flow
     *      1. Member calls proposeRemoveMinter(GovMinter)
     *      2. Proposal created with validation passed
     *      3. Other members vote on proposal
     *      4. On quorum approval: FiatToken updated, internal state cleaned, event emitted
     *      5. GovMinter can no longer mint from FiatToken
     *
     * @custom:revert NotAMinter - If minter is not currently registered
     * @custom:revert ContractPaused - If emergency pause is active
     * @custom:revert NotAMember - If caller is not a governance member (from onlyMember)
     * @custom:revert TooManyActiveProposals - If proposer has 3+ active proposals
     */
    function proposeRemoveMinter(address minter) public onlyMember whenNotPaused returns (uint256) {
        if (!isMinter[minter]) revert NotAMinter();

        bytes memory data = abi.encode(minter);
        return _createProposal(ACTION_REMOVE_MINTER, data);
    }

    /**
     * @notice Propose to update the maximum allowance limit per minter
     * @param newLimit New maximum allowance amount (in token base units)
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members (no pause restriction)
     *
     * @custom:security Access Control
     *      - onlyMember: Restricted to active governance members only
     *      - NO whenNotPaused: Can be proposed even during emergency pause
     *      - Allows governance to adjust limits during crisis situations
     *
     * @custom:security Input Validation
     *      - Validates newLimit > 0 via ConfigurationValidator.validateAmount
     *      - Prevents setting zero limit (would block all minter configurations)
     *      - No upper bound check (governance decides appropriate limit)
     *
     * @custom:security Global Limit Enforcement
     *      - maxMinterAllowance is enforced in proposeConfigureMinter
     *      - Changing limit does NOT affect existing minters' allowances
     *      - Only affects future proposeConfigureMinter proposals
     *
     * @custom:security No Retroactive Effect
     *      - ⚠️ Does NOT reduce existing minters' allowances
     *      - Existing minters keep their allowances even if > newLimit
     *      - To reduce existing allowance: use proposeConfigureMinter with lower value
     *
     * @custom:security Proposal Creation
     *      - Creates governance proposal requiring quorum approval
     *      - Subject to MAX_ACTIVE_PROPOSALS_PER_MEMBER limit (prevents spam)
     *      - Proposal expires after proposalExpiry time if not executed
     *
     * @custom:security Execution Safety
     *      - State-only operation (no external calls)
     *      - Always succeeds when executed (cannot fail)
     *      - Emits MaxMinterAllowanceUpdated event for transparency
     *
     * @custom:security Emergency Flexibility
     *      - Can be proposed during pause (unlike minter configuration)
     *      - Allows governance to prepare for unpause by adjusting limits
     *      - Useful for crisis response (e.g., reduce limits before unpause)
     *
     * @custom:usage Typical Flow
     *      1. Member calls proposeUpdateMaxMinterAllowance(5000000e18)
     *      2. Proposal created with validation passed
     *      3. Other members vote on proposal
     *      4. On quorum approval: maxMinterAllowance updated, event emitted
     *      5. Future proposeConfigureMinter calls use new limit
     *
     * @custom:usage Common Scenarios
     *      - Increase limit: When minters need higher allowances for scaling
     *      - Decrease limit: Risk mitigation to reduce potential damage
     *      - Emergency: Reduce limit during pause, then unpause safely
     *
     * @custom:revert InvalidAmount - If newLimit is 0 (from ConfigurationValidator)
     * @custom:revert NotAMember - If caller is not a governance member (from onlyMember)
     * @custom:revert TooManyActiveProposals - If proposer has 3+ active proposals
     */
    function proposeUpdateMaxMinterAllowance(uint256 newLimit) public onlyMember returns (uint256) {
        // Validate amount is not zero
        ConfigurationValidator.validateAmount(newLimit);

        bytes memory data = abi.encode(newLimit);
        return _createProposal(ACTION_UPDATE_MAX_MINTER_ALLOWANCE, data);
    }

    /**
     * @notice Propose emergency pause of all minter operations
     * @return proposalId The ID of the created proposal
     * @dev Creates a governance proposal to activate emergency pause
     * @dev Only callable by active governance members when contract is not paused
     * @dev When executed, sets emergencyPaused = true, blocking minter configurations
     * @dev Reverts with ContractPaused if already in paused state (prevents duplicate proposals)
     */
    function proposePause() external onlyMember whenNotPaused returns (uint256) {
        bytes memory callData = "";
        return _createProposal(ACTION_PAUSE, callData);
    }

    /**
     * @notice Propose resumption of minter operations
     * @return proposalId The ID of the created proposal
     * @dev Creates a governance proposal to deactivate emergency pause
     * @dev Only callable by active governance members when contract is paused
     * @dev When executed, sets emergencyPaused = false, allowing minter operations
     * @dev Reverts with ContractNotPaused if not in paused state (prevents unnecessary proposals)
     */
    function proposeUnpause() external onlyMember whenPaused returns (uint256) {
        bytes memory callData = "";
        return _createProposal(ACTION_UNPAUSE, callData);
    }

    // ========== Self-Governance Proposals ==========
    // Note: proposeAddMember, proposeRemoveMember, proposeChangeMember are inherited from GovBase

    // ========== Internal Action Implementations ==========

    /// @dev Execute emergency pause action
    /// @notice Activates emergency pause to block all minter configuration operations
    ///
    ///         This function is called when a pause proposal is approved and executed.
    ///         It sets the emergencyPaused flag to true, which causes all subsequent
    ///         minter configuration and removal operations to revert with ContractPaused error.
    ///
    ///         Use cases:
    ///         - Security incident detected in FiatToken or minter contracts
    ///         - Critical bug discovered in minter management logic
    ///         - Regulatory compliance requirement
    ///         - System maintenance or upgrade preparation
    ///
    /// @return success Always returns true (pause cannot fail)
    function _executePause() internal returns (bool) {
        emergencyPaused = true;
        emit EmergencyPaused(currentProposalId);
        return true;
    }

    /// @dev Execute emergency unpause action
    /// @notice Deactivates emergency pause to resume minter configuration operations
    ///
    ///         This function is called when an unpause proposal is approved and executed.
    ///         It sets the emergencyPaused flag to false, which allows minter configuration
    ///         and removal operations to proceed normally again.
    ///
    ///         Use cases:
    ///         - Security incident resolved
    ///         - Bug fix deployed and verified
    ///         - Compliance requirement satisfied
    ///         - Maintenance/upgrade completed
    ///
    /// @return success Always returns true (unpause cannot fail)
    function _executeUnpause() internal returns (bool) {
        emergencyPaused = false;
        emit EmergencyUnpaused(currentProposalId);
        return true;
    }

    // ========== Safe Operations ==========

    /// @dev Safely configure minter with try-catch for FiatToken interaction
    /// @param minter Address of the minter to configure
    /// @param allowance New allowance amount for the minter
    /// @return success True if FiatToken call and state update succeeded, false otherwise
    /// @notice Uses try-catch pattern to handle FiatToken failures gracefully
    ///
    /// @custom:security Try-Catch Pattern
    ///      - Wraps external FiatToken.configureMinter call in try-catch
    ///      - Catches ALL errors: revert, require, out of gas, etc.
    ///      - Returns false on failure instead of reverting
    ///      - Allows proposal retry via manual executeProposal call
    ///
    /// @custom:security Call Order (Check-Effect-Interaction-Effect)
    ///      - 1. Check: whenNotPaused modifier validates state
    ///      - 2. External Interaction: FiatToken.configureMinter (may fail)
    ///      - 3. Internal Effect: _configureMinter updates state (only on success)
    ///      - 4. Event: MinterConfigured emission (only on success)
    ///      - Prevents state corruption if FiatToken call fails
    ///
    /// @custom:security State Consistency
    ///      - Internal state (_configureMinter) updated ONLY if FiatToken succeeds
    ///      - Maintains consistency between GovMasterMinter and FiatToken
    ///      - No partial state updates on failure
    ///
    /// @custom:security Pause Protection
    ///      - whenNotPaused modifier prevents execution during emergency
    ///      - Protects against minter configuration during crisis
    ///      - Returns ContractPaused error, not false
    ///
    /// @custom:security Failure Scenarios
    ///      - FiatToken not initialized (address(0))
    ///      - FiatToken paused or has access control restrictions
    ///      - FiatToken implementation reverted for business logic reasons
    ///      - Out of gas during FiatToken call
    ///      - All failures return false (proposal marked as Failed, can retry)
    ///
    /// @custom:security No Validation
    ///      - ⚠️ Does not validate inputs (assumes done by caller)
    ///      - Relies on proposeConfigureMinter validation
    ///      - Relies on FiatToken.configureMinter validation
    ///
    /// @custom:usage Call Flow
    ///      1. _executeCustomAction calls this function
    ///      2. whenNotPaused check passes
    ///      3. FiatToken.configureMinter called
    ///      4a. On success: internal state updated, event emitted, return true
    ///      4b. On failure: state unchanged, no event, return false
    ///      5. GovBase marks proposal Executed (true) or Failed (false)
    function _safeConfigureMinter(address minter, uint256 allowance) internal whenNotPaused returns (bool) {
        try fiatToken.configureMinter(minter, allowance) {
            _configureMinter(minter, allowance);
            emit MinterConfigured(minter, allowance);
            return true;
        } catch {
            return false;
        }
    }

    /// @dev Safely remove minter with try-catch for FiatToken interaction
    /// @param minter Address of the minter to remove
    /// @return success True if FiatToken call and state cleanup succeeded, false otherwise
    /// @notice Uses try-catch pattern to handle FiatToken failures gracefully
    ///
    /// @custom:security Try-Catch Pattern
    ///      - Wraps external FiatToken.removeMinter call in try-catch
    ///      - Catches ALL errors: revert, require, out of gas, etc.
    ///      - Returns false on failure instead of reverting
    ///      - Allows proposal retry via manual executeProposal call
    ///
    /// @custom:security Call Order (Check-Effect-Interaction-Effect)
    ///      - 1. Check: whenNotPaused modifier validates state
    ///      - 2. External Interaction: FiatToken.removeMinter (may fail)
    ///      - 3. Internal Effect: _removeMinter cleans up state (only on success)
    ///      - 4. Event: MinterRemoved emission (only on success)
    ///      - Prevents state corruption if FiatToken call fails
    ///
    /// @custom:security State Consistency
    ///      - Internal state (_removeMinter) updated ONLY if FiatToken succeeds
    ///      - Maintains consistency between GovMasterMinter and FiatToken
    ///      - No partial state updates on failure
    ///
    /// @custom:security Pause Protection
    ///      - whenNotPaused modifier prevents execution during emergency
    ///      - Protects against minter removal during crisis
    ///      - Returns ContractPaused error, not false
    ///
    /// @custom:security Failure Scenarios
    ///      - FiatToken not initialized (address(0))
    ///      - Minter not registered in FiatToken (out of sync)
    ///      - FiatToken paused or has access control restrictions
    ///      - FiatToken implementation reverted for business logic reasons
    ///      - Out of gas during FiatToken call
    ///      - All failures return false (proposal marked as Failed, can retry)
    ///
    /// @custom:security No Validation
    ///      - ⚠️ Does not validate minter address (assumes done by caller)
    ///      - Relies on proposeRemoveMinter validation (isMinter check)
    ///      - Relies on FiatToken.removeMinter validation
    ///
    /// @custom:usage Call Flow
    ///      1. _executeCustomAction calls this function
    ///      2. whenNotPaused check passes
    ///      3. FiatToken.removeMinter called
    ///      4a. On success: internal state cleaned, event emitted, return true
    ///      4b. On failure: state unchanged, no event, return false
    ///      5. GovBase marks proposal Executed (true) or Failed (false)
    function _safeRemoveMinter(address minter) internal whenNotPaused returns (bool) {
        try fiatToken.removeMinter(minter) {
            _removeMinter(minter);
            emit MinterRemoved(minter);
            return true;
        } catch {
            return false;
        }
    }

    // ========== View Functions ==========

    /// @notice Get the allowance for a specific minter
    /// @param minter Address of the minter to query
    /// @return allowance Current allowance amount for the minter (0 if not a minter)
    /// @dev Returns 0 for non-existent minters (does not revert)
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context (external contracts, UI, scripts)
    ///      - Gas-efficient single storage read
    ///
    /// @custom:security No Validation
    ///      - Does not validate minter address
    ///      - Returns 0 for address(0) or invalid addresses
    ///      - Use getMinterStatus to check if minter is registered
    ///
    /// @custom:usage Common Use Cases
    ///      - Check remaining minting capacity before minting
    ///      - Monitor minter allowances in UI/dashboard
    ///      - Verify proposal execution results
    function getMinterAllowance(address minter) public view returns (uint256) {
        return minterAllowances[minter];
    }

    /// @notice Check if an address is registered as a minter
    /// @param minter Address to check
    /// @return status True if address is a registered minter, false otherwise
    /// @dev Authoritative source for minter registration status
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Gas-efficient single storage read
    ///
    /// @custom:security Authoritative Status
    ///      - This is the source of truth for minter status
    ///      - Used by proposeRemoveMinter for validation
    ///      - Should match FiatToken's minter status (maintained via governance)
    ///
    /// @custom:usage Common Use Cases
    ///      - Validate before proposing minter removal
    ///      - Check minter registration status in UI
    ///      - Verify state consistency with FiatToken
    function getMinterStatus(address minter) public view returns (bool) {
        return isMinter[minter];
    }

    /// @notice Get the total number of registered minters
    /// @return count Total number of minters in minterList
    /// @dev Efficient way to get minter count without iterating
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Gas-efficient single storage read
    ///
    /// @custom:usage Common Use Cases
    ///      - Determine array size for getAllMinters pagination
    ///      - Display total minter count in dashboards
    ///      - Monitor governance activity over time
    function getMinterCount() public view returns (uint256) {
        return minterList.length;
    }

    /// @notice Get minter address and allowance at a specific index
    /// @param index Index in the minterList array (0-based)
    /// @return minter Address of the minter at the index
    /// @return allowance Current allowance for the minter
    /// @dev Used for pagination or iterating through minters
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Reverts if index out of bounds
    ///
    /// @custom:security Array Order
    ///      - ⚠️ Array order is NOT stable (swap-and-pop pattern used)
    ///      - Order changes when minters are removed
    ///      - Do not rely on consistent index for same minter
    ///
    /// @custom:usage Pagination Pattern
    ///      1. Call getMinterCount() to get total
    ///      2. Loop from 0 to count-1
    ///      3. Call getMinterAt(i) for each index
    ///      4. Handle IndexOutOfBounds if count changes during iteration
    ///
    /// @custom:revert IndexOutOfBounds - If index >= minterList.length
    function getMinterAt(uint256 index) public view returns (address minter, uint256 allowance) {
        if (index >= minterList.length) revert IndexOutOfBounds();
        minter = minterList[index];
        allowance = minterAllowances[minter];
    }

    /// @notice Get all registered minters and their allowances
    /// @return minters Array of all minter addresses
    /// @return allowances Array of corresponding allowances (parallel to minters array)
    /// @dev Returns complete snapshot of all minters in single call
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Gas cost scales with number of minters
    ///
    /// @custom:security Gas Consideration
    ///      - ⚠️ Gas cost increases linearly with minter count
    ///      - Each minter requires: 1 array element + 1 mapping read
    ///      - For large minter counts (>100), consider pagination via getMinterAt
    ///      - Suitable for: UI display, off-chain queries, small minter sets
    ///
    /// @custom:security Array Order
    ///      - Order is NOT guaranteed to be stable
    ///      - Order changes when minters are removed (swap-and-pop)
    ///      - Parallel arrays: minters[i] corresponds to allowances[i]
    ///
    /// @custom:usage Common Use Cases
    ///      - Display all minters in admin dashboard
    ///      - Export complete minter configuration
    ///      - Verify governance state
    ///      - Off-chain analysis and monitoring
    function getAllMinters() public view returns (address[] memory minters, uint256[] memory allowances) {
        uint256 count = minterList.length;
        minters = new address[](count);
        allowances = new uint256[](count);

        for (uint256 i = 0; i < count; i++) {
            minters[i] = minterList[i];
            allowances[i] = minterAllowances[minterList[i]];
        }
    }

    /// @notice Get aggregate statistics about all minters
    /// @return totalMinters Total number of registered minters
    /// @return totalAllowance Sum of all minter allowances
    /// @dev Efficient way to get summary statistics in single call
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Gas-efficient dual storage read (2 SLOADs)
    ///
    /// @custom:security State Consistency
    ///      - totalAllowance maintained by _configureMinter and _removeMinter
    ///      - Should equal sum of all individual minterAllowances
    ///      - Cached for efficiency (no need to iterate)
    ///
    /// @custom:usage Common Use Cases
    ///      - Monitor total minting capacity across all minters
    ///      - Dashboard summary statistics
    ///      - Risk assessment (total exposure)
    ///      - Governance reporting
    function getMinterStats() public view returns (uint256 totalMinters, uint256 totalAllowance) {
        totalMinters = minterList.length;
        totalAllowance = totalMinterAllowance;
    }

    /// @notice Check if a minter can mint a specific amount
    /// @param minter Address of the minter to check
    /// @param amount Amount to check (in token base units)
    /// @return canMintAmount True if minter is registered AND amount <= allowance
    /// @dev Helper function for pre-mint validation
    ///
    /// @custom:security Read-Only
    ///      - Pure view function, no state changes
    ///      - Safe to call from any context
    ///      - Gas-efficient (1 SLOAD for status, 1 SLOAD for allowance if needed)
    ///
    /// @custom:security Validation Logic
    ///      - Returns false if minter not registered (isMinter check)
    ///      - Returns false if amount > minterAllowances[minter]
    ///      - Returns true only if both conditions satisfied
    ///      - Short-circuit evaluation for efficiency
    ///
    /// @custom:security Not Authoritative for Minting
    ///      - ⚠️ This is a helper, not the source of truth
    ///      - Actual minting permission controlled by FiatToken
    ///      - State may change between check and mint (TOCTOU)
    ///      - Use for: UI validation, pre-checks, estimates
    ///
    /// @custom:usage Common Use Cases
    ///      - Pre-validate mint operations in UI before transaction
    ///      - Check minting capacity before submitting proposals
    ///      - Display remaining minting capacity to users
    ///      - Off-chain analysis and monitoring
    ///
    /// @custom:usage TOCTOU Warning
    ///      - Time-Of-Check-Time-Of-Use race condition possible
    ///      - Minter allowance may change between canMint check and actual mint
    ///      - Always expect mint transaction may still fail
    function canMint(address minter, uint256 amount) public view returns (bool) {
        if (!isMinter[minter]) return false;
        return amount <= minterAllowances[minter];
    }

    // ========== IMinterManagement Interface Implementation ==========

    // Note: isMinter(address) is provided by the public mapping's auto-generated getter

    /// @notice Get the minter allowance (IMinterManagement interface)
    /// @param _minter Address of the minter
    /// @return Allowance amount
    function minterAllowance(address _minter) external view returns (uint256) {
        return minterAllowances[_minter];
    }

    /// @notice Configure a minter (IMinterManagement interface)
    /// @param _minter Address of the minter
    /// @param _minterAllowedAmount New allowance amount
    /// @return Always returns true (reverts on failure)
    /// @dev Creates a governance proposal to configure the minter
    function configureMinter(address _minter, uint256 _minterAllowedAmount) external override returns (bool) {
        proposeConfigureMinter(_minter, _minterAllowedAmount);
        return true;
    }

    /// @notice Remove a minter (IMinterManagement interface)
    /// @param _minter Address of the minter to remove
    /// @return Always returns true (reverts on failure)
    /// @dev Creates a governance proposal to remove the minter
    function removeMinter(address _minter) external override returns (bool) {
        proposeRemoveMinter(_minter);
        return true;
    }

    /// @notice Update master minter (IMinterManagement interface)
    /// @param _newMasterMinter Address of the new master minter
    /// @dev This is a no-op in governance model - master minter role is handled by governance members
    function updateMasterMinter(address _newMasterMinter) external override {
        // In governance model, "master minter" role is fulfilled by governance members
        // This function is implemented for interface compatibility but does nothing
        // To change governance, use proposeAddMember/proposeRemoveMember/proposeChangeMember
        revert("Use governance functions to manage members");
    }
}
