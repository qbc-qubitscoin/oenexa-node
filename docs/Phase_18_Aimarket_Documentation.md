# AI Model Marketplace (Phase 18)

**Document Version**: v1.0  
**Phase**: Phase 18  
**Category**: AI & Decentralized Machine Learning Marketplace  
**Status**: Specification complete. Prototype inference engine in research.  

---

## 1. Overview

**Phase 18** establishes an open, decentralized marketplace for Artificial Intelligence (AI) models, fine-tuned weights, and real-time inference services on the OENEXA blockchain. 

By unifying decentralized storage (Phase 16) for neural network model weights and the decentralized compute network (Phase 17) for sandbox execution, Phase 18 provides an end-to-end framework where AI developers can monetize machine learning models, and consumers can run confidential, verifiable inference without relying on centralized cloud AI providers.

---

## 2. Architecture

```
+-------------------------------------------------------------------------------+
|                           AI Marketplace Architecture                         |
+-------------------------------------------------------------------------------+
                                        |
     1. Register Model                  |             2. Request Inference
+---------------------------+           |         +---------------------------+
|    AI Model Developer     |           |         |       Consumer Client     |
| (Mints Royalty NFT + URI) |           |         | (Locks OEN Fee in Escrow) |
+---------------------------+           |         +---------------------------+
              |                         |                       |
              v                         v                       v
+-------------------------------------------------------------------------------+
|                            On-Chain Model Registry                            |
|             {modelID, contentHash, pricePerInference, providerAddress}        |
+-------------------------------------------------------------------------------+
                                        |
             +--------------------------+--------------------------+
             |                                                     |
             v                                                     v
+---------------------------+                             +---------------------------+
|    Phase 16 Storage       |                             |     Phase 17 Compute      |
|  (Encrypted Model Weights |                             | (Sandboxed wazero Runtime |
|   ONNX / GGUF / WASM)     |                             |   Encrypted Input Stream) |
+---------------------------+                             +---------------------------+
                                        |
                                        v
+-------------------------------------------------------------------------------+
|                              Settlement Engine                                |
|          Inference Verified -> Escrow Split: Model Creator % / Node %         |
+-------------------------------------------------------------------------------+
```

### Architectural Principles
- **Model Registry Contract**: Maintains an immutable directory of published AI models on-chain. Each entry records `{modelID, contentHash, pricePerInference, providerAddress}`, ensuring transparent, tamper-proof discovery and pricing.
- **Inference Lifecycle**:
  1. *Escrow Deposit*: The consumer locks the required OEN inference fee in the marketplace escrow contract alongside a cryptographic hash of the query input.
  2. *Model Retrieval*: The assigned compute node fetches the model weights from the Phase 16 Decentralized Storage Layer.
  3. *Isolated Execution*: The compute node executes the model inside a deterministic `wazero` WASM sandbox.
  4. *Proof Submission*: The node submits the inference result and corresponding execution proof to the smart contract.
  5. *Atomic Settlement*: The escrow is unlocked, automatically distributing revenue to the model creator and compute node.
- **Client Input Privacy**: Query inputs can be encrypted using hybrid post-quantum key encapsulation mechanisms (ML-KEM) and client ML-DSA-65 keys, shielding sensitive user prompts from unauthorized observers.
- **Model Ownership & Royalty NFTs**: Registering an AI model mints an on-chain ownership NFT. This NFT confers fractionalized royalty streams whenever inferences are paid for, enabling developers to sell or collateralize intellectual property rights.

---

## 3. Supported Model Types

The marketplace supports deterministic runtime targets capable of compiling to WebAssembly:

| Model Format | Runtime Engine | Typical Use Cases |
|--------------|----------------|-------------------|
| **ONNX Models** | WebAssembly runtime compiled via Emscripten / Tract | Computer vision, classification, tabular predictions, tabular ESG scoring |
| **Custom WASM Runtimes** | Native Go / Rust compiled to `wasm32-wasi` | Tailored mathematical pipelines, financial quantitative modeling, signal filtering |
| **Quantized LLMs** | `llama.cpp` compiled to WASM (Q4_K_M, Q8_0) | Natural language processing, automated document summarization, code generation |

---

## 4. Economic Tokenomics

- **Inference Fee Split**: Every inference fee paid in OEN is split programmatically:
  - **70%** to the model owner (streamed directly to the model NFT holder)
  - **25%** to the compute node performing the execution
  - **5%** burned to reduce OEN circulating supply
- **Quality Assurance Staking**: Model creators stake OEN against their model listings. Models that produce verifiable errors or violate safety contracts risk having their stake slashed.

---

## 5. Status & Next Steps

- **Specification Status**: Complete.
- **Current Development**: Prototype WASM inference execution engine undergoing benchmarking in sandbox environments.
- **Prerequisites**: Leverages Phase 16 (Storage) for model weight hosting and Phase 17 (Compute) for distributed task orchestration.
