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

interface INCPExit {
	/**
	 * @dev Sets a new administrator.
	 * @param _newAdministrator Address of the new administrator.
	 */
	function setAdministrator(address _newAdministrator) external;

	/**
	 * @dev Sets a new administrator setter.
	 * @param _newAdministratorSetter Address of the new administrator setter.
	 */
	function setAdministratorSetter(address _newAdministratorSetter) external;

	/**
	 * @dev Exits from the contract.
	 * @param exitNcp Address of the NCP to exit.
	 * @param totalAmount Total amount of ether to exit with.
	 * @param lockedUserBalanceToNCPTotal Total locked user balance to NCP.
	 */
	function depositExitAmount(address exitNcp, uint256 totalAmount, uint256 lockedUserBalanceToNCPTotal) external payable;

	/**
	 * @dev Withdraws amount for a user.
	 * @param exitNcp Address of the ncp
	 * @param exitUser Address of the user to withdraw for.
	 * @param amount Amount to withdraw.
	 */
	function withdrawForUser(address exitNcp, address exitUser, uint256 amount) external;

	/**
	 * @dev Withdraws amount for the administrator.
	 * @param exitNcp Address of the NCP to withdraw from.
	 */
	function withdrawForAdministrator(address exitNcp, uint256 amount, address to) external;

	function getAvailableAmountForAdministrator(address exitNcp) external view returns (uint256);

	function getLockedUserBalanceToNCPTotal(address exitNcp) external view returns (uint256);
}
