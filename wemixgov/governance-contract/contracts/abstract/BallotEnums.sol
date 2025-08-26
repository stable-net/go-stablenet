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

contract BallotEnums {
	enum BallotStates {
		Invalid,
		Ready,
		InProgress,
		Accepted,
		Rejected,
		Canceled
	}

	enum DecisionTypes {
		Invalid,
		Accept,
		Reject
	}

	enum BallotTypes {
		Invalid,
		MemberAdd, // new Member Address, new Node id, new Node ip, new Node port
		MemberRemoval, // old Member Address
		MemberChange, // Old Member Address, New Member Address, new Node id, New Node ip, new Node port
		GovernanceChange, // new Governace Impl Address
		EnvValChange // Env variable name, type , value
	}
}
