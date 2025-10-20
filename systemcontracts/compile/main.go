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

	sc "github.com/ethereum/go-ethereum/systemcontracts"
	compile "github.com/ethereum/go-ethereum/systemcontracts/compile/compiler"
)

var (
	rootFlag         = flag.String("root", "../solidity", "")
	openZeppelinFlag = flag.String("openZeppelin", "../solidity/openzeppelin", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	versions := []string{sc.SYSTEM_CONTRACT_VERSION_1, sc.SYSTEM_CONTRACT_VERSION_2}
	srcFiles := [][]string{
		{ // v1
			filepath.Join(filepath.Join(root, versions[0]), "GovValidator.sol"),
		},
		{ // v2
			filepath.Join(filepath.Join(root, versions[0]), "GovValidator.sol"),
		},
	}
	contractBins := [][]string{
		{ // v1
			sc.CONTRACT_GOV_VALIDATOR,
		},
		{ // v2
			sc.CONTRACT_GOV_VALIDATOR,
		},
	}
	openZeppelin := *openZeppelinFlag

	for i, version := range versions {
		codeDir := filepath.Join(root, "../artifacts/"+version)
		if compiledContracts, err := compile.Compile(openZeppelin, srcFiles[i]...); err != nil {
			panic(err)
		} else if err := compiledContracts.ExportContractCode(codeDir, contractBins[i]); err != nil {
			panic(err)
		}
	}
	fmt.Println("success!")
}
