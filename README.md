# OENEXA (OEN) — Quantum-Safe Layer-1 Blockchain
### *The Post-Quantum Layer-1 for AI, Data Centers, and Everyday Commerce*
#### Master Whitepaper (Version 2.0) Implementation

[![Go Version](https://img.shields.io/badge/Go-1.26.2-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![NIST PQC](https://img.shields.io/badge/Cryptography-NIST%20FIPS%20203%20%7C%20204-7928CA?logo=shield)](https://csrc.nist.gov/projects/post-quantum-cryptography)
[![Consensus: PBFT](https://img.shields.io/badge/Consensus-PBFT%20Proof--of--Stake-FF5722)](internal/consensus)
[![Privacy: Dual-Pool](https://img.shields.io/badge/Privacy-Dual--Pool%20Shielded-8A2BE2)](internal/shielded)
[![Smart Contracts](https://img.shields.io/badge/WASM%20VM-Wazero%20(Pure--Go)-00C7B7)](internal/vm)
[![AI Grid](https://img.shields.io/badge/Cortex-AI%20Data%20Center%20Grid-00E676)](internal/contracts/cortex)
[![D-Commerce](https://img.shields.io/badge/Escrow-Zero--Trust%20D--Commerce-FFB300)](internal/contracts/dcommerce)
[![Coverage](https://img.shields.io/badge/Coverage-100%25%20Statements-brightgreen)](TASK_RECORD.md)
[![Architecture](https://img.shields.io/badge/Architecture-Decoupled%20Backend-00C7B7)](internal/web)
[![Frontend](https://img.shields.io/badge/Frontend-oenexa--frontend-61DAFB?logo=react&logoColor=black)](https://github.com/oenexa/oenexa-frontend)

---

## Table of Contents

1. [Executive Overview](#1-executive-overview)
2. [End-to-End System Architecture](#2-end-to-end-system-architecture)
3. [Post-Quantum Cryptography Stack](#3-post-quantum-cryptography-stack)
4. [Dual-Pool Privacy & Turnstile Conservation](#4-dual-pool-privacy--turnstile-conservation)
5. [PBFT Consensus & Equivocation Slashing](#5-pbft-consensus--equivocation-slashing)
6. [High-Performance Sparse Merkle Trie (SMT)](#6-high-performance-sparse-merkle-trie-smt)
7. [Oenexa Cortex: Green Data Center AI Grid](#7-oenexa-cortex-green-data-center-ai-grid)
8. [Decentralized Everyday Commerce (D-Commerce)](#8-decentralized-everyday-commerce-d-commerce)
9. [OenexaVM WebAssembly Execution Runtime](#9-oenexavm-webassembly-execution-runtime)
10. [Layer-2 Rollups & Scalability](#10-layer-2-rollups--scalability)
11. [Currency, Gas & Tokenomics (EIP-1559)](#11-currency-gas--tokenomics-eip-1559)
12. [Repository Directory Structure](#12-repository-directory-structure)
13. [Building, Running & CLI Guide](#13-building-running--cli-guide)
    - [A. Compilation](#a-compilation)
    - [B. Running a Full Node (`oenexa-node`)](#b-running-a-full-node-oenexa-node)
    - [C. High-Speed Client Wallet CLI (`oenexa-cli`)](#c-high-speed-client-wallet-cli-oenexa-cli)
    - [D. Shielded Pool CLI Operations](#d-shielded-pool-cli-operations)
    - [E. Interactive Web Portal & Explorer](#e-interactive-web-portal--explorer)
14. [Web3 JSON-RPC 2.0 API Specification](#14-web3-json-rpc-20-api-specification)
15. [Testing, Benchmarking & Quality Assurance](#15-testing-benchmarking--quality-assurance)
16. [Educational Curriculum & Student Roadmap](#16-educational-curriculum--student-roadmap)
17. [License](#17-license)

---

## 1. Executive Overview

**OENEXA (OEN)** (*Open Economy, Next Generation Exchange & Assets*) is a Generation-6, post-quantum, privacy-preserving Layer-1 blockchain ecosystem written in pure Go (`CGO_ENABLED=0`). Engineered to endure for the next half-century, OENEXA replaces classical vulnerable algorithms with **NIST Post-Quantum Cryptography (PQC)** standards, while unifying:

- **AI & Hyperscale Infrastructure**: Tokenizing and funding Tier-4 green data centers via **Oenexa Cortex**, providing on-chain Compute-as-a-Service (CaaS) leasing and autonomous yield distribution.
- **Zero-Trust D-Commerce**: Dismantling extractive Web2 delivery platforms (which levy up to 30% fees) through two-party, QR-code mediated physical delivery escrow contracts.
- **Dual-Pool Shielded Privacy**: Zcash-class confidential note commitments and nullifiers with strict turnstile supply conservation guarantees (`TotalSupply == TransparentSupply + ShieldedSupply`) and selective viewing keys.
- **Instant PBFT Finality**: Sub-second deterministic consensus with byte-for-byte state transition validation and automated equivocation (double-voting) slashing.
- **Pure-Go WASM Smart Contracts**: Sandboxed, gas-metered execution powered by `wazero`.
- **Educational Legacy**: Purpose-built as an open engineering textbook for students, researchers, and core engineers exploring quantum-safe distributed systems.

---

## 2. End-to-End System Architecture

```
+----------------------------------------------------------------------------------------------------+
|                                    OENEXA Node Architecture                                        |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|    [ Users & DApps ] <======== JSON-RPC 2.0 / WebSockets (Port 8545) =======> [ React 18 UI / CLI ]|
|            │                                                                         │             |
|            ▼                                                                         ▼             |
|    ┌───────────────┐                  libp2p GossipSub                      ┌─────────────────┐    |
|    │  RPC Ingest   │ ◄─────── Topics: /oenexa/tx, /block, /vote ──────────► │  P2P Mesh Node  │    |
|    └───────┬───────┘                                                        └────────┬────────┘    |
|            │                                                                         │             |
|            ▼                                                                         ▼             |
|    ┌───────────────────────────────────────────────┐                        ┌─────────────────┐    |
|    │               Mempool Engine                  │ ◄───────────────────── │  KEM Handshake  │    |
|    │ - Gas-Priority Heap & Deduplication           │                        │  (ML-KEM-768)   │    |
|    │ - Sender Nonce Queue & Eviction Policies      │                        └─────────────────┘    |
|    └───────────────────────┬───────────────────────┘                                               |
|                            │ Pull Pending Txs                                                      |
|                            ▼                                                                       |
|    ┌─────────────────────────────────────────────────────────────────────────┐                     |
|    │                       PBFT Consensus Engine                             │                     |
|    │  - Round-Robin Proposer Selection (H mod |Validators|)                  │                     |
|    │  - 4-Step State Machine: Propose ──► Prevote ──► Precommit ──► Commit   │                     |
|    │  - Quorum Threshold: > 2/3 Weighted Voting Power                        │                     |
|    │  - Verification: Header, PrevHash, Tx Merkle, StateRoot Replay          │                     |
|    │  - Anti-Equivocation Slashing: Vote Conflict Detection & 10% Burn       │                     |
|    └───────────────────────┬─────────────────────────────────────────────────┘                     |
|                            │ Block Execution & State Transition                                    |
|                            ▼                                                                       |
|    ┌─────────────────────────────────────────────────────────────────────────┐                     |
|    │                  State Transition Engine (ApplyBlock)                   │                     |
|    │  ┌─────────────────────────────┐     ┌────────────────────────────────┐ │                     |
|    │  │     Transparent Pool        │     │         Shielded Pool          │ │                     |
|    │  │  - ML-DSA-65 Signatures     │     │  - Note Commitments (Depth 32) │ │                     |
|    │  │  - EIP-1559 Dynamic Fee Burn│     │  - Nullifier Set (Anti-Replay) │ │                     |
|    │  │  - Account Balances & Nonces│     │  - Viewing Key Disclosures     │ │                     |
|    │  └──────────────┬──────────────┘     └───────────────┬────────────────┘ │                     |
|    │                 │             Turnstile Model        │                  │                     |
|    │                 └─────────── TotalSupply Conserved ──┘                  │                     |
|    │  ┌────────────────────────────────────────────────────────────────────┐ │                     |
|    │  │            OenexaVM WebAssembly Runtime (Pure-Go wazero)           │ │                     |
|    │  │  - Gas Metering & Host Environment Callbacks                       │ │                     |
|    │  │  - Oenexa Cortex (Data Center Bonds & CaaS Compute Yields)         │ │                     |
|    │  │  - D-Commerce Zero-Trust QR Delivery Escrow Contracts             │ │                     |
|    │  │  - DeFi Hub (OenSwap AMM, Collateralized Lending, RWA)             │ │                     |
|    │  └────────────────────────────────────────────────────────────────────┘ │                     |
|    └───────────────────────┬─────────────────────────────────────────────────┘                     |
|                            │ Commit Dirty State                                                    |
|                            ▼                                                                       |
|    ┌─────────────────────────────────────────────────────────────────────────┐                     |
|    │               256-Bit Sparse Merkle Trie (SMT & StateDB)                │                     |
|    │  - In-Memory Dirty-Account Buffer (~91 ns/op O(1) Access)               │                     |
|    │  - Deterministic 256-Bit Binary Key Paths & Lazy Merkleization          │                     |
|    │  - Cryptographic Inclusion Proofs for Mobile Light Clients              │                     |
|    └───────────────────────┬─────────────────────────────────────────────────┘                     |
|                            │ Disk Persistence                                                      |
|                            ▼                                                                       |
|    ┌─────────────────────────────────────────────────────────────────────────┐                     |
|    │            Storage Layer: GoLevelDB / BadgerDB Engine                   │                     |
|    │  - Immutable BlockStore, ReceiptStore, and SMT Node Records             │                     |
|    +-------------------------------------------------------------------------+                     |
+----------------------------------------------------------------------------------------------------+
```

---

## 3. Post-Quantum Cryptography Stack

OENEXA eliminates all legacy asymmetric primitives vulnerable to Shor's algorithm on a quantum computer (RSA, ECDSA, Ed25519, secp256k1, and pairing curves). The entire node operates strictly on NIST Post-Quantum Cryptography standards:

| Primitive | Standard / Spec | Output / Size | Implementation / Role |
| :--- | :--- | :--- | :--- |
| **ML-DSA-65** | **NIST FIPS 204** (Dilithium3) | Public Key: 1,952 B<br>Signature: 3,309 B | Transparent transactions (`t-addr`), block proposal signatures, validator PBFT consensus voting, and governance. |
| **ML-KEM-768** | **NIST FIPS 203** (Kyber768) | Ciphertext: 1,088 B<br>Shared Secret: 32 B | Ephemeral quantum-resistant P2P session key encapsulation, preventing *"Harvest Now, Decrypt Later"* espionage. |
| **SHA-3-256** | **NIST FIPS 202** | 32-byte digest | Cryptographic address derivation ($\text{Addr} = \text{SHA3-256}(\text{PubKey})$), block headers, SMT hashing, and note commitments. |
| **AES-256-GCM** | **NIST SP 800-38D** | 32-byte key, 12-byte IV | Authenticated P2P wire communication encryption and local disk keystore encryption. |
| **Argon2id** | **RFC 9106** | 64 MB memory, 3 iterations | Memory-hard key derivation function (KDF) protecting encrypted private spending keys at rest. |

---

## 4. Dual-Pool Privacy & Turnstile Conservation

OENEXA merges transparent auditability with mathematically verifiable confidentiality through a dual-pool model:

```
┌─────────────────────────────────┐           Turnstile Deposit           ┌─────────────────────────────────┐
│        Transparent Pool         │ ──────────── TxShield ──────────────► │          Shielded Pool          │
│      (t-addr / ML-DSA-65)       │                                       │       (z-addr / Note Tree)      │
│     Public balances & flows     │ ◄─────────── TxUnshield ───────────── │    Hidden amounts & identities  │
└─────────────────────────────────┘           Turnstile Withdraw          └─────────────────────────────────┘
                 │                                                                         │
                 └──────────────────────── Total Supply Invariant ─────────────────────────┘
                                   TotalSupply = TransparentSupply + ShieldedSupply
```

### Key Properties
- **Transparent Pool (`t-addr`)**: Publicly auditable accounts signed with ML-DSA-65. Used for exchange settlement, public audits, and smart contract invocations.
- **Shielded Pool (`z-addr`)**: Obscures sender, recipient, and transfer amounts. Value is encapsulated into cryptographic **Notes** stored in an incremental Merkle accumulator tree of depth 32.
- **Nullifier Invariant**: Spending a note reveals its unique, deterministic nullifier on-chain. Duplicate nullifiers are rejected by the state machine, preventing double-spending without revealing which note was spent.
- **Turnstile Supply Conservation Guarantee**: All crossings between pools (`TxShield` and `TxUnshield`) are strictly accounted for in the turnstile state. The engine validates that `TotalSupply == TransparentSupply + ShieldedSupply` at every block commit, guaranteeing **zero stealth inflation**.
- **Selective Regulatory Disclosure**: Users can export read-only **Viewing Keys** (`vk_oen_...`) to tax authorities or auditors without forfeiting spending control.

---

## 5. PBFT Consensus & Equivocation Slashing

OENEXA delivers sub-second finality using an upgraded Practical Byzantine Fault Tolerant (PBFT) consensus engine (`internal/consensus`):

### 1. Deterministic Proposer Selection
At block height $H$, the active round-robin proposer index is computed deterministically:
$$\text{ProposerIndex} = H \pmod{|\text{ValidatorSet}|}$$

### 2. Four-Step Voting Lifecycle
1. **Propose**: The designated proposer pulls pending transactions from the mempool, constructs the block candidate, signs it via ML-DSA-65, and broadcasts it across `/oenexa/block/1.0.0`.
2. **Prevote**: Validators execute full state transition verification (verifying previous block hash, proposer signature, transaction Merkle root, gas consumption, and deterministic state root byte-for-byte). If valid, the validator broadcasts a signed `VotePrevote`.
3. **Precommit**: Once $> \frac{2}{3}$ of the total voting power in Prevotes is observed, the validator advances to `StepPrecommit` and broadcasts a signed `VotePrecommit`.
4. **Commit & Absolute Finality**: Upon receiving $> \frac{2}{3}$ Precommits, the block is finalized, committed to disk, and committed transactions are purged from the mempool. No chain reorganizations can occur.

### 3. Anti-Equivocation Slashing
If a Byzantine validator attempts a double-voting attack (signing two conflicting block proposals or votes for the same height, round, and step):
- The consensus engine immediately identifies the conflicting signatures.
- The malicious validator's voting power is revoked from the active set.
- **10% of their staked on-chain balance is permanently slashed and burned**.

---

## 6. High-Performance Sparse Merkle Trie (SMT)

The state storage engine (`internal/state`, `internal/mpt`) employs a **256-bit binary Sparse Merkle Trie**:

- **Deterministic 256-Bit Branching**: Account addresses directly dictate the 256-bit traversal path ($0 = \text{left}, 1 = \text{right}$).
- **Dirty-Buffer Cache**: Transaction state updates are buffered in an in-memory dirty map with $O(1)$ read/write latency (~91 ns/op).
- **Lazy Merkleization**: Only modified leaves are re-hashed at block commit, reducing Merkle root generation overhead by over 80%.
- **Cryptographic Inclusion Proofs**: Light clients and mobile wallets can verify any account balance or state item in $\approx 1.0\ \mu\text{s}$ using a 256-hash proof without downloading the blockchain.

---

## 7. Oenexa Cortex: Green Data Center AI Grid

The exponential rise of Artificial Intelligence demands sustainable, high-capacity computing power. OENEXA natively integrates data center financing and compute leasing through **Oenexa Cortex** (`internal/contracts/cortex`):

```
+─────────────────────────────────────────────────────────────────────────────────+
|                                  Oenexa Cortex                                  |
+─────────────────────────────────────────────────────────────────────────────────+
|                                                                                 |
|   1. Fractional Bonds       2. Renewable Telemetry      3. CaaS Compute Leasing |
|   ┌────────────────────┐    ┌────────────────────┐    ┌────────────────────┐    |
|   │ Tier-4 Data Center │    │ Decentralized      │    │ AI Developers pay  │    |
|   │ tokenized into     │    │ Oracles stream     │    │ OEN for GPU/TPU    │    |
|   │ CortexAsset shares │    │ kWh & GPU hours    │    │ cluster compute    │    |
|   └─────────┬──────────┘    └─────────┬──────────┘    └─────────┬──────────┘    |
|             │                         │                         │               |
|             ▼                         ▼                         ▼               |
|   ┌────────────────────────────────────────────────────────────────────────┐    |
|   │                       Cortex Smart Contract                            │    |
|   │  - Infrastructure Bond Registry & Stake Tracking                       │    |
|   │  - RevenuePool Accumulation (100% compute fees in OEN)                 │    |
|   │  - Proportional Autonomous Yield Streaming to Backers:                 │    |
|   │      Yield = (Shares / TotalShares) * RevenuePool - Claimed            │    |
|   └────────────────────────────────────────────────────────────────────────┘    |
+─────────────────────────────────────────────────────────────────────────────────+
```

- **Fractional Infrastructure Bonds**: Hyperscale, 100% renewable-powered data centers are tokenized as `CortexAsset` contracts. Users pool OEN to fund construction and earn fractional ownership.
- **Compute-as-a-Service (CaaS)**: Decentralized GPU/TPU clusters are leased to developers for LLM training and autonomous AI agent execution, settled natively in OEN.
- **SaaS Development & Subscriptions**: Software platforms build on the grid with recurring OEN subscription auto-deductions and decentralized crowdfunding.
- **Fair Yield Streaming**: 100% of compute leasing fees pool into the contract, allowing backers to claim transparent dividends proportional to their ownership.

---

## 8. Decentralized Everyday Commerce (D-Commerce)

To liberate local economies from centralized delivery platforms charging predatory 20–30% commissions, OENEXA deploys a zero-trust physical delivery escrow protocol (`internal/contracts/dcommerce`):

```
Buyer                         Restaurant                    Courier
  │                               │                            │
  │ 1. Place Order & Escrow OEN   │                            │
  │    (Food + Delivery Tip)      │                            │
  │    Submits Hash(S_dropoff)    │                            │
  │───────────────────────────────┼───────────────────────────►│
  │                               │ 2. Prepares Order          │
  │                               │    Submits Hash(S_pickup)  │
  │                               │                            │
  │                               │ 3. Courier Arrives         │
  │                               │    Courier scans QR pickup │
  │                               │ ◄──────────────────────────│
  │                               │    Reveals S_pickup        │
  │                               │    Calls ConfirmPickup()   │
  │                               │    State: PICKED_UP        │
  │ 4. Courier Arrives at Home    │                            │
  │    Buyer displays QR dropoff  │                            │
  │ ──► Courier scans QR ─────────┼───────────────────────────►│
  │    Reveals S_dropoff          │                            │
  │    Calls ConfirmDelivery()    │                            │
  │                               │                            │
  │ 5. Instant Atomic Settlement  │                            │
  │    Restaurant receives 100% food payment                   │
  │    Courier receives 100% delivery commission               │
  ▼    ZERO Middleman Fees        ▼                            ▼
```

- **Zero Platform Cuts**: 100% of food cost settles to the restaurant; 100% of the delivery fee settles to the courier.
- **Two-Factor Cryptographic Handoff**: Secret preimage verification ($S_{\text{pickup}}$ and $S_{\text{dropoff}}$) ensures couriers cannot claim delivery without the customer physically scanning their code.
- **Micro-Merchant POS**: Street vendors and retail shops accept OEN instantly via mobile wallets with sub-second finality and zero chargebacks.

---

## 9. OenexaVM WebAssembly Execution Runtime

Smart contracts in OENEXA are compiled to standard WebAssembly (WASM) and executed via **wazero** (`internal/vm`), a pure-Go, zero-CGo runtime:

- **Polyglot Contracts**: Write smart contracts in Go, Rust, C, or AssemblyScript and compile to `.wasm`.
- **Deterministic Metering**: Every WASM opcode and host call consumes gas according to a strict instruction fee schedule.
- **Isolated Sandbox**: Contracts interact with `StateDB` strictly through secured host callbacks (read storage, write storage, transfer balance, log event). Contracts cannot escape memory bounds or invoke system syscalls.

---

## 10. Layer-2 Rollups & Scalability

To support planetary-scale transaction volume (millions of concurrent micro-transactions across D-Commerce and SaaS subscriptions), OENEXA incorporates an integrated Layer-2 Rollup engine (`internal/rollup`):

- **L2 Sequencer**: Ingests high-throughput transactions off-chain, verifies state transitions, and produces compact batches.
- **L1 Bridge Anchor**: Periodically posts aggregated `RollupTx` batches and cryptographic root assertions to Layer-1.
- **Target Performance**: Scaling OENEXA capacity to over **100,000+ TPS** with negligible transaction overhead.

---

## 11. Currency, Gas & Tokenomics (EIP-1559)

| Parameter | Specification |
| :--- | :--- |
| **Asset Symbol** | **OEN** (*Open Economy, Next Generation Exchange & Assets*) |
| **Atomic Unit** | **nano-OEN** ($1\ \text{OEN} = 10^9\ \text{nano-OEN}$; 9 decimal places) |
| **Hard Cap** | **100,000,000 OEN** |
| **Genesis Allocation** | 10,000,000 OEN (10% allocated to ecosystem development, grants & community reserve) |
| **Mining / Staking Subsidy** | 90,000,000 OEN emitted across 32 halving eras |
| **Era 0 Block Reward** | 45.000000000 OEN per block |
| **Block Time Target** | 2.0 seconds |
| **Block Gas Limit** | 30,000,000 gas |
| **Fee Architecture** | **EIP-1559 Dynamic Fee Mechanism**: Base fee is deterministically adjusted and burned on-chain; priority tip is awarded to the block proposer. |

---

## 12. Repository Directory Structure

```
oenexa/
├── cmd/
│   ├── node/                   # Full node daemon (oenexa-node start / wallet / query / tx)
│   ├── oenexa-cli/             # High-speed client wallet & RPC CLI (keygen / balance / transfer / info)
│   ├── loadtest/               # High-throughput benchmarking & stress-testing tool
│   ├── multisig/               # Post-quantum M-of-N multi-signature CLI utility
│   └── oenexaid/               # Decentralized Identity (DID/VC) credential generator
├── configs/                    # Production and testnet configurations (mainnet.toml, testnet.toml)
├── developer_notes/            # Comprehensive "Why & How" architectural and educational notes (01–08)
├── docs/                       # Specifications, security audits, benchmarks, and 32 phase documentations
├── internal/
│   ├── config/                 # TOML configuration parser and validator
│   ├── consensus/              # PBFT engine, validator set, proposer selection, slashing
│   ├── contracts/              # Native smart contract implementations:
│   │   ├── aimarket/           # Autonomous AI agent discovery, prompt monetization & tasks
│   │   ├── autonomous/         # Self-tuning autonomous economic parameters
│   │   ├── carbonx/            # Carbon credit registry & ecological verification
│   │   ├── cbdc/               # Central Bank Digital Currency cross-border settlement
│   │   ├── compute/            # Decentralized compute task matching and settlement
│   │   ├── cortex/             # Oenexa Cortex AI data center bonds & CaaS yield streaming
│   │   ├── custody/            # Institutional multi-tier cold custody & timelocks
│   │   ├── dcommerce/          # Zero-trust QR-code mediated physical delivery escrow
│   │   ├── dex/                # OenSwap AMM pools (x * y = k) & confidential liquidity
│   │   ├── energyx/            # P2P renewable energy grid trading & settlement
│   │   ├── esgmarket/          # Certified ESG marketplace & compliance registry
│   │   ├── esgnetwork/         # Global IoT environmental sensor verification mesh
│   │   ├── global/             # Multi-jurisdictional compliance routing & state anchors
│   │   ├── govpartner/         # Sovereign & government observer nodes
│   │   ├── greendao/           # DAO governance, quadratic voting & treasury timelocks
│   │   ├── hydrochain/         # Green hydrogen supply-chain lifecycle provenance
│   │   ├── iso20022/           # ISO 20022 financial messaging translation (pacs.008, pain.001)
│   │   ├── lending/            # Collateralized lending & interest rate models
│   │   ├── mainnet/            # Mainnet parameters, staking rules & genesis allocations
│   │   ├── multisig/           # Quantum-resistant M-of-N multi-sig smart contract
│   │   ├── oenexaid/           # Decentralized Identity (DID) & Verifiable Credentials (VC)
│   │   ├── oracle/             # Multi-source price and telemetry oracle feeds
│   │   ├── quantumnet/         # Quantum communication & entanglement simulation
│   │   ├── rwa/                # Real-World Asset (RWA) tokenization & compliance
│   │   ├── sovereign/          # Sovereign wealth fund reserves & legal wrappers
│   │   ├── storage/            # Decentralized storage contracts with storage proofs
│   │   └── superapp/           # Unified multi-service mobile aggregation contract
│   ├── core/                   # Block, Transaction, Genesis, Merkle tree, Tokenomics, EIP-1559
│   ├── crypto/                 # ML-DSA-65 signatures, ML-KEM-768 key encapsulation, SHA-3-256
│   ├── identity/               # W3C DID and Verifiable Credential primitives
│   ├── keystore/               # Encrypted disk keystores using Argon2id + AES-256-GCM
│   ├── mempool/                # Gas-price priority queue with sender throttling & deduplication
│   ├── metrics/                # Prometheus metrics exporter and /healthz endpoint
│   ├── mpt/                    # 256-bit Sparse Merkle Trie (SMT) with dirty-buffer caching
│   ├── node/                   # Lifecycle coordinator wiring consensus, p2p, state, and RPC
│   ├── oracle/                 # Standalone oracle aggregation daemon service
│   ├── p2p/                    # libp2p network, KEM handshake, gossip routing, peer discovery
│   ├── rollup/                 # Layer-2 optimistic/ZK rollup sequencer and L1 bridge anchor
│   ├── rpc/                    # Web3 JSON-RPC 2.0 server (oen_* routes & shielded methods)
│   ├── shielded/               # Dual-pool privacy (Note, Nullifier, Merkle tree, Turnstile, Viewing Key)
│   ├── state/                  # StateDB, block execution (ApplyBlock), turnstile state enforcement
│   ├── storage/                # Embedded LevelDB/BadgerDB persistence engine
│   ├── sync/                   # Initial Block Download (IBD) and batch chain synchronizer
│   ├── upgrade/                # Self-updating binary downloader, scheduler, and rollback manager
│   ├── vm/                     # Wazero pure-Go WebAssembly VM engine and gas meter
│   └── web/                    # REST API provider (/api/status) with CORS support
└── test/
    └── bdd/                    # Ginkgo v2 & Gomega end-to-end BDD integration test suites
```

---

## 13. Building, Running & CLI Guide

### Prerequisites
- **Go 1.22+** (tested and optimized on Go 1.26+)
- Pure Go: No C compiler required (`CGO_ENABLED=0`).
- Cross-Platform: Fully compatible with Windows, Linux, and macOS.

---

### A. Compilation

Build both the full node daemon and client CLI:

```powershell
# In Windows PowerShell:
go build -o oenexa-node.exe ./cmd/node
go build -o oenexa-cli.exe ./cmd/oenexa-cli

# In Linux / macOS / WSL:
go build -o oenexa-node ./cmd/node
go build -o oenexa-cli ./cmd/oenexa-cli
```

---

### B. Running a Full Node (`oenexa-node`)

Start the node in local devnet/standalone mode:

```powershell
# Windows:
.\oenexa-node.exe start

# Linux / macOS:
./oenexa-node start

# Or run directly via Go:
go run ./cmd/node start
```

> [!NOTE]
> On Windows, running `node start` will launch the JavaScript Node.js interpreter if present in your `PATH`. Always invoke `.\oenexa-node.exe start` or `go run ./cmd/node start`.

The node initializes its state at `~/.oenexa`, starts block production every 2.0 seconds, and serves:
- **Web3 JSON-RPC 2.0**: `http://127.0.0.1:8545`
- **Prometheus Metrics**: `http://127.0.0.1:9100/metrics`
- **Health Check**: `http://127.0.0.1:9100/healthz`
- **Interactive Web Portal**: `http://127.0.0.1:8545`

---

### C. High-Speed Client Wallet CLI (`oenexa-cli`)

`oenexa-cli` provides instant, scriptable wallet generation and transaction execution:

```powershell
# 1. Generate a new ML-DSA-65 Post-Quantum Wallet:
.\oenexa-cli.exe keygen

# 2. Check the balance of any address:
.\oenexa-cli.exe balance 0x<address_hex> --rpc http://127.0.0.1:8545

# 3. Send a signed transaction in OEN:
.\oenexa-cli.exe transfer <private_key_hex> <recipient_address_hex> 10.5 --rpc http://127.0.0.1:8545

# 4. Inspect node and chain sync info:
.\oenexa-cli.exe info --rpc http://127.0.0.1:8545
```

---

### D. Shielded Pool CLI Operations

Manage privacy features directly via `oenexa-node`:

```powershell
# Create an encrypted keystore:
.\oenexa-node.exe wallet new

# Display active wallet address:
.\oenexa-node.exe wallet show

# Shield transparent OEN into a private note:
.\oenexa-node.exe wallet shield --amount 1000000000 --zaddr 0x...

# Unshield confidential funds back to a transparent account:
.\oenexa-node.exe wallet unshield --amount 1000000000 --to 0x...

# Export read-only viewing key for audit compliance:
.\oenexa-node.exe wallet export-viewing-key

# Query real-time turnstile supply invariant:
.\oenexa-node.exe query turnstile
```

---

### E. Decoupled Frontend Integration (`oenexa-frontend`)

To maintain clean separation of concerns, high consensus performance, and zero bloat, the web dashboard, block explorer, and wallet portal are decoupled into the independent [`oenexa-frontend`](https://github.com/oenexa/oenexa-frontend) repository.

External frontends and dApps connect to the running node via:
- **Web3 JSON-RPC 2.0**: `http://127.0.0.1:8545` (blocks, transactions, balances, shielded methods)
- **REST Status API**: `http://127.0.0.1:8545/api/status` (lightweight node telemetry with CORS enabled)

To run the decoupled React/Vite frontend against a local node:
```bash
git clone https://github.com/oenexa/oenexa-frontend.git
cd oenexa-frontend
npm install
npm run dev
```

---

## 14. Web3 JSON-RPC 2.0 API Specification

The OENEXA node exposes standard and privacy-enhanced JSON-RPC 2.0 methods over HTTP/WebSockets on port `8545`:

| Method | Parameters | Description |
| :--- | :--- | :--- |
| `oen_blockNumber` | `[]` | Returns the current chain block height. |
| `oen_getBalance` | `[address, "latest"]` | Returns transparent balance in nano-OEN. |
| `oen_getTransactionCount` | `[address, "latest"]` | Returns current nonce for an account. |
| `oen_sendRawTransaction` | `[signedTxHex]` | Ingests, validates, and broadcasts an ML-DSA-65 signed transaction. |
| `oen_getBlockByNumber` | `[height, fullTxBool]` | Returns block header, transactions, and state root. |
| `oen_chainInfo` | `[]` | Returns node status, peer count, genesis hash, and sync state. |
| `oen_getShieldedBalance` | `[viewingKey]` | Queries decrypted shielded note value for a given viewing key. |
| `oen_getTurnstileStatus` | `[]` | Returns transparent supply, shielded supply, and total conservation check. |
| `oen_estimateGas` | `[txObject]` | Calculates gas consumed for WASM contract execution or transfer. |

---

## 15. Testing, Benchmarking & Quality Assurance

OENEXA enforces a zero-compromise quality standard. **All 46 core packages and BDD suites maintain verified 100.0% statement test coverage**:

```bash
# Run all unit tests across the entire repository
go test ./...

# Run targeted subsystems with coverage report
go test -cover ./internal/consensus ./internal/contracts/cortex ./internal/contracts/dcommerce ./internal/mpt ./internal/shielded

# Run Behavior-Driven Development (BDD) end-to-end scenarios
go test -v ./test/bdd
```

Refer to [`TASK_RECORD.md`](TASK_RECORD.md) for the complete verified coverage audit breakdown.

---

## 16. Educational Curriculum & Student Roadmap

OENEXA is intentionally organized as an open educational curriculum. Students, academics, and systems engineers can study the design and working principles of next-generation distributed systems:

- **Consensus & State Verification**: Read [`DEVELOPMENT_PROCESS.md`](DEVELOPMENT_PROCESS.md) for the PBFT voting textbook and equivocation slashing walk-through.
- **Architectural & Design Notes**: Explore [`developer_notes/`](developer_notes/README.md) for deep dives into zero-allocation state mapping, memory safety, and PQC key encapsulation.
- **Master Whitepaper (v2.0)**: Review [`OENEXA_Whitepaper.md`](OENEXA_Whitepaper.md) for the unified vision of AI Data Center Grids, SaaS Subscriptions, and Everyday D-Commerce.
- **Runbooks & Quickstart**: Review [`DEVELOPER_QUICKSTART.md`](DEVELOPER_QUICKSTART.md) and [`DEVELOPMENT.md`](DEVELOPMENT.md).

---

## 17. License

This project is licensed under the [MIT License](LICENSE).
