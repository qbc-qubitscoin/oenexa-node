package vm

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

// Gas costs for host functions.
const (
	gasCostGet         uint64 = 200
	gasCostSet         uint64 = 500
	gasCostLog         uint64 = 100
	gasCostCaller      uint64 = 50
	gasCostBlockHeight uint64 = 10
	gasCostValue       uint64 = 10
	gasCostTransfer    uint64 = 1000
	gasCostBalance     uint64 = 100
)

// hostGet implements env.oen_get(slot i32) -> i64
func hostGet() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil || ec.State == nil {
			stack[0] = 0
			return
		}
		if err := ec.UseGas(gasCostGet); err != nil {
			panic(ErrOutOfGas)
		}
		slot := uint32(stack[0])
		stack[0] = ec.State.GetStorage(ec.ContractAddr, slot)
	}
}

// hostSet implements env.oen_set(slot i32, value i64)
func hostSet() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil || ec.State == nil {
			return
		}
		if ec.ReadOnly {
			panic("write to storage in a read-only context")
		}
		if err := ec.UseGas(gasCostSet); err != nil {
			panic(ErrOutOfGas)
		}
		slot := uint32(stack[0])
		value := stack[1]
		ec.State.SetStorage(ec.ContractAddr, slot, value)
	}
}

// hostTransfer implements env.oen_transfer(to_ptr i32, amount i64)
func hostTransfer() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil || ec.State == nil {
			stack[0] = 0
			return
		}
		if ec.ReadOnly {
			panic("transfer in a read-only context")
		}
		if err := ec.UseGas(gasCostTransfer); err != nil {
			panic(ErrOutOfGas)
		}
		
		mem := mod.Memory()
		if mem == nil {
			panic("no memory")
		}
		
		toPtr := uint32(stack[0])
		amount := stack[1]
		
		b, ok := mem.Read(toPtr, 32)
		if !ok || len(b) != 32 {
			panic("invalid address pointer")
		}
		
		var toAddr [32]byte
		copy(toAddr[:], b)
		
		if err := ec.State.Transfer(ec.ContractAddr, toAddr, amount); err != nil {
			stack[0] = 0 // fail
		} else {
			stack[0] = 1 // success
		}
	}
}

// hostBalance implements env.oen_balance(addr_ptr i32) -> i64
func hostBalance() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil || ec.State == nil {
			stack[0] = 0
			return
		}
		if err := ec.UseGas(gasCostBalance); err != nil {
			panic(ErrOutOfGas)
		}
		
		mem := mod.Memory()
		if mem == nil {
			panic("no memory")
		}
		
		addrPtr := uint32(stack[0])
		b, ok := mem.Read(addrPtr, 32)
		if !ok || len(b) != 32 {
			panic("invalid address pointer")
		}
		
		var addr [32]byte
		copy(addr[:], b)
		
		stack[0] = ec.State.GetBalance(addr)
	}
}

// hostLog implements env.oen_log(ptr i32, len i32)
func hostLog() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec != nil {
			if err := ec.UseGas(gasCostLog); err != nil {
				panic(ErrOutOfGas)
			}
		}
		ptr, length := uint32(stack[0]), uint32(stack[1])
		mem := mod.Memory()
		if mem == nil {
			return
		}
		b, ok := mem.Read(ptr, length)
		if !ok {
			return
		}
		fmt.Printf("[contract log] %s\n", b)
	}
}

// hostCaller implements env.oen_caller() -> i64
func hostCaller() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil {
			stack[0] = 0
			return
		}
		if err := ec.UseGas(gasCostCaller); err != nil {
			panic(ErrOutOfGas)
		}
		var v uint64
		for i := 0; i < 8; i++ {
			v = (v << 8) | uint64(ec.CallerAddr[i])
		}
		stack[0] = v
	}
}

// hostBlockHeight implements env.oen_block_height() -> i64
func hostBlockHeight() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil {
			stack[0] = 0
			return
		}
		if err := ec.UseGas(gasCostBlockHeight); err != nil {
			panic(ErrOutOfGas)
		}
		stack[0] = ec.BlockHeight
	}
}

// hostValue implements env.oen_value() -> i64
func hostValue() api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		ec := GetExecCtx(ctx)
		if ec == nil {
			stack[0] = 0
			return
		}
		if err := ec.UseGas(gasCostValue); err != nil {
			panic(ErrOutOfGas)
		}
		stack[0] = ec.Value
	}
}
