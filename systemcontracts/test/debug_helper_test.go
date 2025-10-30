package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestDebugAddress(t *testing.T) {
	initGov(t)
	defer g.backend.Close()

	// Check if the address in the EOA matches the address derived from the private key
	for i, validator := range customValidators {
		derivedAddr := crypto.PubkeyToAddress(validator.Operator.PrivateKey.PublicKey)
		storedAddr := validator.Operator.Address

		t.Logf("Validator %d:", i)
		t.Logf("  Stored Address:  %s", storedAddr.Hex())
		t.Logf("  Derived Address: %s", derivedAddr.Hex())

		if derivedAddr != storedAddr {
			t.Errorf("Address mismatch for validator %d!", i)
		} else {
			t.Logf("  ✓ Addresses match")
		}
	}
}

func TestDebugMemberCheck(t *testing.T) {
	initGov(t)
	defer g.backend.Close()

	// Check if customValidators[0].Operator is a member
	member, err := g.BaseMembers(g.govValidator, customValidators[0].Operator, customValidators[0].Operator.Address)
	require.NoError(t, err)
	t.Logf("Member active: %v, JoinedAt: %v", member.IsActive, member.JoinedAt)
	require.True(t, member.IsActive, "customValidators[0].Operator should be an active member")

	// Try to call proposeAddMember
	t.Log("Attempting to call proposeAddMember...")
	_, tx, err := g.BaseTxProposeAddMember(t, g.govValidator, customValidators[0].Operator, newValidator.Operator.Address, 2)
	if err != nil {
		t.Logf("Transaction creation error: %v", err)
	}
	receipt, err := g.ExpectedOk(tx, err)
	if err != nil {
		t.Logf("Transaction execution error: %v", err)
	}
	require.NoError(t, err, "Should be able to propose adding a new member")
	t.Logf("Receipt status: %v", receipt.Status)
}

func TestDebugMembershipSimple(t *testing.T) {
	initGov(t)
	defer g.backend.Close()

	t.Logf("customValidators[0].Operator.Address: %s", customValidators[0].Operator.Address.Hex())

	// Check member version
	version, err := g.BaseMemberVersion(g.govValidator, customValidators[0].Operator)
	require.NoError(t, err)
	t.Logf("Member version: %v", version)

	// Check versioned member list
	for i := 0; i < 4; i++ {
		addr, err := g.BaseVersionedMemberList(g.govValidator, customValidators[0].Operator, version, big.NewInt(int64(i)))
		if err != nil {
			t.Logf("Error getting member at index %d: %v", i, err)
		} else {
			t.Logf("Member at index %d: %s", i, addr.Hex())
		}
	}

	// First, verify that the member is initialized correctly
	member, err := g.BaseMembers(g.govValidator, customValidators[0].Operator, customValidators[0].Operator.Address)
	require.NoError(t, err)
	t.Logf("Member isActive: %v, JoinedAt: %v", member.IsActive, member.JoinedAt)
	require.True(t, member.IsActive, "Member should be active after initialization")

	// Check quorum
	quorum, err := g.BaseQuorum(g.govValidator, customValidators[0].Operator)
	require.NoError(t, err)
	t.Logf("Quorum: %v", quorum)

	// Now try to propose adding a new member
	t.Log("Calling proposeAddMember...")
	_, tx, txErr := g.BaseTxProposeAddMember(t, g.govValidator, customValidators[0].Operator, newValidator.Operator.Address, 2)
	if txErr != nil {
		t.Logf("Transaction creation failed: %v", txErr)
	}

	_, err = g.ExpectedOk(tx, txErr)
	if err != nil {
		t.Logf("Transaction execution failed: %v", err)
	}
	require.NoError(t, err, "Should be able to propose adding a new member")

	t.Log("SUCCESS: proposeAddMember worked!")
}
