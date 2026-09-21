package cortex

import (
	"errors"
	"sync"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

// DataCenterState represents the current operational phase of the Data Center
type DataCenterState string

const (
	StateFunding    DataCenterState = "FUNDING"
	StateBuilding   DataCenterState = "BUILDING"
	StateActive     DataCenterState = "ACTIVE"
)

// CortexAsset represents a tokenized Tier-4 Data Center
type CortexAsset struct {
	AssetID         string
	Owner           [crypto.AddressSize]byte // The infrastructure provider
	TotalFunding    uint64                   // Total OEN required to build
	CurrentFunding  uint64                   // OEN raised so far
	TotalShares     uint64                   // Total fractional shares available
	PricePerShare   uint64                   // Cost of 1 share in OEN
	RevenuePool     uint64                   // OEN accumulated from Compute-as-a-Service
	State           DataCenterState
	CreatedAt       int64
	
	// Track how many shares each investor holds
	Investors       map[[crypto.AddressSize]byte]uint64
	
	// Track how much revenue has been claimed by each investor to prevent double-claiming
	ClaimedRevenue  map[[crypto.AddressSize]byte]uint64
}

// CortexContract simulates the OenexaVM smart contract for the AI Data Center Grid
type CortexContract struct {
	mu     sync.RWMutex
	assets map[string]*CortexAsset
}

// NewCortexContract initializes a new Cortex infrastructure registry
func NewCortexContract() *CortexContract {
	return &CortexContract{
		assets: make(map[string]*CortexAsset),
	}
}

// RegisterDataCenter allows an infrastructure provider to propose a new Data Center for funding
func (c *CortexContract) RegisterDataCenter(assetID string, owner [crypto.AddressSize]byte, totalShares, pricePerShare uint64) (*CortexAsset, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.assets[assetID]; exists {
		return nil, errors.New("data center asset already exists")
	}

	asset := &CortexAsset{
		AssetID:        assetID,
		Owner:          owner,
		TotalFunding:   totalShares * pricePerShare,
		CurrentFunding: 0,
		TotalShares:    totalShares,
		PricePerShare:  pricePerShare,
		RevenuePool:    0,
		State:          StateFunding,
		CreatedAt:      time.Now().Unix(),
		Investors:      make(map[[crypto.AddressSize]byte]uint64),
		ClaimedRevenue: make(map[[crypto.AddressSize]byte]uint64),
	}

	c.assets[assetID] = asset
	return asset, nil
}

// PurchaseShares allows an investor to fund the data center by buying shares
func (c *CortexContract) PurchaseShares(assetID string, investor [crypto.AddressSize]byte, shareCount uint64) ([]*core.Transaction, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	asset, exists := c.assets[assetID]
	if !exists {
		return nil, errors.New("data center not found")
	}
	if asset.State != StateFunding {
		return nil, errors.New("data center is no longer accepting funding")
	}

	cost := shareCount * asset.PricePerShare

	// Check if this purchase exceeds the total funding goal
	if asset.CurrentFunding+cost > asset.TotalFunding {
		return nil, errors.New("purchase exceeds remaining funding required")
	}

	// Update asset state
	asset.CurrentFunding += cost
	asset.Investors[investor] += shareCount

	// If fully funded, transition state
	if asset.CurrentFunding == asset.TotalFunding {
		asset.State = StateBuilding
	}

	// Generate the transaction to lock the investor's OEN into the contract
	tx := &core.Transaction{
		Type:   core.TxTransfer,
		From:   investor,
		To:     asset.Owner, // Alternatively, to a locked contract address
		Amount: cost,
	}

	return []*core.Transaction{tx}, nil
}

// ActivateDataCenter is called by the Oracle/Owner when construction is finished
func (c *CortexContract) ActivateDataCenter(assetID string, owner [crypto.AddressSize]byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	asset, exists := c.assets[assetID]
	if !exists {
		return errors.New("data center not found")
	}
	if asset.Owner != owner {
		return errors.New("only owner can activate")
	}
	if asset.State != StateBuilding {
		return errors.New("data center must be fully funded and in building state to activate")
	}

	asset.State = StateActive
	return nil
}

// PayForCompute allows AI developers/SaaS platforms to rent GPU/TPU power, adding OEN to the Revenue Pool
func (c *CortexContract) PayForCompute(assetID string, client [crypto.AddressSize]byte, paymentAmount uint64) ([]*core.Transaction, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	asset, exists := c.assets[assetID]
	if !exists {
		return nil, errors.New("data center not found")
	}
	if asset.State != StateActive {
		return nil, errors.New("data center is not active and cannot process compute workloads yet")
	}

	asset.RevenuePool += paymentAmount

	// Transaction representing the payment into the contract's yield pool
	// In a real VM, 'To' would be the contract's internal address
	tx := &core.Transaction{
		Type:   core.TxTransfer,
		From:   client,
		To:     asset.Owner, // Simplified: sending to owner's pool
		Amount: paymentAmount,
	}

	return []*core.Transaction{tx}, nil
}

// ClaimYield allows an investor to withdraw their proportional share of the accumulated Revenue Pool
func (c *CortexContract) ClaimYield(assetID string, investor [crypto.AddressSize]byte) ([]*core.Transaction, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	asset, exists := c.assets[assetID]
	if !exists {
		return nil, errors.New("data center not found")
	}

	shares := asset.Investors[investor]
	if shares == 0 {
		return nil, errors.New("no shares owned in this data center")
	}

	// Calculate proportional total revenue this investor is entitled to:
	// Entitlement = (Shares / TotalShares) * RevenuePool
	// To prevent precision loss with integer math, multiply first
	entitlement := (shares * asset.RevenuePool) / asset.TotalShares

	// Subtract what they have already claimed in previous periods
	claimed := asset.ClaimedRevenue[investor]
	availableToClaim := entitlement - claimed

	if availableToClaim == 0 {
		return nil, errors.New("no new yield available to claim")
	}

	// Update claimed amount
	asset.ClaimedRevenue[investor] += availableToClaim

	// Generate transaction dispatching yield to investor
	tx := &core.Transaction{
		Type:   core.TxTransfer,
		From:   asset.Owner, // Yield pool
		To:     investor,
		Amount: availableToClaim,
	}

	return []*core.Transaction{tx}, nil
}
