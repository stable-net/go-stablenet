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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/cmd/gstable/gentx"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestCreateCommand_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")
	outputDir := filepath.Join(tempDir, "gentx")

	// Generate a test key
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	// Generate operator address
	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	// Create CLI context
	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", operatorAddr.String(),
		"--name", "Test Validator",
		"--description", "Test Description",
		"--website", "https://test.com",
		"--chain-id", "8282",
		"--output-dir", outputDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.NoError(t, err)

	// Verify output file was created
	entries, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "gentx-")
	assert.Contains(t, entries[0].Name(), ".json")
}

func TestCreateCommand_MissingValidatorKey(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--operator-addr", "0x1234567890123456789012345678901234567890",
		"--name", "Test Validator",
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator-key")
}

func TestCreateCommand_MissingOperatorAddr(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")

	// Generate a test key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--name", "Test Validator",
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operator-addr")
}

func TestCreateCommand_MissingName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")

	// Generate a test key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", "0x1234567890123456789012345678901234567890",
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestCreateCommand_MissingChainID(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")

	// Generate a test key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", "0x1234567890123456789012345678901234567890",
		"--name", "Test Validator",
		"--output-dir", tempDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain-id")
}

func TestCreateCommand_InvalidOperatorAddress(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")

	// Generate a test key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", "invalid-address",
		"--name", "Test Validator",
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operator address")
}

func TestCreateCommand_InvalidValidatorKeyFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "nonexistent.key")

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", "0x1234567890123456789012345678901234567890",
		"--name", "Test Validator",
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator key file")
}

func TestCreateCommand_InvalidMetadata(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")

	// Generate a test key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	// Generate operator address
	cryptoProvider := ethcrypto.NewEthereumProvider()
	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	// Name too long (>70 chars)
	longName := string(make([]byte, 71))
	for i := range longName {
		longName = string(append([]byte(longName[:i]), 'a'))
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", operatorAddr.String(),
		"--name", longName,
		"--chain-id", "8282",
		"--output-dir", tempDir,
	}

	// Act
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
}

func TestCreateCommand_DuplicateValidator(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	keyFile := filepath.Join(tempDir, "validator.key")
	outputDir := filepath.Join(tempDir, "gentx")

	// Generate a test key
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Save key to file
	err = os.WriteFile(keyFile, validatorKey, 0600)
	require.NoError(t, err)

	// Generate operator address
	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CreateCommand},
	}

	args := []string{
		"gstable",
		"create",
		"--validator-key", keyFile,
		"--operator-addr", operatorAddr.String(),
		"--name", "Test Validator",
		"--chain-id", "8282",
		"--output-dir", outputDir,
	}

	// Create first gentx
	err = app.Run(args)
	require.NoError(t, err)

	// Act - Try to create duplicate
	err = app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}
