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

pragma solidity ^0.8.14;

import "@openzeppelin/contracts/proxy/Proxy.sol";

contract GovRewardee is Proxy {
    // keccak-256("GovRewardee.implementation.slot") - 1
    bytes32 internal constant _IMPLEMENTATION_SLOT = 0x4ef8d65ed4f969898f05d331f7b880c9611386779b412e35e117f26e0983c85d;

    constructor(address rewardeeImp) {
        assembly {
            sstore(_IMPLEMENTATION_SLOT, rewardeeImp)
        }
    }

    function _implementation() internal view virtual override returns (address impAddress) {
        assembly {
            impAddress := sload(_IMPLEMENTATION_SLOT)
        }
    }
}
