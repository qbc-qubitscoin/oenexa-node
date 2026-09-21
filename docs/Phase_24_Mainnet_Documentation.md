# Mainnet Launch (Phase 24)

**Version**: v1.0 | **Phase**: 24 | **Status**: Planned after Phase 23 (All prior phases must be complete)

## Overview

Phase 24 marks the full production mainnet launch of the OENEXA network — transitioning from Testnet Alpha (Phase 7) through successive testnets to a multi-validator, globally distributed mainnet securing real economic value.

## Launch Checklist

- [ ] Security audit by two independent firms (Phases 7 Bug Bounty + Phase 24 final audit)
- [ ] Genesis block configuration: initial validator set, genesis OEN distribution, chain parameters
- [ ] Minimum validator count: 21 validators across 3+ continents
- [ ] QPoS+ consensus with hardware attestation (TPM) for validators
- [ ] Emergency governance council configured (Phase 10 GreenDAO multi-sig)
- [ ] Node monitoring stack: Prometheus + Grafana dashboards
- [ ] Public RPC endpoints in 4+ regions
- [ ] Wallet release (Phase 8) on iOS/Android

## Chain Parameters (Mainnet)

```yaml
ChainID:           1 (OEN mainnet)
BlockInterval:     2 seconds
MaxBlockGas:       30,000,000
BaseFeeInitial:    1,000,000,000 wei (1 gwei)
MaxValidators:     100
QPoS_MinStake:     100,000 OEN
SlashingFraction:  0.05 (5%)
```

## Security & Economics
- Validator reward mechanisms activated
- EIP-1559 base fee burning goes live
- BFT tolerance: Up to 1/3 of validators can be malicious without breaking safety

## Post-Launch Operations
- Genesis distribution to Phase 1-23 contributors and foundation
- Transition to community governance via GreenDAO
- Decentralized onboarding of new validators via staking contracts
