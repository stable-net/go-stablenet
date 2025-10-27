// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

/**
 * @title ConfigurationValidator
 * @notice Library for validating governance configuration parameters
 * @dev Provides comprehensive validation for GovMasterMinter configuration values
 * Initial implementation for centralized configuration validation
 */
library ConfigurationValidator {
    // ========== Custom Errors ==========
    error InvalidAddress();
    error InvalidAmount();
    error IntervalTooShort(uint256 provided, uint256 minimum);
    error IntervalTooLong(uint256 provided, uint256 maximum);
    error WindowTooShort(uint256 provided, uint256 minimum);
    error WindowTooLong(uint256 provided, uint256 maximum);
    error AmountTooSmall(uint256 provided, uint256 minimum);
    error AmountTooLarge(uint256 provided, uint256 maximum);
    error InvalidRateLimitRelation();
    error CooldownTooShort(uint256 provided, uint256 minimum);
    error CooldownTooLong(uint256 provided, uint256 maximum);
    error InvalidMintingCapConfiguration();

    // ========== Constants ==========
    // Mint interval bounds (1 minute to 7 days)
    uint256 public constant MIN_MINT_INTERVAL = 1 minutes;
    uint256 public constant MAX_MINT_INTERVAL = 7 days;

    // Minting cap window bounds (1 hour to 30 days)
    uint256 public constant MIN_MINTING_CAP_WINDOW = 1 hours;
    uint256 public constant MAX_MINTING_CAP_WINDOW = 30 days;

    // Minting cap amount bounds (1000 tokens to 1 trillion tokens)
    uint256 public constant MIN_MINTING_CAP_AMOUNT = 1000 * 10 ** 18;
    uint256 public constant MAX_MINTING_CAP_AMOUNT = 1_000_000_000_000 * 10 ** 18; // 1 trillion

    // Rate limit bounds
    uint256 public constant MIN_RATE_LIMIT = 1 * 10 ** 18; // 1 token
    uint256 public constant MAX_RATE_LIMIT = 1_000_000_000_000 * 10 ** 18; // 1 trillion tokens

    // Cooldown bounds (1 minute to 7 days)
    uint256 public constant MIN_COOLDOWN = 1 minutes;
    uint256 public constant MAX_COOLDOWN = 7 days;

    // ========== Address Validation ==========

    /**
     * @notice Validate that address is not zero
     * @param addr Address to validate
     */
    function validateAddress(address addr) internal pure {
        if (addr == address(0)) revert InvalidAddress();
    }

    /**
     * @notice Validate that address is not zero and not same as compareTo
     * @param addr Address to validate
     * @param compareTo Address to compare against
     */
    function validateAddressNotSame(address addr, address compareTo) internal pure {
        validateAddress(addr);
        if (addr == compareTo) revert InvalidAddress();
    }

    // ========== Amount Validation ==========

    /**
     * @notice Validate amount is greater than zero
     * @param amount Amount to validate
     */
    function validateAmount(uint256 amount) internal pure {
        if (amount == 0) revert InvalidAmount();
    }

    /**
     * @notice Validate amount is within bounds
     * @param amount Amount to validate
     * @param min Minimum allowed amount
     * @param max Maximum allowed amount
     */
    function validateAmountRange(uint256 amount, uint256 min, uint256 max) internal pure {
        if (amount == 0) revert InvalidAmount();
        if (amount < min) revert AmountTooSmall(amount, min);
        if (amount > max) revert AmountTooLarge(amount, max);
    }

    // ========== Mint Interval Validation ==========

    /**
     * @notice Validate mint interval (cooldown) parameters
     * @param globalInterval Global mint cooldown in seconds
     * @param perAddressInterval Per-address mint cooldown in seconds
     * @dev Both intervals must be within reasonable bounds
     */
    function validateMintIntervals(uint256 globalInterval, uint256 perAddressInterval) internal pure {
        // Validate global interval
        if (globalInterval < MIN_MINT_INTERVAL) {
            revert IntervalTooShort(globalInterval, MIN_MINT_INTERVAL);
        }
        if (globalInterval > MAX_MINT_INTERVAL) {
            revert IntervalTooLong(globalInterval, MAX_MINT_INTERVAL);
        }

        // Validate per-address interval
        if (perAddressInterval < MIN_MINT_INTERVAL) {
            revert IntervalTooShort(perAddressInterval, MIN_MINT_INTERVAL);
        }
        if (perAddressInterval > MAX_MINT_INTERVAL) {
            revert IntervalTooLong(perAddressInterval, MAX_MINT_INTERVAL);
        }
    }

    // ========== Minting Cap Validation ==========

    /**
     * @notice Validate minting cap configuration
     * @param window Rolling window duration in seconds
     * @param amount Maximum amount within the window
     * @dev Window and amount must be within reasonable bounds
     */
    function validateMintingCap(uint256 window, uint256 amount) internal pure {
        // Validate window
        if (window < MIN_MINTING_CAP_WINDOW) {
            revert WindowTooShort(window, MIN_MINTING_CAP_WINDOW);
        }
        if (window > MAX_MINTING_CAP_WINDOW) {
            revert WindowTooLong(window, MAX_MINTING_CAP_WINDOW);
        }

        // Validate amount
        if (amount < MIN_MINTING_CAP_AMOUNT) {
            revert AmountTooSmall(amount, MIN_MINTING_CAP_AMOUNT);
        }
        if (amount > MAX_MINTING_CAP_AMOUNT) {
            revert AmountTooLarge(amount, MAX_MINTING_CAP_AMOUNT);
        }
    }

    /**
     * @notice Validate minting cap is reasonable relative to window
     * @param window Rolling window duration in seconds
     * @param amount Maximum amount within the window
     * @dev Ensure amount is not disproportionate to window duration
     */
    function validateMintingCapRatio(uint256 window, uint256 amount) internal pure {
        validateMintingCap(window, amount);

        // Additional ratio check: ensure reasonable amount per hour
        // For example, a 1-hour window shouldn't have a 1 trillion token cap
        uint256 amountPerHour = (amount * 1 hours) / window;

        // Maximum reasonable rate: 10 billion tokens per hour
        uint256 maxTokensPerHour = 10_000_000_000 * 10 ** 18;

        if (amountPerHour > maxTokensPerHour) {
            revert InvalidMintingCapConfiguration();
        }
    }

    // ========== Rate Limit Validation ==========

    /**
     * @notice Validate rate limits follow hierarchy: maxSingle <= hourly <= daily
     * @param maxSingle Maximum single mint amount
     * @param hourly Hourly mint limit
     * @param daily Daily mint limit
     */
    function validateRateLimits(uint256 maxSingle, uint256 hourly, uint256 daily) internal pure {
        // Validate individual amounts
        if (maxSingle == 0 || hourly == 0 || daily == 0) {
            revert InvalidAmount();
        }

        // Validate within bounds
        if (maxSingle < MIN_RATE_LIMIT || maxSingle > MAX_RATE_LIMIT) {
            revert AmountTooSmall(maxSingle, MIN_RATE_LIMIT);
        }
        if (daily > MAX_RATE_LIMIT) {
            revert AmountTooLarge(daily, MAX_RATE_LIMIT);
        }

        // Validate hierarchy
        if (maxSingle > hourly || hourly > daily) {
            revert InvalidRateLimitRelation();
        }
    }

    // ========== Cooldown Validation ==========

    /**
     * @notice Validate cooldown parameters
     * @param global Global cooldown in seconds
     * @param perAddress Per-address cooldown in seconds
     */
    function validateCooldowns(uint256 global, uint256 perAddress) internal pure {
        // Validate global cooldown
        if (global < MIN_COOLDOWN) {
            revert CooldownTooShort(global, MIN_COOLDOWN);
        }
        if (global > MAX_COOLDOWN) {
            revert CooldownTooLong(global, MAX_COOLDOWN);
        }

        // Validate per-address cooldown
        if (perAddress < MIN_COOLDOWN) {
            revert CooldownTooShort(perAddress, MIN_COOLDOWN);
        }
        if (perAddress > MAX_COOLDOWN) {
            revert CooldownTooLong(perAddress, MAX_COOLDOWN);
        }
    }

    // ========== Allowance Validation ==========

    /**
     * @notice Validate minter allowance
     * @param allowance Allowance to validate
     * @param maxAllowance Maximum allowed allowance
     */
    function validateAllowance(uint256 allowance, uint256 maxAllowance) internal pure {
        if (allowance == 0) revert InvalidAmount();
        if (allowance > maxAllowance) {
            revert AmountTooLarge(allowance, maxAllowance);
        }
    }

    // ========== Composite Validation ==========

    /**
     * @notice Validate complete governance configuration
     * @param globalInterval Global mint interval
     * @param perAddressInterval Per-address mint interval
     * @param mintingCapWindow Minting cap window
     * @param mintingCapAmount Minting cap amount
     * @dev Performs comprehensive validation of all configuration parameters
     */
    function validateCompleteConfiguration(
        uint256 globalInterval,
        uint256 perAddressInterval,
        uint256 mintingCapWindow,
        uint256 mintingCapAmount
    ) internal pure {
        validateMintIntervals(globalInterval, perAddressInterval);
        validateMintingCapRatio(mintingCapWindow, mintingCapAmount);
    }

    // ========== View Functions ==========

    /**
     * @notice Get validation bounds for mint intervals
     * @return min Minimum allowed interval
     * @return max Maximum allowed interval
     */
    function getMintIntervalBounds() internal pure returns (uint256 min, uint256 max) {
        return (MIN_MINT_INTERVAL, MAX_MINT_INTERVAL);
    }

    /**
     * @notice Get validation bounds for minting cap window
     * @return min Minimum allowed window
     * @return max Maximum allowed window
     */
    function getMintingCapWindowBounds() internal pure returns (uint256 min, uint256 max) {
        return (MIN_MINTING_CAP_WINDOW, MAX_MINTING_CAP_WINDOW);
    }

    /**
     * @notice Get validation bounds for minting cap amount
     * @return min Minimum allowed amount
     * @return max Maximum allowed amount
     */
    function getMintingCapAmountBounds() internal pure returns (uint256 min, uint256 max) {
        return (MIN_MINTING_CAP_AMOUNT, MAX_MINTING_CAP_AMOUNT);
    }

    /**
     * @notice Get validation bounds for cooldowns
     * @return min Minimum allowed cooldown
     * @return max Maximum allowed cooldown
     */
    function getCooldownBounds() internal pure returns (uint256 min, uint256 max) {
        return (MIN_COOLDOWN, MAX_COOLDOWN);
    }
}
