# Decentralized Storage Layer (Phase 16)

**Document Version**: v1.0  
**Phase**: Phase 16  
**Category**: Decentralized Infrastructure & Content-Addressed Storage  
**Status**: Protocol specification complete.  

---

## 1. Overview

**Phase 16** introduces a native, decentralized storage protocol for the OENEXA ecosystem. Large-scale data — such as legal contracts for real-world assets (Phase 15), AI model weight files (Phase 18), institutional KYC records (Phase 9), and environmental sensor telemetry (Phase 10) — cannot be stored directly inside L1 consensus state without causing severe state bloat.

The Decentralized Storage Layer allows high-volume assets to be stored off-chain in content-addressed storage nodes while retaining zero-knowledge verifiable cryptographic anchors within OENEXA's 256-bit Sparse Merkle Trie (SMT).

---

## 2. Architecture

```
+-------------------------------------------------------------------------+
|                      Decentralized Storage Protocol                     |
+-------------------------------------------------------------------------+
                                     |
        +----------------------------+----------------------------+
        |                                                         |
        v                                                         v
+-------------------------+                               +-------------------------+
|   On-Chain State Trie   |                               |  Distributed Off-Chain  |
| (SMT Canonical Anchor   | <===== Merkle Proof Verification ===== |  Storage Nodes (IPFS / |
|   contentHash [32]byte) |                               |     Filecoin Bridge)    |
+-------------------------+                               +-------------------------+
        ^                                                         |
        |                                                         |
        +----------------------------+----------------------------+
                                     |
        +----------------------------+----------------------------+
        |                                                         |
        v                                                         v
+-------------------------+                               +-------------------------+
|  AES-256-GCM Envelope   |                               |   OEN Storage Market    |
| (Encrypted Content      |                               | (Proof-of-Storage Yield |
|  Key-gated via ML-DSA)  |                               |    Priced by Oracles)   |
+-------------------------+                               +-------------------------+
```

### Key Components
- **Storage Backend**: Built upon an IPFS-compatible, content-addressed protocol with optional Filecoin bridge integration for verifiable multi-year persistence and geographic redundancy guarantees.
- **On-Chain Cryptographic Anchoring**: OenexaVM smart contracts store only a 32-byte `contentHash` ($\text{SHA-3}$ of the raw payload). Because this hash resides in the global Sparse Merkle Trie, any party can generate an $O(\log N)$ Merkle inclusion proof verifying that a given off-chain document was committed at a specific block height.
- **Access Control via Post-Quantum Signatures**: Retrieval requests require ML-DSA-65 signed bearer tokens issued by the asset owner. Storage nodes verify signature validity against the owner's on-chain public key before serving data.
- **Envelope Encryption**: Sensitive files are encrypted client-side using AES-256-GCM with unique ephemeral keys. Ephemeral keys are wrapped using hybrid post-quantum key encapsulation mechanisms (ML-KEM-768) and managed through OenexaID decentralized key vaults.
- **OEN Storage Market**: Storage providers stake OEN tokens and advertise available capacity to the network. Nodes earn OEN per megabyte-epoch, with dynamic baseline storage pricing continuously benchmarked by the Phase 11 Oracle network.

---

## 3. Node Integration

OenexaVM smart contracts reference stored data via the canonical `StorageRef` struct:

```go
// StorageRef links an on-chain record to off-chain content
type StorageRef struct {
    ContentHash [32]byte  // SHA-3 of raw content
    SizeBytes   uint64
    MimeType    string
    Provider    [32]byte  // storage node address
}
```

### Storage Node RPC Protocol
Storage nodes implement a lightweight JSON-RPC interface that integrates with OENEXA core clients:
- `storage_pin(contentHash, payload, proof)`: Pins an encrypted blob to the local storage provider node.
- `storage_fetch(contentHash, authSignature)`: Verifies caller's ML-DSA-65 signature and streams the requested data.
- `storage_audit(contentHash)`: Issues a challenge-response proof-of-spacetime confirming the provider retains the bytes.

---

## 4. Security & Fault Tolerance

- **Garbage Collection & Slashing**: Storage nodes that fail periodic on-chain storage audit challenges lose a fraction of their bonded OEN stake.
- **Deduplication**: SHA-3 content addressing natively eliminates duplicate file storage across different contracts and users.
- **Censorship Resistance**: Files are erasure-coded across a minimum of $K$-of-$N$ storage providers, ensuring accessibility even if individual nodes go offline.

---

## 5. Status & Next Steps

- **Specification Status**: Complete.
- **Integration Dependencies**:
  - Serves as the storage foundation for Phase 17 (Decentralized Compute) and Phase 18 (AI Model Marketplace).
  - Supplies document storage for Phase 15 (RWA Registry).
