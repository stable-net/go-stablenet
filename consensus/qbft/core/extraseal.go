package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
)

// addToExtraSeal adds a seal received after consensus to extraSeals.
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) error {
	logger := c.currentLogger(true, msg)
	var (
		block    *types.Block
		ok       bool
		sealType SealType
		valSet   qbft.ValidatorSet
	)

	if c.state == StateAcceptRequest {
		block, ok = c.priorState.Proposal().(*types.Block)
		valSet = c.priorState.Validators()
	} else {
		block, ok = c.current.Proposal().(*types.Block)
		valSet = c.valSet
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
		if err := verifySeal(valSet, block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), sealType,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return err
		}

		c.addToPrepareExtraSeal(msg.(*qbftmessage.Prepare))
	} else if commitMsg, ok := msg.(*qbftmessage.Commit); ok {
		sealType = SealTypeCommit
		// Check digest
		if commitMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid extra COMMIT message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(valSet, block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), sealType,
			commitMsg.CommitSeal, commitMsg.Source()); err != nil {
			return err
		}

		c.addToCommitExtraSeal(msg.(*qbftmessage.Commit))
	} else {
		return errInvalidExtraSealMessage
	}
	return nil
}

func (c *Core) addToPrepareExtraSeal(prepareMsg *qbftmessage.Prepare) {
	logger := c.currentLogger(true, prepareMsg)

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	if c.prepareExtraSeals[prepareMsg.Source()] != nil {
		if existingView, incomingView := c.prepareExtraSeals[prepareMsg.Source()].View(), prepareMsg.View(); existingView.Cmp(&incomingView) >= 0 {
			return
		}
	}
	c.prepareExtraSeals[prepareMsg.Source()] = prepareMsg
	logger.Debug("QBFT: new extra prepare seal message")
}

func (c *Core) addToCommitExtraSeal(commitMsg *qbftmessage.Commit) {
	logger := c.currentLogger(true, commitMsg)

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	if c.commitExtraSeals[commitMsg.Source()] != nil {
		if existingView, incomingView := c.commitExtraSeals[commitMsg.Source()].View(), commitMsg.View(); existingView.Cmp(&incomingView) >= 0 {
			return
		}
	}
	c.commitExtraSeals[commitMsg.Source()] = commitMsg
	logger.Debug("QBFT: new extra commit seal message")
}

// addEffectiveSealToExtraSeal adds a consensus-effective seal to extraSeals used during block creation.
func (c *Core) addEffectiveSealToExtraSeal() error {
	// c.current may be nil on the first function call after system boot.
	if c.current != nil {
		for _, m := range c.current.QBFTPrepares.Values() {
			c.addToPrepareExtraSeal(m.(*qbftmessage.Prepare))
		}
		for _, m := range c.current.QBFTCommits.Values() {
			c.addToCommitExtraSeal(m.(*qbftmessage.Commit))
		}
	}
	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal qbft.Proposal, priorRound *big.Int, valSet qbft.ValidatorSet) ([]qbft.SealData, []qbft.SealData) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	preparedSeal := make([]qbft.SealData, 0)
	committedSeal := make([]qbft.SealData, 0)
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
				idx, _ := valSet.GetByAddress(msg.Source())
				if idx < 0 {
					continue
				}
				preparedSeal = append(preparedSeal, qbft.SealData{
					Sealer: uint32(idx),
					Seal:   append([]byte{}, msg.PrepareSeal...),
				})
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
				idx, _ := valSet.GetByAddress(msg.Source())
				if idx < 0 {
					continue
				}
				committedSeal = append(committedSeal, qbft.SealData{
					Sealer: uint32(idx),
					Seal:   append([]byte{}, msg.CommitSeal...),
				})
			} else {
				delete(c.commitExtraSeals, addr) // erase invalid seal
			}
		}
	}

	return preparedSeal, committedSeal
}
