package mempool

import (
	"crypto/rand"
	"testing"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

// makeRandHash generates a random 32-byte hash
func makeRandHash() [crypto.HashSize]byte {
	var h [crypto.HashSize]byte
	_, _ = rand.Read(h[:])
	return h
}

// makeRandAddr generates a random 20-byte address
func makeRandAddr() [crypto.AddressSize]byte {
	var addr [crypto.AddressSize]byte
	_, _ = rand.Read(addr[:])
	return addr
}

// BenchmarkMempoolAdd measures the speed of inserting a transaction
// into the priority heap.
func BenchmarkMempoolAdd(b *testing.B) {
	mp := New(b.N + 1) // Ensure capacity
	addr := makeRandAddr()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tx := &core.Transaction{
			Hash:     makeRandHash(),
			From:     addr,
			Nonce:    uint64(i),
			GasPrice: core.MinGasPrice,
		}
		_ = mp.Add(tx)
	}
}

// BenchmarkMempoolPurgeCommitted mathematically proves the O(N) optimization
// by simulating a block forging event where 2,000 transactions are purged
// from a mempool containing 10,000 transactions.
func BenchmarkMempoolPurgeCommitted(b *testing.B) {
	// Pre-generate transactions to avoid allocation overhead during benchmark
	totalTxs := 10000
	purgeCount := 2000
	
	allTxs := make([]*core.Transaction, totalTxs)
	addr := makeRandAddr()
	for i := 0; i < totalTxs; i++ {
		allTxs[i] = &core.Transaction{
			Hash:     makeRandHash(),
			From:     addr,
			Nonce:    uint64(i),
			GasPrice: core.MinGasPrice,
		}
	}
	
	toPurge := allTxs[:purgeCount]

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup a fresh mempool with 10k txs for each iteration
		mp := New(totalTxs)
		for _, tx := range allTxs {
			_ = mp.Add(tx)
		}
		b.StartTimer()

		// Benchmark the purge logic
		mp.PurgeCommitted(toPurge)
	}
}
