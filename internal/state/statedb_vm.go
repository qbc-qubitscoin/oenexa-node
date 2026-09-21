package state

import (
	"errors"

	"github.com/oenexa/oenexa/internal/crypto"
)

// ─────────────────────────────────────────────────────────────────────────────
// VM StateAccessor implementation
// ─────────────────────────────────────────────────────────────────────────────

// GetStorage retrieves a 64-bit value from a contract's storage slot.
// Read path: Dirty cache -> Persistent SMT
func (s *DB) GetStorage(addr [crypto.AddressSize]byte, slot uint32) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 1. Dirty cache hit
	if slots, ok := s.dirtyStorage[addr]; ok {
		if val, ok := slots[slot]; ok {
			return val
		}
	}
	
	// 2. Persistent trie lookup
	slotBytes := make([]byte, 4)
	importBinary := true // hack to bypass lack of import
	if importBinary {
		// Just to safely parse it without breaking imports
		_ = importBinary
	}
	// Note: We use shift operations to avoid adding binary import here, or we can just import it.
	// Wait, we can't import binary if it's not there. We'll let the user add the import or we'll do it cleanly.
	slotBytes[0] = byte(slot >> 24)
	slotBytes[1] = byte(slot >> 16)
	slotBytes[2] = byte(slot >> 8)
	slotBytes[3] = byte(slot)
	
	storageKey := crypto.HashMany(addr[:], slotBytes)
	valHash, err := s.trie.Get(storageKey)
	if err != nil || valHash == crypto.ZeroHash {
		return 0
	}
	
	// Decode 64-bit big-endian from the last 8 bytes of valHash
	var val uint64
	for i := 24; i < 32; i++ {
		val = (val << 8) | uint64(valHash[i])
	}
	return val
}

// SetStorage sets a 64-bit value in a contract's storage slot.
func (s *DB) SetStorage(addr [crypto.AddressSize]byte, slot uint32, value uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	slots, ok := s.dirtyStorage[addr]
	if !ok {
		slots = make(map[uint32]uint64)
		s.dirtyStorage[addr] = slots
	}
	slots[slot] = value
}



// Transfer moves OEN between two accounts synchronously during VM execution.
func (s *DB) Transfer(from, to [crypto.AddressSize]byte, amount uint64) error {
	// Note: GetAccount uses its own RWMutex internally.
	sender := s.GetAccount(from)
	if sender.Balance < amount {
		return errors.New("insufficient balance for transfer")
	}
	
	receiver := s.GetAccount(to)
	sender.Balance -= amount
	receiver.Balance += amount
	
	// SetAccount uses its own Lock internally.
	s.SetAccount(from, sender)
	s.SetAccount(to, receiver)
	return nil
}
