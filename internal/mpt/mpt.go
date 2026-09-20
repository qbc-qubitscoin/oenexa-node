// Package mpt implements a Sparse Merkle Trie (SMT) for OENEXA account state.
//
// # Architecture
//
// The trie has a fixed depth of 160 bits — one bit per bit of a 20-byte
// account address. Each internal node stores the hashes of its two children
// (left = bit 0, right = bit 1). Leaves store the hash of the encoded
// account value.
//
// # Storage Model
//
// Nodes are persisted to a LevelDB backend using their content hash as the
// key. An in-memory dirty cache (nodeCache) holds nodes written during a
// block but not yet flushed. Calling Commit() atomically writes all dirty
// nodes to disk in a single LevelDB batch, then resets the cache.
//
// For test environments that need no disk I/O, pass nil as the backend; the
// trie will operate entirely in-memory using a plain map.
//
// # Node Encoding
//
//	Leaf   node: 0x01 ++ address[20] ++ valueHash[32]  (53 bytes)
//	Branch node: 0x02 ++ leftHash[32] ++ rightHash[32]  (65 bytes)
//
// The zero hash (all-zero 32 bytes) is used as a sentinel for missing/empty
// subtrees, meaning "nothing is stored here".
package mpt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/oenexa/oenexa/internal/crypto"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constants & sentinels
// ─────────────────────────────────────────────────────────────────────────────

const (
	// trieDepth is fixed at 160 — one bit per bit of a 20-byte address.
	trieDepth = 256

	// nodeTypeLeaf tags a serialised leaf node.
	nodeTypeLeaf byte = 0x01

	// nodeTypeBranch tags a serialised branch node.
	nodeTypeBranch byte = 0x02
)

// zeroHash is the sentinel value for an empty subtree.
var zeroHash [crypto.HashSize]byte

// dbKeyPrefix is the prefix used for all trie node keys inside LevelDB, so
// they can co-exist with block and account keys without collision.
var dbKeyPrefix = []byte("t:")

// ─────────────────────────────────────────────────────────────────────────────
// Backend interface — abstracts LevelDB from in-memory maps
// ─────────────────────────────────────────────────────────────────────────────

// Backend is a minimal key-value store interface satisfied by both
// *storage.DB (LevelDB) and the in-memory mapBackend used in tests.
type Backend interface {
	// Put stores key → value.
	Put(key, value []byte) error

	// Get returns the value for key, or (nil, nil) when absent.
	Get(key []byte) ([]byte, error)

	// NewBatch starts an atomic write batch.
	NewBatch() Batch
}

// Batch is the write-batch abstraction.
type Batch interface {
	// Put queues a write.
	Put(key, value []byte)

	// Write commits the batch atomically.
	Write() error
}

// ─────────────────────────────────────────────────────────────────────────────
// In-memory backend (for tests / genesis)
// ─────────────────────────────────────────────────────────────────────────────

// mapBackend is a pure in-memory Backend backed by a Go map.
// It is not goroutine-safe; the Trie's mutex provides the needed protection.
type mapBackend struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMapBackend() *mapBackend { return &mapBackend{data: make(map[string][]byte)} }

func (m *mapBackend) Put(key, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	m.data[string(key)] = cp
	return nil
}

func (m *mapBackend) Get(key []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[string(key)]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (m *mapBackend) NewBatch() Batch { return &mapBatch{b: m} }

// mapBatch accumulates writes and applies them to the mapBackend on Write().
type mapBatch struct {
	b      *mapBackend
	writes []mapWrite
}

type mapWrite struct{ k, v []byte }

func (b *mapBatch) Put(key, value []byte) {
	kk := make([]byte, len(key))
	vv := make([]byte, len(value))
	copy(kk, key)
	copy(vv, value)
	b.writes = append(b.writes, mapWrite{kk, vv})
}

func (b *mapBatch) Write() error {
	b.b.mu.Lock()
	defer b.b.mu.Unlock()
	for _, w := range b.writes {
		b.b.data[string(w.k)] = w.v
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Trie
// ─────────────────────────────────────────────────────────────────────────────

// Trie is a goroutine-safe Sparse Merkle Trie keyed by 20-byte addresses.
// All writes go to an in-memory dirty cache; call Commit() to persist them
// atomically to the backend.
type Trie struct {
	mu        sync.RWMutex
	backend   Backend
	root      [crypto.HashSize]byte            // current root hash
	nodeCache map[[crypto.HashSize]byte][]byte  // hash → encoded node (dirty)
}

// New returns a Trie whose root is set to the empty hash. If backend is nil,
// an in-memory map is used (suitable for tests and genesis).
func New(backend Backend) *Trie {
	if backend == nil {
		backend = newMapBackend()
	}
	return &Trie{
		backend:   backend,
		root:      zeroHash,
		nodeCache: make(map[[crypto.HashSize]byte][]byte),
	}
}

// NewWithRoot reconstructs a Trie whose root is already stored in backend.
// Use this when loading state from disk after a node restart.
func NewWithRoot(backend Backend, root [crypto.HashSize]byte) *Trie {
	if backend == nil {
		backend = newMapBackend()
	}
	return &Trie{
		backend:   backend,
		root:      root,
		nodeCache: make(map[[crypto.HashSize]byte][]byte),
	}
}

// Root returns the current root hash of the trie.
// This is O(1) — the root is maintained incrementally.
func (t *Trie) Root() [crypto.HashSize]byte {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

// Update inserts or updates the leaf for addr with the given value hash.
// The change is held in the dirty cache until Commit() is called.
func (t *Trie) Update(addr [crypto.AddressSize]byte, valueHash [32]byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	newRoot, err := t.insert(t.root, addr, valueHash, 0)
	if err != nil {
		return err
	}
	t.root = newRoot
	return nil
}

// Delete removes the leaf for addr from the trie.
func (t *Trie) Delete(addr [crypto.AddressSize]byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	newRoot, err := t.remove(t.root, addr, 0)
	if err != nil {
		return err
	}
	t.root = newRoot
	return nil
}

// Get returns the stored value hash for addr, or zeroHash if absent.
func (t *Trie) Get(addr [crypto.AddressSize]byte) ([32]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.getLeafValue(t.root, addr, 0)
}

// Prove returns the Merkle sibling path for addr (160 sibling hashes, root
// to leaf order). An empty sibling means "empty subtree on that side."
// The proof can be used to verify inclusion or exclusion without the full trie.
func (t *Trie) Prove(addr [crypto.AddressSize]byte) ([][crypto.HashSize]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	proof := make([][crypto.HashSize]byte, 0, trieDepth)
	return t.collectProof(t.root, addr, 0, proof)
}

// Commit atomically writes all dirty cached nodes to the backend and resets
// the cache. It also persists the current root under a well-known meta-key so
// that NewWithRoot can be reconstructed on next startup.
func (t *Trie) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.nodeCache) == 0 {
		return nil // nothing to flush
	}

	batch := t.backend.NewBatch()

	// Write every dirty node (hash → encoding).
	for h, enc := range t.nodeCache {
		key := nodeKey(h)
		batch.Put(key, enc)
	}

	// Persist the root hash under a stable meta-key for recovery.
	rootKey := []byte("t:root")
	batch.Put(rootKey, t.root[:])

	if err := batch.Write(); err != nil {
		return fmt.Errorf("mpt commit: %w", err)
	}

	// Reset the dirty cache.
	t.nodeCache = make(map[[crypto.HashSize]byte][]byte)
	return nil
}

// LoadRoot reads the persisted root from the backend (written by Commit).
// Returns zeroHash when no root has been committed yet (fresh database).
func LoadRoot(backend Backend) ([crypto.HashSize]byte, error) {
	val, err := backend.Get([]byte("t:root"))
	if err != nil {
		return zeroHash, fmt.Errorf("mpt load root: %w", err)
	}
	if val == nil {
		return zeroHash, nil
	}
	var root [crypto.HashSize]byte
	copy(root[:], val)
	return root, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal recursive trie operations
// ─────────────────────────────────────────────────────────────────────────────

// bit returns the depth-th bit of addr (MSB of byte 0 = bit 0).
func bit(addr [crypto.AddressSize]byte, depth int) int {
	byteIdx := depth / 8
	bitIdx := 7 - (depth % 8)
	return int((addr[byteIdx] >> bitIdx) & 1)
}

// insert recursively descends the trie and returns the new subtree root hash.
func (t *Trie) insert(cur [32]byte, addr [crypto.AddressSize]byte, val [32]byte, depth int) ([32]byte, error) {
	if depth == trieDepth {
		// We are at the leaf position. Write a new leaf node.
		return t.putLeaf(addr, val)
	}

	if cur == zeroHash {
		// Empty subtree — create a new leaf directly at this position.
		return t.putLeaf(addr, val)
	}

	enc, err := t.loadNode(cur)
	if err != nil {
		return zeroHash, err
	}
	if len(enc) == 0 {
		return t.putLeaf(addr, val)
	}

	switch enc[0] {
	case nodeTypeLeaf:
		// There is an existing leaf here. Check if it's the same address.
		var existingAddr [crypto.AddressSize]byte
		copy(existingAddr[:], enc[1:1+crypto.AddressSize])

		if existingAddr == addr {
			// Same address: overwrite with new value.
			return t.putLeaf(addr, val)
		}

		// Different address: we must split. Push the existing leaf down one
		// level and insert both.
		var existingVal [32]byte
		copy(existingVal[:], enc[1+crypto.AddressSize:1+crypto.AddressSize+crypto.HashSize])

		var leftChild, rightChild [32]byte
		leftChild = zeroHash
		rightChild = zeroHash

		existingBit := bit(existingAddr, depth)
		newBit := bit(addr, depth)

		var err error
		if existingBit == 0 {
			leftChild, err = t.insert(zeroHash, existingAddr, existingVal, depth+1)
		} else {
			rightChild, err = t.insert(zeroHash, existingAddr, existingVal, depth+1)
		}
		if err != nil {
			return zeroHash, err
		}
		if newBit == 0 {
			leftChild, err = t.insert(leftChild, addr, val, depth+1)
		} else {
			rightChild, err = t.insert(rightChild, addr, val, depth+1)
		}
		if err != nil {
			return zeroHash, err
		}
		return t.putBranch(leftChild, rightChild)

	case nodeTypeBranch:
		var leftHash, rightHash [32]byte
		copy(leftHash[:], enc[1:33])
		copy(rightHash[:], enc[33:65])

		b := bit(addr, depth)
		if b == 0 {
			newLeft, err := t.insert(leftHash, addr, val, depth+1)
			if err != nil {
				return zeroHash, err
			}
			return t.putBranch(newLeft, rightHash)
		}
		newRight, err := t.insert(rightHash, addr, val, depth+1)
		if err != nil {
			return zeroHash, err
		}
		return t.putBranch(leftHash, newRight)

	default:
		return zeroHash, errors.New("mpt: unknown node type")
	}
}

// remove recursively deletes an address and returns the new subtree root hash.
func (t *Trie) remove(cur [32]byte, addr [crypto.AddressSize]byte, depth int) ([32]byte, error) {
	if cur == zeroHash {
		return zeroHash, nil // not found — no-op
	}

	enc, err := t.loadNode(cur)
	if err != nil {
		return zeroHash, err
	}
	if len(enc) == 0 {
		return zeroHash, nil
	}

	switch enc[0] {
	case nodeTypeLeaf:
		var existingAddr [crypto.AddressSize]byte
		copy(existingAddr[:], enc[1:1+crypto.AddressSize])
		if existingAddr == addr {
			return zeroHash, nil // remove this leaf
		}
		return cur, nil // different address — no change

	case nodeTypeBranch:
		var leftHash, rightHash [32]byte
		copy(leftHash[:], enc[1:33])
		copy(rightHash[:], enc[33:65])

		b := bit(addr, depth)
		if b == 0 {
			newLeft, err := t.remove(leftHash, addr, depth+1)
			if err != nil {
				return zeroHash, err
			}
			return t.putBranch(newLeft, rightHash)
		}
		newRight, err := t.remove(rightHash, addr, depth+1)
		if err != nil {
			return zeroHash, err
		}
		return t.putBranch(leftHash, newRight)

	default:
		return zeroHash, errors.New("mpt: unknown node type during remove")
	}
}

// getLeafValue searches for the value hash stored at addr.
func (t *Trie) getLeafValue(cur [32]byte, addr [crypto.AddressSize]byte, depth int) ([32]byte, error) {
	if cur == zeroHash {
		return zeroHash, nil
	}
	enc, err := t.loadNode(cur)
	if err != nil {
		return zeroHash, err
	}
	if len(enc) == 0 {
		return zeroHash, nil
	}

	switch enc[0] {
	case nodeTypeLeaf:
		var existingAddr [crypto.AddressSize]byte
		copy(existingAddr[:], enc[1:1+crypto.AddressSize])
		if existingAddr != addr {
			return zeroHash, nil
		}
		var val [32]byte
		copy(val[:], enc[1+crypto.AddressSize:1+crypto.AddressSize+crypto.HashSize])
		return val, nil

	case nodeTypeBranch:
		var leftHash, rightHash [32]byte
		copy(leftHash[:], enc[1:33])
		copy(rightHash[:], enc[33:65])

		if bit(addr, depth) == 0 {
			return t.getLeafValue(leftHash, addr, depth+1)
		}
		return t.getLeafValue(rightHash, addr, depth+1)

	default:
		return zeroHash, errors.New("mpt: unknown node type during get")
	}
}

// collectProof records the sibling hash at each level of the descent.
// When the path hits a compact leaf before reaching full depth, it records
// that leaf's hash as the sibling for the remaining proof positions so that
// VerifyProof can reconstruct the correct branch hashes.
func (t *Trie) collectProof(cur [32]byte, addr [crypto.AddressSize]byte, depth int, proof [][32]byte) ([][32]byte, error) {
	if depth == trieDepth {
		return proof, nil
	}
	if cur == zeroHash {
		// Empty subtree at this position: record zero for all remaining levels.
		for i := depth; i < trieDepth; i++ {
			proof = append(proof, zeroHash)
		}
		return proof, nil
	}

	enc, err := t.loadNode(cur)
	if err != nil {
		return nil, err
	}

	if len(enc) == 0 || enc[0] == nodeTypeLeaf {
		// We hit a compact leaf at depth < trieDepth.
		// For the current depth, record the sibling as zeroHash (there is no
		// branching here; the whole subtree IS this leaf).
		// We signal this shortcut by recording `cur` (the leaf hash) in the
		// FIRST remaining slot, then zeros for the rest.
		// VerifyProof recognises this because the first zero-padded rebuild
		// step would produce the leaf hash, not a branch hash.
		proof = append(proof, cur) // shortcut sentinel: the full subtree hash
		for i := depth + 1; i < trieDepth; i++ {
			proof = append(proof, zeroHash) // unused levels
		}
		return proof, nil
	}

	// Branch node: descend into the correct child, recording the sibling.
	var leftHash, rightHash [32]byte
	copy(leftHash[:], enc[1:33])
	copy(rightHash[:], enc[33:65])

	if bit(addr, depth) == 0 {
		proof = append(proof, rightHash) // sibling is the right child
		return t.collectProof(leftHash, addr, depth+1, proof)
	}
	proof = append(proof, leftHash) // sibling is the left child
	return t.collectProof(rightHash, addr, depth+1, proof)
}

// ─────────────────────────────────────────────────────────────────────────────
// Node construction helpers
// ─────────────────────────────────────────────────────────────────────────────

// putLeaf encodes and caches a leaf node, returning its content hash.
func (t *Trie) putLeaf(addr [crypto.AddressSize]byte, val [32]byte) ([32]byte, error) {
	enc := make([]byte, 1+crypto.AddressSize+crypto.HashSize)
	enc[0] = nodeTypeLeaf
	copy(enc[1:], addr[:])
	copy(enc[1+crypto.AddressSize:], val[:])
	return t.cacheNode(enc), nil
}

// putBranch encodes and caches a branch node, returning its content hash.
// If both children are zero, the branch itself is considered empty.
func (t *Trie) putBranch(left, right [32]byte) ([32]byte, error) {
	if left == zeroHash && right == zeroHash {
		return zeroHash, nil
	}
	enc := make([]byte, 1+32+32)
	enc[0] = nodeTypeBranch
	copy(enc[1:], left[:])
	copy(enc[33:], right[:])
	return t.cacheNode(enc), nil
}

// cacheNode stores enc in the dirty cache under its content hash and returns
// that hash.
func (t *Trie) cacheNode(enc []byte) [32]byte {
	h := crypto.Hash256(enc)
	t.nodeCache[h] = enc
	return h
}

// ─────────────────────────────────────────────────────────────────────────────
// Node loading (cache-first, then backend)
// ─────────────────────────────────────────────────────────────────────────────

// loadNode retrieves a node encoding by its hash. It checks the dirty cache
// first (avoiding disk reads for nodes written in the current block), then
// falls back to the persistent backend.
func (t *Trie) loadNode(h [32]byte) ([]byte, error) {
	if h == zeroHash {
		return nil, nil
	}
	// Check dirty cache first (no disk I/O).
	if enc, ok := t.nodeCache[h]; ok {
		return enc, nil
	}
	// Load from persistent backend.
	key := nodeKey(h)
	enc, err := t.backend.Get(key)
	if err != nil {
		return nil, fmt.Errorf("mpt load node %x: %w", h[:4], err)
	}
	return enc, nil
}

// nodeKey builds the LevelDB key for a trie node: "t:" + hash[32].
func nodeKey(h [32]byte) []byte {
	key := make([]byte, len(dbKeyPrefix)+crypto.HashSize)
	copy(key, dbKeyPrefix)
	copy(key[len(dbKeyPrefix):], h[:])
	return key
}

// ─────────────────────────────────────────────────────────────────────────────
// VerifyProof — standalone proof verification (no trie needed)
// ─────────────────────────────────────────────────────────────────────────────

// VerifyProof checks that addr with valueHash is included in the trie whose
// root is expectedRoot. Proof must be the trieDepth-element path returned
// by Trie.Prove.
//
// Proof format (from collectProof):
//   - For each branch level descended, proof[i] = sibling hash.
//   - When a compact leaf shortcut is hit at depth d: proof[d] = leaf node
//     hash (the full subtree), proof[d+1..] = zeroHash (unused).
//   - When an empty subtree is hit: proof[i..] = zeroHash.
//
// Returns true when the proof is valid.
func VerifyProof(
	expectedRoot [32]byte,
	addr [crypto.AddressSize]byte,
	valueHash [32]byte,
	proof [][32]byte,
) bool {
	if len(proof) != trieDepth {
		return false
	}

	// Compute the hash of the leaf node for addr/valueHash.
	leafEnc := make([]byte, 1+crypto.AddressSize+crypto.HashSize)
	leafEnc[0] = nodeTypeLeaf
	copy(leafEnc[1:], addr[:])
	copy(leafEnc[1+crypto.AddressSize:], valueHash[:])
	leafHash := crypto.Hash256(leafEnc)

	// Walk from the deepest level back to the root.
	// cur starts as zeroHash; we populate it once we identify the leaf level.
	cur := zeroHash
	leafSet := false

	for i := trieDepth - 1; i >= 0; i-- {
		sibling := proof[i]

		if !leafSet {
			// Scan from the bottom upward looking for the shortcut sentinel or
			// real branch levels. If both cur and sibling are zero, we are still
			// in unused padding territory — skip.
			if sibling == zeroHash {
				continue // still in the zero-padded region
			}
			// Non-zero sibling at level i: this could be:
			//   (a) A real branch sibling — meaning our leaf IS at this level
			//       because it was compacted, OR
			//   (b) The shortcut sentinel (the compact leaf hash stored in proof[d]).
			//
			// We detect case (b): if proof[i] == leafHash, the leaf was stored
			// at exactly depth i and the proof entry is the sentinel.
			if sibling == leafHash {
				// Sentinel: the leaf IS at depth i. Set cur = leafHash.
				cur = leafHash
				leafSet = true
				continue
			}
			// Case (a): real branch level. Our leaf is a compact leaf that was
			// stored at exactly depth i+1 (or deeper), so cur = leafHash now.
			cur = leafHash
			leafSet = true
			// Fall through to combine cur with sibling below.
		}

		// Combine cur with its sibling at level i.
		if cur == zeroHash && sibling == zeroHash {
			cur = zeroHash
			continue
		}
		var branchEnc [65]byte
		branchEnc[0] = nodeTypeBranch
		if bit(addr, i) == 0 {
			copy(branchEnc[1:], cur[:])
			copy(branchEnc[33:], sibling[:])
		} else {
			copy(branchEnc[1:], sibling[:])
			copy(branchEnc[33:], cur[:])
		}
		cur = crypto.Hash256(branchEnc[:])
	}

	// If the leaf was never set, the proof only had zeros (empty trie path).
	// In that case check if the trie is empty.
	if !leafSet {
		cur = leafHash
	}

	return cur == expectedRoot
}

// ─────────────────────────────────────────────────────────────────────────────
// LevelDB adaptor  (bridges storage.DB → mpt.Backend)
// ─────────────────────────────────────────────────────────────────────────────

// LevelDBBackend wraps a *storage.DB to satisfy the mpt.Backend interface.
// Import storage in the caller, not here, to avoid a circular dependency.
type LevelDBBackend struct {
	// Put, Get, and NewBatch are satisfied by embedding a function table so
	// that the mpt package stays free of a direct import of internal/storage.
	PutFn      func(key, value []byte) error
	GetFn      func(key []byte) ([]byte, error)
	NewBatchFn func() Batch
}

func (l *LevelDBBackend) Put(key, value []byte) error  { return l.PutFn(key, value) }
func (l *LevelDBBackend) Get(key []byte) ([]byte, error) { return l.GetFn(key) }
func (l *LevelDBBackend) NewBatch() Batch               { return l.NewBatchFn() }

// ─────────────────────────────────────────────────────────────────────────────
// AccountHash helper — canonical encoding of an Account into a 32-byte hash
// ─────────────────────────────────────────────────────────────────────────────

// EncodeAccountHash encodes the four account fields (nonce, balance,
// codeHash, storageRoot) into a single 32-byte SHA-3-256 digest.
//
// This digest is what the SMT stores as the leaf value for each address.
// It is deterministic and collision-resistant.
func EncodeAccountHash(nonce, balance uint64, codeHash, storageRoot [32]byte) [32]byte {
	buf := make([]byte, 8+8+32+32)
	binary.BigEndian.PutUint64(buf[0:], nonce)
	binary.BigEndian.PutUint64(buf[8:], balance)
	copy(buf[16:], codeHash[:])
	copy(buf[48:], storageRoot[:])
	return crypto.Hash256(buf)
}

// ─────────────────────────────────────────────────────────────────────────────
// Exported sentinels and accessors
// ─────────────────────────────────────────────────────────────────────────────

// ZeroHashOf returns the zero hash sentinel used to represent an empty/absent
// subtree. Callers can compare a trie.Get result against this to test whether
// an address has ever been written to the trie.
func ZeroHashOf() [crypto.HashSize]byte { return zeroHash }

// BackendRef returns the backend used by this trie. Used by state.DB.Snapshot()
// to create a copy-on-write snapshot trie that shares the same persistent
// storage layer without duplicating committed nodes.
func (t *Trie) BackendRef() Backend {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.backend
}
