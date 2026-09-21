# ISO 20022 Financial Messaging (Phase 20)

**Document Version**: v1.0  
**Phase**: Phase 20  
**Category**: Interbank Standards & Financial Protocol Interoperability  
**Status**: Specification complete. Parser module in development.  

---

## 1. Overview

**Phase 20** embeds native **ISO 20022** structured financial messaging directly into OENEXA transactions. ISO 20022 is the universal messaging standard adopted by major global payment systems, including SWIFT, SEPA, FedNow, CHIPS, and Eurosystem TARGET2.

By providing native on-chain encoding, validation, and execution of ISO 20022 messages (MX format), OENEXA bridges the gap between legacy institutional financial infrastructure and quantum-safe blockchain ledgers without requiring fragile, custodial, off-chain translation gateways.

---

## 2. Architecture

```
+---------------------------------------------------------------------------------+
|                       ISO 20022 On-Chain Integration                            |
+---------------------------------------------------------------------------------+
                                         |
         +-------------------------------+-------------------------------+
         |                                                               |
         v                                                               v
+---------------------------------+             +---------------------------------+
|   OENEXA Transaction Payload    |             |      Traditional Banking        |
|  (Includes ISO20022Payload)     | <=========> |   (SWIFT MX, SEPA, FedNow)      |
+---------------------------------+             +---------------------------------+
         |
         v
+---------------------------------+
|     OenexaVM Parser Module      |
|  (Zero-alloc on-chain decoder)  |
+---------------------------------+
         |
         +-------------------------------+-------------------------------+
         |                                                               |
         v                                                               v
+---------------------------------+             +---------------------------------+
|    Phase 19 CBDC Settlement     |             |    Real-Time Audit Streams      |
| (Automatic Metadata Attachment) |             |  (Regulatory Observer Nodes)    |
+---------------------------------+             +---------------------------------+
```

### Architectural Principles
- **Extended Transaction Fields**: The core OENEXA `Transaction` struct is extended to support an optional byte payload:
  ```go
  type Transaction struct {
      // Standard L1 fields
      Nonce     uint64
      To        [32]byte
      Value     uint64
      GasLimit  uint64
      GasPrice  uint64
      Data      []byte
      Signature []byte // ML-DSA-65 signature

      // Phase 20 ISO 20022 extension
      ISO20022Payload []byte // Compressed XML/ASN.1 serialized MX message
  }
  ```
- **On-Chain Parsing & Validation**: OenexaVM incorporates an optimized, deterministic parser written in Go and exposed to WASM contracts via syscalls. Contracts can extract remittance data, creditor/debtor IBANs, and purpose codes directly during transaction execution.
- **Native CBDC Integration**: All transactions executed within the Phase 19 CBDC framework automatically bind ISO 20022 metadata, guaranteeing that central bank settlements maintain identical record-keeping semantics across classical banking and blockchain books.
- **Automated Compliance & Reporting**: Regulatory nodes and compliance custodians can subscribe to an ISO 20022 event stream via the node RPC to capture standardized audit trails for tax compliance, fraud detection, and anti-money laundering analytics.

---

## 3. Supported Message Types

OENEXA prioritizes the core interbank payments and reporting catalogue:

| Message Definition | Type Identifier | Description & Primary Function |
|--------------------|-----------------|--------------------------------|
| **pacs.008** | Financial Institution Customer Credit Transfer | Core payment message routing funds between corporate/retail bank accounts |
| **pacs.009** | Financial Institution Credit Transfer | Interbank high-value treasury settlement and institutional transfers |
| **pain.001** | Customer Credit Transfer Initiation | Corporate payment initiation from enterprise ERPs directly to the blockchain |
| **camt.053** | Bank to Customer Statement | End-of-day electronic account statements and reconciliation records |

---

## 4. Performance & Storage Optimization

Because standard ISO 20022 XML payloads can be verbose ($5\text{ KB} - 50\text{ KB}$), OENEXA applies two primary optimizations:
1. **Canonical Binary Encoding**: Ingested XML is parsed and converted to a compact binary ASN.1/Protobuf schema before inclusion in the transaction, reducing payload sizes by over 80%.
2. **Off-Chain Anchoring for High-Volume Statements**: Large `camt.053` statements store their full XML body in the Phase 16 Decentralized Storage Layer, maintaining only the canonical hash and summary balances on L1.

---

## 5. Status & Next Steps

- **Specification Status**: Complete.
- **Current Development**: The high-efficiency WASM ISO 20022 parser module is actively in development.
- **Roadmap Integration**: Pairs directly with Phase 19 (CBDC Bridge) and Phase 21 (Institutional Custody).
