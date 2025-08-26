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

interface ITestnetGov {
	function isReward(address addr) external view returns (bool);
	function isVoter(address addr) external view returns (bool);
	function isStaker(address addr) external view returns (bool);
	function isMember(address) external view returns (bool);
	function getMember(uint256) external view returns (address);
	function getMemberLength() external view returns (uint256);
	function getReward(uint256) external view returns (address);
	function getNodeIdxFromMember(address) external view returns (uint256);
	function getMemberFromNodeIdx(uint256) external view returns (address);
	function getNodeLength() external view returns (uint256);
	function getNode(uint256) external view returns (bytes memory, bytes memory, bytes memory, uint);
	function getBallotInVoting() external view returns (uint256);
	function getVoter(uint256 idx) external view returns (address);
}
