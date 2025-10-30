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

import { IFiatToken } from "../interfaces/IFiatToken.sol";

/**
 * @title MockFiatToken
 * @notice Mock FiatToken contract for testing GovMinter and GovMasterMinter
 * @dev Implements full IFiatToken interface for comprehensive testing
 */
contract MockFiatToken is IFiatToken {
    mapping(address => bool) private __isMinter;
    mapping(address => uint256) private __minterAllowance;
    mapping(address => uint256) private __balances;
    uint256 private __totalSupply;
    address private __masterMinter;

    bool public shouldFailMint;
    bool public shouldFailBurn;

    event MinterConfigured(address indexed minter, uint256 minterAllowedAmount);
    event MinterRemoved(address indexed oldMinter);
    event MasterMinterChanged(address indexed newMasterMinter);
    event Mint(address indexed minter, address indexed to, uint256 amount);
    event Burn(address indexed burner, uint256 amount);

    // ========== IMinterManagement Functions ==========

    function configureMinter(address minter, uint256 minterAllowedAmount) external returns (bool) {
        __isMinter[minter] = true;
        __minterAllowance[minter] = minterAllowedAmount;
        emit MinterConfigured(minter, minterAllowedAmount);
        return true;
    }

    function removeMinter(address minter) external returns (bool) {
        __isMinter[minter] = false;
        __minterAllowance[minter] = 0;
        emit MinterRemoved(minter);
        return true;
    }

    function isMinter(address account) external view returns (bool) {
        return __isMinter[account];
    }

    function minterAllowance(address minter) external view returns (uint256) {
        return __minterAllowance[minter];
    }

    function updateMasterMinter(address newMasterMinter) external {
        __masterMinter = newMasterMinter;
        emit MasterMinterChanged(newMasterMinter);
    }

    // ========== IFiatToken Functions ==========

    function mint(address to, uint256 amount) external {
        require(!shouldFailMint, "MockFiatToken: mint failed");
        require(to != address(0), "MockFiatToken: mint to zero address");

        __balances[to] += amount;
        __totalSupply += amount;
        emit Mint(msg.sender, to, amount);
    }

    function burn(uint256 amount) external {
        require(!shouldFailBurn, "MockFiatToken: burn failed");
        require(__balances[msg.sender] >= amount, "MockFiatToken: insufficient balance");

        __balances[msg.sender] -= amount;
        __totalSupply -= amount;
        emit Burn(msg.sender, amount);
    }

    // ========== Test Utility Functions ==========

    function setFailMint(bool _shouldFail) external {
        shouldFailMint = _shouldFail;
    }

    function setFailBurn(bool _shouldFail) external {
        shouldFailBurn = _shouldFail;
    }

    function balanceOf(address account) external view returns (uint256) {
        return __balances[account];
    }

    function totalSupply() external view returns (uint256) {
        return __totalSupply;
    }

    // Helper to set balance for testing
    function setBalance(address account, uint256 amount) external {
        uint256 oldBalance = __balances[account];
        __balances[account] = amount;

        if (amount > oldBalance) {
            __totalSupply += amount - oldBalance;
        } else {
            __totalSupply -= oldBalance - amount;
        }
    }
}
