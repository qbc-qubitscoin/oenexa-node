package dcommerce

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

// OrderState represents the current state of a D-Commerce delivery order.
type OrderState string

const (
	StateCreated   OrderState = "CREATED"
	StateAccepted  OrderState = "ACCEPTED"
	StatePickedUp  OrderState = "PICKED_UP"
	StateDelivered OrderState = "DELIVERED"
	StateDisputed  OrderState = "DISPUTED"
	StateRefunded  OrderState = "REFUNDED"
)

// DeliveryOrder represents the state of a restaurant delivery escrow contract.
type DeliveryOrder struct {
	OrderID       string             `json:"order_id"`
	Buyer         [crypto.AddressSize]byte     `json:"buyer"`
	Restaurant    [crypto.AddressSize]byte     `json:"restaurant"`
	Courier       [crypto.AddressSize]byte     `json:"courier"`
	FoodAmount    uint64             `json:"food_amount"`
	DeliveryFee   uint64                       `json:"delivery_fee"`
	PickupQRHash  string                       `json:"pickup_qr_hash"` // Hash of the barcode provided by restaurant
	DropoffQRHash string                       `json:"dropoff_qr_hash"` // Hash of the QR provided by buyer
	State         OrderState                   `json:"state"`
	CreatedAt     int64              `json:"created_at"`
	DeliveredAt   int64              `json:"delivered_at"`
}

// DeliveryEscrowContract acts as a system contract for the OenexaVM to handle P2P/B2P delivery logic.
type DeliveryEscrowContract struct {
	// In a real WASM contract, state would be read/written to the OenexaVM StateDB.
	// For this template, we mock the state storage locally.
	orders map[string]*DeliveryOrder
}

// NewDeliveryEscrowContract initializes the D-Commerce escrow system.
func NewDeliveryEscrowContract() *DeliveryEscrowContract {
	return &DeliveryEscrowContract{
		orders: make(map[string]*DeliveryOrder),
	}
}

// CreateOrder is called by the Buyer to lock funds in escrow for a food order.
func (c *DeliveryEscrowContract) CreateOrder(buyer, restaurant [crypto.AddressSize]byte, foodAmount, deliveryFee uint64, dropoffQRHash string) (*DeliveryOrder, error) {
	// In production, the VM deducts (foodAmount + deliveryFee) from the Buyer's balance and locks it in the contract.
	
	hasher := sha256.New()
	hasher.Write(buyer[:])
	hasher.Write(restaurant[:])
	timeBytes, _ := json.Marshal(time.Now().UnixNano())
	hasher.Write(timeBytes)
	orderID := string(hasher.Sum(nil)[:8]) // shortened for example

	order := &DeliveryOrder{
		OrderID:       orderID,
		Buyer:         buyer,
		Restaurant:    restaurant,
		FoodAmount:    foodAmount,
		DeliveryFee:   deliveryFee,
		DropoffQRHash: dropoffQRHash,
		State:         StateCreated,
		CreatedAt:     time.Now().Unix(),
	}

	c.orders[orderID] = order
	return order, nil
}

// AcceptOrder is called by the Restaurant to acknowledge they are preparing the food.
func (c *DeliveryEscrowContract) AcceptOrder(orderID string, restaurant [crypto.AddressSize]byte, pickupQRHash string) error {
	order, exists := c.orders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	if order.Restaurant != restaurant {
		return errors.New("unauthorized: only the designated restaurant can accept")
	}
	if order.State != StateCreated {
		return errors.New("invalid state transition: order must be CREATED")
	}

	order.PickupQRHash = pickupQRHash
	order.State = StateAccepted
	return nil
}

// AssignCourier allows a courier to accept the delivery job.
func (c *DeliveryEscrowContract) AssignCourier(orderID string, courier [crypto.AddressSize]byte) error {
	order, exists := c.orders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	if order.State != StateAccepted {
		return errors.New("invalid state transition: order must be ACCEPTED by restaurant first")
	}
	
	// Ensure courier is not the buyer or restaurant
	if courier == order.Buyer || courier == order.Restaurant {
		return errors.New("courier cannot be the buyer or the restaurant")
	}

	order.Courier = courier
	return nil
}

// ConfirmPickup is called by the Courier after scanning the Restaurant's QR code.
func (c *DeliveryEscrowContract) ConfirmPickup(orderID string, courier [crypto.AddressSize]byte, pickupSecret string) error {
	order, exists := c.orders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	if order.Courier != courier {
		return errors.New("unauthorized: only the assigned courier can confirm pickup")
	}
	if order.State != StateAccepted {
		return errors.New("invalid state transition")
	}
	
	// Verify the QR code secret
	hash := sha256.Sum256([]byte(pickupSecret))
	hashHex := hex.EncodeToString(hash[:])
	if hashHex != order.PickupQRHash {
		return errors.New("invalid pickup QR code")
	}

	order.State = StatePickedUp
	return nil
}

// ConfirmDelivery is called by the Courier after scanning the Buyer's QR code.
// The VM will transfer the foodAmount to the Restaurant, and deliveryFee to the Courier.
func (c *DeliveryEscrowContract) ConfirmDelivery(orderID string, courier [crypto.AddressSize]byte, dropoffSecret string) ([]*core.Transaction, error) {
	order, exists := c.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	if order.Courier != courier {
		return nil, errors.New("unauthorized: only the assigned courier can confirm delivery")
	}
	if order.State != StatePickedUp {
		return nil, errors.New("invalid state transition: order must be PICKED_UP")
	}

	// Verify the drop-off QR code secret
	hash := sha256.Sum256([]byte(dropoffSecret))
	hashHex := hex.EncodeToString(hash[:])
	if hashHex != order.DropoffQRHash {
		return nil, errors.New("invalid dropoff QR code")
	}

	order.State = StateDelivered
	order.DeliveredAt = time.Now().Unix()

	// Generate the internal settlement transactions (to be processed by the StateDB)
	var settlements []*core.Transaction
	
	// 1. Pay the restaurant
	if order.FoodAmount > 0 {
		settlements = append(settlements, &core.Transaction{
			Type:   core.TxTransfer,
			From:   order.Buyer, // In reality, from the Contract's escrow address
			To:     order.Restaurant,
			Amount: order.FoodAmount,
		})
	}
	
	// 2. Pay the courier
	if order.DeliveryFee > 0 {
		settlements = append(settlements, &core.Transaction{
			Type:   core.TxTransfer,
			From:   order.Buyer, 
			To:     order.Courier,
			Amount: order.DeliveryFee,
		})
	}

	return settlements, nil
}

// DisputeOrder allows the buyer or restaurant to freeze funds if food is missing or not delivered.
func (c *DeliveryEscrowContract) DisputeOrder(orderID string) error {
	order, exists := c.orders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	if order.State == StateDelivered || order.State == StateRefunded {
		return errors.New("order is already finalized")
	}

	order.State = StateDisputed
	// Funds remain locked until decentralized arbitration resolves it.
	return nil
}

// GetOrder returns the current state of an order.
func (c *DeliveryEscrowContract) GetOrder(orderID string) (*DeliveryOrder, error) {
	order, exists := c.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}
