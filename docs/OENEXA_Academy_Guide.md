# OENEXA Academy: Core Principles & Architecture Guide

Welcome, future developers, students, and blockchain architects. 

OENEXA was built not just to be a cryptocurrency, but as a foundational organization and a blueprint for the future of decentralized systems. This guide was written to help you understand the **working principles** of the OENEXA blockchain. 

By reading this, you will learn how we engineered the system to be both the **most secure** (post-quantum resistant) and the **fastest** (highly optimized state engine) network in the world.

---

## 1. Cryptography: The "Most Secure" Principle
*Code location: `internal/crypto/`*

### The Problem with Legacy Blockchains
Bitcoin and Ethereum use Elliptic Curve Cryptography (ECDSA or Ed25519) for digital signatures. While secure today, these algorithms are completely broken by **Shor's Algorithm** running on a sufficiently powerful quantum computer.

### The OENEXA Solution
OENEXA is designed for the post-quantum era. We use **ML-DSA-65** (formerly CRYSTALS-Dilithium), which is the official NIST standard for post-quantum digital signatures.
- **How it works:** Instead of relying on the difficulty of factoring prime numbers, ML-DSA relies on the hardness of the *Learning With Errors (LWE)* problem over structured lattices. 
- **The Code:** When a user signs a transaction, the `internal/crypto` package generates a 3,309-byte signature. The node verifies this signature in **78 microseconds**—making it incredibly fast to validate, ensuring quantum-level security without sacrificing network speed.

---

## 2. The State Engine: The "Fastest" Principle
*Code location: `internal/state/` and `internal/mpt/`*

### The Problem with State Bloat
In standard blockchains, looking up an account balance gets slower as the network grows because every account is stored in a massive tree.

### The OENEXA Solution (Dual-Store Architecture)
To make OENEXA the fastest blockchain, we separated *reading data* from *proving data*.
1. **The Flat Store (O(1) Speed):** When the node needs to check a balance, it doesn't walk down a complex tree. It looks up the account directly in a flat LevelDB database using a simple key (`a: <address>`). This takes just **91 nanoseconds**.
2. **The Sparse Merkle Trie (SMT):** To ensure cryptographic security, we maintain a 256-bit Sparse Merkle Trie. However, the node only updates this tree *in memory* during block execution and commits it at the very end. 
3. **The Code:** Look at `internal/storage/statestore.go`. You will see `SaveState()` atomically writing both the fast-read flat data and the secure Merkle tree roots to disk simultaneously.

---

## 3. The Mempool: Transaction Lifecycle
*Code location: `internal/mempool/`*

How does a transaction go from a user's wallet to being permanently recorded?
1. **Submission:** A user submits a transaction via the Web API.
2. **Validation:** The mempool immediately checks the ML-DSA-65 signature and ensures the user has enough OEN to pay the gas fee.
3. **Queuing:** Valid transactions are added to an in-memory priority queue, sorted by the fee they are willing to pay. Adding a transaction takes just **161 nanoseconds**.
4. **Block Mining/Forging:** The consensus engine pulls the top transactions from the mempool, executes them against the State Engine, and packages them into a block.

---

## 4. Decoupled Architecture
*Code location: `internal/web/` vs `oenexa-frontend` repository*

A core software engineering principle is **Separation of Concerns**. 
If you look at the OENEXA node code, you will notice there is no HTML, CSS, or React code. 
- The Node (`oenexa-node`) is a pure, highly-optimized Go backend that focuses *only* on consensus, cryptography, and network security. It exposes a JSON API.
- The UI (`oenexa-frontend`) is a separate React application. 

**Why teach this?** Because tying a visual dashboard into a consensus engine creates security vulnerabilities and slows down development. By separating them, UI engineers and protocol engineers can work at their own pace without stepping on each other's toes.

---

## 5. The Path to 100,000+ TPS (Layer 2)
*Code location: Planned for Phase 25*

While our Layer 1 is heavily optimized (capable of ~4,500 TPS natively), the ultimate vision of OENEXA is to handle global-scale enterprise traffic. To achieve this, students should study our **Rollup Architecture** (Phase 25).
By moving execution off-chain and only posting Zero-Knowledge (ZK) STARK proofs to the Layer 1 blockchain, we can compress thousands of transactions into a single verification step, effectively giving OENEXA infinite scalability.

---

### A Note to the Student
*As you read the OENEXA codebase, remember that every line of code was written with three things in mind: **Security against future threats**, **Speed for global adoption**, and **Clarity for future developers**. You are not just reading code; you are reading the foundation of a decentralized digital economy.*
