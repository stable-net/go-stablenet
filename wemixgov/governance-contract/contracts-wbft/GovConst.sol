// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

contract GovConst{
    uint256 public constant MINIMUM_STAKING = 500000e18;
    uint256 public constant MAXIMUM_STAKING = type(uint128).max;
    uint256 public constant UNBONDING_PERIOD_STAKER = 1 weeks;
    uint256 public constant UNBONDING_PERIOD_DELEGATOR = 72 hours;
    uint256 public constant BLS_PUBLIC_KEY_LENGTH = 48;
    uint256 public constant MIN_STAKERS = 5;
}