# OENEXA Node Development Notes

This document provides a comprehensive technical overview of the core logic, architectural patterns, performance optimizations, and development guidelines for the OENEXA Node architecture (aligned with Master Whitepaper v2.0).

---

## 1. Core State & 256-Bit Sparse Merkle Trie (`internal/state`, `internal/mpt`)

The `StateDB` is the core accounting and state-transition engine of the OENEXA network:
* **State Representation:** `StateDB` tracks transparent account balances, nonces, smart contract bytecode, and contract storage slots.
* **Performance Upgrade (Zero-Allocation Maps):** State maps are keyed using native `[20]byte` and `[32]byte` arrays instead of string hex representations. This eliminates dynamic heap allocations and string conversions inside high-frequency transaction loops, avoiding Garbage Collection pressure.
* **256-Bit Binary Sparse Merkle Trie (SMT):** Account addresses directly define the 256-step branch path ($0 = \text{left}, 1 = \text{right}$).
* **Dirty-Buffer Cache:** During block execution, state updates are buffered in an in-memory dirty map with $O(1)$ read and write latency (~91 ns/op).
* **Lazy Merkleization:** Only dirty leaves are re-hashed at block commit, reducing root calculation overhead by >80%.
* **Cryptographic Inclusion Proofs:** Light clients can verify any balance or storage slot using a 256-hash Merkle proof without downloading the full blockchain.

---

## 2. Mempool & Transaction Deduplication (`internal/mempool`)

The Mempool buffers and orders unconfirmed transactions broadcast across the P2P network:
* **Gas-Priority Heap:** Maintains a max-heap prioritizing transactions with the highest fee-to-gas ratio, ensuring optimal economic incentives for block proposers.
* **Batch Purge Optimization ($O(N)$ Complexity):** Committed transactions in a newly forged block are purged in a single batch-rebuild operation, replacing naive $O(N \cdot K)$ individual heap deletions that would stall block ingestion under load.
* **Deduplication & Throttling:** Nonces are validated sequentially per sender, and duplicate transaction hashes are discarded to protect against memory exhaustion and spam attacks.

---

## 3. Cryptography & Consensus (`internal/crypto`, `internal/consensus`)

* **Post-Quantum Cryptography (PQC):** OENEXA enforces NIST Post-Quantum Cryptography standards:
  - **ML-DSA-65 (FIPS 204)**: Lattice-based digital signatures for all transactions, block headers, and consensus votes.
  - **ML-KEM-768 (FIPS 203)**: Ephemeral key encapsulation for secure peer-to-peer transport handshakes.
  - **SHA-3-256 (FIPS 202)**: Standard hashing across Merkle trees, addresses, and note commitments.
* **PBFT Consensus State Machine:** 4-step voting lifecycle (Propose ──► Prevote ──► Precommit ──► Commit) achieving deterministic finality once $> 2/3$ weighted voting power signs.
* **State Transition Validation:** When receiving a block proposal, validators verify the proposer's ML-DSA-65 signature, previous block hash, transaction Merkle root, and replay the transactions to verify that gas used, burned fees, and the resulting state root match the header byte-for-byte.
* **Anti-Equivocation Slashing:** Conflicting votes or proposals from the same validator at the same height/round/step are flagged as Byzantine attacks, triggering immediate voting power revocation and a **10% stake burning penalty**.

---

## 4. Dual-Pool Privacy & Turnstile Conservation (`internal/shielded`)

* **Dual-Pool Model:** Operates transparent (`t-addr`) and shielded (`z-addr`) pools.
* **Note Commitments & Nullifiers:** Shielded funds exist as encrypted notes in an incremental Merkle accumulator tree of depth 32. Spending a note exposes a deterministic nullifier, preventing double-spending without revealing note ancestry.
* **Turnstile Supply Conservation:** Moving funds between pools (`TxShield` / `TxUnshield`) is audited on-chain at every block commit:
  $$\text{TotalSupply} = \text{TransparentSupply} + \text{ShieldedSupply}$$
* **Viewing Keys:** Users can generate read-only viewing keys (`vk_oen_...`) for tax and audit compliance without granting spending authority.

---

## 5. Smart Contracts & OenexaVM (`internal/vm`, `internal/contracts`)

* **Pure-Go WebAssembly Engine:** OenexaVM is built with `wazero` (`CGO_ENABLED=0`), offering memory safety and cross-platform compatibility without C dependencies.
* **Gas Metering & Sandbox Security:** Every WASM opcode and host call consumes gas according to a deterministic schedule. Contracts access state only through secure host functions.
* **Oenexa Cortex (`internal/contracts/cortex`):** Tokenized Tier-4 renewable data center bonds, on-chain Compute-as-a-Service (CaaS) leasing paid in OEN, and automated yield streaming to backers.
* **D-Commerce Escrow (`internal/contracts/dcommerce`):** Zero-trust physical delivery escrow using two-step QR code secrets ($S_{\text{pickup}}$ and $S_{\text{dropoff}}$) for atomic settlement, zero merchant platform fees, and micro-merchant POS.

---

## 6. Scalability & Layer-2 Rollups (`internal/rollup`)

* **Off-Chain Transaction Aggregation:** The Layer-2 sequencer collects and batches high-throughput transactions off-chain.
* **L1 State Anchoring:** Aggregated batches and cryptographic state roots are committed periodically to the Layer-1 chain via `RollupTx`, enabling throughput exceeding **100,000+ TPS**.

---

## 7. Network Protocol & Web3 RPC (`internal/p2p`, `internal/rpc`)

* **libp2p GossipSub:** Messages propagate over dedicated pubsub topics (`/oenexa/tx/1.0.0`, `/oenexa/block/1.0.0`, `/oenexa/vote/1.0.0`).
* **Web3 JSON-RPC 2.0:** Exposes standard and privacy-enhanced APIs (`oen_blockNumber`, `oen_getBalance`, `oen_sendRawTransaction`, `oen_chainInfo`, `oen_getShieldedBalance`, `oen_getTurnstileStatus`) on port `8545`.

---

## 8. Decoupled Architecture & Frontend API Contracts (`internal/web`)

* **Separation of Concerns:** The core node contains zero visual UI, npm dependencies, or HTML bundles. This keeps consensus binaries lean and secure.
* **REST Node Telemetry (`/api/status`):** `internal/web` exposes a lightweight health and status endpoint with permissive CORS headers to support decoupled browser applications.
* **Independent Client Repositories:** All user-facing web portals and dashboards are maintained in standalone repositories (such as [`oenexa-frontend`](https://github.com/oenexa/oenexa-frontend)), allowing rapid UI evolution without touching consensus code.

---

## Development & Contribution Workflow

1. **Test-Driven Development (TDD):** In accordance with [`PROJECT_RULES.md`](PROJECT_RULES.md), write unit tests before writing feature logic.
2. **Zero-Allocation Mindset:** Avoid string conversions or dynamic heap allocations inside core execution loops (`ApplyTx`, `CommitRoot`, `Mempool.Add`).
3. **Deterministic State:** Always sort map keys before hashing to guarantee consensus across all node implementations.
4. **Decoupled Boundary:** Never introduce JavaScript, HTML, or npm dependencies into the core node repository.
5. **Educational Curriculum:** Reference [`DEVELOPMENT_PROCESS.md`](DEVELOPMENT_PROCESS.md) and [`developer_notes/`](developer_notes/README.md) for deep conceptual and architectural notes.
