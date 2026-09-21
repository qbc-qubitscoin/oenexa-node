# Global ESG Impact Network (Phase 31)

**Version**: v1.0 | **Phase**: 31 | **Status**: Data model complete, oracle integration in design

## Overview

Phase 31 connects all OENEXA ESG infrastructure into a unified, globally visible ESG Impact Network — a public dashboard and data protocol for verified environmental and social impact metrics from blockchain-native sources.

## Data Sources

The network aggregates data from previous phases:
- **Phase 5 CarbonX**: Corporate carbon footprint, offset credits.
- **Phase 10 GreenDAO**: Governance votes on ESG proposals and fund allocation.
- **Phase 14 HydroChain**: Renewable energy production (kWh) telemetry.
- **Phase 26 ESG Market**: Carbon credit retirements and NFT burns.
- **Phase 28 EnergyX**: P2P renewable energy traded amounts.

## Network Architecture

### ESG Oracle Network
- Specialised oracle nodes aggregate ESG data from off-chain APIs and on-chain events, publishing signed attestations.

### ESG Score NFT
- Companies and wallets hold a non-transferable `ESGScoreNFT`.
- Reflects their on-chain ESG activity, updated dynamically each epoch.
- Used as a credential in DeFi (Phase 12/13) to access green yields.

### Public API & Reporting
- RESTful API for ESG data queries.
- Natively formats data for GRI, SASB, and TCFD sustainability reporting standards.
- Enables automated corporate sustainability report generation directly from the blockchain.

### ESG Derivatives
- Phase 12 DEX integration supports ESG score-linked financial products.
- Example: Interest rate swaps where the rate is tied to the counterparty's real-time ESGScoreNFT.
