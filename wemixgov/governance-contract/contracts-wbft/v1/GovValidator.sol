// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {GovBase} from "./GovBase.sol";

contract GovValidator is GovBase {
    using EnumerableSet for EnumerableSet.AddressSet;

    uint256 public constant BLS_PUBLIC_KEY_LENGTH = 48;
    uint256 public constant BLS_SIGNATURE_LENGTH = 96;

    // 0x0 ~ 0x31: reserved for GovBase
    address public blsPoP; // 0x32; Precompiled contract address for BLS PoP verification
    EnumerableSet.AddressSet private __validators; // 0x33, 0x34; validator addresses
    mapping(address => address) public validatorToOperator; // 0x35
    mapping(address => address) public operatorToValidator; // 0x36
    mapping(address => bytes) public validatorToBlsKey; // 0x37
    mapping(bytes => address) public blsKeyToValidator; // 0x38

    function isValidator(address _validator) external view returns (bool) {
        return __validators.contains(_validator);
    }

    function validatorList() external view returns (address[] memory) {
        return __validators.values();
    }

    function validatorCount() external view returns (uint256) {
        return __validators.length();
    }

    function configureValidator(address _newValidator, bytes calldata _blsPK, bytes calldata _blsSig) external onlyMember {
        require(!__validators.contains(_newValidator), "validator exists");
        _checkBLSPublicKey(_blsPK, _blsSig);

        address _oldValidator = operatorToValidator[msg.sender];
        if (_oldValidator != address(0)) {
            // already registered validator
            __validators.remove(_oldValidator);
            delete validatorToOperator[_oldValidator];
            delete blsKeyToValidator[validatorToBlsKey[_oldValidator]];
            delete validatorToBlsKey[_oldValidator];
        }
        operatorToValidator[msg.sender] = _newValidator;
        validatorToOperator[_newValidator] = msg.sender;
        validatorToBlsKey[_newValidator] = _blsPK;
        blsKeyToValidator[_blsPK] = _newValidator;
    }

    function onMemberRemoved(address _member) internal override {
        address _validator = operatorToValidator[_member];
        if (_validator != address(0)) {
            __validators.remove(_validator);
            delete validatorToOperator[_validator];
            delete blsKeyToValidator[validatorToBlsKey[_validator]];
            delete validatorToBlsKey[_validator];
            delete validatorToBlsKey[_validator];
            delete operatorToValidator[_member];
        }
    }

    function onMemberAdded(address _member) internal override {
        // do nothing
    }

    function onMemberChanged(address _oldMember, address _newMember) internal override {
        address _validator = operatorToValidator[_oldMember];
        if (_validator != address(0)) {
            operatorToValidator[_newMember] = _validator;
            validatorToOperator[_validator] = _newMember;
            delete operatorToValidator[_oldMember];
        }
    }

    function _executeInternalAction(bytes32 actionType) internal override returns (bool) {
        return true;
    }

    function _checkBLSPublicKey(bytes calldata _blsPK, bytes calldata _blsSig) internal view {
        require(_blsPK.length == BLS_PUBLIC_KEY_LENGTH, "invalid bls public key length");
        require(_blsSig.length == BLS_SIGNATURE_LENGTH, "invalid bls signature length");
        require(blsKeyToValidator[_blsPK] == address(0), "already registered bls public key");

        (bool _success, bytes memory _result) = blsPoP.staticcall(abi.encodePacked(_blsPK, _blsSig));
        require(_success, "failed to verify bls pop");
        require(abi.decode(_result, (bool)), "invalid bls public key");
    }
}
