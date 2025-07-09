// SPDX-License-Identifier: MIT

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
