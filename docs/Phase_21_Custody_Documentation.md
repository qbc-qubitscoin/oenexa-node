# Institutional Custody Layer (Phase 21)

**Version**: v1.0 | **Phase**: 21 | **Status**: Architecture complete, HSM vendor evaluation in progress

## Overview

Phase 21 provides an institutional-grade custody framework for OENEXA assets — enabling regulated custodians, exchanges, and asset managers to hold OEN and tokenized RWA assets with the security controls required by financial regulations (MiCA Article 70, SEC Rule 17f-2, Basel III operational risk frameworks).

## Architecture

### Multi-Signature Custody Contracts

All custody operations use N-of-M threshold schemes:

```go
// CustodyPolicy defines the multi-sig rules for a custody account.
type CustodyPolicy struct {
    Threshold      uint8      // M required signers
    Signers        [][32]byte // N authorised ML-DSA-65 public key hashes
    TimelockBlocks uint64     // blocks to wait before large withdrawals execute
    LargeThreshold uint64     // OEN wei amount above which timelock applies
    AuditOracleID  [32]byte   // Phase 11 oracle receiving compliance reports
}
```

- Keys never leave HSMs (Hardware Security Modules — FIPS 140-3 Level 3)
- Geographic key splitting: shards distributed across 3+ jurisdictions
- Biometric authentication required for human co-signers
- No single point of failure: any (M-1) key compromise cannot authorise a transfer

### Time-Locked Withdrawals

Large withdrawals (`amount > LargeThreshold`) enter a timelock queue:
1. Withdrawal request signed by M signers → queued on-chain
2. 48-hour observation window (configurable per institution)
3. Compliance officer human review during window
4. Automatic execution at `requestBlock + TimelockBlocks` if not cancelled

### Insurance Module

- On-chain insurance pool funded by 0.1% custody fee on all deposits
- Pool held in a DAO-governed smart contract (Phase 10 GreenDAO)
- Payout triggered by on-chain proof of loss (double-spend, smart contract exploit)
- Maximum single payout: 10% of pool (prevents pool depletion)

### Audit Trail

Every custody operation emits an immutable on-chain event:
- ISO 20022 `camt.053` fields embedded (Phase 20)
- ML-DSA-65 signed by all M authorising signers
- Retrievable via the OENEXA Merkle proof API for third-party auditors

## Regulatory Compliance

| Regulation | Coverage |
|------------|----------|
| MiCA Article 70 | CASP custody obligations — time-lock, segregation, insurance |
| EU DORA | Operational resilience — multi-region HSM distribution |
| SEC Rule 17f-2 | Qualified custodian requirements for US-regulated assets |
| FATF R.15 | VASP controls — all custody accounts require OenexaID KYC |

## HSM Integration

- PKCS#11 interface for key operations
- Supported: Thales Luna, Utimaco, YubiHSM 2 (development)
- Key ceremony: Shamir's Secret Sharing adapted for ML-DSA-65 (4,032-byte private keys)
- Remote attestation via TPM 2.0 for HSM authenticity verification
