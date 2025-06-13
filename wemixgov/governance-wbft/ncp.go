package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_NCP_LIST          = "0x0" // ,0x1
	SLOT_NCP_LAST_ID       = "0x2"
	SLOT_NCP_ID_TO_ADDRESS = "0x3"
	SLOT_NCP_ADDRESS_TO_ID = "0x4"
)

func NCPLength(govNCPAddress common.Address, state StateReader) uint64 {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Length(state, govNCPAddress)
}

func IsNCP(govNCPAddress common.Address, state StateReader, ncp common.Address) bool {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Contains(state, govNCPAddress, ncp)
}

func NCPList(govNCPAddress common.Address, state StateReader) []common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Values(state, govNCPAddress)
}

func NCPAt(govNCPAddress common.Address, state StateReader, index *big.Int) common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.At(state, govNCPAddress, index)
}
