package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_NCP_LIST = "0x0" // ,0x1
)

func NCPLength(state StateReader) uint64 {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Length(state, GovNCPAddress)
}

func IsNCP(state StateReader, ncp common.Address) bool {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Contains(state, GovNCPAddress, ncp)
}

func NCPList(state StateReader) []common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Values(state, GovNCPAddress)
}

func NCPAt(state StateReader, index *big.Int) common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.At(state, GovNCPAddress, index)
}
