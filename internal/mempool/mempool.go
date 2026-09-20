package mempool

import (
	"container/heap"
	"errors"
	"sync"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

const (
	DefaultMaxSize    = 10_000
	MaxTxDataSize     = 1 << 20 // 1 MiB
	MaxSenderQueueLen = 64
)

// Mempool is a thread-safe priority queue of unconfirmed transactions.
type Mempool struct {
	mu       sync.RWMutex
	txs      map[[crypto.HashSize]byte]*core.Transaction   // hash -> tx
	bySender map[[crypto.AddressSize]byte][]*core.Transaction // addr -> txs
	heap     txHeap
	maxSize  int
}

// New creates a new Mempool with the given maximum capacity.
func New(maxSize int) *Mempool {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	mp := &Mempool{
		txs:      make(map[[crypto.HashSize]byte]*core.Transaction),
		bySender: make(map[[crypto.AddressSize]byte][]*core.Transaction),
		maxSize:  maxSize,
	}
	heap.Init(&mp.heap)
	return mp
}

// Add validates and inserts a transaction into the pool.
func (mp *Mempool) Add(tx *core.Transaction) error {
	if err := tx.BasicValidate(); err != nil {
		return err
	}
	if len(tx.Data) > MaxTxDataSize {
		return errors.New("tx data exceeds maximum size")
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	if _, exists := mp.txs[tx.Hash]; exists {
		return errors.New("duplicate transaction")
	}
	if len(mp.txs) >= mp.maxSize {
		return errors.New("mempool full")
	}

	senderKey := tx.From
	if len(mp.bySender[senderKey]) >= MaxSenderQueueLen {
		return errors.New("sender queue full")
	}

	mp.txs[tx.Hash] = tx
	mp.bySender[senderKey] = append(mp.bySender[senderKey], tx)
	heap.Push(&mp.heap, tx)
	return nil
}

// Pending returns up to n highest-priority transactions without removing them.
func (mp *Mempool) Pending(n int) []*core.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Copy heap to avoid mutating it.
	tmp := make(txHeap, len(mp.heap))
	copy(tmp, mp.heap)
	heap.Init(&tmp)

	result := make([]*core.Transaction, 0, n)
	for len(result) < n && tmp.Len() > 0 {
		result = append(result, heap.Pop(&tmp).(*core.Transaction))
	}
	return result
}

// Get retrieves a transaction by its hash hex string.
func (mp *Mempool) Get(hash [crypto.HashSize]byte) (*core.Transaction, bool) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	tx, ok := mp.txs[hash]
	return tx, ok
}

// Remove deletes a transaction by hash.
func (mp *Mempool) Remove(hash [crypto.HashSize]byte) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.remove(hash)
	mp.rebuildHeap()
}

func (mp *Mempool) remove(hash [crypto.HashSize]byte) {
	tx, ok := mp.txs[hash]
	if !ok {
		return
	}
	delete(mp.txs, hash)

	senderKey := tx.From
	senderTxs := mp.bySender[senderKey]
	for i, t := range senderTxs {
		if t.Hash == hash {
			mp.bySender[senderKey] = append(senderTxs[:i], senderTxs[i+1:]...)
			break
		}
	}
	if len(mp.bySender[senderKey]) == 0 {
		delete(mp.bySender, senderKey)
	}
}

func (mp *Mempool) rebuildHeap() {
	mp.heap = make(txHeap, 0, len(mp.txs))
	for _, t := range mp.txs {
		mp.heap = append(mp.heap, t)
	}
	heap.Init(&mp.heap)
}

// PurgeCommitted removes all transactions whose hashes appear in the committed set.
func (mp *Mempool) PurgeCommitted(committed []*core.Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	for _, tx := range committed {
		mp.remove(tx.Hash)
	}
	mp.rebuildHeap()
}

// Size returns the current number of transactions in the pool.
func (mp *Mempool) Size() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.txs)
}

// Len is an alias for Size, satisfying common interface expectations.
func (mp *Mempool) Len() int { return mp.Size() }
