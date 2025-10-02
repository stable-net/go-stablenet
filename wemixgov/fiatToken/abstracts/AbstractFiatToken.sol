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

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

abstract contract AbstractFiatToken is IERC20 {
    function _approve(address owner, address spender, uint256 value) internal virtual;

    function _transfer(address from, address to, uint256 value) internal virtual;

    function _increaseAllowance(address owner, address spender, uint256 increment) internal virtual;

    function _decreaseAllowance(address owner, address spender, uint256 decrement) internal virtual;
}
