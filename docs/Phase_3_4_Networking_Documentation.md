Title: Node Networking & P2P Engine (Phase 3 & 4)
Version: v2.0 (libp2p upgraded) | Phase: 3 & 4

## Overview
Phase 3 & 4 establishes the fundamental network topology of the OENEXA blockchain. Nodes must seamlessly discover each other, form a resilient mesh, and broadcast transactions and blocks globally.

We have fully replaced our original custom TCP engine with an enterprise-grade `go-libp2p` implementation.

## Architecture

### 1. The Libp2p Host (`internal/p2p/node.go`)
- **Transport & Security:** The node automatically binds to an available TCP port. It derives a stable `Ed25519` libp2p identity from the node's long-term `ML-DSA-65` private key hash, ensuring peer IDs remain consistent across restarts.
- **Multiplexing:** Utilizing standard libp2p multiplexers (Yamux/Mplex), OENEXA nodes can handle hundreds of concurrent streams over a single TCP connection.

### 2. Peer Discovery
- **Local Network (mDNS):** Nodes leverage `mDNS` to instantly discover and connect to other OENEXA nodes on the local network without relying on static bootstrap nodes. This enables zero-configuration localized clusters.
- **Global Network (DHT):** For WAN networks, standard `Kademlia DHT` routing can easily map external peer IDs.

### 3. GossipSub Routing
Instead of our naive custom gossip algorithm, we now use `pubsub.GossipSub` — the same robust message dissemination protocol used by Ethereum 2.0 and Filecoin. 

We subscribe to four primary topics:
- `/oenexa/tx/1.0.0` - Propagates pending mempool transactions.
- `/oenexa/block/1.0.0` - Propagates newly mined blocks from validators.
- `/oenexa/syncreq/1.0.0` - Nodes falling behind broadcast a catch-up request.
- `/oenexa/syncres/1.0.0` - Healthy nodes respond with chunks of historical blocks.

### 4. Integration with the Full Node (`internal/node/node.go`)
- The P2P node runs autonomously in the background.
- It provides non-blocking event hooks (`OnTxReceived`, `OnBlockReceived`, `OnSyncReq`, `OnSyncRes`) back to the main OENEXA orchestrator.
- When an incoming transaction is gossiped, the `OnTxReceived` hook pushes it directly to the `mempool` for validation.

## Security Considerations
While OENEXA transactions rely on Post-Quantum `ML-DSA-65` signatures, the underlying transport currently uses standard libp2p Noise/TLS. Because transaction signatures and state roots are cryptographically secure *above* the transport layer, intercepting or decrypting the P2P traffic yields no systemic vulnerabilities to quantum adversaries — the data inside the Gossip payloads remains perfectly unforgeable.
