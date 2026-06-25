package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftmsg "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestIsSequenceTooFarAhead(t *testing.T) {
	c := &Core{}
	threshold := int64(1)

	tests := []struct {
		viewSeq  int64
		currSeq  int64
		expected bool
	}{
		{100, 100, false}, // same
		{101, 100, false}, // next (diff=1) - SHOULD NOT BE TOO FAR
		{102, 100, true},  // diff=2 - TOO FAR
	}

	for _, tt := range tests {
		_, tooFar := c.isSequenceTooFarAhead(big.NewInt(tt.viewSeq), big.NewInt(tt.currSeq), threshold)
		if tooFar != tt.expected {
			t.Errorf("isSequenceTooFarAhead(%d, %d, %d) = %v; want %v", tt.viewSeq, tt.currSeq, threshold, tooFar, tt.expected)
		}
	}
}

func TestIsRoundTooFarAhead(t *testing.T) {
	c := &Core{}
	threshold := int64(10)

	tests := []struct {
		viewRound int64
		currRound int64
		expected  bool
	}{
		{0, 0, false},
		{5, 0, false},
		{10, 0, false}, // diff=10 - SHOULD NOT BE TOO FAR
		{11, 0, true},  // diff=11 - TOO FAR
	}

	for _, tt := range tests {
		_, tooFar := c.isRoundTooFarAhead(big.NewInt(tt.viewRound), big.NewInt(tt.currRound), threshold)
		if tooFar != tt.expected {
			t.Errorf("isRoundTooFarAhead(%d, %d, %d) = %v; want %v", tt.viewRound, tt.currRound, threshold, tooFar, tt.expected)
		}
	}
}

// TestIsTooFarFutureMessageRoundThreshold verifies that future-sequence messages
// with a round exceeding roundThreshold are rejected, while same-sequence messages
// use a relative round check against the current round.
func TestIsTooFarFutureMessageRoundThreshold(t *testing.T) {
	// Current view: seq=100, round=0
	core := makeCoreForTest(common.Big0, common.Big0, big.NewInt(100), nil)

	tests := []struct {
		name     string
		seq      int64
		round    int64
		expected bool
	}{
		// Future sequence (n+1): absolute round check from 0
		{"future seq, round below threshold", 101, roundThreshold - 1, false},
		{"future seq, round at threshold", 101, roundThreshold, true},
		{"future seq, round exceeds threshold", 101, roundThreshold + 1, true},
		{"future seq, round way over threshold", 101, 1000000, true},
		// Same sequence: relative round check from current round (0)
		{"same seq, round at threshold", 100, roundThreshold, false},
		{"same seq, round exceeds threshold", 100, roundThreshold + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &wbft.View{
				Sequence: big.NewInt(tt.seq),
				Round:    big.NewInt(tt.round),
			}
			result := core.isTooFarFutureMessage(view)
			if result != tt.expected {
				t.Errorf("isTooFarFutureMessage(seq=%d, round=%d) = %v; want %v",
					tt.seq, tt.round, result, tt.expected)
			}
		})
	}
}

// TestAddToBacklogSizeLimit verifies that messages beyond maxBacklogSizePerValidator
// are dropped for a given source address.
func TestAddToBacklogSizeLimit(t *testing.T) {
	proposal := makeProposal(big.NewInt(100))
	core := makeCoreForTest(common.Big0, common.Big0, big.NewInt(101), proposal)

	src := crypto.PubkeyToAddress(signers[1].PublicKey)
	seq := big.NewInt(101)

	// Fill the backlog to the limit using distinct rounds to avoid dedup.
	for i := 0; i < maxBacklogSizePerValidator; i++ {
		addBacklogMsg(t, core, wbftmsg.PrepareCode, 1, proposal, seq, big.NewInt(int64(i)))
	}

	if core.backlogs[src].Size() != maxBacklogSizePerValidator {
		t.Fatalf("expected backlog size %d, got %d", maxBacklogSizePerValidator, core.backlogs[src].Size())
	}

	// One more message should be dropped.
	addBacklogMsg(t, core, wbftmsg.PrepareCode, 1, proposal, seq, big.NewInt(maxBacklogSizePerValidator))

	if core.backlogs[src].Size() != maxBacklogSizePerValidator {
		t.Errorf("expected backlog size to remain %d after overflow, got %d",
			maxBacklogSizePerValidator, core.backlogs[src].Size())
	}
}

// TestAddToBacklogDedup verifies that a second message with the same
// (code, sequence, round) from the same validator is dropped regardless of payload.
func TestAddToBacklogDedup(t *testing.T) {
	proposal := makeProposal(big.NewInt(100))
	otherProposal := makeProposal(big.NewInt(99)) // distinct proposal to produce a different digest
	core := makeCoreForTest(common.Big0, common.Big0, big.NewInt(101), proposal)

	tests := []struct {
		name      string
		signerIdx int
		proposal  *types.Block
		seq       int64
		round     int64
		code      uint64
		wantSize  int // expected backlog size for this signer after adding
	}{
		{
			name:      "first message",
			signerIdx: 1, proposal: proposal, seq: 101, round: 5, code: wbftmsg.PrepareCode,
			wantSize: 1,
		},
		{
			name:      "duplicate same slot different digest - dropped",
			signerIdx: 1, proposal: otherProposal, seq: 101, round: 5, code: wbftmsg.PrepareCode,
			wantSize: 1,
		},
		{
			name:      "same validator different round - accepted",
			signerIdx: 1, proposal: proposal, seq: 101, round: 6, code: wbftmsg.PrepareCode,
			wantSize: 2,
		},
		{
			name:      "same validator same slot different code - accepted",
			signerIdx: 1, proposal: proposal, seq: 101, round: 5, code: wbftmsg.CommitCode,
			wantSize: 3,
		},
		{
			name:      "different validator same slot - accepted",
			signerIdx: 2, proposal: proposal, seq: 101, round: 5, code: wbftmsg.PrepareCode,
			wantSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := crypto.PubkeyToAddress(signers[tt.signerIdx].PublicKey)
			addBacklogMsg(t, core, tt.code, tt.signerIdx, tt.proposal, big.NewInt(tt.seq), big.NewInt(tt.round))

			if core.backlogs[src].Size() != tt.wantSize {
				t.Errorf("backlog size = %d; want %d", core.backlogs[src].Size(), tt.wantSize)
			}
		})
	}
}

// TestProcessBacklogPurgesOnValidatorRemoval verifies that both the backlog queue
// and backlogKeys are fully purged for a validator that has left the validator set.
func TestProcessBacklogPurgesOnValidatorRemoval(t *testing.T) {
	proposal := makeProposal(big.NewInt(100))
	core := makeCoreForTest(common.Big0, common.Big0, big.NewInt(101), proposal)

	src := crypto.PubkeyToAddress(signers[1].PublicKey)
	remaining := crypto.PubkeyToAddress(signers[2].PublicKey)
	seq := big.NewInt(101)

	// Add several messages from signers[1] (to be removed) and signers[2] (to remain).
	for _, round := range []int64{1, 2, 3} {
		addBacklogMsg(t, core, wbftmsg.PrepareCode, 1, proposal, seq, big.NewInt(round))
	}
	addBacklogMsg(t, core, wbftmsg.PrepareCode, 2, proposal, seq, big.NewInt(1))

	if core.backlogs[src].Size() != 3 {
		t.Fatalf("expected 3 messages in backlog before removal, got %d", core.backlogs[src].Size())
	}
	if len(core.backlogKeys[src]) != 3 {
		t.Fatalf("expected 3 keys before removal, got %d", len(core.backlogKeys[src]))
	}

	// Remove signers[1] from the validator set.
	core.valSet.RemoveValidator(src)

	core.processBacklog()

	// Removed validator's backlog and keys must be purged.
	if _, exists := core.backlogs[src]; exists {
		t.Errorf("expected backlogs entry for removed validator to be deleted")
	}
	if len(core.backlogKeys[src]) != 0 {
		t.Errorf("expected backlogKeys for removed validator to be empty, got %d entries",
			len(core.backlogKeys[src]))
	}

	// Remaining validator's backlog and keys must be intact.
	if core.backlogs[remaining] == nil || core.backlogs[remaining].Size() != 1 {
		t.Errorf("expected remaining validator backlog to be intact with 1 message")
	}
	if len(core.backlogKeys[remaining]) != 1 {
		t.Errorf("expected remaining validator backlogKeys to be intact with 1 key, got %d",
			len(core.backlogKeys[remaining]))
	}
}

// TestProcessBacklogRemovesKeysOnDrop verifies that backlogKeys are removed when
// a backlogged message is dropped because it has become expired or invalid.
func TestProcessBacklogRemovesKeysOnDrop(t *testing.T) {
	proposal := makeProposal(big.NewInt(100))
	// Core at seq=101, round=0
	core := makeCoreForTest(common.Big0, common.Big0, big.NewInt(101), proposal)

	seq := big.NewInt(101)
	round := big.NewInt(5)
	src := crypto.PubkeyToAddress(signers[1].PublicKey)

	// Add a future-round message to the backlog.
	addBacklogMsg(t, core, wbftmsg.PrepareCode, 1, proposal, seq, round)

	if _, exists := core.backlogKeys[src][backlogKey{wbftmsg.PrepareCode, seq.Uint64(), round.Uint64()}]; !exists {
		t.Fatal("expected key to be present in backlogKeys after addToBacklog")
	}

	// Advance to round=16 (> round=5) so the backlogged message becomes old.
	core.current = newRoundState(
		&wbft.View{Sequence: big.NewInt(101), Round: big.NewInt(16)},
		core.valSet, nil, nil, nil, nil, func(hash common.Hash) bool { return false },
	)

	core.processBacklog()

	if len(core.backlogKeys[src]) != 0 {
		t.Errorf("expected backlogKeys to be empty after processBacklog, got %d entries",
			len(core.backlogKeys[src]))
	}
}

// addBacklogMsg creates a message of the given code from signers[signerIdx]
// and adds it to the backlog.
func addBacklogMsg(t *testing.T, core *Core, code uint64, signerIdx int, proposal *types.Block, seq, round *big.Int) {
	t.Helper()
	switch code {
	case wbftmsg.PrepareCode:
		core.addToBacklog(createPrepareMsg(signers[signerIdx], proposal.Header(), seq, round, proposal.Hash().Bytes()))
	case wbftmsg.CommitCode:
		core.addToBacklog(createCommitMsg(signers[signerIdx], proposal.Header(), seq, round, proposal.Hash().Bytes()))
	default:
		t.Fatalf("addBacklogMsg: unsupported message code %d", code)
	}
}
