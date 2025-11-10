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
)

func TestInspectCommand_SingleGenTx_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create a test GenTx
	validatorAddr := createTestGenTx(t, gentxDir, "8282")

	app := &cli.App{
		Commands: []*cli.Command{gentx.InspectCommand},
	}

	args := []string{
		"gstable",
		"inspect",
		"--input-dir", gentxDir,
		"--validator-addr", validatorAddr.String(),
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestInspectCommand_AllGenTxs_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	// Create 3 test GenTxs
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")
	createTestGenTx(t, gentxDir, "8282")

	app := &cli.App{
		Commands: []*cli.Command{gentx.InspectCommand},
	}

	args := []string{
		"gstable",
		"inspect",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert
	require.NoError(t, err)
}

func TestInspectCommand_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")
	os.MkdirAll(gentxDir, 0755)

	app := &cli.App{
		Commands: []*cli.Command{gentx.InspectCommand},
	}

	args := []string{
		"gstable",
		"inspect",
		"--input-dir", gentxDir,
	}

	// Act
	err := app.Run(args)

	// Assert - Should succeed with no GenTxs
	require.NoError(t, err)
}

func TestInspectCommand_InvalidValidatorAddress(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")

	app := &cli.App{
		Commands: []*cli.Command{gentx.InspectCommand},
	}

	args := []string{
		"gstable",
		"inspect",
		"--input-dir", gentxDir,
		"--validator-addr", "invalid-address",
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator address")
}

func TestInspectCommand_ValidatorNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	gentxDir := filepath.Join(tempDir, "gentx")
	os.MkdirAll(gentxDir, 0755)

	app := &cli.App{
		Commands: []*cli.Command{gentx.InspectCommand},
	}

	args := []string{
		"gstable",
		"inspect",
		"--input-dir", gentxDir,
		"--validator-addr", "0x1234567890123456789012345678901234567890",
	}

	// Act
	err := app.Run(args)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load GenTx")
}
