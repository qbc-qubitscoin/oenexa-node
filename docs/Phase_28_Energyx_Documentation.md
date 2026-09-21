# EnergyX — Energy Trading Platform (Phase 28)

**Version**: v1.0 | **Phase**: 28 | **Status**: Oracle data model complete, smart meter vendor integrations in scoping

## Overview

EnergyX is OENEXA's peer-to-peer energy trading platform. It enables prosumers (solar panel owners, wind generators) to sell excess renewable energy directly to buyers on a transparent, automated blockchain market, bypassing centralized utilities.

## Architecture

### Smart Meter Oracle
- Phase 11 Oracle nodes pull kWh production and consumption data from IoT smart meters via secure APIs.
- Signatures on IoT data are verified on-chain to prevent spoofing.

### Energy Credit Token
- 1 Energy Credit = 1 kWh of verified renewable energy.
- Tokens are minted dynamically based on oracle feed data.

### Automated Settlement
- Smart contracts automatically transfer Energy Credits from seller to buyer, alongside the OEN payment.
- Runs every 15-minute settlement period (standard European energy market interval).

### Grid Integration
- Energy Credit transfers trigger real-world grid instructions via DSO (Distribution System Operator) APIs.
- OenexaVM logic balances local grid demand before routing long-distance.

## Market Structure

```go
type EnergyListing struct {
    ProducerID  [32]byte   // ML-DSA-65 verified prosumer (OenexaID)
    KWhAvail    uint64
    PricePerKWh uint64     // OEN wei
    ExpiryBlock uint64
    GridZone    string     // e.g., "DE-TNG" (TenneT Germany)
}
```

## ESG Link (Phase 5)
Energy Credit purchases directly contribute to the buyers' ESG score (Phase 5) and CarbonX offset portfolio (Phase 26), creating a closed-loop sustainability economy.
