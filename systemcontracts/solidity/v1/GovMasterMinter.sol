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

import { GovBaseV2 } from "../abstracts/GovBaseV2.sol";
import { IFiatToken } from "../interfaces/IFiatToken.sol";

/**
 * @title GovMasterMinter
 * @dev Governance contract for managing FiatToken minter addresses
 * Supports both EOA and CA as members
 */
contract GovMasterMinter is GovBaseV2 {
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
    // Note: Member management actions (ADD_MEMBER, REMOVE_MEMBER, CHANGE_MEMBER) are inherited from GovBase
    bytes32 public constant ACTION_CONFIGURE_MINTER = keccak256("CONFIGURE_MINTER");
    bytes32 public constant ACTION_REMOVE_MINTER = keccak256("REMOVE_MINTER");
    bytes32 public constant ACTION_UPDATE_MAX_MINTER_ALLOWANCE = keccak256("UPDATE_MAX_MINTER_ALLOWANCE");
    bytes32 public constant ACTION_PAUSE = keccak256("PAUSE");
    bytes32 public constant ACTION_UNPAUSE = keccak256("UNPAUSE");

    // ========== State Variables ==========
    // GovBaseV2 uses slots 0x0-0xb, __gap uses 0xc-0x31

    // Slot 0x32: fiatToken (address, 20 bytes)
    IFiatToken public fiatToken; // The FiatToken contract

    // Slot 0x33: emergencyPaused (bool, 1 byte)
    bool public emergencyPaused; // Emergency pause state

    // Slot 0x34: maxMinterAllowance (uint256, 32 bytes)
    // Note: Default value 10B tokens (10000000000 * 10^18 = 10000000000000000000000000000 wei)
    // Genesis state should override this via initializeMasterMinter() but due to simulated backend limitation,
    // this fallback value is used
    uint256 public maxMinterAllowance = 10000000000000000000000000000;

    // Slot 0x35: isMinter mapping (base slot)
    mapping(address => bool) public isMinter; // Minter status

    // Slot 0x36: minterList array (length slot, elements stored at keccak256(0x36))
    address[] private __minterList; // List of all minters

    // Slot 0x37: minterIndex mapping (base slot)
    mapping(address => uint256) private __minterIndex; // Index in minterList + 1 (1-based indexing)

    // ========== Events ==========

    /// @notice Emitted when a minter is configured with a new allowance
    event MinterConfigured(address indexed minter, uint256 allowance);

    /// @notice Emitted when a minter is removed
    event MinterRemoved(address indexed minter);

    /// @notice Emitted when the maximum minter allowance limit is updated
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

    // ========== Proposal Functions ==========

    /**
     * @notice Propose to configure a minter (add or update allowance)
     * @param minter Address of the minter contract
     * @param minterAllowedAmount Allowance amount for the minter (in token base units)
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members when contract is not paused
     *      Creates proposal requiring quorum approval before execution
     *      Validation: minter != address(0), 0 < allowance <= maxMinterAllowance
     */
    function proposeConfigureMinter(address minter, uint256 minterAllowedAmount) external onlyActiveMember whenNotPaused returns (uint256) {
        return _proposeConfigureMinter(minter, minterAllowedAmount);
    }

    /**
     * @notice Propose to remove a minter from FiatToken
     * @param minter Address of the minter to remove
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members when contract is not paused
     *      Creates proposal requiring quorum approval before execution
     *      Validation: minter != address(0), isMinter[minter] == true
     */
    function proposeRemoveMinter(address minter) external onlyActiveMember whenNotPaused returns (uint256) {
        return _proposeRemoveMinter(minter);
    }

    /**
     * @notice Propose to update the maximum allowance limit per minter
     * @param newLimit New maximum allowance amount (in token base units)
     * @return proposalId The ID of the created proposal
     * @dev Only callable by active governance members (can be proposed during pause)
     *      Creates proposal requiring quorum approval before execution
     *      Validation: newLimit > 0
     *      Note: Does NOT affect existing minters' allowances, only future proposals
     */
    function proposeUpdateMaxMinterAllowance(uint256 newLimit) external onlyActiveMember returns (uint256) {
        if (newLimit == 0) revert InvalidAllowance();

        bytes memory data = abi.encode(newLimit);
        return _createProposal(ACTION_UPDATE_MAX_MINTER_ALLOWANCE, data);
    }

    /**
     * @notice Propose emergency pause of all minter operations
     * @return proposalId The ID of the created proposal
     * @dev Only callable by governance members when not paused
     *      When executed, blocks all minter configuration and removal operations
     */
    function proposePause() external onlyActiveMember whenNotPaused returns (uint256) {
        bytes memory callData = "";
        return _createProposal(ACTION_PAUSE, callData);
    }

    /**
     * @notice Propose resumption of minter operations
     * @return proposalId The ID of the created proposal
     * @dev Only callable by governance members when paused
     *      When executed, resumes normal minter configuration and removal operations
     */
    function proposeUnpause() external onlyActiveMember whenPaused returns (uint256) {
        bytes memory callData = "";
        return _createProposal(ACTION_UNPAUSE, callData);
    }

    // Note: proposeAddMember, proposeRemoveMember, proposeChangeMember are inherited from GovBase

    // ========== Internal Proposal Helpers ==========

    /// @dev Internal helper for proposeConfigureMinter logic
    /// @param minter Address of the minter contract
    /// @param minterAllowedAmount Allowance amount for the minter
    /// @return proposalId The ID of the created proposal
    function _proposeConfigureMinter(address minter, uint256 minterAllowedAmount) internal returns (uint256) {
        // Validate minter address
        if (minter == address(0)) revert InvalidMinterAddress();

        // Validate allowance
        if (minterAllowedAmount == 0) revert InvalidAllowance();
        if (minterAllowedAmount > maxMinterAllowance) revert InvalidAllowance();

        bytes memory data = abi.encode(minter, minterAllowedAmount);
        return _createProposal(ACTION_CONFIGURE_MINTER, data);
    }

    /// @dev Internal helper for proposeRemoveMinter logic
    /// @param minter Address of the minter to remove
    /// @return proposalId The ID of the created proposal
    function _proposeRemoveMinter(address minter) internal returns (uint256) {
        if (minter == address(0)) revert InvalidMinterAddress();
        if (!isMinter[minter]) revert NotAMinter();

        bytes memory data = abi.encode(minter);
        return _createProposal(ACTION_REMOVE_MINTER, data);
    }

    // ========== IFiatToken Interface Implementation ==========

    // Note: isMinter(address) is provided by the public mapping's auto-generated getter

    /// @notice Get the minter allowance (IFiatToken interface)
    /// @param _minter Address of the minter
    /// @return Allowance amount
    /// @dev Delegates to FiatToken for current allowance
    function minterAllowance(address _minter) external view returns (uint256) {
        return fiatToken.minterAllowance(_minter);
    }

    // ========== Internal Action Implementation ==========

    /// @dev Override _executeCustomAction to implement GovMasterMinter-specific actions
    /// @param actionType The type of action to execute (one of the ACTION_* constants)
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if action executed successfully, false if unknown action
    /// @notice Called by GovBase after proposal approved and quorum reached
    ///         Protected by nonReentrant modifier in GovBase
    ///         Validates all parameters at execution time (defense in depth)
    ///         Reverts on invalid parameters or FiatToken call failures
    function _executeCustomAction(bytes32 actionType, bytes memory callData) internal override returns (bool) {
        if (actionType == ACTION_CONFIGURE_MINTER) {
            (address minter, uint256 allowance) = abi.decode(callData, (address, uint256));
            // Re-validate at execution time (defense in depth)
            if (minter == address(0)) revert InvalidMinterAddress();
            if (allowance == 0) revert InvalidAllowance();
            if (allowance > maxMinterAllowance) revert InvalidAllowance();
            return _safeConfigureMinter(minter, allowance);
        }

        if (actionType == ACTION_REMOVE_MINTER) {
            address minter = abi.decode(callData, (address));
            // Re-validate at execution time (defense in depth)
            if (minter == address(0)) revert InvalidMinterAddress();
            if (!isMinter[minter]) revert NotAMinter();
            return _safeRemoveMinter(minter);
        }

        if (actionType == ACTION_UPDATE_MAX_MINTER_ALLOWANCE) {
            uint256 newLimit = abi.decode(callData, (uint256));
            // Re-validate at execution time (defense in depth)
            if (newLimit == 0) revert InvalidAllowance();
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

    /// @dev Execute emergency pause action
    /// @return success Always returns true (pause cannot fail)
    function _executePause() internal returns (bool) {
        if (emergencyPaused) revert ContractPaused();
        emergencyPaused = true;
        emit EmergencyPaused(currentProposalId);
        return true;
    }

    /// @dev Execute emergency unpause action
    /// @return success Always returns true (unpause cannot fail)
    function _executeUnpause() internal returns (bool) {
        if (!emergencyPaused) revert ContractNotPaused();
        emergencyPaused = false;
        emit EmergencyUnpaused(currentProposalId);
        return true;
    }

    /// @dev Configure minter following CEI pattern (Check-Effect-Interaction)
    /// @param minter Address of the minter to configure
    /// @param allowance New allowance amount for the minter
    /// @return success Always returns true (reverts on failure)
    /// @notice Updates internal state before external call to prevent reentrancy
    ///         Protected by GovBase's nonReentrant modifier on _executeAction
    ///         Reverts with MinterConfigurationFailed if FiatToken call fails
    function _safeConfigureMinter(address minter, uint256 allowance) internal whenNotPaused returns (bool) {
        // Effect: Update internal state first (CEI pattern)
        // Add minter to tracking if not already present
        if (!isMinter[minter]) {
            __minterList.push(minter);
            __minterIndex[minter] = __minterList.length; // 1-based index
            isMinter[minter] = true;
        }
        // Note: allowance parameter used for FiatToken call, not stored locally

        // Interaction: Call external FiatToken
        try fiatToken.configureMinter(minter, allowance) {
            emit MinterConfigured(minter, allowance);
            return true;
        } catch {
            // Revert all state changes on failure
            revert MinterConfigurationFailed();
        }
    }

    /// @dev Remove minter following CEI pattern (Check-Effect-Interaction)
    /// @param minter Address of the minter to remove
    /// @return success Always returns true (reverts on failure)
    /// @notice Updates internal state before external call to prevent reentrancy
    ///         Protected by GovBase's nonReentrant modifier on _executeAction
    ///         Reverts with MinterRemovalFailed if FiatToken call fails
    function _safeRemoveMinter(address minter) internal whenNotPaused returns (bool) {
        // Effect: Update internal state first (CEI pattern)
        // Remove from array using swap-and-pop
        uint256 index = __minterIndex[minter];
        if (index > 0) {
            uint256 arrayIndex = index - 1;
            uint256 lastIndex = __minterList.length - 1;

            if (arrayIndex != lastIndex) {
                address lastMinter = __minterList[lastIndex];
                __minterList[arrayIndex] = lastMinter;
                __minterIndex[lastMinter] = index;
            }

            __minterList.pop();
            delete __minterIndex[minter];
        }

        isMinter[minter] = false;
        // Note: FiatToken.removeMinter() handles actual allowance removal

        // Interaction: Call external FiatToken
        try fiatToken.removeMinter(minter) {
            emit MinterRemoved(minter);
            return true;
        } catch {
            // Revert all state changes on failure
            revert MinterRemovalFailed();
        }
    }

    // ========== View Functions ==========

    /// @notice Query FiatToken to check if an address has actual minting authority
    /// @param account Address to check for minting authority
    /// @return status True if account has minting authority in FiatToken, false otherwise
    /// @dev Delegates to FiatToken.isMinter() to query the actual source of truth
    ///      This is the authoritative check for real minting permissions
    function getIsMinter(address account) external view returns (bool) {
        return fiatToken.isMinter(account);
    }

    /// @notice Get the allowance for a specific minter from FiatToken
    /// @param minter Address of the minter to query
    /// @return allowance Current allowance amount for the minter (0 if not a minter)
    /// @dev Queries FiatToken.minterAllowance() directly for current allowance
    ///      FiatToken is the single source of truth for allowance values
    function getMinterAllowance(address minter) public view returns (uint256) {
        return fiatToken.minterAllowance(minter);
    }

    /// @notice Check if an address is tracked as a minter in GovMasterMinter's local state
    /// @param minter Address to check
    /// @return status True if address is tracked in local isMinter mapping, false otherwise
    /// @dev Queries local isMinter mapping (GovMasterMinter tracking state only)
    ///      This does NOT check actual FiatToken minting authority
    function getMinterStatus(address minter) public view returns (bool) {
        return isMinter[minter];
    }

    /// @notice Get the total number of minters tracked in GovMasterMinter
    /// @return count Total number of minters in minterList array
    /// @dev Returns the length of minterList (local tracking state)
    ///      This count represents minters managed by this governance contract
    function getMinterCount() public view returns (uint256) {
        return __minterList.length;
    }

    /// @notice Get minter address and its current allowance at a specific index
    /// @param index Index in the minterList array (0-based indexing)
    /// @return minter Address of the minter at the specified index
    /// @return allowance Current allowance for the minter from FiatToken
    /// @dev Reverts with IndexOutOfBounds if index >= minterList.length
    ///      Array order is not stable due to swap-and-pop removal pattern
    ///      Queries FiatToken.minterAllowance() for current allowance value
    function getMinterAt(uint256 index) public view returns (address minter, uint256 allowance) {
        if (index >= __minterList.length) revert IndexOutOfBounds();
        minter = __minterList[index];
        allowance = fiatToken.minterAllowance(minter);
    }

    /// @notice Get all minters tracked in GovMasterMinter and their current allowances
    /// @return minters Array of all minter addresses from minterList
    /// @return allowances Array of current allowances from FiatToken (parallel to minters array)
    /// @dev Queries FiatToken.minterAllowance() for each minter's current allowance
    ///      Gas cost: O(n) where n = minterList.length
    ///      Performance: ~100K gas for 10 minters, ~1M gas for 100 minters
    function getAllMinters() public view returns (address[] memory minters, uint256[] memory allowances) {
        uint256 count = __minterList.length;
        minters = new address[](count);
        allowances = new uint256[](count);

        for (uint256 i = 0; i < count; i++) {
            minters[i] = __minterList[i];
            allowances[i] = fiatToken.minterAllowance(__minterList[i]);
        }
    }

    /// @notice Get aggregate statistics for all minters tracked in GovMasterMinter
    /// @return totalMinters Total number of minters in minterList
    /// @return totalAllowance Sum of all minter allowances from FiatToken
    /// @dev Queries FiatToken.minterAllowance() for each minter and calculates sum
    ///      Gas cost: O(n) where n = minterList.length, similar to getAllMinters()
    function getMinterStats() public view returns (uint256 totalMinters, uint256 totalAllowance) {
        totalMinters = __minterList.length;
        totalAllowance = 0;
        for (uint256 i = 0; i < totalMinters; i++) {
            totalAllowance += fiatToken.minterAllowance(__minterList[i]);
        }
    }

    /// @notice Check if a minter can mint a specific amount based on current allowance
    /// @param minter Address of the minter to check
    /// @param amount Amount to check (in token base units with decimals)
    /// @return canMintAmount True if minter is tracked locally AND amount <= FiatToken allowance
    /// @dev Checks local isMinter mapping first, then queries FiatToken.minterAllowance()
    ///      Returns false immediately if minter is not tracked in local state
    function canMint(address minter, uint256 amount) public view returns (bool) {
        if (!isMinter[minter]) return false;
        return amount <= fiatToken.minterAllowance(minter);
    }
}
