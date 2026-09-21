# OENEXA (OEN) Documentation

Welcome to the official documentation repository for **OENEXA (OEN)** — a high-performance, quantum-safe, post-quantum Layer-1 blockchain built from the ground up to protect decentralized finance, real-world assets, and sovereign infrastructure against classical and quantum computing threats.

---

## Current Status

- **Version**: `v0.9.0` (Testnet Alpha)
- **Signature Scheme**: Post-Quantum ML-DSA-65 (NIST FIPS 204 / Dilithium3 standard) providing 128-bit quantum security level
- **State Engine**: High-performance 256-bit Sparse Merkle Trie (SMT) with dirty-account buffer caching and O(1) in-memory account access
- **Throughput**: ~4,500 TPS raw RPC ingestion on single-node testnet alpha
- **Smart Contracts**: WASM runtime powered by `wazero` sandbox environment
- **Consensus**: Quantum-resistant Proof-of-Stake with verifiable random selection

---

## Table of Contents

### 1. Architecture & Design Docs
Technical specifications, system designs, benchmarks, and infrastructure architecture:
- [Testnet Alpha Benchmark Report](./Testnet_Alpha_Benchmark.md) — Comprehensive cryptographic, state engine, Merkle proof, and mempool benchmarks (v0.9.0)
- [Wallet Architecture Design](./Wallet_Architecture_Design.md) — Quantum-safe wallet core, key derivation, and transaction signing protocols
- [Oracle Design Specification](./Oracle_Design.md) — Multi-source decentralized oracle network with threshold signature validation
- [DeFi Architecture Specification](./DeFi_Architecture.md) — Automated Market Maker (AMM), liquidity pools, and lending protocol architecture
- [CarbonX ESG Framework](./CarbonX_ESG_Framework.md) — Environmental, Social, and Governance (ESG) verification and carbon credit architecture
- [CarbonX Registry Integration](./CarbonX_Registry_Integration.md) — Technical integration guide for certified external carbon registries

### 2. Security & Audit Reports
Formal verification, security audits, threat models, and vulnerability assessments:
- [DeFi Security Audit Report](./DeFi_Audit_Report.md) — Comprehensive smart contract audit of AMM, liquidity pools, and math libraries
- [Oracle Security Review](./Oracle_Security_Review.md) — Attack surface assessment, Sybil resistance, and data integrity evaluation
- [Wallet Security Audit](./Wallet_Security_Audit.md) — Key generation, memory safety, side-channel analysis, and post-quantum security audit
- [Bug Bounty Report (Alpha)](./Bug_Bounty_Report_Alpha.md) — Summary of vulnerability disclosures, bounties awarded, and testnet alpha patches

### 3. Identity & Privacy
Decentralized identity, zero-knowledge verifiable credentials, and shielded privacy rails:
- [OenexaID Schema Specification](./OenexaID_Schema.md) — W3C-compliant Decentralized Identifier (DID) and Verifiable Credential schemas
- [OenexaID Legal & Privacy Review](./OenexaID_Legal_Privacy_Review.md) — GDPR, CCPA, and global regulatory compliance analysis for decentralized identity

### 4. DeFi & Governance
Decentralized finance mechanics, tokenomics, and decentralized autonomous organization frameworks:
- [DeFi Architecture Specification](./DeFi_Architecture.md) — Core DeFi primitives, quantum-safe swap algorithms, and liquidity math
- [DeFi Security Audit Report](./DeFi_Audit_Report.md) — Risk mitigation, reentrancy guards, and arithmetic security audit
- [GreenDAO Governance Framework](./GreenDAO_Governance_Framework.md) — On-chain proposal lifecycle, quadratic voting, and treasury governance
- [GreenDAO Legal Entity Structure](./GreenDAO_Legal_Entity_Structure.md) — Decentralized legal wrapper, DAO legal personality, and jurisdictional compliance

---

## Phase Roadmap (Phases 1–32)

The 32-phase evolutionary roadmap from foundational node engineering to a fully autonomous post-quantum global economy:

### Foundations & Core Protocol (Phases 1–8)
- **Phase 1: Core Node Architecture** — Core daemon, peer-to-peer gossip network, and configuration engine.
- **Phase 2: Post-Quantum Cryptography** — Native integration of NIST FIPS 204 ML-DSA-65 digital signatures.
- **Phase 3: State Storage Engine** — 256-bit Sparse Merkle Trie (SMT) with dirty-buffer optimizations and cryptographic proofs.
- **Phase 4: Consensus Engine** — Post-quantum Proof-of-Stake consensus and block proposal protocols.
- **Phase 5: OenexaVM Execution Runtime** — Sandboxed WebAssembly (WASM) execution engine built on `wazero`.
- **Phase 6: Privacy & Shielded Transactions** — NoteCommitmentTree, shielded transfers, and zero-knowledge privacy rails.
- **Phase 7: Testnet Alpha & Performance Tuning** — Benchmarking, load testing, and bug bounty remediation.  
  📄 [Testnet Alpha Benchmark Report](./Testnet_Alpha_Benchmark.md) | [Bug Bounty Report Alpha](./Bug_Bounty_Report_Alpha.md)
- **Phase 8: Quantum-Safe Wallet Infrastructure** — Non-custodial CLI/SDK wallet with quantum-safe key management.  
  📄 [Wallet Architecture Design](./Wallet_Architecture_Design.md) | [Wallet Security Audit](./Wallet_Security_Audit.md)

### Identity, Oracles & D-Commerce (Phases 9–13)
- **Phase 9: OenexaID Decentralized Identity** — Self-sovereign DIDs and verifiable credentials for KYC/accreditation.  
  📄 [OenexaID Schema Specification](./OenexaID_Schema.md) | [OenexaID Legal & Privacy Review](./OenexaID_Legal_Privacy_Review.md)
- **Phase 10: Decentralized Oracle Network** — Verifiable multi-source data feeds with threshold signing.  
  📄 [Oracle Design Specification](./Oracle_Design.md) | [Oracle Security Review](./Oracle_Security_Review.md)
- **Phase 11: DeFi Hub (OenexaSwap)** — Constant-product AMM pools, liquidity mining, and over-collateralized lending.  
  📄 [DeFi Architecture Specification](./DeFi_Architecture.md) | [DeFi Security Audit Report](./DeFi_Audit_Report.md)
- **Phase 12: D-Commerce Escrow Platform** — Zero-trust, QR-code mediated delivery mechanics for P2P/B2P food and goods delivery.  
  📄 [D-Commerce Escrow Documentation](./Phase_12_DCommerce_Escrow_Documentation.md)
- **Phase 13: Micro-Merchant Gateway** — Fast, zero-fee Point-of-Sale integrations for mobile merchants.

### AI, Infrastructure & SaaS (Phases 14–21)
- **Phase 14: Oenexa Cortex** — Fractionalized tokenization and financing of Tier-4 green data centers.  
  📄 [Cortex Documentation](./Phase_14_Cortex_Documentation.md)
- **Phase 15: Compute-as-a-Service (CaaS)** — Tokenized leasing of decentralized GPU/TPU power for AI models.
- **Phase 16: SaaS Development Platform** — Smart contract frameworks for recurring SaaS subscriptions and product crowdfunding.
- **Phase 17: Real-World Asset (RWA) Engine** — Securities, real estate, and carbon credit tokenization on-chain.
- **Phase 18: AI Agent Marketplace** — Autonomous discovery and payment routing for specialized AI agents.
- **Phase 19: CBDC Gateway** — Institutional bridge for Central Bank Digital Currencies.
- **Phase 20: ISO 20022 Integration** — Traditional finance message parsing and clearing integration.
- **Phase 21: Institutional Custody** — HSM integration and quantum-safe multisignature cold storage.

### Global Impact & Frontier Technology (Phases 26–32)
- **Phase 26: Global ESG Marketplace** — Scaled global trading platform for verified ecological and renewable credits.  
  📄 [ESG Marketplace Documentation](./Phase_26_Esgmarket_Documentation.md)
- **Phase 27: Sovereign Regulatory Partnerships** — Standardized sovereign regulatory observation nodes and compliance reporting.  
  📄 [Government Partnership Documentation](./Phase_27_Govpartner_Documentation.md)
- **Phase 28: EnergyX Renewable Smart Grid** — IoT-connected smart-grid automated energy settlement and microgrid trading.  
  📄 [EnergyX Documentation](./Phase_28_Energyx_Documentation.md)
- **Phase 29: Sovereign Asset & National Identity Rails** — National-scale sovereign DID rollouts and sovereign bond digitization.  
  📄 [Sovereign Infrastructure Documentation](./Phase_29_Sovereign_Documentation.md)
- **Phase 30: Quantum Communication Network (QKD)** — Quantum Key Distribution hardware network interface and entanglement validation.  
  📄 [Quantum Network Documentation](./Phase_30_Quantumnet_Documentation.md)
- **Phase 31: Global Decentralized ESG Network** — Worldwide consortium verification mesh for real-time sensor environmental telemetry.  
  📄 [ESG Network Documentation](./Phase_31_Esgnetwork_Documentation.md)
- **Phase 32: Fully Autonomous Post-Quantum Economy** — Autonomous AI-driven economic coordination, self-tuning consensus, and planetary-scale settlement.  
  📄 [Autonomous Economy Documentation](./Phase_32_Autonomous_Documentation.md)
