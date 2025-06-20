package core

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/consensus/wbft"
)

// priorState stores prior consensus state
// for collecting extra seals during StateAcceptRequest
type priorState struct {
	mu           *sync.RWMutex
	round        *big.Int
	proposal     wbft.Proposal
	validatorSet wbft.ValidatorSet
}

func (c *Core) updatePriorState() {
	c.priorState.mu.Lock()
	defer c.priorState.mu.Unlock()
	c.priorState.round = c.current.Round()
	if c.current.Proposal() != nil {
		c.priorState.proposal = c.current.Proposal()
	}
	if c.valSet != nil {
		c.priorState.validatorSet = c.valSet
	}
}

func (p *priorState) Round() *big.Int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.round
}

func (p *priorState) Proposal() wbft.Proposal {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.proposal
}

func (p *priorState) Validators() wbft.ValidatorSet {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.validatorSet
}
