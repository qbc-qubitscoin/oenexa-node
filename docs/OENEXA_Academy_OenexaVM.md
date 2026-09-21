# OenexaVM Architecture: WebAssembly Smart Contracts

In standard blockchains like Ethereum, the Ethereum Virtual Machine (EVM) executes opcodes that modify the blockchain's state. OENEXA takes a vastly more performant approach: executing WebAssembly (WASM) via **OenexaVM**.

Today, we officially bridged the WASM execution environment to the core Cryptographic State Engine (Sparse Merkle Trie).

## How it Works

When a transaction of type `Deploy` or `Call` is executed on the network, it spins up a Sandboxed WASM Runtime (`wazero`). 

Normally, a WASM runtime is isolated from the host machine for security. To allow the contract to interact with the blockchain, we inject a highly secure **Host API Interface**:

### The Host API (`env`)
Smart contracts import the following functions directly from the OENEXA Node `env`:

- `oen_get(slot)`: Reads a 64-bit value from the contract's persistent storage inside the Sparse Merkle Trie.
- `oen_set(slot, value)`: Modifies the contract's storage. (Cost: 500 Gas).
- `oen_transfer(to_addr, amount)`: Allows the smart contract to securely send OEN coins to another address.
- `oen_balance(addr)`: Reads the OEN balance of any wallet on the blockchain.
- `oen_caller()`: Returns the address of the user who initiated the transaction.

## The State Integration
Prior to this upgrade, smart contract storage was floating in a temporary memory map. Now, the `StateAccessor` interface connects the VM directly into the `state.DB`. 

When `oen_set` is called, it caches the value in a `dirtyStorage` map. At the end of the block, `CommitRoot()` flushes every single contract storage slot across the entire blockchain into a massive 256-bit Sparse Merkle Trie, computing a unified cryptographically secure `StateRoot`.

### Why This Matters for Phases 12 & 14
With the ability to read balances, store data, and execute mathematical logic at near-native CPU speeds, **OENEXA** is now fully capable of hosting:
1. **DeFi (Phase 12)**: Automated Market Makers (AMMs) that read and write pool reserves into storage slots.
2. **HydroChain (Phase 14)**: Tokenized real-world assets that track ownership arrays and execute complex revenue distribution mathematics inside WASM.
