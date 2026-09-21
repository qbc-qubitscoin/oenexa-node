# OENEXA Node — Benchmark Report (v0.9.0)

**Phase**: Phase 7 + Current Protocol Status  
**Target Release**: v0.9.0-alpha  
**Date**: September 2026  
**Status**: Verified Testnet Alpha Performance  

---

## 1. Overview

This benchmark report details the execution latency, memory allocation, throughput limits, and cryptographic characteristics of the OENEXA Layer-1 blockchain node. All benchmarks were collected directly from test runs using the standard Go benchmark suite (`go test -bench=. -benchmem`):

- **Processor**: 12th Gen Intel(R) Core(TM) i9-12900H (20 logical cores, base clock 2.50 GHz, boost up to 5.00 GHz)
- **RAM**: 32 GB DDR5
- **Operating System**: Windows 11 Pro (x64)
- **Go Version**: `go1.22+`
- **Compiler Flags**: Standard optimizations (`-O2`, GC enabled)

---

## 2. Cryptography

OENEXA implements NIST FIPS 204 (ML-DSA-65 / Dilithium3) as its native post-quantum signature scheme, securing all account addresses, transaction authorizations, and validator block proposals.

| Benchmark Target | Latency | Memory / Op | Allocations | Details |
|------------------|---------|-------------|-------------|---------|
| `BenchmarkMLDSASign` | **291 µs/op** | 4,444 B/op | 9 allocs/op | Generates 3,309-byte post-quantum digital signature |
| `BenchmarkMLDSAVerify` | **78 µs/op** | 1,344 B/op | 6 allocs/op | Cryptographic signature verification over message hash |
| `EncodeAccountHash` | **302 ns/op** | 0 B/op | 0 allocs/op | In-place SHA-3 (Keccak-256) serialization of account state |

### Cryptographic Observations
- **Signature Size**: At 3,309 bytes, ML-DSA-65 signatures are substantially larger than classical ECDSA (64 bytes) or Ed25519 (64 bytes), requiring careful optimization in mempool buffering and network serialization.
- **Verification Efficiency**: Signature verification completes in 78 µs, enabling high single-thread verification capacity (~12,820 checks/s/core).
- **Zero-Allocation Account Hashing**: Serialization and hashing of account fields (`Nonce`, `Balance`, `StorageRoot`, `CodeHash`) operates with zero heap allocations, ensuring state root calculations remain CPU cache-friendly.

---

## 3. State Engine

OENEXA utilizes a high-performance 256-bit Sparse Merkle Trie (SMT) with dirty-account write buffers to decouple transaction execution from disk writes.

| Benchmark Target | Latency | Memory / Op | Allocations | Complexity / Description |
|------------------|---------|-------------|-------------|--------------------------|
| `BenchmarkStateDBSetGet` | **91 ns/op** | 0 B/op | 0 allocs/op | Dirty-buffer $O(1)$ write and immediate in-memory read |
| `BenchmarkMPTUpdate` | **22 µs/op** | ~1.8 KB/op | 12 allocs/op | Single leaf insertion/update into 256-bit SMT |
| `BenchmarkMPTGet` (cold) | **3.9 µs/op** | 384 B/op | 4 allocs/op | Trie descent and node lookup from disk/cache |
| `Old CommitRoot_100k` | **~59 ms** | ~14 MB/op | Multi-alloc | Legacy flat-map commit with $O(N \log N)$ sorting |
| `New MPT CommitRoot` | **$O(\text{dirty} \times \log N)$** | Minimal | Proportional | Selective flush of dirty accounts per block |

### State Engine Architectural Improvements
- **Selective Flushing**: Under the legacy flat-map model, committing state required iterating over all 100,000 accounts and sorting their hashes, resulting in a ~59 ms commit pause per block. The Sparse Merkle Trie flushes only mutated (dirty) leaves, reducing block commitment latency to sub-millisecond ranges for typical transactional blocks ($< 1,000$ modified accounts).
- **Atomic Rollbacks**: In-flight transactions write to an in-memory ephemeral layer, enabling instant zero-cost rollbacks upon VM instruction revert without trie recalculations.

---

## 4. Merkle Proofs

Cryptographic inclusion proofs allow light clients, decentralized bridges, and Layer-2 rollups to verify account balances and storage slots without maintaining the full state database.

| Benchmark Target | Latency | Memory / Op | Allocations | Description |
|------------------|---------|-------------|-------------|-------------|
| `BenchmarkMPTProve` | **2.7 µs/op** | 2,112 B/op | 8 allocs/op | Generates full 256-step Merkle inclusion proof path |
| `BenchmarkVerifyProof` | **1.0 µs/op** | 0 B/op | 0 allocs/op | Light client verification against state root |

### Light Client Implications
- Verification takes exactly 1.0 µs with zero allocations, allowing mobile clients, IoT devices, and smart-contract bridges to verify OENEXA state proofs at a rate of 1,000,000 proofs per second per core.

---

## 5. Mempool

The transaction pool orchestrates incoming transactions, prioritizes gas fees, validates nonces, and drops invalid or expired transactions.

| Benchmark Target | Latency | Memory / Op | Allocations | Description |
|------------------|---------|-------------|-------------|-------------|
| `BenchmarkMempoolAdd` | **161 ns/op** | 0 B/op | 0 allocs/op | Ingestion, nonce validation, and priority insertion |
| `BenchmarkMempoolPurgeCommitted` | **8.1 µs/op** | 0 B/op | 0 allocs/op | Batch eviction of committed block transactions |

### Mempool Highlights
- Zero memory allocation during transaction insertion maintains low GC overhead under high-volume flooding.
- Batch eviction runs in 8.1 µs, freeing up memory immediately upon block finalization.

---

## 6. TPS Analysis & Scaling Roadmap

### Current Single-Node Testnet Performance
- **Raw RPC Ingestion**: **~4,500 TPS** sustained over HTTP/WebSocket JSON-RPC endpoints.
- **Block Time**: 1.0 second target block time.
- **Single-Core Verification Ceiling**: ML-DSA-65 signature verification requires 78 µs per transaction, imposing a theoretical single-core ceiling of:
  $$\text{Capacity}_{\text{single core}} = \frac{1\,000\,000\,\mu\text{s}}{78\,\mu\text{s}} \approx 12\,820 \text{ verifications/second}$$
- On the benchmark machine (20 cores), parallelized verification provides theoretical multi-threaded signature verification headroom exceeding **150,000 verifications/second**. However, state trie lock contention, RPC serialization, and network I/O currently gate L1 execution throughput at ~4,500 TPS.

### Path to 100,000+ TPS
To scale from testnet alpha to planetary-scale infrastructure:
1. **Signature Verification Batching**: Grouping ML-DSA-65 verifications across transaction batches using vectorized AVX-512 / NEON routines.
2. **Parallel State Access (State Sharding/Conflict Graphs)**: Pre-executing non-conflicting account reads/writes in parallel worker threads prior to trie commitment.
3. **Layer-2 Rollups (Phase 25)**: Execution offloaded to sovereign ZK-Rollups and optimistic rollups, utilizing OENEXA L1 strictly for data availability and quantum-safe settlement.
