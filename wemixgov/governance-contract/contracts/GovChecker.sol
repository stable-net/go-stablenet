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

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "./interface/IRegistry.sol";
import "./interface/IGov.sol";

/**
 * @title GovChecker
 * @dev GovChecker Contract that uses Registry contract
 */
contract GovChecker is OwnableUpgradeable {
	IRegistry public reg;

	bytes32 public constant GOV_NAME = "GovernanceContract";
	bytes32 public constant STAKING_NAME = "Staking";
	bytes32 public constant BALLOT_STORAGE_NAME = "BallotStorage";
	bytes32 public constant ENV_STORAGE_NAME = "EnvStorage";
	bytes32 public constant REWARD_POOL_NAME = "RewardPool";
	bytes32 public constant MAINTENANCE_NAME = "Maintenance";
	bytes32 public constant ECOSYSTEM_NAME = "Ecosystem";
	bytes32 public constant STAKING_REWARD_NAME = "StakingReward";

	/*
	 * @dev Function to set registry address. Contract that wants to use registry should setRegistry first.
	 * @param _addr address of registry
	 * @return A boolean that indicates if the operation was successful.
	 */
	event SetRegistry(address indexed addr);

	function setRegistry(address _addr) public onlyOwner {
		require(_addr != address(0), "Address should be non-zero");
		reg = IRegistry(_addr);
		emit SetRegistry(_addr);
	}

	modifier onlyGov() {
		require(getGovAddress() == msg.sender, "No Permission");
		_;
	}

	modifier onlyGovMem() {
		require(IGov(getGovAddress()).isMember(msg.sender), "No Permission");
		_;
	}

	modifier onlyGovStaker() {
		require(IGov(getGovAddress()).isStaker(msg.sender), "No Permission");
		_;
	}

	modifier anyGov() {
		require(getGovAddress() == msg.sender || IGov(getGovAddress()).isMember(msg.sender), "No Permission");
		_;
	}

	function getContractAddress(bytes32 name) internal view returns (address) {
		return reg.getContractAddress(name);
	}

	function getGovAddress() internal view returns (address) {
		return getContractAddress(GOV_NAME);
	}

	function getStakingAddress() internal view returns (address) {
		return getContractAddress(STAKING_NAME);
	}

	function getBallotStorageAddress() internal view returns (address) {
		return getContractAddress(BALLOT_STORAGE_NAME);
	}

	function getEnvStorageAddress() internal view returns (address) {
		return getContractAddress(ENV_STORAGE_NAME);
	}

	function getRewardPoolAddress() internal view returns (address) {
		return getContractAddress(REWARD_POOL_NAME);
	}

	function getEcosystemAddress() internal view returns (address) {
		return getContractAddress(ECOSYSTEM_NAME);
	}

	function getStakingRewardAddress() internal view returns (address) {
		return getContractAddress(STAKING_REWARD_NAME);
	}

	function getMaintenanceAddress() internal view returns (address) {
		return getContractAddress(MAINTENANCE_NAME);
	}
}
