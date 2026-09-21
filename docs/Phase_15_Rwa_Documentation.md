# Real-World Assets (RWA) Registry (Phase 15)

**Document Version**: v1.0  
**Phase**: Phase 15  
**Category**: Asset Tokenization & Institutional Capital Markets  
**Status**: Specification complete. Depends on Phase 14 (HydroChain) contract framework.  

---

## 1. Overview

**Phase 15** generalizes the asset tokenization infrastructure introduced in HydroChain (Phase 14) into a universal, compliant **Real-World Asset (RWA) Registry**. This framework enables issuers to tokenize physical, legal, and financial instruments onto the OENEXA blockchain, including:
- Commercial and residential real estate
- Physical commodities (precious metals, crude oil, refined fuels)
- Sovereign and corporate infrastructure debt / green bonds
- Private equity, venture debt, and syndicated loan portfolios

Every tokenized asset on OENEXA is legally enforceable, backed 1:1 by real-world property or contractual claims held within bankruptcy-remote Special Purpose Vehicles (SPVs), and safeguarded against post-quantum cryptographic vulnerabilities.

---

## 2. Architecture

```
+-----------------------------------------------------------------------------------+
|                            OENEXA RWA Architecture                                |
+-----------------------------------------------------------------------------------+
                                          |
         +--------------------------------+--------------------------------+
         |                                                                 |
         v                                                                 v
+-----------------------------+                               +-----------------------------+
|    Off-Chain SPV Wrapper    |                               |    Universal RWA Registry   |
| (Legal title, Deeds, Custody| -- SHA-3 Document Hash -----> | (On-chain Asset Ledger      |
|     Trust Indentures)       |                               |      AssetID -> Record)     |
+-----------------------------+                               +-----------------------------+
         |                                                                 |
         +--------------------------------+--------------------------------+
                                          |
         +--------------------------------+--------------------------------+
         |                                                                 |
         v                                                                 v
+-----------------------------+                               +-----------------------------+
|   Compliance Guard (DID)    |                               |  Institutional Custody      |
|  (OenexaID VC Check:        |                               |  (Phase 21 Cold Storage     |
|   ISO 3166-1 Jurisdiction)  |                               |   Audited Asset Backing)    |
+-----------------------------+                               +-----------------------------+
```

### Architectural Components
- **RWA Registry**: An on-chain state trie ledger maintaining an immutable mapping of `assetID → RWARecord`. Prior to minting on-chain fractional tokens, each asset must undergo verified legal wrapping (creation of an SPV or trust structure) whose legal title documents are permanently anchored on-chain.
- **Supported Asset Classes**:
  1. `RealEstate` (0): Commercial developments, multifamily complexes, land parcels.
  2. `Commodity` (1): Physical warehouse receipts (gold, silver, agricultural commodities) linked to Phase 11 Oracles for real-time market mark-to-market valuations.
  3. `Bond` (2): Fixed-income securities, municipal infrastructure debt, and green corporate debt.
  4. `PrivateCredit` (3): Short-term receivables and private credit facilities.
- **Compliance & Jurisdictional Filtering**: RWA transfers are strictly gated at the contract layer. Every transfer operation interrogates the recipient's OenexaID Verifiable Credential to verify regulatory eligibility according to ISO 3166-1 jurisdictional mandates (e.g., verifying US Reg D / Reg S exemptions or European MiCA/prospectus compliance).
- **Institutional Custody Integration**: Phase 21 (Custody) establishes an auditable bridge between off-chain physical asset vaults (e.g., bullion depositories, title escrow services) and on-chain representations.

---

## 3. Data Model

The canonical `RWARecord` data structure stored within the OenexaVM state trie:

```
RWARecord {
  AssetID     [32]byte   // SHA-3 of legal identifier
  AssetClass  uint8      // 0=RealEstate, 1=Commodity, 2=Bond
  Jurisdiction string    // ISO 3166-1
  TotalValue  uint64     // USD cents
  ShareCount  uint64
  LegalHash   [32]byte   // hash of off-chain legal docs
}
```

### Struct Field Descriptions
- `AssetID`: 32-byte cryptographic identifier computed as $\text{SHA-3}(\text{IssuerPrefix} \parallel \text{RegistrationNumber} \parallel \text{Nonce})$.
- `AssetClass`: Enum differentiating regulatory and pricing treatment across real estate, commodities, and bonds.
- `Jurisdiction`: Two-letter ISO 3166-1 alpha-2 code indicating the governing legal jurisdiction of the SPV.
- `TotalValue`: Total appraised valuation of the underlying asset denominated in USD cents (64-bit integer to avoid floating-point discrepancies).
- `ShareCount`: Number of fractional security tokens minted against the underlying asset.
- `LegalHash`: Merkle root / SHA-3 hash of legal documents stored securely in the Phase 16 Decentralized Storage Layer.

---

## 4. Transfer & Compliance Rules

1. **KYC/AML Attestation**: A transfer function call fails immediately unless both sender and recipient provide cryptographic proof of valid OenexaID KYC and sanctions clearance.
2. **Jurisdiction Whitelisting**: If an asset specifies `Jurisdiction = "US"`, transfers are restricted exclusively to holders of an accredited investor credential with valid US residency tags, or non-US investors operating under explicit Reg S allowances.
3. **Lockup Schedules**: Enforces programmatic lockups (e.g., Rule 144 one-year restriction on secondary market resales) stored within the token metadata.

---

## 5. Status & Roadmap

- **Specification Status**: Complete.
- **Dependencies**: Leverages the WASM contract scaffolding developed in Phase 14 (HydroChain).
- **Next Steps**:
  - Integration with Phase 16 (Decentralized Storage) for automated retrieval and legal document verification.
  - Development of standardized RWA factory templates for institutional issuers.
