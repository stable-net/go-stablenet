// Copyright 2025 The go-wemix-wbft Authors
// This file is sourced from the Prysm project, licensed under the GPLv3.
// Original source: https://github.com/OffchainLabs/prysm/blob/develop/crypto/bls/blst/init.go
// Copyright The Prysm Authors.

//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && !blst_disabled

package blst

import (
	"runtime"

	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/crypto/bls/common"
	blst "github.com/supranational/blst/bindings/go"
)

func init() {
	// Reserve 1 core for general application work
	maxProcs := runtime.GOMAXPROCS(0) - 1
	if maxProcs <= 0 {
		maxProcs = 1
	}
	blst.SetMaxProcs(maxProcs)
	pubkeyCache = lru.NewCache[[common.BLS_PUBLIC_KEY_LENGTH]byte, common.PublicKey](maxKeys)
}
