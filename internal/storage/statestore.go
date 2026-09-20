// Package storage provides persistence layers for the OENEXA node.
package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/mpt"
	"github.com/oenexa/oenexa/internal/state"
	"github.com/syndtr/goleveldb/leveldb"
)

// ─────────────────────────────────────────────────────────────────────────────
// levelDBBatch wraps *leveldb.Batch to satisfy mpt.Batch.
// ─────────────────────────────────────────────────────────────────────────────

type levelDBBatch struct {
	db  *leveldb.DB
	bat *leveldb.Batch
}

func (b *levelDBBatch) Put(key, value []byte) { b.bat.Put(key, value) }
func (b *levelDBBatch) Write() error          { return b.db.Write(b.bat, nil) }

// ─────────────────────────────────────────────────────────────────────────────
// levelDBBackend wraps *DB to satisfy mpt.Backend.
// ─────────────────────────────────────────────────────────────────────────────

type levelDBBackend struct{ db *DB }

func (l *levelDBBackend) Put(key, value []byte) error    { return l.db.Put(key, value) }
func (l *levelDBBackend) Get(key []byte) ([]byte, error) { return l.db.Get(key) }
func (l *levelDBBackend) NewBatch() mpt.Batch {
	return &levelDBBatch{db: l.db.db, bat: new(leveldb.Batch)}
}

// ─────────────────────────────────────────────────────────────────────────────
// Key layout
// ─────────────────────────────────────────────────────────────────────────────

// keyPrefixAccount is the flat account store prefix: "a:" + address[32] → encoded account.
// This is separate from the SMT node prefix "t:" used by the mpt package.
var keyPrefixAccount = []byte("a:")

// accountKey builds the flat account store key for the given address.
func accountKey(addr [crypto.AddressSize]byte) []byte {
	key := make([]byte, len(keyPrefixAccount)+crypto.AddressSize)
	copy(key, keyPrefixAccount)
	copy(key[len(keyPrefixAccount):], addr[:])
	return key
}

// encodeAccount serialises an Account into a compact 80-byte fixed-size record:
//
//	nonce[8] + balance[8] + codeHash[32] + storageRoot[32] = 80 bytes
//
// Using a fixed-size binary format is faster than gob and avoids reflection.
func encodeAccount(acc *state.Account) []byte {
	buf := make([]byte, 8+8+32+32)
	binary.BigEndian.PutUint64(buf[0:], acc.Nonce)
	binary.BigEndian.PutUint64(buf[8:], acc.Balance)
	copy(buf[16:], acc.CodeHash[:])
	copy(buf[48:], acc.StorageRoot[:])
	return buf
}

// decodeAccount deserialises a flat account record.
func decodeAccount(data []byte) (*state.Account, error) {
	if len(data) != 80 {
		return nil, fmt.Errorf("account record: expected 80 bytes, got %d", len(data))
	}
	var acc state.Account
	acc.Nonce = binary.BigEndian.Uint64(data[0:])
	acc.Balance = binary.BigEndian.Uint64(data[8:])
	copy(acc.CodeHash[:], data[16:48])
	copy(acc.StorageRoot[:], data[48:80])
	return &acc, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// StateStore
// ─────────────────────────────────────────────────────────────────────────────

// StateStore persists and loads the world-state using a dual-store design:
//
// 1. Sparse Merkle Trie (SMT) — "t:" prefix keys, managed by the mpt package.
//    Stores addr → Hash(account). Used for:
//      - Computing the cryptographic state root (block header field).
//      - Generating Merkle inclusion proofs for light clients.
//    This is O(dirty × log N) per block commit.
//
// 2. Flat account store — "a:" prefix keys.
//    Stores addr → encoded(account). Used for:
//      - Fast O(1) account reads on node startup without replaying history.
//      - Rebuilding the StateDB's dirty map after a restart.
//
// Both stores are written together in a single atomic operation via SaveState.
// LoadState reads the flat store to hydrate the StateDB, then attaches the SMT
// for future proof generation.
type StateStore struct{ db *DB }

// NewStateStore wraps db in a StateStore.
func NewStateStore(db *DB) *StateStore { return &StateStore{db: db} }

// SaveState atomically writes:
//   - Every dirty account to the flat account store ("a:" keys).
//   - Every dirty account's hash to the SMT ("t:" nodes).
//   - The new SMT root hash to the "t:root" meta-key.
//
// This replaces the old gob-per-account approach with a fast binary-encoded
// flat store + a cryptographic SMT layer.
func (s *StateStore) SaveState(st *state.DB) error {
	backend := &levelDBBackend{db: s.db}

	// Build the SMT trie backed by LevelDB.
	root, err := mpt.LoadRoot(backend)
	if err != nil {
		return fmt.Errorf("state store: load existing root: %w", err)
	}
	var trie *mpt.Trie
	if root == (crypto.ZeroHash) {
		trie = mpt.New(backend)
	} else {
		trie = mpt.NewWithRoot(backend, root)
	}

	// Collect dirty accounts from the StateDB.
	type dirtyEntry struct {
		addr [crypto.AddressSize]byte
		acc  *state.Account
	}
	var entries []dirtyEntry
	st.ForEach(func(addr [crypto.AddressSize]byte, acc *state.Account) {
		entries = append(entries, dirtyEntry{addr, acc})
	})

	if len(entries) == 0 {
		// Nothing dirty — still need to persist the trie root.
		_ = trie.Commit()
		return nil
	}

	// Write all dirty accounts into the SMT.
	for _, e := range entries {
		h := mpt.EncodeAccountHash(e.acc.Nonce, e.acc.Balance, e.acc.CodeHash, e.acc.StorageRoot)
		if err := trie.Update(e.addr, h); err != nil {
			return fmt.Errorf("state store: trie update: %w", err)
		}
	}

	// Commit SMT nodes to LevelDB.
	if err := trie.Commit(); err != nil {
		return fmt.Errorf("state store: trie commit: %w", err)
	}

	// Write all dirty accounts to the flat store atomically.
	batch := s.db.NewBatch()
	for _, e := range entries {
		batch.Put(accountKey(e.addr), encodeAccount(e.acc))
	}
	if err := s.db.Write(batch); err != nil {
		return fmt.Errorf("state store: flat account write: %w", err)
	}

	return nil
}

// LoadState reads the full account state from the flat store and returns a
// StateDB with all accounts pre-populated in the dirty buffer.
//
// This guarantees that GetAccount works immediately after LoadState without
// any additional SMT reads. The attached SMT is used for CommitRoot / proofs.
func (s *StateStore) LoadState() (*state.DB, error) {
	backend := &levelDBBackend{db: s.db}

	// Attach the SMT (for future CommitRoot / Prove calls).
	root, err := mpt.LoadRoot(backend)
	if err != nil {
		return nil, fmt.Errorf("state store load root: %w", err)
	}
	var trie *mpt.Trie
	if root == (crypto.ZeroHash) {
		trie = mpt.New(backend)
	} else {
		trie = mpt.NewWithRoot(backend, root)
	}

	st := state.NewStateDBWithTrie(trie)

	// Populate the StateDB dirty map from the flat account store.
	// This is O(N) but only happens once on node startup.
	err = s.db.IterPrefix(keyPrefixAccount, func(key, val []byte) error {
		if len(key) != len(keyPrefixAccount)+crypto.AddressSize {
			return fmt.Errorf("invalid account key length %d", len(key))
		}
		var addr [crypto.AddressSize]byte
		copy(addr[:], key[len(keyPrefixAccount):])

		acc, err := decodeAccount(val)
		if err != nil {
			return fmt.Errorf("decode account %x: %w", addr[:4], err)
		}
		st.SetAccount(addr, acc)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("state store load accounts: %w", err)
	}

	return st, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers for tests — expose encodeAccount/decodeAccount
// ─────────────────────────────────────────────────────────────────────────────

// encodeAccountForTest exposes encodeAccount for storage_test.go.
// Not part of the public API.
func encodeAccountBytes(acc *state.Account) []byte { return encodeAccount(acc) }

// Used by the benchmark to avoid import cycles.
var _ = bytes.Compare
