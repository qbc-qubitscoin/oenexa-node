package shielded

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/sha3"
)

const TreeDepth = 32

var MaxTreeLeaves = uint64(1) << TreeDepth

// emptyRoots stores canonical empty sub-tree roots at each depth level.
var emptyRoots [TreeDepth + 1][32]byte

func init() {
	// Base level 0 is empty leaf (all zeros)
	for d := 0; d < TreeDepth; d++ {
		emptyRoots[d+1] = hashNodes(emptyRoots[d], emptyRoots[d])
	}
}

func hashNodes(left, right [32]byte) [32]byte {
	h := sha3.New256()
	h.Write([]byte("OENEXA_TREE_NODE"))
	h.Write(left[:])
	h.Write(right[:])
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// NoteCommitmentTree is an incremental Merkle tree storing all historical note commitments.
type NoteCommitmentTree struct {
	mu     sync.RWMutex
	leaves [][32]byte
}

// NewNoteCommitmentTree initializes an empty commitment accumulator tree.
func NewNoteCommitmentTree() *NoteCommitmentTree {
	return &NoteCommitmentTree{
		leaves: make([][32]byte, 0, 1024),
	}
}

// Append adds a new note commitment to the tree and returns its leaf index.
func (t *NoteCommitmentTree) Append(cm [32]byte) (uint32, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if uint64(len(t.leaves)) >= MaxTreeLeaves {
		return 0, errors.New("shielded: commitment tree is full")
	}
	idx := uint32(len(t.leaves))
	t.leaves = append(t.leaves, cm)
	return idx, nil
}

// LeafCount returns the number of commitments currently in the tree.
func (t *NoteCommitmentTree) LeafCount() uint32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return uint32(len(t.leaves))
}

// Clone returns an independent deep copy of the commitment tree.
func (t *NoteCommitmentTree) Clone() *NoteCommitmentTree {
	t.mu.RLock()
	defer t.mu.RUnlock()
	leavesCopy := make([][32]byte, len(t.leaves))
	copy(leavesCopy, t.leaves)
	return &NoteCommitmentTree{
		leaves: leavesCopy,
	}
}

// Root computes the current Merkle root of the commitment tree.
func (t *NoteCommitmentTree) Root() [32]byte {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.leaves) == 0 {
		return emptyRoots[TreeDepth]
	}

	curr := make([][32]byte, len(t.leaves))
	copy(curr, t.leaves)

	for d := 0; d < TreeDepth; d++ {
		var next [][32]byte
		for i := 0; i < len(curr); i += 2 {
			left := curr[i]
			var right [32]byte
			if i+1 < len(curr) {
				right = curr[i+1]
			} else {
				right = emptyRoots[d]
			}
			next = append(next, hashNodes(left, right))
		}
		curr = next
	}
	return curr[0]
}

// Witness generates an authentication path (Merkle proof) for the leaf at index.
func (t *NoteCommitmentTree) Witness(index uint32) ([][32]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if int(index) >= len(t.leaves) {
		return nil, fmt.Errorf("shielded: index %d out of bounds (%d leaves)", index, len(t.leaves))
	}

	path := make([][32]byte, TreeDepth)
	levelLeaves := make([][32]byte, len(t.leaves))
	copy(levelLeaves, t.leaves)
	currIdx := index

	for d := 0; d < TreeDepth; d++ {
		isRight := (currIdx % 2) == 1
		var sibling [32]byte
		if isRight {
			sibling = levelLeaves[currIdx-1]
		} else {
			if int(currIdx+1) < len(levelLeaves) {
				sibling = levelLeaves[currIdx+1]
			} else {
				sibling = emptyRoots[d]
			}
		}
		path[d] = sibling

		// Step to next level
		var next [][32]byte
		for i := 0; i < len(levelLeaves); i += 2 {
			left := levelLeaves[i]
			var right [32]byte
			if i+1 < len(levelLeaves) {
				right = levelLeaves[i+1]
			} else {
				right = emptyRoots[d]
			}
			next = append(next, hashNodes(left, right))
		}
		levelLeaves = next
		currIdx /= 2
	}
	return path, nil
}

// VerifyWitness verifies that a commitment exists at index under the given root.
func VerifyWitness(cm [32]byte, index uint32, root [32]byte, path [][32]byte) bool {
	if len(path) != TreeDepth {
		return false
	}
	curr := cm
	currIdx := index
	for d := 0; d < TreeDepth; d++ {
		sibling := path[d]
		if (currIdx % 2) == 1 {
			curr = hashNodes(sibling, curr)
		} else {
			curr = hashNodes(curr, sibling)
		}
		currIdx /= 2
	}
	return curr == root
}
