# Global Expansion & Localisation (Phase 23)

**Version**: v1.0 | **Phase**: 23 | **Status**: Regulatory mapping complete, VARA and MAS engagements in progress

## Overview

Phase 23 prepares OENEXA for global regulatory compliance and multi-language/multi-currency support, enabling compliant deployment across diverse jurisdictions with varying legal frameworks for digital assets and financial services.

## Regulatory Coverage

| Region | Framework | OENEXA Approach | Status |
|--------|-----------|-----------------|--------|
| **European Union** | MiCA (Markets in Crypto-Assets Regulation) | Full compliance: CASP registration, custody (Phase 21), whitepaper | Architecture complete |
| **United States** | SEC / CFTC framework | RWA tokens as securities (Phase 15); OEN as commodity; no US retail CBDC | Legal review ongoing |
| **UAE** | VARA (Virtual Assets Regulatory Authority) | VARA VASP licence; Phase 19 CBDC pilot with UAE Central Bank | Active engagement |
| **Singapore** | MAS (Monetary Authority of Singapore) | Major Payment Institution licence; DPT service provider | Architecture complete |
| **United Kingdom** | FCA Digital Asset Framework | Cryptoasset business registration; stablecoin framework | In review |
| **Switzerland** | FINMA DLT Act | DLT trading facility licence; Swiss Foundation for GreenDAO (Phase 10) | Complete |
| **Japan** | FSA PVTA | Crypto exchange registration; JVCEA membership | Planning |

## Compliance Engine

The compliance engine is implemented as an OenexaVM WASM contract that:
1. Reads the buyer's OenexaID jurisdiction from their VC (Phase 9)
2. Checks the asset's `AllowedJurisdictions` list
3. Applies the applicable rule set (MiCA, SEC, VARA, etc.)
4. Returns `allow` or `deny` + reason code

**Hot-updatable**: Rules are updatable via Phase 10 GreenDAO governance vote without node restart.

## Localisation

### Language Support
- 30+ languages via i18n in the SuperApp (Phase 22)
- Right-to-left (RTL) support: Arabic, Hebrew, Farsi
- Date/number/currency formatting per locale

### Multi-Currency Fiat Off-Ramps

| Region | Currency | Partner Type |
|--------|----------|--------------|
| EU | EUR | Licensed EMI (Electronic Money Institution) |
| UK | GBP | FCA-registered payment institution |
| UAE | AED | CBUAE-licensed exchange |
| Singapore | SGD | MAS-licensed DPT service provider |
| USA | USD | FinCEN-registered MSB |

### FATF Travel Rule Compliance
- Automatic Travel Rule enforcement for transfers > 1,000 OEN / CBDC equivalent
- Sender and receiver institution details embedded in ISO 20022 `pacs.008` (Phase 20)
- IVMS 101 data standard for inter-VASP messaging

## Phase 27 Dependency
Phase 27 (Government & Partner Integration) extends Phase 23 by adding formal government MoUs, BIS connectivity, and enterprise ERP integrations.
