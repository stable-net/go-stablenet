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

import "../interface/IStaking.sol";
import "../interface/IGov.sol";
import "../GovChecker.sol";

abstract contract AGov is GovChecker, IGov {
	uint public modifiedBlock;

	// For voting member
	mapping(uint256 => address) internal voters;
	mapping(address => uint256) public voterIdx;
	uint256 internal memberLength;

	// For reward member
	mapping(uint256 => address) internal rewards;
	mapping(address => uint256) public rewardIdx;

	//For staking member
	mapping(uint256 => address) internal stakers;
	mapping(address => uint256) public stakerIdx;

	//For a node duplicate check
	// mainnet value is here
	// mapping(bytes32=>bool) internal checkNodeInfo;
	mapping(bytes => bool) internal checkNodeName;
	mapping(bytes => bool) internal checkNodeEnode;
	mapping(bytes32 => bool) internal checkNodeIpPort;

	// For enode
	struct Node {
		bytes name;
		bytes enode;
		bytes ip;
		uint port;
	}

	mapping(uint256 => Node) internal nodes;
	mapping(address => uint256) internal nodeIdxFromMember;
	mapping(uint256 => address) internal nodeToMember;
	uint256 internal nodeLength;

	// For ballot
	uint256 public ballotLength;
	uint256 public voteLength;
	uint256 internal ballotInVoting;

	function isReward(address addr) public view override returns (bool) {
		return (rewardIdx[addr] != 0);
	}
	function isVoter(address addr) public view override returns (bool) {
		return (voterIdx[addr] != 0);
	}
	function isStaker(address addr) public view override returns (bool) {
		return (stakerIdx[addr] != 0);
	}
	function isMember(address addr) public view override returns (bool) {
		return (isStaker(addr) || isVoter(addr));
	}
	function getMember(uint256 idx) public view override returns (address) {
		return stakers[idx];
	}
	function getMemberLength() public view override returns (uint256) {
		return memberLength;
	}
	function getReward(uint256 idx) public view override returns (address) {
		return rewards[idx];
	}
	function getNodeIdxFromMember(address addr) public view override returns (uint256) {
		return nodeIdxFromMember[addr];
	}
	function getMemberFromNodeIdx(uint256 idx) public view override returns (address) {
		return nodeToMember[idx];
	}
	function getNodeLength() public view override returns (uint256) {
		return nodeLength;
	}
	//====NxtMeta=====/
	function getVoter(uint256 idx) public view override returns (address) {
		return voters[idx];
	}

	function getNode(uint256 idx) public view override returns (bytes memory name, bytes memory enode, bytes memory ip, uint port) {
		return (nodes[idx].name, nodes[idx].enode, nodes[idx].ip, nodes[idx].port);
	}

	function getBallotInVoting() public view override returns (uint256) {
		return ballotInVoting;
	}
}
