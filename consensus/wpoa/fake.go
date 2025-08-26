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

package wpoa

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/wemixgov"
)

type WemixFakePoA struct {
	wpoa   *WemixPoA
	prvKey *ecdsa.PrivateKey
}

func NewWemixFakeEngine(prvKey *ecdsa.PrivateKey, backend wemixgov.GovBackend) consensus.Engine {
	wkpoa := &WemixFakePoA{
		wpoa: &WemixPoA{
			govCli: NewWemixGov(backend),
		},
		prvKey: prvKey,
	}

	return wkpoa
}

func (wfpoa *WemixFakePoA) Author(header *types.Header) (common.Address, error) {
	return wfpoa.wpoa.Author(header)
}

func (wfpoa *WemixFakePoA) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return wfpoa.wpoa.VerifyHeader(chain, header)
}

func (wfpoa *WemixFakePoA) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	return wfpoa.wpoa.VerifyHeaders(chain, headers)
}

func (wfpoa *WemixFakePoA) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return wfpoa.wpoa.VerifyUncles(chain, block)
}

func (wfpoa *WemixFakePoA) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return wfpoa.wpoa.Prepare(chain, header)
}

func (wfpoa *WemixFakePoA) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) error {
	return wfpoa.wpoa.Finalize(chain, header, state, txs, uncles, withdrawals)
}

func (wfpoa *WemixFakePoA) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	header.Coinbase = crypto.PubkeyToAddress(wfpoa.prvKey.PublicKey)
	wfpoa.wpoa.Finalize(chain, header, state, txs, uncles, withdrawals)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	var err error
	header.MinerNodeSig, err = crypto.Sign(crypto.Keccak256(append(header.Number.Bytes(), header.Root.Bytes()...)), wfpoa.prvKey)
	if err != nil {
		return nil, err
	}
	return types.NewBlock(header, txs, uncles, receipts, trie.NewStackTrie(nil)), nil
}

func (wfpoa *WemixFakePoA) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := types.CopyHeader(block.Header())
	nonce := block.NumberU64()
	header.Nonce = types.EncodeNonce(nonce)

	hash := sealHash(block.Header()).Bytes()
	digest, _ := hashimeta(hash, nonce)
	header.MixDigest = common.BytesToHash(digest)
	go func() { results <- block.WithSeal(header) }()
	return nil
}

func (wfpoa *WemixFakePoA) SealHash(header *types.Header) common.Hash {
	return wfpoa.wpoa.SealHash(header)
}

func (wfpoa *WemixFakePoA) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return wfpoa.wpoa.CalcDifficulty(chain, time, parent)
}

func (wfpoa *WemixFakePoA) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return wfpoa.wpoa.APIs(chain)
}

func (wfpoa *WemixFakePoA) Close() error {
	return wfpoa.wpoa.Close()
}

func (wfpoa *WemixFakePoA) CallEngineSpecific(method string, args ...interface{}) interface{} {
	return nil
}
