# OENEXA Developer Quickstart & Runbook

Welcome to the OENEXA core repository. This project comprises a custom blockchain built in Go, featuring a post-quantum cryptographic layer (ML-DSA-65 / ML-KEM-768), an embedded WASM virtual machine (OenexaVM), dual-pool shielded privacy, and a comprehensive ecosystem of decentralized applications (Phases 1-32, Master Whitepaper v2.0).

This document outlines the step-by-step process for running the project locally for development, interacting via the CLI suite, and the strategy for moving to a production server environment.

---

## Part 1: Running Locally (Development Mode)

Local development runs a single-node "devnet" or local testnet. This allows you to deploy WASM smart contracts, test the JSON-RPC interface, and interact with the web dashboards without requiring a full peer-to-peer network.

### Step 1: Prerequisites
Ensure you have the following installed on your local machine:
1. **Go 1.22+** (tested and verified on Go 1.26+).
2. **Pure Go Environment**: No C compiler required (`CGO_ENABLED=0`).
3. **TinyGo** (optional, required to compile custom Go smart contracts into WebAssembly for OenexaVM).

---

### Step 2: Build the Core Node and Client CLI
Compile both the main blockchain daemon (`oenexa-node`) and the high-speed client wallet tool (`oenexa-cli`):

```bash
# In PowerShell / Windows:
go build -o oenexa-node.exe ./cmd/node
go build -o oenexa-cli.exe ./cmd/oenexa-cli

# In Linux / macOS / Git Bash:
go build -o oenexa-node ./cmd/node
go build -o oenexa-cli ./cmd/oenexa-cli
```

---

### Step 3: Run the Local Node
Start the node in standalone/dev mode:

```bash
# In Windows PowerShell:
.\oenexa-node.exe start

# In Linux / macOS / Git Bash:
./oenexa-node start

# Or run directly via Go without building an executable:
go run ./cmd/node start
```

*Note: The node initializes state at `~/.oenexa`, produces blocks every 2.0 seconds, serves Web3 JSON-RPC on `http://127.0.0.1:8545`, and exposes Prometheus metrics at `http://127.0.0.1:9100/metrics`.*

---

### Step 4: Interact via `oenexa-cli` Wallet

While the node is running, open a separate terminal to manage wallets and broadcast transactions:

```bash
# Generate a new ML-DSA-65 Post-Quantum keypair:
.\oenexa-cli.exe keygen

# Query balance of an account via RPC:
.\oenexa-cli.exe balance 0x<address_hex> --rpc http://127.0.0.1:8545

# Transfer OEN to another account:
.\oenexa-cli.exe transfer <private_key_hex> <to_address_hex> 5.0 --rpc http://127.0.0.1:8545

# Query node status and chain info:
.\oenexa-cli.exe info --rpc http://127.0.0.1:8545
```

---

### Step 5: Interact with the Shielded Pool (`oenexa-node`)

The node binary includes wallet management for transparent and shielded privacy transactions:

```bash
# Generate encrypted keystore:
.\oenexa-node.exe wallet new

# Deposit transparent OEN into a confidential shielded note:
.\oenexa-node.exe wallet shield --amount 1000000000 --zaddr 0x...

# Withdraw confidential shielded funds back to a transparent account:
.\oenexa-node.exe wallet unshield --amount 1000000000 --to 0x...

# Export read-only viewing key for auditing:
.\oenexa-node.exe wallet export-viewing-key

# Check turnstile supply conservation invariant:
.\oenexa-node.exe query turnstile
```

---

### Step 6: Run the Test Suites
The repository contains comprehensive unit and integration tests across all packages with verified **100% statement coverage**:

```bash
# Run all Go tests recursively
go test ./...

# Run targeted core packages
go test -cover ./internal/consensus ./internal/contracts/cortex ./internal/contracts/dcommerce ./internal/mpt

# Run Ginkgo BDD end-to-end integration tests
go test -v ./test/bdd
```

---

### Step 7: Connect the Decoupled Web Dashboard (`oenexa-frontend`)
To preserve node performance and security, the frontend dashboard, explorer, and wallet portal are decoupled into the standalone [`oenexa-frontend`](https://github.com/oenexa/oenexa-frontend) repository:

```bash
git clone https://github.com/oenexa/oenexa-frontend.git
cd oenexa-frontend
npm install
npm run dev
```

The frontend runs at `http://localhost:5173` and automatically connects to the local node's JSON-RPC 2.0 API at `http://127.0.0.1:8545` and REST telemetry at `http://127.0.0.1:8545/api/status`.

---

## Part 2: Production Server Deployment

Moving to production requires transitioning from a local devnet to a distributed Peer-to-Peer (P2P) network. 

### 1. Infrastructure Preparation
- **Servers**: Provision cloud instances (AWS EC2, Google Compute Engine, or bare metal) with at least 4 Cores, 16GB RAM, and NVMe SSDs for fast state I/O.
- **Networking**: Open port `30303` (TCP/UDP) for the libp2p network layer, and optionally port `8545` if the node is intended to be a public JSON-RPC endpoint.

### 2. Compilation and Binary Distribution
- Do not build on the production server. Use a CI/CD pipeline (e.g., GitHub Actions) to compile static Linux binaries (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`).
- Distribute the compiled `oenexa-node` binary to your server nodes.

### 3. Bootstrap Nodes (Seed Nodes)
A blockchain needs initial connection points:
1. Deploy 3 to 5 highly available "Seed Nodes" across different geographic regions.
2. Note their libp2p multiaddresses (e.g., `/ip4/<ip>/tcp/30303/p2p/<peerID>`).

### 4. Running the Mainnet Validator Node
On a production server, run the node pointing to the seed nodes and using a secure production configuration.

```bash
# Example Systemd execution command
./oenexa-node start \
  --config /etc/oenexa/mainnet.toml \
  --datadir /var/lib/oenexa
```

### 5. Process Management and Monitoring
- **Systemd/Docker**: Wrap the execution in a Systemd service file (`configs/oenexa-node.service`) or a Docker container ensuring the process auto-restarts on failure.
- **Telemetry**: Hook the node logs into a monitoring stack (Prometheus + Grafana). Monitor metric endpoints (`http://127.0.0.1:9100/metrics`) for block propagation times, ML-DSA verification latencies, and memory usage.
- **Security**: Keep validator private keys secure, ideally utilizing Hardware Security Modules (HSMs) or secure cloud enclaves (e.g., AWS KMS) via OENEXA's Phase 21 custody integrations.
