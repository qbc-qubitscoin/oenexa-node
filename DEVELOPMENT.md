# OENEXA Node Development Notes

This document provides a comprehensive overview of the core logic, optimizations, and development notes for the OENEXA Node architecture.

## 1. Core State & Data Storage (`internal/state`)
The `StateDB` is the beating heart of the OENEXA network. It tracks all account balances, smart contract bytecode, and contract storage.
*   **Logic Overview:** `StateDB` uses an in-memory map protected by a Read-Write Mutex. It maps 20-byte cryptographic addresses to `Account` structures.
*   **Performance Upgrade (Zero-Allocation Maps):** Originally, the state maps were keyed using string representations (hex codes). This forced the Go runtime to dynamically allocate memory and run hex-encoding on the heap on every single transaction, destroying performance and saturating Garbage Collection. We refactored this to use strongly typed native `[20]byte` and `[32]byte` keys, achieving **zero-allocation** mapping.
*   **State Root Hash (`CommitRoot`):** To achieve consensus, all nodes must derive the exact same State Root hash. The node computes this by extracting all native byte keys, sorting them deterministically, and sequentially hashing the account data.

## 2. Mempool & Transaction Queue (`internal/mempool`)
The Mempool holds unconfirmed transactions broadcasted to the network before they are forged into a block.
*   **Logic Overview:** It maintains a priority queue (implemented as a Go `heap`) to ensure transactions with the highest fee-to-gas ratio are prioritized by validators.
*   **Performance Upgrade (O(N) Purge Optimization):** Originally, when a block was forged, the node deleted committed transactions from the mempool individually. Re-sorting the heap after every individual deletion resulted in a catastrophic `O(N * K)` complexity (where N is mempool size, K is transactions committed). This would stall the node under load. We upgraded the logic to batch-delete and rebuild the heap only **once** per block, reducing complexity to `O(N)`. 
*   **Memory Efficiency:** Just like the `StateDB`, the mempool was upgraded to bypass string keys and use native byte arrays (`[32]byte`), significantly accelerating transaction lookup times.

## 3. Cryptography & Consensus (`internal/crypto`, `internal/consensus`)
*   **Post-Quantum Cryptography:** OENEXA utilizes ML-DSA-65 (Dilithium) for signing and verifying all transactions and blocks. This provides quantum resistance.
*   **Consensus Mechanism:** The node uses a hybrid PoS (Proof of Stake) gossip architecture where validators vote on block proposals. The system validates the ML-DSA signatures of all validators before committing a block.

## 4. Smart Contracts VM (`internal/vm`)
*   **WebAssembly (WASM):** OENEXA uses Wazero to instantiate isolated WebAssembly virtual machines for each smart contract execution.
*   **Logic Overview:** The VM is injected with "host functions" that allow the WASM bytecode to securely interact with the `StateDB` (e.g., reading/writing storage, sending tokens) without escaping the VM sandbox.

## 5. Network Protocol (`internal/p2p`)
*   **Gossip Network:** Nodes communicate using a libp2p-based gossip protocol, caching recent messages to prevent infinite message loops.

## Development Workflow
When making upgrades to the codebase:
1. Ensure `go test ./...` passes to verify deterministic state execution.
2. Beware of heap allocations (like `crypto.ToHex` or string conversions) inside core loops, as blockchain systems require maximum memory efficiency.
3. Keep `StateDB` deterministic; any map iteration must be sorted before hashing, as Go map iteration order is highly randomized.
