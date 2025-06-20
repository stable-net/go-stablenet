// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/wbft/types/decode.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func Decode(code uint64, data []byte) (WBFTMessage, error) {
	switch code {
	case PreprepareCode:
		var preprepare Preprepare
		if err := rlp.DecodeBytes(data, &preprepare); err != nil {
			return nil, wbftcommon.ErrFailedDecodePreprepare
		}
		preprepare.code = PreprepareCode
		return &preprepare, nil
	case PrepareCode:
		var prepare Prepare
		if err := rlp.DecodeBytes(data, &prepare); err != nil {
			return nil, wbftcommon.ErrFailedDecodeCommit
		}
		prepare.code = PrepareCode
		return &prepare, nil
	case CommitCode:
		var commit Commit
		if err := rlp.DecodeBytes(data, &commit); err != nil {
			return nil, wbftcommon.ErrFailedDecodeCommit
		}
		commit.code = CommitCode
		return &commit, nil
	case RoundChangeCode:
		var roundChange RoundChange
		if err := rlp.DecodeBytes(data, &roundChange); err != nil {
			return nil, wbftcommon.ErrFailedDecodeRoundChange
		}
		roundChange.code = RoundChangeCode
		return &roundChange, nil
	}
	return nil, wbftcommon.ErrInvalidMessage
}
