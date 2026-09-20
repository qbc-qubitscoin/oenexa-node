package state

import (
	"crypto/rand"
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
)

// makeRandAddr generates a random 20-byte address for benchmarking
func makeRandAddr() [crypto.AddressSize]byte {
	var addr [crypto.AddressSize]byte
	_, _ = rand.Read(addr[:])
	return addr
}

// BenchmarkStateDBSetGet measures the raw overhead of writing to and reading from
// the zero-allocation state maps.
func BenchmarkStateDBSetGet(b *testing.B) {
	st := NewStateDB()
	addr := makeRandAddr()
	acc := &Account{Balance: 1000, Nonce: 1}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Set
		st.SetAccount(addr, acc)
		// Get
		_ = st.GetAccount(addr)
	}
}

// BenchmarkStateDBCommitRoot_10k measures the time it takes to sort and hash the
// entire StateDB when it contains 10,000 accounts.
func BenchmarkStateDBCommitRoot_10k(b *testing.B) {
	st := NewStateDB()
	for i := 0; i < 10000; i++ {
		st.SetAccount(makeRandAddr(), &Account{Balance: uint64(i), Nonce: 1})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = st.CommitRoot()
	}
}

// BenchmarkStateDBCommitRoot_100k scales the commit root benchmark to 100,000 accounts.
// This reveals the O(N log N) sorting overhead impact as the blockchain grows.
func BenchmarkStateDBCommitRoot_100k(b *testing.B) {
	st := NewStateDB()
	for i := 0; i < 100000; i++ {
		st.SetAccount(makeRandAddr(), &Account{Balance: uint64(i), Nonce: 1})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = st.CommitRoot()
	}
}
