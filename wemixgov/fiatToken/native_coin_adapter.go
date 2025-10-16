package systemcontracts

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
)

const (
	SLOT_COIN_ADAPTER_MASTER_MINTER  = "0x0" // address masterMinter
	SLOT_COIN_ADAPTER_MINTERS        = "0x1" // mapping(address => bool) _minters
	SLOT_COIN_ADAPTER_MINTER_ALLOWED = "0x2" // mapping(address => uint256) _minterAllowed
	SLOT_COIN_ADAPTER_COIN_MANAGER   = "0x8" // address _coinManager
	SLOT_COIN_ADAPTER_NAME           = "0x9" // string name
	SLOT_COIN_ADAPTER_SYMBOL         = "0xa" // string symbol
	SLOT_COIN_ADAPTER_DECIMALS       = "0xb" // uint8 decimals
	SLOT_COIN_ADAPTER_CURRENCY       = "0xc" // string currency
	SLOT_COIN_ADAPTER_TOTAL_SUPPLY   = "0xe" // uint256 _totalSupply

	COIN_ADAPTER_PARAM_MASTER_MINTER  = "masterMinter"
	COIN_ADAPTER_PARAM_MINTERS        = "minters"
	COIN_ADAPTER_PARAM_MINTER_ALLOWED = "minterAllowed"
	COIN_ADAPTER_PARAM_NAME           = "name"
	COIN_ADAPTER_PARAM_SYMBOL         = "symbol"
	COIN_ADAPTER_PARAM_DECIMALS       = "decimals"
	COIN_ADAPTER_PARAM_CURRENCY       = "currency"
)

func InitializeCoinAdatper(coinAdapterAddress common.Address, param map[string]string, alloc *types.GenesisAlloc) ([]params.StateParam, error) {
	sp := make([]params.StateParam, 0)

	// SLOT_COIN_ADAPTER_MASTER_MINTER
	masterMinter, ok := param[COIN_ADAPTER_PARAM_MASTER_MINTER]
	if !ok || len(masterMinter) == 0 {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: missing parameter: %s", COIN_ADAPTER_PARAM_MASTER_MINTER)
	}
	sp = append(sp, params.StateParam{
		Address: coinAdapterAddress,
		Key:     common.HexToHash(SLOT_COIN_ADAPTER_MASTER_MINTER),
		Value:   common.BytesToHash(common.HexToAddress(masterMinter).Bytes()),
	})

	// SLOT_COIN_ADAPTER_MINTERS, SLOT_COIN_ADAPTER_MINTER_ALLOWED
	if mintersStr, ok := param[COIN_ADAPTER_PARAM_MINTERS]; ok && len(mintersStr) > 0 {
		minters := strings.Split(mintersStr, ",")
		minterAllowedAmounts := make([]*big.Int, len(minters))
		if minterAllowedStr, ok := param[COIN_ADAPTER_PARAM_MINTER_ALLOWED]; ok && len(minterAllowedStr) > 0 {
			minterAllowed := strings.Split(minterAllowedStr, ",")
			if len(minters) != len(minterAllowed) {
				return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: the number of minters and minterAllowedAmounts must be the same")
			}
			for i, allowed := range minterAllowed {
				allowedAmount, ok := new(big.Int).SetString(allowed, 10)
				if !ok {
					return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: invalid minterAllowed, must be decimal")
				}
				minterAllowedAmounts[i] = allowedAmount
			}
		}
		mintersSlot := common.HexToHash(SLOT_COIN_ADAPTER_MINTERS)
		minterAllowedSlot := common.HexToHash(SLOT_COIN_ADAPTER_MINTER_ALLOWED)
		for i, minter := range minters {
			minterAddress := common.HexToAddress(minter)
			sp = append(sp, params.StateParam{
				Address: coinAdapterAddress,
				Key:     govwbft.CalculateMappingSlot(mintersSlot, minterAddress),
				Value:   common.BytesToHash([]byte{1}), // == true
			})

			if minterAllowedAmounts[i] != nil {
				sp = append(sp, params.StateParam{
					Address: coinAdapterAddress,
					Key:     govwbft.CalculateMappingSlot(minterAllowedSlot, minterAddress),
					Value:   common.BigToHash(minterAllowedAmounts[i]),
				})
			}
		}
	}

	// SLOT_COIN_ADAPTER_COIN_MANAGER
	sp = append(sp, params.StateParam{
		Address: coinAdapterAddress,
		Key:     common.HexToHash(SLOT_COIN_ADAPTER_COIN_MANAGER),
		Value:   common.BytesToHash(params.NativeCoinManagerAddress.Bytes()),
	})

	// SLOT_COIN_ADAPTER_NAME
	coinName, ok := param[COIN_ADAPTER_PARAM_NAME]
	if !ok || len(coinName) == 0 {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: missing parameter: %s", COIN_ADAPTER_PARAM_NAME)
	}
	for slot, value := range govwbft.EncodeBytesToSlots(common.HexToHash(SLOT_COIN_ADAPTER_NAME), []byte(coinName)) {
		sp = append(sp, params.StateParam{
			Address: coinAdapterAddress,
			Key:     slot,
			Value:   value,
		})
	}

	// SLOT_COIN_ADAPTER_SYMBOL
	symbol, ok := param[COIN_ADAPTER_PARAM_SYMBOL]
	if !ok || len(symbol) == 0 {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: missing parameter: %s", COIN_ADAPTER_PARAM_SYMBOL)
	}
	for slot, value := range govwbft.EncodeBytesToSlots(common.HexToHash(SLOT_COIN_ADAPTER_NAME), []byte(symbol)) {
		sp = append(sp, params.StateParam{
			Address: coinAdapterAddress,
			Key:     slot,
			Value:   value,
		})
	}

	// SLOT_COIN_ADAPTER_DECIMALS
	decimalsStr, ok := param[COIN_ADAPTER_PARAM_DECIMALS]
	if !ok || len(decimalsStr) == 0 {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: missing parameter: %s", COIN_ADAPTER_PARAM_DECIMALS)
	}
	decimals, err := strconv.ParseUint(decimalsStr, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params.decimals`: %w", err)
	}
	sp = append(sp, params.StateParam{
		Address: coinAdapterAddress,
		Key:     common.HexToHash(SLOT_COIN_ADAPTER_DECIMALS),
		Value:   common.BytesToHash([]byte{uint8(decimals)}),
	})

	// SLOT_COIN_ADAPTER_CURRENCY
	currency, ok := param[COIN_ADAPTER_PARAM_CURRENCY]
	if !ok || len(currency) == 0 {
		return nil, fmt.Errorf("`systemContracts.nativeCoinAdapter.params`: missing parameter: %s", COIN_ADAPTER_PARAM_CURRENCY)
	}
	for slot, value := range govwbft.EncodeBytesToSlots(common.HexToHash(SLOT_COIN_ADAPTER_NAME), []byte(currency)) {
		sp = append(sp, params.StateParam{
			Address: coinAdapterAddress,
			Key:     slot,
			Value:   value,
		})
	}

	// SLOT_COIN_ADAPTER_TOTAL_SUPPLY
	if alloc != nil {
		totalSupply := new(big.Int)
		for _, account := range *alloc {
			totalSupply.Add(totalSupply, account.Balance)
		}
		sp = append(sp, params.StateParam{
			Address: coinAdapterAddress,
			Key:     common.HexToHash(SLOT_COIN_ADAPTER_TOTAL_SUPPLY),
			Value:   common.BigToHash(totalSupply),
		})
	}

	return sp, nil
}
