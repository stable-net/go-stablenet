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
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestCollectCommand_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")
	os.MkdirAll(gentxDir, 0755)

	app := &cli.App{
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert - Should succeed with no GenTxs
	require.NoError(t, err)
}

func TestCollectCommand_AllValid(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create 3 valid GenTxs
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")

	app := &cli.App{
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestCollectCommand_FilterByChainID(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create GenTxs with different chain IDs
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "9999")

	app := &cli.App{
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
		"--chain-id", "8282",
	}

	// Act
	err := app.Run(args)

	// Assert - Should only collect 2 GenTxs for chain 8282
	require.NoError(t, err)
}

func TestCollectCommand_WithInvalidGenTx(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create 2 valid GenTxs
	createTestGenTx(t, gentxDir, "8282")
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
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
	}

	// Act
	err = app.Run(args)

	// Assert - Should fail due to invalid GenTx
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestCollectCommand_NonExistentDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "nonexistent")

	app := &cli.App{
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert - Should succeed with no GenTxs
	require.NoError(t, err)
}

func TestCollectCommand_NoMatchingChainID(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create GenTxs with different chain ID
	createTestGenTx(t, gentxDir, "9999")
	createTestGenTx(t, gentxDir, "9999")

	app := &cli.App{
		Commands: []*cli.Command{gentx.CollectCommand},
	}

	args := []string{
		"gstable",
		"collect",
		"--input-dir", gentxDir,
		"--chain-id", "8282",
	}

	// Act
	err := app.Run(args)

	// Assert - Should succeed with no matching GenTxs
	require.NoError(t, err)
}
