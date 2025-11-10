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

import { GovBase } from "../abstracts/GovBase.sol";
import { AddressSetLib } from "../libraries/AddressSetLib.sol";

/**
 * @title GovCouncil
 * @notice Governance contract for managing blacklist and authorized account lists
 * @dev Inherits from GovBase for proposal-based governance mechanism
 *
 * ## Features
 * - Blacklist management with governance approval
 * - Authorized account list management
 * - O(1) current state queries
 * - Pagination support for large lists
 * - Batch operations for gas efficiency
 *
 * ## Security
 * - All changes require governance proposal + approval
 * - Zero address protection
 * - Duplicate prevention
 * - Reentrancy protection from GovBase
 * - Comprehensive event logging
 *
 * @custom:security-contact security@stablenet.io
 */
contract GovCouncil is GovBase {
    using AddressSetLib for AddressSetLib.AddressSet;

    // ========== Storage ==========

    // Current active lists (O(1) operations)
    AddressSetLib.AddressSet private _currentBlacklist;
    AddressSetLib.AddressSet private _currentAuthorizedAccounts;

    // ========== Action Types ==========

    // Action types for single operations
    bytes32 public constant ACTION_ADD_BLACKLIST = keccak256("ADD_BLACKLIST");
    bytes32 public constant ACTION_REMOVE_BLACKLIST = keccak256("REMOVE_BLACKLIST");
    bytes32 public constant ACTION_ADD_AUTHORIZED_ACCOUNT = keccak256("ADD_AUTHORIZED_ACCOUNT");
    bytes32 public constant ACTION_REMOVE_AUTHORIZED_ACCOUNT = keccak256("REMOVE_AUTHORIZED_ACCOUNT");

    // Action types for batch operations
    bytes32 public constant ACTION_ADD_BLACKLIST_BATCH = keccak256("ADD_BLACKLIST_BATCH");
    bytes32 public constant ACTION_REMOVE_BLACKLIST_BATCH = keccak256("REMOVE_BLACKLIST_BATCH");
    bytes32 public constant ACTION_ADD_AUTHORIZED_ACCOUNT_BATCH = keccak256("ADD_AUTHORIZED_ACCOUNT_BATCH");
    bytes32 public constant ACTION_REMOVE_AUTHORIZED_ACCOUNT_BATCH = keccak256("REMOVE_AUTHORIZED_ACCOUNT_BATCH");

    // ========== Events ==========

    /// @notice Emitted when address is added to blacklist
    /// @param account The blacklisted address
    /// @param proposalId The proposal that caused this change
    event AddressBlacklisted(address indexed account, uint256 indexed proposalId);

    /// @notice Emitted when address is removed from blacklist
    /// @param account The address removed from blacklist
    /// @param proposalId The proposal that caused this change
    event AddressUnblacklisted(address indexed account, uint256 indexed proposalId);

    /// @notice Emitted when address is added to authorized account list
    /// @param account The authorized account address
    /// @param proposalId The proposal that caused this change
    event AuthorizedAccountAdded(address indexed account, uint256 indexed proposalId);

    /// @notice Emitted when address is removed from authorized account list
    /// @param account The authorized account address
    /// @param proposalId The proposal that caused this change
    event AuthorizedAccountRemoved(address indexed account, uint256 indexed proposalId);

    // ========== Custom Errors ==========

    error AlreadyInBlacklist();
    error NotInBlacklist();
    error AlreadyInAuthorizedAccountList();
    error NotInAuthorizedAccountList();

    // ========================================
    // BLACKLIST: Proposal Functions
    // ========================================

    /**
     * @notice Propose to add an address to blacklist
     * @dev Creates governance proposal. Requires approval to execute.
     *
     * Security checks:
     * - Zero address validation
     * - Duplicate prevention
     * - Member-only access (via GovBase)
     *
     * @param account Address to blacklist
     * @return proposalId The created proposal ID
     */
    function proposeAddBlacklist(address account)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Security: Validate address
        if (account == address(0)) {
            revert AddressSetLib.ZeroAddressNotAllowed();
        }

        // Security: Prevent duplicates
        if (_currentBlacklist.contains(account)) {
            revert AlreadyInBlacklist();
        }

        return _createProposal(ACTION_ADD_BLACKLIST, abi.encode(account));
    }

    /**
     * @notice Propose to remove an address from blacklist
     * @dev Creates governance proposal. Requires approval to execute.
     *
     * Security checks:
     * - Existence validation
     * - Member-only access (via GovBase)
     *
     * @param account Address to remove from blacklist
     * @return proposalId The created proposal ID
     */
    function proposeRemoveBlacklist(address account)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Security: Validate existence
        if (!_currentBlacklist.contains(account)) {
            revert NotInBlacklist();
        }

        return _createProposal(ACTION_REMOVE_BLACKLIST, abi.encode(account));
    }

    /**
     * @notice Propose to add multiple addresses to blacklist (batch operation)
     * @dev More gas-efficient than multiple single proposals
     *
     * @param accounts Array of addresses to blacklist
     * @return proposalId The created proposal ID
     */
    function proposeAddBlacklistBatch(address[] calldata accounts)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Validate all addresses first
        for (uint256 i = 0; i < accounts.length; i++) {
            if (accounts[i] == address(0)) {
                revert AddressSetLib.ZeroAddressNotAllowed();
            }
            if (_currentBlacklist.contains(accounts[i])) {
                revert AlreadyInBlacklist();
            }
        }

        return _createProposal(ACTION_ADD_BLACKLIST_BATCH, abi.encode(accounts));
    }

    /**
     * @notice Propose to remove multiple addresses from blacklist (batch operation)
     * @param accounts Array of addresses to remove
     * @return proposalId The created proposal ID
     */
    function proposeRemoveBlacklistBatch(address[] calldata accounts)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Validate all addresses exist
        for (uint256 i = 0; i < accounts.length; i++) {
            if (!_currentBlacklist.contains(accounts[i])) {
                revert NotInBlacklist();
            }
        }

        return _createProposal(ACTION_REMOVE_BLACKLIST_BATCH, abi.encode(accounts));
    }

    // ========================================
    // AUTHORIZED ACCOUNT: Proposal Functions
    // ========================================

    /**
     * @notice Propose to add an address to authorized account list
     * @param account Address to add
     * @return proposalId The created proposal ID
     */
    function proposeAddAuthorizedAccount(address account)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        if (account == address(0)) {
            revert AddressSetLib.ZeroAddressNotAllowed();
        }

        if (_currentAuthorizedAccounts.contains(account)) {
            revert AlreadyInAuthorizedAccountList();
        }

        return _createProposal(ACTION_ADD_AUTHORIZED_ACCOUNT, abi.encode(account));
    }

    /**
     * @notice Propose to remove an address from authorized account list
     * @param account Address to remove
     * @return proposalId The created proposal ID
     */
    function proposeRemoveAuthorizedAccount(address account)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        if (!_currentAuthorizedAccounts.contains(account)) {
            revert NotInAuthorizedAccountList();
        }

        return _createProposal(ACTION_REMOVE_AUTHORIZED_ACCOUNT, abi.encode(account));
    }

    /**
     * @notice Propose to add multiple addresses to authorized account list (batch operation)
     * @param accounts Array of addresses to add
     * @return proposalId The created proposal ID
     */
    function proposeAddAuthorizedAccountBatch(address[] calldata accounts)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Validate all addresses first
        for (uint256 i = 0; i < accounts.length; i++) {
            if (accounts[i] == address(0)) {
                revert AddressSetLib.ZeroAddressNotAllowed();
            }
            if (_currentAuthorizedAccounts.contains(accounts[i])) {
                revert AlreadyInAuthorizedAccountList();
            }
        }

        return _createProposal(ACTION_ADD_AUTHORIZED_ACCOUNT_BATCH, abi.encode(accounts));
    }

    /**
     * @notice Propose to remove multiple addresses from authorized account list (batch operation)
     * @param accounts Array of addresses to remove
     * @return proposalId The created proposal ID
     */
    function proposeRemoveAuthorizedAccountBatch(address[] calldata accounts)
        external
        onlyActiveMember
        returns (uint256 proposalId)
    {
        // Validate all addresses exist
        for (uint256 i = 0; i < accounts.length; i++) {
            if (!_currentAuthorizedAccounts.contains(accounts[i])) {
                revert NotInAuthorizedAccountList();
            }
        }

        return _createProposal(ACTION_REMOVE_AUTHORIZED_ACCOUNT_BATCH, abi.encode(accounts));
    }

    // ========================================
    // Internal Execution (Called by GovBase)
    // ========================================

    /**
     * @notice Execute approved proposal actions
     * @dev Called by GovBase when proposal is approved
     *
     * @param actionType The type of action to execute
     * @param callData Encoded action parameters
     * @return success True if execution succeeded
     */
    function _executeCustomAction(bytes32 actionType, bytes memory callData)
        internal
        override
        returns (bool success)
    {
        // Blacklist operations
        if (actionType == ACTION_ADD_BLACKLIST) {
            address account = abi.decode(callData, (address));
            return _addToBlacklist(account, currentProposalId);
        } else if (actionType == ACTION_ADD_BLACKLIST_BATCH) {
            address[] memory accounts = abi.decode(callData, (address[]));
            for (uint256 i = 0; i < accounts.length; i++) {
                _addToBlacklist(accounts[i], currentProposalId);
            }
            return true;
        } else if (actionType == ACTION_REMOVE_BLACKLIST) {
            address account = abi.decode(callData, (address));
            return _removeFromBlacklist(account, currentProposalId);
        } else if (actionType == ACTION_REMOVE_BLACKLIST_BATCH) {
            address[] memory accounts = abi.decode(callData, (address[]));
            for (uint256 i = 0; i < accounts.length; i++) {
                _removeFromBlacklist(accounts[i], currentProposalId);
            }
            return true;
        }
        // Authorized account operations
        else if (actionType == ACTION_ADD_AUTHORIZED_ACCOUNT) {
            address account = abi.decode(callData, (address));
            return _addToAuthorizedAccount(account, currentProposalId);
        } else if (actionType == ACTION_ADD_AUTHORIZED_ACCOUNT_BATCH) {
            address[] memory accounts = abi.decode(callData, (address[]));
            for (uint256 i = 0; i < accounts.length; i++) {
                _addToAuthorizedAccount(accounts[i], currentProposalId);
            }
            return true;
        } else if (actionType == ACTION_REMOVE_AUTHORIZED_ACCOUNT) {
            address account = abi.decode(callData, (address));
            return _removeFromAuthorizedAccount(account, currentProposalId);
        } else if (actionType == ACTION_REMOVE_AUTHORIZED_ACCOUNT_BATCH) {
            address[] memory accounts = abi.decode(callData, (address[]));
            for (uint256 i = 0; i < accounts.length; i++) {
                _removeFromAuthorizedAccount(accounts[i], currentProposalId);
            }
            return true;
        }

        return false;
    }

    // ========================================
    // Core Blacklist Operations
    // ========================================

    /**
     * @notice Add address to blacklist (internal)
     * @dev Updates current state and emits event
     *
     * @param account Address to blacklist
     * @param proposalId Proposal ID that triggered this
     * @return success True if successful
     */
    function _addToBlacklist(address account, uint256 proposalId) private returns (bool success) {
        // Add to current set
        _currentBlacklist.add(account);

        // Emit event
        emit AddressBlacklisted(account, proposalId);

        return true;
    }

    /**
     * @notice Remove address from blacklist (internal)
     * @dev Updates current state and emits event
     *
     * @param account Address to remove
     * @param proposalId Proposal ID that triggered this
     * @return success True if successful
     */
    function _removeFromBlacklist(address account, uint256 proposalId) private returns (bool success) {
        // Remove from current set
        _currentBlacklist.remove(account);

        // Emit event
        emit AddressUnblacklisted(account, proposalId);

        return true;
    }

    // ========================================
    // Core Authorized Account Operations
    // ========================================

    /**
     * @notice Add address to authorized account list (internal)
     */
    function _addToAuthorizedAccount(address account, uint256 proposalId) private returns (bool success) {
        _currentAuthorizedAccounts.add(account);

        emit AuthorizedAccountAdded(account, proposalId);

        return true;
    }

    /**
     * @notice Remove address from authorized account list (internal)
     */
    function _removeFromAuthorizedAccount(address account, uint256 proposalId) private returns (bool success) {
        _currentAuthorizedAccounts.remove(account);

        emit AuthorizedAccountRemoved(account, proposalId);

        return true;
    }

    // ========================================
    // BLACKLIST: Query Functions (O(1))
    // ========================================

    /**
     * @notice Check if address is currently blacklisted
     * @param account Address to check
     * @return isBlacklisted True if currently in blacklist
     */
    function isBlacklisted(address account) external view returns (bool) {
        return _currentBlacklist.contains(account);
    }

    /**
     * @notice Get total number of currently blacklisted addresses
     * @return count Number of addresses
     */
    function getBlacklistCount() external view returns (uint256) {
        return _currentBlacklist.length();
    }

    /**
     * @notice Get blacklisted address at specific index
     * @dev Use with getBlacklistCount() for iteration
     * @param index Index to query (0-based)
     * @return account Address at index
     */
    function getBlacklistedAddress(uint256 index) external view returns (address) {
        return _currentBlacklist.at(index);
    }

    /**
     * @notice Get range of blacklisted addresses (pagination)
     * @dev Recommended for large lists. O(n) where n = end - start + 1
     * @param startIndex Start index (inclusive)
     * @param endIndex End index (inclusive)
     * @return addresses Array of addresses in range
     */
    function getBlacklistRange(uint256 startIndex, uint256 endIndex)
        external
        view
        returns (address[] memory)
    {
        return _currentBlacklist.valuesInRange(startIndex, endIndex);
    }

    /**
     * @notice Get all currently blacklisted addresses
     * @dev WARNING: Can be expensive for large lists. Use getBlacklistRange() for pagination.
     * @return addresses Array of all blacklisted addresses
     */
    function getAllBlacklisted() external view returns (address[] memory) {
        return _currentBlacklist.values();
    }

    // ========================================
    // AUTHORIZED ACCOUNT: Query Functions
    // ========================================

    /**
     * @notice Check if address is currently in authorized account list
     */
    function isAuthorizedAccount(address account) external view returns (bool) {
        return _currentAuthorizedAccounts.contains(account);
    }

    /**
     * @notice Get total number of authorized accounts
     */
    function getAuthorizedAccountCount() external view returns (uint256) {
        return _currentAuthorizedAccounts.length();
    }

    /**
     * @notice Get authorized account address at specific index
     */
    function getAuthorizedAccountAddress(uint256 index) external view returns (address) {
        return _currentAuthorizedAccounts.at(index);
    }

    /**
     * @notice Get range of authorized account addresses (pagination)
     */
    function getAuthorizedAccountRange(uint256 startIndex, uint256 endIndex)
        external
        view
        returns (address[] memory)
    {
        return _currentAuthorizedAccounts.valuesInRange(startIndex, endIndex);
    }

    /**
     * @notice Get all authorized account addresses
     */
    function getAllAuthorizedAccounts() external view returns (address[] memory) {
        return _currentAuthorizedAccounts.values();
    }
}
