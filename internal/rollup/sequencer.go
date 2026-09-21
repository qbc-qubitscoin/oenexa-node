package rollup

import (
	"errors"
	"sync"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/state"
)

// Sequencer represents an L2 Rollup node that aggregates user transactions,
// executes them off-chain, and produces batches for L1.
type Sequencer struct {
	mu            sync.RWMutex
	l2State       *state.DB
	mempool       []*core.Transaction
	lastBatchID   uint64
	lastRoot      [crypto.HashSize]byte
	batchInterval time.Duration
	
	// Mock proof generator function
	proofGenerator func(pre, post [crypto.HashSize]byte, txs []*core.Transaction) []byte
}

func NewSequencer() *Sequencer {
	return &Sequencer{
		l2State:       state.NewStateDB(),
		mempool:       make([]*core.Transaction, 0),
		batchInterval: 2 * time.Second,
		proofGenerator: mockSTARKProof,
	}
}

// SetLastRoot sets the initial L2 root (e.g. Genesis)
func (s *Sequencer) SetLastRoot(r [crypto.HashSize]byte) {
	s.lastRoot = r
}

// SubmitTransaction accepts a transaction into the L2 mempool.
// In a real network, this would be an RPC endpoint.
func (s *Sequencer) SubmitTransaction(tx *core.Transaction) error {
	// 1. Verify ML-DSA-65 Signature
	if err := tx.Verify(); err != nil {
		return errors.New("invalid post-quantum signature: " + err.Error())
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.mempool = append(s.mempool, tx)
	return nil
}

// BuildBatch takes the current mempool, executes all txs in the L2 StateDB,
// computes the new state root, and generates a ZK-STARK proof.
func (s *Sequencer) BuildBatch() (*Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.mempool) == 0 {
		return nil, errors.New("mempool is empty")
	}
	
	// preRoot is the state root from the previous batch
	preRoot := s.lastRoot
	
	// Execute transactions
	txs := make([]*core.Transaction, len(s.mempool))
	copy(txs, s.mempool)
	s.mempool = s.mempool[:0] // Clear mempool
	
	// Execute on L2 state
	// Note: We use state.ApplyTransaction directly for this mock, ignoring VM context for pure token transfers if VM isn't strictly needed,
	// or we can just mock token transfers on L2 for now.
	for _, tx := range txs {
		// In full L2, we would use transition.ApplyMessage / state engine.
		// For simplicity in the Sequencer, we do a basic transfer if it's a standard Tx.
		if tx.Type == core.TxTransfer {
			sender := s.l2State.GetAccount(tx.From)
			if sender.Balance >= tx.Amount+tx.GasLimit {
				receiver := s.l2State.GetAccount(tx.To)
				sender.Balance -= tx.Amount
				sender.Nonce++
				receiver.Balance += tx.Amount
				s.l2State.SetAccount(tx.From, sender)
				s.l2State.SetAccount(tx.To, receiver)
			}
		}
	}

	postRoot := s.l2State.CommitRoot()
	s.lastBatchID++
	s.lastRoot = postRoot

	proof := s.proofGenerator(preRoot, postRoot, txs)

	batch := &Batch{
		BatchID:      s.lastBatchID,
		PreRoot:      preRoot,
		PostRoot:     postRoot,
		Transactions: txs,
		ZkStarkProof: proof,
	}

	return batch, nil
}

// GetL2State returns the current state of the L2.
func (s *Sequencer) GetL2State() *state.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.l2State
}

// mockSTARKProof simulates the generation of a post-quantum ZK-STARK proof.
// Real STARKs are hash-based and resilient to quantum computers.
func mockSTARKProof(pre, post [crypto.HashSize]byte, txs []*core.Transaction) []byte {
	// A real STARK proof would be tens of kilobytes. We'll simulate a 4KB proof.
	proof := make([]byte, 4096)
	// Seed it deterministically for tests
	copy(proof[:32], pre[:])
	copy(proof[32:64], post[:])
	return proof
}
