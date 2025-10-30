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

pragma solidity ^0.8.14;

/**
 * @dev A contract that implements the MinterManagementInterface has external
 * functions for adding and removing minters and modifying their allowances.
 * An example is the FiatToken contract.
 */
interface IMinterManagement {
    function isMinter(address _account) external view returns (bool);

    function minterAllowance(address _minter) external view returns (uint256);

    function configureMinter(address _minter, uint256 _minterAllowedAmount) external returns (bool);

    function removeMinter(address _minter) external returns (bool);

    function updateMasterMinter(address _newMasterMinter) external;
}
