# OENEXA Development Process & Engineering Curriculum

This document outlines the evolutionary development process used to implement the 32 Phases of the OENEXA ecosystem, completely aligned with the Master Whitepaper (Version 2.0). It also serves as an **educational curriculum and architectural textbook** for students, researchers, and engineers studying the design and working principles of a high-performance, post-quantum blockchain organization.

---

## Part I: 32-Phase Roadmap & Current Status

### Stage 1: Core Foundation & Quantum Security (Phases 1 - 5)
1. **Cryptography & Primitives (Phase 1 & 2)**: Implementation of NIST FIPS 204 ML-DSA-65 (digital signatures) and NIST FIPS 203 ML-KEM-768 (key encapsulation) for post-quantum security.
2. **State & Mempool (Phase 3)**: 256-bit Sparse Merkle Trie (SMT), dirty-buffer state management, account balance/nonce model, and priority-fee mempool with EIP-1559 dynamic base fee burning.
3. **P2P Networking (Phase 4)**: Global peer discovery via mDNS and Kademlia DHT, pubsub topic gossiping (`/oenexa/tx/1.0.0`, `/oenexa/block/1.0.0`, `/oenexa/vote/1.0.0`) using `go-libp2p`.
4. **OenexaVM (Phase 5A)**: Pure-Go WebAssembly (WASM) execution engine (`wazero`) with strict gas metering, host-call state access, and memory sandboxing.
5. **PBFT Consensus & Slashing (Phase 5B)**: Practical Byzantine Fault Tolerant consensus engine with 2-phase voting (Prevote, Precommit), leader rotation, full state transition verification, and automatic equivocation (double-voting) slashing.

### Stage 2: Tooling & Utilities (Phases 6 - 8)
1. **Oenexa CLI & Node**: Multi-command wallet (`oenexa-cli`) supporting ML-DSA-65 key generation, balance inspection, and JSON-RPC transfers.
2. **Testnet Tools & Benchmarks**: High-load benchmarking suite (4,500 TPS testnet alpha), signature batch verification, and cryptographic fuzzing.
3. **Smart Contract Tooling**: Bytecode deployment utilities and gas profilers for the OenexaVM.

### Stage 3: Real-World Escrows & D-Commerce (Phases 9 - 13)
1. **OenexaID (Phase 9)**: W3C-compliant Decentralized Identifier (DID) and Verifiable Credential (VC) registry for merchant accreditation and privacy-preserving KYC.
2. **Oracle Network (Phase 10)**: Decentralized median-aggregated telemetry feeds for real-world environmental and pricing data.
3. **DeFi Hub (Phase 11)**: OenexaSwap automated market maker (AMM) and over-collateralized lending protocols.
4. **D-Commerce Escrow (Phase 12)**: Zero-trust, QR-code mediated physical delivery escrow smart contracts (`internal/contracts/dcommerce`) enabling zero-fee food delivery and open market trade.
5. **Micro-Merchant Gateway (Phase 13)**: Point-of-Sale (POS) mobile interfaces for street vendors and retail stores to instantly accept OEN with zero chargebacks.

### Stage 4: Infrastructure, AI, and SaaS (Phases 14 - 21)
1. **Oenexa Cortex (Phase 14)**: Green Data Center tokenized infrastructure bonds and automated yield distribution contracts (`internal/contracts/cortex`).
2. **Compute-as-a-Service (CaaS) (Phase 15)**: On-chain leasing of decentralized GPU/TPU clusters for AI agents and LLM training paid natively in OEN.
3. **SaaS Development (Phase 16)**: Smart contract frameworks for recurring SaaS subscriptions (auto-deducted OEN) and product crowdfunding.
4. **Real World Assets (RWA) (Phase 17)**: Fractionalization of real estate, renewable assets, and traditional debt instruments.
5. **AI Marketplace (Phase 18)**: Autonomous AI agent registration, discovery, and task execution escrow.
6. **CBDC & ISO 20022 (Phases 19 - 20)**: Central bank digital currency gateways and ISO 20022 financial messaging adapters.
7. **Institutional Custody (Phase 21)**: Hardware security module (HSM) multi-party computation (MPC) cold storage.

### Stage 5: Global Expansion & Scalability (Phases 22 - 32)
1. **Super App & Global Compliance (Phases 22 - 23)**: Cross-platform user interface unifying D-Commerce, Cortex compute leasing, and wallet management.
2. **Mainnet & Layer-2 Rollups (Phases 24 - 25)**: Optimistic/ZK rollup sequencers targeting 100,000+ TPS for micro-transactions.
3. **ESG Markets & EnergyX (Phases 26 - 28)**: Peer-to-peer renewable electricity trading and automated carbon offset retirement.
4. **Sovereign Wealth Funds (Phase 29)**: Multi-jurisdictional reserves and institutional governance wrappers.
5. **Quantum Internet & Autonomous AI (Phases 30 - 32)**: Quantum-entangled communications and autonomous self-tuning protocol economics.

---

## Part II: Working Principles & Student Curriculum

This section provides future students, researchers, and core contributors with the foundational knowledge required to understand how each subsystem of the OENEXA node functions under the hood.

```
+-----------------------------------------------------------------------------+
|                          OENEXA Node Architecture                           |
+-----------------------------------------------------------------------------+
|                                                                             |
|   [ JSON-RPC / CLI ] <---------- HTTP / WebSockets ---------> [ User App ]  |
|           |                                                                 |
|           v                                                                 |
|   [ Mempool Engine ] <---- P2P Gossiping (/oenexa/tx) ----> [ Peer Nodes ]  |
|           |                                                                 |
|           v                                                                 |
|   [ Consensus Engine ] <--- Votes & Proposals (/vote) ----> [ Validators ]  |
|           |  (PBFT 2-Step Voting: Prevote & Precommit)                      |
|           v                                                                 |
|   [ State Transition (ApplyTx) ]                                            |
|       |--> VM Execution: Pure-Go WebAssembly (wazero)                       |
|       |--> Fee Model: EIP-1559 (BaseFee Burned + Validator Tip)             |
|       |--> Smart Contracts: D-Commerce Escrow, Cortex AI Grid               |
|           |                                                                 |
|           v                                                                 |
|   [ Sparse Merkle Trie (SMT) ] ---> Deterministic StateRoot                 |
|           |                                                                 |
|           v                                                                 |
|   [ Persistent BlockStore & StateStore ] ---> BadgerDB Storage Engine       |
|                                                                             |
+-----------------------------------------------------------------------------+
```

---

### 1. Post-Quantum Cryptography (PQC)
Modern blockchains rely on ECDSA or Ed25519 elliptic curve signatures, which will be completely broken by Shor's algorithm on a cryptographically relevant quantum computer. OENEXA replaces all legacy curves with **lattice-based post-quantum cryptography**:

- **Digital Signatures (ML-DSA-65 / FIPS 204)**: Formerly known as Dilithium3. Provides NIST Security Level 3 protection. Public keys are 1,952 bytes; signatures are 3,309 bytes. Every account address is computed as:
  $$\text{Address} = \text{SHA3-256}(\text{PublicKey})$$
- **Key Encapsulation (ML-KEM-768 / FIPS 203)**: Formerly known as Kyber768. Used in the P2P transport layer to establish encrypted, quantum-safe peer sessions resistant to "Harvest Now, Decrypt Later" espionage.

---

### 2. State Storage Engine: 256-Bit Sparse Merkle Trie (SMT)
The state of OENEXA records account nonces, balances, and smart contract storage. Rather than using an Ethereum-style hexary Patricia Trie (which incurs high branch-traversal overhead), OENEXA utilizes a **256-bit binary Sparse Merkle Trie**:

- **Deterministic Key Path**: The 32-byte address directly defines the 256-step branch path ($0 = \text{left}, 1 = \text{right}$).
- **Dirty-Buffer Cache**: During block execution, state updates are buffered in an in-memory dirty map with $O(1)$ read and write latency (~91 ns/op).
- **Lazy Merkleization**: Only dirty leaves are re-hashed at block commit, reducing Merkle root generation time by over 80%.
- **Cryptographic Inclusion Proofs**: Light clients and mobile devices can verify an account balance in $1.0\ \mu\text{s}$ using a 256-hash Merkle proof without downloading the full chain.

---

### 3. Practical Byzantine Fault Tolerant (PBFT) Consensus
OENEXA achieves sub-second finality using an upgraded PBFT Proof-of-Stake consensus algorithm (`internal/consensus`):

#### Step 1: Proposal (Proposer Turn)
At height $H$, the round-robin leader is selected by:
$$\text{ProposerIndex} = H \pmod{|\text{ValidatorSet}|}$$
The proposer pulls pending transactions from the mempool, computes the gas and burned fees, updates the state snapshot, signs the block header with ML-DSA-65, and broadcasts the block over the P2P network.

#### Step 2: Verification & Prevote
When peer validators receive the proposed block, they do not blindly trust it. They execute full state transition verification:
1. Validate block header signature against the designated proposer's ML-DSA-65 public key.
2. Confirm parent block hash matches the local tip: $\text{PrevHash} == \text{Chain}[H-1].\text{Hash}$.
3. Verify transaction Merkle root: $\text{ComputeMerkleRoot}(\text{Txs}) == \text{Header.MerkleRoot}$.
4. Replay transactions against a state snapshot, verifying that the gas used, burned fees, and resulting state root match the header byte-for-byte.
5. If valid, the validator signs and broadcasts a `VotePrevote`.

#### Step 3: Precommit Quorum
When a validator receives $> \frac{2}{3}$ of the total voting power in Prevotes for the block hash:
$$\sum \text{VotingPower}(\text{Voters}) > \frac{2}{3} \times \text{TotalVotingPower}$$
The validator advances its state machine to `StepPrecommit` and broadcasts a `VotePrecommit`.

#### Step 4: Commit & Finality
When $> \frac{2}{3}$ of Precommits arrive, the block is finalized:
- The snapshot state is committed to the persistent database.
- Committed transactions are purged from the mempool.
- The block is appended to the immutable chain.
- Unlike Proof-of-Work, finality is absolute—no chain reorganizations can occur.

#### Step 5: Equivocation Detection & Slashing
If a Byzantine validator signs two different block hashes for the same height, round, and step (double-voting attack), the engine detects the conflicting signatures:
- The duplicate vote is rejected.
- The malicious validator's voting power is immediately slashed in the validator set.
- 10% of their on-chain balance is burned as a financial slashing penalty.

---

### 4. Zero-Trust D-Commerce Escrow (The QR-Code Delivery Protocol)
To dismantle extractive Web2 delivery platforms (which charge up to 30% merchant commission), OENEXA implements a zero-trust physical delivery escrow (`internal/contracts/dcommerce`):

1. **Buyer Orders**: Buyer deposits OEN (food cost + courier tip) into the smart contract. The buyer's wallet generates a random secret $S_{\text{dropoff}}$ and submits its hash $H(S_{\text{dropoff}})$ on-chain.
2. **Restaurant Prepares**: Restaurant accepts the order and submits $H(S_{\text{pickup}})$ on-chain.
3. **Courier Pickup**: The courier arrives at the restaurant and scans the restaurant's physical barcode/QR code, revealing $S_{\text{pickup}}$. The courier calls `ConfirmPickup(orderID, S_{\text{pickup}})`. The contract computes $\text{SHA256}(S_{\text{pickup}})$ and verifies it matches $H(S_{\text{pickup}})$. The state transitions to `PICKED_UP`.
4. **Customer Drop-off**: The courier arrives at the buyer's home. The buyer displays their QR code, revealing $S_{\text{dropoff}}$. The courier scans it and calls `ConfirmDelivery(orderID, S_{\text{dropoff}})`.
5. **Instant Two-Party Settlement**: The contract verifies the hash and atomically releases the funds: 100% of food cost to the restaurant, and 100% of delivery fee to the courier. No middleman, zero platform cuts.

---

### 5. Oenexa Cortex: Green Data Center Infrastructure
OENEXA powers the future of decentralized AI through **Oenexa Cortex** (`internal/contracts/cortex`):

- **Fractional Infrastructure Bonds**: Physical Tier-4 data centers running on 100% renewable energy are tokenized as `CortexAsset` contracts. Users pool OEN to fund construction and receive fractional shares.
- **Compute-as-a-Service (CaaS)**: Once the data center is active, AI developers pay OEN to lease GPU/TPU compute clusters.
- **Autonomous Yield Streaming**: All compute fees accumulate in the contract's `RevenuePool`. Investors claim their proportional dividend with guaranteed mathematical fairness:
  $$\text{InvestorYield} = \frac{\text{SharesHeld}}{\text{TotalShares}} \times \text{RevenuePool} - \text{PreviouslyClaimed}$$

---

## Part III: Codebase Directory Guide

| Directory | Purpose | Key Files |
| :--- | :--- | :--- |
| `cmd/node/` | Main daemon entrypoint | `main.go` |
| `cmd/oenexa-cli/` | Multi-command wallet CLI | `main.go` |
| `internal/crypto/` | Post-quantum ML-DSA-65 & ML-KEM-768 | `crypto.go`, `wallet.go` |
| `internal/core/` | Blocks, transactions, headers, genesis | `block.go`, `transaction.go`, `genesis.go` |
| `internal/state/` | Sparse Merkle Trie & StateDB | `statedb.go`, `smt.go`, `account.go` |
| `internal/consensus/` | PBFT engine, validator set, slashing | `engine.go`, `validator.go`, `vote.go` |
| `internal/p2p/` | libp2p pubsub gossip network | `node.go`, `identity.go` |
| `internal/vm/` | Wazero pure-Go WASM runtime | `vm.go`, `gas.go` |
| `internal/contracts/dcommerce/`| Zero-trust delivery escrow | `escrow.go`, `escrow_test.go` |
| `internal/contracts/cortex/` | AI Data Center CaaS & bonds | `cortex.go`, `cortex_test.go` |
| `internal/mempool/` | EIP-1559 gas-priority pool | `mempool.go` |
| `internal/rpc/` | JSON-RPC 2.0 web API | `server.go`, `api.go` |
