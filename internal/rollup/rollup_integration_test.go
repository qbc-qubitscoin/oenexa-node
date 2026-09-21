package rollup_test

import (
	"testing"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/rollup"
	"github.com/oenexa/oenexa/internal/state"
)

func TestRollupEndToEnd(t *testing.T) {
	// 1. Initialize Sequencer
	seq := rollup.NewSequencer()

	// 2. Fund some L2 accounts (mocking a bridge deposit)
	l2State := seq.GetL2State()
	
	// Create keys
	w1, _ := crypto.NewWallet()
	addr1 := w1.Address
	
	w2, _ := crypto.NewWallet()
	addr2 := w2.Address

	l2State.SetAccount(addr1, &state.Account{Balance: 1000})
	l2State.SetAccount(addr2, &state.Account{Balance: 0})
	
	// Genesis Root
	genesisRoot := l2State.CommitRoot()
	seq.SetLastRoot(genesisRoot)
	
	// Prototype limitation: CommitRoot clears the dirty map, but our GetAccount
	// only reads from the dirty map. Restore for execution.
	l2State.SetAccount(addr1, &state.Account{Balance: 1000})
	l2State.SetAccount(addr2, &state.Account{Balance: 0})

	// 3. Initialize L1 Bridge with L2 genesis root
	bridge := rollup.NewL1Bridge(genesisRoot)

	// 4. Create L2 Transactions
	tx := &core.Transaction{
		Version:   1,
		Type:      core.TxTransfer,
		Nonce:     1,
		From:      addr1,
		To:        addr2,
		Amount:    100,
		GasLimit:  0,
		GasPrice:  0,
		Timestamp: time.Now().Unix(),
		PublicKey: w1.PublicKey,
	}
	_ = tx.Sign(w1.PrivateKey)

	// 5. Submit Tx to Sequencer
	if err := seq.SubmitTransaction(tx); err != nil {
		t.Fatalf("failed to submit tx to sequencer: %v", err)
	}

	// 6. Sequencer builds Batch and ZK-STARK proof
	batch, err := seq.BuildBatch()
	if err != nil {
		t.Fatalf("failed to build batch: %v", err)
	}


	if err := bridge.SubmitBatch(batch); err != nil {
		t.Fatalf("L1 bridge rejected batch: %v", err)
	}

	// 8. Verify L1 Bridge updated the L2 Root properly
	if bridge.GetL2Root() != batch.PostRoot {
		t.Fatalf("L1 bridge root mismatch. Expected %x, got %x", batch.PostRoot, bridge.GetL2Root())
	}
	
	t.Logf("Successfully processed L2 batch with ZK-STARK. New L2 Root: %x", bridge.GetL2Root())
}
