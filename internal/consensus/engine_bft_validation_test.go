package consensus

import (
	"testing"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/mempool"
	"github.com/oenexa/oenexa/internal/state"
)

func setupTestBFT(t *testing.T) (*Engine, *Engine, *crypto.Wallet, *crypto.Wallet) {
	// Val 1 (Proposer for height 1)
	val1Wallet, err := crypto.NewWallet()
	if err != nil {
		t.Fatalf("val1 wallet: %v", err)
	}
	// Val 2 (Follower validator)
	val2Wallet, err := crypto.NewWallet()
	if err != nil {
		t.Fatalf("val2 wallet: %v", err)
	}

	st1 := state.NewStateDB()
	st2 := state.NewStateDB()

	// Give balances to both validators in state
	st1.SetAccount(val1Wallet.Address, &state.Account{Balance: 1_000_000 * core.OneOEN})
	st1.SetAccount(val2Wallet.Address, &state.Account{Balance: 1_000_000 * core.OneOEN})
	st2.SetAccount(val1Wallet.Address, &state.Account{Balance: 1_000_000 * core.OneOEN})
	st2.SetAccount(val2Wallet.Address, &state.Account{Balance: 1_000_000 * core.OneOEN})

	cfg := core.DefaultGenesisConfig(val1Wallet.Address)
	cfg.Allocations = map[[crypto.AddressSize]byte]uint64{
		val1Wallet.Address: 1_000_000 * core.OneOEN,
		val2Wallet.Address: 1_000_000 * core.OneOEN,
	}
	genesis := cfg.Build()

	// 2-validator set: Val2 at idx 0 (h%2==0), Val1 at idx 1 (h%2==1)
	vs1, _ := NewValidatorSet([]*Validator{
		{Address: val2Wallet.Address, PublicKey: val2Wallet.PublicKey, VotingPower: 10},
		{Address: val1Wallet.Address, PublicKey: val1Wallet.PublicKey, VotingPower: 10},
	})
	vs2, _ := NewValidatorSet([]*Validator{
		{Address: val2Wallet.Address, PublicKey: val2Wallet.PublicKey, VotingPower: 10},
		{Address: val1Wallet.Address, PublicKey: val1Wallet.PublicKey, VotingPower: 10},
	})

	pool1 := mempool.New(100)
	pool2 := mempool.New(100)

	e1 := NewEngine(val1Wallet.Address, val1Wallet.PublicKey, val1Wallet.PrivateKey, vs1, st1, pool1, genesis, nil, nil)
	e2 := NewEngine(val2Wallet.Address, val2Wallet.PublicKey, val2Wallet.PrivateKey, vs2, st2, pool2, genesis, nil, nil)

	return e1, e2, val1Wallet, val2Wallet
}

func TestEngine_ProcessProposal_Valid(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	// Proposer (e1) produces a valid block
	blk, snap, err := e1.buildBlock(1)
	if err != nil {
		t.Fatalf("buildBlock: %v", err)
	}
	_ = snap

	// Follower (e2) receives and validates the proposal
	if err := e2.ProcessProposal(blk); err != nil {
		t.Fatalf("ProcessProposal failed for valid block: %v", err)
	}

	if e2.step != StepPrevote {
		t.Fatalf("Expected step StepPrevote, got %d", e2.step)
	}
}

func TestEngine_ProcessProposal_InvalidProposer(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	fakeProposer, _ := crypto.NewWallet()
	blk, _, _ := e1.buildBlock(1)

	// Tamper: claim it was proposed by a different address
	blk.Header.ValidatorAddr = fakeProposer.Address
	_ = blk.SignHeader(fakeProposer.PrivateKey)

	err := e2.ProcessProposal(blk)
	if err == nil {
		t.Fatalf("Expected error for invalid proposer, got nil")
	}
}

func TestEngine_ProcessProposal_InvalidSignature(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	blk, _, _ := e1.buildBlock(1)

	// Corrupt signature
	blk.Signature = make([]byte, crypto.SignatureSize)

	err := e2.ProcessProposal(blk)
	if err == nil {
		t.Fatalf("Expected error for corrupt signature, got nil")
	}
}

func TestEngine_ProcessProposal_InvalidPrevHash(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	blk, _, _ := e1.buildBlock(1)

	// Corrupt PrevHash
	blk.Header.PrevHash[0] ^= 0xff
	_ = blk.SignHeader(e1.validatorPriv)

	err := e2.ProcessProposal(blk)
	if err == nil {
		t.Fatalf("Expected error for corrupt PrevHash, got nil")
	}
}

func TestEngine_ProcessProposal_InvalidStateRoot(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	blk, _, _ := e1.buildBlock(1)

	// Corrupt StateRoot (e.g. Byzantine proposer trying to forge fake state)
	blk.Header.StateRoot[0] ^= 0xff
	_ = blk.SignHeader(e1.validatorPriv)

	err := e2.ProcessProposal(blk)
	if err == nil {
		t.Fatalf("Expected error for corrupt StateRoot, got nil")
	}
}

func TestEngine_ProcessProposal_InvalidMerkleRoot(t *testing.T) {
	e1, e2, _, _ := setupTestBFT(t)

	blk, _, _ := e1.buildBlock(1)

	// Corrupt MerkleRoot
	blk.Header.MerkleRoot[0] ^= 0xff
	_ = blk.SignHeader(e1.validatorPriv)

	err := e2.ProcessProposal(blk)
	if err == nil {
		t.Fatalf("Expected error for corrupt MerkleRoot, got nil")
	}
}

func TestEngine_ProcessVote_EquivocationSlashing(t *testing.T) {
	e1, _, _, val2Wallet := setupTestBFT(t)

	hashA := [crypto.HashSize]byte{1, 2, 3}
	hashB := [crypto.HashSize]byte{4, 5, 6}

	// Val2 votes for HashA
	voteA := &Vote{
		Type:      VotePrevote,
		Height:    1,
		Round:     0,
		BlockHash: hashA,
		Voter:     val2Wallet.Address,
		PublicKey: val2Wallet.PublicKey,
	}
	if err := voteA.Sign(val2Wallet.PrivateKey); err != nil {
		t.Fatalf("sign voteA: %v", err)
	}

	if err := e1.ProcessVote(voteA); err != nil {
		t.Fatalf("ProcessVote voteA failed: %v", err)
	}

	initialBalance := e1.state.GetAccount(val2Wallet.Address).Balance
	initialPower := e1.validatorSet.Get(val2Wallet.Address).VotingPower

	// Val2 equivocates (submits a conflicting vote for HashB at the same height & round)
	voteB := &Vote{
		Type:      VotePrevote,
		Height:    1,
		Round:     0,
		BlockHash: hashB,
		Voter:     val2Wallet.Address,
		PublicKey: val2Wallet.PublicKey,
	}
	if err := voteB.Sign(val2Wallet.PrivateKey); err != nil {
		t.Fatalf("sign voteB: %v", err)
	}

	err := e1.ProcessVote(voteB)
	if err == nil {
		t.Fatalf("Expected error for equivocation, got nil")
	}

	// Verify slashing: voting power reduced and 10% balance slashed
	newPower := e1.validatorSet.Get(val2Wallet.Address).VotingPower
	if newPower >= initialPower {
		t.Fatalf("Expected voting power to be slashed from %d, got %d", initialPower, newPower)
	}

	newBalance := e1.state.GetAccount(val2Wallet.Address).Balance
	expectedBalance := initialBalance - (initialBalance / 10)
	if newBalance != expectedBalance {
		t.Fatalf("Expected balance to be slashed to %d, got %d", expectedBalance, newBalance)
	}
}

func TestValidatorSet_Slash(t *testing.T) {
	w, _ := crypto.NewWallet()
	vs, err := NewValidatorSet([]*Validator{
		{Address: w.Address, PublicKey: w.PublicKey, VotingPower: 100},
	})
	if err != nil {
		t.Fatalf("NewValidatorSet: %v", err)
	}

	// Slash 30 power
	if !vs.Slash(w.Address, 30) {
		t.Fatalf("Slash returned false")
	}
	if vs.TotalPower() != 70 {
		t.Fatalf("Expected total power 70, got %d", vs.TotalPower())
	}
	if vs.Get(w.Address).VotingPower != 70 {
		t.Fatalf("Expected voting power 70, got %d", vs.Get(w.Address).VotingPower)
	}

	// Slash unknown address
	unknown := [crypto.AddressSize]byte{9, 9, 9}
	if vs.Slash(unknown, 10) {
		t.Fatalf("Expected false when slashing unknown address")
	}
}
