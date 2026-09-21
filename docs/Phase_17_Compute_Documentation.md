# Decentralized Compute Network (Phase 17)

**Document Version**: v1.0  
**Phase**: Phase 17  
**Category**: Off-Chain Verifiable Compute & Distributed Processing  
**Status**: Architecture defined. Depends on Phase 16 (Storage) for program distribution.  

---

## 1. Overview

**Phase 17** introduces a high-throughput, verifiable off-chain compute network for OENEXA. Heavy computations — such as complex mathematical simulations, machine learning training, and high-frequency risk modeling — are cost-prohibitive to run directly on-chain within block gas limits.

The Decentralized Compute Network allows clients to submit task specifications on-chain, while compute providers execute the jobs off-chain in deterministic sandboxes. Computational integrity is guaranteed cryptographically before escrowed OEN rewards are disbursed.

---

## 2. Architecture

```
+---------------------------------------------------------------------------------+
|                       Decentralized Compute Lifecycle                           |
+---------------------------------------------------------------------------------+
                                         |
     1. Submit Task                      |                 2. Fetch Code & Inputs
+----------------------+                 |             +----------------------------+
|   Client Contract    |                 |             |  Phase 16 Storage Nodes    |
| (ComputeTask Escrow) |                 |             | (WASM Binary + Input Data) |
+----------------------+                 |             +----------------------------+
           |                             |                           |
           v                             v                           v
+---------------------------------------------------------------------------------+
|                  Compute Provider Node (wazero WASM Sandbox)                    |
|             Executes deterministic task with strictly bounded gas               |
+---------------------------------------------------------------------------------+
                                         |
                                         v
                         3. Submit Result + Execution Proof
                                         |
           +-----------------------------+-----------------------------+
           |                                                           |
           v                                                           v
+----------------------+                                     +----------------------+
|  Optimistic Window   |                                     |  Zero-Knowledge (ZK) |
| (7-day dispute bond) |                                     | (Phase 25 SNARK/STARK|
+----------------------+                                     +----------------------+
           |                                                           |
           +-----------------------------+-----------------------------+
                                         |
                                         v
                       4. Escrow Settled & OEN Paid to Node
```

### Architectural Components
- **Task Submission**: A user or smart contract initiates a job by submitting a `ComputeTask` transaction to the compute coordinator contract, locking the agreed-upon reward in escrow:
  ```go
  type ComputeTask struct {
      ProgramHash [32]byte // Content hash of the WASM binary (Phase 16)
      InputHash   [32]byte // Hash of task input data
      MaxGas      uint64   // Deterministic instruction quota
      Reward      uint64   // OEN bounty held in escrow
  }
  ```
- **Off-Chain Execution**: Registered compute nodes listen for task events, retrieve the compiled WASM binary and input payload from Phase 16 storage, and execute the program inside an isolated `wazero` runtime.
- **Verification Protocol**:
  - *Optimistic Mode*: Default mode featuring a 7-day challenge window. Compute providers bond OEN collateral; invalid results can be challenged by any verifier via interactive bisection games.
  - *ZK-Proof Mode*: Providers generate succinct cryptographic proofs of correct execution, enabling instant on-chain settlement without dispute windows (integrating with Phase 25 rollups).
- **Escrow Settlement**: Once the verification criteria are satisfied (challenge window expires or ZK-proof verifies), the escrow contract atomically releases the OEN reward to the compute provider and slashes any failed bonds.

---

## 3. Integration with OenexaVM

Because OENEXA's L1 smart contract engine is built natively upon the pure-Go zero-dependency WebAssembly runtime `wazero`, the compute network shares identical execution mechanics:
- **Binary Compatibility**: Developers write compute tasks in Go, Rust, or C/C++ and compile directly to `wasm32-wasi`.
- **Read-Only State Introspection**: Compute tasks can securely query read-only snapshots of the OENEXA global Sparse Merkle Trie (SMT), allowing off-chain computations to read live account balances, oracle feeds, or token states with cryptographic proof validation.
- **Deterministic Metering**: Gas measurement in the compute sandbox mirrors L1 gas accounting, ensuring consistent termination conditions across all worker nodes.

---

## 4. Economic Security & Slashing

- **Provider Collateral**: Compute nodes must stake a minimum threshold of OEN tokens to participate in job routing.
- **Dispute Slashing**: In the event an optimistic challenge proves a provider submitted fraudulent or non-deterministic outputs, 100% of the provider's bond is slashed (50% awarded to the challenger, 50% burned).
- **Timeouts & Refunds**: If a claimed task is not completed before the specified deadline, the job is reassigned and the delinquent provider is penalized.

---

## 5. Status & Dependencies

- **Specification Status**: Complete architecture defined.
- **Dependencies**: Depends on Phase 16 (Decentralized Storage) for distributing WASM executables and datasets.
- **Upcoming Phases**: Serves as the computational backbone for Phase 18 (AI Model Marketplace).
