## WBFT Protocol Specification (WEMIX 4.0)

WBFT(WEMIX Byzantine Fault Tolerant) is a consensus algorithm that emphasizes decentralization, adapting Istanbul BFT(https://github.com/ethereum/EIPs/issues/650) and QBFT(https://github.com/Consensys/qbft-formal-spec-and-verification) for use in public blockchains. The following improvements have been implemented:

- Adoption of DPoS(Delegated Proof of Stake): Allows anyone to participate as a validator through staking.
- Validator selection: Chosen via VRF based on staking amount and validation diligence.
- Reward system and diligence metrics.
- Concept of epoch: Defines a unit where the validator set changes.
- Inclusion of consensus proof in the agreement process.
- Improved miner worker operations.
- Enhanced block header verification.

### Terminology
- `Staker`: A participant who stakes an amount exceeding the minimum staking threshold. Candidates for validators during an epoch.
- `Validator`: Block validation participant(Same as Validator in IBFT). Must be a staker to become a validator.
- `Proposer`: A block validation participant that is chosen to propose block in a consensus round(Same as Proposer in IBFT).
- `Epoch`: The duration during which a fixed validator set remains active, represented in the number of blocks.
- `Epoch block`: The last block of an epoch. It records cumulative diligence for all stakers and defines the validator set for the next epoch.
- `Sequence`: Sequence number of a proposal. A sequence number should be greater than all previous sequence numbers. Currently each proposed block height is its associated sequence number(Same as Sequence in IBFT).
- `Round`: Consensus round. A round starts with the proposer creating a block proposal and ends with a block commitment or round change(Same as Round in IBFT).
- `Diligence`: Measures validation activity based on the total number of previous seals (previous prepare and commit seals) recorded in the epoch block during an epoch
- `Round state`: Consensus messages of a specific sequence and round, including pre-prepare message, prepare message, and commit message(Same as Round state in IBFT).
- `Backlog`: The storage to keep future consensus messages(Same as Backlog in IBFT).
- `Consensus proof`: The commitment signatures of a block that can prove the block has gone through the consensus process in IBFT. Unlike IBFT, which only had commit seals, WBFT includes prepare seals and seals for previous block verification. Previous seals are part of the consensus process.

Removed from IBFT
- `Snapshot`
- `Validator voting`

### Adoption of DPoS(Delegated Proof of Stake)
In IBFT or QBFT, new validators could be added or removed via validator set voting, which is suitable for PoA chains but not for public blockchains. WBFT allows anyone to participate as a validator through staking. By staking at least the minimum amount, one can become a staker. Stakers have the following attributes:

- `Staker node address`: The nodekey address used in consensus if selected as a validator.
- `Staker BLS public key`: The BLS public key used in verifing WBFT seal by others.
- `Staker operator address`: The wallet address for performing stake/unstake operations.
- `Staker reward address`: The address for receiving block rewards.
- `Staking amount`: The amount staked, which must exceed the minimum staking amount.
- `Delegated amount`: The amount received from delegations.

Rules related to staking:
- Staking/Unstaking
  - First staking must exceed the minimum staking threshold and then it becomes a staker.
  - No restrictions on amounts for additional staking.
  - Staking amount is staker's own staking amount + delegated amount.
  - Staker's own staking amount should be greater than or equal to the minimum staking threshold.
  - When unstaking, if the remaining staking amount falls below the minimum threshold and is not zero, the unstaking will fail.  
  - If the own staking amount reaches under minimum staking threshold by unstaking or slashing, it is removed from staker set.
  - If a staker is removed from staker set, its remaining own staking amounts are unstaked and its delegated amounts are refunded to each delegator.
  - Unstaked funds are released to the _staker operator address_ after the unbonding period (1 weeks).
- Slashing
  - Stakers considered as malicious may be slashed, paying their some own staked amounts to WEMIX ecosystem.
  - Funds in the unbonding state can still be slashed.
- Delegation
  - Anyone can delegate funds to the staker with no amount limit.
  - Delegated amounts are credited to the recipient staker's staking amount.
  - Undelegated funds are released to the _delegator address_ after the unbonding period (72 hours).

### Validator Selection
In WBFT, the proposer of the last block in an epoch (referred to as the epoch block) selects the validator set for the next epoch and records it in the block. Validator selection rules:
- Stabilization stage
  - The stabilization stage is the period from the first epoch until just before the first epoch when the number of stakers reaches the minimum required.
  - If current epoch is in a stabilization stage and the number of stakers is below the minimum stakers in last block of an epoch, next epoch will be in a stabilization stage.
  - An epoch in a stabilization stage has the validator set which is same to the previous epoch.
  - The very first validator set is defined in genesis.json, and validators in the genesis block initially have a staking amount of zero.
- After stabilization stage, validator selection follows below rules
  - `minimum stakers <= number of stakers <= target validators`: every stakers become validators.
  - `number of stakers > target validators`: validators are selected using VRF, considering staking amount and diligence.
  - `number of stakers < minimum stakers`: all remaining stakers become validators, which should not occur in a public network after stabilization stage for the sake of network security.

Validators are selected to act as proposers in a round-robin manner.

### Reward System and Diligence Metrics
WBFT rewards consist of two types:
- `Transaction Gas Fees`: Granted to the block proposer.
- `Block Minting Rewards`: Distributed to validators proportionally to their staking power.
  - The minting amount is defined in the WBFT configuration.
  - After the "brioche" hard fork, block minting rewards follow a halving cycle.

Diligence influences validator selection directly, encouraging validators to perform their roles sincerely.
Block minting rewards play an important role in BFT. In public network BFTs, if validators do not faithfully 
perform block validation, round changes can occur, leading to longer block generation times. Therefore, it is necessary 
to provide rewards for the validation role as well. However, direct coin rewards based on validation actions are avoided 
due to potential misuse (e.g., block proposers deliberately omitting a specific validator's consensus proof). 
Instead, validation rewards are indirectly tied to validator selection. Validators who diligently perform their 
validation duties have a better chance of being selected and can earn greater rewards in this structure.

Diligence is calculated at the end of each epoch:
- Let `e` be the number of blocks in the epoch.
- Let `v` be the number of validators.
- Let `w` be the number of times a validator is selected as a proposer during an epoch(including times by round change).
- Let `p` be the total seals (prepare and commit) included in the proposed blocks by a validator.
  - p can be equal to `2*v*w` at most.
- Let `s` be the seals submitted by the validator during the epoch.
  - s can be equal to `2*e` at most.
- Diligence `d = p / (2*v*w) + s / (2*e)`.
  - The maximum value of `d` is 2.
  - The minimum value of `d` is 0 (no proposals or seals).
- ex) if `e=200, v=20, w=10` then `d = p / 400 + s / 400`.

When a proposer writes the diligence in a epoch block, it uses cumulative diligence for each staker.

Cumulative diligence `D(n) = D(n-1) * 0.9 + d(n) * 0.1`.
- `D(n)` is the cumulative diligence until the n-th epoch.
- `d(n)` is the diligence of the n-th epoch.
- `D(n-1)` is the cumulative diligence until the (n-1)th epoch. If it becomes a staker at first, its `D(n-1) = 1.9` (default. 95% of the maximum value of diligence).
  - If the default value is too low, the probability of being selected as a validator when first becoming a staker will be low. Conversely, if it is too high, it may be advantageous to become a new staker again after even minor mistakes. Therefore, an appropriate value is necessary.

### Concept of Epoch
The WBFT configuration allows defining the size of an epoch. An epoch represents the period (in terms of block count) during which a predetermined validator set remains active. The last block of an epoch is referred to as an epoch block. The genesis block is considered an epoch block; hence, the first epoch starts from block 1. When the proposer suggests a block that is an epoch block, the following steps are performed:

- Reflect the diligence shown by the current staker set during this epoch in their cumulative diligence and record it in the extra field of the block header (if an staker was not part of the validator set during this epoch, its cumulative diligence is not updated).
- Select a new validator set and record it in the extra field:
  - Retrieve stakers from the GovStaking contract.
  - Select validators based on their staking power and diligence.

### Inclusion of Consensus Proof in the Consensus Process
Traditional IBFT and QBFT protocols only included and stored the minimum necessary consensus proof (commit seal) collected locally by each validator for the finalized block. Since these commit seals could vary by validator, they were not part of the block hash. However, to select validators based on consensus proof, it must be included in the consensus process. The consensus proofs included in WBFT blocks are as follows:

- `Previous Prepare Seal`: The prepare seal for the previous block. It is included in the block hash and consensus.
- `Previous Commit Seal`: The commit seal for the previous block. It is included in the block hash and consensus.
- `Prepare Seal`: The prepare seal for the current block. It is not included in the block hash or consensus.
- `Commit Seal`: The commit seal for the current block. It is not included in the block hash or consensus.

When a validator becomes the proposer, they create the current block by combining the prepare and commit seals stored in the previous block with any additional prepare and commit messages received. These are recorded in the current block as the `Previous Prepare Seal` and `Previous Commit Seal`.

The prepare and commit seals stored in the previous block are messages collected from peers just before the block is finalized. However, before the next block is created, the block period allows additional prepare and commit messages to be received. WBFT improves upon IBFT/QBFT by aiming to collect as many of these messages as possible, rather than stopping at just two-third.

This approach incentivizes proposers to include as many previous seals as possible in the block, which enhances their diligence score. It also encourages sealers whose seals are included to improve their diligence scores.

Since each seal is 65 bytes in size and seals are recorded for all validators, block headers could become excessively large. To address this issue, BLS signatures are utilized to reduce the size of the seals.

### Improved Miner Worker Operations
The existing miner worker is designed to be fit to the ethash algorithm. When a new block is received, the worker starts new work and enters a loop to find the nonce. After a certain time (recommit time), it starts new work to include newly in-came transactions in the block. However, this behavior is not suitable for the IBFT algorithm. While it is correct to start work when a new block is received, starting new work at each recommit time is inefficient. Instead, it is more appropriate to start new work when a new round begins. In IBFT, there are cases where consensus fails in a round, and in such cases, a new proposer must start new work. The timing for this should be determined by a round change, not by recommit time. Therefore, in WBFT, the worker is modified to start work at the beginning of each round. The following protocol is applied:
- When the worker receives a new block, it notifies the WBFT engine to perform the final commit (Same to IBFT).
- The WBFT engine notifies the worker to start new work whenever a new sequence begins with a new round (round 0) or when a round change occurs.
- The WBFT engine waits for the block period before notifying the worker(In the existing IBFT, the block period was waited for when sealing the block).
- When new work starts, the worker begins the process of preparing the block.

### Enhanced Block Header Verification
In the existing IBFT, the block verification process involved validating the consensus proof of the validators. This process required knowledge of the validator set, which is available for blocks being added to the canonical chain. However, during the snap sync process, the target block for verification could belong to a distant future in the canonical chain, making it impossible to determine the validator set. As a result, errors would escalate, and the snap sync process was implemented to ignore such errors. This logic had the potential to overlook errors that should have been detected, necessitating improvement.

In WBFT, verification requiring the validator set is performed only when the block is attached to the canonical chain. The validation step has been removed from the standard block verification logic to address this issue.

### Modified Sturctures
The existing QBFT Config was revised by removing unnecessary fields and adding required ones, resulting in the following structure.
```
"qbft": {
      "epochLength": 200,
      "blockPeriodSeconds": 1,
      "requestTimeoutSeconds": 2,
      "maxRequestTimeoutSeconds": 2,
      "targetValidators": 20,
      "minStakers": 13,
      "blockReward": 1000000000000000000,
      "blockRewardBeneficiary":  {
         "denominator": 10000,
         "beneficiaries": [
           {
             "name": "Maintenance",
             "addr": "0xadd0000000000000000000000000000000000000",
             "numerator": 2500,
           },
           {
             "name": "EcoSystem",
             "addr": "0xadd0000000000000000000000000000000000001",
             "numerator": 2500
           }]},
      "validators": ["0xadd0000000000000000000000000000000000002", ...]
    }
```

`blockRewardBeneficiary` defines the address that will receive the block minting rewards consistently.
`targetValidators` should be less than or equals to `epochLength`.
`minStakers` should be less than or equals to `targetValidators`.

```
type Staker struct {
    Addr      common.Address
    BLSPubKey []byte
    Diligence uint64
}
type EpochInfo struct {
    Stakers    []*Staker
    Validators []uint32
}
type WBFTExtra struct {
	VanityData                  []byte
	Round                       uint32
	PreparedSeal                [][]byte
	CommittedSeal               [][]byte
	AggregatedPrevPreparedSeal  []byte
	AggregatedPrevCommittedSeal []byte
	Epoch                       EpochInfo
}
```

### WEMIX 3.5

WEMIX 3.5 defines a hard fork from the existing WEMIX SPoA consensus to the WBFT consensus and adopts an intermediate consensus method with some features disabled for a safe transition to WBFT.

#### MontBlanc hard fork

WBFT is designed to record the first epoch information in the genesis block. However, since the WEMIX chain needs to transition from an already existing chain to a WBFT chain, exception rules for this are defined in the mont blanc hard fork. The following protocol applies to this hard fork:
- WPoA validator nodes stop block creation when it is time to create the mont blanc block (i.e., they create blocks up to just before the mont blanc hard fork).
- WBFT validator nodes receive block propagation normally up to the block before mont blanc.
- WBFT validator nodes recognize the mont blanc hard fork and can obtain the validator set from the WBFT config when it is their turn to create this block.
- WBFT validator nodes can create blocks and proceed with consensus once they can obtain the validator set.
- The mont blanc block is a special block. Validators that receive this block perform additional tasks to deploy and initialize the NCP contract and GovStaking contract.
- The mont blanc block is the first epoch block of WBFT. Therefore, the first validator set is recorded.
- The first epoch starts from the block after the mont blanc block, during which stakers start staking from zero.
- If the number of stakers is equal to or greater than the minimum stakers during the first epoch, these stakers become validators from the next epoch.
- If the number of stakers is less than the minimum stakers during the first epoch, the initial validator set is maintained.

#### NCP

Although WBFT is designed and implemented to be used as a public chain, the Wemix chain continues selecting validators from NCPs for a safer transition to a public chain as an intermediate step. NCPs are selected from the existing Wemix 3.0 NCPs and are defined by contract. They have the obligation to run (mining) nodes for maintaining WEMIX 3.5 safely. The addition/removal of NCPs is decided by voting among NCPs, so it can proceed without a separate hard fork. Anyone can know the NCP list by querying the NCP contract. The NCP system is a temporary feature used only in WEMIX 3.5 and will not be used in WEMIX 4.0.

During the WEMIX 3.5 phase, the validator set selection rules are as follows:
- Retrieve stakers from the GovStaking contract.
- Among these stakers, nodes that are NCPs are selected for the validator set.
- Stakers who are not NCPs are not included in the validator set.

The process of obtaining the validator set at any block height is as follows (replacing the snapshot function in the existing IBFT):
- If the current block is an epoch block, obtain the validator set from the current block.
- If the current block is not an epoch block, traverse blocks backward from the height - 1 block to find the most recent epoch block and obtain the validator set.
- It should meet the mont blanc hard fork block or the genesis block during traversing, otherwise it is a failure.

#### Not implemented in WEMIX 3.5
- Slashing: The NCP system is used to ensure the safety of the chain during the transition period, and the slashing mechanism is not necessary.
- VRF for validator selection: All NCPs are validators, so the VRF is not used for validator selection.

## Building the source

For prerequisites and detailed build instructions please read the [Installation Instructions](https://geth.ethereum.org/docs/getting-started/installing-geth).

Building `geth` requires both a Go (version 1.19 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make geth
```

or, to build the full suite of utilities:

```shell
make all
```

## Executables

The go-ethereum project comes with several wrappers/executables found in the `cmd`
directory.

|  Command   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| :--------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`geth`** | Our main Ethereum CLI client. It is the entry point into the Ethereum network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Ethereum network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `geth --help` and the [CLI page](https://geth.ethereum.org/docs/fundamentals/command-line-options) for command line options. |
|   `clef`   | Stand-alone signing tool, which can be used as a backend signer for `geth`.                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|  `devp2p`  | Utilities to interact with nodes on the networking layer, without running a full blockchain.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|  `abigen`  | Source code generator to convert Ethereum contract definitions into easy-to-use, compile-time type-safe Go packages. It operates on plain [Ethereum contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) with expanded functionality if the contract bytecode is also available. However, it also accepts Solidity source files, making development much more streamlined. Please see our [Native DApps](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings) page for details.                                  |
| `bootnode` | Stripped down version of our Ethereum client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.                                                                                                                                                                                                                                               |
|   `evm`    | Developer utility version of the EVM (Ethereum Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug run`).                                                                                                                                                                                                                                               |
| `rlpdump`  | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp)) dumps (data encoding used by the Ethereum protocol both network as well as consensus wise) to user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                                |

## Running `geth`

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://geth.ethereum.org/docs/fundamentals/command-line-options)),
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `geth` instance.

### Hardware Requirements

Minimum:

* CPU with 2+ cores
* 4GB RAM
* 1TB free storage space to sync the Mainnet
* 8 MBit/sec download Internet service

Recommended:

* Fast CPU with 4+ cores
* 16GB+ RAM
* High-performance SSD with at least 1TB of free space
* 25+ MBit/sec download Internet service

### Full node on the main Ethereum network

By far the most common scenario is people wanting to simply interact with the Ethereum
network: create accounts; transfer funds; deploy and interact with contracts. For this
particular use case, the user doesn't care about years-old historical data, so we can
sync quickly to the current state of the network. To do so:

```shell
$ geth console
```

This command will:
 * Start `geth` in snap sync mode (default, can be changed with the `--syncmode` flag),
   causing it to download more data in exchange for avoiding processing the entire history
   of the Ethereum network, which is very CPU intensive.
 * Start the built-in interactive [JavaScript console](https://geth.ethereum.org/docs/interacting-with-geth/javascript-console),
   (via the trailing `console` subcommand) through which you can interact using [`web3` methods](https://github.com/ChainSafe/web3.js/blob/0.20.7/DOCUMENTATION.md) 
   (note: the `web3` version bundled within `geth` is very old, and not up to date with official docs),
   as well as `geth`'s own [management APIs](https://geth.ethereum.org/docs/interacting-with-geth/rpc).
   This tool is optional and if you leave it out you can always attach it to an already running
   `geth` instance with `geth attach`.

### A Full node on the Görli test network

Transitioning towards developers, if you'd like to play around with creating Ethereum
contracts, you almost certainly would like to do that without any real money involved until
you get the hang of the entire system. In other words, instead of attaching to the main
network, you want to join the **test** network with your node, which is fully equivalent to
the main network, but with play-Ether only.

```shell
$ geth --goerli console
```

The `console` subcommand has the same meaning as above and is equally
useful on the testnet too.

Specifying the `--goerli` flag, however, will reconfigure your `geth` instance a bit:

 * Instead of connecting to the main Ethereum network, the client will connect to the Görli
   test network, which uses different P2P bootnodes, different network IDs and genesis
   states.
 * Instead of using the default data directory (`~/.ethereum` on Linux for example), `geth`
   will nest itself one level deeper into a `goerli` subfolder (`~/.ethereum/goerli` on
   Linux). Note, on OSX and Linux this also means that attaching to a running testnet node
   requires the use of a custom endpoint since `geth attach` will try to attach to a
   production node endpoint by default, e.g.,
   `geth attach <datadir>/goerli/geth.ipc`. Windows users are not affected by
   this.

*Note: Although some internal protective measures prevent transactions from
crossing over between the main network and test network, you should always
use separate accounts for play and real money. Unless you manually move
accounts, `geth` will by default correctly separate the two networks and will not make any
accounts available between them.*

### Configuration

As an alternative to passing the numerous flags to the `geth` binary, you can also pass a
configuration file via:

```shell
$ geth --config /path/to/your_config.toml
```

To get an idea of how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ geth --your-favourite-flags dumpconfig
```

*Note: This works only with `geth` v1.6.0 and above.*

#### Docker quick start

One of the quickest ways to get Ethereum up and running on your machine is by using
Docker:

```shell
docker run -d --name ethereum-node -v /Users/alice/ethereum:/root \
           -p 8545:8545 -p 30303:30303 \
           ethereum/client-go
```

This will start `geth` in snap-sync mode with a DB memory allowance of 1GB, as the
above command does.  It will also create a persistent volume in your home directory for
saving your blockchain as well as map the default ports. There is also an `alpine` tag
available for a slim version of the image.

Do not forget `--http.addr 0.0.0.0`, if you want to access RPC from other containers
and/or hosts. By default, `geth` binds to the local interface and RPC endpoints are not
accessible from the outside.

### Programmatically interfacing `geth` nodes

As a developer, sooner rather than later you'll want to start interacting with `geth` and the
Ethereum network via your own programs and not manually through the console. To aid
this, `geth` has built-in support for a JSON-RPC based APIs ([standard APIs](https://ethereum.github.io/execution-apis/api-documentation/)
and [`geth` specific APIs](https://geth.ethereum.org/docs/interacting-with-geth/rpc)).
These can be exposed via HTTP, WebSockets and IPC (UNIX sockets on UNIX based
platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `geth`,
whereas the HTTP and WS interfaces need to manually be enabled and only expose a
subset of APIs due to security reasons. These can be turned on/off and configured as
you'd expect.

HTTP based JSON-RPC API options:

  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `8545`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
  * `--ws.origins` Origins from which to accept WebSocket requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,eth,miner,net,personal,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to
connect via HTTP, WS or IPC to a `geth` node configured with the above flags and you'll
need to speak [JSON-RPC](https://www.jsonrpc.org/specification) on all transports. You
can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based
transport before doing so! Hackers on the internet are actively trying to subvert
Ethereum nodes with exposed APIs! Further, all browser tabs can access locally
running web servers, so malicious web pages could try to subvert locally available
APIs!**

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for
granted in the official networks need to be manually set up. 

#### Generating genesis.json
First, you'll need to create the genesis state of your networks, which all nodes need to be
aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):
```json
{
  "config": {
    "chainId": <arbitrary positive integer>,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0
  },
  "alloc": {},
  "coinbase": "0x0000000000000000000000000000000000000000",
  "difficulty": "0x20000",
  "extraData": "",
  "gasLimit": "0x2fefd8",
  "nonce": "0x0000000000000042",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

The genesis file determines which consensus engine will be used, which hardfork changes will be supported, and other key configurations. 
Instead of wandering through countless docs to find a suitable Genesis file for the chain you want to create, you may just use **genesis_generator**

Make sure you built every debian packages by `make all`

```shell 
$ genesis_generator
```

This will help you generate genesis file by simply choosing the options it gives like below : 
``` shell
Which consensus engine to use? (default = Wemix)
 1. Ethash (proof-of-work)
 2. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)
 3. Clique (proof-of-authority)
 4. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)
 5. WBFT (wemix-byzantine-fault-tolerance)
 6. WEMIX (wemix-byzantine-fault-tolerance), merged from Wemix3.0 (proof-of-authority)
 ```

If you want more specific genesis file settings,simply modify the desired fields after it has been generated.


With the genesis state defined in the above JSON file, you'll need to initialize **every**
`geth` node with it prior to starting it up to ensure all blockchain parameters are correctly
set:

```shell
$ geth init path/to/genesis.json
```

#### Setting local private chain with wbft engine

Here's a simple example running single node for private chain with wbft engine.  
Note that this setting is not recommended for production.
 
1. Make `working directory`  
  <br>
2. Make geth folder inside `working directory`

    ```shell
    $ mkdir {working directory}/geth
    ```

2. Clone `go-wemix-qbft` inside `working directory` ( not mandatory. you can clone wherever you want. ) and move to `go-wemix-qbft`

    ```shell
    $ cd {path you clone go-wemix-qbft}/go-wemix-qbft
    ```

3. Make build file

    ```shell
    $ make all
    ```

4. Make `nodekey` inside `geth`

    ```shell
    $ ./build/bin/bootnode --genkey {working directory}/geth/nodekey
    ```

5. Check your enode address

    ```shell
    $ ./build/bin/bootnode -nodekey {working directory}/geth/nodekey
    
    Example)
    enode://adc70110af20a4e06b63c1b7c94bcaf61cd91f610afbdaf15d16cd279279438eded69763da2c7f861eb682594150d76900c126a15e50ccfb7989d1028fe26baf@127.0.0.1:0?discport=30301
    Note: you're using cmd/bootnode, a developer tool.
    We recommend using a regular node as bootstrap node for production deployments.
    INFO [12-20|10:53:21.527] New local node record                    seq=1,734,659,601,526 id=02148abb6456716e ip=<nil> udp=0 tcp=0
    ^C
    ```

6. Make genesis file (  From the following instructions, we assume that the genesis file has been created under the `working directory`. We also recommend you to create `config.toml` file)

    ```shell
   Example) 
   
   Which consensus engine to use? (default = Wbft)
     1. Ethash (proof-of-work)
     2. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)
     3. Clique (proof-of-authority)
     4. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)
     5. WBFT (wemix-byzantine-fault-tolerance)
     6. WEMIX (wemix-byzantine-fault-tolerance), merged from Wemix3.0 (proof-of-authority)
    > 5
    
    Which accounts are allowed to seal? (mandatory at least one)
    > 0x6B0d675682f92a771042a70F60b2f199628A2Ad0
    > 0x
    
    Want to generate config.toml file to configure static nodes to connect?
    Else you have to manage peer node manually (default true)
    > yes
    
    Enter enode URLs for static nodes (press enter with empty input when done):
    > enode://8937d3a33683e5395c4be88f1dcebf8d105bec5a88130177407ca2b960a68a4271c46b97b8f0aa097d7c18c96e410dbd0a38fdc956371e8dbae742bdc380428e@127.0.0.1:0?discport=30301
    > 
    
     Do you want to export generated config file?
     If not it will be just printed (default true)
    > yes
    
    Which folder to save the config.toml into? (default = current)
    > /my/working/directory
    
    Which accounts should be pre-funded? (advisable at least one)
    > 0x
    
    Specify your chain/network ID if you want an explicit one (default = random)
    > 
    
     Do you want to export generated genesis file?
     If not it will be just printed (default true)
    > yes
    
    Which folder to save the genesis spec into? (default = current)
    It will create genesis.json
    > /my/working/directory
    
    ```

7. init genesis block

    ```shell
    $ ./build/bin/geth --datadir {working directory} init {working directory}/genesis.json
    ```

8. run geth

    ```shell
    $ ./build/bin/geth --datadir {working directory} --http --http.addr "0.0.0.0" --http.port {httpPortNum}  --syncmode full --port {portNum}  --mine 
    ```

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to
start a bootstrap node that others can use to find each other in your network and/or over
the internet. The clean way is to configure and run a dedicated bootnode:

```shell
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an [`enode` URL](https://ethereum.org/en/developers/docs/networking-layer/network-addresses/#enode)
that other nodes can use to connect to it and exchange peer information. Make sure to
replace the displayed IP address information (most probably `[::]`) with your externally
accessible IP to get the actual `enode` URL.

*Note: You could also use a full-fledged `geth` node as a bootnode, but it's the less
recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try
`telnet <ip> <port>` to ensure it's indeed reachable), start every subsequent `geth`
node pointed to the bootnode for peer discovery via the `--bootnodes` flag. It will
probably also be desirable to keep the data directory of your private network separated, so
do also specify a custom `--datadir` flag.

```shell
$ geth --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll
also need to configure a miner to process transactions and create new blocks for you.*

#### Running a private miner


In a private network setting a single CPU miner instance is more than enough for
practical purposes as it can produce a stable stream of blocks at the correct intervals
without needing heavy resources (consider running on a single thread, no need for multiple
ones either). To start a `geth` instance for mining, run it with all your usual flags, extended
by:

```shell
$ geth <usual-flags> --mine --miner.threads=1 --miner.etherbase=0x0000000000000000000000000000000000000000
```

Which will start mining blocks and transactions on a single CPU thread, crediting all
proceedings to the account specified by `--miner.etherbase`. You can further tune the mining
by changing the default gas limit blocks converge to (`--miner.targetgaslimit`) and the price
transactions are accepted at (`--miner.gasprice`).

## Contribution

Thank you for considering helping out with the source code! We welcome contributions
from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-ethereum, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit
more complex changes though, please check up with the core devs first on [our Discord Server](https://discord.gg/invite/nthXNEv)
to ensure those changes are in line with the general philosophy of the project and/or get
some early feedback which can make both your efforts much lighter as well as our review
and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://geth.ethereum.org/docs/developers/geth-developer/dev-guide)
for more details on configuring your environment, managing project dependencies, and
testing procedures.

### Contributing to geth.ethereum.org

For contributions to the [go-ethereum website](https://geth.ethereum.org), please checkout and raise pull requests against the `website` branch.
For more detailed instructions please see the `website` branch [README](https://github.com/ethereum/go-ethereum/tree/website#readme) or the 
[contributing](https://geth.ethereum.org/docs/developers/geth-developer/contributing) page of the website.

## License

The go-ethereum library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The go-ethereum binaries (i.e. all code inside of the `cmd` directory) are licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
