# Phase 5: Consensus Engine (PBFT / Proof-of-Stake)

## Overview
Oenexa uses a highly optimized Proof-of-Stake (PoS) consensus mechanism inspired by Practical Byzantine Fault Tolerance (PBFT) and Tendermint. The consensus engine ensures that all nodes in the decentralized network agree on the exact sequence of blocks and transactions, even in the presence of malicious actors or network partitions.

This document serves as an educational guide to how consensus works in Oenexa, the cryptographic primitives it uses, and how it achieves finality.

---

## 1. The PBFT State Machine

Instead of nodes blindly accepting the longest chain (like Proof-of-Work in Bitcoin), Oenexa validators actively vote on block proposals. A block is only committed to the blockchain when a supermajority (>2/3 of total voting power) of validators explicitly sign off on it.

The state machine for each consensus height is divided into four main steps:

### Step 1: Propose (`StepPropose`)
A deterministic leader is chosen for the current height (usually based on a round-robin schedule weighted by stake). The leader bundles transactions from the mempool, executes them against the `state.DB` (Sparse Merkle Trie), generates a new `core.Block`, signs the block, and gossips it to the network.

### Step 2: Prevote (`StepPrevote`)
Upon receiving a valid proposal, validators verify the transactions, state transitions, and the ML-DSA-65 post-quantum signature of the proposer. If everything is valid, they broadcast a `VotePrevote` message.

### Step 3: Precommit (`StepPrecommit`)
Once a validator sees `> 2/3` of the network's voting power casting a `VotePrevote` for the same block hash, it considers the block "locked." The validator then broadcasts a `VotePrecommit` message.

### Step 4: Commit (`StepCommit`)
Once `> 2/3` voting power casts a `VotePrecommit` for the same block hash, the block reaches cryptographic finality. The node writes the block to disk (`storage/blockchain.go`), applies the state changes, and begins the next height.

---

## 2. Cryptographic Voting (ML-DSA-65)

Because Oenexa is designed to be quantum-resistant, all consensus votes are signed using **ML-DSA-65** (Dilithium).

```go
type Vote struct {
	Height    uint64
	Round     uint32
	Type      VoteType // Prevote or Precommit
	BlockHash [32]byte
	Validator string   // ML-DSA-65 Public Key (Hex)
	Signature []byte   // Post-Quantum Signature
}
```

Whenever a node receives a vote via the P2P network (on the `/oenexa/vote/1.0.0` libp2p topic), it calls `pqc.VerifyMLDSASignature` before processing the vote. This ensures that a quantum computer cannot spoof a validator's vote.

---

## 3. Fast-Path for Local Dev / Single Node

In the early stages of network bootstrapping (or in local testing), the network might consist of only a single validator with 100% of the voting power. 

The consensus engine intelligently recognizes this: when a single validator proposes a block, it immediately casts a Prevote. Since its single vote constitutes 100% of the voting power (which is > 2/3), the `HasQuorum` check instantly triggers the Precommit. The Precommit vote is cast, `HasQuorum` is reached again, and the block is committed instantaneously. This provides incredible performance in single-node environments while perfectly retaining the PBFT security model for decentralized setups.

---

## 4. Where to Find the Code

Students and researchers exploring the Oenexa codebase should look at the following files to understand the Consensus Engine:

- **`internal/consensus/engine.go`**: The core PBFT state machine and the main orchestrator of rounds, steps, and vote tallying.
- **`internal/consensus/vote.go`**: Definitions for `Vote` structs and the logic to serialize/deserialize them for the network.
- **`internal/consensus/validator.go`**: The `ValidatorSet` which tracks the voting power of all active validators and calculates `HasQuorum`.
- **`internal/node/node.go`**: The integration layer where the libp2p P2P network hooks into the consensus engine (`OnVoteReceived`, `OnBlockBroadcast`, etc.).

---

## 5. Security Properties

1. **Instant Finality**: Because Oenexa requires 2/3+ Precommits before applying a block, forks are mathematically impossible (as long as < 1/3 of the network is malicious). Users do not need to wait 6 confirmations; once a block is committed, it is final forever.
2. **Post-Quantum Secure**: Unlike classical PBFT networks that use ECDSA, the Oenexa engine is fully hardened against Shor's algorithm via ML-DSA-65.
3. **Byzantine Fault Tolerance**: The network continues to operate correctly even if up to 33% of validators crash, go offline, or actively try to sabotage consensus.
