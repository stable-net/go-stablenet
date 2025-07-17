// Modification Copyright 2024 The Wemix Authors
// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//

package params

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params/forks"
)

// Genesis hashes to enforce below configs on.
var (
	WemixMainnetGenesisHash = common.HexToHash("0xa4e17f33b057aa2f78685aba4d6a60a79fc0a366fe8a521a3af1d6eceb1cf3cd")
	WemixTestnetGenesisHash = common.HexToHash("0x15f522870e65c4e66230341b71b63ca421a61d4b5432f59d4906f021802435c6")
	MainnetGenesisHash      = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	HoleskyGenesisHash      = common.HexToHash("0xb5f7f912443c940f21fd611f12828d75b534364ed9e95ca4e307729a4661bde4")
	SepoliaGenesisHash      = common.HexToHash("0x25a5cc106eea7138acab33231d7160d69cb777ee0c2c553fcddf5138993e6dd9")
	GoerliGenesisHash       = common.HexToHash("0xbf7e331f7f7c1dd2e05159666b3bf8bc7a8a3a9eb1d518969eab529dd9b88c1a")
)

func newUint64(val uint64) *uint64 { return &val }
func newBool(val bool) *bool       { return &val }

var (
	MainnetTerminalTotalDifficulty, _ = new(big.Int).SetString("58_750_000_000_000_000_000_000", 0)

	WemixMainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1111),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0), // We shouldn't have applied DAO hard fork because we didn't have to
		DAOForkSupport:      true,          // Why on earth?
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
		MergeNetsplitBlock:  nil,
		ShanghaiTime:        nil,
		CancunTime:          nil,
		PragueTime:          nil,
		VerkleTime:          nil,
		PangyoBlock:         big.NewInt(0),
		ApplepieBlock:       big.NewInt(20_476_911),
		BriocheBlock:        big.NewInt(53_525_500),  // target date: 24-07-01 00:00:00 (GMT+09)
		CroissantBlock:      big.NewInt(100_000_000), // TODO: decide the block number
		Ethash:              new(EthashConfig),
		Brioche: &BriocheConfig{
			BlockReward:       big.NewInt(1e18),
			FirstHalvingBlock: big.NewInt(53_525_500),
			HalvingPeriod:     big.NewInt(63_115_200),
			FinishRewardBlock: big.NewInt(2_467_714_000), // target date: 2101-01-01 00:00:00 (GMT+09)
			HalvingTimes:      16,
			HalvingRate:       50,
		},
		Croissant: &CroissantConfig{
			WBFT: &WBFTConfig{ // TODO: this is just for test on mainnet
				EpochLength:              100,
				BlockPeriodSeconds:       1,
				RequestTimeoutSeconds:    1000,
				ProposerPolicy:           newUint64(0),
				BlockReward:              (*math.HexOrDecimal256)(big.NewInt(1000000000000000000)),
				MaxRequestTimeoutSeconds: &mrts,
				BlockRewardBeneficiary: &BeneficiaryInfo{Denominator: 10000, Beneficiaries: []*Beneficiary{{
					Name:      "Maintenance",
					Addr:      common.HexToAddress("0x1620cf4bD57087236025516cdd8D70e127da5331"),
					Numerator: 2500,
				}, {
					Name:      "EcoSystem",
					Addr:      common.HexToAddress("0xC79d535f6EDD3E0Fa648D6169eA5b8e1Aa38921e"),
					Numerator: 2500,
				}}},
				TargetValidators:            newUint64(1), // TODO: define validators
				StabilizingStakersThreshold: newUint64(1), // TODO: define min stakers
			},
			Init: &WbftInit{
				Validators:    []common.Address{common.HexToAddress("0x5b5682ab6952f96f5e68c7dd34c8018c71748248")}, // TODO: define initial validators
				BLSPublicKeys: []string{"0x935344a9e431d256fd4fcb819fd5497fb80ce4cd402b4f93ea0cd585dfb4dc433e962a55a153f8c041a773304ef8833d"},
			},
			GovContracts: &GovContracts{
				GovConfig: &GovContract{
					Address: DefaultGovConfigAddress,
					Version: DefaultGovVersion,
					Params: map[string]string{
						"minimumStaking":           "500000000000000000000000",
						"maximumStaking":           "1000000000000000000000000000",
						"unbondingPeriodStaker":    "604800", // 7 days
						"unbondingPeriodDelegator": "259200", // 3 days
						"feePrecision":             "10000",  // 0.01%
						"changeFeeDelay":           "604800", // 7 days
						"govCouncil":               DefaultGovNCPAddress.String(),
					},
				},
				GovStaking: &GovContract{
					Address: DefaultGovStakingAddress,
					Version: DefaultGovVersion,
				},
				GovRewardeeImp: &GovContract{
					Address: DefaultGovRewardeeImpAddress,
					Version: DefaultGovVersion,
				},
				GovNCP: &GovContract{
					Address: DefaultGovNCPAddress,
					Version: DefaultGovVersion,
					Params: map[string]string{
						"ncps": "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697", // comma separated; TODO: define NCPs
					},
				},
			},
		},
	}

	// WemixTestnetChainConfig contains the chain parameters to run a node on the Wemix test network.
	WemixTestnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1112),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		DAOForkSupport:      true,
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
		PangyoBlock:         big.NewInt(10_000_000),
		ApplepieBlock:       big.NewInt(26_240_268),
		BriocheBlock:        big.NewInt(59_414_700),  // target date: 24-06-04 11:00:41 (GMT+09)
		CroissantBlock:      big.NewInt(100_000_000), // TODO: decide the block number
		Ethash:              new(EthashConfig),
		Brioche: &BriocheConfig{
			BlockReward:       big.NewInt(1e18),
			FirstHalvingBlock: big.NewInt(59_414_700),
			HalvingPeriod:     big.NewInt(63_115_200),
			FinishRewardBlock: big.NewInt(2_473_258_000), // target date: 2100-12-01 11:02:21 (GMT+09)
			HalvingTimes:      16,
			HalvingRate:       50,
		},
		Croissant: &CroissantConfig{
			WBFT: &WBFTConfig{ // TODO: this is just for test on mainnet
				EpochLength:              100,
				BlockPeriodSeconds:       1,
				RequestTimeoutSeconds:    1000,
				ProposerPolicy:           newUint64(0),
				BlockReward:              (*math.HexOrDecimal256)(big.NewInt(1000000000000000000)),
				MaxRequestTimeoutSeconds: &mrts,
				BlockRewardBeneficiary: &BeneficiaryInfo{Denominator: 10000, Beneficiaries: []*Beneficiary{{
					Name:      "Maintenance",
					Addr:      common.HexToAddress("0x1620cf4bD57087236025516cdd8D70e127da5331"),
					Numerator: 2500,
				}, {
					Name:      "EcoSystem",
					Addr:      common.HexToAddress("0xC79d535f6EDD3E0Fa648D6169eA5b8e1Aa38921e"),
					Numerator: 2500,
				}}},
				TargetValidators:            newUint64(1), // TODO: define validators
				StabilizingStakersThreshold: newUint64(1),
			},
			Init: &WbftInit{
				Validators:    []common.Address{common.HexToAddress("0x5b5682ab6952f96f5e68c7dd34c8018c71748248")}, // TODO: define initial validators
				BLSPublicKeys: []string{"0x935344a9e431d256fd4fcb819fd5497fb80ce4cd402b4f93ea0cd585dfb4dc433e962a55a153f8c041a773304ef8833d"},
			},
			GovContracts: &GovContracts{
				GovConfig: &GovContract{
					Address: DefaultGovConfigAddress,
					Version: DefaultGovVersion,
					Params: map[string]string{
						"minimumStaking":           "500000000000000000000000",
						"maximumStaking":           "1000000000000000000000000000",
						"unbondingPeriodStaker":    "604800", // 7 days
						"unbondingPeriodDelegator": "259200", // 3 days
						"feePrecision":             "10000",  // 0.01%
						"changeFeeDelay":           "604800", // 7 days
						"govCouncil":               DefaultGovNCPAddress.String(),
					},
				},
				GovStaking: &GovContract{
					Address: DefaultGovStakingAddress,
					Version: DefaultGovVersion,
				},
				GovRewardeeImp: &GovContract{
					Address: DefaultGovRewardeeImpAddress,
					Version: DefaultGovVersion,
				},
				GovNCP: &GovContract{
					Address: DefaultGovNCPAddress,
					Version: DefaultGovVersion,
					Params: map[string]string{
						"ncps": "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697", // comma separated
					},
				},
			},
		},
	}

	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(1),
		HomesteadBlock:                big.NewInt(1_150_000),
		DAOForkBlock:                  big.NewInt(1_920_000),
		DAOForkSupport:                true,
		EIP150Block:                   big.NewInt(2_463_000),
		EIP155Block:                   big.NewInt(2_675_000),
		EIP158Block:                   big.NewInt(2_675_000),
		ByzantiumBlock:                big.NewInt(4_370_000),
		ConstantinopleBlock:           big.NewInt(7_280_000),
		PetersburgBlock:               big.NewInt(7_280_000),
		IstanbulBlock:                 big.NewInt(9_069_000),
		MuirGlacierBlock:              big.NewInt(9_200_000),
		BerlinBlock:                   big.NewInt(12_244_000),
		LondonBlock:                   big.NewInt(12_965_000),
		ArrowGlacierBlock:             big.NewInt(13_773_000),
		GrayGlacierBlock:              big.NewInt(15_050_000),
		TerminalTotalDifficulty:       MainnetTerminalTotalDifficulty, // 58_750_000_000_000_000_000_000
		TerminalTotalDifficultyPassed: true,
		ShanghaiTime:                  newUint64(1681338455),
		CancunTime:                    newUint64(1710338135),
		Ethash:                        new(EthashConfig),
	}
	// HoleskyChainConfig contains the chain parameters to run a node on the Holesky test network.
	HoleskyChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(17000),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                true,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              nil,
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		GrayGlacierBlock:              nil,
		TerminalTotalDifficulty:       big.NewInt(0),
		TerminalTotalDifficultyPassed: true,
		MergeNetsplitBlock:            nil,
		ShanghaiTime:                  newUint64(1696000704),
		CancunTime:                    newUint64(1707305664),
		Ethash:                        new(EthashConfig),
	}
	// SepoliaChainConfig contains the chain parameters to run a node on the Sepolia test network.
	SepoliaChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(11155111),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                true,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		GrayGlacierBlock:              nil,
		TerminalTotalDifficulty:       big.NewInt(17_000_000_000_000_000),
		TerminalTotalDifficultyPassed: true,
		MergeNetsplitBlock:            big.NewInt(1735371),
		ShanghaiTime:                  newUint64(1677557088),
		CancunTime:                    newUint64(1706655072),
		Ethash:                        new(EthashConfig),
	}
	// GoerliChainConfig contains the chain parameters to run a node on the Görli test network.
	GoerliChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(5),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                true,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(1_561_651),
		MuirGlacierBlock:              nil,
		BerlinBlock:                   big.NewInt(4_460_644),
		LondonBlock:                   big.NewInt(5_062_605),
		ArrowGlacierBlock:             nil,
		TerminalTotalDifficulty:       big.NewInt(10_790_000),
		TerminalTotalDifficultyPassed: true,
		ShanghaiTime:                  newUint64(1678832736),
		CancunTime:                    newUint64(1705473120),
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}
	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	AllEthashProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		MergeNetsplitBlock:            nil,
		ShanghaiTime:                  nil,
		CancunTime:                    nil,
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       nil,
		TerminalTotalDifficultyPassed: true,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
	}

	mrts = uint64(4)

	// customized for WEMIX chain
	AllDevChainProtocolChanges = &ChainConfig{
		ChainID:             big.NewInt(1337),
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
		BriocheBlock:        big.NewInt(0),
		CroissantBlock:      big.NewInt(0),
		Brioche: &BriocheConfig{
			BlockReward:       big.NewInt(1e18),
			FirstHalvingBlock: big.NewInt(50),
			HalvingPeriod:     big.NewInt(50),
			FinishRewardBlock: big.NewInt(450),
			HalvingTimes:      8,
			HalvingRate:       50,
		},
		Croissant: &CroissantConfig{
			WBFT: &WBFTConfig{
				EpochLength:              100,
				BlockPeriodSeconds:       1,
				RequestTimeoutSeconds:    1000,
				ProposerPolicy:           newUint64(0),
				BlockReward:              (*math.HexOrDecimal256)(big.NewInt(1000000000000000000)),
				MaxRequestTimeoutSeconds: &mrts,
				BlockRewardBeneficiary: &BeneficiaryInfo{Denominator: 10000, Beneficiaries: []*Beneficiary{{
					Name:      "Wemix Foundation",
					Addr:      common.HexToAddress("0x7014F43c5BC7f7F3b4FBdf1599E5e1394548607a"),
					Numerator: 5000,
				}}},
				TargetValidators:            newUint64(1),
				StabilizingStakersThreshold: newUint64(1),
			},
			Init: &WbftInit{
				Validators:    []common.Address{common.HexToAddress("0x7014F43c5BC7f7F3b4FBdf1599E5e1394548607a")},
				BLSPublicKeys: []string{"0xb1ae18fdcbcc6a80d7a0c4cfec1a04bc1bee78e519eaadd689108077d946e0849a2c30ac96462be32023f34ca67ebcf6"},
			},
			GovContracts: &GovContracts{
				GovConfig: &GovContract{
					Address: DefaultGovConfigAddress,
					Version: DefaultGovVersion,
					Params:  DefaultGovConfigParams,
				},
				GovStaking: &GovContract{
					Address: DefaultGovStakingAddress,
					Version: DefaultGovVersion,
				},
				GovRewardeeImp: &GovContract{
					Address: DefaultGovRewardeeImpAddress,
					Version: DefaultGovVersion,
				},
				GovNCP: &GovContract{
					Address: DefaultGovNCPAddress,
					Version: DefaultGovVersion,
					Params: map[string]string{
						"ncps": "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697", // comma separated
					},
				},
			},
		},
	}

	EthashChainProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		TerminalTotalDifficultyPassed: false,
	}

	BeaconEthashChainProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		TerminalTotalDifficultyPassed: true,
	}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	AllCliqueProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		GrayGlacierBlock:              nil,
		MergeNetsplitBlock:            nil,
		ShanghaiTime:                  nil,
		CancunTime:                    nil,
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       nil,
		TerminalTotalDifficultyPassed: false,
		Ethash:                        nil,
		Clique:                        &CliqueConfig{Period: 0, Epoch: 30000},
	}

	// TestChainConfig contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers for testing purposes.
	TestChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(1111),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		MergeNetsplitBlock:            nil,
		BriocheBlock:                  big.NewInt(0),
		ShanghaiTime:                  nil,
		CancunTime:                    nil,
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       nil,
		TerminalTotalDifficultyPassed: false,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
		Brioche: &BriocheConfig{
			BlockReward:       big.NewInt(1e18),
			FirstHalvingBlock: big.NewInt(50),
			HalvingPeriod:     big.NewInt(50),
			FinishRewardBlock: big.NewInt(450),
			HalvingTimes:      8,
			HalvingRate:       50,
		},
	}

	// TestWBFTChainConfig contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers for testing purposes
	// and used for WBFT engine tests.
	TestWBFTChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(1111),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		MergeNetsplitBlock:            nil,
		BriocheBlock:                  big.NewInt(0),
		CroissantBlock:                big.NewInt(0),
		ShanghaiTime:                  nil,
		CancunTime:                    nil,
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       nil,
		TerminalTotalDifficultyPassed: false,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
		Brioche: &BriocheConfig{
			BlockReward:       big.NewInt(1e18),
			FirstHalvingBlock: big.NewInt(50),
			HalvingPeriod:     big.NewInt(50),
			FinishRewardBlock: big.NewInt(450),
			HalvingTimes:      8,
			HalvingRate:       50,
		},
		Croissant: &CroissantConfig{
			WBFT: &WBFTConfig{
				EpochLength:              100,
				BlockPeriodSeconds:       1,
				RequestTimeoutSeconds:    1000,
				ProposerPolicy:           newUint64(0),
				BlockReward:              (*math.HexOrDecimal256)(big.NewInt(1000000000000000000)),
				MaxRequestTimeoutSeconds: &mrts,
				BlockRewardBeneficiary: &BeneficiaryInfo{Denominator: 10000, Beneficiaries: []*Beneficiary{{
					Name:      "Wemix Foundation",
					Addr:      common.HexToAddress("0x7014F43c5BC7f7F3b4FBdf1599E5e1394548607a"),
					Numerator: 5000,
				}}},
				TargetValidators:            newUint64(1),
				StabilizingStakersThreshold: newUint64(1),
				UseNCP:                      newBool(false),
			},
			Init: &WbftInit{
				Validators:    []common.Address{common.HexToAddress("0x7014F43c5BC7f7F3b4FBdf1599E5e1394548607a")},
				BLSPublicKeys: []string{"0xb1ae18fdcbcc6a80d7a0c4cfec1a04bc1bee78e519eaadd689108077d946e0849a2c30ac96462be32023f34ca67ebcf6"},
			},
			GovContracts: &GovContracts{
				GovConfig: &GovContract{
					Address: DefaultGovConfigAddress,
					Version: DefaultGovVersion,
					Params:  DefaultGovConfigParams,
				},
				GovStaking: &GovContract{
					Address: DefaultGovStakingAddress,
					Version: DefaultGovVersion,
				},
				GovRewardeeImp: &GovContract{
					Address: DefaultGovRewardeeImpAddress,
					Version: DefaultGovVersion,
				},
			},
		},
	}

	// MergedTestChainConfig contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers for testing purposes.
	MergedTestChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(1),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		MergeNetsplitBlock:            big.NewInt(0),
		ShanghaiTime:                  newUint64(0),
		CancunTime:                    newUint64(0),
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       big.NewInt(0),
		TerminalTotalDifficultyPassed: true,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
	}

	// NonActivatedConfig defines the chain configuration without activating
	// any protocol change (EIPs).
	NonActivatedConfig = &ChainConfig{
		ChainID:                       big.NewInt(1),
		HomesteadBlock:                nil,
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   nil,
		EIP155Block:                   nil,
		EIP158Block:                   nil,
		ByzantiumBlock:                nil,
		ConstantinopleBlock:           nil,
		PetersburgBlock:               nil,
		IstanbulBlock:                 nil,
		MuirGlacierBlock:              nil,
		BerlinBlock:                   nil,
		LondonBlock:                   nil,
		ArrowGlacierBlock:             nil,
		GrayGlacierBlock:              nil,
		MergeNetsplitBlock:            nil,
		ShanghaiTime:                  nil,
		CancunTime:                    nil,
		PragueTime:                    nil,
		VerkleTime:                    nil,
		TerminalTotalDifficulty:       nil,
		TerminalTotalDifficultyPassed: false,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
	}
)

// NetworkNames are user friendly names to use in the chain spec banner.
var NetworkNames = map[string]string{
	WemixMainnetChainConfig.ChainID.String(): "mainnet",
	WemixTestnetChainConfig.ChainID.String(): "testnet",
	GoerliChainConfig.ChainID.String():       "goerli",
	SepoliaChainConfig.ChainID.String():      "sepolia",
	HoleskyChainConfig.ChainID.String():      "holesky",
	MainnetChainConfig.ChainID.String():      "ethmainnet",
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)
	BerlinBlock         *big.Int `json:"berlinBlock,omitempty"`         // Berlin switch block (nil = no fork, 0 = already on berlin)
	LondonBlock         *big.Int `json:"londonBlock,omitempty"`         // London switch block (nil = no fork, 0 = already on london)
	ArrowGlacierBlock   *big.Int `json:"arrowGlacierBlock,omitempty"`   // Eip-4345 (bomb delay) switch block (nil = no fork, 0 = already activated)
	GrayGlacierBlock    *big.Int `json:"grayGlacierBlock,omitempty"`    // Eip-5133 (bomb delay) switch block (nil = no fork, 0 = already activated)
	MergeNetsplitBlock  *big.Int `json:"mergeNetsplitBlock,omitempty"`  // Virtual fork after The Merge to use as a network splitter
	PangyoBlock         *big.Int `json:"pangyoBlock,omitempty"`         // Pangyo switch block (nil = no fork, 0 = already on Pangyo)
	ApplepieBlock       *big.Int `json:"applepieBlock,omitempty"`       // Applepie switch block (nil = no fork, 0 = already on Applepie)
	BriocheBlock        *big.Int `json:"briocheBlock,omitempty"`        // Brioche switch block (nil = no fork, 0 = already on Brioche)
	CroissantBlock      *big.Int `json:"croissantBlock,omitempty"`      // Croissant switch block (nil = no fork, 0 = already on Croissant)

	// Fork scheduling was switched from blocks to timestamps here

	ShanghaiTime *uint64 `json:"shanghaiTime,omitempty"` // Shanghai switch time (nil = no fork, 0 = already on shanghai)
	CancunTime   *uint64 `json:"cancunTime,omitempty"`   // Cancun switch time (nil = no fork, 0 = already on cancun)
	PragueTime   *uint64 `json:"pragueTime,omitempty"`   // Prague switch time (nil = no fork, 0 = already on prague)
	VerkleTime   *uint64 `json:"verkleTime,omitempty"`   // Verkle switch time (nil = no fork, 0 = already on verkle)

	// TerminalTotalDifficulty is the amount of total difficulty reached by
	// the network that triggers the consensus upgrade.
	TerminalTotalDifficulty *big.Int `json:"terminalTotalDifficulty,omitempty"`

	// TerminalTotalDifficultyPassed is a flag specifying that the network already
	// passed the terminal total difficulty. Its purpose is to disable legacy sync
	// even without having seen the TTD locally (safer long term).
	TerminalTotalDifficultyPassed bool `json:"terminalTotalDifficultyPassed,omitempty"`

	// Various consensus engines
	Ethash  *EthashConfig  `json:"ethash,omitempty"`
	Clique  *CliqueConfig  `json:"clique,omitempty"`
	Brioche *BriocheConfig `json:"brioche,omitempty"` // if this config is nil, brioche halving is not applied

	Croissant   *CroissantConfig `json:"croissant,omitempty"`
	Transitions []Transition     `json:"transitions,omitempty"`
}

// Brioche halving configuration
type BriocheConfig struct {
	// if the chain is on brioche hard fork, `RewardAmount` of gov contract is not used rather this BlockReward is used
	BlockReward       *big.Int `json:"blockReward,omitempty"`       // nil - use default block reward(1e18)
	FirstHalvingBlock *big.Int `json:"firstHalvingBlock,omitempty"` // nil - halving is not work. including this block
	HalvingPeriod     *big.Int `json:"halvingPeriod,omitempty"`     // nil - halving is not work
	FinishRewardBlock *big.Int `json:"finishRewardBlock,omitempty"` // nil - block reward goes on endlessly
	HalvingTimes      uint64   `json:"halvingTimes,omitempty"`      // 0 - no halving
	HalvingRate       uint32   `json:"halvingRate,omitempty"`       // 0 - no reward on halving; 100 - no halving; >100 - increasing reward
}

func (bc *BriocheConfig) GetBriocheBlockReward(defaultReward *big.Int, num *big.Int) *big.Int {
	blockReward := new(big.Int).Set(defaultReward) // default brioche block reward
	if bc != nil {
		if bc.BlockReward != nil {
			blockReward = new(big.Int).Set(bc.BlockReward)
		}
		if bc.FinishRewardBlock != nil &&
			bc.FinishRewardBlock.Cmp(num) <= 0 {
			blockReward = big.NewInt(0)
		} else if bc.FirstHalvingBlock != nil &&
			bc.HalvingPeriod != nil &&
			bc.HalvingTimes > 0 &&
			num.Cmp(bc.FirstHalvingBlock) >= 0 {
			blockReward = bc.calcHalvedReward(blockReward, num)
		}
	}
	return blockReward
}

func (bc *BriocheConfig) calcHalvedReward(baseReward *big.Int, num *big.Int) *big.Int {
	elapsed := new(big.Int).Sub(num, bc.FirstHalvingBlock)
	times := new(big.Int).Add(common.Big1, new(big.Int).Div(elapsed, bc.HalvingPeriod))
	if times.Uint64() > bc.HalvingTimes {
		times = big.NewInt(int64(bc.HalvingTimes))
	}

	reward := new(big.Int).Set(baseReward)
	numerator := new(big.Int).Exp(big.NewInt(int64(bc.HalvingRate)), times, nil)
	denominator := new(big.Int).Exp(big.NewInt(100), times, nil)
	return reward.Div(reward.Mul(reward, numerator), denominator)
}

func (bc *BriocheConfig) String() string {
	return fmt.Sprintf("{BlockReward: %v FirstHalvingBlock: %v HalvingPeriod: %v FinishRewardBlock: %v HalvingTimes: %v HalvingRate: %v}",
		bc.BlockReward,
		bc.FirstHalvingBlock,
		bc.HalvingPeriod,
		bc.FinishRewardBlock,
		bc.HalvingTimes,
		bc.HalvingRate,
	)
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// Description returns a human-readable description of ChainConfig.
func (c *ChainConfig) Description() string {
	var banner string

	// Create some basic network config output
	network := NetworkNames[c.ChainID.String()]
	if network == "" {
		network = "unknown"
	}
	banner += fmt.Sprintf("Chain ID:  %v (%s)\n", c.ChainID, network)
	switch {
	case c.Ethash != nil:
		if c.TerminalTotalDifficulty == nil {
			banner += "Consensus: Ethash (proof-of-work)\n"
		} else if !c.TerminalTotalDifficultyPassed {
			banner += "Consensus: Beacon (proof-of-stake), merging from Ethash (proof-of-work)\n"
		} else {
			banner += "Consensus: Beacon (proof-of-stake), merged from Ethash (proof-of-work)\n"
		}
	case c.Clique != nil:
		if c.TerminalTotalDifficulty == nil {
			banner += "Consensus: Clique (proof-of-authority)\n"
		} else if !c.TerminalTotalDifficultyPassed {
			banner += "Consensus: Beacon (proof-of-stake), merging from Clique (proof-of-authority)\n"
		} else {
			banner += "Consensus: Beacon (proof-of-stake), merged from Clique (proof-of-authority)\n"
		}
	case c.CroissantEnabled():
		banner += "Consensus: WBFT (wemix-byzantine-fault-tolerance)\n"
	default:
		banner += "Consensus: WEMIX PoA\n"
	}
	banner += "\n"

	// Create a list of forks with a short description of them. Forks that only
	// makes sense for mainnet should be optional at printing to avoid bloating
	// the output for testnets and private networks.
	banner += "Pre-Merge hard forks (block based):\n"
	banner += fmt.Sprintf(" - Homestead:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/homestead.md)\n", c.HomesteadBlock)
	if c.DAOForkBlock != nil {
		banner += fmt.Sprintf(" - DAO Fork:                    #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/dao-fork.md)\n", c.DAOForkBlock)
	}
	banner += fmt.Sprintf(" - Tangerine Whistle (EIP 150): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/tangerine-whistle.md)\n", c.EIP150Block)
	banner += fmt.Sprintf(" - Spurious Dragon/1 (EIP 155): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Spurious Dragon/2 (EIP 158): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Byzantium:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/byzantium.md)\n", c.ByzantiumBlock)
	banner += fmt.Sprintf(" - Constantinople:              #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/constantinople.md)\n", c.ConstantinopleBlock)
	banner += fmt.Sprintf(" - Petersburg:                  #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/petersburg.md)\n", c.PetersburgBlock)
	banner += fmt.Sprintf(" - Istanbul:                    #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/istanbul.md)\n", c.IstanbulBlock)
	if c.MuirGlacierBlock != nil {
		banner += fmt.Sprintf(" - Muir Glacier:                #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/muir-glacier.md)\n", c.MuirGlacierBlock)
	}
	banner += fmt.Sprintf(" - Berlin:                      #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/berlin.md)\n", c.BerlinBlock)
	banner += fmt.Sprintf(" - London:                      #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/london.md)\n", c.LondonBlock)
	if c.ArrowGlacierBlock != nil {
		banner += fmt.Sprintf(" - Arrow Glacier:               #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/arrow-glacier.md)\n", c.ArrowGlacierBlock)
	}
	if c.GrayGlacierBlock != nil {
		banner += fmt.Sprintf(" - Gray Glacier:                #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/gray-glacier.md)\n", c.GrayGlacierBlock)
	}
	banner += fmt.Sprintf(" - Pangyo:                      #%-8v\n", c.PangyoBlock)
	banner += fmt.Sprintf(" - Applepie:                    #%-8v\n", c.ApplepieBlock)
	banner += fmt.Sprintf(" - Brioche:                     #%-8v\n", c.BriocheBlock)
	if c.Brioche != nil {
		banner += fmt.Sprintf("   - FirstHalvingBlock:         #%-8v\n", c.Brioche.FirstHalvingBlock)
		banner += fmt.Sprintf("   - HalvingPeriod:             %-8v\n", c.Brioche.HalvingPeriod)
		banner += fmt.Sprintf("   - FinishRewardBlock:         #%-8v\n", c.Brioche.FinishRewardBlock)
		banner += fmt.Sprintf("   - HalvingTimes:              %-8v\n", c.Brioche.HalvingTimes)
		banner += fmt.Sprintf("   - HalvingRate:               %-8v\n", c.Brioche.HalvingRate)
	}
	banner += fmt.Sprintf(" - Croissant:                   #%-8v\n", c.CroissantBlock)
	if c.Croissant != nil {
		if c.Croissant.WBFT != nil {
			banner += "   - WBFT\n"
			banner += fmt.Sprintf("     - EpochLength:               %-8v\n", c.Croissant.WBFT.EpochLength)
			banner += fmt.Sprintf("     - BlockPeriodSeconds:        %-8v\n", c.Croissant.WBFT.BlockPeriodSeconds)
			banner += fmt.Sprintf("     - RequestTimeoutSeconds:     %-8v\n", c.Croissant.WBFT.RequestTimeoutSeconds)
			banner += fmt.Sprintf("     - ProposerPolicy:            %-8v\n", c.Croissant.WBFT.ProposerPolicy)
			if c.Croissant.WBFT.BlockReward == nil {
				banner += fmt.Sprintf("     - BlockReward:               %-8v\n", 0)
			} else {
				banner += fmt.Sprintf("     - BlockReward:               %-8v\n", ((*big.Int)(c.Croissant.WBFT.BlockReward)).Int64())
			}
			if c.Croissant.WBFT.BlockRewardBeneficiary == nil {
				banner += fmt.Sprintf("     - BlockRewardBeneficiary:    %v\n", nil)
			} else {
				banner += fmt.Sprintf("     - BlockRewardBeneficiary.Denominator: %v\n", c.Croissant.WBFT.BlockRewardBeneficiary.Denominator)
				for i, b := range c.Croissant.WBFT.BlockRewardBeneficiary.Beneficiaries {
					banner += fmt.Sprintf("     - BlockRewardBeneficiary[%v]: %v\n", i, b)
				}
			}
			if c.Croissant.WBFT.MaxRequestTimeoutSeconds == nil {
				banner += fmt.Sprintf("     - MaxRequestTimeoutSeconds:  %-8v\n", 0)
			} else {
				banner += fmt.Sprintf("     - MaxRequestTimeoutSeconds:  %-8v\n", *c.Croissant.WBFT.MaxRequestTimeoutSeconds)
			}
		}
		if c.Croissant.Init != nil {
			if c.Croissant.Init != nil {
				banner += "   - Init\n"
				banner += fmt.Sprintf("     - Validators:         %-8v\n", c.Croissant.Init.Validators)
				banner += fmt.Sprintf("     - BLSPublicKeys:      %-8v\n", c.Croissant.Init.BLSPublicKeys)
			}
		}
		if c.Croissant.GovContracts != nil {
			banner += "   - GovContracts:\n"
			banner += fmt.Sprintf("     - GovConfig:        %-8v\n", c.Croissant.GovContracts.GovConfig)
			banner += fmt.Sprintf("     - GovStaking:       %-8v\n", c.Croissant.GovContracts.GovStaking)
			banner += fmt.Sprintf("     - GovRewardeeImp:   %-8v\n", c.Croissant.GovContracts.GovRewardeeImp)
			banner += fmt.Sprintf("     - GovNCP:           %-8v\n", c.Croissant.GovContracts.GovNCP)
		}
	}
	banner += "\n"

	// Create a list of forks post-merge
	banner += "Post-Merge hard forks (timestamp based):\n"
	if c.ShanghaiTime != nil {
		banner += fmt.Sprintf(" - Shanghai:                    @%-10v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/shanghai.md)\n", *c.ShanghaiTime)
	}
	if c.CancunTime != nil {
		banner += fmt.Sprintf(" - Cancun:                      @%-10v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/cancun.md)\n", *c.CancunTime)
	}
	if c.PragueTime != nil {
		banner += fmt.Sprintf(" - Prague:                      @%-10v\n", *c.PragueTime)
	}
	if c.VerkleTime != nil {
		banner += fmt.Sprintf(" - Verkle:                      @%-10v\n", *c.VerkleTime)
	}
	return banner
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return isBlockForked(c.HomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isBlockForked(c.DAOForkBlock, num)
}

// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isBlockForked(c.EIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isBlockForked(c.EIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isBlockForked(c.EIP158Block, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isBlockForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return isBlockForked(c.ConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
	return isBlockForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return isBlockForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isBlockForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
	return isBlockForked(c.IstanbulBlock, num)
}

// IsBerlin returns whether num is either equal to the Berlin fork block or greater.
func (c *ChainConfig) IsBerlin(num *big.Int) bool {
	return isBlockForked(c.BerlinBlock, num)
}

// IsLondon returns whether num is either equal to the London fork block or greater.
func (c *ChainConfig) IsLondon(num *big.Int) bool {
	return isBlockForked(c.LondonBlock, num)
}

// IsPangyo returns whether num is either equal to the Pangyo fork block or greater.
func (c *ChainConfig) IsPangyo(num *big.Int) bool {
	return isBlockForked(c.PangyoBlock, num)
}

// fee delegation
// IsApplepie returns whether num is either equal to the Applepie fork block or greater.
func (c *ChainConfig) IsApplepie(num *big.Int) bool {
	return isBlockForked(c.ApplepieBlock, num)
}

func (c *ChainConfig) IsBrioche(num *big.Int) bool {
	return isBlockForked(c.BriocheBlock, num)
}

// IsArrowGlacier returns whether num is either equal to the Arrow Glacier (EIP-4345) fork block or greater.
func (c *ChainConfig) IsArrowGlacier(num *big.Int) bool {
	return isBlockForked(c.ArrowGlacierBlock, num)
}

// IsGrayGlacier returns whether num is either equal to the Gray Glacier (EIP-5133) fork block or greater.
func (c *ChainConfig) IsGrayGlacier(num *big.Int) bool {
	return isBlockForked(c.GrayGlacierBlock, num)
}

// IsTerminalPoWBlock returns whether the given block is the last block of PoW stage.
func (c *ChainConfig) IsTerminalPoWBlock(parentTotalDiff *big.Int, totalDiff *big.Int) bool {
	if c.TerminalTotalDifficulty == nil {
		return false
	}
	return parentTotalDiff.Cmp(c.TerminalTotalDifficulty) < 0 && totalDiff.Cmp(c.TerminalTotalDifficulty) >= 0
}

// IsShanghai returns whether time is either equal to the Shanghai fork time or greater.
func (c *ChainConfig) IsShanghai(num *big.Int, time uint64) bool {
	return c.IsLondon(num) && isTimestampForked(c.ShanghaiTime, time)
}

// IsCancun returns whether num is either equal to the Cancun fork time or greater.
func (c *ChainConfig) IsCancun(num *big.Int, time uint64) bool {
	return c.IsLondon(num) && isTimestampForked(c.CancunTime, time)
}

// IsPrague returns whether num is either equal to the Prague fork time or greater.
func (c *ChainConfig) IsPrague(num *big.Int, time uint64) bool {
	return c.IsLondon(num) && isTimestampForked(c.PragueTime, time)
}

// IsVerkle returns whether num is either equal to the Verkle fork time or greater.
func (c *ChainConfig) IsVerkle(num *big.Int, time uint64) bool {
	return c.IsLondon(num) && isTimestampForked(c.VerkleTime, time)
}

func (c *ChainConfig) IsCroissant(num *big.Int) bool {
	return isBlockForked(c.CroissantBlock, num)
}

func (c *ChainConfig) CroissantEnabled() bool {
	return c.CroissantBlock != nil && c.Croissant != nil
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64, time uint64) *ConfigCompatError {
	var (
		bhead = new(big.Int).SetUint64(height)
		btime = time
	)
	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead, btime)
		if err == nil || (lasterr != nil && err.RewindToBlock == lasterr.RewindToBlock && err.RewindToTime == lasterr.RewindToTime) {
			break
		}
		lasterr = err

		if err.RewindToTime > 0 {
			btime = err.RewindToTime
		} else {
			bhead.SetUint64(err.RewindToBlock)
		}
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, geth isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	type fork struct {
		name      string
		block     *big.Int // forks up to - and including the merge - were defined with block numbers
		timestamp *uint64  // forks after the merge are scheduled using timestamps
		optional  bool     // if true, the fork may be nil and next fork is still allowed
	}
	var lastFork fork
	for _, cur := range []fork{
		{name: "homesteadBlock", block: c.HomesteadBlock},
		{name: "daoForkBlock", block: c.DAOForkBlock, optional: true},
		{name: "eip150Block", block: c.EIP150Block},
		{name: "eip155Block", block: c.EIP155Block},
		{name: "eip158Block", block: c.EIP158Block},
		{name: "byzantiumBlock", block: c.ByzantiumBlock},
		{name: "constantinopleBlock", block: c.ConstantinopleBlock},
		{name: "petersburgBlock", block: c.PetersburgBlock},
		{name: "istanbulBlock", block: c.IstanbulBlock},
		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
		{name: "berlinBlock", block: c.BerlinBlock},
		{name: "londonBlock", block: c.LondonBlock},
		{name: "arrowGlacierBlock", block: c.ArrowGlacierBlock, optional: true},
		{name: "grayGlacierBlock", block: c.GrayGlacierBlock, optional: true},
		{name: "pangyoBlock", block: c.PangyoBlock, optional: true},
		{name: "applepieBlock", block: c.ApplepieBlock, optional: true},
		{name: "briocheBlock", block: c.BriocheBlock, optional: true},
		{name: "croissantBlock", block: c.CroissantBlock, optional: true},
		{name: "mergeNetsplitBlock", block: c.MergeNetsplitBlock, optional: true},
		{name: "shanghaiTime", timestamp: c.ShanghaiTime, optional: true},
		{name: "cancunTime", timestamp: c.CancunTime, optional: true},
		{name: "pragueTime", timestamp: c.PragueTime, optional: true},
		{name: "verkleTime", timestamp: c.VerkleTime, optional: true},
	} {
		if lastFork.name != "" {
			switch {
			// Non-optional forks must all be present in the chain config up to the last defined fork
			case lastFork.block == nil && lastFork.timestamp == nil && (cur.block != nil || cur.timestamp != nil):
				if cur.block != nil {
					return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at block %v",
						lastFork.name, cur.name, cur.block)
				} else {
					return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at timestamp %v",
						lastFork.name, cur.name, *cur.timestamp)
				}

			// Fork (whether defined by block or timestamp) must follow the fork definition sequence
			case (lastFork.block != nil && cur.block != nil) || (lastFork.timestamp != nil && cur.timestamp != nil):
				if lastFork.block != nil && lastFork.block.Cmp(cur.block) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at block %v, but %v enabled at block %v",
						lastFork.name, lastFork.block, cur.name, cur.block)
				} else if lastFork.timestamp != nil && *lastFork.timestamp > *cur.timestamp {
					return fmt.Errorf("unsupported fork ordering: %v enabled at timestamp %v, but %v enabled at timestamp %v",
						lastFork.name, *lastFork.timestamp, cur.name, *cur.timestamp)
				}

				// Timestamp based forks can follow block based ones, but not the other way around
				if lastFork.timestamp != nil && cur.block != nil {
					return fmt.Errorf("unsupported fork ordering: %v used timestamp ordering, but %v reverted to block ordering",
						lastFork.name, cur.name)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || (cur.block != nil || cur.timestamp != nil) {
			lastFork = cur
		}
	}
	return nil
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, headNumber *big.Int, headTimestamp uint64) *ConfigCompatError {
	if isForkBlockIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, headNumber) {
		return newBlockCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkBlockIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, headNumber) {
		return newBlockCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(headNumber) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newBlockCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkBlockIncompatible(c.EIP150Block, newcfg.EIP150Block, headNumber) {
		return newBlockCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkBlockIncompatible(c.EIP155Block, newcfg.EIP155Block, headNumber) {
		return newBlockCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkBlockIncompatible(c.EIP158Block, newcfg.EIP158Block, headNumber) {
		return newBlockCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(headNumber) && !configBlockEqual(c.ChainID, newcfg.ChainID) {
		return newBlockCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkBlockIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, headNumber) {
		return newBlockCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, headNumber) {
		return newBlockCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if isForkBlockIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, headNumber) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
		if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, headNumber) {
			return newBlockCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if isForkBlockIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, headNumber) {
		return newBlockCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if isForkBlockIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, headNumber) {
		return newBlockCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}
	if isForkBlockIncompatible(c.BerlinBlock, newcfg.BerlinBlock, headNumber) {
		return newBlockCompatError("Berlin fork block", c.BerlinBlock, newcfg.BerlinBlock)
	}
	if isForkBlockIncompatible(c.LondonBlock, newcfg.LondonBlock, headNumber) {
		return newBlockCompatError("London fork block", c.LondonBlock, newcfg.LondonBlock)
	}
	if isForkBlockIncompatible(c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock, headNumber) {
		return newBlockCompatError("Arrow Glacier fork block", c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock)
	}
	if isForkBlockIncompatible(c.GrayGlacierBlock, newcfg.GrayGlacierBlock, headNumber) {
		return newBlockCompatError("Gray Glacier fork block", c.GrayGlacierBlock, newcfg.GrayGlacierBlock)
	}
	if isForkBlockIncompatible(c.PangyoBlock, newcfg.PangyoBlock, headNumber) {
		return newBlockCompatError("Pangyo fork block", c.PangyoBlock, newcfg.PangyoBlock)
	}
	if isForkBlockIncompatible(c.ApplepieBlock, newcfg.ApplepieBlock, headNumber) {
		return newBlockCompatError("Applepie fork block", c.ApplepieBlock, newcfg.ApplepieBlock)
	}
	if isForkBlockIncompatible(c.BriocheBlock, newcfg.BriocheBlock, headNumber) {
		return newBlockCompatError("Brioche fork block", c.BriocheBlock, newcfg.BriocheBlock)
	}
	if isForkBlockIncompatible(c.CroissantBlock, newcfg.CroissantBlock, headNumber) {
		return newBlockCompatError("Croissant fork block", c.CroissantBlock, newcfg.CroissantBlock)
	}
	if isForkBlockIncompatible(c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock, headNumber) {
		return newBlockCompatError("Merge netsplit fork block", c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock)
	}
	if isForkTimestampIncompatible(c.ShanghaiTime, newcfg.ShanghaiTime, headTimestamp) {
		return newTimestampCompatError("Shanghai fork timestamp", c.ShanghaiTime, newcfg.ShanghaiTime)
	}
	if isForkTimestampIncompatible(c.CancunTime, newcfg.CancunTime, headTimestamp) {
		return newTimestampCompatError("Cancun fork timestamp", c.CancunTime, newcfg.CancunTime)
	}
	if isForkTimestampIncompatible(c.PragueTime, newcfg.PragueTime, headTimestamp) {
		return newTimestampCompatError("Prague fork timestamp", c.PragueTime, newcfg.PragueTime)
	}
	if isForkTimestampIncompatible(c.VerkleTime, newcfg.VerkleTime, headTimestamp) {
		return newTimestampCompatError("Verkle fork timestamp", c.VerkleTime, newcfg.VerkleTime)
	}
	return nil
}

// BaseFeeChangeDenominator bounds the amount the base fee can change between blocks.
func (c *ChainConfig) BaseFeeChangeDenominator() uint64 {
	return DefaultBaseFeeChangeDenominator
}

// ElasticityMultiplier bounds the maximum gas limit an EIP-1559 block may have.
func (c *ChainConfig) ElasticityMultiplier() uint64 {
	return DefaultElasticityMultiplier
}

// LatestFork returns the latest time-based fork that would be active for the given time.
func (c *ChainConfig) LatestFork(time uint64) forks.Fork {
	// Assume last non-time-based fork has passed.
	london := c.LondonBlock

	switch {
	case c.IsPrague(london, time):
		return forks.Prague
	case c.IsCancun(london, time):
		return forks.Cancun
	case c.IsShanghai(london, time):
		return forks.Shanghai
	default:
		return forks.Paris
	}
}

// isForkBlockIncompatible returns true if a fork scheduled at block s1 cannot be
// rescheduled to block s2 because head is already past the fork.
func isForkBlockIncompatible(s1, s2, head *big.Int) bool {
	return (isBlockForked(s1, head) || isBlockForked(s2, head)) && !configBlockEqual(s1, s2)
}

// isBlockForked returns whether a fork scheduled at block s is active at the
// given head block. Whilst this method is the same as isTimestampForked, they
// are explicitly separate for clearer reading.
func isBlockForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configBlockEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// isForkTimestampIncompatible returns true if a fork scheduled at timestamp s1
// cannot be rescheduled to timestamp s2 because head is already past the fork.
func isForkTimestampIncompatible(s1, s2 *uint64, head uint64) bool {
	return (isTimestampForked(s1, head) || isTimestampForked(s2, head)) && !configTimestampEqual(s1, s2)
}

// isTimestampForked returns whether a fork scheduled at timestamp s is active
// at the given head timestamp. Whilst this method is the same as isBlockForked,
// they are explicitly separate for clearer reading.
func isTimestampForked(s *uint64, head uint64) bool {
	if s == nil {
		return false
	}
	return *s <= head
}

func configTimestampEqual(x, y *uint64) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return *x == *y
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string

	// block numbers of the stored and new configurations if block based forking
	StoredBlock, NewBlock *big.Int

	// timestamps of the stored and new configurations if time based forking
	StoredTime, NewTime *uint64

	// the block number to which the local chain must be rewound to correct the error
	RewindToBlock uint64

	// the timestamp to which the local chain must be rewound to correct the error
	RewindToTime uint64
}

func newBlockCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{
		What:          what,
		StoredBlock:   storedblock,
		NewBlock:      newblock,
		RewindToBlock: 0,
	}
	if rew != nil && rew.Sign() > 0 {
		err.RewindToBlock = rew.Uint64() - 1
	}
	return err
}

func newTimestampCompatError(what string, storedtime, newtime *uint64) *ConfigCompatError {
	var rew *uint64
	switch {
	case storedtime == nil:
		rew = newtime
	case newtime == nil || *storedtime < *newtime:
		rew = storedtime
	default:
		rew = newtime
	}
	err := &ConfigCompatError{
		What:         what,
		StoredTime:   storedtime,
		NewTime:      newtime,
		RewindToTime: 0,
	}
	if rew != nil {
		err.RewindToTime = *rew - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	if err.StoredBlock != nil {
		return fmt.Sprintf("mismatching %s in database (have block %d, want block %d, rewindto block %d)", err.What, err.StoredBlock, err.NewBlock, err.RewindToBlock)
	}
	return fmt.Sprintf("mismatching %s in database (have timestamp %d, want timestamp %d, rewindto timestamp %d)", err.What, err.StoredTime, err.NewTime, err.RewindToTime)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
	IsBerlin, IsLondon                                      bool
	IsPangyo, IsApplepie, IsBrioche, IsCroissant            bool
	IsMerge, IsShanghai, IsCancun, IsPrague                 bool
	IsVerkle                                                bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int, isMerge bool, timestamp uint64) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	// disallow setting Merge out of order
	isMerge = isMerge && c.IsLondon(num)
	return Rules{
		ChainID:          new(big.Int).Set(chainID),
		IsHomestead:      c.IsHomestead(num),
		IsEIP150:         c.IsEIP150(num),
		IsEIP155:         c.IsEIP155(num),
		IsEIP158:         c.IsEIP158(num),
		IsByzantium:      c.IsByzantium(num),
		IsConstantinople: c.IsConstantinople(num),
		IsPetersburg:     c.IsPetersburg(num),
		IsIstanbul:       c.IsIstanbul(num),
		IsBerlin:         c.IsBerlin(num),
		IsLondon:         c.IsLondon(num),
		IsPangyo:         c.IsPangyo(num),
		IsApplepie:       c.IsApplepie(num),
		IsBrioche:        c.IsBrioche(num),
		IsCroissant:      c.IsCroissant(num),
		IsMerge:          isMerge,
		IsShanghai:       isMerge && c.IsShanghai(num, timestamp),
		IsCancun:         isMerge && c.IsCancun(num, timestamp),
		IsPrague:         isMerge && c.IsPrague(num, timestamp),
		IsVerkle:         isMerge && c.IsVerkle(num, timestamp),
	}
}
