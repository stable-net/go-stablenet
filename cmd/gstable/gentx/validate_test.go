// Copyright 2025 The go-stablenet Authors
// This file is part of go-stablenet.
//
// go-stablenet is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-stablenet. If not, see <http://www.gnu.org/licenses/>.

package gentx_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/cmd/gstable/gentx"
	"github.com/ethereum/go-ethereum/internal/genutils/application"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func createTestGenTx(t *testing.T, outputDir string, chainID string) domain.Address {
	cryptoProvider := ethcrypto.NewEthereumProvider()
	repo := repository.NewFileRepository(outputDir)
	validator := validation.NewValidationService(cryptoProvider)
	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Generate validator key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Generate operator address
	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	// Create metadata
	metadata, err := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	require.NoError(t, err)

	// Create GenTx
	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      chainID,
		Timestamp:    time.Now().UTC(),
	}

	gentx, err := createUseCase.Execute(request)
	require.NoError(t, err)

	return gentx.ValidatorAddress()
}

func TestValidateCommand_SingleGenTx_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create a test GenTx
	validatorAddr := createTestGenTx(t, gentxDir, "8282")

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
		"--validator-addr", validatorAddr.String(),
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestValidateCommand_AllGenTxs_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create 3 test GenTxs
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestValidateCommand_FilterByChainID(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create GenTxs with different chain IDs
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "9999")

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
		"--chain-id", "8282",
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestValidateCommand_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")
	os.MkdirAll(gentxDir, 0755)

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert - Should succeed with no GenTxs
	require.NoError(t, err)
}

func TestValidateCommand_InvalidValidatorAddress(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
		"--validator-addr", "invalid-address",
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator address")
}

func TestValidateCommand_ValidatorNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")
	os.MkdirAll(gentxDir, 0755)

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
		"--validator-addr", "0x1234567890123456789012345678901234567890",
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidateCommand_WithInvalidGenTx(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create a valid GenTx first
	createTestGenTx(t, gentxDir, "8282")

	// Create an invalid GenTx manually
	cryptoProvider := ethcrypto.NewEthereumProvider()
	repo := repository.NewFileRepository(gentxDir)

	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	validatorSig, err := cryptoProvider.Sign(validatorKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, validatorSig)
	require.NoError(t, err)

	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	blsPublicKey, err := cryptoProvider.DeriveBLSPublicKey(validatorKey)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Invalid Validator", "Invalid Description", "https://invalid.com")
	require.NoError(t, err)

	// Create an invalid signature (wrong message)
	invalidSignature, err := cryptoProvider.Sign(validatorKey, []byte("wrong message"))
	require.NoError(t, err)

	invalidGenTx, err := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsPublicKey,
		metadata,
		invalidSignature,
		"8282",
		time.Now().UTC(),
	)
	require.NoError(t, err)

	// Save directly (bypassing validation)
	err = repo.Save(invalidGenTx)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.ValidateCommand},
	}

	args := []string{
		"gstable",
		"validate",
		"--input-dir", gentxDir,
	}

	// Act
	err = app.Run(args)

	// Assert - Should fail due to invalid GenTx
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
