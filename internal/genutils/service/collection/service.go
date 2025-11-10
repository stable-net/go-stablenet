package collection

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// Repository defines the interface for GenTx storage operations.
type Repository interface {
	// FindAll retrieves all GenTx objects from storage.
	FindAll() ([]domain.GenTx, error)
}

// Validator defines the interface for GenTx validation.
type Validator interface {
	// Validate performs comprehensive validation on a GenTx.
	Validate(gentx domain.GenTx) error
}

// CollectionService provides functionality for collecting and organizing GenTx files.
//
// The service is responsible for:
//   - Collecting GenTx files from directories
//   - Validating each GenTx
//   - Deduplicating by validator address
//   - Sorting deterministically by validator address
//
// CollectionService is safe for concurrent use by multiple goroutines.
type CollectionService struct {
	repository Repository
	validator  Validator
}

// NewCollectionService creates a new CollectionService instance.
//
// Parameters:
//   - repository: Repository for loading GenTx files
//   - validator: Validator for validating GenTx integrity
//
// Returns a new CollectionService that can collect and organize GenTx files.
func NewCollectionService(repository Repository, validator Validator) *CollectionService {
	return &CollectionService{
		repository: repository,
		validator:  validator,
	}
}

// CollectFromDirectory collects all GenTx files from the specified directory.
//
// The method performs the following operations:
//  1. Validates directory exists and is accessible
//  2. Loads all GenTx files from the directory using the repository
//  3. Validates each GenTx using the validator
//  4. Creates a GenTxCollection (automatically deduplicated and sorted)
//
// Parameters:
//   - dirPath: Path to directory containing GenTx files
//
// Returns:
//   - GenTxCollection containing all valid, unique GenTxs sorted by validator address
//   - Error if directory doesn't exist, validation fails, or collection creation fails
//
// Example:
//
//	service := NewCollectionService(repo, validator)
//	collection, err := service.CollectFromDirectory("/path/to/gentxs")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Collected %d GenTxs\n", collection.Count())
func (s *CollectionService) CollectFromDirectory(dirPath string) (*domain.GenTxCollection, error) {
	// Validate directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", dirPath)
		}
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Load all GenTx files from the repository
	gentxs, err := s.repository.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load GenTx files: %w", err)
	}

	// Create GenTxCollection
	collection := domain.NewGenTxCollection()

	// Validate and add each GenTx to the collection
	for _, gentx := range gentxs {
		if err := s.validator.Validate(gentx); err != nil {
			return nil, fmt.Errorf("validation failed for GenTx %s: %w",
				gentx.ValidatorAddress().String(), err)
		}

		if err := collection.Add(gentx); err != nil {
			return nil, fmt.Errorf("failed to add GenTx to collection: %w", err)
		}
	}

	return collection, nil
}
