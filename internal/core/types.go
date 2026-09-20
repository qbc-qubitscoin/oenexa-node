package core

// TxType enumerates supported transaction types.
type TxType uint8

const (
	TxTransfer         TxType = 0x01
	TxDeploy           TxType = 0x02
	TxCall             TxType = 0x03
	TxStake            TxType = 0x04
	TxUnstake          TxType = 0x05
	TxShield           TxType = 0x06 // Transparent -> Shielded Pool
	TxUnshield         TxType = 0x07 // Shielded Pool -> Transparent
	TxShieldedTransfer TxType = 0x08 // Shielded -> Shielded Pool
)

// Currency units — 9 decimal places (1 OEN = 1_000_000_000 nano-OEN).
const (
	NanoOEN   uint64 = 1
	OneOEN    uint64 = 1_000_000_000        // 10^9 nano-OEN
	MaxSupply uint64 = 100_000_000 * OneOEN // 100M OEN = 10^17 nano-OEN (fits uint64)

	// Backwards compatibility aliases
	Oenexa uint64 = NanoOEN
)

// ─────────────────────────────────────────────────────────────────────────────
// Gas constants — ultra-low to make OEN the world's cheapest-fee chain.
// ─────────────────────────────────────────────────────────────────────────────
const (
	GasTransfer         uint64 = 21          // standard transfer
	GasDeploy           uint64 = 5_000       // WASM contract deployment
	GasCall             uint64 = 500         // WASM contract call
	GasShield           uint64 = 50          // transparent to shielded deposit
	GasUnshield         uint64 = 100         // shielded to transparent withdraw
	GasShieldedTransfer uint64 = 150         // shielded-to-shielded transfer
	BlockGasLimit       uint64 = 500_000_000 // high block gas limit
	MinGasPrice         uint64 = 1           // 1 nano-OEN/gas absolute floor
)

// Chain metadata.
const (
	Ticker          = "OEN"
	LegacyTicker    = "OEN"
	Decimals        = 9
	ProtocolVersion = 1
	ChainID         = 1
)

// BlockIntervalSec is the target block time in seconds.
const BlockIntervalSec = 2
