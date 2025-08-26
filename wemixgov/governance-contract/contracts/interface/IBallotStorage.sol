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

interface IBallotStorage {
	function createBallotForMember(
		uint256,
		uint256,
		uint256,
		address,
		address,
		address,
		address,
		address,
		bytes memory,
		bytes memory,
		bytes memory,
		uint
	) external;

	function createBallotForAddress(uint256, uint256, uint256, address, address) external returns (uint256);
	function createBallotForVariable(uint256, uint256, uint256, address, bytes32, uint256, bytes memory) external returns (uint256);
	function createBallotForExit(uint256, uint256, uint256) external;
	function createVote(uint256, uint256, address, uint256, uint256) external;
	function finalizeBallot(uint256, uint256) external;
	function startBallot(uint256, uint256, uint256) external;
	function updateBallotMemo(uint256, bytes memory) external;
	function updateBallotDuration(uint256, uint256) external;
	function updateBallotMemberLockAmount(uint256, uint256) external;

	function getBallotPeriod(uint256) external view returns (uint256, uint256, uint256);
	function getBallotVotingInfo(uint256) external view returns (uint256, uint256, uint256);
	function getBallotState(uint256) external view returns (uint256, uint256, bool);

	function getBallotBasic(
		uint256
	) external view returns (uint256, uint256, uint256, address, bytes memory, uint256, uint256, uint256, uint256, bool, uint256);

	function getBallotMember(
		uint256
	) external view returns (address, address, address, address, bytes memory, bytes memory, bytes memory, uint256, uint256);
	function getBallotAddress(uint256) external view returns (address);
	function getBallotVariable(uint256) external view returns (bytes32, uint256, bytes memory);
	function getBallotForExit(uint256) external view returns (uint256, uint256);
}
