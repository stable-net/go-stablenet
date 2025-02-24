package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
)

// addToExtraSeal adds the message to extraSeals which is read when making block
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) error {
	logger := c.currentLogger(true, msg)
	var (
		block    *types.Block
		ok       bool
		sealType SealType
	)

	if c.state == StateAcceptRequest {
		block, ok = c.priorState.Proposal().(*types.Block)
	} else {
		block, ok = c.current.Proposal().(*types.Block)
	}
	if !ok {
		// ignore if block is not found
		return nil
	}

	// validate seal
	if prepareMsg, ok := msg.(*qbftmessage.Prepare); ok {
		sealType = SealTypePrepare
		// Check digest
		if prepareMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid extra PREPARE message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), sealType,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return errInvalidSeal
		}
		// store seal
		c.extraSealsMu.Lock()
		defer c.extraSealsMu.Unlock()
		if c.prepareExtraSeals[msg.Source()] != nil {
			if existingView, incomingView := c.prepareExtraSeals[msg.Source()].View(), prepareMsg.View(); existingView.Cmp(&incomingView) >= 0 {
				return nil
			}
		}
		c.prepareExtraSeals[msg.Source()] = prepareMsg
		logger.Debug("QBFT: new extra prepare seal message")
	} else if commitMsg, ok := msg.(*qbftmessage.Commit); ok {
		sealType = SealTypeCommit
		// Check digest
		if commitMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid extra COMMIT message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), sealType,
			commitMsg.CommitSeal, commitMsg.Source()); err != nil {
			return errInvalidSeal
		}
		// store seal
		c.extraSealsMu.Lock()
		defer c.extraSealsMu.Unlock()
		if c.commitExtraSeals[msg.Source()] != nil {
			if existingView, incomingView := c.commitExtraSeals[msg.Source()].View(), commitMsg.View(); existingView.Cmp(&incomingView) >= 0 {
				return nil
			}
		}
		c.commitExtraSeals[msg.Source()] = commitMsg
		logger.Debug("QBFT: new extra commit seal message")
	} else {
		return errInvalidExtraSealMessage
	}

	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal qbft.Proposal, priorRound *big.Int) (map[common.Hash][]byte, map[common.Hash][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	preparedSeal := make(map[common.Hash][]byte)
	committedSeal := make(map[common.Hash][]byte)
	// latestView is view to process extra seal
	latestView := qbft.View{
		Round:    priorRound,
		Sequence: lastProposal.Number(),
	}

	// process prepare seal
	for addr, msg := range c.prepareExtraSeals {
		if msg != nil {
			view := msg.View()
			if latestView.Cmp(&view) == 0 && msg.Digest == lastProposal.Hash() {
				// this seal(c.prepareExtraSeals[addr]) is valid and re-usable for this sequence
				preparedSeal[common.BytesToHash(msg.PrepareSeal[:])] = msg.PrepareSeal[:]
			} else {
				delete(c.prepareExtraSeals, addr) // erase invalid seal
			}
		}
	}

	// process commit seal
	for addr, msg := range c.commitExtraSeals {
		if msg != nil {
			view := msg.View()
			if latestView.Cmp(&view) == 0 && msg.Digest == lastProposal.Hash() {
				// this seal(c.commitExtraSeals[addr]) is valid and re-usable for this sequence
				committedSeal[common.BytesToHash(msg.CommitSeal[:])] = msg.CommitSeal[:]
			} else {
				delete(c.commitExtraSeals, addr) // erase invalid seal
			}
		}
	}

	return preparedSeal, committedSeal
}
