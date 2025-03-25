package testutils

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestGeneratingGenesisExtra(t *testing.T) {
	QBFTExtra := &types.QBFTExtra{
		VanityData:        []byte{},
		PrevRound:         0,
		PreparedSeal:      &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		PrevPreparedSeal:  &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		Round:             0,
		CommittedSeal:     &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		PrevCommittedSeal: &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		EpochInfo: &types.EpochInfo{
			Stakers: []*types.Staker{
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xa62987A40E094CbE020313f19F71aeeB3E48B86f")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xE1bD8108149FCa703B9FCD8ea967B6b7e660f13b")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xd19f9374f4549B2fB182ED766d6b7501494a3634")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0x02cF1E577C79EF0E93947cCd82a4D41E0485Be73")), Diligence: types.DefaultDiligence},
			},
			Validators:    []uint32{0, 1, 2, 3},
			BLSPublicKeys: [][]byte{{}, {}, {}, {}},
		},
	}
	genesis := GenesisWithSeals(QBFTExtra.EpochInfo.GetStakers(), QBFTExtra.EpochInfo.BLSPublicKeys)
	t.Log("Genesis Extra Data: ", hex.EncodeToString(genesis.ExtraData))
	qbftExtra := new(types.QBFTExtra)
	err := rlp.DecodeBytes(genesis.ExtraData[:], qbftExtra)
	if err != nil {
		t.Errorf("Failed to decode genesis qbft extra data : %v", err)
	}
	qbftExtra.VanityData = []byte{} // clear vanity
	if !reflect.DeepEqual(QBFTExtra, qbftExtra) {
		t.Errorf("decoded extra object is different from origin(decoded=%v, origin=%v)", QBFTExtra, qbftExtra)
	}
}
