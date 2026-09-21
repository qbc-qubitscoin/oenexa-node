package dcommerce

import (
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
)

func TestDeliveryEscrowFlow(t *testing.T) {
	contract := NewDeliveryEscrowContract()

	// Generate some mock addresses
	buyerWallet, _ := crypto.NewWallet()
	restaurantWallet, _ := crypto.NewWallet()
	courierWallet, _ := crypto.NewWallet()

	foodAmount := uint64(500)
	deliveryFee := uint64(50)

	// 1. Create Order
	order, err := contract.CreateOrder(buyerWallet.Address, restaurantWallet.Address, foodAmount, deliveryFee)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if order.State != StateCreated {
		t.Fatalf("Expected state %s, got %s", StateCreated, order.State)
	}

	// 2. Accept Order
	err = contract.AcceptOrder(order.OrderID, restaurantWallet.Address)
	if err != nil {
		t.Fatalf("Failed to accept order: %v", err)
	}

	// 3. Assign Courier
	err = contract.AssignCourier(order.OrderID, courierWallet.Address)
	if err != nil {
		t.Fatalf("Failed to assign courier: %v", err)
	}

	// 4. Confirm Delivery
	settlements, err := contract.ConfirmDelivery(order.OrderID, buyerWallet.Address)
	if err != nil {
		t.Fatalf("Failed to confirm delivery: %v", err)
	}

	if len(settlements) != 2 {
		t.Fatalf("Expected 2 settlement transactions, got %d", len(settlements))
	}

	if settlements[0].Amount != foodAmount || settlements[0].To != restaurantWallet.Address {
		t.Fatalf("Invalid restaurant settlement")
	}

	if settlements[1].Amount != deliveryFee || settlements[1].To != courierWallet.Address {
		t.Fatalf("Invalid courier settlement")
	}

	if order.State != StateDelivered {
		t.Fatalf("Expected state %s, got %s", StateDelivered, order.State)
	}
}

func TestDeliveryEscrowDispute(t *testing.T) {
	contract := NewDeliveryEscrowContract()
	buyerWallet, _ := crypto.NewWallet()
	restaurantWallet, _ := crypto.NewWallet()

	order, _ := contract.CreateOrder(buyerWallet.Address, restaurantWallet.Address, 100, 10)
	
	err := contract.DisputeOrder(order.OrderID)
	if err != nil {
		t.Fatalf("Failed to dispute order: %v", err)
	}

	if order.State != StateDisputed {
		t.Fatalf("Expected state %s, got %s", StateDisputed, order.State)
	}
}
