package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

var (
	g                 *GovWBFT
	customValidators  []*TestCandidate
	nonValidator      *TestCandidate
	newValidator      *TestCandidate
	anotherValidator  *TestCandidate
	anotherValidator2 *TestCandidate
)

func initGov(t *testing.T) {
	customValidators = []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
	nonValidator = NewTestCandidate()
	newValidator = NewTestCandidate()
	anotherValidator = NewTestCandidate()
	anotherValidator2 = NewTestCandidate()

	var err error
	g, err = NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[3].Operator.Address: {Balance: towei(1_000_000)},
		nonValidator.Operator.Address:        {Balance: towei(1_000_000)},
		newValidator.Operator.Address:        {Balance: towei(1_000_000)},
		anotherValidator.Operator.Address:    {Balance: towei(1_000_000)},
		anotherValidator2.Operator.Address:   {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members, validators, blsPubKeys string
		if len(customValidators) > 0 {
			for i, v := range customValidators {
				if i > 0 {
					members = members + ","
					validators = validators + ","
					blsPubKeys = blsPubKeys + ","
				}
				members = members + v.Operator.Address.String()
				validators = validators + v.Validator.Address.String()
				blsPubKeys = blsPubKeys + hexutil.Encode(v.GetBLSPublicKey(t).Marshal())
			}
			govValidator.Params = map[string]string{
				"members":       members,
				"quorum":        "2",
				"expiry":        "604800", // 7 days
				"memberVersion": "1",
				"validators":    validators,
				"blsPublicKeys": blsPubKeys,
			}
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
}

func TestGovValidator_configureValidator(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		initGov(t)
		defer g.backend.Close()

		quorum, err := g.BaseQuorum(g.govValidator, nonValidator.Operator)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorum)

		expiry, err := g.BaseProposalExpiry(g.govValidator, nonValidator.Operator)
		require.NoError(t, err)
		require.Equal(t, uint64(604800), expiry.Uint64()) // 7 days

		member, err := g.BaseMembers(g.govValidator, nonValidator.Operator, customValidators[0].Operator.Address)
		require.NoError(t, err)
		require.True(t, member.IsActive)
		require.Zero(t, member.JoinedAt)
		member, err = g.BaseMembers(g.govValidator, nonValidator.Operator, customValidators[1].Operator.Address)
		require.NoError(t, err)
		require.True(t, member.IsActive)
		require.Zero(t, member.JoinedAt)
		member, err = g.BaseMembers(g.govValidator, nonValidator.Operator, customValidators[2].Operator.Address)
		require.NoError(t, err)
		require.True(t, member.IsActive)
		require.Zero(t, member.JoinedAt)
		member, err = g.BaseMembers(g.govValidator, nonValidator.Operator, customValidators[3].Operator.Address)
		require.NoError(t, err)
		require.True(t, member.IsActive)
		require.Zero(t, member.JoinedAt)
		member, err = g.BaseMembers(g.govValidator, nonValidator.Operator, nonValidator.Operator.Address)
		require.NoError(t, err)
		require.False(t, member.IsActive)
		require.Zero(t, member.JoinedAt)

		version, err := g.BaseMemberVersion(g.govValidator, nonValidator.Operator)
		require.NoError(t, err)
		require.Equal(t, uint64(1), version.Uint64())

		// Verify all 4 members are in versionedMemberList (order-independent check)
		// Build a set of expected member addresses
		expectedMembers := make(map[common.Address]bool)
		for _, cv := range customValidators {
			expectedMembers[cv.Operator.Address] = true
		}

		// Check each position in versionedMemberList
		for i := uint64(0); i < 4; i++ {
			memberAddr, err := g.BaseVersionedMemberList(g.govValidator, nonValidator.Operator, version, new(big.Int).SetUint64(i))
			require.NoError(t, err, "Should be able to read member at index %d", i)
			// Verify this address is one of our expected members
			require.True(t, expectedMembers[memberAddr], "Member at index %d (%s) should be in customValidators list", i, memberAddr.Hex())
			// Mark as found (prevent duplicates)
			delete(expectedMembers, memberAddr)
		}

		// Verify all expected members were found
		require.Empty(t, expectedMembers, "All customValidators should be in versionedMemberList")

		// Verify index 4 is out of bounds
		_, err = g.BaseVersionedMemberList(g.govValidator, nonValidator.Operator, version, new(big.Int).SetUint64(4))
		require.Error(t, err, "Index 4 should be out of bounds")

		isValidator, err := g.IsValidator(nonValidator.Operator, customValidators[0].Validator.Address)
		require.NoError(t, err)
		require.True(t, isValidator)
		isValidator, err = g.IsValidator(nonValidator.Operator, customValidators[1].Validator.Address)
		require.NoError(t, err)
		require.True(t, isValidator)
		isValidator, err = g.IsValidator(nonValidator.Operator, customValidators[2].Validator.Address)
		require.NoError(t, err)
		require.True(t, isValidator)
		isValidator, err = g.IsValidator(nonValidator.Operator, customValidators[3].Validator.Address)
		require.NoError(t, err)
		require.True(t, isValidator)
		isValidator, err = g.IsValidator(nonValidator.Operator, nonValidator.Validator.Address)
		require.NoError(t, err)
		require.False(t, isValidator)

		valCount, err := g.ValidatorCount(nonValidator.Operator)
		require.NoError(t, err)
		require.Equal(t, uint64(4), valCount.Uint64())

		operator, err := g.ValidatorToOperator(nonValidator.Operator, customValidators[0].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[0].Operator.Address, operator)
		operator, err = g.ValidatorToOperator(nonValidator.Operator, customValidators[1].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[1].Operator.Address, operator)
		operator, err = g.ValidatorToOperator(nonValidator.Operator, customValidators[2].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[2].Operator.Address, operator)
		operator, err = g.ValidatorToOperator(nonValidator.Operator, customValidators[3].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[3].Operator.Address, operator)
		operator, err = g.ValidatorToOperator(nonValidator.Operator, nonValidator.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, common.Address{}, operator)

		validator, err := g.OperatorToValidator(nonValidator.Operator, customValidators[0].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[0].Validator.Address, validator)
		validator, err = g.OperatorToValidator(nonValidator.Operator, customValidators[1].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[1].Validator.Address, validator)
		validator, err = g.OperatorToValidator(nonValidator.Operator, customValidators[2].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[2].Validator.Address, validator)
		validator, err = g.OperatorToValidator(nonValidator.Operator, customValidators[3].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[3].Validator.Address, validator)
		validator, err = g.OperatorToValidator(nonValidator.Operator, nonValidator.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, common.Address{}, validator)

		blsKey, err := g.ValidatorToBlsKey(nonValidator.Operator, customValidators[0].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[0].GetBLSPublicKey(t).Marshal(), blsKey)
		blsKey, err = g.ValidatorToBlsKey(nonValidator.Operator, customValidators[1].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[1].GetBLSPublicKey(t).Marshal(), blsKey)
		blsKey, err = g.ValidatorToBlsKey(nonValidator.Operator, customValidators[2].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[2].GetBLSPublicKey(t).Marshal(), blsKey)
		blsKey, err = g.ValidatorToBlsKey(nonValidator.Operator, customValidators[3].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[3].GetBLSPublicKey(t).Marshal(), blsKey)
		blsKey, err = g.ValidatorToBlsKey(nonValidator.Operator, nonValidator.Validator.Address)
		require.NoError(t, err)
		require.Empty(t, blsKey)

		validator, err = g.BlsKeyToValidator(nonValidator.Operator, customValidators[0].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, customValidators[0].Validator.Address, validator)
		validator, err = g.BlsKeyToValidator(nonValidator.Operator, customValidators[1].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, customValidators[1].Validator.Address, validator)
		validator, err = g.BlsKeyToValidator(nonValidator.Operator, customValidators[2].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, customValidators[2].Validator.Address, validator)
		validator, err = g.BlsKeyToValidator(nonValidator.Operator, customValidators[3].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, customValidators[3].Validator.Address, validator)
		validator, err = g.BlsKeyToValidator(nonValidator.Operator, nonValidator.GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, common.Address{}, validator)
	})

	t.Run("configureValidator", func(t *testing.T) {
		initGov(t)
		defer g.backend.Close()

		// error cases
		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				nonValidator.Operator,
				nonValidator.Validator.Address,
				nonValidator.GetBLSPublicKey(t).Marshal(),
				nonValidator.GetBLSPoPSignature(t).Marshal())),
			"NotAMember",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				common.Address{},
				nonValidator.GetBLSPublicKey(t).Marshal(),
				nonValidator.GetBLSPoPSignature(t).Marshal())),
			"InvalidValidator",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[1].Validator.Address,
				customValidators[1].GetBLSPublicKey(t).Marshal(),
				customValidators[1].GetBLSPoPSignature(t).Marshal())),
			"AlreadyValidatorExists",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[0].GetBLSPublicKey(t).Marshal(),
				customValidators[1].GetBLSPoPSignature(t).Marshal())),
			"InvalidBlsKey",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[0].GetBLSPublicKey(t).Marshal()[1:],
				customValidators[0].GetBLSPoPSignature(t).Marshal())),
			"InvalidBlsKeyLength",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[0].GetBLSPublicKey(t).Marshal(),
				customValidators[0].GetBLSPoPSignature(t).Marshal()[1:])),
			"InvalidSignatureLength",
		)

		invalidSig := customValidators[0].GetBLSPoPSignature(t).Marshal()
		invalidSig[0] ^= 0xFF
		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[0].GetBLSPublicKey(t).Marshal(),
				invalidSig)),
			"FailedToVerifyBlsKey",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[1].GetBLSPublicKey(t).Marshal(),
				customValidators[1].GetBLSPoPSignature(t).Marshal())),
			"AlreadyRegisteredBlsKey",
		)

		ExpectedRevert(t,
			g.ExpectedFail(g.validatorContractTx(
				t,
				"configureValidator",
				customValidators[0].Operator,
				customValidators[0].Validator.Address,
				customValidators[0].GetBLSPublicKey(t).Marshal(),
				customValidators[0].GetBLSPoPSignature(t).Marshal())),
			"NoConfigurationChanging",
		)

		// success case 1: register new validator
		_, tx, err := g.BaseTxProposeAddMember(t,
			g.govValidator,
			customValidators[0].Operator,
			newValidator.Operator.Address,
			2)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)
		proposalId, err := g.BaseCurrentProposalId(g.govValidator, customValidators[0].Operator)
		require.NoError(t, err)
		currentProposal, err := g.BaseGetProposal(g.govValidator, customValidators[0].Operator, proposalId)
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, currentProposal.Status) // Voting
		require.Equal(t, crypto.Keccak256Hash([]byte("ACTION_ADD_MEMBER")), common.BytesToHash(currentProposal.ActionType[:]))
		require.Equal(t, uint32(2), currentProposal.RequiredApprovals)

		_, err = g.ExpectedOk(g.BaseTxApproveProposal(t,
			g.govValidator,
			customValidators[1].Operator, proposalId))
		require.NoError(t, err)
		currentProposal, err = g.BaseGetProposal(g.govValidator, customValidators[0].Operator, proposalId)
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusExecuted, currentProposal.Status) // Executed
		memberVersion, err := g.BaseMemberVersion(g.govValidator, customValidators[0].Operator)
		require.NoError(t, err)
		require.Equal(t, uint64(2), memberVersion.Uint64())

		newMember, err := g.BaseMembers(g.govValidator, nonValidator.Operator, newValidator.Operator.Address)
		require.NoError(t, err)
		require.True(t, newMember.IsActive)

		noVal, err := g.OperatorToValidator(customValidators[0].Operator, newValidator.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, common.Address{}, noVal)
		_, err = g.ExpectedOk(g.validatorContractTx(t,
			"configureValidator",
			newValidator.Operator,
			newValidator.Validator.Address,
			newValidator.GetBLSPublicKey(t).Marshal(),
			newValidator.GetBLSPoPSignature(t).Marshal()))
		require.NoError(t, err)
		// verify
		isVal, err := g.IsValidator(newValidator.Operator, newValidator.Validator.Address)
		require.NoError(t, err)
		require.True(t, isVal)
		val, err := g.OperatorToValidator(newValidator.Operator, newValidator.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, newValidator.Validator.Address, val)
		valCount, err := g.ValidatorCount(newValidator.Operator)
		require.NoError(t, err)
		require.Equal(t, uint64(5), valCount.Uint64())
		blsKey, err := g.ValidatorToBlsKey(newValidator.Operator, newValidator.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, newValidator.GetBLSPublicKey(t).Marshal(), blsKey)
		val, err = g.BlsKeyToValidator(newValidator.Operator, newValidator.GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, newValidator.Validator.Address, val)
		op, err := g.ValidatorToOperator(newValidator.Operator, newValidator.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, newValidator.Operator.Address, op)

		// success case 2: change BlsKey of existing validator
		_, err = g.ExpectedOk(g.validatorContractTx(t,
			"configureValidator",
			customValidators[0].Operator,
			customValidators[0].Validator.Address,
			nonValidator.GetBLSPublicKey(t).Marshal(),
			nonValidator.GetBLSPoPSignature(t).Marshal()))
		require.NoError(t, err)

		// verify
		blsKey, err = g.ValidatorToBlsKey(nonValidator.Operator, customValidators[0].Validator.Address)
		require.NoError(t, err)
		require.Equal(t, nonValidator.GetBLSPublicKey(t).Marshal(), blsKey)
		val, err = g.BlsKeyToValidator(nonValidator.Operator, customValidators[0].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, common.Address{}, val)
		val, err = g.BlsKeyToValidator(nonValidator.Operator, nonValidator.GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, customValidators[0].Validator.Address, val)

		// success case 3: change validator address of existing validator(blsKey unchanged)
		_, err = g.ExpectedOk(g.validatorContractTx(t,
			"configureValidator",
			customValidators[1].Operator,
			anotherValidator.Validator.Address,
			customValidators[1].GetBLSPublicKey(t).Marshal(),
			customValidators[1].GetBLSPoPSignature(t).Marshal()))
		require.NoError(t, err)
		// verify
		isVal, err = g.IsValidator(anotherValidator.Operator, anotherValidator.Validator.Address)
		require.NoError(t, err)
		require.True(t, isVal)
		isVal, err = g.IsValidator(anotherValidator.Operator, customValidators[1].Validator.Address)
		require.NoError(t, err)
		require.False(t, isVal)
		val, err = g.OperatorToValidator(anotherValidator.Operator, customValidators[1].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, anotherValidator.Validator.Address, val)
		op, err = g.ValidatorToOperator(anotherValidator.Operator, anotherValidator.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[1].Operator.Address, op)
		blsKey, err = g.ValidatorToBlsKey(anotherValidator.Operator, anotherValidator.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[1].GetBLSPublicKey(t).Marshal(), blsKey)
		val, err = g.BlsKeyToValidator(anotherValidator.Operator, customValidators[1].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, anotherValidator.Validator.Address, val)

		// success case 4: change validator address and blsKey of existing validator
		_, err = g.ExpectedOk(g.validatorContractTx(t,
			"configureValidator",
			customValidators[2].Operator,
			anotherValidator2.Validator.Address,
			anotherValidator2.GetBLSPublicKey(t).Marshal(),
			anotherValidator2.GetBLSPoPSignature(t).Marshal()))
		require.NoError(t, err)
		// verify
		isVal, err = g.IsValidator(anotherValidator2.Operator, anotherValidator2.Validator.Address)
		require.NoError(t, err)
		require.True(t, isVal)
		isVal, err = g.IsValidator(anotherValidator2.Operator, customValidators[2].Validator.Address)
		require.NoError(t, err)
		require.False(t, isVal)
		val, err = g.OperatorToValidator(anotherValidator2.Operator, customValidators[2].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, anotherValidator2.Validator.Address, val)
		op, err = g.ValidatorToOperator(anotherValidator2.Operator, anotherValidator2.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, customValidators[2].Operator.Address, op)
		blsKey, err = g.ValidatorToBlsKey(anotherValidator2.Operator, anotherValidator2.Validator.Address)
		require.NoError(t, err)
		require.Equal(t, anotherValidator2.GetBLSPublicKey(t).Marshal(), blsKey)
		val, err = g.BlsKeyToValidator(anotherValidator2.Operator, anotherValidator2.GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, anotherValidator2.Validator.Address, val)
		val, err = g.BlsKeyToValidator(anotherValidator2.Operator, customValidators[2].GetBLSPublicKey(t).Marshal())
		require.NoError(t, err)
		require.Equal(t, common.Address{}, val)
	})
}
