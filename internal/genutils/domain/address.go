package domain

import (
	"github.com/ethereum/go-ethereum/common"
)

// Address represents an Ethereum address value object.
// It is immutable and ensures valid Ethereum address format.
type Address struct {
	value common.Address
}

// NewAddress creates a new Address from a hex string.
// Returns ErrInvalidAddress if the hex string is not a valid Ethereum address.
// The hex string must start with "0x" prefix.
func NewAddress(hex string) (Address, error) {
	// Check for 0x prefix
	if len(hex) < 2 || hex[:2] != "0x" {
		return Address{}, ErrInvalidAddress
	}

	// Validate as Ethereum address
	if !common.IsHexAddress(hex) {
		return Address{}, ErrInvalidAddress
	}

	return Address{value: common.HexToAddress(hex)}, nil
}

// String returns the checksummed hex representation of the address.
func (a Address) String() string {
	return a.value.Hex()
}

// Bytes returns the address as a 20-byte array.
// Returns a copy to maintain immutability.
func (a Address) Bytes() []byte {
	bytes := a.value.Bytes()
	// Return a copy to prevent external modification
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result
}

// Equals compares two addresses for equality.
// Comparison is case-insensitive as addresses are normalized.
func (a Address) Equals(other Address) bool {
	return a.value == other.value
}

// IsZero returns true if the address is the zero address (0x0000...0000).
func (a Address) IsZero() bool {
	return a.value == common.Address{}
}
