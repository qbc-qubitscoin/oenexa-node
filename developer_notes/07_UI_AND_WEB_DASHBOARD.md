# Developer Note 07: Decoupled Web Architecture & API Gateway

## 1. Why Decouple the Web Dashboard from the Core Node

In early blockchain prototypes, developers often bundle an embedded HTML/JS dashboard into the node executable. In production Generation-6 blockchains like OENEXA, this pattern is explicitly abandoned in favor of a **Decoupled Multi-Tier Architecture**:

### Key Architectural Rationale
1. **Minimizing Attack Surface**: Embedding JavaScript bundlers, npm dependencies, and static HTML templates directly into the core consensus binary exposes validators to supply-chain vulnerabilities and parser exploits.
2. **Binary Leanliness**: A pure Go binary (`CGO_ENABLED=0`) compiles in seconds, has a tiny memory footprint, and runs deterministically across server environments without UI baggage.
3. **Independent Release Velocity**: Frontend developers can release new features, fix responsive mobile layouts, and enhance UX in the standalone [`oenexa-frontend`](https://github.com/oenexa/oenexa-frontend) repository without requiring blockchain hardforks, validator restarts, or node redeployments.
4. **Clean API Abstraction**: Forcing all visual clients to interact strictly via JSON-RPC 2.0 and REST ensures that any external tool (third-party wallets, mobile apps, block explorers) enjoys the exact same first-class access as the official web dashboard.

---

## 2. Core Node Web API Provider (`internal/web`)

The `internal/web` package provides a lightweight, pure-Go HTTP server designed to report live node health and telemetry to external frontends.

### The `/api/status` Endpoint
The `/api/status` endpoint provides real-time node operational data:

```json
{
  "version": "v0.9.0",
  "network": "mainnet",
  "chain_height": 1420,
  "peer_count": 8,
  "pending_tx": 3,
  "state_root": "0x3f8a...",
  "timestamp": "2026-09-22T09:30:00Z",
  "online": true
}
```

### CORS & Browser Integration
To allow browsers running the decoupled frontend (e.g. at `http://localhost:5173` or a production web domain) to connect seamlessly to local or remote nodes without cross-origin blocks, `internal/web` automatically injects CORS headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, HEAD, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`
- `Cache-Control: no-store`

---

## 3. Web3 JSON-RPC 2.0 Gateway (`internal/rpc`)

While `internal/web` provides high-level health telemetry via `/api/status`, `internal/rpc` provides the primary Web3 state manipulation interface on port `8545`:

```
┌──────────────────────────────────────────────────────────────┐
│       Decoupled Frontend Application (oenexa-frontend)       │
└──────────────────────────────┬───────────────────────────────┘
                               │
            ┌──────────────────┴──────────────────┐
            │ (HTTP POST / JSON-RPC 2.0)          │ (HTTP GET)
            ▼                                     ▼
┌──────────────────────────────┐       ┌───────────────────────┐
│     internal/rpc (:8545)     │       │ internal/web (:8545)  │
│  - oen_blockNumber           │       │  - /api/status        │
│  - oen_getBalance            │       │  - CORS Middleware    │
│  - oen_sendRawTransaction    │       │  - Health Telemetry   │
│  - oen_getShieldedBalance    │       └───────────────────────┘
│  - oen_getTurnstileStatus    │
└──────────────────────────────┘
```

---

## 4. The Decoupled Frontend Ecosystem (`oenexa-frontend`)

The official web dashboard lives in its own dedicated repository: [`github.com/oenexa/oenexa-frontend`](https://github.com/oenexa/oenexa-frontend).

### Frontend Stack
- **Framework**: React 18 + TypeScript 5 + Vite.
- **Client Library**: Typed JSON-RPC 2.0 client (`rpcClient.ts`) communicating over HTTP/WebSockets.
- **Features**:
  - Real-time block explorer and transaction inspector.
  - NIST ML-DSA-65 post-quantum wallet interface.
  - Shielded pool deposit/withdrawal manager and viewing key exporter.
  - OenexaSwap AMM calculator and GreenDAO quadratic voting.
  - CarbonX ESG carbon offset registry integration.

---

## 5. Automated Testing & Verification

The `internal/web` package maintains **100.0% statement test coverage**:

```bash
go test -v ./internal/web
```

The test suite validates:
- Standard `GET` request handling and JSON payload correctness.
- `HEAD` request handling without response body.
- HTTP `405 Method Not Allowed` on invalid verbs (e.g., `POST` to `/api/status`).
- Injection and verification of CORS response headers.
- Custom node state delivery via `ServeWithInfo`.
