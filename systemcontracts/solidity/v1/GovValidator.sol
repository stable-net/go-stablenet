// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import { GovBase } from "../abstracts/GovBase.sol";

contract GovValidator is GovBase {
    using EnumerableSet for EnumerableSet.AddressSet;

    uint256 public constant BLS_PUBLIC_KEY_LENGTH = 48;
    uint256 public constant BLS_SIGNATURE_LENGTH = 96;

    error InvalidValidator();
    error AlreadyValidatorExists();
    error NoConfigurationChanging();
    error InvalidBlsKeyLength();
    error InvalidSignatureLength();
    error AlreadyRegisteredBlsKey();
    error FailedToVerifyBlsKey();
    error InvalidBlsKey();
    error InvalidMinerTip();

    // 0x0 ~ 0x31: reserved for GovBase
    address public blsPoP; // 0x32; Precompiled contract address for BLS PoP verification
    EnumerableSet.AddressSet private __validators; // 0x33, 0x34; validator addresses
    mapping(address => address) public validatorToOperator; // 0x35
    mapping(address => address) public operatorToValidator; // 0x36
    mapping(address => bytes) public validatorToBlsKey; // 0x37
    mapping(bytes => address) public blsKeyToValidator; // 0x38
    uint256 public minerTip; // 0x39; Minimum miner tip in wei (e.g., 100 Gwei = 100000000000)

    function isValidator(address _validator) external view returns (bool) {
        return __validators.contains(_validator);
    }

    function validatorList() external view returns (address[] memory) {
        return __validators.values();
    }

    function validatorCount() external view returns (uint256) {
        return __validators.length();
    }

    function configureValidator(address _newValidator, bytes calldata _blsKey, bytes calldata _blsSig) external onlyActiveMember {
        if (_newValidator == address(0)) {
            revert InvalidValidator();
        }
        if (validatorToOperator[_newValidator] != address(0) && validatorToOperator[_newValidator] != msg.sender) {
            // _newValidator is already registered by other operator
            revert AlreadyValidatorExists();
        }
        _checkBlsKey(_blsKey, _blsSig);

        address _oldValidator = operatorToValidator[msg.sender];
        if (blsKeyToValidator[_blsKey] != address(0) && blsKeyToValidator[_blsKey] != _oldValidator) {
            revert AlreadyRegisteredBlsKey();
        }

        // case 1: register new validator
        // case 2: change BLS key only
        // case 3: change validator address with changing BLS key (bls key can be same as old one)
        if (_oldValidator == address(0)) {
            // case 1
            _setValidatorAndBlsKey(_newValidator, _blsKey);
        } else if (_oldValidator == _newValidator) {
            // case 2
            if (blsKeyToValidator[_blsKey] == _oldValidator) {
                revert NoConfigurationChanging();
            }
            _changeBlsKey(_oldValidator, _blsKey);
        } else {
            // case 3
            _changeValidator(_oldValidator, _newValidator, _blsKey);
        }
    }

    function _setValidatorAndBlsKey(address _newValidator, bytes calldata _blsKey) internal {
        __validators.add(_newValidator);
        operatorToValidator[msg.sender] = _newValidator;
        validatorToOperator[_newValidator] = msg.sender;
        _setBlsKey(_newValidator, _blsKey);
    }

    function _setBlsKey(address _validator, bytes calldata _blsKey) internal {
        validatorToBlsKey[_validator] = _blsKey;
        blsKeyToValidator[_blsKey] = _validator;
    }

    function _changeBlsKey(address _validator, bytes calldata _blsKey) internal {
        delete blsKeyToValidator[validatorToBlsKey[_validator]];
        _setBlsKey(_validator, _blsKey);
    }

    function _changeValidator(address _oldValidator, address _newValidator, bytes calldata _blsKey) internal {
        _removeValidator(_oldValidator, msg.sender);

        _setValidatorAndBlsKey(_newValidator, _blsKey);
    }

    function _removeValidator(address _validator, address _member) internal {
        __validators.remove(_validator);
        delete validatorToOperator[_validator];
        delete operatorToValidator[_member];
        delete blsKeyToValidator[validatorToBlsKey[_validator]];
        delete validatorToBlsKey[_validator];
    }

    function _onMemberRemoved(address _member) internal override {
        address _validator = operatorToValidator[_member];
        if (_validator != address(0)) {
            _removeValidator(_validator, _member);
        }
    }

    function _onMemberAdded(address _member) internal override {
        // do nothing
    }

    function _onMemberChanged(address _oldMember, address _newMember) internal override {
        address _validator = operatorToValidator[_oldMember];
        if (_validator != address(0)) {
            operatorToValidator[_newMember] = _validator;
            validatorToOperator[_validator] = _newMember;
            delete operatorToValidator[_oldMember];
        }
    }

    function _executeCustomAction(bytes32 actionType, bytes memory callData) internal override returns (bool) {
        // GovValidator doesn't have custom actions beyond member management
        // All validator configuration is done directly via configureValidator
        return false;
    }

    function _checkBlsKey(bytes calldata _blsKey, bytes calldata _blsSig) internal view {
        if (_blsKey.length != BLS_PUBLIC_KEY_LENGTH) {
            revert InvalidBlsKeyLength();
        }
        if (_blsSig.length != BLS_SIGNATURE_LENGTH) {
            revert InvalidSignatureLength();
        }

        (bool _success, bytes memory _result) = blsPoP.staticcall(abi.encodePacked(_blsKey, _blsSig));
        if (!_success) {
            revert FailedToVerifyBlsKey();
        }
        if (!abi.decode(_result, (bool))) {
            revert InvalidBlsKey();
        }
    }

    // ========== Events ==========
    event MinerTipUpdated(uint256 oldTip, uint256 newTip, address indexed updater);

    // ========== Fee Policy Management ==========
    function proposeMinerTip(uint256 _newTip) 
        external 
        onlyMember 
        noActiveProposal 
        returns (uint256) 
    {
        if (_newTip == 0) revert InvalidMinerTip();
        bytes4 _selector = this.setMinerTip.selector;
        bytes memory _encodedParams = abi.encode(_newTip);
        return _createProposal(keccak256("SET_MINER_TIP"), abi.encodePacked(_selector, _encodedParams));
    }

    function setMinerTip(uint256 _newTip) external onlyMe {
        uint256 oldTip = minerTip;
        minerTip = _newTip;
        emit MinerTipUpdated(oldTip, _newTip, msg.sender);
    }

    function getMinerTipGwei() external view returns (uint256) {
        return minerTip / 1e9;
    }
}
