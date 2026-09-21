package p2p_test

import (
	"context"
	"testing"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/p2p"
)

func TestLibp2pNode_TxGossip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Create two identities
	w1, _ := crypto.NewWallet()
	id1 := p2p.NewIdentity(w1.PublicKey, w1.PrivateKey, "")
	
	w2, _ := crypto.NewWallet()
	id2 := p2p.NewIdentity(w2.PublicKey, w2.PrivateKey, "")

	// 2. Start two libp2p nodes
	node1, err := p2p.NewNode(id1, 0)
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	
	node2, err := p2p.NewNode(id2, 0)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}

	// 3. Setup hooks
	txChan := make(chan *core.Transaction, 1)
	node2.OnTxReceived = func(tx *core.Transaction) {
		txChan <- tx
	}

	if err := node1.Start(ctx); err != nil {
		t.Fatalf("node1 start: %v", err)
	}
	if err := node2.Start(ctx); err != nil {
		t.Fatalf("node2 start: %v", err)
	}

	// 4. M-DNS takes time to discover (or we can skip it and manually connect)
	// We'll manually connect to speed up the test using mDNS port (which is randomly assigned because port 0).
	// But libp2p node doesn't expose Multiaddrs publicly yet. Let's just wait a bit for mDNS.
	// Actually, mDNS discovery can be flaky in tests. We'll wait 2 seconds.
	time.Sleep(2 * time.Second)

	// 5. Broadcast a Transaction
	tx := &core.Transaction{
		Version:   1,
		Type:      core.TxTransfer,
		Nonce:     1,
		From:      w1.Address,
		To:        w2.Address,
		Amount:    100,
		Timestamp: time.Now().Unix(),
		PublicKey: w1.PublicKey,
	}
	_ = tx.Sign(w1.PrivateKey)

	if err := node1.BroadcastTx(ctx, tx); err != nil {
		t.Fatalf("broadcast tx: %v", err)
	}

	// 6. Wait for node2 to receive it
	select {
	case receivedTx := <-txChan:
		if receivedTx.Hash != tx.Hash {
			t.Fatalf("received wrong tx hash")
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for gossip message. mDNS discovery might have failed.")
	}
}
