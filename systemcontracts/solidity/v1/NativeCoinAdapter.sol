/**
 * Original Apache-2.0 License:
 * Copyright 2023 Circle Internet Group, Inc. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Modifications Copyright 2025 The stable-one Authors
 *
 * Original code based on: https://github.com/circlefin/stablecoin-evm/tree/c8c31b249341bf3ffb2e8dbff41977c392a260c5/contracts
 *
 * NOTE: This contract is included in a GPL-3.0 project.
 *       When distributed as part of the project, it is subject to GPL-3.0 terms.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

pragma solidity 0.8.14;

import { SafeMath } from "@openzeppelin/contracts/utils/math/SafeMath.sol";

import { AbstractFiatToken } from "../abstracts/AbstractFiatToken.sol";
import { Blacklistable } from "../abstracts/Blacklistable.sol";
import { Mintable } from "../abstracts/Mintable.sol";

import { EIP712Domain } from "../abstracts/eip/EIP712Domain.sol";
import { EIP3009 } from "../abstracts/eip/EIP3009.sol";
import { EIP2612 } from "../abstracts/eip/EIP2612.sol";
import { EIP712 } from "../libraries/EIP712.sol";

import { ICoinManager } from "../interfaces/ICoinManager.sol";

/**
 * @title NativeCoinAdapter
 * @dev ERC20 Token backed by fiat reserves
 */
contract NativeCoinAdapter is AbstractFiatToken, Mintable, Blacklistable, EIP3009, EIP2612 {
    using SafeMath for uint256;
    /**
     * [Mintable]
     * address masterMinter (slot 0x0)
     * mapping(address => bool) _minters (slot 0x1)
     * mapping(address => uint256) _minterAllowed (slot 0x2)
     *
     * [Blacklistable]
     * address blacklister (slot 0x3)
     * mapping(address => bool) _blacklisted (slot 0x4)
     *
     * [EIP3009]
     * mapping(address => mapping(bytes32 => bool)) _authorizationStates (slot 0x5)
     *
     * [EIP2612]
     * mapping(address => uint256) _permitNonces (slot 0x6)
     *
     * [EIP712Domain]
     * bytes32 _DEPRECATED_CACHED_DOMAIN_SEPARATOR (slot 0x7)
     */

    address private _coinManager; // (slot 0x8)

    string public name; // (slot 0x9)
    string public symbol; // (slot 0xa)
    uint8 public decimals; // (slot 0xb)
    string public currency; // (slot 0xc)

    mapping(address => mapping(address => uint256)) private _allowed; // (slot 0xd)
    uint256 private _totalSupply = 0; // (slot 0xe)

    /**
     * @notice Gets the totalSupply of the fiat token.
     * @return The totalSupply of the fiat token.
     */
    function totalSupply() external view override returns (uint256) {
        return _totalSupply;
    }

    /**
     * @notice Gets the fiat token balance of an account.
     * @param account  The address to check.
     * @return balance The fiat token balance of the account.
     */
    function balanceOf(address account) external view override returns (uint256) {
        return _balanceOf(account);
    }

    /**
     * @dev Helper method to obtain the balance of an account.
     * @param _account  The address of the account.
     * @return          The fiat token balance of the account.
     */
    function _balanceOf(address _account) internal view returns (uint256) {
        return _account.balance;
    }

    // ============================================================================
    // Mint & Burn
    // ============================================================================
    /**
     * @notice Mints fiat tokens to an address.
     * @param _to The address that will receive the minted tokens.
     * @param _amount The amount of tokens to mint. Must be less than or equal
     * to the minterAllowance of the caller.
     * @return True if the operation was successful.
     */
    function mint(address _to, uint256 _amount) external override onlyMinters notBlacklisted(msg.sender) notBlacklisted(_to) returns (bool) {
        require(_to != address(0), "NativeCoinAdapter: mint to the zero address");
        require(_amount > 0, "NativeCoinAdapter: mint amount not greater than 0");

        uint256 mintingAllowedAmount = _minterAllowed[msg.sender];
        require(_amount <= mintingAllowedAmount, "NativeCoinAdapter: mint amount exceeds minterAllowance");

        _minterAllowed[msg.sender] = mintingAllowedAmount.sub(_amount);

        (bool _success, ) = _coinManager.call(abi.encodeWithSelector(ICoinManager.mint.selector, _to, _amount));
        require(_success, "mint failed");

        _totalSupply = _totalSupply.add(_amount);
        emit Mint(msg.sender, _to, _amount);
        // Transfer event is emitted by the EVM
        return true;
    }

    /**
     * @notice Allows a minter to burn some of its own tokens.
     * @dev The caller must be a minter, must not be blacklisted, and the amount to burn
     * should be less than or equal to the account's balance.
     * @param _amount the amount of tokens to be burned.
     */
    function burn(uint256 _amount) external override onlyMinters notBlacklisted(msg.sender) {
        uint256 balance = _balanceOf(msg.sender);
        require(_amount > 0, "NativeCoinAdapter: burn amount not greater than 0");
        require(balance >= _amount, "NativeCoinAdapter: burn amount exceeds balance");

        (bool success, ) = _coinManager.call(abi.encodeWithSelector(ICoinManager.burn.selector, msg.sender, _amount));
        require(success, "burn failed");

        _totalSupply = _totalSupply.sub(_amount);
        emit Burn(msg.sender, _amount);
        // Transfer event is emitted by the EVM
    }

    // ============================================================================
    // Transfer
    // ============================================================================
    /**
     * @notice Transfers tokens from an address to another by spending the caller's allowance.
     * @dev The caller must have some fiat token allowance on the payer's tokens.
     * @param from  Payer's address.
     * @param to    Payee's address.
     * @param value Transfer amount.
     * @return True if the operation was successful.
     */
    function transferFrom(
        address from,
        address to,
        uint256 value
    ) external override notBlacklisted(msg.sender) notBlacklisted(from) notBlacklisted(to) returns (bool) {
        require(value <= _allowed[from][msg.sender], "ERC20: transfer amount exceeds allowance");
        _transfer(from, to, value);
        _allowed[from][msg.sender] = _allowed[from][msg.sender].sub(value);
        return true;
    }

    /**
     * @notice Transfers tokens from the caller.
     * @param to    Payee's address.
     * @param value Transfer amount.
     * @return True if the operation was successful.
     */
    function transfer(address to, uint256 value) external override notBlacklisted(msg.sender) notBlacklisted(to) returns (bool) {
        _transfer(msg.sender, to, value);
        return true;
    }

    /**
     * @dev Internal function to process transfers.
     * @param from  Payer's address.
     * @param to    Payee's address.
     * @param value Transfer amount.
     */
    function _transfer(address from, address to, uint256 value) internal override {
        require(from != address(0), "ERC20: transfer from the zero address");
        require(to != address(0), "ERC20: transfer to the zero address");
        require(value <= _balanceOf(from), "ERC20: transfer amount exceeds balance");

        (bool success, ) = _coinManager.call(abi.encodeWithSelector(ICoinManager.transfer.selector, from, to, value));
        require(success, "transfer failed");
        // Transfer event is emitted by the EVM
    }

    // ============================================================================
    // Allowance
    // ============================================================================
    /**
     * @notice Gets the remaining amount of fiat tokens a spender is allowed to transfer on
     * behalf of the token owner.
     * @param owner   The token owner's address.
     * @param spender The spender's address.
     * @return The remaining allowance.
     */
    function allowance(address owner, address spender) external view override returns (uint256) {
        return _allowed[owner][spender];
    }

    /**
     * @notice Sets a fiat token allowance for a spender to spend on behalf of the caller.
     * @param spender The spender's address.
     * @param value   The allowance amount.
     * @return True if the operation was successful.
     */
    function approve(address spender, uint256 value) external override notBlacklisted(msg.sender) notBlacklisted(spender) returns (bool) {
        _approve(msg.sender, spender, value);
        return true;
    }

    /**
     * @dev Internal function to set allowance.
     * @param owner     Token owner's address.
     * @param spender   Spender's address.
     * @param value     Allowance amount.
     */
    function _approve(address owner, address spender, uint256 value) internal override {
        require(owner != address(0), "ERC20: approve from the zero address");
        require(spender != address(0), "ERC20: approve to the zero address");
        _allowed[owner][spender] = value;
        emit Approval(owner, spender, value);
    }

    /**
     * @notice Increase the allowance by a given increment
     * @param spender   Spender's address
     * @param increment Amount of increase in allowance
     * @return True if successful
     */
    function increaseAllowance(address spender, uint256 increment) external notBlacklisted(msg.sender) notBlacklisted(spender) returns (bool) {
        _increaseAllowance(msg.sender, spender, increment);
        return true;
    }

    /**
     * @notice Decrease the allowance by a given decrement
     * @param spender   Spender's address
     * @param decrement Amount of decrease in allowance
     * @return True if successful
     */
    function decreaseAllowance(address spender, uint256 decrement) external notBlacklisted(msg.sender) notBlacklisted(spender) returns (bool) {
        _decreaseAllowance(msg.sender, spender, decrement);
        return true;
    }

    /**
     * @dev Internal function to increase the allowance by a given increment
     * @param owner     Token owner's address
     * @param spender   Spender's address
     * @param increment Amount of increase
     */
    function _increaseAllowance(address owner, address spender, uint256 increment) internal override {
        _approve(owner, spender, _allowed[owner][spender].add(increment));
    }

    /**
     * @dev Internal function to decrease the allowance by a given decrement
     * @param owner     Token owner's address
     * @param spender   Spender's address
     * @param decrement Amount of decrease
     */
    function _decreaseAllowance(address owner, address spender, uint256 decrement) internal override {
        _approve(owner, spender, _allowed[owner][spender].sub(decrement, "ERC20: decreased allowance below zero"));
    }

    // ============================================================================
    // EIP-3009
    // ============================================================================
    /**
     * @notice Execute a transfer with a signed authorization
     * @param from          Payer's address (Authorizer)
     * @param to            Payee's address
     * @param value         Amount to be transferred
     * @param validAfter    The time after which this is valid (unix time)
     * @param validBefore   The time before which this is valid (unix time)
     * @param nonce         Unique nonce
     * @param v             v of the signature
     * @param r             r of the signature
     * @param s             s of the signature
     */
    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external notBlacklisted(from) notBlacklisted(to) {
        _transferWithAuthorization(from, to, value, validAfter, validBefore, nonce, v, r, s);
    }

    /**
     * @notice Execute a transfer with a signed authorization
     * @dev EOA wallet signatures should be packed in the order of r, s, v.
     * @param from          Payer's address (Authorizer)
     * @param to            Payee's address
     * @param value         Amount to be transferred
     * @param validAfter    The time after which this is valid (unix time)
     * @param validBefore   The time before which this is valid (unix time)
     * @param nonce         Unique nonce
     * @param signature     Signature bytes signed by an EOA wallet or a contract wallet
     */
    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes memory signature
    ) external notBlacklisted(from) notBlacklisted(to) {
        _transferWithAuthorization(from, to, value, validAfter, validBefore, nonce, signature);
    }

    /**
     * @notice Receive a transfer with a signed authorization from the payer
     * @dev This has an additional check to ensure that the payee's address
     * matches the caller of this function to prevent front-running attacks.
     * @param from          Payer's address (Authorizer)
     * @param to            Payee's address
     * @param value         Amount to be transferred
     * @param validAfter    The time after which this is valid (unix time)
     * @param validBefore   The time before which this is valid (unix time)
     * @param nonce         Unique nonce
     * @param v             v of the signature
     * @param r             r of the signature
     * @param s             s of the signature
     */
    function receiveWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external notBlacklisted(from) notBlacklisted(to) {
        _receiveWithAuthorization(from, to, value, validAfter, validBefore, nonce, v, r, s);
    }

    /**
     * @notice Receive a transfer with a signed authorization from the payer
     * @dev This has an additional check to ensure that the payee's address
     * matches the caller of this function to prevent front-running attacks.
     * EOA wallet signatures should be packed in the order of r, s, v.
     * @param from          Payer's address (Authorizer)
     * @param to            Payee's address
     * @param value         Amount to be transferred
     * @param validAfter    The time after which this is valid (unix time)
     * @param validBefore   The time before which this is valid (unix time)
     * @param nonce         Unique nonce
     * @param signature     Signature bytes signed by an EOA wallet or a contract wallet
     */
    function receiveWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes memory signature
    ) external notBlacklisted(from) notBlacklisted(to) {
        _receiveWithAuthorization(from, to, value, validAfter, validBefore, nonce, signature);
    }

    /**
     * @notice Attempt to cancel an authorization
     * @dev Works only if the authorization is not yet used.
     * @param authorizer    Authorizer's address
     * @param nonce         Nonce of the authorization
     * @param v             v of the signature
     * @param r             r of the signature
     * @param s             s of the signature
     */
    function cancelAuthorization(address authorizer, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) external {
        _cancelAuthorization(authorizer, nonce, v, r, s);
    }

    /**
     * @notice Attempt to cancel an authorization
     * @dev Works only if the authorization is not yet used.
     * EOA wallet signatures should be packed in the order of r, s, v.
     * @param authorizer    Authorizer's address
     * @param nonce         Nonce of the authorization
     * @param signature     Signature bytes signed by an EOA wallet or a contract wallet
     */
    function cancelAuthorization(address authorizer, bytes32 nonce, bytes memory signature) external {
        _cancelAuthorization(authorizer, nonce, signature);
    }

    // ============================================================================
    // EIP-2612
    // ============================================================================
    /**
     * @notice Update allowance with a signed permit
     * @param owner       Token owner's address (Authorizer)
     * @param spender     Spender's address
     * @param value       Amount of allowance
     * @param deadline    The time at which the signature expires (unix time), or max uint256 value to signal no expiration
     * @param v           v of the signature
     * @param r           r of the signature
     * @param s           s of the signature
     */
    function permit(
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external notBlacklisted(owner) notBlacklisted(spender) {
        _permit(owner, spender, value, deadline, v, r, s);
    }

    /**
     * @notice Update allowance with a signed permit
     * @dev EOA wallet signatures should be packed in the order of r, s, v.
     * @param owner       Token owner's address (Authorizer)
     * @param spender     Spender's address
     * @param value       Amount of allowance
     * @param deadline    The time at which the signature expires (unix time), or max uint256 value to signal no expiration
     * @param signature   Signature bytes signed by an EOA wallet or a contract wallet
     */
    function permit(
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        bytes memory signature
    ) external notBlacklisted(owner) notBlacklisted(spender) {
        _permit(owner, spender, value, deadline, signature);
    }

    // ============================================================================
    // Blacklistable
    // ============================================================================
    /**
     * @inheritdoc Blacklistable
     */
    function updateBlacklister(address) external pure override {
        revert("Blacklister role merged into MasterMinter");
    }

    /**
     * @inheritdoc Blacklistable
     */
    function _isBlacklister(address _account) internal view override returns (bool) {
        return _account == masterMinter;
    }

    // ============================================================================
    // Metadata
    // ============================================================================
    /**
     * @dev Internal function to get the current chain id.
     * @return The current chain id.
     */
    function _chainId() internal view returns (uint256) {
        uint256 chainId;
        assembly {
            chainId := chainid()
        }
        return chainId;
    }

    /**
     * @inheritdoc EIP712Domain
     */
    function _domainSeparator() internal view override returns (bytes32) {
        return EIP712.makeDomainSeparator(name, version(), _chainId());
    }

    /**
     * @notice Version string for the EIP712 domain separator
     * @return Version string
     */
    function version() public pure returns (string memory) {
        return "1";
    }
}
