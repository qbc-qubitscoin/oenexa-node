package crypto

import (
	"testing"
)

// BenchmarkMLDSASign measures the performance of signing a typical 32-byte transaction hash
// using the Post-Quantum ML-DSA-65 algorithm.
func BenchmarkMLDSASign(b *testing.B) {
	wallet, err := NewWallet()
	if err != nil {
		b.Fatalf("failed to create wallet: %v", err)
	}

	msg := Hash256([]byte("benchmark_tx_hash_simulation"))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Sign(wallet.PrivateKey, msg[:])
		if err != nil {
			b.Fatalf("signing failed: %v", err)
		}
	}
}

// BenchmarkMLDSAVerify measures the performance of verifying an ML-DSA-65 signature.
// This is critical because validators must verify hundreds of signatures per block.
func BenchmarkMLDSAVerify(b *testing.B) {
	wallet, err := NewWallet()
	if err != nil {
		b.Fatalf("failed to create wallet: %v", err)
	}

	msg := Hash256([]byte("benchmark_tx_hash_simulation"))
	sig, err := Sign(wallet.PrivateKey, msg[:])
	if err != nil {
		b.Fatalf("signing failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := Verify(wallet.PublicKey, msg[:], sig)
		if err != nil || !valid {
			b.Fatalf("verification failed or invalid")
		}
	}
}
