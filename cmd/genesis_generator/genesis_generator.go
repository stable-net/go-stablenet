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
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type genesisGenerator struct {
	Genesis *core.Genesis `json:"genesis,omitempty"`
}

func newUint64(val uint64) *uint64 { return &val }

func makeGenerator(network string) *genesisGenerator {
	// Construct a default genesis block
	return &genesisGenerator{
		Genesis: &core.Genesis{
			Timestamp:  uint64(time.Now().Unix()),
			GasLimit:   4700000,
			Difficulty: big.NewInt(524288),
			Alloc:      make(types.GenesisAlloc),
			Config: &params.ChainConfig{
				HomesteadBlock:      big.NewInt(0),
				EIP150Block:         big.NewInt(0),
				EIP155Block:         big.NewInt(0),
				EIP158Block:         big.NewInt(0),
				ByzantiumBlock:      big.NewInt(0),
				ConstantinopleBlock: big.NewInt(0),
				PetersburgBlock:     big.NewInt(0),
				IstanbulBlock:       big.NewInt(0),
				MuirGlacierBlock:    big.NewInt(0),
				BerlinBlock:         big.NewInt(0),
				LondonBlock:         big.NewInt(0),
				ArrowGlacierBlock:   big.NewInt(0),
				GrayGlacierBlock:    big.NewInt(0),
			},
		},
	}
}

func (g *genesisGenerator) run() {
	fmt.Println("+-----------------------------------------------------------+")
	fmt.Println("| Genesis Generator is a tool that generates an genesis file  |")
	fmt.Println("| according to the desired consensus engines in wemix chain   |")
	fmt.Println("| from your cli inputs.                                       |")
	fmt.Println("| Don't wander through vast docs, just simply generate it!    |")
	fmt.Println("+-----------------------------------------------------------+")
	fmt.Println()

	g.makeGenesis()
}

func (g *genesisGenerator) makeGenesis() {
	// Figure out which consensus engine to choose
	fmt.Println()
	fmt.Println("Which consensus engine to use? (default = Anzeon)")
	fmt.Println(" 1. Anzeon (WBFT for StableNet)")
	fmt.Println(" 2. Ethash (proof-of-work)")
	fmt.Println(" 3. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)")
	fmt.Println(" 4. Clique (proof-of-authority)")
	fmt.Println(" 5. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)")

	choice := read()
	switch {
	case choice == "1" || choice == "":
		fmt.Println()
		fmt.Println("Which type of network would you like to configure? (default = single-node)")
		fmt.Println(" 1. single-node")
		fmt.Println(" 2. multi-node")

		typeChoice := read()
		switch {
		case typeChoice == "1" || typeChoice == "":
			g.wbftSingleNodeConfig()
			return
		case typeChoice == "2":
			g.wbftChainConfig()
		}

	case choice == "2":
		g.ethashConfig()

	case choice == "3":
		g.ethashConfig()
		g.beaconChainConfig()

	case choice == "4":
		g.cliqueConfig()

	case choice == "5":
		g.cliqueConfig()
		g.beaconChainConfig()

	default:
		log.Crit("Invalid consensus engine choice", "choice", choice)
	}
	// Consensus all set, just ask for initial funds and go
	fmt.Println()
	fmt.Println("Which accounts should be pre-funded? (advisable at least one)")
	for {
		// Read the address of the account to fund
		if address := readAddress(); address != nil {
			g.Genesis.Alloc[*address] = types.Account{
				Balance: new(big.Int).Lsh(big.NewInt(1), 256-7), // 2^256 / 128 (allow many pre-funds without balance overflows)
			}
			continue
		}
		break
	}

	// Query the user for some custom extras
	fmt.Println()
	fmt.Println("Specify your chain/network ID if you want an explicit one (default = random)")
	g.Genesis.Config.ChainID = new(big.Int).SetUint64(uint64(readDefaultInt(rand.Intn(65536))))

	// All done, store the genesis and flush to disk
	log.Info("Configured new genesis block")

	fmt.Println()
	fmt.Println(" Do you want to export generated genesis file?")
	fmt.Println(" If not it will be just printed (default true)")

	if readDefaultYesNo(true) {
		fmt.Println()
		fmt.Printf("Which folder to save the genesis spec into? (default = current)\n")
		fmt.Printf("It will create genesis.json\n")

		folder := readDefaultString(".")
		g.genGenesisFile(folder)
	} else {
		g.genGenesisFile("")
	}
}

type NodeAccount struct {
	address   common.Address
	blsPubKey string
}

func deriveAccount(nodeKey *ecdsa.PrivateKey) (*NodeAccount, error) {
	address := crypto.PubkeyToAddress(nodeKey.PublicKey)
	blsKey, err := bls.DeriveFromECDSA(nodeKey)
	if err != nil {
		return nil, err
	}
	blsPubKeyBytes := blsKey.PublicKey().Marshal()
	blsPubKey := hexutil.Encode(blsPubKeyBytes)

	return &NodeAccount{address, blsPubKey}, nil
}

func (g *genesisGenerator) setAnzeonConfig(validators []common.Address, blsPublicKeys []string, quorum int) {
	g.Genesis.Config.Anzeon = params.DefaultAnzeonConfig

	var vals, blsKeys string
	for i, val := range validators {
		g.Genesis.Config.Anzeon.Init.Validators = append(g.Genesis.Config.Anzeon.Init.Validators, val)
		g.Genesis.Config.Anzeon.Init.BLSPublicKeys = append(g.Genesis.Config.Anzeon.Init.BLSPublicKeys, blsPublicKeys[i])
		vals += val.String() + ","
		blsKeys += blsPublicKeys[i] + ","
	}
	vals = strings.TrimRight(vals, ",")
	blsKeys = strings.TrimRight(blsKeys, ",")

	quorumStr := strconv.Itoa(quorum)
	g.Genesis.Config.Anzeon.SystemContracts.GovValidator.Params = map[string]string{
		"members":       vals,
		"quorum":        quorumStr,
		"expiry":        "604800",
		"memberVersion": "1",
		"validators":    vals,
		"blsPublicKeys": blsKeys,
		"gasTip":        "5000000000000",
	}
	g.Genesis.Config.Anzeon.SystemContracts.NativeCoinAdapter.Params = map[string]string{
		"masterMinter":  "0x0000000000000000000000000000000000001002",
		"minters":       "0x0000000000000000000000000000000000001003",
		"minterAllowed": "10000000000000000000000000000",
		"name":          "KRC1",
		"symbol":        "KRC1",
		"decimals":      "18",
		"currency":      "KRW",
	}
	g.Genesis.Config.Anzeon.SystemContracts.GovMasterMinter.Params = map[string]string{
		"quorum":             quorumStr,
		"expiry":             "604800",
		"members":            vals,
		"memberVersion":      "1",
		"fiatToken":          "0x0000000000000000000000000000000000001000",
		"maxMinterAllowance": "10000000000000000000000000000",
	}
	g.Genesis.Config.Anzeon.SystemContracts.GovMinter.Params = map[string]string{
		"quorum":        quorumStr,
		"expiry":        "604800",
		"members":       vals,
		"memberVersion": "1",
		"fiatToken":     "0x0000000000000000000000000000000000001000",
	}
}

func genDefaultConfigFile() {
	var buf bytes.Buffer
	// Write Node.P2P section with StaticNodes
	buf.WriteString("[Node.P2P]\n")
	buf.WriteString("StaticNodes = [\n")
	buf.WriteString("\n]\n")

	writeConfigFile(buf, ".")
}

func (g *genesisGenerator) wbftSingleNodeConfig() {
	fmt.Println()
	fmt.Println("Enter the path to the nodekey file to use (default = ./nodekey):")

	var account *NodeAccount
	for {
		keyPath := readDefaultString("./nodekey")
		nodeKey, err := crypto.LoadECDSA(keyPath)
		if err != nil {
			fmt.Printf("Failed to load nodekey from '%s': %v\n", keyPath, err)
			continue
		}
		account, err = deriveAccount(nodeKey)
		if err != nil {
			fmt.Printf("Failed to derive account from nodekey: %v\n", err)
			continue
		}
		break
	}

	validators := []common.Address{account.address}
	blsPubKeys := []string{account.blsPubKey}

	g.Genesis.Difficulty = types.WBFTDefaultDifficulty

	g.setAnzeonConfig(validators, blsPubKeys, 1)

	genDefaultConfigFile()

	g.Genesis.Alloc[account.address] = types.Account{
		Balance: new(big.Int).Lsh(big.NewInt(1), 256-7), // 2^256 / 128 (allow many pre-funds without balance overflows)
	}

	g.Genesis.Config.ChainID = new(big.Int).SetUint64(uint64(rand.Intn(65536)))

	g.genGenesisFile(".")

	fmt.Println("genesis.json successfully generated")
}

func (g *genesisGenerator) wbftChainConfig() {
	g.Genesis.Difficulty = types.WBFTDefaultDifficulty

	fmt.Println()
	fmt.Println("Which accounts are allowed to seal? (mandatory at least one)")

	var validators []common.Address
	var blsPublicKeys []string
	for {
		if address := readAddress(); address != nil {
			validators = append(validators, *address)
			blsPublicKeys = append(blsPublicKeys, readBLSPubKey())
			continue
		}
		if len(validators) > 0 {
			break
		}
	}

	minQuorum := 2
	if len(validators) == 1 {
		minQuorum = 1
	}

	fmt.Println()
	fmt.Printf("Enter the quorum for governance: (default = %d)\n", minQuorum)

	var quorum int
	for {
		quorum = readDefaultInt(minQuorum)
		if quorum < minQuorum {
			fmt.Printf("Quorum must be at least %d\n", minQuorum)
			continue
		}
		if quorum > len(validators) {
			fmt.Printf("Quorum cannot exceed number of validators: %d\n", len(validators))
			continue
		}
		break
	}

	g.setAnzeonConfig(validators, blsPublicKeys, quorum)

	// you can add config file for static nodes if you want
	fmt.Println()
	fmt.Println("Want to generate config.toml file to configure static nodes to connect?")
	fmt.Println("Else you have to manage peer node manually (default true)")
	if readDefaultYesNo(true) {
		genConfigFile()
	}
}

func (g *genesisGenerator) beaconChainConfig() {
	fmt.Println()
	fmt.Println("Do you want to start beacon chain immediately? (default yes)")
	if readDefaultYesNo(true) {
		g.Genesis.Config.TerminalTotalDifficulty = common.Big0
		g.Genesis.Config.TerminalTotalDifficultyPassed = true
		g.Genesis.Config.ShanghaiTime = newUint64(0)
		g.Genesis.Config.CancunTime = newUint64(0)
	} else {
		g.Genesis.Config.TerminalTotalDifficultyPassed = false
		fmt.Println()
		fmt.Println("Enter TerminalTotalDifficulty value you want to set (default current TTD*10)")
		g.Genesis.Config.TerminalTotalDifficulty = readDefaultBigInt(new(big.Int).Mul(g.Genesis.Difficulty, big.NewInt(10)))
		fmt.Println()
		fmt.Println("Enter timestamp you want to enable Shanghai Fork (default currentTimeStamp+10000)")

		g.Genesis.Config.ShanghaiTime = newUint64(uint64(readDefaultInt(int(time.Now().Unix()) + 10000)))
		fmt.Println()
		fmt.Println("Enter timestamp you want to enable Cancun Fork (default currentTimeStamp+10000)")
		g.Genesis.Config.CancunTime = newUint64(uint64(readDefaultInt(int(time.Now().Unix()) + 10000)))
	}
}

func (g *genesisGenerator) ethashConfig() {
	g.Genesis.Config.Ethash = new(params.EthashConfig)
	g.Genesis.ExtraData = make([]byte, 32)
}

func (g *genesisGenerator) cliqueConfig() {
	g.Genesis.Difficulty = big.NewInt(1)
	g.Genesis.Config.Clique = &params.CliqueConfig{
		Period: 15,
		Epoch:  30000,
	}
	fmt.Println()
	fmt.Println("How many seconds should blocks take? (default = 15)")
	g.Genesis.Config.Clique.Period = uint64(readDefaultInt(15))

	// We also need the initial list of signers
	fmt.Println()
	fmt.Println("Which accounts are allowed to seal? (mandatory at least one)")

	var signers []common.Address
	for {
		if address := readAddress(); address != nil {
			signers = append(signers, *address)
			continue
		}
		if len(signers) > 0 {
			break
		}
	}
	// Sort the signers and embed into the extra-data section
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if bytes.Compare(signers[i][:], signers[j][:]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	g.Genesis.ExtraData = make([]byte, 32+len(signers)*common.AddressLength+65)
	for i, signer := range signers {
		copy(g.Genesis.ExtraData[32+i*common.AddressLength:], signer[:])
	}

	// you can add config file for static nodes if you want
	fmt.Println()
	fmt.Println("Want to generate config.toml file to configure static nodes to connect?")
	fmt.Println("Else you have to manage peer node manually (default true)")
	if readDefaultYesNo(true) {
		genConfigFile()
	}
}

// flush dumps the contents of config to disk or print.
func (g *genesisGenerator) genGenesisFile(folder string) {
	out, _ := json.MarshalIndent(g.Genesis, "", "  ")

	if folder != "" {
		if err := os.MkdirAll(folder, 0755); err != nil {
			log.Error("Failed to create spec folder", "folder", folder, "err", err)
			return
		}
		gethJson := filepath.Join(folder, "genesis.json")
		if err := os.WriteFile(gethJson, out, 0644); err != nil {
			log.Error("Failed to save genesis file", "err", err)
			return
		}
		log.Info("Saved native genesis chain spec", "path", gethJson)
	} else {
		fmt.Println(string(out))
	}
}

// genConfigFile creates config.toml file that defines Node.P2P.StaticNodes
func genConfigFile() {
	// Create a buffer to write TOML content
	var buf bytes.Buffer
	// Write Node.P2P section with StaticNodes
	buf.WriteString("[Node.P2P]\n")
	buf.WriteString("StaticNodes = [\n")

	fmt.Println()
	fmt.Println("Enter enode URLs for static nodes (press enter with empty input when done):")
	var enodes []string
	for {
		enode := readDefaultString("")
		if enode == "" {
			break
		}
		if validateEnodeURL(enode) {
			enodes = append(enodes, fmt.Sprintf("    %q", enode))
		} else {
			fmt.Println("Invalid enode URL. Try again:")
			continue
		}
	}

	buf.WriteString(strings.Join(enodes, ",\n"))
	buf.WriteString("\n]\n")

	fmt.Println()
	fmt.Println(" Do you want to export generated config file?")
	fmt.Println(" If not it will be just printed (default true)")

	if readDefaultYesNo(true) {
		fmt.Println()
		fmt.Printf("Which folder to save the config.toml into? (default = current)\n")
		folder := readDefaultString(".")
		writeConfigFile(buf, folder)
	} else {
		fmt.Println(buf.String())
	}
}

func writeConfigFile(buf bytes.Buffer, folder string) {
	if err := os.MkdirAll(folder, 0755); err != nil {
		log.Error("Failed to create spec folder", "folder", folder, "err", err)
		return
	}
	configPath := filepath.Join(folder, "config.toml")
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		log.Error("Failed to save config file", "err", err)
		return
	}
	log.Info("Saved config.toml file", "path", configPath)
}

// validateEnodeURL checks if the given string is a valid enode URL
func validateEnodeURL(enode string) bool {
	if !strings.HasPrefix(enode, "enode://") {
		log.Error("Invalid enode URL: must start with 'enode://'")
		return false
	}

	u, err := url.Parse(enode)
	if err != nil {
		log.Error("Invalid enode URL format", "err", err)
		return false
	}

	// Check if the hex part is valid (should be 128 characters after enode://)
	if len(u.User.String()) != 128 {
		log.Error("Invalid public key in enode URL: must be 128 hex characters")
		return false
	}

	// Validate the host:port part
	hostPort := u.Host
	if hostPort == "" {
		log.Error("Missing host:port in enode URL")
		return false
	}

	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		log.Error("Invalid host:port format in enode URL", "err", err)
		return false
	}

	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		log.Error("Invalid port number in enode URL")
		return false
	}

	// Validate IP address
	if net.ParseIP(host) == nil {
		log.Error("Invalid IP address in enode URL")
		return false
	}

	return true
}
