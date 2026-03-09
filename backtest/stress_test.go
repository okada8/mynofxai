package backtest

import (
	"fmt"
	"math"
)

// StressScenario defines a specific market crash or stress event.
type StressScenario struct {
	Name                 string
	BTCDrop              float64 // e.g. 0.50 for 50% drop
	ETHDrop              float64 // e.g. 0.60 for 60% drop
	AltcoinDrop          float64 // e.g. 0.80 for 80% drop
	VolatilityMultiplier float64 // e.g. 2.0 for 2x volatility
}

// StressTestResult holds the outcome of a stress test.
type StressTestResult struct {
	ScenarioName   string
	EquityBefore   float64
	EquityAfter    float64
	Loss           float64
	LossPct        float64
	MaxDrawdownPct float64
	Liquidated     bool
}

// StressTester runs stress tests on a given portfolio.
type StressTester struct {
	Scenarios []StressScenario
}

// NewStressTester creates a new StressTester with default scenarios.
func NewStressTester() *StressTester {
	return &StressTester{
		Scenarios: []StressScenario{
			{
				Name:                 "Black Thursday (Mar 2020)",
				BTCDrop:              0.50,
				ETHDrop:              0.55,
				AltcoinDrop:          0.60,
				VolatilityMultiplier: 3.0,
			},
			{
				Name:                 "LUNA Crash (May 2022)",
				BTCDrop:              0.30,
				ETHDrop:              0.35,
				AltcoinDrop:          0.80, // Altcoins crushed
				VolatilityMultiplier: 2.5,
			},
			{
				Name:                 "FTX Collapse (Nov 2022)",
				BTCDrop:              0.25,
				ETHDrop:              0.30,
				AltcoinDrop:          0.40,
				VolatilityMultiplier: 2.0,
			},
			{
				Name:                 "Flash Crash (Generic)",
				BTCDrop:              0.15,
				ETHDrop:              0.20,
				AltcoinDrop:          0.25,
				VolatilityMultiplier: 5.0, // Extreme short-term volatility
			},
		},
	}
}

// RunStressTest simulates the impact of a scenario on the current portfolio.
// This is a simplified static simulation: it assumes instantaneous price drops.
func (st *StressTester) RunStressTest(account *BacktestAccount, scenario StressScenario) StressTestResult {
	// Get current equity
	currentPrices := getCurrentPrices(account)
	equityBefore, _, _ := account.TotalEquity(currentPrices)
	
	// Simulate portfolio value after crash
	// We calculate PnL impact for each position based on the scenario drop
	marginUsed := 0.0
	unrealizedPnL := 0.0
	
	// Track if account is liquidated
	isLiquidated := false

	for _, pos := range account.positions {
		// Determine drop rate based on symbol type
		dropRate := scenario.AltcoinDrop
		if pos.Symbol == "BTCUSDT" {
			dropRate = scenario.BTCDrop
		} else if pos.Symbol == "ETHUSDT" {
			dropRate = scenario.ETHDrop
		}
		
		// Simulate new price
		// Current price is approximated by EntryPrice (or we could pass current market prices)
		// For stress test, assuming entry price is "current" is a reasonable baseline for "what if crash happens NOW"
		currentPrice := pos.EntryPrice 
		
		// Apply drop rate: Price drops by dropRate%
		newPrice := currentPrice * (1.0 - dropRate)
		
		// Calculate PnL impact
		var pnl float64
		if pos.Side == "long" {
			// Long: Loss = (New - Entry) * Qty
			// If Price drops, New < Entry, PnL is negative
			pnl = (newPrice - pos.EntryPrice) * pos.Quantity
		} else {
			// Short: Profit = (Entry - New) * Qty
			// If Price drops, New < Entry, PnL is positive
			pnl = (pos.EntryPrice - newPrice) * pos.Quantity
		}
		
		unrealizedPnL += pnl
		marginUsed += pos.Margin
		
		// Check individual position liquidation condition
		// Liquidation Price for Long: Entry * (1 - 1/Lev)
		// If NewPrice <= LiquidationPrice, position is wiped out.
		if pos.Side == "long" {
			if newPrice <= pos.LiquidationPrice {
				// Position liquidated -> Loss equals margin + any extra if gap risk
				// Simplified: assume max loss is margin
				// pnl = -pos.Margin // Or keep calculated negative PnL if cross margin
			}
		} else {
			// Short liquidation: NewPrice >= LiquidationPrice
			// In a crash, shorts are generally safe.
			if newPrice >= pos.LiquidationPrice {
				// Unlikely in crash unless volatility spike up
			}
		}
	}
	
	// Equity After = Cash + Margin + Unrealized PnL
	// Note: Account Cash is assumed safe (unless exchange risk modeled)
	equityAfter := account.Cash() + marginUsed + unrealizedPnL
	
	// Check for bankruptcy (Account Liquidation)
	if equityAfter <= 0 {
		equityAfter = 0
		isLiquidated = true
	}
	
	loss := equityBefore - equityAfter
	lossPct := 0.0
	if equityBefore > 0 {
		lossPct = loss / equityBefore
	}
	
	return StressTestResult{
		ScenarioName:   scenario.Name,
		EquityBefore:   equityBefore,
		EquityAfter:    equityAfter,
		Loss:           loss,
		LossPct:        lossPct,
		Liquidated:     isLiquidated,
	}
}

// RunAllScenarios runs all predefined scenarios against the account.
func (st *StressTester) RunAllScenarios(account *BacktestAccount) []StressTestResult {
	results := make([]StressTestResult, 0, len(st.Scenarios))
	for _, scenario := range st.Scenarios {
		results = append(results, st.RunStressTest(account, scenario))
	}
	return results
}

// Helper to get current prices from account positions (using EntryPrice as proxy for current snapshot)
func getCurrentPrices(acc *BacktestAccount) map[string]float64 {
	prices := make(map[string]float64)
	for _, pos := range acc.positions {
		prices[pos.Symbol] = pos.EntryPrice
	}
	return prices
}
