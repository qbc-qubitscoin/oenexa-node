package state

import (
	"testing"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/shielded"
)

func TestStateDB_ShieldedMethods(t *testing.T) {
	st := NewStateDB()

	// 1. Initial shielded pool is 0
	if bal := st.GetShieldedBalance(); bal != 0 {
		t.Fatalf("expected 0 initial shielded balance, got %d", bal)
	}

	// 2. Credit and Debit
	st.CreditShieldedPool(1000)
	if bal := st.GetShieldedBalance(); bal != 1000 {
		t.Fatalf("expected 1000 shielded balance, got %d", bal)
	}

	if err := st.DebitShieldedPool(400); err != nil {
		t.Fatalf("DebitShieldedPool failed: %v", err)
	}
	if bal := st.GetShieldedBalance(); bal != 600 {
		t.Fatalf("expected 600 shielded balance, got %d", bal)
	}

	if err := st.DebitShieldedPool(700); err == nil {
		t.Fatal("expected error on debiting more than balance")
	}

	// 3. Nullifier operations
	var nf [32]byte
	nf[0] = 0xAA
	if st.HasNullifier(nf) {
		t.Fatal("expected nullifier to be unspent initially")
	}
	if err := st.AddNullifier(nf); err != nil {
		t.Fatalf("AddNullifier failed: %v", err)
	}
	if !st.HasNullifier(nf) {
		t.Fatal("expected nullifier to be spent after AddNullifier")
	}
	if err := st.AddNullifier(nf); err == nil {
		t.Fatal("expected error adding duplicate nullifier")
	}

	// 4. Commitment tree operations
	var cm [32]byte
	cm[0] = 0xBB
	idx, err := st.AppendNoteCommitment(cm)
	if err != nil {
		t.Fatalf("AppendNoteCommitment failed: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected index 0, got %d", idx)
	}
	if st.CommitmentTree() == nil {
		t.Fatal("expected non-nil CommitmentTree")
	}
	root := st.CommitmentTreeRoot()

	// 5. Total transparent & total supply
	addr1 := makeAddr(50)
	addr2 := makeAddr(51)
	st.SetAccount(addr1, &Account{Balance: 2000})
	st.SetAccount(addr2, &Account{Balance: 3000})

	if trans := st.TotalTransparentSupply(); trans != 5000 {
		t.Fatalf("expected 5000 transparent supply, got %d", trans)
	}
	if total := st.TotalSupply(); total != 5600 {
		t.Fatalf("expected 5600 total supply (5000 + 600), got %d", total)
	}

	// 6. CommitRoot includes shielded pool and tree
	cRoot := st.CommitRoot()
	if cRoot == [32]byte{} {
		t.Fatal("unexpected zero commit root")
	}

	// 7. Snapshot deep copy
	snap := st.Snapshot()
	if snap.GetShieldedBalance() != 600 {
		t.Fatalf("snapshot shielded balance mismatch: %d", snap.GetShieldedBalance())
	}
	if !snap.HasNullifier(nf) {
		t.Fatal("snapshot missing nullifier")
	}
	if snap.CommitmentTreeRoot() != root {
		t.Fatal("snapshot commitment tree root mismatch")
	}

	// Mutate original and ensure snapshot is isolated
	st.CreditShieldedPool(50)
	if snap.GetShieldedBalance() != 600 {
		t.Fatal("snapshot was mutated by original state change")
	}

	// 8. Apply
	stTarget := NewStateDB()
	stTarget.Apply(st)
	if stTarget.GetShieldedBalance() != 650 {
		t.Fatalf("applied shielded balance mismatch: %d", stTarget.GetShieldedBalance())
	}
	if !stTarget.HasNullifier(nf) {
		t.Fatal("applied state missing nullifier")
	}
	if stTarget.CommitmentTreeRoot() != root {
		t.Fatal("applied state commitment tree root mismatch")
	}
	stTarget.Apply(nil)
	stTarget.Apply(stTarget)
}

func TestTransition_Shield_Operations(t *testing.T) {
	st := NewStateDB()
	senderAddr := makeAddr(60)
	st.SetAccount(senderAddr, &Account{Balance: 200_000, Nonce: 0})

	var cm [32]byte
	cm[0] = 0x11

	// 1. Direct applyShield error checks
	// Amount == 0
	txZeroAmt := &core.Transaction{Type: core.TxShield, From: senderAddr, Amount: 0, Data: cm[:]}
	if _, err := applyShield(st, txZeroAmt); err == nil {
		t.Fatal("expected error on 0 shield amount")
	}

	// Invalid note commitment length
	txBadCm := &core.Transaction{Type: core.TxShield, From: senderAddr, Amount: 100, Data: []byte("too-short")}
	if _, err := applyShield(st, txBadCm); err == nil {
		t.Fatal("expected error on invalid commitment length")
	}

	// Insufficient balance
	txTooMuch := &core.Transaction{Type: core.TxShield, From: senderAddr, Amount: 500_000, Data: cm[:]}
	if _, err := applyShield(st, txTooMuch); err == nil {
		t.Fatal("expected error on insufficient balance for shield")
	}

	// Append failure
	origMax := shielded.MaxTreeLeaves
	shielded.MaxTreeLeaves = 0
	txValid := &core.Transaction{Type: core.TxShield, From: senderAddr, Amount: 1000, Data: cm[:]}
	if _, err := applyShield(st, txValid); err == nil {
		t.Fatal("expected error when commitment tree is full")
	}
	shielded.MaxTreeLeaves = origMax

	// 2. Direct applyShield success
	gasUsed, err := applyShield(st, txValid)
	if err != nil {
		t.Fatalf("applyShield failed: %v", err)
	}
	if gasUsed != core.GasShield {
		t.Fatalf("expected gas %d, got %d", core.GasShield, gasUsed)
	}
	if st.GetBalance(senderAddr) != 199_000 {
		t.Fatalf("sender balance mismatch: %d", st.GetBalance(senderAddr))
	}
	if st.GetShieldedBalance() != 1000 {
		t.Fatalf("shielded pool balance mismatch: %d", st.GetShieldedBalance())
	}

	// 3. Via ApplyTransaction
	dummyPK := make([]byte, crypto.PublicKeySize)
	txShield := core.NewShield(senderAddr, dummyPK, 0, 2000, 1, cm, nil)
	txShield.GasLimit = core.GasShield + 10_000
	res, err := ApplyTransaction(st, txShield, 1_000_000, nil, 1)
	if err != nil {
		t.Fatalf("ApplyTransaction(TxShield) failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("ApplyTransaction(TxShield) returned failure: %v", res.Error)
	}
	if st.GetShieldedBalance() != 3000 {
		t.Fatalf("expected 3000 shielded balance, got %d", st.GetShieldedBalance())
	}
}

func TestTransition_Unshield_Operations(t *testing.T) {
	st := NewStateDB()
	recipientAddr := makeAddr(70)
	senderGasAddr := makeAddr(71)
	st.SetAccount(senderGasAddr, &Account{Balance: 200_000, Nonce: 0})
	st.CreditShieldedPool(5000)

	var nf [32]byte
	nf[0] = 0x22

	// 1. Direct applyUnshield error checks
	// Amount == 0
	txZeroAmt := &core.Transaction{Type: core.TxUnshield, To: recipientAddr, Amount: 0, Data: nf[:]}
	if _, err := applyUnshield(st, txZeroAmt); err == nil {
		t.Fatal("expected error on 0 unshield amount")
	}

	// Bad nullifier length
	txBadNf := &core.Transaction{Type: core.TxUnshield, To: recipientAddr, Amount: 100, Data: []byte("short")}
	if _, err := applyUnshield(st, txBadNf); err == nil {
		t.Fatal("expected error on bad nullifier length")
	}

	// Insufficient shielded pool
	txTooMuch := &core.Transaction{Type: core.TxUnshield, To: recipientAddr, Amount: 99_999, Data: nf[:]}
	if _, err := applyUnshield(st, txTooMuch); err == nil {
		t.Fatal("expected error on insufficient shielded pool balance")
	}

	// 2. Direct applyUnshield success
	txValid := &core.Transaction{Type: core.TxUnshield, To: recipientAddr, Amount: 1500, Data: nf[:]}
	gasUsed, err := applyUnshield(st, txValid)
	if err != nil {
		t.Fatalf("applyUnshield failed: %v", err)
	}
	if gasUsed != core.GasUnshield {
		t.Fatalf("expected gas %d, got %d", core.GasUnshield, gasUsed)
	}
	if st.GetBalance(recipientAddr) != 1500 {
		t.Fatalf("recipient balance mismatch: %d", st.GetBalance(recipientAddr))
	}
	if st.GetShieldedBalance() != 3500 {
		t.Fatalf("shielded pool mismatch: %d", st.GetShieldedBalance())
	}
	if !st.HasNullifier(nf) {
		t.Fatal("nullifier not marked as spent")
	}

	// Double-unshield fails (AddNullifier error branch)
	if _, err := applyUnshield(st, txValid); err == nil {
		t.Fatal("expected error on double-unshield with same nullifier")
	}

	// 3. Via ApplyTransaction
	var nf2 [32]byte
	nf2[0] = 0x33
	dummyPK := make([]byte, crypto.PublicKeySize)
	txUnshield := core.NewUnshield(recipientAddr, dummyPK, 0, 1000, 1, nf2, nil)
	txUnshield.From = senderGasAddr
	txUnshield.GasLimit = core.GasUnshield + 5000
	res, err := ApplyTransaction(st, txUnshield, 1_000_000, nil, 1)
	if err != nil {
		t.Fatalf("ApplyTransaction(TxUnshield) failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("ApplyTransaction(TxUnshield) returned failure: %v", res.Error)
	}
	if st.GetBalance(recipientAddr) != 2500 {
		t.Fatalf("expected recipient balance 2500, got %d", st.GetBalance(recipientAddr))
	}
	if st.GetShieldedBalance() != 2500 {
		t.Fatalf("expected shielded pool 2500, got %d", st.GetShieldedBalance())
	}
}

func TestTransition_ShieldedTransfer_Operations(t *testing.T) {
	st := NewStateDB()
	senderGasAddr := makeAddr(80)
	st.SetAccount(senderGasAddr, &Account{Balance: 200_000, Nonce: 0})

	var nf, cm [32]byte
	nf[0] = 0x44
	cm[0] = 0x55
	payload := append(nf[:], cm[:]...)

	// 1. Direct applyShieldedTransfer errors
	// Too short
	txShort := &core.Transaction{Type: core.TxShieldedTransfer, Data: []byte("too-short")}
	if _, err := applyShieldedTransfer(st, txShort); err == nil {
		t.Fatal("expected error on short payload")
	}

	// Tree full error
	origMax := shielded.MaxTreeLeaves
	shielded.MaxTreeLeaves = 0
	txValid := &core.Transaction{Type: core.TxShieldedTransfer, Data: payload}
	if _, err := applyShieldedTransfer(st, txValid); err == nil {
		t.Fatal("expected error when commitment tree is full")
	}
	shielded.MaxTreeLeaves = origMax

	// 2. Direct success
	gasUsed, err := applyShieldedTransfer(st, txValid)
	if err != nil {
		t.Fatalf("applyShieldedTransfer failed: %v", err)
	}
	if gasUsed != core.GasShieldedTransfer {
		t.Fatalf("expected gas %d, got %d", core.GasShieldedTransfer, gasUsed)
	}
	if !st.HasNullifier(nf) {
		t.Fatal("nullifier should be marked spent")
	}

	// Double-spend nullifier error
	if _, err := applyShieldedTransfer(st, txValid); err == nil {
		t.Fatal("expected error on duplicate nullifier")
	}

	// 3. Via ApplyTransaction
	var nf2, cm2 [32]byte
	nf2[0] = 0x66
	cm2[0] = 0x77
	payload2 := append(nf2[:], cm2[:]...)
	txST := &core.Transaction{
		Type:     core.TxShieldedTransfer,
		From:     senderGasAddr,
		Data:     payload2,
		GasLimit: core.GasShieldedTransfer + 5000,
		GasPrice: 1,
		Nonce:    0,
	}
	res, err := ApplyTransaction(st, txST, 1_000_000, nil, 1)
	if err != nil {
		t.Fatalf("ApplyTransaction(TxShieldedTransfer) failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("ApplyTransaction(TxShieldedTransfer) failed: %v", res.Error)
	}
	if !st.HasNullifier(nf2) {
		t.Fatal("nullifier2 should be marked spent")
	}
}

func TestTransition_IntrinsicGas_Shielded(t *testing.T) {
	txShield := &core.Transaction{Type: core.TxShield}
	if IntrinsicGas(txShield) != core.GasShield {
		t.Errorf("IntrinsicGas(TxShield): want %d, got %d", core.GasShield, IntrinsicGas(txShield))
	}

	txUnshield := &core.Transaction{Type: core.TxUnshield}
	if IntrinsicGas(txUnshield) != core.GasUnshield {
		t.Errorf("IntrinsicGas(TxUnshield): want %d, got %d", core.GasUnshield, IntrinsicGas(txUnshield))
	}

	txShieldedTransfer := &core.Transaction{Type: core.TxShieldedTransfer}
	if IntrinsicGas(txShieldedTransfer) != core.GasShieldedTransfer {
		t.Errorf("IntrinsicGas(TxShieldedTransfer): want %d, got %d", core.GasShieldedTransfer, IntrinsicGas(txShieldedTransfer))
	}
}

func TestApplyBlock_ShieldedTransactions(t *testing.T) {
	st := NewStateDB()
	validatorAddr := makeAddr(90)
	senderAddr := makeAddr(91)
	recipientAddr := makeAddr(92)
	st.SetAccount(senderAddr, &Account{Balance: 100_000, Nonce: 0})

	var cm [32]byte
	cm[0] = 0x88
	dummyPK := make([]byte, crypto.PublicKeySize)
	tx1 := core.NewShield(senderAddr, dummyPK, 0, 10_000, 1, cm, nil)
	tx1.GasLimit = core.GasShield + 1000

	var nf [32]byte
	nf[0] = 0x99
	tx2 := core.NewUnshield(recipientAddr, dummyPK, 1, 4000, 1, nf, nil)
	tx2.From = senderAddr
	tx2.GasLimit = core.GasUnshield + 1000

	blk := &core.Block{
		Header: core.BlockHeader{Height: 1, BaseFee: 1},
		Txs:    []*core.Transaction{tx1, tx2},
	}

	res, err := ApplyBlock(st, blk, validatorAddr, nil)
	if err != nil {
		t.Fatalf("ApplyBlock failed: %v", err)
	}
	if res.ShieldedBalance != 6000 {
		t.Fatalf("expected 6000 shielded balance in block result, got %d", res.ShieldedBalance)
	}
	if st.GetShieldedBalance() != 6000 {
		t.Fatalf("expected 6000 shielded balance in state, got %d", st.GetShieldedBalance())
	}
	if st.GetBalance(recipientAddr) != 4000 {
		t.Fatalf("expected 4000 recipient balance, got %d", st.GetBalance(recipientAddr))
	}
}
