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

/// @author @seunghwalee
interface INCPStaking {
	struct UserInfo {
		uint256 amount;
		uint256 rewardDebt;
		uint256 pendingReward;
		uint256 pendingAmountReward;
		uint256 lastRewardClaimed;
	}
	function ncpDeposit(uint256 amount, address payable to) external payable;
	function ncpWithdraw(uint256 amount, address payable to) external payable;
	function getUserInfo(uint256 pid, address account) external view returns (UserInfo memory info);
	function ncpToIdx(address ncp) external view returns (uint256);
}
