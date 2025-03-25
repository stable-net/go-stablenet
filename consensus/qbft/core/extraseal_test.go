package core

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	signers    []*ecdsa.PrivateKey
	validators []common.Address
	blsPubKeys [][]byte
)

func init() {
	_, signers, _ = testutils.GenesisAndKeys(5)
	validators = make([]common.Address, 0)
	blsPubKeys = make([][]byte, 0)
	for _, key := range signers {
		blsKey, _ := bls.DeriveFromECDSA(key)
		validators = append(validators, crypto.PubkeyToAddress(key.PublicKey))
		blsPubKeys = append(blsPubKeys, blsKey.PublicKey().Marshal())
	}
}

// makeCoreForTest returns core object with empty backend.
// Its purpose is to test pure qbft/core functions.
// not recommended for testing logic that includes qbft/backend function
func makeCoreForTest(priorRound, currentRound, currentSequence *big.Int, lastProposal *types.Block) *Core {
	// set core with empty backend.
	// current state is StateAcceptRequest.
	valSet := validator.NewSet(validators, blsPubKeys, qbft.NewRoundRobinProposerPolicy())
	core := &Core{
		config:             nil,
		address:            crypto.PubkeyToAddress(signers[0].PublicKey),
		state:              StateAcceptRequest,
		handlerWg:          new(sync.WaitGroup),
		logger:             log.New("address", crypto.PubkeyToAddress(signers[0].PublicKey)),
		backend:            nil,
		backlogs:           make(map[common.Address]*prque.Prque[int64, messages.QBFTMessage]),
		backlogsMu:         new(sync.Mutex),
		prepareExtraSeals:  make(map[common.Address]*messages.Prepare),
		commitExtraSeals:   make(map[common.Address]*messages.Commit),
		extraSealsMu:       new(sync.Mutex),
		pendingRequests:    prque.New[int64, *Request](nil),
		pendingRequestsMu:  new(sync.Mutex),
		consensusTimestamp: time.Time{},
		priorState:         priorState{new(sync.RWMutex), priorRound, lastProposal, valSet},
	}
	core.validateFn = core.checkValidatorSignature
	core.valSet = valSet
	// Set core current view and proposal
	core.current = newRoundState(
		&qbft.View{
			Round:    currentRound,
			Sequence: currentSequence,
		}, valSet,
		nil, nil, nil, nil, func(hash common.Hash) bool { return false })
	return core
}

// TestAddToExtraSeal is to test core.addToExtraSeal function
// test case is when core.state is StateAcceptedRequest
func TestAddToExtraSeal(t *testing.T) {
	// make proposals
	lastProposal := makeProposal(common.Big2)
	invalidLastProposal := makeProposal(common.Big3)

	// make core instance  with empty backend
	core := makeCoreForTest(common.Big2, common.Big0, common.Big3, lastProposal)

	type testMessage struct {
		message       messages.QBFTMessage
		expectedError error
	}

	malformedSeal := func(commitMsg *messages.Commit) *messages.Commit {
		seal := []byte("malformedSeal")
		commitMsg.CommitSeal = seal
		return commitMsg
	}
	expectedSigner1Prepare := createPrepareMsg(signers[1], lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes())
	expectedSigner2Commit := createCommitMsg(signers[2], lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes())

	testExtraSealMessages := []testMessage{
		{
			expectedSigner1Prepare,
			nil,
		},
		{
			expectedSigner2Commit,
			nil,
		},
		{
			createPreprepareMsg(common.Big2, common.Big2, lastProposal),
			errInvalidExtraSealMessage,
		},
		{
			// fail to verify digest -  message's proposal is not same with currentProposal
			createPrepareMsg(signers[3], invalidLastProposal.Header(), common.Big2, common.Big2, invalidLastProposal.Hash().Bytes()),
			errInvalidMessage,
		},
		{
			// fail to verify seal -  seal doesn't match with message
			malformedSeal(createCommitMsg(signers[4], lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes())),
			errInvalidSeal,
		},
		{
			// ignored - lower view than existing one
			createPrepareMsg(signers[1], lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes()),
			nil,
		},
		{
			// ignored - lower view than existing one - ignored
			createCommitMsg(signers[2], lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes()),
			nil,
		},
	}

	for _, tm := range testExtraSealMessages {
		err := core.addToExtraSeal(tm.message)
		if !errors.Is(err, tm.expectedError) {
			t.Errorf("unexpected error adding to extraSeal. want %v, have %v", tm.expectedError, err)
		}
	}

	signer1Prepare := core.prepareExtraSeals[crypto.PubkeyToAddress(signers[1].PublicKey)]
	signer2Commit := core.commitExtraSeals[crypto.PubkeyToAddress(signers[2].PublicKey)]
	if signer1Prepare != expectedSigner1Prepare {
		t.Errorf("unexpected stored extraSeal message. want %v, have %v", expectedSigner1Prepare, signer1Prepare)
	}
	if signer2Commit != expectedSigner2Commit {
		t.Errorf("unexpected stored extraSeal message. want %v, have %v", expectedSigner2Commit, signer2Commit)
	}
}

// TestProcessExtraSeal is to test core.ProcessExtraSeal
// 1. if mesasge have lower view than LatestView, drop it
// 2. if wrong digest, drop it
// 3. message other than prepare/commit, drop it
func TestProcessExtraSeal(t *testing.T) {
	// make proposals
	lastProposal := makeProposal(common.Big2)

	// make core instance  with empty backend
	core := makeCoreForTest(common.Big2, common.Big0, common.Big3, lastProposal)

	// assume situation when consensus enters new round and preparing for new block.
	// set core.current.proposal. Proposal's block number should be current.Sequence -1
	core.current.SetPreprepare(createPreprepareMsg(common.Big2, common.Big2, lastProposal))

	testExtraSealMessages := []messages.QBFTMessage{
		// 1. 2 valid extra seal messages
		createPrepareMsg(signers[1], lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(signers[2], lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		// 2. 3 extra seal messages with lower view than latestView
		createPrepareMsg(signers[3], lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes()),
		createCommitMsg(signers[4], lastProposal.Header(), common.Big1, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(signers[0], lastProposal.Header(), common.Big1, common.Big2, lastProposal.Hash().Bytes()),
	}

	for _, tm := range testExtraSealMessages {
		if err := core.addToExtraSeal(tm); err != nil {
			t.Errorf("error adding to core.extraSeals : %v", err)
		}
	}

	preparedSeal, committedSeal := core.ProcessExtraSeal(lastProposal, common.Big2, core.PriorValidators())
	if len(preparedSeal) != 1 {
		t.Errorf("unexpected length of preparedSeal. want %d, have %d", 1, len(preparedSeal))
	}
	if len(committedSeal) != 1 {
		t.Errorf("unexpected length of committedSeal. want %d, have %d", 1, len(committedSeal))
	}
}

func createPreprepareMsg(sequence, round *big.Int, proposal qbft.Proposal) *messages.Preprepare {
	return messages.NewPreprepare(sequence, round, proposal)
}

func createPrepareMsg(signer *ecdsa.PrivateKey, header *types.Header, sequence, round *big.Int, digest []byte) *messages.Prepare {
	blsKey, _ := bls.DeriveFromECDSA(signer)
	seal := blsKey.Sign(PrepareSeal(header, uint32(round.Uint64()), SealTypePrepare)).Marshal()
	prepare := messages.NewPrepare(sequence, round, common.BytesToHash(digest), seal)
	prepare.SetSource(crypto.PubkeyToAddress(signer.PublicKey))
	return prepare
}

func createCommitMsg(signer *ecdsa.PrivateKey, header *types.Header, sequence, round *big.Int, digest []byte) *messages.Commit {
	blsKey, _ := bls.DeriveFromECDSA(signer)
	seal := blsKey.Sign(PrepareSeal(header, uint32(round.Uint64()), SealTypeCommit)).Marshal()
	commit := messages.NewCommit(sequence, round, common.BytesToHash(digest), seal)
	commit.SetSource(crypto.PubkeyToAddress(signer.PublicKey))
	return commit
}

func makeProposal(blockNumber *big.Int) *types.Block {
	header := &types.Header{
		ParentHash: common.BytesToHash([]byte("parentBlockHash")),
		Number:     blockNumber,
		GasLimit:   0,
		GasUsed:    0,
		Time:       uint64(12345678),
		Difficulty: types.QBFTDefaultDifficulty,
	}
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}
