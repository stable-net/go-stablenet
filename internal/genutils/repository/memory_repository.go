package repository

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// MemoryRepository implements GenTxRepository using in-memory storage.
// It is intended for testing purposes and provides thread-safe operations.
// All data is lost when the program terminates.
type MemoryRepository struct {
	mu     sync.RWMutex
	gentxs map[string]domain.GenTx
}

// NewMemoryRepository creates a new MemoryRepository instance.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		gentxs: make(map[string]domain.GenTx),
	}
}

// Save persists a GenTx to memory.
// Returns ErrDuplicateValidatorAddress if the validator address already exists.
// This method is thread-safe.
func (r *MemoryRepository) Save(gentx domain.GenTx) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate key from validator address (lowercase for case-insensitive comparison)
	key := r.genKey(gentx.ValidatorAddress())

	// Check if already exists
	if _, exists := r.gentxs[key]; exists {
		return domain.ErrDuplicateValidatorAddress
	}

	// Store gentx
	r.gentxs[key] = gentx
	return nil
}

// FindAll retrieves all GenTxs from memory.
// Returns an empty slice if no GenTxs are found.
// This method is thread-safe.
func (r *MemoryRepository) FindAll() ([]domain.GenTx, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create slice from map values
	gentxs := make([]domain.GenTx, 0, len(r.gentxs))
	for _, gentx := range r.gentxs {
		gentxs = append(gentxs, gentx)
	}

	return gentxs, nil
}

// FindByValidator retrieves a GenTx by validator address.
// Returns ErrGenTxNotFound if the GenTx does not exist.
// This method is thread-safe.
func (r *MemoryRepository) FindByValidator(validatorAddress domain.Address) (domain.GenTx, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.genKey(validatorAddress)

	gentx, exists := r.gentxs[key]
	if !exists {
		return domain.GenTx{}, domain.ErrGenTxNotFound
	}

	return gentx, nil
}

// Exists checks if a GenTx with the given validator address exists.
// This method is thread-safe.
func (r *MemoryRepository) Exists(validatorAddress domain.Address) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.genKey(validatorAddress)
	_, exists := r.gentxs[key]
	return exists, nil
}

// Delete removes a GenTx from memory by validator address.
// Returns ErrGenTxNotFound if the GenTx does not exist.
// This method is thread-safe.
func (r *MemoryRepository) Delete(validatorAddress domain.Address) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.genKey(validatorAddress)

	if _, exists := r.gentxs[key]; !exists {
		return domain.ErrGenTxNotFound
	}

	delete(r.gentxs, key)
	return nil
}

// Count returns the total number of GenTxs in memory.
// This method is thread-safe.
func (r *MemoryRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.gentxs), nil
}

// genKey generates a storage key from a validator address.
// Uses lowercase for case-insensitive comparison.
func (r *MemoryRepository) genKey(address domain.Address) string {
	return strings.ToLower(address.String())
}
