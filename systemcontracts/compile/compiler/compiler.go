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

package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/systemcontracts/compile/solcdownloader"
	"github.com/pkg/errors"
)

var (
	solcVersion string = "0.8.14"
)

func Compile(openzeppelinPath string, sourceFiles ...string) (compiledTy, error) {
	solcPath, err := solcdownloader.GetSolcBin(solcVersion)
	if err != nil {
		return nil, err
	}
	if openzeppelinPath == "" {
		openzeppelinPath = "../solidity/openzeppelin"
	}
	args := []string{
		"--combined-json", "bin,bin-runtime,srcmap,srcmap-runtime,abi,userdoc,devdoc,metadata,hashes",
		"--optimize",                // code optimizer switched on
		"--allow-paths", ".,./,../", //default to support relative path： ./  ../  .
		fmt.Sprintf("@openzeppelin/contracts/=%s/contracts/contracts/", openzeppelinPath),
		fmt.Sprintf("@openzeppelin/contracts-upgradeable/=%s/contracts-upgradeable/contracts/", openzeppelinPath),
		"--",
	}
	// ~/.solc-bin/solc-0.8.14
	cmd := exec.Command(solcPath, append(args, sourceFiles...)...)

	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("solc: %v\n%s", err, stderr.Bytes())
	}

	// compiler version
	var compilerVersion = struct {
		Version string
	}{}

	if err := json.Unmarshal(stdout.Bytes(), &compilerVersion); err != nil {
		return nil, err
	} else if contracts, err := compiler.ParseCombinedJSON(stdout.Bytes(), "", "", compilerVersion.Version, strings.Join(args, " ")); err != nil {
		return nil, err
	} else {
		out := make(compiledTy)
		for name, v := range contracts {
			n := strings.Split(name, ":")[1]
			if _, ok := out[n]; ok {
				return nil, fmt.Errorf("duplicated contract name: %s", name)
			} else {
				out[n] = v
			}
		}
		return out, nil
	}
}

type compiledTy map[string]*compiler.Contract

func (compiled compiledTy) BindContracts(pkg, filename string, contracts ...string) error {
	var (
		length    = len(contracts)
		types     = make([]string, length)
		abis      = make([]string, length)
		bytecodes = make([]string, length)
		sigs      = make([]map[string]string, length)
	)

	var err error
	for i, name := range contracts {
		contract, ok := compiled[name]
		if !ok {
			return fmt.Errorf("not found contract : %v", name)
		}
		types[i] = name
		if abis[i], err = abiToString(contract); err != nil {
			return errors.Wrap(err, name)
		}
		bytecodes[i] = contract.Code
		sigs[i] = contract.Hashes
	}

	str, err := bind.Bind(types, abis, bytecodes, sigs, pkg, bind.LangGo, nil, nil)
	if err != nil {
		return err
	}

	filedata := []byte(str)

	return writeFile(filename, filedata)
}

func (compiled compiledTy) ExportContractCode(outDir string, outFiles []string) error {
	for _, outFileName := range outFiles {
		contract := compiled[outFileName]
		err := writeFile(filepath.Join(outDir, outFileName), []byte(contract.RuntimeCode))
		if err != nil {
			return fmt.Errorf("failed to write contract %s: %v", outFileName, err)
		}
	}
	return nil
}

func abiToString(contract *compiler.Contract) (abi string, err error) {
	switch v := contract.Info.AbiDefinition.(type) {
	case string:
		abi = v
	default:
		var bytes []byte
		if bytes, err = json.Marshal(v); err == nil {
			abi = string(bytes)
		}
	}
	return
}

func writeFile(name string, data []byte) error {
	if read, err := os.ReadFile(name); err == nil {
		// if out file is already exists, compare the file contents
		if crypto.Keccak256Hash(read) == crypto.Keccak256Hash(data) {
			return nil
		}
	} else {
		// check dir is exist
		outDir := filepath.Dir(name)
		if _, err := os.Stat(outDir); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err = os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
		}
	}

	return os.WriteFile(name, data, 0644)
}
