package test

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/crypto/bls/common"
	"github.com/stretchr/testify/require"
)

const TestSignature = "test signature"

func TestCopySignatureSet(t *testing.T) {
	t.Run("blst", func(t *testing.T) {
		key, err := bls.RandKey()
		require.NoError(t, err)
		key2, err := bls.RandKey()
		require.NoError(t, err)
		key3, err := bls.RandKey()
		require.NoError(t, err)

		message := [32]byte{'C', 'D'}
		message2 := [32]byte{'E', 'F'}
		message3 := [32]byte{'H', 'I'}

		sig := key.Sign(message[:])
		sig2 := key2.Sign(message2[:])
		sig3 := key3.Sign(message3[:])

		set := &bls.SignatureBatch{
			Signatures:   [][]byte{sig.Marshal()},
			PublicKeys:   []bls.PublicKey{key.PublicKey()},
			Messages:     [][32]byte{message},
			Descriptions: createDescriptions(1),
		}
		set2 := &bls.SignatureBatch{
			Signatures:   [][]byte{sig2.Marshal()},
			PublicKeys:   []bls.PublicKey{key.PublicKey()},
			Messages:     [][32]byte{message},
			Descriptions: createDescriptions(1),
		}
		set3 := &bls.SignatureBatch{
			Signatures:   [][]byte{sig3.Marshal()},
			PublicKeys:   []bls.PublicKey{key.PublicKey()},
			Messages:     [][32]byte{message},
			Descriptions: createDescriptions(1),
		}
		aggSet := set.Join(set2).Join(set3)
		aggSet2 := aggSet.Copy()

		require.Equal(t, aggSet, aggSet2)
	})
}

func TestVerifyVerbosely_AllSignaturesValid(t *testing.T) {
	set := NewValidSignatureSet(t, "good", 3)
	valid, err := set.VerifyVerbosely()
	require.NoError(t, err)
	require.True(t, valid, "SignatureSet is expected to be valid")
}

func TestVerifyVerbosely_SomeSignaturesInvalid(t *testing.T) {
	goodSet := NewValidSignatureSet(t, "good", 3)
	badSet := NewInvalidSignatureSet(t, "bad", 3, false)
	set := bls.NewSet().Join(goodSet).Join(badSet)
	valid, err := set.VerifyVerbosely()
	require.False(t, valid, "SignatureSet is expected to be invalid")
	require.Contains(t, err.Error(), "signature 'signature of bad0' is invalid")
	require.Contains(t, err.Error(), "signature 'signature of bad1' is invalid")
	require.Contains(t, err.Error(), "signature 'signature of bad2' is invalid")
	require.NotContains(t, err.Error(), "signature 'signature of good0' is invalid")
	require.NotContains(t, err.Error(), "signature 'signature of good1' is invalid")
	require.NotContains(t, err.Error(), "signature 'signature of good2' is invalid")
}

func TestVerifyVerbosely_VerificationThrowsError(t *testing.T) {
	goodSet := NewValidSignatureSet(t, "good", 1)
	badSet := NewInvalidSignatureSet(t, "bad", 1, true)
	set := bls.NewSet().Join(goodSet).Join(badSet)
	valid, err := set.VerifyVerbosely()
	require.False(t, valid, "SignatureSet is expected to be invalid")
	require.Contains(t, err.Error(), "signature 'signature of bad0' is invalid")
	require.Contains(t, err.Error(), "could not unmarshal bytes into signature")
	require.NotContains(t, err.Error(), "signature 'signature of good0' is invalid")
}

func TestSignatureBatch_RemoveDuplicates(t *testing.T) {
	var keys []bls.SecretKey
	for i := 0; i < 100; i++ {
		key, err := bls.RandKey()
		require.NoError(t, err)
		keys = append(keys, key)
	}
	tests := []struct {
		name         string
		batchCreator func() (input *bls.SignatureBatch, output *bls.SignatureBatch)
		want         int
	}{
		{
			name: "empty batch",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				return &bls.SignatureBatch{}, &bls.SignatureBatch{}
			},
			want: 0,
		},
		{
			name: "valid duplicates in batch",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:20]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				allSigs := append(signatures, signatures...)
				allPubs := append(pubs, pubs...)
				allMsgs := append(messages, messages...)
				return &bls.SignatureBatch{
						Signatures:   allSigs,
						PublicKeys:   allPubs,
						Messages:     allMsgs,
						Descriptions: createDescriptions(len(allMsgs)),
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(allMsgs)),
					}
			},
			want: 20,
		},
		{
			name: "valid duplicates in batch with multiple messages",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				allSigs := append(signatures, signatures...)
				allPubs := append(pubs, pubs...)
				allMsgs := append(messages, messages...)
				return &bls.SignatureBatch{
						Signatures:   allSigs,
						PublicKeys:   allPubs,
						Messages:     allMsgs,
						Descriptions: createDescriptions(len(allMsgs)),
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(allMsgs)),
					}
			},
			want: 30,
		},
		{
			name: "no duplicates in batch with multiple messages",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				return &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}
			},
			want: 0,
		},
		{
			name: "valid duplicates and invalid duplicates in batch with multiple messages",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				allSigs := append(signatures, signatures...)
				// Make it a non-unique entry
				allSigs[10] = make([]byte, 96)
				allPubs := append(pubs, pubs...)
				allMsgs := append(messages, messages...)
				// Insert it back at the end
				signatures = append(signatures, signatures[10])
				pubs = append(pubs, pubs[10])
				messages = append(messages, messages[10])
				// Zero out to expected result
				signatures[10] = make([]byte, 96)
				return &bls.SignatureBatch{
						Signatures:   allSigs,
						PublicKeys:   allPubs,
						Messages:     allMsgs,
						Descriptions: createDescriptions(len(allMsgs)),
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(allMsgs)),
					}
			},
			want: 29,
		},
		{
			name: "valid duplicates and invalid duplicates with signature,pubkey,message in batch with multiple messages",
			batchCreator: func() (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				allSigs := append(signatures, signatures...)
				// Make it a non-unique entry
				allSigs[10] = make([]byte, 96)

				allPubs := append(pubs, pubs...)
				allPubs[20] = keys[len(keys)-1].PublicKey()

				allMsgs := append(messages, messages...)
				allMsgs[29] = [32]byte{'j', 'u', 'n', 'k'}

				// Insert it back at the end
				signatures = append(signatures, signatures[10])
				pubs = append(pubs, pubs[10])
				messages = append(messages, messages[10])
				// Zero out to expected result
				signatures[10] = make([]byte, 96)

				// Insert it back at the end
				signatures = append(signatures, signatures[20])
				pubs = append(pubs, pubs[20])
				messages = append(messages, messages[20])
				// Zero out to expected result
				pubs[20] = keys[len(keys)-1].PublicKey()

				// Insert it back at the end
				signatures = append(signatures, signatures[29])
				pubs = append(pubs, pubs[29])
				messages = append(messages, messages[29])
				messages[29] = [32]byte{'j', 'u', 'n', 'k'}

				return &bls.SignatureBatch{
						Signatures:   allSigs,
						PublicKeys:   allPubs,
						Messages:     allMsgs,
						Descriptions: createDescriptions(len(allMsgs)),
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}
			},
			want: 27,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, output := tt.batchCreator()
			num, res, err := input.RemoveDuplicates()
			require.NoError(t, err)
			if num != tt.want {
				t.Errorf("RemoveDuplicates() got = %v, want %v", num, tt.want)
			}
			if !reflect.DeepEqual(res.Signatures, output.Signatures) {
				t.Errorf("RemoveDuplicates() Signatures output = %v, want %v", res.Signatures, output.Signatures)
			}
			if !reflect.DeepEqual(res.PublicKeys, output.PublicKeys) {
				t.Errorf("RemoveDuplicates() Publickeys output = %v, want %v", res.PublicKeys, output.PublicKeys)
			}
			if !reflect.DeepEqual(res.Messages, output.Messages) {
				t.Errorf("RemoveDuplicates() Messages output = %v, want %v", res.Messages, output.Messages)
			}
		})
	}
}

func TestSignatureBatch_AggregateBatch(t *testing.T) {
	var keys []bls.SecretKey
	for i := 0; i < 100; i++ {
		key, err := bls.RandKey()
		require.NoError(t, err)
		keys = append(keys, key)
	}
	tests := []struct {
		name         string
		batchCreator func(t *testing.T) (input *bls.SignatureBatch, output *bls.SignatureBatch)
		wantErr      bool
	}{
		{
			name: "empty batch",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				return &bls.SignatureBatch{Signatures: nil, Messages: nil, PublicKeys: nil, Descriptions: nil},
					&bls.SignatureBatch{Signatures: nil, Messages: nil, PublicKeys: nil, Descriptions: nil}
			},
			wantErr: false,
		},
		{
			name: "mismatch number of signatures and messages in batch",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				key1 := keys[0]
				key2 := keys[1]
				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				sig1 := key1.Sign(msg[:])
				sig2 := key2.Sign(msg[:])
				signatures := [][]byte{sig1.Marshal(), sig2.Marshal()}
				pubs := []common.PublicKey{key1.PublicKey(), key2.PublicKey()}
				messages := [][32]byte{msg}
				descs := createDescriptions(2)
				return &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: descs,
					}, &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: descs,
					}
			},
			wantErr: true,
		},
		{
			name: "valid signatures in batch",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:20]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				aggSig, err := bls.AggregateCompressedSignatures(signatures)
				require.NoError(t, err)
				aggPub := bls.AggregateMultiplePubkeys(pubs)
				return &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}, &bls.SignatureBatch{
						Signatures:   [][]byte{aggSig.Marshal()},
						PublicKeys:   []bls.PublicKey{aggPub},
						Messages:     [][32]byte{msg},
						Descriptions: createDescriptions(1, bls.AggregatedSignature),
					}
			},
			wantErr: false,
		},
		{
			name: "invalid signatures in batch",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:20]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				signatures[10] = make([]byte, 96)
				return &bls.SignatureBatch{
					Signatures:   signatures,
					PublicKeys:   pubs,
					Messages:     messages,
					Descriptions: createDescriptions(len(messages)),
				}, nil
			},
			wantErr: true,
		},
		{
			name: "valid aggregates in batch with multiple messages",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				aggSig1, err := bls.AggregateCompressedSignatures(signatures[:10])
				require.NoError(t, err)
				aggSig2, err := bls.AggregateCompressedSignatures(signatures[10:20])
				require.NoError(t, err)
				aggSig3, err := bls.AggregateCompressedSignatures(signatures[20:30])
				require.NoError(t, err)
				aggPub1 := bls.AggregateMultiplePubkeys(pubs[:10])
				aggPub2 := bls.AggregateMultiplePubkeys(pubs[10:20])
				aggPub3 := bls.AggregateMultiplePubkeys(pubs[20:30])
				return &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}, &bls.SignatureBatch{
						Signatures:   [][]byte{aggSig1.Marshal(), aggSig2.Marshal(), aggSig3.Marshal()},
						PublicKeys:   []bls.PublicKey{aggPub1, aggPub2, aggPub3},
						Messages:     [][32]byte{msg, msg1, msg2},
						Descriptions: createDescriptions(3, bls.AggregatedSignature),
					}
			},
			wantErr: false,
		},
		{
			name: "common and uncommon messages in batch with multiple messages",
			batchCreator: func(t *testing.T) (*bls.SignatureBatch, *bls.SignatureBatch) {
				chosenKeys := keys[:30]

				msg := [32]byte{'r', 'a', 'n', 'd', 'o', 'm'}
				msg1 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '1'}
				msg2 := [32]byte{'r', 'a', 'n', 'd', 'o', 'm', '2'}
				var signatures [][]byte
				var messages [][32]byte
				var pubs []bls.PublicKey
				for _, k := range chosenKeys[:10] {
					s := k.Sign(msg[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[10:20] {
					s := k.Sign(msg1[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg1)
					pubs = append(pubs, k.PublicKey())
				}
				for _, k := range chosenKeys[20:30] {
					s := k.Sign(msg2[:])
					signatures = append(signatures, s.Marshal())
					messages = append(messages, msg2)
					pubs = append(pubs, k.PublicKey())
				}
				// Set a custom message
				messages[5][31] ^= byte(100)
				messages[15][31] ^= byte(100)
				messages[25][31] ^= byte(100)

				var newSigs [][]byte
				newSigs = append(newSigs, signatures[:5]...)
				newSigs = append(newSigs, signatures[6:10]...)

				aggSig1, err := bls.AggregateCompressedSignatures(newSigs)
				require.NoError(t, err)

				newSigs = [][]byte{}
				newSigs = append(newSigs, signatures[10:15]...)
				newSigs = append(newSigs, signatures[16:20]...)
				aggSig2, err := bls.AggregateCompressedSignatures(newSigs)
				require.NoError(t, err)

				newSigs = [][]byte{}
				newSigs = append(newSigs, signatures[20:25]...)
				newSigs = append(newSigs, signatures[26:30]...)
				aggSig3, err := bls.AggregateCompressedSignatures(newSigs)
				require.NoError(t, err)

				var newPubs []bls.PublicKey
				newPubs = append(newPubs, pubs[:5]...)
				newPubs = append(newPubs, pubs[6:10]...)

				aggPub1 := bls.AggregateMultiplePubkeys(newPubs)

				newPubs = []bls.PublicKey{}
				newPubs = append(newPubs, pubs[10:15]...)
				newPubs = append(newPubs, pubs[16:20]...)
				aggPub2 := bls.AggregateMultiplePubkeys(newPubs)

				newPubs = []bls.PublicKey{}
				newPubs = append(newPubs, pubs[20:25]...)
				newPubs = append(newPubs, pubs[26:30]...)
				aggPub3 := bls.AggregateMultiplePubkeys(newPubs)

				return &bls.SignatureBatch{
						Signatures:   signatures,
						PublicKeys:   pubs,
						Messages:     messages,
						Descriptions: createDescriptions(len(messages)),
					}, &bls.SignatureBatch{
						Signatures:   [][]byte{aggSig1.Marshal(), signatures[5], aggSig2.Marshal(), signatures[15], aggSig3.Marshal(), signatures[25]},
						PublicKeys:   []bls.PublicKey{aggPub1, pubs[5], aggPub2, pubs[15], aggPub3, pubs[25]},
						Messages:     [][32]byte{msg, messages[5], msg1, messages[15], msg2, messages[25]},
						Descriptions: []string{bls.AggregatedSignature, TestSignature, bls.AggregatedSignature, TestSignature, bls.AggregatedSignature, TestSignature},
					}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, output := tt.batchCreator(t)
			got, err := input.AggregateBatch()
			if (err != nil) != tt.wantErr {
				t.Errorf("AggregateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got = sortSet(got)
			output = sortSet(output)

			if !reflect.DeepEqual(got.Signatures, output.Signatures) {
				t.Errorf("AggregateBatch() Signatures got = %v, want %v", got.Signatures, output.Signatures)
			}
			if !reflect.DeepEqual(got.PublicKeys, output.PublicKeys) {
				t.Errorf("AggregateBatch() PublicKeys got = %v, want %v", got.PublicKeys, output.PublicKeys)
			}
			if !reflect.DeepEqual(got.Messages, output.Messages) {
				t.Errorf("AggregateBatch() Messages got = %v, want %v", got.Messages, output.Messages)
			}
			if !reflect.DeepEqual(got.Descriptions, output.Descriptions) {
				t.Errorf("AggregateBatch() Descriptions got = %v, want %v", got.Descriptions, output.Descriptions)
			}
		})
	}
}

func NewValidSignatureSet(t *testing.T, msgBody string, num int) *bls.SignatureBatch {
	set := &bls.SignatureBatch{
		Signatures:   make([][]byte, num),
		PublicKeys:   make([]common.PublicKey, num),
		Messages:     make([][32]byte, num),
		Descriptions: make([]string, num),
	}

	for i := 0; i < num; i++ {
		priv, err := bls.RandKey()
		require.NoError(t, err)
		pubkey := priv.PublicKey()
		msg := messageBytes(fmt.Sprintf("%s%d", msgBody, i))
		sig := priv.Sign(msg[:]).Marshal()
		desc := fmt.Sprintf("signature of %s%d", msgBody, i)

		set.Signatures[i] = sig
		set.PublicKeys[i] = pubkey
		set.Messages[i] = msg
		set.Descriptions[i] = desc
	}

	return set
}

func NewInvalidSignatureSet(t *testing.T, msgBody string, num int, throwErr bool) *bls.SignatureBatch {
	set := &bls.SignatureBatch{
		Signatures:   make([][]byte, num),
		PublicKeys:   make([]common.PublicKey, num),
		Messages:     make([][32]byte, num),
		Descriptions: make([]string, num),
	}

	for i := 0; i < num; i++ {
		priv, err := bls.RandKey()
		require.NoError(t, err)
		pubkey := priv.PublicKey()
		msg := messageBytes(fmt.Sprintf("%s%d", msgBody, i))
		var sig []byte
		if throwErr {
			sig = make([]byte, 96)
		} else {
			badMsg := messageBytes("badmsg")
			sig = priv.Sign(badMsg[:]).Marshal()
		}
		desc := fmt.Sprintf("signature of %s%d", msgBody, i)

		set.Signatures[i] = sig
		set.PublicKeys[i] = pubkey
		set.Messages[i] = msg
		set.Descriptions[i] = desc
	}

	return set
}

func messageBytes(message string) [32]byte {
	var bytes [32]byte
	copy(bytes[:], message)
	return bytes
}

func createDescriptions(length int, text ...string) []string {
	desc := make([]string, length)
	for i := range desc {
		if len(text) > 0 {
			desc[i] = text[0]
		} else {
			desc[i] = TestSignature
		}
	}
	return desc
}

func sortSet(s *bls.SignatureBatch) *bls.SignatureBatch {
	sort.Sort(sorter{set: s})
	return s
}

type sorter struct {
	set *bls.SignatureBatch
}

func (s sorter) Len() int {
	return len(s.set.Messages)
}

func (s sorter) Swap(i, j int) {
	s.set.Signatures[i], s.set.Signatures[j] = s.set.Signatures[j], s.set.Signatures[i]
	s.set.PublicKeys[i], s.set.PublicKeys[j] = s.set.PublicKeys[j], s.set.PublicKeys[i]
	s.set.Messages[i], s.set.Messages[j] = s.set.Messages[j], s.set.Messages[i]
	s.set.Descriptions[i], s.set.Descriptions[j] = s.set.Descriptions[j], s.set.Descriptions[i]
}

func (s sorter) Less(i, j int) bool {
	return bytes.Compare(s.set.Messages[i][:], s.set.Messages[j][:]) == -1
}
