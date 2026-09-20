package mpt

import (
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
)

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkMPTUpdate — single leaf insertion cost
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkMPTUpdate measures the amortised cost of inserting one new
// address+value into the trie. The trie is pre-populated with 1,000 accounts
// before the timer starts so the result reflects average tree depth, not a
// best-case empty-trie insertion.
func BenchmarkMPTUpdate(b *testing.B) {
	trie := New(nil)

	// Pre-populate with 1,000 accounts.
	for i := 0; i < 1_000; i++ {
		_ = trie.Update(makeAddr(i), makeHash(i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Use (1_000 + i) to avoid hitting existing leaves on every iteration.
		_ = trie.Update(makeAddr(1_000+i), makeHash(i))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkMPTCommitRoot — flush N dirty accounts and compute new root
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkMPTCommitRoot_10k benchmarks inserting 10,000 dirty accounts and
// flushing them to the in-memory backend (equivalent of one block commit).
//
// Compare against the old BenchmarkStateDBCommitRoot_10k baseline (~5.8 ms).
func BenchmarkMPTCommitRoot_10k(b *testing.B) {
	benchmarkCommitRoot(b, 10_000)
}

// BenchmarkMPTCommitRoot_100k benchmarks inserting 100,000 dirty accounts and
// flushing them (large-state-size stress test).
//
// Compare against the old BenchmarkStateDBCommitRoot_100k baseline (~59 ms).
func BenchmarkMPTCommitRoot_100k(b *testing.B) {
	benchmarkCommitRoot(b, 100_000)
}

// benchmarkCommitRoot is the shared implementation for commit root benchmarks.
// Each iteration creates a fresh trie, bulk-updates N accounts in memory
// (simulating a block of N state changes), calls Commit(), and reads the root.
// This mirrors the exact call path in state.DB.CommitRoot().
func benchmarkCommitRoot(b *testing.B, n int) {
	b.ReportAllocs()

	// Pre-build addresses and hashes outside the timer to measure only the
	// trie operations themselves.
	addrs := make([][crypto.AddressSize]byte, n)
	hashes := make([][32]byte, n)
	for i := 0; i < n; i++ {
		addrs[i] = makeAddr(i)
		hashes[i] = makeHash(i)
	}

	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		b.StopTimer()
		trie := New(nil) // fresh in-memory trie per iteration
		b.StartTimer()

		// Simulate block execution: update all N accounts.
		for i := 0; i < n; i++ {
			_ = trie.Update(addrs[i], hashes[i])
		}

		// Commit (flush dirty cache to backend) and read the new root.
		_ = trie.Commit()
		_ = trie.Root()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkMPTGet — read path (cache-miss, load from backend)
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkMPTGet measures the read latency for an account that is already
// committed to the backend (i.e. no dirty-cache hit — forces a backend lookup).
func BenchmarkMPTGet(b *testing.B) {
	backend := newMapBackend()
	trie := New(backend)

	const N = 10_000
	for i := 0; i < N; i++ {
		_ = trie.Update(makeAddr(i), makeHash(i))
	}
	_ = trie.Commit()

	// Reconstruct trie without dirty cache (simulates post-restart state).
	root := trie.Root()
	coldTrie := NewWithRoot(backend, root)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = coldTrie.Get(makeAddr(i % N))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkMPTProve — Merkle proof generation
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkMPTProve measures the time to generate a 160-step Merkle proof for
// a leaf in a trie containing 10,000 accounts. This is the call that will be
// served to light clients.
func BenchmarkMPTProve(b *testing.B) {
	trie := New(nil)
	const N = 10_000
	for i := 0; i < N; i++ {
		_ = trie.Update(makeAddr(i), makeHash(i))
	}

	target := makeAddr(5_000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = trie.Prove(target)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkVerifyProof — Merkle proof verification (light-client side)
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkVerifyProof measures the computational cost of verifying an
// inclusion proof on the light-client side. This is a pure in-memory
// computation: 160 SHA-3-256 hash operations.
func BenchmarkVerifyProof(b *testing.B) {
	trie := New(nil)
	addr := makeAddr(42)
	val := makeHash(42)
	_ = trie.Update(addr, val)

	proof, _ := trie.Prove(addr)
	root := trie.Root()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = VerifyProof(root, addr, val, proof)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BenchmarkEncodeAccountHash — leaf value encoding
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkEncodeAccountHash measures the cost of encoding one account's
// fields into the 32-byte trie leaf value. This runs on every dirty account
// during CommitRoot(), so it should be as fast as possible.
func BenchmarkEncodeAccountHash(b *testing.B) {
	var code, storage [32]byte
	code[0] = 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeAccountHash(uint64(i), 1_000_000, code, storage)
	}
}
