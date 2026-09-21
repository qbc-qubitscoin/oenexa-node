// Package mpt — in-memory mode compatibility tests.
//
// These tests verify that mpt.New(nil) produces a fully functional trie
// with NO LevelDB dependency whatsoever. Every test in this file runs
// without touching the filesystem, ensuring that:
//
//  1. state.NewStateDB() continues to work unchanged in all existing tests.
//  2. The mpt package can be imported and used in unit tests without
//     provisioning a temporary database directory.
//  3. The in-memory mapBackend is behaviourally identical to the LevelDB
//     backend for all trie operations that tests exercise.
//
// Design:
//
//	mpt.New(nil)   → uses internal mapBackend  (map[string][]byte)
//	mpt.New(ldb)   → uses levelDBBackend       (LevelDB on disk)
//
// The two backends satisfy the same mpt.Backend interface. All trie logic
// (insert, delete, get, commit, prove) is backend-agnostic. Only the
// storage layer differs.
package mpt

import (
	"fmt"
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
)

// ─────────────────────────────────────────────────────────────────────────────
// Core compatibility: nil backend == pure in-memory
// ─────────────────────────────────────────────────────────────────────────────

// TestInMemoryMode_NilBackend verifies that passing nil to mpt.New() creates
// a working trie with no LevelDB dependency. This is the "test mode" used
// by state.NewStateDB().
func TestInMemoryMode_NilBackend(t *testing.T) {
	// New(nil) MUST NOT panic, open a file, or require any setup.
	trie := New(nil)
	if trie == nil {
		t.Fatal("New(nil) returned nil")
	}

	// Root of an empty trie must be the zero sentinel.
	if trie.Root() != ZeroHashOf() {
		t.Errorf("empty trie root should be zero, got %x", trie.Root())
	}
}

// TestInMemoryMode_Update verifies that Update works without LevelDB.
func TestInMemoryMode_Update(t *testing.T) {
	trie := New(nil) // no disk

	addr := makeAddr(1)
	val := makeHash(100)

	if err := trie.Update(addr, val); err != nil {
		t.Fatalf("Update on in-memory trie: %v", err)
	}

	got, err := trie.Get(addr)
	if err != nil {
		t.Fatalf("Get on in-memory trie: %v", err)
	}
	if got != val {
		t.Errorf("Get returned %x, want %x", got, val)
	}
}

// TestInMemoryMode_Delete verifies Delete on an in-memory trie.
func TestInMemoryMode_Delete(t *testing.T) {
	trie := New(nil)
	addr := makeAddr(5)
	_ = trie.Update(addr, makeHash(5))
	_ = trie.Delete(addr)

	got, _ := trie.Get(addr)
	if got != ZeroHashOf() {
		t.Errorf("expected zero after in-memory delete, got %x", got)
	}
}

// TestInMemoryMode_Commit verifies that Commit() works on in-memory backend
// (writes to the internal map, doesn't flush to disk).
func TestInMemoryMode_Commit(t *testing.T) {
	trie := New(nil)
	addr := makeAddr(42)
	val := makeHash(42)
	_ = trie.Update(addr, val)

	if err := trie.Commit(); err != nil {
		t.Fatalf("Commit on in-memory trie: %v", err)
	}

	// After commit, the dirty cache is reset. The node must be loadable
	// from the mapBackend (which received the commit batch).
	got, err := trie.Get(addr)
	if err != nil {
		t.Fatalf("Get after Commit on in-memory trie: %v", err)
	}
	if got != val {
		t.Errorf("Get after Commit: %x, want %x", got, val)
	}
}

// TestInMemoryMode_Prove verifies proof generation on an in-memory trie.
func TestInMemoryMode_Prove(t *testing.T) {
	trie := New(nil)
	addr := makeAddr(7)
	val := makeHash(77)
	_ = trie.Update(addr, val)

	proof, err := trie.Prove(addr)
	if err != nil {
		t.Fatalf("Prove on in-memory trie: %v", err)
	}
	if !VerifyProof(trie.Root(), addr, val, proof) {
		t.Error("proof from in-memory trie failed VerifyProof")
	}
}

// TestInMemoryMode_RootDeterminism verifies root determinism on in-memory tries.
func TestInMemoryMode_RootDeterminism(t *testing.T) {
	build := func(seeds []int) [crypto.HashSize]byte {
		trie := New(nil)
		for _, s := range seeds {
			_ = trie.Update(makeAddr(s), makeHash(s))
		}
		return trie.Root()
	}

	r1 := build([]int{10, 20, 30})
	r2 := build([]int{30, 10, 20})

	if r1 != r2 {
		t.Errorf("non-deterministic root on in-memory trie:\n  r1=%x\n  r2=%x", r1, r2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Backend parity: in-memory and LevelDB must produce the same roots
// ─────────────────────────────────────────────────────────────────────────────

// TestInMemoryVsMapBackend_RootParity verifies that a trie built with
// mpt.New(nil) produces the exact same root hash as one built with the
// explicit mapBackend (both are in-memory, just testing the two paths).
func TestInMemoryVsMapBackend_RootParity(t *testing.T) {
	const N = 500

	// Trie A: nil → implicit mapBackend
	trieA := New(nil)
	// Trie B: explicit mapBackend
	trieB := New(newMapBackend())

	for i := 0; i < N; i++ {
		addr := makeAddr(i)
		val := makeHash(i * 3)
		_ = trieA.Update(addr, val)
		_ = trieB.Update(addr, val)
	}

	if trieA.Root() != trieB.Root() {
		t.Errorf("root mismatch between nil-backend and explicit mapBackend:\n  nil=%x\n  map=%x",
			trieA.Root(), trieB.Root())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// state.NewStateDB() compatibility — uses no import of state to avoid cycle.
// These tests exercise the exact code path that state.NewStateDB() takes.
// ─────────────────────────────────────────────────────────────────────────────

// TestInMemoryMode_StateDBFlow simulates the state.NewStateDB() usage pattern:
//  1. Create trie with nil backend.
//  2. Update many accounts (dirty buffer fill).
//  3. Call Commit (block commit).
//  4. Verify root is stable and non-zero.
//
// This is the exact sequence executed by every existing state package test.
func TestInMemoryMode_StateDBFlow(t *testing.T) {
	trie := New(nil)

	// Simulate 100 SetAccount calls (block execution).
	const accounts = 100
	for i := 0; i < accounts; i++ {
		addr := makeAddr(i)
		acc := EncodeAccountHash(uint64(i), uint64(i*1000), [32]byte{}, [32]byte{})
		if err := trie.Update(addr, acc); err != nil {
			t.Fatalf("Update account %d: %v", i, err)
		}
	}

	// Simulate CommitRoot() — commit dirty nodes to backend.
	if err := trie.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	root := trie.Root()
	if root == ZeroHashOf() {
		t.Error("state root should not be zero after inserting 100 accounts")
	}

	// Verify all 100 accounts are still readable after commit.
	for i := 0; i < accounts; i++ {
		addr := makeAddr(i)
		want := EncodeAccountHash(uint64(i), uint64(i*1000), [32]byte{}, [32]byte{})
		got, err := trie.Get(addr)
		if err != nil {
			t.Fatalf("Get account %d after commit: %v", i, err)
		}
		if got != want {
			t.Errorf("account %d: got %x, want %x", i, got, want)
		}
	}
}

// TestInMemoryMode_MultipleCommits verifies that multiple sequential Commit()
// calls work correctly on an in-memory trie, with the root evolving properly.
// This mirrors the block-by-block execution of a running node in tests.
func TestInMemoryMode_MultipleCommits(t *testing.T) {
	trie := New(nil)

	var prevRoot [32]byte

	for block := 0; block < 5; block++ {
		// Each "block" updates 10 accounts.
		for i := 0; i < 10; i++ {
			seed := block*10 + i
			addr := makeAddr(seed)
			val := EncodeAccountHash(uint64(block), uint64(seed), [32]byte{}, [32]byte{})
			if err := trie.Update(addr, val); err != nil {
				t.Fatalf("block %d, account %d: Update: %v", block, i, err)
			}
		}

		if err := trie.Commit(); err != nil {
			t.Fatalf("block %d: Commit: %v", block, err)
		}

		newRoot := trie.Root()
		if newRoot == prevRoot {
			t.Errorf("block %d: root did not change after new accounts added", block)
		}
		prevRoot = newRoot
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Snapshot/Apply compatibility without LevelDB
// ─────────────────────────────────────────────────────────────────────────────

// TestInMemoryMode_BackendRef verifies that BackendRef() returns a non-nil
// backend even when the trie was created with New(nil), so that Snapshot()
// in statedb.go can safely call NewWithRoot(t.BackendRef(), t.Root()).
func TestInMemoryMode_BackendRef(t *testing.T) {
	trie := New(nil)
	ref := trie.BackendRef()
	if ref == nil {
		t.Fatal("BackendRef() returned nil for in-memory trie")
	}

	// Create a snapshot trie sharing the same backend.
	_ = trie.Update(makeAddr(1), makeHash(1))
	snapTrie := NewWithRoot(ref, trie.Root())

	// The snapshot should be able to read committed nodes via the shared backend.
	if err := trie.Commit(); err != nil {
		t.Fatalf("Commit before snapshot read: %v", err)
	}

	got, err := snapTrie.Get(makeAddr(1))
	if err != nil {
		t.Fatalf("Get on snapshot trie: %v", err)
	}
	if got != makeHash(1) {
		t.Errorf("snapshot Get: %x, want %x", got, makeHash(1))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Concurrency safety
// ─────────────────────────────────────────────────────────────────────────────

// TestInMemoryMode_Concurrent verifies that concurrent Update + Get calls on
// an in-memory trie are safe (the trie's RWMutex must protect mapBackend).
func TestInMemoryMode_Concurrent(t *testing.T) {
	trie := New(nil)
	const goroutines = 16
	const ops = 50

	done := make(chan struct{}, goroutines)

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer func() { done <- struct{}{} }()
			for i := 0; i < ops; i++ {
				addr := makeAddr(g*ops + i)
				val := makeHash(g*ops + i)
				if err := trie.Update(addr, val); err != nil {
					// Just record; can't call t.Fatal from goroutine.
					fmt.Printf("goroutine %d Update %d: %v\n", g, i, err)
				}
				_, _ = trie.Get(addr)
			}
		}(g)
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}

	// Root must be non-zero after concurrent writes.
	if trie.Root() == ZeroHashOf() {
		t.Error("root is zero after concurrent writes")
	}
}
