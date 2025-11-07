package domain

import (
	"strings"
)

// ValidatorMetadata represents validator metadata value object.
// It is immutable and ensures valid metadata format.
type ValidatorMetadata struct {
	name        string
	description string
	website     string
}

// NewValidatorMetadata creates a new ValidatorMetadata.
// Name is required (1-70 chars), description is optional (max 280 chars), website is optional.
// All fields are trimmed of leading/trailing whitespace.
func NewValidatorMetadata(name, description, website string) (ValidatorMetadata, error) {
	// Trim whitespace
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	website = strings.TrimSpace(website)

	// Validate name (required, 1-70 chars)
	if name == "" {
		return ValidatorMetadata{}, ErrMissingValidatorName
	}

	if len(name) > 70 {
		return ValidatorMetadata{}, ErrValidatorNameTooLong
	}

	// Validate description (optional, max 280 chars)
	if len(description) > 280 {
		return ValidatorMetadata{}, ErrDescriptionTooLong
	}

	// Website is optional, no validation required
	// (URL validation can be done at application layer if needed)

	return ValidatorMetadata{
		name:        name,
		description: description,
		website:     website,
	}, nil
}

// Name returns the validator name.
func (m ValidatorMetadata) Name() string {
	return m.name
}

// Description returns the validator description.
func (m ValidatorMetadata) Description() string {
	return m.description
}

// Website returns the validator website.
func (m ValidatorMetadata) Website() string {
	return m.website
}

// Equals compares two ValidatorMetadata for equality.
func (m ValidatorMetadata) Equals(other ValidatorMetadata) bool {
	return m.name == other.name &&
		m.description == other.description &&
		m.website == other.website
}

// IsZero returns true if the metadata is the zero value.
func (m ValidatorMetadata) IsZero() bool {
	return m.name == "" && m.description == "" && m.website == ""
}
