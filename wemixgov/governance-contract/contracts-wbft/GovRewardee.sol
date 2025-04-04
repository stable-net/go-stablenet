// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/proxy/Proxy.sol";

contract GovRewardee is Proxy {
    address public constant GOV_REWARDEE_IMP = address(0x1003);

    function _implementation() internal view virtual override returns (address) {
        return GOV_REWARDEE_IMP;
    }
}
