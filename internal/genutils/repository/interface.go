package repository

import (
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// GenTxRepository defines the interface for GenTx persistence.
// It follows the Repository pattern from DDD, abstracting the
// underlying storage mechanism and providing domain-centric operations.
type GenTxRepository interface {
	// Save persists a GenTx to the repository.
	// If a GenTx with the same validator address already exists, it returns an error.
	// Returns ErrDuplicateValidatorAddress if the validator address already exists.
	Save(gentx domain.GenTx) error

	// FindAll retrieves all GenTxs from the repository.
	// Returns an empty slice if no GenTxs are found.
	// The order is not guaranteed unless sorted by the caller.
	FindAll() ([]domain.GenTx, error)

	// FindByValidator retrieves a GenTx by validator address.
	// Returns ErrGenTxNotFound if the GenTx does not exist.
	FindByValidator(validatorAddress domain.Address) (domain.GenTx, error)

	// Exists checks if a GenTx with the given validator address exists.
	// Returns true if the GenTx exists, false otherwise.
	Exists(validatorAddress domain.Address) (bool, error)

	// Delete removes a GenTx from the repository by validator address.
	// Returns ErrGenTxNotFound if the GenTx does not exist.
	// This operation should be used with caution in production.
	Delete(validatorAddress domain.Address) error

	// Count returns the total number of GenTxs in the repository.
	Count() (int, error)
}
