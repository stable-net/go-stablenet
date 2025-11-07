package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestValidatorMetadata_NewValidatorMetadata_ValidData(t *testing.T) {
	// Arrange
	name := "My Validator"
	description := "A reliable validator for the network"
	website := "https://myvalidator.com"

	// Act
	metadata, err := domain.NewValidatorMetadata(name, description, website)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, name, metadata.Name())
	assert.Equal(t, description, metadata.Description())
	assert.Equal(t, website, metadata.Website())
}

func TestValidatorMetadata_NewValidatorMetadata_MinimalData(t *testing.T) {
	// Arrange - only name is required
	name := "V"
	description := ""
	website := ""

	// Act
	metadata, err := domain.NewValidatorMetadata(name, description, website)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, name, metadata.Name())
	assert.Equal(t, "", metadata.Description())
	assert.Equal(t, "", metadata.Website())
}

func TestValidatorMetadata_NewValidatorMetadata_InvalidName(t *testing.T) {
	tests := []struct {
		name        string
		validatorName string
		description string
	}{
		{
			name:          "empty name",
			validatorName: "",
			description:   "should reject empty validator name",
		},
		{
			name:          "whitespace only",
			validatorName: "   ",
			description:   "should reject whitespace-only name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := domain.NewValidatorMetadata(tt.validatorName, "description", "https://test.com")

			// Assert
			assert.Error(t, err, tt.description)
			assert.ErrorIs(t, err, domain.ErrMissingValidatorName)
		})
	}
}

func TestValidatorMetadata_NewValidatorMetadata_NameTooLong(t *testing.T) {
	// Arrange - 71 characters (max is 70)
	name := strings.Repeat("a", 71)

	// Act
	_, err := domain.NewValidatorMetadata(name, "description", "https://test.com")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrValidatorNameTooLong)
}

func TestValidatorMetadata_NewValidatorMetadata_DescriptionTooLong(t *testing.T) {
	// Arrange - 281 characters (max is 280)
	description := strings.Repeat("a", 281)

	// Act
	_, err := domain.NewValidatorMetadata("ValidName", description, "https://test.com")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDescriptionTooLong)
}

func TestValidatorMetadata_Name_MaxLength(t *testing.T) {
	// Arrange - exactly 70 characters (max allowed)
	name := strings.Repeat("a", 70)

	// Act
	metadata, err := domain.NewValidatorMetadata(name, "description", "https://test.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, name, metadata.Name())
}

func TestValidatorMetadata_Description_MaxLength(t *testing.T) {
	// Arrange - exactly 280 characters (max allowed)
	description := strings.Repeat("a", 280)

	// Act
	metadata, err := domain.NewValidatorMetadata("ValidName", description, "https://test.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, description, metadata.Description())
}

func TestValidatorMetadata_Website_Optional(t *testing.T) {
	// Arrange
	name := "My Validator"
	description := "Description"

	// Act - website is optional
	metadata, err := domain.NewValidatorMetadata(name, description, "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "", metadata.Website())
}

func TestValidatorMetadata_Website_ValidURLs(t *testing.T) {
	tests := []struct {
		name    string
		website string
	}{
		{
			name:    "https URL",
			website: "https://validator.com",
		},
		{
			name:    "http URL",
			website: "http://validator.com",
		},
		{
			name:    "URL with path",
			website: "https://validator.com/info",
		},
		{
			name:    "URL with subdomain",
			website: "https://staking.validator.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			metadata, err := domain.NewValidatorMetadata("ValidName", "description", tt.website)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.website, metadata.Website())
		})
	}
}

func TestValidatorMetadata_Equals(t *testing.T) {
	// Arrange
	metadata1, _ := domain.NewValidatorMetadata("Validator1", "Description1", "https://test1.com")
	metadata2, _ := domain.NewValidatorMetadata("Validator1", "Description1", "https://test1.com")
	metadata3, _ := domain.NewValidatorMetadata("Validator2", "Description2", "https://test2.com")

	// Act & Assert
	assert.True(t, metadata1.Equals(metadata2), "identical metadata should be equal")
	assert.False(t, metadata1.Equals(metadata3), "different metadata should not be equal")
}

func TestValidatorMetadata_Equals_NameOnly(t *testing.T) {
	// Arrange - same name, different description and website
	metadata1, _ := domain.NewValidatorMetadata("Validator1", "Description1", "https://test1.com")
	metadata2, _ := domain.NewValidatorMetadata("Validator1", "Description2", "https://test2.com")

	// Act & Assert
	assert.False(t, metadata1.Equals(metadata2), "metadata with different description should not be equal")
}

func TestValidatorMetadata_IsZero(t *testing.T) {
	// Test zero metadata
	t.Run("zero metadata", func(t *testing.T) {
		zero := domain.ValidatorMetadata{}
		assert.True(t, zero.IsZero(), "uninitialized metadata should be zero")
	})

	// Test non-zero metadata
	t.Run("non-zero metadata", func(t *testing.T) {
		nonZero, _ := domain.NewValidatorMetadata("ValidName", "", "")
		assert.False(t, nonZero.IsZero(), "initialized metadata should not be zero")
	})
}

func TestValidatorMetadata_NameTrimming(t *testing.T) {
	// Arrange - name with leading/trailing spaces
	name := "  My Validator  "

	// Act
	metadata, err := domain.NewValidatorMetadata(name, "description", "https://test.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "My Validator", metadata.Name(), "name should be trimmed")
}

func TestValidatorMetadata_DescriptionTrimming(t *testing.T) {
	// Arrange - description with leading/trailing spaces
	description := "  This is a description  "

	// Act
	metadata, err := domain.NewValidatorMetadata("ValidName", description, "https://test.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "This is a description", metadata.Description(), "description should be trimmed")
}

func TestValidatorMetadata_WebsiteTrimming(t *testing.T) {
	// Arrange - website with leading/trailing spaces
	website := "  https://validator.com  "

	// Act
	metadata, err := domain.NewValidatorMetadata("ValidName", "description", website)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "https://validator.com", metadata.Website(), "website should be trimmed")
}

func TestValidatorMetadata_Immutability(t *testing.T) {
	// Arrange
	originalName := "Original Name"
	originalDesc := "Original Description"
	originalWebsite := "https://original.com"

	metadata, _ := domain.NewValidatorMetadata(originalName, originalDesc, originalWebsite)

	// Act - verify that getters return copies/immutable values
	name := metadata.Name()
	desc := metadata.Description()
	website := metadata.Website()

	// Assert - original values should remain unchanged
	assert.Equal(t, originalName, name)
	assert.Equal(t, originalDesc, desc)
	assert.Equal(t, originalWebsite, website)

	// Verify we can't modify through another instance
	metadata2, _ := domain.NewValidatorMetadata("Different", "Different", "https://different.com")
	assert.NotEqual(t, metadata.Name(), metadata2.Name())
}
