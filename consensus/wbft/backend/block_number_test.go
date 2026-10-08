// Copyright 2026 The go-stablenet Authors
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

package backend

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
)

// makeProposalWithNumber builds an unsealed proposal on top of genesis whose
// header number is overridden, with the randao reveal signed for that number
// as a proposer would.
func makeProposalWithNumber(t *testing.T, chain *core.BlockChain, engine *Backend, number *big.Int) *types.Block {
	t.Helper()

	// Drive the core into a new round so Prepare can sign the header.
	makeBlockWithoutSeal(chain, engine, chain.Genesis())

	header := makeHeader(chain.Config(), engine.config, chain.Genesis())
	header.Number = number
	if err := engine.Prepare(chain, header); err != nil {
		t.Fatalf("prepare header %s: %v", number, err)
	}
	state, err := chain.State()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	block, err := engine.FinalizeAndAssemble(chain, header, state, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("assemble block %s: %v", number, err)
	}
	return block
}

// TestVerifyRejectsOversizedBlockNumber verifies that a block whose number
// only matches the expected height in its low 64 bits is rejected both as a
// consensus proposal and as an imported header.
func TestVerifyRejectsOversizedBlockNumber(t *testing.T) {
	two64 := new(big.Int).Lsh(common.Big1, 64)

	tests := []struct {
		name   string
		number *big.Int
	}{
		{"2^64 + 1", new(big.Int).Add(two64, common.Big1)},
		{"2^65 + 1", new(big.Int).Add(new(big.Int).Lsh(two64, 1), common.Big1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, engine, _ := newBlockChain(1)
			defer engine.Stop()

			block := makeProposalWithNumber(t, chain, engine, tt.number)

			if _, err := engine.Verify(block); !errors.Is(err, wbftcommon.ErrInvalidBlockNumber) {
				t.Errorf("Verify(number=%s) = %v; want %v", tt.number, err, wbftcommon.ErrInvalidBlockNumber)
			}
			if err := engine.VerifyHeader(chain, block.Header()); !errors.Is(err, wbftcommon.ErrInvalidBlockNumber) {
				t.Errorf("VerifyHeader(number=%s) = %v; want %v", tt.number, err, wbftcommon.ErrInvalidBlockNumber)
			}
		})
	}
}

// TestVerifyAcceptsNextBlockNumber guards against the number check rejecting
// an honest proposal built the same way as the oversized ones above.
func TestVerifyAcceptsNextBlockNumber(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()

	block := makeProposalWithNumber(t, chain, engine, big.NewInt(1))
	if _, err := engine.Verify(block); err != nil {
		t.Fatalf("Verify(number=1) = %v; want nil", err)
	}
}
