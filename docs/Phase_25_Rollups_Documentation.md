Title: Layer-2 Rollups & Scalability (Phase 25)
Version: v1.0 | Phase: 25

## Overview
OENEXA aims to process over 100,000 Transactions Per Second (TPS). To achieve this, we have architected a **Layer-2 Rollup** engine in `internal/rollup`.

Instead of burdening the Layer-1 (L1) blockchain with every single transaction, transactions are routed to an off-chain **Sequencer**. The Sequencer acts as a high-speed execution environment, batching thousands of transfers into a single mathematical payload.

## Architecture

### 1. The L2 Sequencer (`sequencer.go`)
- Maintains an in-memory `Mempool` of Layer-2 transactions.
- Re-uses OENEXA's cryptographic `state.DB` to execute state transitions identically to the L1 chain.
- Periodically calls `BuildBatch()` to bundle all transactions, calculate the new L2 State Root (`PostRoot`), and emit a cryptographic proof.

### 2. The L1 Bridge (`l1_bridge.go`)
- Acts as the main chain's anchor for the Layer-2 network.
- Accepts `Batch` payloads from the Sequencer.
- Instead of re-executing all 10,000 transactions, the Bridge simply verifies the ZK-STARK (Zero-Knowledge) proof provided by the Sequencer.
- Because OENEXA is quantum-safe, these STARK proofs rely heavily on hash-based mechanics which resist quantum cryptanalysis.

### 3. The Batch Payload (`batch.go`)
Each batch contains:
- `BatchID`: Sequential identifier.
- `PreRoot`: The L2 State Root prior to the transactions.
- `PostRoot`: The resulting L2 State Root.
- `ZkStarkProof`: The cryptographic guarantee that the state transition from PreRoot to PostRoot is mathematically valid.

## Integration Testing
The full lifecycle is proven in `rollup_integration_test.go`:
1. Users generate quantum-safe (ML-DSA-65) keypairs and sign transactions.
2. The Sequencer accepts, validates, and executes them off-chain.
3. The L1 Bridge securely verifies the ZK-STARK proof and updates the canonical state.
