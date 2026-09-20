package rwa

import (
	"errors"
	"fmt"
)

// RWA Platform Launch represents the core module for Phase 15
type RwaModule struct {
	Active bool
}

func NewRwaModule() *RwaModule {
	return &RwaModule{Active: true}
}

func (m *RwaModule) Execute() string {
	return fmt.Sprintf("Phase %d: RWA Platform Launch executed successfully", 15)
}

// Asset represents an on-chain tokenized Real World Asset supporting transparent and shielded units.
type Asset struct {
	ID            string
	Name          string
	TotalUnits    uint64
	ShieldedUnits uint64
	ValuationUSD  uint64 // in USD cents
}

// TokenizeAsset creates a new RWA token specification.
func (m *RwaModule) TokenizeAsset(id, name string, totalUnits, valuationUSD uint64) (*Asset, error) {
	if id == "" || name == "" || totalUnits == 0 {
		return nil, errors.New("invalid asset parameters")
	}
	return &Asset{
		ID:           id,
		Name:         name,
		TotalUnits:   totalUnits,
		ValuationUSD: valuationUSD,
	}, nil
}

// TransparentUnits returns the publicly visible units.
func (a *Asset) TransparentUnits() uint64 {
	return a.TotalUnits - a.ShieldedUnits
}

// ShieldUnits moves transparent units into the shielded pool.
func (a *Asset) ShieldUnits(units uint64) error {
	if a.ShieldedUnits+units > a.TotalUnits {
		return errors.New("insufficient transparent units to shield")
	}
	a.ShieldedUnits += units
	return nil
}

// UnshieldUnits withdraws shielded units back to transparent circulation.
func (a *Asset) UnshieldUnits(units uint64) error {
	if units > a.ShieldedUnits {
		return errors.New("insufficient shielded units to unshield")
	}
	a.ShieldedUnits -= units
	return nil
}
