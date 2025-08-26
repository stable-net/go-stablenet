// Copyright 2025 The go-wemix-wbft Authors
// This file is sourced from the Prysm project, licensed under the GPLv3.
// Original source: https://github.com/OffchainLabs/prysm/blob/develop/crypto/bls/common/constants.go
// Copyright The Prysm Authors.

package common

const (
	BLS_SECRET_KEY_LENGTH = 32
	BLS_PUBLIC_KEY_LENGTH = 48
	BLS_SIGNATURE_LENGTH  = 96
)

var (
	// ZeroSecretKey represents a zero secret key.
	ZeroSecretKey = [32]byte{}
	// InfinitePublicKey represents an infinite public key (G1 Point at Infinity).
	InfinitePublicKey = [BLS_PUBLIC_KEY_LENGTH]byte{0xC0}
	// InfiniteSignature represents an infinite signature (G2 Point at Infinity).
	InfiniteSignature = [96]byte{0xC0}
)
