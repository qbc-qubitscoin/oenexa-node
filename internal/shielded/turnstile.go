package shielded

import (
	"errors"
	"fmt"
	"sync"
)

// Turnstile enforces strict on-chain supply conservation across the dual-pool model:
// TransparentSupply + ShieldedSupply == TotalSupply at all times.
type Turnstile struct {
	mu                sync.RWMutex
	transparentSupply uint64
	shieldedSupply    uint64
}

// NewTurnstile initializes the turnstile with the initial transparent supply.
func NewTurnstile(initialTransparentSupply uint64) *Turnstile {
	return &Turnstile{
		transparentSupply: initialTransparentSupply,
		shieldedSupply:    0,
	}
}

// Shield moves value from the transparent pool into the shielded pool.
func (t *Turnstile) Shield(amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if amount == 0 {
		return errors.New("shielded: turnstile shield amount must be > 0")
	}
	if t.transparentSupply < amount {
		return fmt.Errorf("shielded: insufficient transparent supply to shield: have %d, want %d", t.transparentSupply, amount)
	}
	t.transparentSupply -= amount
	t.shieldedSupply += amount
	return nil
}

// Unshield moves value from the shielded pool back into the transparent pool.
func (t *Turnstile) Unshield(amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if amount == 0 {
		return errors.New("shielded: turnstile unshield amount must be > 0")
	}
	if t.shieldedSupply < amount {
		return fmt.Errorf("shielded: insufficient shielded supply to unshield: have %d, want %d", t.shieldedSupply, amount)
	}
	t.shieldedSupply -= amount
	t.transparentSupply += amount
	return nil
}

// Balances returns the current transparent and shielded supply totals.
func (t *Turnstile) Balances() (transparent uint64, shielded uint64) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.transparentSupply, t.shieldedSupply
}

// VerifyInvariant verifies that the sum of transparent and shielded pools equals expected total.
func (t *Turnstile) VerifyInvariant(expectedTotal uint64) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sum := t.transparentSupply + t.shieldedSupply
	if sum != expectedTotal {
		return fmt.Errorf("shielded: turnstile invariant violation! sum=%d (transparent=%d, shielded=%d), expected=%d",
			sum, t.transparentSupply, t.shieldedSupply, expectedTotal)
	}
	return nil
}

// Snapshot returns a shallow copy of turnstile state.
func (t *Turnstile) Snapshot() *Turnstile {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return &Turnstile{
		transparentSupply: t.transparentSupply,
		shieldedSupply:    t.shieldedSupply,
	}
}
