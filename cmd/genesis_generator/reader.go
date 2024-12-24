package main

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/peterh/liner"
)

// prompts the user for input with the given prompt string.  Returns when a value is entered.
// Causes the genesisGenerator to exit if ctrl-d is pressed
func promptInput(p string) string {
	for {
		text, err := prompt.Stdin.PromptInput(p)
		if err != nil {
			if err != liner.ErrPromptAborted {
				log.Crit("Failed to read user input", "err", err)
			}
		} else {
			return text
		}
	}
}

// read reads a single line from stdin, trimming if from spaces.
func read() string {
	text := promptInput("> ")
	return strings.TrimSpace(text)
}

// readDefaultString reads a single line from stdin, trimming if from spaces. If
// an empty line is entered, the default value is returned.
func readDefaultString(def string) string {
	text := promptInput("> ")
	if text = strings.TrimSpace(text); text != "" {
		return text
	}
	return def
}

// readDefaultYesNo reads a single line from stdin, trimming if from spaces and
// interpreting it as a 'yes' or a 'no'. If an empty line is entered, the default
// value is returned.
func readDefaultYesNo(def bool) bool {
	for {
		text := promptInput("> ")
		if text = strings.ToLower(strings.TrimSpace(text)); text == "" {
			return def
		}
		if text == "y" || text == "yes" {
			return true
		}
		if text == "n" || text == "no" {
			return false
		}
		log.Error("Invalid input, expected 'y', 'yes', 'n', 'no' or empty")
	}
}

// readDefaultInt reads a single line from stdin, trimming if from spaces, enforcing
// it to parse into an integer. If an empty line is entered, the default value is
// returned.
func readDefaultInt(def int) int {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			log.Error("Invalid input, expected integer", "err", err)
			continue
		}
		return val
	}
}

// readDefaultBigInt reads a single line from stdin, trimming if from spaces,
// enforcing it to parse into a big integer. If an empty line is entered, the
// default value is returned.
func readDefaultBigInt(def *big.Int) *big.Int {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		val, ok := new(big.Int).SetString(text, 0)
		if !ok {
			log.Error("Invalid input, expected big integer")
			continue
		}
		return val
	}
}

// readAddress reads a single line from stdin, trimming if from spaces and converts
// it to an Ethereum address.
func readAddress() *common.Address {
	for {
		text := promptInput("> 0x")
		if text = strings.TrimSpace(text); text == "" {
			return nil
		}
		// Make sure it looks ok and return it if so
		if len(text) != 40 {
			log.Error("Invalid address length, please retry")
			continue
		}
		bigaddr, _ := new(big.Int).SetString(text, 16)
		address := common.BigToAddress(bigaddr)
		return &address
	}
}
