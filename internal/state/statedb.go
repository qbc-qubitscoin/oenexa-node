package state

import (
	"encoding/binary"
	"errors"
	"sort"
	"sync"

	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/shielded"
)

// DB StateDB is an in-memory trie-like account store protected by a mutex.
type DB struct {
	mu             sync.RWMutex
	accounts       map[[crypto.AddressSize]byte]*Account // address -> account
	shieldedPool   uint64              // Total nano-OEN in the shielded pool
	nullifiers     map[[32]byte]bool   // Spent note nullifiers
	commitmentTree *shielded.NoteCommitmentTree
}

// NewStateDB creates an empty state database.
func NewStateDB() *DB {
	return &DB{
		accounts:       make(map[[crypto.AddressSize]byte]*Account),
		nullifiers:     make(map[[32]byte]bool),
		commitmentTree: shielded.NewNoteCommitmentTree(),
	}
}

// GetAccount returns the account for the given address, or a zero account if absent.
func (s *DB) GetAccount(addr [crypto.AddressSize]byte) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if a, ok := s.accounts[addr]; ok {
		return a.Clone()
	}
	return &Account{}
}

// SetAccount stores the account for the given address.
func (s *DB) SetAccount(addr [crypto.AddressSize]byte, acc *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[addr] = acc.Clone()
}

// GetBalance returns the oenexa balance of an address.
func (s *DB) GetBalance(addr [crypto.AddressSize]byte) uint64 {
	return s.GetAccount(addr).Balance
}

// GetNonce returns the nonce of an address.
func (s *DB) GetNonce(addr [crypto.AddressSize]byte) uint64 {
	return s.GetAccount(addr).Nonce
}

// Credit adds amount to the balance of addr (used for block rewards).
func (s *DB) Credit(addr [crypto.AddressSize]byte, amount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.accounts[addr]
	if acc == nil {
		acc = &Account{}
	} else {
		acc = acc.Clone()
	}
	acc.Balance += amount
	s.accounts[addr] = acc
}

// CommitRoot computes a deterministic state root = SHA-3-256 (sorted account entries + shielded state).
func (s *DB) CommitRoot() [crypto.HashSize]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	
	// Compute hashes directly without string hex conversion
	var keys [][crypto.AddressSize]byte
	for k := range s.accounts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		for x := 0; x < crypto.AddressSize; x++ {
			if keys[i][x] != keys[j][x] {
				return keys[i][x] < keys[j][x]
			}
		}
		return false
	})

	var buf []byte
	b8 := make([]byte, 8)
	for _, k := range keys {
		a := s.accounts[k]
		buf = append(buf, k[:]...)
		binary.BigEndian.PutUint64(b8, a.Nonce)
		buf = append(buf, b8...)
		binary.BigEndian.PutUint64(b8, a.Balance)
		buf = append(buf, b8...)
		buf = append(buf, a.CodeHash[:]...)
		buf = append(buf, a.StorageRoot[:]...)
	}
	if s.shieldedPool > 0 || s.commitmentTree.LeafCount() > 0 {
		binary.BigEndian.PutUint64(b8, s.shieldedPool)
		buf = append(buf, b8...)
		root := s.commitmentTree.Root()
		buf = append(buf, root[:]...)
	}
	return crypto.Hash256(buf)
}

// GetShieldedBalance returns the total nano-OEN held in the shielded pool.
func (s *DB) GetShieldedBalance() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shieldedPool
}

// CreditShieldedPool adds amount to the shielded pool.
func (s *DB) CreditShieldedPool(amount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shieldedPool += amount
}

// DebitShieldedPool deducts amount from the shielded pool.
func (s *DB) DebitShieldedPool(amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shieldedPool < amount {
		return errors.New("insufficient shielded pool balance")
	}
	s.shieldedPool -= amount
	return nil
}

// HasNullifier returns true if the nullifier has already been revealed/spent.
func (s *DB) HasNullifier(nf [32]byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nullifiers[nf]
}

// AddNullifier registers a nullifier as spent.
func (s *DB) AddNullifier(nf [32]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nullifiers[nf] {
		return errors.New("nullifier already spent")
	}
	s.nullifiers[nf] = true
	return nil
}

// RemoveNullifier unregisters a nullifier (used for rollback on execution failure).
func (s *DB) RemoveNullifier(nf [32]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nullifiers, nf)
}

// AppendNoteCommitment adds a note commitment to the tree accumulator.
func (s *DB) AppendNoteCommitment(cm [32]byte) (uint32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commitmentTree.Append(cm)
}

// CommitmentTreeRoot returns the root of the shielded note commitment tree.
func (s *DB) CommitmentTreeRoot() [32]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.commitmentTree.Root()
}

// CommitmentTree returns the commitment tree.
func (s *DB) CommitmentTree() *shielded.NoteCommitmentTree {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.commitmentTree
}

// TotalTransparentSupply calculates the total transparent balance across all accounts.
func (s *DB) TotalTransparentSupply() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var total uint64
	for _, a := range s.accounts {
		total += a.Balance
	}
	return total
}

// TotalSupply returns the total supply (transparent + shielded).
func (s *DB) TotalSupply() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var total uint64
	for _, a := range s.accounts {
		total += a.Balance
	}
	return total + s.shieldedPool
}

// Snapshot returns a deep copy of the current state (for rollback).
func (s *DB) Snapshot() *DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := NewStateDB()
	for k, v := range s.accounts {
		snap.accounts[k] = v.Clone()
	}
	snap.shieldedPool = s.shieldedPool
	for k := range s.nullifiers {
		snap.nullifiers[k] = true
	}
	snap.commitmentTree = s.commitmentTree.Clone()
	return snap
}

// Len returns the number of accounts.
func (s *DB) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.accounts)
}

// ForEach calls fn for every account in the database (read-locked).
// The callback must not call any StateDB method that acquires the write lock.
func (s *DB) ForEach(fn func(addr [crypto.AddressSize]byte, acc *Account)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for addr, acc := range s.accounts {
		fn(addr, acc.Clone())
	}
}

// Apply replaces this state with the contents of another (used after dry-run).
func (s *DB) Apply(other *DB) {
	if s == other || other == nil {
		return
	}
	other.mu.RLock()
	defer other.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts = make(map[[crypto.AddressSize]byte]*Account, len(other.accounts))
	for k, v := range other.accounts {
		s.accounts[k] = v.Clone()
	}
	s.shieldedPool = other.shieldedPool
	s.nullifiers = make(map[[32]byte]bool, len(other.nullifiers))
	for k := range other.nullifiers {
		s.nullifiers[k] = true
	}
	s.commitmentTree = other.commitmentTree.Clone()
}
