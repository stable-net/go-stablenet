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

interface IEnvStorage {
	function setBlocksPerByBytes(bytes memory) external;
	function setBallotDurationMinByBytes(bytes memory) external;
	function setBallotDurationMaxByBytes(bytes memory) external;
	function setStakingMinByBytes(bytes memory) external;
	function setStakingMaxByBytes(bytes memory) external;
	function setMaxIdleBlockIntervalByBytes(bytes memory) external;
	function setBlockCreationTimeByBytes(bytes memory _value) external;
	function setBlockRewardAmountByBytes(bytes memory _value) external;
	function setMaxPriorityFeePerGasByBytes(bytes memory _value) external;
	function setBallotDurationMinMax(uint256 _min, uint256 _max) external;
	function setBlockRewardDistributionMethodByBytes(bytes memory _value) external;
	function setGasLimitAndBaseFeeByBytes(bytes memory _value) external;
	function setMaxBaseFeeByBytes(bytes memory _value) external;
	function setBallotDurationMinMaxByBytes(bytes memory _value) external;
	function setStakingMinMaxByBytes(bytes memory _value) external;
	function getBlockCreationTime() external view returns (uint256);
	function getBlockRewardAmount() external view returns (uint256);
	function getMaxPriorityFeePerGas() external view returns (uint256);
	function getStakingMinMax() external view returns (uint256, uint256);
	function getBlockRewardDistributionMethod() external view returns (uint256, uint256, uint256, uint256);
	function getGasLimitAndBaseFee() external view returns (uint256, uint256, uint256);
	function getMaxBaseFee() external view returns (uint256);
	function getBlocksPer() external view returns (uint256);
	function getStakingMin() external view returns (uint256);
	function getStakingMax() external view returns (uint256);
	function getBallotDurationMin() external view returns (uint256);
	function getBallotDurationMax() external view returns (uint256);
	function getBallotDurationMinMax() external view returns (uint256, uint256);
	function getMaxIdleBlockInterval() external view returns (uint256);
	function checkVariableCondition(bytes32 envKey, bytes memory envVal) external pure returns (bool);
	function setVariable(bytes32 envKey, bytes memory envVal) external;
}
