package domain

import (
	"sort"
	"strings"
)

// GenTxCollection represents a collection of genesis transactions.
// It enforces uniqueness constraints and maintains collection invariants.
type GenTxCollection struct {
	gentxs             []GenTx
	validatorAddresses map[string]bool // validator address hex -> exists
	operatorAddresses  map[string]bool // operator address hex -> exists
	blsPublicKeys      map[string]bool // BLS key hex -> exists
}

// NewGenTxCollection creates a new empty GenTxCollection.
func NewGenTxCollection() *GenTxCollection {
	return &GenTxCollection{
		gentxs:             make([]GenTx, 0),
		validatorAddresses: make(map[string]bool),
		operatorAddresses:  make(map[string]bool),
		blsPublicKeys:      make(map[string]bool),
	}
}

// Add adds a GenTx to the collection.
// Returns error if duplicate validator address, operator address, or BLS key is detected.
func (c *GenTxCollection) Add(gentx GenTx) error {
	// Check for duplicate validator address
	validatorHex := strings.ToLower(gentx.ValidatorAddress().String())
	if c.validatorAddresses[validatorHex] {
		return ErrDuplicateValidatorAddress
	}

	// Check for duplicate operator address
	operatorHex := strings.ToLower(gentx.OperatorAddress().String())
	if c.operatorAddresses[operatorHex] {
		return ErrDuplicateOperatorAddress
	}

	// Check for duplicate BLS public key
	blsHex := strings.ToLower(gentx.BLSPublicKey().String())
	if c.blsPublicKeys[blsHex] {
		return ErrDuplicateBLSKey
	}

	// Add to collection
	c.gentxs = append(c.gentxs, gentx)
	c.validatorAddresses[validatorHex] = true
	c.operatorAddresses[operatorHex] = true
	c.blsPublicKeys[blsHex] = true

	return nil
}

// Remove removes a GenTx from the collection by validator address.
// Returns error if the GenTx is not found.
func (c *GenTxCollection) Remove(validatorAddr Address) error {
	validatorHex := strings.ToLower(validatorAddr.String())

	// Find the GenTx
	index := -1
	var targetGenTx GenTx
	for i, gentx := range c.gentxs {
		if strings.ToLower(gentx.ValidatorAddress().String()) == validatorHex {
			index = i
			targetGenTx = gentx
			break
		}
	}

	if index == -1 {
		return ErrGenTxNotFound
	}

	// Remove from slice
	c.gentxs = append(c.gentxs[:index], c.gentxs[index+1:]...)

	// Remove from maps
	delete(c.validatorAddresses, validatorHex)
	delete(c.operatorAddresses, strings.ToLower(targetGenTx.OperatorAddress().String()))
	delete(c.blsPublicKeys, strings.ToLower(targetGenTx.BLSPublicKey().String()))

	return nil
}

// ContainsValidator checks if the collection contains a GenTx with the given validator address.
func (c *GenTxCollection) ContainsValidator(validatorAddr Address) bool {
	validatorHex := strings.ToLower(validatorAddr.String())
	return c.validatorAddresses[validatorHex]
}

// ContainsOperator checks if the collection contains a GenTx with the given operator address.
func (c *GenTxCollection) ContainsOperator(operatorAddr Address) bool {
	operatorHex := strings.ToLower(operatorAddr.String())
	return c.operatorAddresses[operatorHex]
}

// Size returns the number of GenTxs in the collection.
func (c *GenTxCollection) Size() int {
	return len(c.gentxs)
}

// IsEmpty returns true if the collection is empty.
func (c *GenTxCollection) IsEmpty() bool {
	return len(c.gentxs) == 0
}

// Clear removes all GenTxs from the collection.
func (c *GenTxCollection) Clear() {
	c.gentxs = make([]GenTx, 0)
	c.validatorAddresses = make(map[string]bool)
	c.operatorAddresses = make(map[string]bool)
	c.blsPublicKeys = make(map[string]bool)
}

// GetAll returns a copy of all GenTxs in the collection.
// The returned slice is a defensive copy to prevent external modification.
func (c *GenTxCollection) GetAll() []GenTx {
	// Return defensive copy
	result := make([]GenTx, len(c.gentxs))
	copy(result, c.gentxs)
	return result
}

// GetSorted returns all GenTxs sorted by validator address (lexicographically).
// This ensures deterministic ordering for consensus operations.
func (c *GenTxCollection) GetSorted() []GenTx {
	// Create defensive copy
	result := make([]GenTx, len(c.gentxs))
	copy(result, c.gentxs)

	// Sort by validator address (case-insensitive lexicographic order)
	sort.Slice(result, func(i, j int) bool {
		addrI := strings.ToLower(result[i].ValidatorAddress().String())
		addrJ := strings.ToLower(result[j].ValidatorAddress().String())
		return addrI < addrJ
	})

	return result
}
