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
