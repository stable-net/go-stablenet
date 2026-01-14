//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && !blst_disabled

package blst_test

import (
	"crypto/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/bls/blst"
	"github.com/stretchr/testify/require"
)

func TestBLS_Stress_Concurrency(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	const (
		GoroutineCount = 500
		LoopPerRoutine = 50
		AggregateSize  = 16 // max aggregated BLS signatures in test scenario
	)

	// ---------------------------------------------------------
	// 1. Prepare test data
	// ---------------------------------------------------------
	t.Log("!!! Preparing test data ")

	pubKeyBytes := make([][]byte, AggregateSize)
	sigBytes := make([][]byte, AggregateSize)

	msg := []byte("Stress Test Payload")

	for i := 0; i < AggregateSize; i++ {
		priv, err := blst.RandKey()
		require.NoError(t, err)

		pubKeyBytes[i] = priv.PublicKey().Marshal()
		sigBytes[i] = priv.Sign(msg).Marshal()
	}

	// ---------------------------------------------------------
	// 2. Start concurrency test
	// ---------------------------------------------------------
	var wg sync.WaitGroup
	wg.Add(GoroutineCount)

	t.Logf(" !!! Starting Concurrency Stress Test: %d goroutines...", GoroutineCount)
	start := time.Now()

	for i := 0; i < GoroutineCount; i++ {
		go func(id int) {
			defer wg.Done()

			// Panic detection (Defend against crashes caused by CGO memory errors)
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("CRITICAL: Panic in routine %d: %v", id, r)
				}
			}()

			for j := 0; j < LoopPerRoutine; j++ {
				subsetLen := (j % AggregateSize) + 1
				currentPubKeyBytes := pubKeyBytes[:subsetLen]
				currentSigBytes := sigBytes[:subsetLen]

				// Force re-parsing by periodically purging the cache
				if j%5 == 0 {
					blst.PurgePubKeyCache()
				}

				// ==========================================================
				// [Scenario A] Block generation: Aggregation of compressed signatures
				// ==========================================================
				aggSig, err := blst.AggregateCompressedSignatures(currentSigBytes)
				if err != nil {
					t.Errorf("Routine %d: AggregateCompressedSignatures failed: %v", id, err)
					return
				}

				// ==========================================================
				// [Scenario B] Block verification: Parsing public key bytes & aggregation -> signature verification
				// ==========================================================

				// 1. Aggregate public keys
				aggPubKey, err := blst.AggregatePublicKeys(currentPubKeyBytes)
				if err != nil {
					t.Errorf("Routine %d: AggregatePublicKeys failed: %v", id, err)
					return
				}

				// 2. Verify aggregated signature
				if !aggSig.Verify(aggPubKey, msg) {
					t.Errorf("Routine %d: Verify failed (SubsetLen: %d)", id, subsetLen)
					return
				}

				// ==========================================================
				// [Scenario C] Network attack defense: Check if panic occurs when malicious node sends invalid bytes
				// ==========================================================
				trash := make([]byte, 96)

				if _, err := rand.Read(trash); err != nil {
					t.Errorf("Routine %d: rand.Read failed: %v", id, err)
					return
				}

				blst.SignatureFromBytes(trash)
			}
		}(i)
	}

	wg.Wait()
	t.Logf(" !!! Passed Concurrency Stress Test in %v", time.Since(start))
}
