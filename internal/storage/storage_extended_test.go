package storage

import (
	"os"
	"testing"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/state"
)

func TestStorageExtended(t *testing.T) {
	// Create a temp db
	path, _ := os.MkdirTemp("", "testdb_ext")
	defer os.RemoveAll(path)
	db, _ := Open(path)

	bs := NewBlockStore(db)

	// Test encodeBlock error by passing a block with a nil transaction
	err := bs.PutBlock(&core.Block{Txs: []*core.Transaction{nil}})
	if err == nil {
		t.Fatal("expected error on encodeBlock with nil tx")
	}

	hash := crypto.Hash256([]byte("test"))

	// Put a bad block gob to trigger decodeBlock error
	db.Put(blockKey(hash), []byte("badgob"))
	_, err = bs.GetBlock(hash)
	if err == nil {
		t.Fatal("expected decodeBlock error")
	}

	// Put a bad gob for GetBlockByHeight
	db.Put(heightKey(10), hash[:])
	_, err = bs.GetBlockByHeight(10)
	if err == nil {
		t.Fatal("expected decodeBlock error in GetBlockByHeight")
	}

	// Test HasBlock on valid key
	has, _ := bs.HasBlock(hash)
	if !has {
		t.Fatal("expected HasBlock to be true")
	}

	// Test ErrNotFound for GetBlock and GetBlockByHeight
	notFoundHash := crypto.Hash256([]byte("missing"))
	_, err = bs.GetBlock(notFoundHash)
	if err != ErrNotFound {
		t.Fatal("expected ErrNotFound for missing block")
	}

	_, err = bs.GetBlockByHeight(999)
	if err != ErrNotFound {
		t.Fatal("expected ErrNotFound for missing block by height")
	}

	// db Close to trigger db errors
	db.Close()

	validBlk := &core.Block{}
	err = bs.PutBlock(validBlk)
	if err == nil {
		t.Fatal("expected error when db is closed")
	}

	err = bs.UpdateTip(1, hash)
	if err == nil {
		t.Fatal("expected error when db is closed")
	}

	_, err = bs.GetBlock(hash)
	if err == nil {
		t.Fatal("expected error when db is closed")
	}

	_, err = bs.GetBlockByHeight(10)
	if err == nil {
		t.Fatal("expected error when db is closed")
	}

	_, err = bs.HasBlock(hash)
	if err == nil {
		t.Fatal("expected error when db is closed")
	}
}

func TestDBOpenError(t *testing.T) {
	// open a bad path (using a file instead of directory to trigger leveldb error)
	f, _ := os.CreateTemp("", "bad_db")
	f.Close()
	defer os.Remove(f.Name())
	_, err := Open(f.Name())
	if err == nil {
		t.Fatal("expected error opening invalid path")
	}
}

func TestDBIterPrefixError(t *testing.T) {
	path, _ := os.MkdirTemp("", "testdb_iter")
	defer os.RemoveAll(path)
	db, _ := Open(path)
	db.Put([]byte("prefix_1"), []byte("val1"))

	// trigger fn returning error
	err := db.IterPrefix([]byte("prefix_"), func(k, v []byte) error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected error from fn")
	}
	db.Close()
}

func TestStateStoreExtended(t *testing.T) {
	path, _ := os.MkdirTemp("", "testdb_state")
	defer os.RemoveAll(path)
	db, _ := Open(path)
	ss := NewStateStore(db)

	// Save an empty state — should succeed and write the trie root meta-key.
	st := state.NewStateDB()
	if err := ss.SaveState(st); err != nil {
		t.Fatalf("SaveState on empty state: %v", err)
	}

	// LoadState on a fresh (empty) state must return a usable StateDB.
	loaded, err := ss.LoadState()
	if err != nil {
		t.Fatalf("LoadState after empty save: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadState returned nil StateDB")
	}

	// The loaded state should report zero dirty accounts (nothing was written).
	if n := loaded.Len(); n != 0 {
		t.Errorf("expected 0 dirty accounts, got %d", n)
	}

	db.Close()

	// LoadState on a closed DB must return an error (cannot read root key).
	_, err = ss.LoadState()
	if err == nil {
		t.Fatal("expected error loading state on closed db")
	}
}
