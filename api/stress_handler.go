package api

import (
	"net/http"
	"nofx/backtest"
	"time"

	"github.com/gin-gonic/gin"
)

// handleStressTest runs stress test scenarios on a trader's current portfolio
func (s *Server) handleStressTest(c *gin.Context) {
	traderID := c.Param("id")
	if traderID == "" {
		SafeBadRequest(c, "Trader ID is required")
		return
	}

	// Get trader
	at, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		SafeNotFound(c, "Trader not found")
		return
	}

	// Get current account info and positions
	accountInfo, err := at.GetAccountInfo()
	if err != nil {
		SafeInternalError(c, "Failed to get account info", err)
		return
	}

	positions, err := at.GetPositions()
	if err != nil {
		SafeInternalError(c, "Failed to get positions", err)
		return
	}

	// Create BacktestAccount for simulation
	// Assuming 0.05% fee and 0.05% slippage for stress test estimation
	initialBalance := 0.0
	if bal, ok := accountInfo["total_equity"].(float64); ok {
		initialBalance = bal
	}
	
	// Create mock account
	mockAccount := backtest.NewBacktestAccount(initialBalance, 5, 5)

	// Prepare snapshots for restoration
	var snapshots []backtest.PositionSnapshot
	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		quantity := pos["quantity"].(float64)
		entryPrice := pos["entry_price"].(float64)
		leverage := int(pos["leverage"].(float64)) // Assuming it comes as float64 from JSON/Map
		marginUsed := pos["margin_used"].(float64)
		
		// Liquidation price might be 0 if not set, that's fine
		liqPrice := 0.0
		if lp, ok := pos["liquidation_price"].(float64); ok {
			liqPrice = lp
		}

		snapshots = append(snapshots, backtest.PositionSnapshot{
			Symbol:           symbol,
			Side:             side,
			Quantity:         quantity,
			AvgPrice:         entryPrice,
			Leverage:         leverage,
			LiquidationPrice: liqPrice,
			MarginUsed:       marginUsed,
			OpenTime:         time.Now().UnixMilli(), // Mock time
		})
	}

	// Restore state (cash = equity - margin - unrealized, but simplified here we just set cash = equity and let restore handle positions?)
	// Actually RestoreFromSnapshots takes (cash, realized, snaps).
	// We need to calculate Cash from Equity.
	// Equity = Cash + Margin + Unrealized.
	// So Cash = Equity - Margin - Unrealized.
	
	totalEquity := initialBalance
	marginUsed := 0.0
	if m, ok := accountInfo["margin_used"].(float64); ok {
		marginUsed = m
	}
	unrealizedPnL := 0.0
	if u, ok := accountInfo["unrealized_profit"].(float64); ok {
		unrealizedPnL = u
	}

	cash := totalEquity - marginUsed - unrealizedPnL
	
	mockAccount.RestoreFromSnapshots(cash, 0, snapshots)

	// Run stress tests
	tester := backtest.NewStressTester()
	results := tester.RunAllScenarios(mockAccount)

	c.JSON(http.StatusOK, gin.H{
		"trader_id": traderID,
		"equity":    totalEquity,
		"results":   results,
	})
}
