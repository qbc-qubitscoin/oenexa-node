package state

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/vm"
)

// TxResult holds the outcome of applying a single transaction.
//
// Fee split (EIP-1559 model):
//
//	FeeCollected = GasUsed × tx.GasPrice (total deducted from sender)
//	BurnedFee = GasUsed × baseFee (removed from supply — deflationary)
//	ValidatorTip = FeeCollected − BurnedFee (goes to block validator)
type TxResult struct {
	GasUsed      uint64
	FeeCollected uint64 // total fee paid by sender (gasUsed × gasPrice)
	BurnedFee    uint64 // baseFee portion — burned, not credited to anyone
	ValidatorTip uint64 // priority tip credited to the validator
	Success      bool
	Error        error
	ReturnData   []byte
}

// ApplyTransaction executes a transaction against the state.
//
//   - baseFee is the current block's base fee per gas (oenexa).
//     Transactions whose GasPrice < baseFee are rejected.
//   - execVM may be nil for transfer-only operation.
func ApplyTransaction(
	st *DB,
	tx *core.Transaction,
	remainingBlockGas uint64,
	execVM *vm.VM,
	baseFee uint64,
) (*TxResult, error) {

	// ── Intrinsic gas ────────────────────────────────────────────────────────
	intrinsic := IntrinsicGas(tx)
	if intrinsic > tx.GasLimit {
		return nil, errors.New("gas limit below intrinsic cost")
	}
	if tx.GasLimit > remainingBlockGas {
		return nil, errors.New("tx gas limit exceeds the remaining block gas")
	}

	// ── Base-fee check (EIP-1559) ────────────────────────────────────────────
	if tx.GasPrice < baseFee {
		return nil, fmt.Errorf("gas price %d below current base fee %d", tx.GasPrice, baseFee)
	}

	// ── Sender account checks ────────────────────────────────────────────────
	sender := st.GetAccount(tx.From)
	maxCost := tx.GasLimit * tx.GasPrice
	requiredBalance := maxCost
	if tx.Type == core.TxTransfer || tx.Type == core.TxDeploy || tx.Type == core.TxCall || tx.Type == core.TxShield {
		requiredBalance += tx.Amount
	}
	if sender.Balance < requiredBalance {
		return nil, errors.New("insufficient balance")
	}
	if sender.Nonce != tx.Nonce {
		return nil, fmt.Errorf("nonce mismatch: expected %d, got %d", sender.Nonce, tx.Nonce)
	}

	// ── Pre-deduct max gas cost ──────────────────────────────────────────────
	sender.Balance -= maxCost
	sender.Nonce++
	st.SetAccount(tx.From, sender)

	// ── Dispatch by tx type ──────────────────────────────────────────────────
	var gasUsed uint64
	var returnData []byte
	var execErr error

	switch tx.Type {
	case core.TxTransfer:
		gasUsed, execErr = applyTransfer(st, tx)

	case core.TxDeploy:
		gasUsed, returnData, execErr = applyDeploy(st, tx, execVM)

	case core.TxCall:
		gasUsed, returnData, execErr = applyCall(st, tx, execVM, tx.GasLimit-intrinsic)

	case core.TxShield:
		gasUsed, execErr = applyShield(st, tx)

	case core.TxUnshield:
		gasUsed, execErr = applyUnshield(st, tx)

	case core.TxShieldedTransfer:
		gasUsed, execErr = applyShieldedTransfer(st, tx)

	default:
		gasUsed = intrinsic
		execErr = fmt.Errorf("unknown tx type: 0x%02x", tx.Type)
	}

	// ── Refund unused gas ────────────────────────────────────────────────────
	if gasUsed < tx.GasLimit {
		refund := (tx.GasLimit - gasUsed) * tx.GasPrice
		sender = st.GetAccount(tx.From)
		sender.Balance += refund
		st.SetAccount(tx.From, sender)
	}

	// ── Fee split: burn baseFee portion, tip goes to validator ───────────────
	feeCollected := gasUsed * tx.GasPrice

	effectiveBaseFee := baseFee
	burnedFee := gasUsed * effectiveBaseFee
	validatorTip := feeCollected - burnedFee // priority tip = gasPrice − baseFee

	success := execErr == nil
	return &TxResult{
		GasUsed:      gasUsed,
		FeeCollected: feeCollected,
		BurnedFee:    burnedFee,
		ValidatorTip: validatorTip,
		Success:      success,
		Error:        execErr,
		ReturnData:   returnData,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Transfer
// ─────────────────────────────────────────────────────────────────────────────

func applyTransfer(st *DB, tx *core.Transaction) (uint64, error) {
	sender := st.GetAccount(tx.From)
	if sender.Balance < tx.Amount {
		return core.GasTransfer, errors.New("insufficient balance for transfer")
	}
	sender.Balance -= tx.Amount
	st.SetAccount(tx.From, sender)

	recipient := st.GetAccount(tx.To)
	recipient.Balance += tx.Amount
	st.SetAccount(tx.To, recipient)

	return core.GasTransfer, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Deploy
// ─────────────────────────────────────────────────────────────────────────────

func applyDeploy(
	st *DB,
	tx *core.Transaction,
	execVM *vm.VM,
) (uint64, []byte, error) {

	if len(tx.Data) == 0 {
		return core.GasDeploy, nil, errors.New("deploy tx has no bytecode")
	}

	// Contract address = SHA-3-256(deployer ‖ nonce).
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], tx.Nonce)
	contractAddr := crypto.HashMany(tx.From[:], nonceBytes[:])

	codeHash := crypto.Hash256(tx.Data)

	if execVM != nil {
		ctx := context.Background()
		if err := execVM.Deploy(ctx, contractAddr, tx.Data); err != nil {
			return core.GasDeploy, nil, fmt.Errorf("deploy failed: %w", err)
		}
		ec := &vm.ExecutionContext{
			ContractAddr: contractAddr,
			CallerAddr:   tx.From,
			Value:        tx.Amount,
			BlockHeight:  0,
			GasLimit:     tx.GasLimit - core.GasDeploy,
		}
		_, _ = execVM.Call(context.WithValue(ctx, vm.ExecCtxKey{}, ec),
			ec, codeHash, "_init")
	}

	contract := st.GetAccount(contractAddr)
	contract.CodeHash = codeHash
	st.SetAccount(contractAddr, contract)

	if tx.Amount > 0 {
		sender := st.GetAccount(tx.From)
		sender.Balance -= tx.Amount
		st.SetAccount(tx.From, sender)

		contract = st.GetAccount(contractAddr)
		contract.Balance += tx.Amount
		st.SetAccount(contractAddr, contract)
	}

	return core.GasDeploy, contractAddr[:], nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Call
// ─────────────────────────────────────────────────────────────────────────────

func applyCall(
	st *DB,
	tx *core.Transaction,
	execVM *vm.VM,
	gasAvail uint64,
) (uint64, []byte, error) {

	contractAcct := st.GetAccount(tx.To)
	if contractAcct.CodeHash == crypto.ZeroHash {
		return core.GasCall, nil, errors.New("call target is not a contract")
	}

	if tx.Amount > 0 {
		sender := st.GetAccount(tx.From)
		if sender.Balance < tx.Amount {
			return core.GasCall, nil, errors.New("insufficient balance for call value")
		}
		sender.Balance -= tx.Amount
		st.SetAccount(tx.From, sender)

		contractAcct = st.GetAccount(tx.To)
		contractAcct.Balance += tx.Amount
		st.SetAccount(tx.To, contractAcct)
	}

	if execVM == nil {
		return core.GasCall, nil, nil
	}

	funcName := "call"
	var params []uint64
	if len(tx.Data) >= 1 {
		nameLen := int(tx.Data[0])
		if nameLen > 0 && len(tx.Data) >= 1+nameLen {
			funcName = string(tx.Data[1 : 1+nameLen])
			rest := tx.Data[1+nameLen:]
			for i := 0; i+8 <= len(rest); i += 8 {
				params = append(params, binary.BigEndian.Uint64(rest[i:]))
			}
		}
	}

	ec := &vm.ExecutionContext{
		ContractAddr: tx.To,
		CallerAddr:   tx.From,
		Value:        tx.Amount,
		GasLimit:     gasAvail,
	}

	ctx := context.WithValue(context.Background(), vm.ExecCtxKey{}, ec)
	result, err := execVM.Call(ctx, ec, contractAcct.CodeHash, funcName, params...)
	if err != nil {
		return core.GasCall, nil, fmt.Errorf("call reverted: %w", err)
	}

	gasUsed := core.GasCall + ec.GasUsed

	var returnData []byte
	b8 := make([]byte, 8)
	for _, r := range result {
		binary.BigEndian.PutUint64(b8, r)
		returnData = append(returnData, b8...)
	}

	return gasUsed, returnData, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Shield (t-addr -> z-addr)
// ─────────────────────────────────────────────────────────────────────────────

func applyShield(st *DB, tx *core.Transaction) (uint64, error) {
	if tx.Amount == 0 {
		return core.GasShield, errors.New("shield amount must be positive")
	}
	if len(tx.Data) < 32 {
		return core.GasShield, errors.New("invalid note commitment length")
	}
	sender := st.GetAccount(tx.From)
	if sender.Balance < tx.Amount {
		return core.GasShield, errors.New("insufficient balance for shield")
	}
	sender.Balance -= tx.Amount
	st.SetAccount(tx.From, sender)

	st.CreditShieldedPool(tx.Amount)

	var cm [32]byte
	copy(cm[:], tx.Data[:32])
	if _, err := st.AppendNoteCommitment(cm); err != nil {
		sender.Balance += tx.Amount
		st.SetAccount(tx.From, sender)
		_ = st.DebitShieldedPool(tx.Amount)
		return core.GasShield, err
	}
	return core.GasShield, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Unshield (z-addr -> t-addr)
// ─────────────────────────────────────────────────────────────────────────────

func applyUnshield(st *DB, tx *core.Transaction) (uint64, error) {
	if tx.Amount == 0 {
		return core.GasUnshield, errors.New("unshield amount must be positive")
	}
	if len(tx.Data) < 32 {
		return core.GasUnshield, errors.New("invalid nullifier length")
	}
	var nf [32]byte
	copy(nf[:], tx.Data[:32])
	if err := st.DebitShieldedPool(tx.Amount); err != nil {
		return core.GasUnshield, fmt.Errorf("insufficient shielded pool balance: %w", err)
	}
	if err := st.AddNullifier(nf); err != nil {
		st.CreditShieldedPool(tx.Amount)
		return core.GasUnshield, err
	}

	recipient := st.GetAccount(tx.To)
	recipient.Balance += tx.Amount
	st.SetAccount(tx.To, recipient)

	return core.GasUnshield, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Shielded Transfer (z-addr -> z-addr)
// ─────────────────────────────────────────────────────────────────────────────

func applyShieldedTransfer(st *DB, tx *core.Transaction) (uint64, error) {
	if len(tx.Data) < 64 {
		return core.GasShieldedTransfer, errors.New("invalid shielded transfer payload")
	}
	var nf, cm [32]byte
	copy(nf[:], tx.Data[:32])
	copy(cm[:], tx.Data[32:64])

	if err := st.AddNullifier(nf); err != nil {
		return core.GasShieldedTransfer, err
	}
	if _, err := st.AppendNoteCommitment(cm); err != nil {
		st.RemoveNullifier(nf)
		return core.GasShieldedTransfer, err
	}
	return core.GasShieldedTransfer, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// IntrinsicGas
// ─────────────────────────────────────────────────────────────────────────────

// IntrinsicGas returns the base gas cost before execution.
func IntrinsicGas(tx *core.Transaction) uint64 {
	switch tx.Type {
	case core.TxTransfer:
		return core.GasTransfer
	case core.TxDeploy:
		return core.GasDeploy + uint64(len(tx.Data))*68
	case core.TxCall:
		return core.GasCall + uint64(len(tx.Data))*16
	case core.TxShield:
		return core.GasShield
	case core.TxUnshield:
		return core.GasUnshield
	case core.TxShieldedTransfer:
		return core.GasShieldedTransfer
	default:
		return core.GasTransfer
	}
}
