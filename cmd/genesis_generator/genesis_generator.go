package main

import (
	"bytes"
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
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
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
	fmt.Println("Which consensus engine to use? (default = Wbft)")
	fmt.Println(" 1. WBFT (wemix-byzantine-fault-tolerance)")
	fmt.Println(" 2. WEMIX (wemix-byzantine-fault-tolerance), merged from Wemix3.0 (proof-of-authority)")
	fmt.Println(" 3. Ethash (proof-of-work)")
	fmt.Println(" 4. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)")
	fmt.Println(" 5. Clique (proof-of-authority)")
	fmt.Println(" 6. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)")

	choice := read()
	switch {
	case choice == "1" || choice == "":
		g.wbftChainConfig()
		// allocate governanace contract code in genesis block
		g.Genesis.Alloc[govwbft.GovConstAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovConstContract), Balance: common.Big0}
		g.Genesis.Alloc[govwbft.GovStakingAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovStakingContract), Balance: common.Big0}

	case choice == "2":
		g.wbftChainConfig()
		fmt.Println()
		fmt.Println("Enter block number you want to enable Montblanc Fork (default 1)")
		montblancBlock := readDefaultBigInt(common.Big1)
		g.Genesis.Config.MontBlancBlock = montblancBlock
		// allocate governanace contract code in genesis block if montblanc block is genesis block
		if montblancBlock.Cmp(common.Big0) == 0 {
			g.Genesis.Alloc[govwbft.GovConstAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovConstContract), Balance: common.Big0}
			g.Genesis.Alloc[govwbft.GovStakingAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovStakingContract), Balance: common.Big0}
		}

	case choice == "3":
		g.ethashConfig()

	case choice == "4":
		g.ethashConfig()
		g.beaconChainConfig()

	case choice == "5":
		g.cliqueConfig()

	case choice == "6":
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

func (g *genesisGenerator) wbftChainConfig() {
	g.Genesis.Difficulty = types.QBFTDefaultDifficulty
	fmt.Println()
	fmt.Println("Which accounts are allowed to seal? (mandatory at least one)")

	var validators []common.Address
	for {
		if address := readAddress(); address != nil {
			validators = append(validators, *address)
			continue
		}
		if len(validators) > 0 {
			break
		}
	}

	g.Genesis.Config.QBFT = &params.QBFTConfig{
		BlockReward:           (*math.HexOrDecimal256)(big.NewInt(params.Ether)),
		EpochLength:           100,
		BlockPeriodSeconds:    1,
		RequestTimeoutSeconds: 2,
		ProposerPolicy:        0,
		Validators:            validators,
	}

	// make extra data
	vanity := append(g.Genesis.ExtraData, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(g.Genesis.ExtraData))...)
	ist := &types.QBFTExtra{
		VanityData:        vanity,
		Validators:        validators,
		PrevRound:         0,
		PrevPreparedSeal:  [][]byte{},
		PrevCommittedSeal: [][]byte{},
		Round:             0,
		PreparedSeal:      [][]byte{},
		CommittedSeal:     [][]byte{},
	}

	istPayload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		log.Crit("failed to encode qbft extra")
	}
	g.Genesis.ExtraData = istPayload

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
	} else {
		fmt.Println(buf.String())
	}
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
