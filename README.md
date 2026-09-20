# OENEXA (OEN) — Quantum-Shielded Layer-1 Blockchain
### *Open Economy, Next Generation Exchange & Assets*

A production-grade, privacy-preserving, post-quantum Layer-1 blockchain written in Go. Built on NIST Post-Quantum Cryptography (PQC) standards and featuring a **Zcash-class dual-pool transaction architecture** (Transparent + Shielded pools), quantum-resistant hash-based note commitments and nullifiers, viewing keys for selective regulatory disclosure, turnstile supply conservation guarantees, an embedded WebAssembly (WASM) virtual machine, and a modern reactive web portal.

[![Go](https://img.shields.io/badge/Go-1.26.2-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![NIST PQC](https://img.shields.io/badge/crypto-NIST%20PQC%20Only-blueviolet)](https://csrc.nist.gov/projects/post-quantum-cryptography)
[![Privacy](https://img.shields.io/badge/privacy-Dual--Pool%20Shielded-purple)](OENEXA_Whitepaper.md)
[![Coverage](https://img.shields.io/badge/coverage-100%25%20Statements-brightgreen)](TASK_RECORD.md)
[![UI](https://img.shields.io/badge/UI-React%2018%20%7C%20TypeScript%205%20%7C%20Vite-61DAFB?logo=react)](web)

---

## 1. Post-Quantum Cryptography & Privacy Stack

OENEXA eliminates classical asymmetric algorithms vulnerable to Shor's algorithm (RSA, ECDSA, Ed25519, secp256k1, and pairing-based curves). All operations conform strictly to NIST Post-Quantum Cryptography standards and quantum-safe cryptographic constructions.

| Primitive | Standard / Specification | Output / Size | Usage |
|-----------|--------------------------|---------------|-------|
| **ML-DSA-65** | FIPS 204 (Dilithium) | Public Key: 1,952 B<br>Signature: 3,309 B | Transparent transactions (`t-addr`), block signing, validator consensus votes, governance |
| **ML-KEM-768** | FIPS 203 (Kyber) | Ciphertext: 1,088 B<br>Shared Secret: 32 B | Ephemeral quantum-resistant P2P key exchange and shielded note payload encryption |
| **SHA-3-256** | FIPS 202 | 32-byte digest | Cryptographic hashing, address derivation, Merkle trees, note commitments & nullifiers |
| **AES-256-GCM** | NIST SP 800-38D | 32-byte key, 12-byte IV | P2P transport encryption and confidential note encryption |
| **Argon2id** | RFC 9106 | 64 MB memory, 3 iterations | Keystore key derivation function (KDF) for private spending keys at rest |

---

## 2. Dual-Pool Architecture & Turnstile Model

OENEXA integrates a two-tier transaction model combining public auditability with mathematically verifiable confidentiality:

```
┌─────────────────────────┐               Turnstile Transfer               ┌─────────────────────────┐
│    Transparent Pool     │ ─────────────── TxShield ────────────────────► │      Shielded Pool      │
│   (t-addr / ML-DSA-65)  │                                                │   (z-addr / Note Tree)  │
│  Public amounts & flows │ ◄────────────── TxUnshield ─────────────────── │ Hidden amounts & actors │
└─────────────────────────┘                                                └─────────────────────────┘
             │                                                                          │
             └─────────────────────── Total Supply Invariant ───────────────────────────┘
                           TotalSupply = TransparentSupply + ShieldedSupply
```

### A. Transparent Pool (`t-addr`)
- Publicly auditable balances and transfers signed via quantum-resistant **ML-DSA-65**.
- Suitable for exchange settlement, institutional compliance, and smart contract execution.

### B. Shielded Pool (`z-addr`)
- Cryptographically hides sender, receiver, and transaction amounts.
- **Note Commitments**: Quantum-safe cryptographic commitments stored in an incremental Merkle accumulator tree of depth 32.
- **Nullifiers**: Unique, deterministic nullifiers revealed upon spending to strictly prevent double-spending without exposing which note was spent.

### C. Turnstile Supply-Conservation Invariant
- Moving funds between pools (`TxShield` and `TxUnshield`) is strictly tracked on-chain.
- The state transition engine enforces `TotalSupply == TransparentSupply + ShieldedSupply` at every block, guaranteeing **zero stealth inflation**.

### D. Selective Disclosure & Compliance
- **Viewing Keys**: Users can export viewing keys (`vk_oen_...`) to provide auditors, tax authorities, or compliance officers with read-only visibility into incoming transfers without exposing private spending keys.
- **Payment Disclosure Proofs**: Cryptographic proofs allowing senders to prove specific payment details without revealing their wider shielded history.

---

## 3. Currency & Tokenomics

| Parameter | Specification |
|-----------|---------------|
| **Asset Symbol** | **OEN** (*Open Economy, Next Generation Exchange & Assets*) |
| **Atomic Unit** | **nano-OEN** (1 OEN = 1,000,000,000 nano-OEN; 9 decimal places) |
| **Hard Cap** | **100,000,000 OEN** |
| **Genesis Allocation** | 10,000,000 OEN (10% allocated across core ecosystem funds) |
| **Block Mining Subsidy** | 90,000,000 OEN emitted over 32 halving eras |
| **Era 0 Reward** | 45.000000000 OEN per block |
| **Block Time Target** | 2.0 seconds |
| **Block Gas Limit** | 30,000,000 gas |
| **Fee Model** | EIP-1559 dynamic base fee (burned) + priority tip (awarded to block proposer) |

---

## 4. Repository Structure

```
oenexa/
├── cmd/
│   ├── node/          # Full node entry point (CLI: oenexa-node start / wallet / tx / query)
│   ├── loadtest/      # High-throughput load testing and benchmarking tool
│   ├── multisig/      # Standalone CLI tool for quantum-safe M-of-N multisig operations
│   └── oenexaid/       # CLI utility for decentralized identity credentials (DID/VC)
├── configs/           # Network configurations (mainnet.toml, testnet.toml, oenexa-node.service)
├── docs/              # Comprehensive architectural designs, whitepapers, and phase specifications
├── internal/
│   ├── shielded/      # Quantum-resistant shielded pool (Note, Nullifier, Merkle Tree, Turnstile, Viewing Key)
│   ├── config/        # TOML configuration parser and validator
│   ├── consensus/     # BFT consensus engine, validator rotation, proposer selection
│   ├── contracts/     # Native and WASM smart contract modules:
│   │   ├── dex/       # OenSwap automated market maker (AMM) pools & confidential liquidity
│   │   ├── rwa/       # Real-World Asset (RWA) tokenization with transparent & shielded allocation
│   │   ├── aimarket/  # AI model marketplace, prompt monetization & inference settlements
│   │   ├── autonomous/# Autonomous AI agent execution, tasks & governance
│   │   ├── carbonx/   # Carbon credit registry, verifications & tokenization
│   │   ├── cbdc/      # Central Bank Digital Currency cross-border settlement rail
│   │   ├── compute/   # Decentralized cloud compute task matching and settlement
│   │   ├── custody/   # Institutional multi-tier custody, timelocks & emergency freezes
│   │   ├── energyx/   # Peer-to-peer renewable energy trading & grid settlement
│   │   ├── esgmarket/ # ESG credit marketplace & compliance validation
│   │   ├── esgnetwork/# Global ESG tracking, auditor sign-offs & emissions ledger
│   │   ├── global/    # Global multi-region jurisdiction routing & state anchors
│   │   ├── govpartner/# Government & institutional partner identity and permissions
│   │   ├── greendao/  # DAO governance, token voting, timelocks & treasury grants
│   │   ├── hydrochain/# Green hydrogen supply chain lifecycle tracking
│   │   ├── iso20022/  # ISO 20022 financial messaging translation (pacs.008, pain.001)
│   │   ├── lending/   # Decentralized collateralized lending & borrowing markets
│   │   ├── mainnet/   # Mainnet deployment parameters, staking rules & genesis allocations
│   │   ├── multisig/  # Quantum-resistant M-of-N multi-signature smart contract
│   │   ├── oracle/    # Decentralized price feeds, cross-chain attestation & quorum feeds
│   │   ├── quantumnet/# Quantum network simulation, oenexa teleportation & entanglement
│   │   ├── oenexaid/   # Decentralized Identity (DID) & Verifiable Credentials (VC)
│   │   ├── rollups/   # Layer-2 optimistic & validity rollup state anchors
│   │   ├── sovereign/ # Sovereign wealth fund reserves & allocation management
│   │   ├── storage/   # Decentralized storage contracts with storage proofs
│   │   └── superapp/  # SuperApp multi-service aggregation contract
│   ├── core/          # Block, Transaction, Genesis, Merkle tree, Tokenomics, EIP-1559 fee model
│   ├── crypto/        # ML-DSA-65 key management, ML-KEM-768 key encapsulation, SHA-3-256
│   ├── identity/      # W3C-compliant DID and Verifiable Credential primitives
│   ├── keystore/      # Encrypted disk keystores using Argon2id + AES-256-GCM
│   ├── mempool/       # Gas-price priority queue with sender account throttling
│   ├── metrics/       # Prometheus telemetry metrics and /healthz endpoint
│   ├── node/          # Full node lifecycle coordinator wiring all subsystems
│   ├── oracle/        # Oracle node backend service
│   ├── p2p/           # Encrypted P2P network, KEM handshake, gossip routing, peer discovery
│   ├── rpc/           # JSON-RPC 2.0 server (oen_* routes, batch queries, shielded methods)
│   ├── state/         # StateDB with shielded pool tracking, block processor (ApplyBlock), gas metering
│   ├── storage/       # Embedded LevelDB database persistence for blocks and states
│   ├── sync/          # Initial Block Download (IBD) and batch chain synchronizer
│   ├── upgrade/       # Self-updating binary downloader, verification, and on-chain scheduler
│   ├── vm/            # WebAssembly virtual machine powered by wazero (pure-Go, zero CGo)
│   └── web/           # Embedded HTTP web server serving JSON-RPC and the static UI bundle
├── test/
│   └── bdd/           # Behavior-Driven Development (BDD) end-to-end scenario suites
└── web/               # Standalone React 18 + TypeScript 5 + Vite web dashboard
```

---

## 5. Building & Running

### Prerequisites
- **Go 1.22+** (tested and verified on Go 1.26+)
- **Node.js 18+ & npm** (only required if building or editing the web frontend)

### A. Build the Blockchain Node Executable
```powershell
# In Windows PowerShell:
go build -o oenexa-node.exe ./cmd/node

# In Linux / macOS:
go build -o oenexa-node ./cmd/node
```

### B. Start the Full Node
```powershell
# In Windows PowerShell:
.\oenexa-node.exe start

# In Linux / macOS:
./oenexa-node start

# Or run directly via Go without pre-compiling:
go run ./cmd/node start
```

> [!NOTE]
> Running `node start` on Windows will invoke the Node.js JavaScript interpreter from your `PATH`. Always use `.\oenexa-node.exe start` or `go run ./cmd/node start`.

The node starts up, initializes the state database at `~/.oenexa`, begins producing blocks every 2 seconds, and serves:
- **JSON-RPC 2.0 Endpoint**: `http://127.0.0.1:8545`
- **Prometheus Metrics & Healthz**: `http://127.0.0.1:9100/metrics`, `/healthz`
- **Interactive Web Portal**: `http://127.0.0.1:8545`

### C. CLI Wallet & Shielded Commands
```powershell
# Create a new ML-DSA-65 post-quantum keypair:
.\oenexa-node.exe wallet new

# Show current wallet addresses:
.\oenexa-node.exe wallet show

# Shield funds from transparent account into a confidential note:
.\oenexa-node.exe wallet shield --amount 1000000000 --zaddr 0x...

# Unshield funds back into a transparent address:
.\oenexa-node.exe wallet unshield --amount 1000000000 --to 0x...

# Export read-only viewing key for audit disclosure:
.\oenexa-node.exe wallet export-viewing-key

# Query on-chain turnstile invariant & pool balance status:
.\oenexa-node.exe query turnstile
```

### D. Run the Web Dashboard (Frontend Development)
The frontend dashboard is built with React 18, TypeScript 5, and Vite:
```bash
cd web
npm install
npm run dev
```

To build production static assets that are embedded into the Go binary:
```bash
cd web
npm run build
```

---

## 6. Testing & Quality Assurance

OENEXA enforces strict quality and correctness requirements. **Every internal package and test suite has achieved 100.0% statement coverage**:

```bash
# Run all unit and package tests
go test ./...

# Run targeted package with coverage report
go test -cover ./internal/shielded ./internal/state ./internal/core

# Run BDD integration test scenarios
go test -v ./test/bdd

# Run frontend unit tests (Vitest)
cd web && npm test -- --run
```

Refer to [`TASK_RECORD.md`](TASK_RECORD.md) for the verified 100% coverage audit breakdown across all packages.

---

## 7. License

Distributed under the [MIT License](LICENSE).
