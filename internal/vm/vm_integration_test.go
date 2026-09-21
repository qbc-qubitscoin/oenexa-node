package vm_test

import (
	"context"
	"testing"

	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/state"
	"github.com/oenexa/oenexa/internal/vm"
)

func TestVMIntegration_StateDB(t *testing.T) {
	// Initialize the cryptographic StateDB
	st := state.NewStateDB()
	
	// Create the VM
	ctx := context.Background()
	execVM, err := vm.NewVM(ctx)
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}
	defer execVM.Close(ctx)

	// Addresses
	callerAddr := crypto.Hash256([]byte("caller"))
	contractAddr := crypto.Hash256([]byte("contract"))

	// 1. Deploy the Counter contract
	st.SetAccount(contractAddr, &state.Account{Balance: 0})
	if err := execVM.Deploy(ctx, contractAddr, vm.CounterContractWASM); err != nil {
		t.Fatalf("failed to deploy contract: %v", err)
	}
	
	// Ensure codeHash was computed (this would normally happen in ApplyTransaction)
	codeHash := crypto.Hash256(vm.CounterContractWASM)
	contractAcct := st.GetAccount(contractAddr)
	contractAcct.CodeHash = codeHash
	st.SetAccount(contractAddr, contractAcct)

	// 2. Execute `increment` function via VM
	// We pass the StateDB as the StateAccessor in the context.
	ec := &vm.ExecutionContext{
		State:        st,
		ContractAddr: contractAddr,
		CallerAddr:   callerAddr,
		GasLimit:     1_000_000,
	}
	execCtx := vm.WithExecCtx(ctx, ec)

	// The counter should start at 0 and become 1.
	_, err = execVM.Call(execCtx, ec, codeHash, "increment")
	if err != nil {
		t.Fatalf("call increment failed: %v", err)
	}

	// 3. Verify that the VM correctly updated the actual StateDB.
	// Slot 0 should now be 1.
	val := st.GetStorage(contractAddr, 0)
	if val != 1 {
		t.Fatalf("expected storage slot 0 to be 1, got %d", val)
	}

	// 4. Execute `increment` again.
	_, err = execVM.Call(execCtx, ec, codeHash, "increment")
	if err != nil {
		t.Fatalf("call increment 2 failed: %v", err)
	}

	val = st.GetStorage(contractAddr, 0)
	if val != 2 {
		t.Fatalf("expected storage slot 0 to be 2, got %d", val)
	}

	// 5. Execute `get_count` function to read from StateDB.
	results, err := execVM.Call(execCtx, ec, codeHash, "get_count")
	if err != nil {
		t.Fatalf("call get_count failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 return value, got %d", len(results))
	}
	if results[0] != 2 {
		t.Fatalf("expected get_count to return 2, got %d", results[0])
	}
	
	// 6. Test Gas Consumption
	if ec.GasUsed == 0 {
		t.Fatalf("expected gas to be consumed, got 0")
	}
	t.Logf("Total gas used for 2 increments + 1 read: %d", ec.GasUsed)
}
