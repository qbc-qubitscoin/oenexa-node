package state

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/vm"
)

// ApplyBlockParallel executes transactions concurrently if they do not touch
// overlapping accounts. This significantly increases transaction throughput (TPS)
// on multi-core hardware, moving OENEXA towards the 100,000+ TPS target.
//
// BurnedFees are implicitly removed from supply (deflationary).
func ApplyBlockParallel(
	st *DB,
	blk *core.Block,
	validatorAddr [crypto.AddressSize]byte,
	execVM *vm.VM,
) (*BlockResult, error) {

	baseFee := blk.Header.BaseFee
	var remainingGas int64 = int64(core.BlockGasLimit)
	
	result := &BlockResult{
		TxResults: make([]*TxResult, len(blk.Txs)),
	}

	// ── Phase 1: Static Dependency Analysis ───────────────────────────────────
	// We group transactions into sequential "layers". All transactions in a
	// layer are guaranteed to touch strictly disjoint addresses and can thus
	// be executed entirely in parallel.
	
	type Layer [][]*core.Transaction
	var layers [][]int // stores indices of blk.Txs
	
	for i, tx := range blk.Txs {
		sender := tx.From
		recipient := tx.To
		
		// Find the earliest layer where this tx does not conflict with existing txs
		placed := false
		for layerIdx, layer := range layers {
			conflict := false
			for _, prevIdx := range layer {
				prevTx := blk.Txs[prevIdx]
				pSender := prevTx.From
				
				// Conflict: they share a sender, recipient, or cross-touch.
				// (0x0 addresses like contract creation also count as conflicts for safety)
				if sender == pSender || sender == prevTx.To ||
				   recipient == pSender || recipient == prevTx.To {
					conflict = true
					break
				}
			}
			
			if !conflict {
				layers[layerIdx] = append(layers[layerIdx], i)
				placed = true
				break
			}
		}
		
		if !placed {
			// Create a new layer
			layers = append(layers, []int{i})
		}
	}

	// ── Phase 2: Parallel Execution ──────────────────────────────────────────
	
	var gasUsed, feeCollected, burnedFees, validatorTip uint64
	var globalErr atomic.Value
	
	for _, layer := range layers {
		var wg sync.WaitGroup
		
		// We execute the layer in parallel.
		for _, txIdx := range layer {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				
				tx := blk.Txs[idx]
				
				// Deduct max gas conservatively before execution to prevent race condition on block limit
				maxGas := int64(tx.GasLimit)
				if atomic.AddInt64(&remainingGas, -maxGas) < 0 {
					// Revert the deduction if we blew the limit (unlikely since mempool checks this)
					atomic.AddInt64(&remainingGas, maxGas)
					globalErr.Store(fmt.Errorf("block gas limit exceeded"))
					return
				}
				
				// Check if another goroutine failed
				if globalErr.Load() != nil {
					return
				}
				
				txRes, err := ApplyTransaction(st, tx, uint64(maxGas), execVM, baseFee)
				if err != nil {
					globalErr.Store(fmt.Errorf("tx %s failed: %w", crypto.ToHex(tx.Hash), err))
					return
				}
				
				// Refund unused gas to the remaining pool
				atomic.AddInt64(&remainingGas, maxGas - int64(txRes.GasUsed))
				
				// Safely accumulate stats
				atomic.AddUint64(&gasUsed, txRes.GasUsed)
				atomic.AddUint64(&feeCollected, txRes.FeeCollected)
				atomic.AddUint64(&burnedFees, txRes.BurnedFee)
				atomic.AddUint64(&validatorTip, txRes.ValidatorTip)
				
				result.TxResults[idx] = txRes
			}(txIdx)
		}
		
		wg.Wait()
		
		if errVal := globalErr.Load(); errVal != nil {
			return nil, errVal.(error)
		}
	}

	result.GasUsed = gasUsed
	result.FeeCollected = feeCollected
	result.BurnedFees = burnedFees
	result.ValidatorTip = validatorTip

	// ── Phase 3: Credit Validator (Sequential) ───────────────────────────────
	reward := core.BlockReward(blk.Header.Height)
	result.BlockReward = reward

	income := result.ValidatorTip + reward
	if income > 0 {
		validator := st.GetAccount(validatorAddr)
		validator.Balance += income
		st.SetAccount(validatorAddr, validator)
	}

	result.ShieldedBalance = st.GetShieldedBalance()
	return result, nil
}
