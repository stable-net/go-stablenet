package main

import (
	"flag"
	"fmt"
	"path/filepath"

	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
)

const pkg string = "gov"

var (
	rootFlag         = flag.String("root", "../contracts-wbft", "")
	openZeppelinFlag = flag.String("openZeppelin", "../contracts", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	openZeppelin := *openZeppelinFlag
	bindDir := filepath.Join(root, "../../bind")
	codeDir := filepath.Join(root, "../../governance-wbft/govcontracts")
	if contracts, err := compile.Compile(openZeppelin,
		filepath.Join(root, "GovStaking.sol"),
		filepath.Join(root, "GovNCP.sol"),
		filepath.Join(root, "GovConst.sol"),
		filepath.Join(root, "GovRewardee.sol"),
		filepath.Join(root, "GovRewardeeImp.sol"),
	); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(bindDir, "gen_govStaking_abi.go"), gov.CONTRACT_GOV_STAKING); err != nil {
		panic(err)
	} else if err := contracts.ExportContractCode(codeDir, gov.CONTRACT_GOV_STAKING); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(bindDir, "gen_govNCP_abi.go"), gov.CONTRACT_GOV_NCP); err != nil {
		panic(err)
	} else if err := contracts.ExportContractCode(codeDir, gov.CONTRACT_GOV_NCP); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(bindDir, "gen_govConst_abi.go"), gov.CONTRACT_GOV_CONST); err != nil {
		panic(err)
	} else if err := contracts.ExportContractCode(codeDir, gov.CONTRACT_GOV_CONST); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(bindDir, "gen_govRewardee_abi.go"), gov.CONTRACT_GOV_REWARDEE); err != nil {
		panic(err)
	} else if err := contracts.ExportContractCode(codeDir, gov.CONTRACT_GOV_REWARDEE); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(bindDir, "gen_govRewardeeImp_abi.go"), gov.CONTRACT_GOV_REWARDEE_IMP); err != nil {
		panic(err)
	} else if err := contracts.ExportContractCode(codeDir, gov.CONTRACT_GOV_REWARDEE_IMP); err != nil {
		panic(err)
	}
	fmt.Println("success!")
}
