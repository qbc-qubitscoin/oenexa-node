package shielded

import (
	"errors"
	"sync"

	"golang.org/x/crypto/sha3"
)

// ComputeNullifier derives a unique, deterministic nullifier for spending a note:
// nf = SHA-3-256("OENEXA_NULLIFIER" || spendingKey || rho || commitment)
func ComputeNullifier(spendingKey []byte, rho [32]byte, commitment [32]byte) [32]byte {
	h := sha3.New256()
	h.Write([]byte("OENEXA_NULLIFIER"))
	h.Write(spendingKey)
	h.Write(rho[:])
	h.Write(commitment[:])
	var nf [32]byte
	copy(nf[:], h.Sum(nil))
	return nf
}

// NullifierSet maintains the set of spent nullifiers to prevent double spending.
type NullifierSet struct {
	mu         sync.RWMutex
	nullifiers map[[32]byte]bool
}

// NewNullifierSet initializes an empty nullifier set.
func NewNullifierSet() *NullifierSet {
	return &NullifierSet{
		nullifiers: make(map[[32]byte]bool),
	}
}

// Has returns true if the nullifier has already been spent.
func (ns *NullifierSet) Has(nf [32]byte) bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.nullifiers[nf]
}

// Add marks a nullifier as spent. Returns an error if it was already spent.
func (ns *NullifierSet) Add(nf [32]byte) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if ns.nullifiers[nf] {
		return errors.New("shielded: double spend detected, nullifier already exists")
	}
	ns.nullifiers[nf] = true
	return nil
}

// Len returns the count of spent nullifiers.
func (ns *NullifierSet) Len() int {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return len(ns.nullifiers)
}

// Clone creates an independent copy of the nullifier set.
func (ns *NullifierSet) Clone() *NullifierSet {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	c := NewNullifierSet()
	for k, v := range ns.nullifiers {
		c.nullifiers[k] = v
	}
	return c
}
