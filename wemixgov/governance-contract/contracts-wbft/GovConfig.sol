// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

contract GovConfig {
    uint256 public minimumStaking;
    uint256 public maximumStaking;
    uint256 public unbondingPeriodStaker;
    uint256 public unbondingPeriodDelegator;
    uint256 public feePrecision;
    uint256 public changeFeeDelay;
    uint256 public minStakers;
}
