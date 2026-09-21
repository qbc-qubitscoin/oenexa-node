package state_test

import (
	"crypto/rand"
	"testing"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/state"
)

// generateParallelBlock creates a block with entirely independent transactions
// to maximize parallel execution throughput.
func generateParallelBlock(count int) (*state.DB, *core.Block) {
	db := state.NewStateDB()

	var txs []*core.Transaction
	
	// Pre-fund accounts and generate txs
	for i := 0; i < count; i++ {
		var from [crypto.AddressSize]byte
		var to [crypto.AddressSize]byte
		rand.Read(from[:])
		rand.Read(to[:])
		
		db.SetAccount(from, &state.Account{Balance: 10 * core.OneOEN})
		
		tx := &core.Transaction{
			From:     from,
			To:       to,
			Amount:   core.OneOEN,
			GasLimit: 21000,
			GasPrice: 100,
			Nonce:    0,
		}
		txs = append(txs, tx)
	}

	header := core.BlockHeader{
		Height:   1,
		BaseFee:  50,
		GasLimit: 3000000000, // Very high block limit for 10k tests
	}

	blk := &core.Block{
		Header: header,
		Txs:    txs,
	}

	return db, blk
}

func BenchmarkApplyBlockSequential(b *testing.B) {
	db, blk := generateParallelBlock(10_000)
	var validator [crypto.AddressSize]byte
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap := db.Snapshot()
		_, err := state.ApplyBlock(snap, blk, validator, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyBlockParallel(b *testing.B) {
	db, blk := generateParallelBlock(10_000)
	var validator [crypto.AddressSize]byte
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap := db.Snapshot()
		_, err := state.ApplyBlockParallel(snap, blk, validator, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestParallelExecutionConsistency(t *testing.T) {
	dbSeq, blk := generateParallelBlock(1000)
	var validator [crypto.AddressSize]byte
	
	dbPar := dbSeq.Snapshot()

	resSeq, err := state.ApplyBlock(dbSeq, blk, validator, nil)
	if err != nil {
		t.Fatal(err)
	}
	
	resPar, err := state.ApplyBlockParallel(dbPar, blk, validator, nil)
	if err != nil {
		t.Fatal(err)
	}
	
	if resSeq.GasUsed != resPar.GasUsed {
		t.Errorf("GasUsed mismatch: seq=%d par=%d", resSeq.GasUsed, resPar.GasUsed)
	}
	if resSeq.FeeCollected != resPar.FeeCollected {
		t.Errorf("FeeCollected mismatch")
	}
	if resSeq.BurnedFees != resPar.BurnedFees {
		t.Errorf("BurnedFees mismatch")
	}
	
	rootSeq := dbSeq.CommitRoot()
	rootPar := dbPar.CommitRoot()
	
	if rootSeq != rootPar {
		t.Errorf("State Root mismatch! seq=%x par=%x", rootSeq, rootPar)
	}
}
