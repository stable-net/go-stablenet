// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

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
