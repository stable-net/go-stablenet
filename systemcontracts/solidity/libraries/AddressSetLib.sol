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

/**
 * @title AddressSetLib
 * @notice Internalized implementation of address set with O(1) operations
 * @dev Based on OpenZeppelin's EnumerableSet but simplified and optimized for address-only usage
 *
 * ## Core Design
 * - Uses array + mapping for O(1) add/remove/contains
 * - 1-based position indexing: position 0 means "not in set"
 * - "Swap and pop" removal: O(1) but order not preserved
 *
 * ## Security Features
 * - No external dependencies
 * - Explicit zero address handling
 * - Overflow protection with uint256
 * - Duplicate prevention
 *
 * @custom:security-contact security@stablenet.io
 */
library AddressSetLib {
    /**
     * @dev Internal set structure
     *
     * Storage layout:
     * - _values: Dynamic array of addresses (enumerable)
     * - _positions: Mapping of address to 1-based position (0 = not exists)
     *
     * Position is 1-based to distinguish between:
     * - 0: Address not in set
     * - 1+: Array index + 1 (actual position)
     */
    struct AddressSet {
        address[] _values;
        mapping(address => uint256) _positions;
    }

    // ========== Errors ==========

    /// @notice Attempted to add zero address to set
    error ZeroAddressNotAllowed();

    /// @notice Attempted to add address that already exists in set
    error AddressAlreadyExists();

    /// @notice Attempted to remove address that doesn't exist in set
    error AddressNotFound();

    /// @notice Index out of bounds for array access
    error IndexOutOfBounds();

    // ========== Core Operations ==========

    /**
     * @notice Add an address to the set
     * @dev O(1) operation. Reverts if address already exists or is zero address.
     *
     * Security checks:
     * - Zero address validation
     * - Duplicate prevention
     *
     * @param set The set to add to
     * @param value The address to add
     * @return success True if added successfully
     */
    function add(AddressSet storage set, address value) internal returns (bool success) {
        // Security: Prevent zero address
        if (value == address(0)) {
            revert ZeroAddressNotAllowed();
        }

        // Security: Prevent duplicates
        if (contains(set, value)) {
            revert AddressAlreadyExists();
        }

        // Add to array
        set._values.push(value);

        // Store 1-based position (array.length at this point)
        set._positions[value] = set._values.length;

        return true;
    }

    /**
     * @notice Remove an address from the set
     * @dev O(1) operation using "swap and pop" technique.
     * The order of elements is NOT preserved after removal.
     *
     * Algorithm:
     * 1. Get position of value to remove (1-based)
     * 2. Get the last element in array
     * 3. Swap: Move last element to position of removed element
     * 4. Update: Update position mapping for moved element
     * 5. Pop: Remove last element from array
     * 6. Delete: Remove position mapping for removed element
     *
     * @param set The set to remove from
     * @param value The address to remove
     * @return success True if removed successfully
     */
    function remove(AddressSet storage set, address value) internal returns (bool success) {
        // Security: Check existence
        uint256 position = set._positions[value];
        if (position == 0) {
            revert AddressNotFound();
        }

        // Convert to 0-based index
        uint256 valueIndex = position - 1;
        uint256 lastIndex = set._values.length - 1;

        // If not the last element, swap with last
        if (valueIndex != lastIndex) {
            address lastValue = set._values[lastIndex];

            // Move last element to the position of removed element
            set._values[valueIndex] = lastValue;

            // Update position of moved element
            set._positions[lastValue] = position; // Keep 1-based
        }

        // Remove last element
        set._values.pop();

        // Delete position mapping
        delete set._positions[value];

        return true;
    }

    /**
     * @notice Check if an address exists in the set
     * @dev O(1) operation. Returns false for zero address.
     *
     * @param set The set to check
     * @param value The address to check
     * @return exists True if address is in set
     */
    function contains(AddressSet storage set, address value) internal view returns (bool exists) {
        // Zero address is never in set
        if (value == address(0)) {
            return false;
        }

        return set._positions[value] != 0;
    }

    /**
     * @notice Get the number of addresses in the set
     * @dev O(1) operation
     *
     * @param set The set to query
     * @return count Number of addresses
     */
    function length(AddressSet storage set) internal view returns (uint256 count) {
        return set._values.length;
    }

    /**
     * @notice Get address at specific index
     * @dev O(1) operation. Reverts if index out of bounds.
     *
     * Security: Index validation to prevent out-of-bounds access
     *
     * @param set The set to query
     * @param index The index to retrieve (0-based)
     * @return value The address at given index
     */
    function at(AddressSet storage set, uint256 index) internal view returns (address value) {
        if (index >= set._values.length) {
            revert IndexOutOfBounds();
        }

        return set._values[index];
    }

    /**
     * @notice Get all addresses in the set
     * @dev O(n) operation. WARNING: Can be expensive for large sets.
     * Recommended for off-chain queries only. Use pagination for large sets.
     *
     * @param set The set to query
     * @return values Array of all addresses
     */
    function values(AddressSet storage set) internal view returns (address[] memory) {
        return set._values;
    }

    /**
     * @notice Get a range of addresses from the set
     * @dev O(n) operation where n = endIndex - startIndex + 1
     * Use this for pagination instead of values() for large sets.
     *
     * Security: Bounds checking on both start and end indices
     *
     * @param set The set to query
     * @param startIndex Start index (inclusive, 0-based)
     * @param endIndex End index (inclusive, 0-based)
     * @return range Array of addresses in specified range
     */
    function valuesInRange(
        AddressSet storage set,
        uint256 startIndex,
        uint256 endIndex
    ) internal view returns (address[] memory range) {
        // Security: Validate range
        if (endIndex < startIndex) {
            revert IndexOutOfBounds();
        }

        uint256 setLength = set._values.length;
        if (endIndex >= setLength) {
            revert IndexOutOfBounds();
        }

        // Calculate range size
        uint256 rangeSize = endIndex - startIndex + 1;
        range = new address[](rangeSize);

        // Copy values in range
        for (uint256 i = 0; i < rangeSize; i++) {
            range[i] = set._values[startIndex + i];
        }

        return range;
    }

    /**
     * @notice Clear all addresses from the set
     * @dev O(n) operation. Use with caution for large sets.
     *
     * Note: This is a destructive operation. Consider if you need this
     * or if removing specific addresses is more appropriate.
     *
     * @param set The set to clear
     */
    function clear(AddressSet storage set) internal {
        // Delete all position mappings
        for (uint256 i = 0; i < set._values.length; i++) {
            delete set._positions[set._values[i]];
        }

        // Clear array
        delete set._values;
    }
}
