// SPDX-License-Identifier: MIT

pragma solidity ^0.8.14;

interface IGovCouncil {
    function inspectOperation(bytes4 _selector, address _sender, bytes memory _arguments) external view returns (bool);
}
