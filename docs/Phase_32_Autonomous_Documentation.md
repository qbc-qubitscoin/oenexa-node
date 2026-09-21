# Autonomous Agent Economy (Phase 32)

**Version**: v1.0 | **Phase**: 32 | **Status**: Research phase (Agent policy framework specification in progress)

## Overview

Phase 32 is the capstone phase of the OENEXA roadmap — enabling autonomous AI agents to participate as first-class economic actors on the network. Agents hold wallets, sign transactions with ML-DSA-65 keys, earn OEN, and interact with all prior infrastructure without human oversight.

## Agent Framework

- **Agent Identity**: Each autonomous agent holds an OenexaID DID (Phase 9) and ML-DSA-65 keypair, provisioned at deployment time.
- **On-chain Action Space**: Agents can call any OenexaVM contract — executing DeFi trades (Phase 12), purchasing ESG credits (Phase 26), or calling AI inference APIs (Phase 18).
- **Economic Incentives**: Agents earn OEN by providing automated services (compute provision in Phase 17, storage in Phase 16, or oracle data in Phase 11).

## Safety Architecture

To prevent runaway agent behaviour, smart contracts enforce strict agent policies:

```yaml
AgentPolicy:
  MaxSpendPerBlock:  1000000000000000000 # 1 OEN ceiling per block
  AllowedContracts:
    - 0xDeFiRouterAddress...
    - 0xESGMarketAddress...
  RequireHumanApproval: true             # For transfers > 100 OEN
  ExpiryBlock:       8500000             # Auto-shutdown mechanism
```

- **Human Override**: Emergency governance (Phase 10 GreenDAO multi-sig) can pause, throttle, or terminate rogue agent contracts.

## AI + Blockchain Convergence

This phase represents the full convergence of AI, quantum-safe cryptography, and decentralised finance:
1. Agents use **Phase 18 AI Marketplace** for intelligence and decision-making.
2. **Phase 17 Compute** provides verifiable execution of their tasks.
3. **Phase 16 Storage** holds the agent's long-term memory and model weights.
4. **Phase 25 Rollups** provide the cheap execution environment needed for high-frequency agent operations.
