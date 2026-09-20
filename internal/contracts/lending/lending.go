package lending

// Position represents a user's collateralized debt position.
type Position struct {
	CollateralOEN uint64
	DebtQUSD      uint64 // USD stablecoin debt
}

// OenexaLend is the money market contract.
type OenexaLend struct {
	Positions map[string]*Position
	// Price of 1 OEN in QUSD (e.g., from OenexaOracle)
	OraclePrice uint64
}

func NewOenexaLend() *OenexaLend {
	return &OenexaLend{
		Positions:   make(map[string]*Position),
		OraclePrice: 100, // Default to $100 per OEN
	}
}

// UpdateOraclePrice receives the median price from the Phase 10 Oracle.
func (ql *OenexaLend) UpdateOraclePrice(newPrice uint64) {
	ql.OraclePrice = newPrice
}

// DepositCollateral adds OEN to a user's position.
func (ql *OenexaLend) DepositCollateral(user string, amount uint64) {
	if ql.Positions[user] == nil {
		ql.Positions[user] = &Position{}
	}
	ql.Positions[user].CollateralOEN += amount
}

// BorrowQUSD borrows QUSD against deposited OEN.
// Requires a 150% Collateralization Ratio.
func (ql *OenexaLend) BorrowQUSD(user string, borrowAmount uint64) bool {
	pos := ql.Positions[user]
	if pos == nil {
		return false
	}

	collateralValueUSD := pos.CollateralOEN * ql.OraclePrice
	newTotalDebt := pos.DebtQUSD + borrowAmount

	// Check if (Collateral Value / New Debt) >= 1.5
	// Rewritten as integers: Collateral Value * 100 >= New Debt * 150
	if collateralValueUSD*100 < newTotalDebt*150 {
		return false // Under-collateralized
	}

	pos.DebtQUSD = newTotalDebt
	return true
}

// Liquidate allows any user to liquidate a position if the collateral ratio falls below 150%.
// For simplicity, we just zero out the debt and take the collateral.
func (ql *OenexaLend) Liquidate(targetUser string) bool {
	pos := ql.Positions[targetUser]
	if pos == nil || pos.DebtQUSD == 0 {
		return false // Nothing to liquidate
	}

	collateralValueUSD := pos.CollateralOEN * ql.OraclePrice

	// If ratio >= 150%, cannot be liquidated
	if collateralValueUSD*100 >= pos.DebtQUSD*150 {
		return false
	}

	// Position is underwater. Liquidate it.
	pos.CollateralOEN = 0
	pos.DebtQUSD = 0

	return true
}
