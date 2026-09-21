# D-Commerce Escrow Platform (Phase 12)

**Document Version**: v2.0  
**Phase**: Phase 12  
**Category**: Decentralized Commerce & Smart Escrow  
**Status**: Architecture complete. WASM contract deployed locally.

---

## 1. Overview

**The Oenexa D-Commerce Escrow** represents a major shift away from extractive Web2 delivery apps and e-commerce platforms. By leveraging the OenexaVM and post-quantum smart contracts, users can now buy and sell goods, or order food delivery, directly over the blockchain with zero intermediary fees.

This phase implements a **QR-Code Mediated Zero-Trust Escrow**, ensuring that funds are cryptographically locked until physical delivery is undeniably verified in the real world.

---

## 2. Architecture & Transaction Flow

```
+----------------+          +-------------------+          +----------------+
|     Buyer      |          |  Oenexa Escrow    |          |   Restaurant   |
| (Mobile Wallet)|          | (Smart Contract)  |          | (Merchant App) |
+----------------+          +-------------------+          +----------------+
        |                             |                             |
        | 1. CreateOrder(OEN)         |                             |
        |---------------------------->|                             |
        | (Generates Dropoff QR)      |                             |
        |                             | 2. Notify Restaurant        |
        |                             |---------------------------->|
        |                             |                             |
        |                             | 3. AcceptOrder()            |
        |                             |<----------------------------|
        |                             | (Generates Pickup QR)       |
```

### The Delivery Flow
1. **Order Creation**: The Buyer calls `CreateOrder` on the contract, escrowing the total cost (Food + Delivery Fee). They generate a `DropoffQRHash` which is stored on-chain.
2. **Acceptance**: The Restaurant accepts the order by generating a `PickupQRHash`.
3. **Pickup**: A decentralized Courier accepts the delivery job (`AssignCourier`). When the Courier arrives at the restaurant, they scan the physical Pickup QR code. The QR code contains the plaintext secret, which the Courier sends to the `ConfirmPickup` function. The smart contract hashes it and verifies it matches the `PickupQRHash`.
4. **Delivery**: The Courier arrives at the Buyer's location. The Buyer presents their Dropoff QR code. The Courier scans it, obtaining the dropoff secret, and calls `ConfirmDelivery`.
5. **Instant Settlement**: The smart contract verifies the hash and immediately releases the reserved OEN, splitting the payout to the Restaurant (Food Cost) and the Courier (Delivery Commission).

---

## 3. Code Implementation

The Escrow contract is built in pure Go and executes within the OenexaVM. 

**Core Struct:**
```go
type DeliveryOrder struct {
	OrderID       string
	Buyer         [32]byte
	Restaurant    [32]byte
	Courier       [32]byte
	FoodAmount    uint64
	DeliveryFee   uint64
	PickupQRHash  string 
	DropoffQRHash string 
	State         OrderState
}
```

The contract guarantees atomic, zero-trust state transitions:
`CREATED` $\rightarrow$ `ACCEPTED` $\rightarrow$ `PICKED_UP` $\rightarrow$ `DELIVERED`

---

## 4. Security & Advantages

- **Zero-Trust Physical Handoff**: By utilizing cryptographically hashed secrets embedded in QR codes, the blockchain perfectly mirrors physical custody transfer without requiring trusted third parties.
- **Zero Extractive Fees**: Unlike traditional platforms that take up to 30%, Oenexa charges only the sub-cent network gas fee for the transaction.
- **Micro-Merchant Empowerment**: Empowers street vendors, independent restaurants, and retail stores to instantly accept digital payments with immediate settlement and no chargeback fraud.
