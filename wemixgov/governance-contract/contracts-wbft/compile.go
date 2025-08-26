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

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
)

var (
	rootFlag         = flag.String("root", "../contracts-wbft", "")
	openZeppelinFlag = flag.String("openZeppelin", "../contracts", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	versions := []string{govwbft.GOV_CONTRACT_VERSION_1, govwbft.GOV_CONTRACT_VERSION_2}
	srcFiles := [][]string{
		{ // v1
			filepath.Join(filepath.Join(root, versions[0]), "GovStaking.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovNCP.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovConfig.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovRewardee.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovRewardeeImp.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "OperatorSample.sol"),
		},
		{ // v2
			filepath.Join(filepath.Join(root, versions[1]), "GovStaking.sol"),
		},
	}
	contractBins := [][]string{
		{ // v1
			govwbft.CONTRACT_GOV_STAKING,
			govwbft.CONTRACT_GOV_NCP,
			govwbft.CONTRACT_GOV_CONFIG,
			govwbft.CONTRACT_GOV_REWARDEE,
			govwbft.CONTRACT_GOV_REWARDEE_IMP,
			govwbft.CONTRACT_OPERATOR_SAMPLE,
		},
		{ // v2
			govwbft.CONTRACT_GOV_STAKING,
		},
	}
	openZeppelin := *openZeppelinFlag

	for i, version := range versions {
		codeDir := filepath.Join(root, "../../governance-wbft/govcontracts/"+version)
		if compiledContracts, err := compile.Compile(openZeppelin, srcFiles[i]...,
		); err != nil {
			panic(err)
		} else if err := compiledContracts.ExportContractCode(codeDir, contractBins[i]); err != nil {
			panic(err)
		}
	}
	fmt.Println("success!")
}
