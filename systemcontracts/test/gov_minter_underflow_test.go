// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

// TestProposeMint_UnderflowGuard verifies that v2 GovMinter reverts with
// InsufficientMinterAllowance (not arithmetic underflow panic) when
// totalAllowance < reservedMintAmount.
//
// Scenario:
//  1. Deploy v1 GovMinter, configure allowance = 100M
//  2. proposeMint(5M) → reservedMintAmount = 5M
//  3. Reduce minter allowance to 3M (simulates external allowance reduction)
//  4. Apply v2 hardfork (bytecode swap)
//  5. proposeMint(1M) → totalAllowance(3M) < reservedMintAmount(5M)
//
// Expected: v2 reverts with InsufficientMinterAllowance instead of Panic(0x11).
func TestProposeMint_UnderflowGuard(t *testing.T) {
	// Use v1 environment first to build up reservedMintAmount
	env := createGovMinterTestEnv(t)
	ctx := env.GMinter
	defer ctx.backend.Close()

	member := env.MinterMembers[0].Operator
	initialAmount := big.NewInt(5_000_000)

	// Step 1: Increase allowance for test
	increasedAllowance := big.NewInt(100_000_000)
	tx, err := ctx.ConfigureMockFiatTokenMinter(t, member, TestGovMinterAddress, increasedAllowance)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Step 2: Create a mint proposal for 5M to build up reservedMintAmount
	tx, err = ctx.TxProposeMint(t, member, member.Address, initialAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Log("Created mint proposal for 5M on v1")

	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(initialAmount))
	t.Logf("reservedMintAmount: %s", reserved.String())

	// Step 3: Reduce minter allowance to 3M
	reducedAllowance := big.NewInt(3_000_000)
	tx, err = ctx.ConfigureMockFiatTokenMinter(t, member, TestGovMinterAddress, reducedAllowance)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Reduced minter allowance to %s (< reservedMintAmount %s)", reducedAllowance, reserved)

	// Step 4: Apply v2 hardfork
	env.GMinter.backend.CommitWithState(&params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: TestGovMinterAddress,
			Version: sc.SYSTEM_CONTRACT_VERSION_2,
		},
	}, nil)
	env.GMinter.govMinter = compiledGovMinterV2.New(env.GMinter.backend.Client(), TestGovMinterAddress)
	t.Log("Applied v2 hardfork")

	// Step 5: proposeMint(1M) on v2 — should revert with InsufficientMinterAllowance
	smallAmount := big.NewInt(1_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, smallAmount)
	err = ctx.ExpectedFail(tx, err)

	require.Error(t, err, "proposeMint should fail when totalAllowance < reservedMintAmount")
	t.Logf("proposeMint reverted: %v", err)

	revertErr, ok := err.(*RevertError)
	require.True(t, ok, "should revert with custom error, not arithmetic panic (got: %v)", err)
	require.Equal(t, "InsufficientMinterAllowance", revertErr.ABI.Name,
		"should revert with InsufficientMinterAllowance")
	t.Log("v2 correctly reverts with InsufficientMinterAllowance instead of underflow panic")
}
