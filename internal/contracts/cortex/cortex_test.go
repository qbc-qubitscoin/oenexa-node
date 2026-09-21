package cortex

import (
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
)

func TestCortexDataCenterFlow(t *testing.T) {
	contract := NewCortexContract()

	ownerWallet, _ := crypto.NewWallet()
	investor1Wallet, _ := crypto.NewWallet()
	investor2Wallet, _ := crypto.NewWallet()
	aiDeveloperWallet, _ := crypto.NewWallet()

	assetID := "DC_FRANKFURT_01"
	totalShares := uint64(100)
	pricePerShare := uint64(1000)

	// 1. Register the Data Center
	asset, err := contract.RegisterDataCenter(assetID, ownerWallet.Address, totalShares, pricePerShare)
	if err != nil {
		t.Fatalf("Failed to register data center: %v", err)
	}

	if asset.State != StateFunding {
		t.Fatalf("Expected state FUNDING, got %s", asset.State)
	}

	// 2. Investor 1 buys 60% of shares
	txs1, err := contract.PurchaseShares(assetID, investor1Wallet.Address, 60)
	if err != nil {
		t.Fatalf("Failed to purchase shares: %v", err)
	}
	if len(txs1) != 1 || txs1[0].Amount != 60000 {
		t.Fatalf("Invalid transaction generated for investor 1")
	}

	// 3. Investor 2 buys remaining 40% of shares
	_, err = contract.PurchaseShares(assetID, investor2Wallet.Address, 40)
	if err != nil {
		t.Fatalf("Failed to purchase shares: %v", err)
	}

	// Contract should automatically transition to BUILDING because funding is met
	if asset.State != StateBuilding {
		t.Fatalf("Expected state BUILDING, got %s", asset.State)
	}

	// 4. Activate the Data Center (Construction Complete)
	err = contract.ActivateDataCenter(assetID, ownerWallet.Address)
	if err != nil {
		t.Fatalf("Failed to activate data center: %v", err)
	}
	
	if asset.State != StateActive {
		t.Fatalf("Expected state ACTIVE, got %s", asset.State)
	}

	// 5. AI Developer leases GPUs for 5000 OEN
	txsCompute, err := contract.PayForCompute(assetID, aiDeveloperWallet.Address, 5000)
	if err != nil {
		t.Fatalf("Failed to pay for compute: %v", err)
	}
	if len(txsCompute) != 1 || txsCompute[0].Amount != 5000 {
		t.Fatalf("Invalid transaction generated for AI compute payment")
	}

	if asset.RevenuePool != 5000 {
		t.Fatalf("Expected RevenuePool 5000, got %d", asset.RevenuePool)
	}

	// 6. Investor 1 claims yield (60% of 5000 = 3000)
	txsYield1, err := contract.ClaimYield(assetID, investor1Wallet.Address)
	if err != nil {
		t.Fatalf("Failed to claim yield: %v", err)
	}
	if txsYield1[0].Amount != 3000 {
		t.Fatalf("Expected yield 3000, got %d", txsYield1[0].Amount)
	}

	// 7. Investor 2 claims yield (40% of 5000 = 2000)
	txsYield2, err := contract.ClaimYield(assetID, investor2Wallet.Address)
	if err != nil {
		t.Fatalf("Failed to claim yield: %v", err)
	}
	if txsYield2[0].Amount != 2000 {
		t.Fatalf("Expected yield 2000, got %d", txsYield2[0].Amount)
	}

	// 8. Try claiming again (should fail because no new yield)
	_, err = contract.ClaimYield(assetID, investor1Wallet.Address)
	if err == nil {
		t.Fatalf("Expected error when claiming without new yield, but got nil")
	}

	// 9. AI Developer uses more compute, adding 10000 OEN
	_, err = contract.PayForCompute(assetID, aiDeveloperWallet.Address, 10000)
	if err != nil {
		t.Fatalf("Failed to pay for compute 2nd time: %v", err)
	}

	// 10. Investor 1 claims new yield (60% of 10000 = 6000)
	txsYield1_Round2, err := contract.ClaimYield(assetID, investor1Wallet.Address)
	if err != nil {
		t.Fatalf("Failed to claim second yield: %v", err)
	}
	if txsYield1_Round2[0].Amount != 6000 {
		t.Fatalf("Expected second yield 6000, got %d", txsYield1_Round2[0].Amount)
	}
}
