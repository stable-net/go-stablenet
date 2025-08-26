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

package gov

type GovContract string

const (
	CONTRACT_REGISTRY          = "Registry"
	CONTRACT_GOV               = "Gov"
	CONTRACT_GOV_IMP           = "GovImp"
	CONTRACT_NCPEXIT           = "NCPExit"
	CONTRACT_NCPEXIT_IMP       = "NCPExitImp"
	CONTRACT_STAKING           = "Staking"
	CONTRACT_STAKING_IMP       = "StakingImp"
	CONTRACT_BALLOTSTORAGE     = "BallotStorage"
	CONTRACT_BALLOTSTORAGE_IMP = "BallotStorageImp"
	CONTRACT_ENVSTORAGE        = "EnvStorage"
	CONTRACT_ENVSTORAGE_IMP    = "EnvStorageImp"
)

type GovDomain string

const (
	DOMAIN_Gov           = "GovernanceContract"
	DOMAIN_NCPExit       = "NCPExit"
	DOMAIN_Staking       = "Staking"
	DOMAIN_BallotStorage = "BallotStorage"
	DOMAIN_EnvStorage    = "EnvStorage"

	DOMAIN_StakingReward = "StakingReward"
	DOMAIN_Ecosystem     = "Ecosystem"
	DOMAIN_Maintenance   = "Maintenance"
	DOMAIN_FeeCollector  = "FeeCollector"
)
