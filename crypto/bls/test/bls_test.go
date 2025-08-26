// Copyright 2025 The go-wemix-wbft Authors
// This file is sourced from the Prysm project, licensed under the GPLv3.
// Original source: https://github.com/OffchainLabs/prysm/blob/develop/crypto/bls/bls_test.go
// Copyright The Prysm Authors.

package test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/crypto/bls/common"
	"github.com/stretchr/testify/require"
)

func TestDeriveFromECDSA(t *testing.T) {
	privECDSA, err := crypto.GenerateKey()
	require.NoError(t, err, "Failed to generate ECDSA private key")

	deriveFirst, err := bls.DeriveFromECDSA(privECDSA)
	require.NoError(t, err)
	deriveSecond, err := bls.DeriveFromECDSA(privECDSA)
	require.NoError(t, err)

	require.Equal(t, deriveFirst.Marshal(), deriveSecond.Marshal())
}

func TestDisallowZeroSecretKeys(t *testing.T) {
	t.Run("blst", func(t *testing.T) {
		// Blst does a zero check on the key during deserialization.
		_, err := bls.SecretKeyFromBytes(common.ZeroSecretKey[:])
		require.Equal(t, common.ErrSecretUnmarshal, err)
	})
}

func TestDisallowZeroPublicKeys(t *testing.T) {
	t.Run("blst", func(t *testing.T) {
		_, err := bls.PublicKeyFromBytes(common.InfinitePublicKey[:])
		require.Equal(t, common.ErrInfinitePubKey, err)
	})
}

func TestDisallowZeroPublicKeys_AggregatePubkeys(t *testing.T) {
	t.Run("blst", func(t *testing.T) {
		_, err := bls.AggregatePublicKeys([][]byte{common.InfinitePublicKey[:], common.InfinitePublicKey[:]})
		require.Equal(t, common.ErrInfinitePubKey, err)
	})
}
