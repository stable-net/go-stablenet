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

interface IMultiSigWallet {
    // struct of transaction
    struct Transaction {
        address to;
        uint256 value;
        bytes data;
        bool executed;
        uint256 currentNumberOfConfirmations;
    }

    /* ========== FUNCTION ========== */

    function submitTransaction(address _to, uint256 _value, bytes memory _data) external;
    function confirmTransaction(uint256 _txIndex) external;
    function executeTransaction(uint256 _txIndex) external payable;
    function revokeConfirmation(uint256 _txIndex) external;

    function addOwner(address _newOwner, bool _increaseQuorum) external;
    function removeOwner(address _owner, bool _reduceQuorum) external;
    function replaceOwner(address _owner, address _newOwner) external;
    function changeQuorum(uint256 _quorum) external;

    /* ========== VIEW FUNCTION ========== */

    function getOwners() external view returns (address[] memory);
    function getOwnerCount() external view returns (uint256);
    function getTransaction(
        uint256 _txIndex
    ) external view returns (address to, uint256 value, bytes memory data, bool executed, uint256 numConfirmations);
    function getTransactionCount() external view returns (uint256);

    /* ========== EVENTS ========== */

    event Deposit(address indexed sender, uint256 amount, uint256 balance);
    event SubmitTransaction(address indexed owner, uint256 indexed txIndex, address indexed to, uint256 value, bytes data);
    event ConfirmTransaction(address indexed owner, uint256 indexed txIndex);
    event RevokeConfirmation(address indexed owner, uint256 indexed txIndex);
    event ExecuteTransaction(address indexed owner, uint256 indexed txIndex);

    event AddOwner(address indexed newOwner);
    event RemoveOwner(address indexed owner);

    event ChangeQuorum(uint256 indexed quorum);
}
