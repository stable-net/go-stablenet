// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of the go-stablenet library.
// Copyright 2025 The go-stablenet Authors
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package types

import "fmt"

// Extra field bit masks for StateAccount.
// The Extra field is a 64-bit value where each bit can represent a specific state or property.
//
// Bit positions (MSB to LSB):
//   - Bit 63 (0x8000000000000000): Blacklisted
//   - Bit 62 (0x4000000000000000): Authorized
//   - Bit 61 (0x2000000000000000): Reserved for future use
//   - ... (add more as needed)
const (
	// AccountExtraMaskBlacklisted defines the bit mask used to mark an account as blacklisted (bit 63, MSB).
	AccountExtraMaskBlacklisted uint64 = 1 << 63

	// AccountExtraMaskAuthorized defines the bit mask used to mark an account as authorized (bit 62).
	AccountExtraMaskAuthorized uint64 = 1 << 62

	// AccountExtraMaskC defines the bit mask used to mark ... (bit 61)
	// Uncomment and adjust as needed:
	// AccountExtraMaskC uint64 = 1 << 61

	// AccountExtraValidMask is the union of all defined bit masks.
	// When adding a new bit mask, it must also be included here;
	// otherwise ValidateExtra will reject accounts that use the new bit.
	AccountExtraValidMask uint64 = AccountExtraMaskBlacklisted | AccountExtraMaskAuthorized
)

// ============================================================================
// Generic utility functions (internal helpers)
// ============================================================================

// hasBit returns true if the specified bit mask is set in the extra value.
func hasBit(extra uint64, mask uint64) bool {
	return (extra & mask) != 0
}

// setBit sets the specified bit mask in the extra value and returns the result.
func setBit(extra uint64, mask uint64) uint64 {
	return extra | mask
}

// clearBit clears the specified bit mask in the extra value and returns the result.
func clearBit(extra uint64, mask uint64) uint64 {
	return extra &^ mask
}

// ============================================================================
// Specific bit functions (public API)
// ============================================================================

// IsBlacklisted returns true if the blacklisted bit is set in the extra value.
func IsBlacklisted(extra uint64) bool {
	return hasBit(extra, AccountExtraMaskBlacklisted)
}

// SetBlacklisted sets the blacklisted bit in the extra value and returns the result.
func SetBlacklisted(extra uint64) uint64 {
	return setBit(extra, AccountExtraMaskBlacklisted)
}

// ClearBlacklisted clears the blacklisted bit in the extra value and returns the result.
func ClearBlacklisted(extra uint64) uint64 {
	return clearBit(extra, AccountExtraMaskBlacklisted)
}

// IsAuthorized reports whether the authorized bit is set in extra.
func IsAuthorized(extra uint64) bool {
	return hasBit(extra, AccountExtraMaskAuthorized)
}

// SetAuthorized sets the authorized bit in extra and returns the new value.
func SetAuthorized(extra uint64) uint64 {
	return setBit(extra, AccountExtraMaskAuthorized)
}

// ClearAuthorized clears the authorized bit in extra and returns the new value.
func ClearAuthorized(extra uint64) uint64 {
	return clearBit(extra, AccountExtraMaskAuthorized)
}

// ValidateExtra checks that no undefined bits are set in the extra value.
// Returns an error if any bit outside AccountExtraValidMask is set.
func ValidateExtra(extra uint64) error {
	if extra&^AccountExtraValidMask != 0 {
		return fmt.Errorf("unknown bits set in account extra: 0x%016x", extra)
	}
	return nil
}
