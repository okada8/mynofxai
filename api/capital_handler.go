package api

import (
	"net/http"
	"nofx/backtest"
	"strings"

	"github.com/gin-gonic/gin"
)

// handleGetCapitalAllocation checks capital distribution across exchanges
func (s *Server) handleGetCapitalAllocation(c *gin.Context) {
	// Get all traders
	traders := s.traderManager.GetAllTraders()

	// Aggregate capital by exchange
	// Map: ExchangeName -> TotalEquity
	currentBalances := make(map[string]float64)
	totalCapital := 0.0

	for _, t := range traders {
		account, err := t.GetAccountInfo()
		if err != nil {
			continue // Skip traders with error
		}

		exchange := t.GetExchange()
		// Normalize exchange name (capital, title case)
		// e.g. "binance" -> "Binance"
		exchange = normalizeExchangeName(exchange)

		if equity, ok := account["total_equity"].(float64); ok {
			currentBalances[exchange] += equity
			totalCapital += equity
		}
	}

	// Create allocator
	allocator := backtest.NewCapitalAllocator()

	// Generate rebalance plan
	plan := allocator.GenerateRebalancePlan(currentBalances)

	// Calculate current allocation percentages for display
	currentAllocation := make(map[string]float64)
	for ex, bal := range currentBalances {
		if totalCapital > 0 {
			currentAllocation[ex] = bal / totalCapital
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_capital":      totalCapital,
		"current_balances":   currentBalances,
		"current_allocation": currentAllocation,
		"target_allocation":  allocator.Allocation,
		"rebalance_plan":     plan,
	})
}

// normalizeExchangeName converts exchange code to CapitalAllocator format
func normalizeExchangeName(code string) string {
	code = strings.ToLower(code)
	switch code {
	case "binance":
		return "Binance"
	case "okx":
		return "OKX"
	case "bybit":
		return "Bybit"
	case "bitget":
		return "Bitget"
	case "gate":
		return "Gate"
	case "kucoin":
		return "KuCoin"
	case "hyperliquid":
		return "Hyperliquid"
	default:
		// Capitalize first letter
		if len(code) > 0 {
			return strings.ToUpper(code[:1]) + code[1:]
		}
		return "Other"
	}
}
