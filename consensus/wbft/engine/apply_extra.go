// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/wbft/engine/apply_extra.go (2024.07.25).
// Modified and improved for the wemix development.

package wbftengine

import "github.com/ethereum/go-ethereum/core/types"

type ApplyWBFTExtra func(*types.WBFTExtra) error

func Combine(applies ...ApplyWBFTExtra) ApplyWBFTExtra {
	return func(extra *types.WBFTExtra) error {
		for _, apply := range applies {
			err := apply(extra)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func ApplyHeaderWBFTExtra(header *types.Header, applies ...ApplyWBFTExtra) (*types.WBFTExtra, error) {
	extra, err := getExtra(header)
	if err != nil {
		return nil, err
	}

	err = Combine(applies...)(extra)
	if err != nil {
		return nil, err
	}

	return extra, setExtra(header, extra)
}
