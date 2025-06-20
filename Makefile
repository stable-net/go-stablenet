# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: geth android ios evm all test clean

GOBIN = ./build/bin
GO ?= latest
GORUN = go run

SUCCESS_TESTS = github.com/ethereum/go-ethereum/accounts github.com/ethereum/go-ethereum/accounts/abi github.com/ethereum/go-ethereum/cmd/abigen github.com/ethereum/go-ethereum/cmd/clef github.com/ethereum/go-ethereum/cmd/evm/... github.com/ethereum/go-ethereum/cmd/devp2p/...  github.com/ethereum/go-ethereum/cmd/ethkey github.com/ethereum/go-ethereum/cmd/rlpdump github.com/ethereum/go-ethereum/common/... github.com/ethereum/go-ethereum/core/... github.com/ethereum/go-ethereum/ethclient/... github.com/ethereum/go-ethereum/consensus/misc/...  github.com/ethereum/go-ethereum/consensus/wbft/...  github.com/ethereum/go-ethereum/crypto/...  github.com/ethereum/go-ethereum/ethdb/...  github.com/ethereum/go-ethereum/ethstats github.com/ethereum/go-ethereum/event github.com/ethereum/go-ethereum/log github.com/ethereum/go-ethereum/metrics/...  github.com/ethereum/go-ethereum/miner github.com/ethereum/go-ethereum/node github.com/ethereum/go-ethereum/p2p/...  github.com/ethereum/go-ethereum/params github.com/ethereum/go-ethereum/rlp/...  github.com/ethereum/go-ethereum/signer/...  github.com/ethereum/go-ethereum/tests/fuzzers/...  github.com/ethereum/go-ethereum/trie/...  github.com/ethereum/go-ethereum/triedb/...  github.com/ethereum/go-ethereum/graphql/...  github.com/ethereum/go-ethereum/internal/...  github.com/ethereum/go-ethereum/rpc github.com/ethereum/go-ethereum/cmd/geth github.com/ethereum/go-ethereum/cmd/utils github.com/ethereum/go-ethereum/console github.com/ethereum/go-ethereum/eth/... github.com/ethereum/go-ethereum/wemixgov/...

# TODO: move this cases to SUCCESS_TESTS one by one after making it to be success
FAILURE_TESTS = github.com/ethereum/go-ethereum/tests

#? geth: Build geth
geth:
	$(GORUN) build/ci.go install ./cmd/geth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/geth\" to launch geth."

#? geth: Build genesis_generator
genesis_generator:
	$(GORUN) build/ci.go install ./cmd/genesis_generator
	@echo "Done building."
	@echo "Run \"$(GOBIN)/genesis_generator\" to launch genesis_generator."

#? all: Build all packages and executables
all:
	$(GORUN) build/ci.go install

#? test: Run the tests
test: all
#	$(GORUN) build/ci.go test
	$(GORUN) build/ci.go test $(SUCCESS_TESTS)

test-short: all
	$(GORUN) build/ci.go test -short $(SUCCESS_TESTS)

#? lint: Run certain pre-selected linters
lint: ## Run linters.
	$(GORUN) build/ci.go lint

#? clean: Clean go cache, built executables, and the auto generated folder
clean:
	go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

#? devtools: Install recommended developer tools
devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install github.com/golang/protobuf/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

#? help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in go-ethereum:"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
.PHONY: help
