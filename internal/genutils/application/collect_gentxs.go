package application

import (
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// CollectGenTxsRequest contains parameters for collecting GenTxs.
type CollectGenTxsRequest struct {
	// ChainID filters GenTxs by chain ID (optional)
	ChainID string
}

// InvalidGenTx represents a GenTx that failed validation.
type InvalidGenTx struct {
	// GenTx is the invalid GenTx
	GenTx domain.GenTx

	// Error is the validation error message
	Error string
}

// CollectGenTxsResult contains the result of collecting GenTxs.
type CollectGenTxsResult struct {
	// ValidGenTxs contains all valid GenTxs
	ValidGenTxs []domain.GenTx

	// InvalidGenTxs contains all invalid GenTxs with their errors
	InvalidGenTxs []InvalidGenTx

	// TotalCount is the total number of GenTxs processed
	TotalCount int

	// ValidCount is the number of valid GenTxs
	ValidCount int

	// InvalidCount is the number of invalid GenTxs
	InvalidCount int
}

// CollectGenTxsUseCase implements the use case for collecting and validating GenTxs.
//
// The use case performs the following operations:
//  1. Loads all GenTxs from repository
//  2. Optionally filters by chain ID
//  3. Validates each GenTx
//  4. Categorizes GenTxs as valid or invalid
//  5. Returns aggregated results with statistics
//
// This use case is typically used when:
//   - Collecting gentx files from multiple validators for genesis creation
//   - Validating a directory of gentx files before genesis creation
//   - Aggregating validator submissions for network initialization
type CollectGenTxsUseCase struct {
	repository Repository
	validator  Validator
}

// NewCollectGenTxsUseCase creates a new CollectGenTxsUseCase instance.
//
// Parameters:
//   - repository: Repository for retrieving GenTxs
//   - validator: Validator for validating GenTxs
//
// Returns a new CollectGenTxsUseCase ready for use.
func NewCollectGenTxsUseCase(
	repository Repository,
	validator Validator,
) *CollectGenTxsUseCase {
	return &CollectGenTxsUseCase{
		repository: repository,
		validator:  validator,
	}
}

// Execute executes the CollectGenTxs use case.
//
// The method performs the following operations:
//  1. Loads all GenTxs from repository
//  2. Applies optional filters (chain ID)
//  3. Validates each GenTx
//  4. Categorizes as valid or invalid
//  5. Returns aggregated results
//
// Parameters:
//   - request: CollectGenTxsRequest with optional filters (can be nil)
//
// Returns:
//   - CollectGenTxsResult: Aggregated results with valid/invalid GenTxs and statistics
//   - error: Error if repository access fails (validation errors are captured in result)
//
// Example:
//
//	request := &CollectGenTxsRequest{
//	    ChainID: "stable-testnet-1",
//	}
//	result, err := useCase.Execute(request)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Collected %d valid GenTxs, %d invalid\n", result.ValidCount, result.InvalidCount)
func (uc *CollectGenTxsUseCase) Execute(request *CollectGenTxsRequest) (*CollectGenTxsResult, error) {
	// Use default request if nil
	if request == nil {
		request = &CollectGenTxsRequest{}
	}

	// Load all GenTxs from repository
	allGenTxs, err := uc.repository.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load GenTxs: %w", err)
	}

	// Filter by chain ID if specified
	var filteredGenTxs []domain.GenTx
	if request.ChainID != "" {
		for _, gentx := range allGenTxs {
			if gentx.ChainID() == request.ChainID {
				filteredGenTxs = append(filteredGenTxs, gentx)
			}
		}
	} else {
		filteredGenTxs = allGenTxs
	}

	// Validate each GenTx and categorize
	var validGenTxs []domain.GenTx
	var invalidGenTxs []InvalidGenTx

	for _, gentx := range filteredGenTxs {
		err := uc.validator.Validate(gentx)
		if err != nil {
			// GenTx failed validation
			invalidGenTxs = append(invalidGenTxs, InvalidGenTx{
				GenTx: gentx,
				Error: err.Error(),
			})
		} else {
			// GenTx is valid
			validGenTxs = append(validGenTxs, gentx)
		}
	}

	// Build result
	result := &CollectGenTxsResult{
		ValidGenTxs:   validGenTxs,
		InvalidGenTxs: invalidGenTxs,
		TotalCount:    len(filteredGenTxs),
		ValidCount:    len(validGenTxs),
		InvalidCount:  len(invalidGenTxs),
	}

	return result, nil
}
