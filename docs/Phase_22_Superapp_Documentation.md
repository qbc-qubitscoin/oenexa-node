# OENEXA SuperApp (Phase 22)

**Version**: v1.0 | **Phase**: 22 | **Status**: UX wireframes complete, React Native scaffold in development

## Overview

Phase 22 delivers the OENEXA SuperApp — a single mobile application combining wallet, DeFi, identity, payments, storage, AI services, and encrypted messaging into one quantum-safe financial platform. Available on iOS, Android, and as a progressive web app.

## App Modules

| Module | Features | Depends On |
|--------|----------|------------|
| **Wallet** | OEN + RWA token management, transaction history, address book | Phase 8 |
| **Pay** | QR-code OEN payments, CBDC wires, ISO 20022 bank transfers | Phase 8, 19, 20 |
| **DeFi** | DEX trading, liquidity provision, lending, yield vaults | Phase 12, 13 |
| **Identity** | OenexaID VC management, ZK proof generation, DID resolution | Phase 9 |
| **Invest** | RWA tokens, HydroChain assets, accreditation onboarding flow | Phase 14, 15 |
| **Carbon** | ESG score dashboard, carbon credit purchase and retirement | Phase 5, 26 |
| **AI** | AI Marketplace model browser, inference request submission | Phase 18 |
| **Messages** | End-to-end encrypted P2P messaging using ML-DSA-65 key agreement | Phase 8 |
| **Storage** | Personal encrypted file vault, document sharing | Phase 16 |

## Technical Architecture

### Frontend Stack
- **Framework**: React Native + TypeScript
- **State management**: Zustand + React Query
- **Navigation**: React Navigation 6 (stack + tab + drawer)
- **UI**: Custom design system (dark theme, OEN green `#39d353` accent)

### Cryptographic Core
- Core crypto compiled to WebAssembly from `internal/crypto`
- ML-DSA-65 key generation, signing, and verification run client-side
- Private keys stored in platform Secure Enclave (iOS) / Android Keystore
- Biometric authentication (Face ID, Touch ID, Fingerprint) gates all signing operations

### Node Communication
- JSON-RPC 2.0 over WebSocket for real-time subscriptions (new blocks, mempool events)
- HTTPS REST fallback for simple queries
- Phase 11 Oracle data displayed via dedicated feed connection

### Key Management

```
Key Generation:
  ML-DSA-65 keypair generated in Secure Enclave / Android Keystore
  \u2192 Private key NEVER leaves hardware
  \u2192 Public key exported \u2192 SHA-3 address derived

Signing:
  Transaction bytes \u2192 Biometric prompt \u2192 Hardware signs \u2192 3,309-byte ML-DSA-65 signature
  \u2192 Broadcast via JSON-RPC
```

## UX Design Principles

1. **Progressive disclosure**: advanced DeFi features hidden until user unlocks them
2. **Quantum-aware**: prominently display PQC security badge on all signing screens
3. **Offline capable**: wallet can generate and sign transactions without internet; broadcast later
4. **Accessibility**: WCAG 2.2 AA compliant, VoiceOver/TalkBack support
5. **Multi-wallet**: multiple accounts per app instance, hardware wallet passthrough (Phase 8)

## Distribution

| Platform | Distribution | Launch Target |
|----------|-------------|---------------|
| iOS | App Store | Phase 22 milestone |
| Android | Google Play + APK sideload | Phase 22 milestone |
| Web (PWA) | Progressive Web App via HTTPS | Phase 22 milestone |
| Desktop | Electron wrapper (macOS, Windows, Linux) | Phase 23 |
