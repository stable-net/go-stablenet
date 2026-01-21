This test verifies `AccountManager.blacklist` functionality using a Helper contract acting as a Governance member.

The Helper contract (`Helper.sol`) compiled and used in this test:

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IGovCouncil {
    function proposeAddBlacklist(address account) external returns (uint256);
}

contract Helper {
    address constant GOV_COUNCIL = 0x0000000000000000000000000000000000001004;

    function run(address target) external {
        IGovCouncil(GOV_COUNCIL).proposeAddBlacklist(target);
    }
}
```

Process:

1. Transaction calls `Helper.run(target)`.
2. Helper invokes `GovCouncil.proposeAddBlacklist(target)`.
3. `GovCouncil` is configured with Quorum 1 and AutoExecute enabled.
4. The proposal is automatically approved and executed, triggering `AccountManager.blacklist`.
5. The tracer captures the state changes in `GovCouncil` (Storage) and `AccountManager` targets.

> **Note**: The `extra` value `9223372036854775808` observed in the trace result corresponds to `1 << 63`.
> Binary representation: `1000000000000000000000000000000000000000000000000000000000000000` (64 bits), which represents the Blacklist flag bit.
