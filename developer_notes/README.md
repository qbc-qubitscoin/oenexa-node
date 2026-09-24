# OENEXA (OEN) — Developer Notes Directory

Welcome to the **OENEXA Developer Notes** folder. This directory contains the complete technical architecture notes explaining **WHY** each system was designed and **HOW** it works under the hood.

---

## Directory Index

| File | Subsystem | Description |
|---|---|---|
| [`01_TDD_AND_TESTING_GUIDE.md`](01_TDD_AND_TESTING_GUIDE.md) | Quality & Engineering | Test-Driven Development (TDD) workflow, Ginkgo BDD specs, fuzzing, and concurrency testing |
| [`02_CORE_BLOCKCHAIN_AND_TOKENOMICS.md`](02_CORE_BLOCKCHAIN_AND_TOKENOMICS.md) | Core Blockchain | Transaction lifecycle, EIP-1559 dynamic base fee, Merkle trees, StateDB, and halving emission curve |
| [`03_POST_QUANTUM_SECURITY.md`](03_POST_QUANTUM_SECURITY.md) | Cryptography & Keystore | NIST FIPS 204 ML-DSA-65 digital signatures, ML-KEM-768 quantum key encapsulation, SHA-3-256, and Argon2id wallet encryption |
| [`04_CONSENSUS_ENGINE.md`](04_CONSENSUS_ENGINE.md) | Consensus | Deterministic round-robin proposer selection, BFT quorum voting, 2-second block intervals, and state commit pipeline |
| [`05_SMART_CONTRACTS_AND_VM.md`](05_SMART_CONTRACTS_AND_VM.md) | WebAssembly VM & DeFi | Wazero pure-Go WASM host runtime, gas metering, storage slots, OenSwap AMM ($x \cdot y = k$), Lending, CarbonX ESG, GreenDAO, and OenexaID |
| [`06_NETWORKING_AND_RPC.md`](06_NETWORKING_AND_RPC.md) | P2P & JSON-RPC | Frame wire protocol, AES-256-GCM secure connections, priority Mempool heap, and JSON-RPC 2.0 dispatch engine |
| [`07_UI_AND_WEB_DASHBOARD.md`](07_UI_AND_WEB_DASHBOARD.md) | Decoupled Web & APIs | Decoupled frontend architecture, REST telemetry provider (/api/status), CORS middleware, and Web3 JSON-RPC gateway |
| [`08_AUTO_UPGRADE_AND_SCHEDULER.md`](08_AUTO_UPGRADE_AND_SCHEDULER.md) | Upgrades & Hardforks | Autonomous quantum-resistant binary downloads, cryptographic checksum verification, and on-chain hardfork scheduler |

---

## The Rule for Developers

Before implementing any new feature, fix, or smart contract:
1. **Read the corresponding note** to understand current state invariants.
2. **Write unit tests first** in accordance with [`PROJECT_RULES.md`](../PROJECT_RULES.md).
3. **Update or add to these notes** so that the *Why* and *How* of your changes are permanently documented for future developers.
4. **Consult the curriculum** in [`DEVELOPMENT_PROCESS.md`](../DEVELOPMENT_PROCESS.md) to understand overall system architecture.
