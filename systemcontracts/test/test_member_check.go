package test

import (
	"testing"
	"github.com/stretchr/testify/require"
)

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
	tx, err := g.BaseTxProposeAddMember(t, g.govValidator, customValidators[0].Operator, newValidator.Operator.Address, 2)
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
