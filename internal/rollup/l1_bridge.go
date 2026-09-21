package rollup

import (
	"bytes"
	"errors"
	"sync"

	"github.com/oenexa/oenexa/internal/crypto"
)

// L1Bridge acts as the Smart Contract on the main OENEXA chain (L1)
// that verifies and records L2 batches submitted by the Sequencer.
type L1Bridge struct {
	mu             sync.RWMutex
	currentL2Root  [crypto.HashSize]byte
	batchHistory   map[uint64]*Batch
	lastBatchID    uint64
}

func NewL1Bridge(genesisRoot [crypto.HashSize]byte) *L1Bridge {
	return &L1Bridge{
		currentL2Root: genesisRoot,
		batchHistory:  make(map[uint64]*Batch),
		lastBatchID:   0,
	}
}

// SubmitBatch is called by the Sequencer on the L1 chain.
func (b *L1Bridge) SubmitBatch(batch *Batch) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Verify sequential ordering
	if batch.BatchID != b.lastBatchID+1 {
		return errors.New("invalid batch sequence ID")
	}

	// 2. Verify state transition continuity
	if !bytes.Equal(batch.PreRoot[:], b.currentL2Root[:]) {
		return errors.New("pre-root does not match current L2 state root")
	}

	// 3. Verify ZK-STARK Proof (Post-Quantum Security)
	// In a real system, we would run a STARK verifier here.
	if !verifySTARK(batch.PreRoot, batch.PostRoot, batch.ZkStarkProof) {
		return errors.New("invalid ZK-STARK proof")
	}

	// 4. Accept Batch and update L1's view of L2 state
	b.currentL2Root = batch.PostRoot
	b.batchHistory[batch.BatchID] = batch
	b.lastBatchID = batch.BatchID

	return nil
}

// GetL2Root returns the latest finalized L2 state root on L1.
func (b *L1Bridge) GetL2Root() [crypto.HashSize]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentL2Root
}

// verifySTARK is a mock verification function for our ZK-STARKs.
func verifySTARK(pre, post [crypto.HashSize]byte, proof []byte) bool {
	if len(proof) < 64 {
		return false
	}
	// Check our mock STARK's deterministic seeds
	return bytes.Equal(proof[:32], pre[:]) && bytes.Equal(proof[32:64], post[:])
}
