package dex

// QubitSwap represents a basic Automated Market Maker (AMM) Liquidity Pool.
// It implements the constant product formula: x * y = k.
type QubitSwap struct {
	ReserveA uint64 // Reserve of Token A
	ReserveB uint64 // Reserve of Token B
}

func NewQubitSwap() *QubitSwap {
	return &QubitSwap{}
}

// AddLiquidity adds tokens to the reserves. In a real contract, this would
// mint LP tokens to the provider based on their proportional share.
func (qs *QubitSwap) AddLiquidity(amountA, amountB uint64) {
	qs.ReserveA += amountA
	qs.ReserveB += amountB
}

// OenSwap is an alias for QubitSwap representing the OENEXA Native DEX.
type OenSwap = QubitSwap

// NewOenSwap creates an empty OenSwap liquidity pool.
func NewOenSwap() *OenSwap {
	return NewQubitSwap()
}

// SwapAforB allows a user to trade Token A for Token B.
// It calculates the output amount ensuring (ReserveA + amountA) * (ReserveB - amountBOut) >= ReserveA * ReserveB.
// It applies a 0.3% fee to amountA.
func (qs *QubitSwap) SwapAforB(amountAIn uint64) (amountBOut uint64, success bool) {
	if amountAIn == 0 || qs.ReserveA == 0 || qs.ReserveB == 0 {
		return 0, false
	}

	// Apply 0.3% fee: amountAInWithFee = amountAIn * 997 / 1000
	amountAInWithFee := (amountAIn * 997) / 1000

	// yOut = (y * xIn) / (x + xIn)
	numerator := qs.ReserveB * amountAInWithFee
	denominator := qs.ReserveA + amountAInWithFee

	amountBOut = numerator / denominator

	qs.ReserveA += amountAIn
	qs.ReserveB -= amountBOut

	return amountBOut, true
}

// SwapBforA allows a user to trade Token B for Token A.
// It applies a 0.3% fee to amountB.
func (qs *QubitSwap) SwapBforA(amountBIn uint64) (amountAOut uint64, success bool) {
	if amountBIn == 0 || qs.ReserveA == 0 || qs.ReserveB == 0 {
		return 0, false
	}

	amountBInWithFee := (amountBIn * 997) / 1000

	numerator := qs.ReserveA * amountBInWithFee
	denominator := qs.ReserveB + amountBInWithFee

	amountAOut = numerator / denominator

	qs.ReserveB += amountBIn
	qs.ReserveA -= amountAOut

	return amountAOut, true
}
