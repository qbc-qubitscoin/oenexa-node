# QuantumNet — Quantum Communication Layer (Phase 30)

**Version**: v1.0 | **Phase**: 30 | **Status**: ML-KEM-768 TLS research complete, QKD hardware integration in scoping

## Overview

Phase 30 prepares OENEXA for the post-quantum networking era. While Phase 1 delivered post-quantum *transaction signatures* (ML-DSA-65), Phase 30 integrates Quantum Key Distribution (QKD) and quantum-safe VPN tunnels for *network transport*, protecting against "harvest now, decrypt later" attacks on node communications.

## Components

### 1. Quantum Key Distribution (QKD)
- Integration with commercial QKD hardware (e.g., ID Quantique, Toshiba).
- QKD-derived symmetric keys are used to protect inter-node communication for high-value sovereign deployments (Phase 29).
- OENEXA node peers negotiate QKD session keys as an additional layer beneath TLS 1.3.

### 2. Post-Quantum TLS
- All node-to-node P2P connections upgraded to TLS 1.3 + NIST PQC KEM (ML-KEM-768 / CRYSTALS-Kyber).
- Protects the gossip protocol and block propagation from quantum interception.

### 3. Key Encapsulation Architecture

```go
// QuantumTunnel wraps a net.Conn with PQ-KEM key exchange.
type QuantumTunnel struct {
    Conn     net.Conn
    KEMKey   [32]byte  // ML-KEM-768 shared secret
    QKDKey   []byte    // QKD hardware key (if available via API)
}
```

## Threat Model Mitigation
Current classical TLS relies on RSA or ECC for key exchange, which is vulnerable to Shor's algorithm on a future quantum computer. By upgrading to ML-KEM-768, OENEXA guarantees that intercepted network traffic cannot be decrypted retroactively.
