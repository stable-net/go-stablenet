package messages

import (
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func Decode(code uint64, data []byte) (QBFTMessage, error) {
	switch code {
	case PreprepareCode:
		var preprepare Preprepare
		if err := rlp.DecodeBytes(data, &preprepare); err != nil {
			return nil, qbftcommon.ErrFailedDecodePreprepare
		}
		preprepare.code = PreprepareCode
		return &preprepare, nil
	case PrepareCode:
		var prepare Prepare
		if err := rlp.DecodeBytes(data, &prepare); err != nil {
			return nil, qbftcommon.ErrFailedDecodeCommit
		}
		prepare.code = PrepareCode
		return &prepare, nil
	case CommitCode:
		var commit Commit
		if err := rlp.DecodeBytes(data, &commit); err != nil {
			return nil, qbftcommon.ErrFailedDecodeCommit
		}
		commit.code = CommitCode
		return &commit, nil
	case RoundChangeCode:
		var roundChange RoundChange
		if err := rlp.DecodeBytes(data, &roundChange); err != nil {
			return nil, qbftcommon.ErrFailedDecodeRoundChange
		}
		roundChange.code = RoundChangeCode
		return &roundChange, nil
	}

	return nil, qbftcommon.ErrInvalidMessage
}
