package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// FileRepository implements GenTxRepository using the filesystem.
// Each GenTx is stored as a separate JSON file named gentx-{validator-address}.json.
// Provides atomic write operations to prevent corruption.
type FileRepository struct {
	baseDir string
}

// NewFileRepository creates a new FileRepository with the given base directory.
// The directory will be created if it doesn't exist.
func NewFileRepository(baseDir string) *FileRepository {
	// Ensure directory exists
	os.MkdirAll(baseDir, 0755)
	return &FileRepository{
		baseDir: baseDir,
	}
}

// Save persists a GenTx to the filesystem.
// Returns ErrDuplicateValidatorAddress if the validator address already exists.
func (r *FileRepository) Save(gentx domain.GenTx) error {
	// Check if already exists
	exists, err := r.Exists(gentx.ValidatorAddress())
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if exists {
		return domain.ErrDuplicateValidatorAddress
	}

	// Generate file path
	filePath := r.genFilePath(gentx.ValidatorAddress())

	// Write to file atomically
	if err := r.writeGenTxFile(filePath, gentx); err != nil {
		return fmt.Errorf("failed to write gentx file: %w", err)
	}

	return nil
}

// FindAll retrieves all GenTxs from the filesystem.
// Returns an empty slice if no GenTxs are found.
func (r *FileRepository) FindAll() ([]domain.GenTx, error) {
	// Read directory
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.GenTx{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Read all gentx files
	var gentxs []domain.GenTx
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process gentx-*.json files
		if !strings.HasPrefix(entry.Name(), "gentx-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(r.baseDir, entry.Name())
		gentx, err := r.readGenTxFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read gentx file %s: %w", entry.Name(), err)
		}

		gentxs = append(gentxs, gentx)
	}

	return gentxs, nil
}

// FindByValidator retrieves a GenTx by validator address.
// Returns ErrGenTxNotFound if the GenTx does not exist.
func (r *FileRepository) FindByValidator(validatorAddress domain.Address) (domain.GenTx, error) {
	// Check if exists
	exists, err := r.Exists(validatorAddress)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to check existence: %w", err)
	}
	if !exists {
		return domain.GenTx{}, domain.ErrGenTxNotFound
	}

	// Read file
	filePath := r.genFilePath(validatorAddress)
	gentx, err := r.readGenTxFile(filePath)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to read gentx file: %w", err)
	}

	return gentx, nil
}

// Exists checks if a GenTx with the given validator address exists.
func (r *FileRepository) Exists(validatorAddress domain.Address) (bool, error) {
	filePath := r.genFilePath(validatorAddress)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// Delete removes a GenTx from the filesystem by validator address.
// Returns ErrGenTxNotFound if the GenTx does not exist.
func (r *FileRepository) Delete(validatorAddress domain.Address) error {
	// Check if exists
	exists, err := r.Exists(validatorAddress)
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if !exists {
		return domain.ErrGenTxNotFound
	}

	// Delete file
	filePath := r.genFilePath(validatorAddress)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Count returns the total number of GenTxs in the filesystem.
func (r *FileRepository) Count() (int, error) {
	gentxs, err := r.FindAll()
	if err != nil {
		return 0, fmt.Errorf("failed to find all gentxs: %w", err)
	}
	return len(gentxs), nil
}

// genFilePath generates the file path for a given validator address.
// Format: {baseDir}/gentx-{address-lowercase-without-0x}.json
func (r *FileRepository) genFilePath(address domain.Address) string {
	// Get address string and convert to lowercase
	addrStr := strings.ToLower(address.String())
	// Remove 0x prefix
	if len(addrStr) > 2 && addrStr[:2] == "0x" {
		addrStr = addrStr[2:]
	}
	fileName := fmt.Sprintf("gentx-%s.json", addrStr)
	return filepath.Join(r.baseDir, fileName)
}

// readGenTxFile reads a GenTx from a JSON file.
func (r *FileRepository) readGenTxFile(filePath string) (domain.GenTx, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON
	var dto GenTxDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Convert to domain object
	gentx, err := dto.ToDomain()
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to convert DTO to domain: %w", err)
	}

	return gentx, nil
}

// writeGenTxFile writes a GenTx to a JSON file atomically.
// Uses a temporary file and renames it to prevent corruption.
func (r *FileRepository) writeGenTxFile(filePath string, gentx domain.GenTx) error {
	// Convert to DTO
	dto := ToDTO(gentx)

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (this is atomic on POSIX systems)
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
