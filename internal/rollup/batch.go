package rollup

import (
	"encoding/binary"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

// Batch represents a bundle of L2 transactions compressed into a single L1 payload.
type Batch struct {
	BatchID      uint64
	PreRoot      [crypto.HashSize]byte
	PostRoot     [crypto.HashSize]byte
	Transactions []*core.Transaction
	
	// Quantum-Safe Zero-Knowledge Proof (STARK) representing the valid execution
	// of all transactions transitioning the state from PreRoot to PostRoot.
	ZkStarkProof []byte 
}

// Hash computes the unique identifier of the batch.
func (b *Batch) Hash() [crypto.HashSize]byte {
	var buf []byte
	
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, b.BatchID)
	buf = append(buf, idBytes...)
	buf = append(buf, b.PreRoot[:]...)
	buf = append(buf, b.PostRoot[:]...)
	
	for _, tx := range b.Transactions {
		buf = append(buf, tx.Hash[:]...)
	}
	
	buf = append(buf, b.ZkStarkProof...)
	
	return crypto.Hash256(buf)
}
