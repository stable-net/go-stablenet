// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

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
