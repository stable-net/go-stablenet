## StableNet Specification
This is the official Go implementation of the StableNet protocol, a fork of the WBFT protocol(https://github.com/wemixarchive/go-wbft) which is fully EVM compatible.

StableNet is a Chain Protocol with a Proof-of-Authority (PoA) architecture underpinned by a BFT consensus algorithm tailored for stablecoins. Our BFT implementation leverages the WBFT engine, an enhanced version of QBFT, designed for more practical applications. (For an in-depth understanding of WBFT, please refer to: https://github.com/wemixarchive/go-wbft?tab=readme-ov-file#wbft-protocol-specification)

The core innovation of the StableNet protocol lies in its ability to facilitate gas fee payments using stable tokens. To achieve this, we have implemented the following key features:

- Base Coin Policy for Stablecoins: A novel approach to managing the chain's native currency.
- Comprehensive Governance System: Establishing robust oversight for validators, minters, and master minters.
- NativeCoinAdapter: A built-in system contract enabling stablecoins to function seamlessly as ERC20 tokens.
- Optimized Gas Fee Policy: A carefully calibrated gas fee structure designed specifically for the stablecoin ecosystem.

Furthermore, we are actively researching and developing the following advanced features:
- High TPS for Real-World Payments: Aiming for exceptionally high transaction throughput to support everyday commercial transactions.
- Prioritized Mempool Policy for Minters: A sophisticated mempool mechanism designed to give precedence to minter transactions.
- Private Bank Functionality: A feature allowing users to obscure transaction amounts while maintaining transparency for financial regulatory bodies.

### A Pioneering Base Coin Policy for Stablecoins

Traditional blockchain protocols have been largely confined to a model where the base coin, used for gas fees, is pre-minted during the genesis phase. This structure often leads to a scenario where, with increasing adoption and successful DApps on the chain, the base coin's value appreciates, significantly enriching the initial genesis allocation wallet holders. This economic design, born from the powerful financial incentives of chain developers, has been widely adopted by many protocols without critical examination, often overlooking the inherent costs borne by users.

When a gas token, initially low in price, gains popularity and its value soars, users are compelled to pay higher transaction fees. The volatile nature of such tokens subjects holders to constant anxiety, making it impossible to separate coin ownership from investment speculation. This describes an optimistic scenario. Conversely, if a chain fails to gain market traction, the gas token's value plummets, leading to reduced profitability and the eventual departure of validators, undermining the chain's economic viability. In either case, whether a chain flourishes or becomes a relic, user convenience is often compromised.

This fundamental challenge underscores the necessity of a stablecoin chain.

Users desire stable and predictable transaction fees when interacting with DApps on a chain. They should be able to forecast their monthly fee expenditures based on their usage frequency. The ability to hold an easily usable base coin without concerns about price volatility is paramount. Furthermore, users should have the flexibility to redeem their stablecoins for fiat currency at any desired moment. Once a chain offers such capabilities, it opens the door for services catering to everyday individuals unfamiliar with blockchain technology – imagine seamlessly paying for a Starbucks coffee.

While existing blockchains have launched numerous financial services under the guise of "financial innovation," their deepest ambition often remains rooted in the desire for explosive coin price surges. As long as this ambition persists, blockchain's potential for genuine, real-world utility remains constrained. It is now time for the blockchain industry to acknowledge that shedding this ambition is crucial for blockchain to truly integrate into our daily lives.

The StableNet protocol has adopted the following policies to utilize stablecoins as its base coin:
- Minimal Genesis Pre-issuance: Only the minimum funds required for initial chain operation are pre-issued at genesis.
- Dynamic Token Management: Authorized Minters can issue and burn gas tokens throughout the chain's operation.
- Minter as Financial Oracle: Minters function as a form of oracle, verifying traditional financial systems. They are obligated to issue tokens only upon fiat currency deposits and to return fiat currency in proportion to burned tokens.
- Inflation-Free Model: The absence of a block minting system for validator rewards ensures an inflation-free environment.
- Regulatory Responsiveness and Stability: We have adopted a PoA structure to actively respond to government regulations and to move beyond profit-driven consortiums, fostering a more stable chain operation.
- Transparent Token Supply: The total issuance of all gas tokens on the chain is easily auditable. (It's surprisingly remarkable that such a fundamental feature is often absent in conventional chain protocols.)
  - This value remains constant unless mint/burn activities occur.

### A Refined Governance System for Validators, Minters, and Master Minters
In conventional public blockchains, token holders traditionally constituted the governance. While governance activities typically occur off-chain, the outcomes of these decisions are ultimately reflected in the chain node's code. The decentralized ethos of blockchain is indeed profound, yet its application to many existing systems of human economic activity can sometimes compromise efficiency. Particularly, when leveraging fiat currency deposit and withdrawal certifications as on-chain data, the immutable assumption is that someone must act as an oracle. While this oracle role can and should be distributed among multiple entities, it cannot be open to all participants.

This forms the basis for the necessity of governance within our chain. Without compromising the decentralized philosophy of blockchain, we have introduced a minimal yet effective governance structure for stablecoins, comprising:
- Validator Governance:
  - This collective operates the chain nodes and acts as the miner group.
  - They can register and remove new validators through consensus (voting).
  - Initially, transaction fees serve as the sole source of income; however, if insufficient, mechanisms like taxation will be introduced to ensure profitability.
  - Adopting the WBFT consensus algorithm, malicious Byzantine attacks are prevented unless more than 1/3 of the group is compromised.
  - A single validator is constituted by three distinct components: an operator wallet used for voting and governance management, a validator key for block generation, and a BLS key for signing WBFT consensus messages.
- Minter Governance:
  - This group is authorized to mint and burn the base coin.
  - They act as oracles for fiat currency deposits and withdrawals at banks.
  - They are strictly obligated to mint coins only in proportion to the deposited fiat currency and to withdraw fiat currency equivalent to the burned amount.
  - Minting and burning can only occur through the collective vote of all minters. (No single minter can unilaterally issue or burn tokens. However, in the future, minters following on-chain protocols like a native bridge might be able to act unilaterally.)
  - Minter admission and departure are also determined by minter votes.
- Master Minter Governance:
  - From the perspective of the base coin, Minter Governance is simply one type of minter. Another potential minter could be a native bridge.
  - The group responsible for managing the registration and removal of such base coin minters is termed Master Minter Governance.
  - This concept mirrors the master minter role in existing FiatToken implementations.
  - While Minter Governance membership (joining/leaving) is decided by its members' votes, the registration/removal of minters for the base coin is determined by Master Minter Governance.

The governance system is realized through the GovValidator, GovMinter, and GovMasterMinter contracts, which are deployed by the system at genesis without an owner. Upgrades are exclusively possible through hard forks, a testament to StableNet's unwavering commitment to the philosophy of decentralization.

#### Mint/Burn Protocol
As previously explained, the authority to mint and burn the base coin resides with the GovMinter. The GovMinter is composed of minter members, all of whom possess equal rights and responsibilities.

The protocol that minter members must adhere to when performing a mint operation is as follows:
- A minter member must be linked to a collateral account designated for token issuance.
- Upon detecting a deposit into the collateral account, the following information is bound together, forming what is referred to as a "minting proof":
  - deposit id
  - amount
  - beneficiary address
  - timestamp
- A minter member may then perform a mint operation for the beneficiary in an amount identical to that specified in the minting proof.
- If one minter member proposes a mint, the remaining minter members must validate the corresponding minting proof. If no issues are found, they will approve the proposal.
- The mint operation proceeds once a quorum of minter members have approved.

The protocol that minter members must adhere to when performing a burn operation is as follows:
- A minter member receives an off-chain burn request.
- The receiving minter member issues a withdrawal id and, based on this, generates a "burn proof":
  - withdrawal id
  - amount
  - token owner (from)
  - timestamp
- It is assumed that the withdrawal id can be shared among minter members off-chain.
- The receiving minter member proposes the burn using the burn proof.
- The remaining minter members verify and approve the proposal.
- The burn operation is executed once a quorum of minter members have approved.

### NativeCoinAdapter: Enabling ERC20 Compatibility for the Base Coin via a Built-in System Contract
One of StableNet's most distinctive features, unparalleled in any other chain, is its NativeCoinAdapter. Historically, blockchains have maintained separate APIs for interacting with their base coin and ERC20 tokens. Many established stablecoin services are built on the premise that stablecoins will inherently possess an ERC20 interface. While the concept of using a stablecoin as a base coin is not new and has been implemented by various projects, few have earnestly considered compatibility with legacy services that rely on this ERC20 assumption.

To avoid disrupting these legacy assumptions, we sought a method to transmit and query the base coin via an ERC20 interface. Our solution is to use the base coin through an ERC20 wrapper contract, which we call NativeCoinAdapter.

The specifications of this contract are as follows:
- Genesis System Contract: It is a system contract deployed at genesis.
- ERC20 Standard Compliance: The NativeCoinAdapter adheres to the ERC20 contract standard and fully supports the functionalities of FiatTokenV2_2 (https://github.com/circlefin/stablecoin-evm) implemented by Circle.
- Base Coin Wrapper: The NativeCoinAdapter acts as an ERC20-formulated wrapper for the base coin. This concept differs from the wrapped tokens for ETH commonly found in L2 solutions.
- Seamless Base Coin Transfer: Transfers executed via NativeCoinAdapter.transfer have the exact same effect as directly sending the base coin.
- Comprehensive Event Logging: All base coin transfers, even those for gas fee payments, generate a NativeCoinAdapter.Transfer event.
- Unified Balance Representation: An account's native balance and NativeCoinAdapter.balanceOf will always return identical values.
- Direct Base Coin Reference: This wrapper contract does not utilize its own storage for balances; instead, it directly references the account's base coin balance.
- Exclusive Mint/Burn Mechanism: The minting and burning of the base coin are exclusively performed through the mint and burn functions of the NativeCoinAdapter.
- Allowance Management: The allowance amount for NativeCoinAdapter.approve() is stored in the contract's storage, and precompile code is used to manage the approved account's base coin usage.
- Full ERC20 Compatibility and Traceability: Consequently, all existing ERC20 functions for stable token usage are fully compatible, with the added benefit of being able to track all base coin movements via Transfer events.

### A Gas Fee Policy Optimized for Stablecoin Environments
Ethereum-based blockchains commonly incorporate two primary components in their transaction gas fees: base fee and priority fee. These two fees serve distinct purposes. The base fee is designed to mitigate excessive transaction requests that exceed the blockchain's capacity. It dynamically adjusts with block congestion, serving as a defense mechanism against attacks like DDoS. Reflecting this public utility, the base fee is burned by the chain rather than being paid to miners. The priority fee, conversely, is paid to miners and functions as a 'tip' for users to request faster processing of their transactions.

These two fee policies necessitate modifications within a stablecoin chain. Firstly, due to the fundamental premise that stablecoins should never be burned without fiat currency redemption, the base fee cannot be incinerated. While an increased base fee, even when generated for public utility according to a defined protocol, could reasonably be paid to validators, it is not inherently irrational. In an inflation-free chain, allocating the base fee to miners might be more appropriate. Of course, there's a possibility that the base fee increase formula could be manipulated by validator collusion to be more sensitive to block congestion. However, under the assumption of a PoA consensus body, such adjustments would not be easily made if they compromise the chain's success. We have modified the existing Ethereum implementation, where the base fee began to rise when block capacity exceeded half, to now increase only when the block is nearly full (at 90%). The gradient of this increase has also been adjusted to reflect realistic usage.

The concept of the priority fee is fundamentally misaligned with stablecoins, thus requiring a complete re-evaluation. While in traditional systems miners could freely set maxPriorityFeePerGas, StableNet replaces this with a value determined by validator governance, forcing all miners to use the collectively voted-upon value. This means validators, through consensus and voting, determine the priority fee, and once set, it applies uniformly to all miners, precluding individual maxPriorityFeePerGas settings. Users can still employ EIP-1559 dynamic fee transactions, but any value higher than the set priority fee will be disregarded, and only the mandated fee price will be deducted. Conversely, setting a value lower than the determined fee will result in transaction rejection. Therefore, providing a higher priority fee does not entitle a user to prioritized transaction processing. However, a higher priority fee can be used to replace a previously sent transaction with the same nonce. StableNet has increased the maximum block gas limit to 105,000,000, allowing for approximately 5,000 basic transaction types per block. We assume that, unlike Ethereum, transactions rarely remain in a pending state within the mempool. Given this assumption, the concept of a 'tip' for priority processing is deemed invalid. Should blocks consistently fill with 5,000 transactions, the base fee would continuously rise, leading to a decrease in demand due to increased transaction costs. Nevertheless, we will continue to explore transaction prioritization in a manner suitable for a stable chain, seeking methods to prioritize transactions from a public utility perspective rather than solely through priority fees.

Stablenet fundamentally pegs its gas fees to the KRW currency. Protocol constants related to gas fees are defined in the protocol_params.go file with the following constants. If a different currency, such as USD, were to be used as the base coin, these constants would need to be adjusted to calculate appropriate gas fees.

```
GasTargetPercentage uint64 = 90               // Target gas usage as a percentage of the maximum gas limit
BaseFeeChangeRate   uint64 = 15               // Percentage rate by which the base fee can change
MinBaseFee          uint64 = 5000000000000    // Minimum base fee
MaxBaseFee          uint64 = 5000000000000000 // Maximum base fee. If set to 0, this limit is disabled.
InitialGasTip uint64 = 5000000000000 // Initial gas tip
```

### Anzeon WBFT Engine: A Tailored Evolution of WBFT for StableNet
The Anzeon WBFT Engine is a specially adapted version of the WEMIX Byzantine Fault Tolerance (WBFT) engine, meticulously modified to suit the unique requirements of the StableNet blockchain.

The original Byzantine Fault Tolerance (QBFT) engine was inherently limited to Proof-of-Authority (PoA) based systems, as validator participation was restricted to permissioned members. The WBFT engine was developed as an evolution of QBFT, introducing concepts such as diligence, staking, and slashing. These additions were crucial for its application in a fully public blockchain environment. Furthermore, WBFT was designed with the versatility to facilitate a seamless transition from existing legacy consensus engines (like PoW or WEMIX 3.0's non-BFT algorithms) to the WBFT engine via a hard fork at a specific block height. The Anzeon WBFT engine refines this general-purpose WBFT for the specific context of the StableNet chain.

Given StableNet's emphasis on public utility over pure commercial viability, a Proof-of-Authority (PoA) structure remains essential. Consequently, certain features vital for public blockchains, such as staking and diligence, are not strictly mandatory. Nevertheless, WBFT was chosen over QBFT due to its enhanced Byzantine fault tolerance and the robust framework of its built-in governance contracts. Even if not directly utilized for validator selection, the diligence mechanism can still serve as a valuable monitoring tool for assessing validator operational status.

Let's delve into the specific modifications implemented in the Anzeon WBFT engine:

The most significant change is the elimination of the Staking mechanism. While staking is a prerequisite for WBFT validator eligibility in a public blockchain, it has been removed from Anzeon WBFT. This decision was driven by several factors, including the contract-based nature of the validator set and the inability to introduce inflation to compensate for staking interest. The BLS key information, previously housed within GovStaking, has been migrated to GovValidator.

Further changes include:
- Removal of WPoA (WEMIX 3.0 legacy engine).
- Elimination of Block Rewards (including the associated Brioche hard fork logic).
- Overhaul of the Governance System:
  - Deprecated existing contracts: GovStaking, GovConfig, GovNCP, GovRewardeeImp.
  - Introduced new governance contracts: GovValidator, GovMinter, GovMasterMinter.
- Transition from Croissant config to Anzeon config:
  - The Croissant configuration could be activated via a hard fork at a specific block; the Anzeon configuration is applied from genesis.
- Removal of specific properties: stabilizingStakersThreshold, targetValidators, and useNCP.

Here's a sample code snippet for the Anzeon config:
```
"anzeon": {
  "wbft": {
    "requestTimeoutSeconds": 2,
    "blockPeriodSeconds": 1,
    "epochLength": 10,
    "proposerPolicy": 0,
    "maxRequestTimeoutSeconds": null
  },
  "init": {
    "validators": [
      "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697"
    ],
    "blsPublicKeys": [
      "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b"
    ]
  },
  "systemContracts": {
    "govValidator": {
      "address": "0x0000000000000000000000000000000000001001",
      "version": "v1",
      "params": {
          "members": "0xC3C49d11659170e525c3ed3E0D4560d485EF9229",
          "quorum": "1",
          "expiry": "604800",
          "memberVersion": "1",
          "validators": "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
          "blsPublicKeys": "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b"
      }
    }
  }
}
```
- The init section specifies the validators that will be active from block 1 through the first epoch.
- The systemContracts/govValidator section must define the following mandatory parameters. These values are used for the initialization of the GovValidator contract:
  - members: The addresses of each validator's operation key.
    - This can be either an EOA (Externally Owned Account) wallet address or a multisig contract address.
  - validators: The addresses of each validator's validator key.
    - This is the block signing address and corresponds to the coinbase of each block.
  - blsPublicKeys: The BLS public key for each validator.
    - This is used for signing WBFT consensus messages.
- These three lists must be comma-separated, and keys at the same index across the lists collectively form a single validator.
- It's important to note that the validator set specified in govValidator will take effect starting from the second epoch.
  - For the first epoch, the active validators are determined not by the GovValidator's list, but by the init configuration in the genesis block. 
  - Unless there is a specific reason otherwise, it is recommended to set the govValidator's validator set to be identical to the init configuration.


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

|          Command          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|:-------------------------:|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|        **`geth`**         | Our main Ethereum CLI client. It is the entry point into the Ethereum network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Ethereum network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `geth --help` and the [CLI page](https://geth.ethereum.org/docs/fundamentals/command-line-options) for command line options. |
| `genesis_generator` | Genesis generator tool. It is used for the first genesis.json file. |
|          `clef`           | Stand-alone signing tool, which can be used as a backend signer for `geth`.                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|         `devp2p`          | Utilities to interact with nodes on the networking layer, without running a full blockchain.                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|         `abigen`          | Source code generator to convert Ethereum contract definitions into easy-to-use, compile-time type-safe Go packages. It operates on plain [Ethereum contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) with expanded functionality if the contract bytecode is also available. However, it also accepts Solidity source files, making development much more streamlined. Please see our [Native DApps](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings) page for details.                |
|        `bootnode`         | Stripped down version of our Ethereum client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.                                                                                                                                                                                                                                                  |
|           `evm`           | Developer utility version of the EVM (Ethereum Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug run`).                                                                                                                                                                                                                                                  |
|         `rlpdump`         | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp)) dumps (data encoding used by the Ethereum protocol both network as well as consensus wise) to user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                                   |

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

#### Generating nodekey
To generate a nodekey, you can use the `bootnode` tool. This will create a new nodekey file in the current directory:

```
> bootnode -genkey mynodekey
```

And you can see the public key, address, and a bls public key derived from the nodekey:
```
> bootnode -nodekey mynodekey -writeaddress

public key: 0x9afad81eb6c7807b2f0edd2e4b55e07084d3ee66f28b563fa8ef1ca7147534f6acc734d069636263de70aa09f9f235398146684927fd147d75dcabb9c0762998
address: 0x6e817a2b618bdcabcf15587af4f04b787a8894bc
derived bls public key: 0xb7cfb997db1ccb18f4d84c5754ed0cbf1126791dec2ffa490c402c88bb0d5ea67df67a6d83c25dd047796376b59ac7b3
```

Use these information to define the genesis file. The address and bls key are used to define the `validators` and `blsPublicKeys` in the genesis file.

#### Generating genesis.json
First, you'll need to create the genesis state of your networks, which all nodes need to be
aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):
```json
{
  "config": {
    "chainId": 8282,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "muirGlacierBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "arrowGlacierBlock": 0,
    "grayGlacierBlock": 0,
    "anzeon": {
      "wbft": {
        "requestTimeoutSeconds": 2,
        "blockPeriodSeconds": 1,
        "epochLength": 10,
        "proposerPolicy": 0,
        "maxRequestTimeoutSeconds": null
      },
      "init": {
        "validators": [
          "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697"
        ],
        "blsPublicKeys": [
          "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b"
        ]
      },
      "systemContracts": {
        "govValidator": {
          "address": "0x0000000000000000000000000000000000001001",
          "version": "v1",
          "params": {
            "members": "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697",
            "quorum": "1",
            "expiry": "604800",
            "memberVersion": "1",
            "validators": "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
            "blsPublicKeys": "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b"
          }
        }
      }
    }
  },
  "nonce": "0x0",
  "timestamp": "0x68c93de9",
  "extraData": "0x",
  "gasLimit": "0x47b760",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {
    "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697": {
      "balance": "0x200000000000000000000000000000000000000000000000000000000000000"
    }
  },
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "baseFeePerGas": null,
  "fees": null,
  "excessBlobGas": null,
  "blobGasUsed": null
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
Which consensus engine to use? (default = Anzeon)
 1. Anzeon (WBFT for StableOne)
 2. Ethash (proof-of-work)
 3. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)
 4. Clique (proof-of-authority)
 5. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)
```

If you want more specific genesis file settings, simply modify the desired fields after it has been generated.


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

2. Clone `go-wemix-wbft` inside `working directory` ( not mandatory. you can clone wherever you want. ) and move to `go-wemix-wbft`

    ```shell
    $ cd {path you clone go-wemix-wbft}/go-wemix-wbft
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

    Which consensus engine to use? (default = Anzeon)
     1. Anzeon (WBFT for StableOne)
     2. Ethash (proof-of-work)
     3. Beacon (proof-of-stake), merging/merged from Ethash (proof-of-work)
     4. Clique (proof-of-authority)
     5. Beacon (proof-of-stake), merging/merged from Clique (proof-of-authority)
    > 1

    Which accounts are allowed to seal? (mandatory at least one)
    > 0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697
    └> BLS Public Key : 0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b
    > 0x
    
    Want to generate config.toml file to configure static nodes to connect?
    Else you have to manage peer node manually (default true)
    > no
    
    Which accounts should be pre-funded? (advisable at least one)
    > 0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697
    > 0x
    
    Specify your chain/network ID if you want an explicit one (default = random)
    >
    
    Do you want to export generated genesis file?
    If not it will be just printed (default true)
    >
    
    Which folder to save the genesis spec into? (default = current)
    It will create genesis.json
    >

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

If you'd like to contribute to go-wemix-wbft, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit
more complex changes though, please check up with the core devs first on [our team](mailto:developer@wemix.com)
to ensure those changes are in line with the general philosophy of the project and/or get
some early feedback which can make both your efforts much lighter as well as our review
and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `dev` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://geth.ethereum.org/docs/developers/geth-developer/dev-guide)
for more details on configuring your environment, managing project dependencies, and
testing procedures.

## License

The go-wemix-wbft library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The go-wemix-wbft binaries (i.e. all code inside of the `cmd` directory) are licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
