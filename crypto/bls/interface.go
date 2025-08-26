// This file is sourced from the Prysm project, licensed under the GPLv3.
// Original source: https://github.com/OffchainLabs/prysm/blob/develop/crypto/bls/interface.go
// Copyright The Prysm Authors.

package bls

import "github.com/ethereum/go-ethereum/crypto/bls/common"

// PublicKey represents a BLS public key.
type PublicKey = common.PublicKey

// SecretKey represents a BLS secret or private key.
type SecretKey = common.SecretKey

// Signature represents a BLS signature.
type Signature = common.Signature
