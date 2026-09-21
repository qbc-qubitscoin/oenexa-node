# Sovereign Infrastructure (Phase 29)

**Version**: v1.0 | **Phase**: 29 | **Status**: Sovereign deployment guide in development, two government pilots scoped

## Overview

Phase 29 delivers a sovereign deployment mode for OENEXA — allowing nation-states and large institutions to run a private or consortium OENEXA chain with full data sovereignty, while maintaining optional bridge connectivity to the public OENEXA mainnet.

## Deployment Modes

| Mode | Description | Security |
|------|-------------|----------|
| **Public Mainnet** | Fully decentralised, permissionless (default Phase 24) | High, economically secured |
| **Consortium** | Permissioned validators, shared KYB (Know Your Business) | Trust-based, BFT |
| **Sovereign** | Nation-state operated, optional air-gapped deployment | Cryptographic, state-level |

## Sovereign Features

- **Custom Chain Parameters**: Block time, gas limits, and validator set are centrally controlled by the sovereign entity.
- **Integrated National Digital Identity**: Bridges OenexaID directly to government eID databases.
- **Air-Gapped Deployment**: Support for isolated networks with offline transaction signing using HSMs.
- **Cross-Chain Bridge**: Secure, selective data disclosure bridge to the public mainnet.
- **Hardware Security**: HSM-backed validator keys (FIPS 140-3 Level 3) mandated for consensus nodes.

## Use Cases

- National CBDC (Phase 19) networks on sovereign chains
- Government procurement and bidding ledgers
- National land registries utilizing OenexaVM for smart property titles
- Central bank inter-institutional settlement networks
