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
	rootFlag = flag.String("root", "../contracts", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	outDir := filepath.Join(root, "../../bind")
	if contracts, err := compile.Compile(root,
		filepath.Join(root, "Registry.sol"),
		filepath.Join(root, "Gov.sol"),
		filepath.Join(root, "GovImp.sol"),
		filepath.Join(root, "NCPExit.sol"),
		filepath.Join(root, "NCPExitImp.sol"),
		filepath.Join(root, "Staking.sol"),
		filepath.Join(root, "StakingImp.sol"),
		filepath.Join(root, "storage", "BallotStorage.sol"),
		filepath.Join(root, "storage", "BallotStorageImp.sol"),
		filepath.Join(root, "storage", "EnvStorage.sol"),
		filepath.Join(root, "storage", "EnvStorageImp.sol"),
	); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_registry_abi.go"), gov.CONTRACT_REGISTRY); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_gov_abi.go"), gov.CONTRACT_GOV, gov.CONTRACT_GOV_IMP); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_ncpExit_abi.go"), gov.CONTRACT_NCPEXIT, gov.CONTRACT_NCPEXIT_IMP); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_staking_abi.go"), gov.CONTRACT_STAKING, gov.CONTRACT_STAKING_IMP); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_ballotStorage_abi.go"), gov.CONTRACT_BALLOTSTORAGE, gov.CONTRACT_BALLOTSTORAGE_IMP); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_envStorage_abi.go"), gov.CONTRACT_ENVSTORAGE, gov.CONTRACT_ENVSTORAGE_IMP); err != nil {
		panic(err)
	} else {
		fmt.Println("success!")
	}
}
