# Layer-2 Rollups (Phase 25)

**Version**: v1.0 | **Phase**: 25 | **Status**: Research phase (ZK-STARK library evaluation in progress)

## Overview

Phase 25 introduces Layer-2 optimistic and ZK rollups on top of OENEXA L1. This targets 100,000+ TPS — the whitepaper goal that L1 ML-DSA-65 signatures alone cannot achieve on commodity hardware.

## Why Rollups Are Necessary

> ML-DSA-65 verification at 78 µs/op caps L1 throughput at ~12,800 verifications/second per core. Rollups batch thousands of L2 transactions into a single L1 proof, amortizing the L1 verification cost across the batch.

## Rollup Architectures

### 1. Optimistic Rollups
- Batches transactions off-chain; posts compressed state diff to L1
- Fraud proof window (7 days)
- L1 acts as data availability + dispute resolution layer
- Fraud proofs executed in OenexaVM (WASM)
- Achieves ~10-20x throughput improvement over L1

### 2. ZK Rollups (STARK-based)
- Generates STARK proofs of batch validity off-chain
- Instant finality on L1 (no challenge period)
- OENEXA-specific: uses post-quantum-friendly STARK proofs (FRI-based, no elliptic curve dependency)
- Target: 100,000 TPS with sub-second L2 finality

## Bridging (L1 ↔ L2)
- Native bridge contract for L1 ↔ L2 OEN transfers
- 7-day withdrawal delay (Optimistic) / Instant (ZK)
- RWA tokens bridgeable to L2 with Merkle proofs of L1 ownership

## Ecosystem Integration
- L2 state roots committed to L1 Sparse Merkle Trie
- DeFi (Phase 12) deployed directly to L2 for high-frequency trading
- Custody (Phase 21) operators interact primarily via L1 for large settlements
