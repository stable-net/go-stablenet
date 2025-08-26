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

pragma solidity ^0.8.0;

interface IStaking {
	function deposit() external payable;
	function withdraw(uint256) external;
	function lock(address, uint256) external;
	function unlock(address, uint256) external;
	function transferLocked(address, uint256, uint256) external;
	function balanceOf(address) external view returns (uint256);
	function lockedBalanceOf(address) external view returns (uint256);
	function availableBalanceOf(address) external view returns (uint256);
	function calcVotingWeight(address) external view returns (uint256);
	function calcVotingWeightWithScaleFactor(address, uint32) external view returns (uint256);
	// function isAllowed(address voter, address staker) external view returns(bool);
	function userBalanceOf(address ncp, address user) external view returns (uint256);
	function userTotalBalanceOf(address ncp) external view returns (uint256);
	function getRatioOfUserBalance(address ncp) external view returns (uint256);
	function delegateDepositAndLockMore(address ncp) external payable;
	function delegateUnlockAndWithdraw(address ncp, uint256 amount) external;
	function getTotalLockedBalance() external view returns (uint256);
}
