package risk

import (
	"fmt"
	"math"
	"nofx/kernel"
	"nofx/market"
)

// RiskLevel constants
const (
	RiskLevelSafe    = "Safe"
	RiskLevelCaution = "Caution"
	RiskLevelDanger  = "Danger"
)

// RiskMonitor monitors portfolio risk
type RiskMonitor struct {
	MaxDrawdownLimit float64 // Maximum allowed drawdown percentage (e.g. 20.0)
	VaRLimit         float64 // Value at Risk limit percentage (e.g. 5.0)
	MaxExposure      float64 // Maximum exposure limit percentage (e.g. 200.0 for 2x leverage)
}

// NewRiskMonitor creates a new risk monitor
func NewRiskMonitor(maxDrawdown, varLimit, maxExposure float64) *RiskMonitor {
	if maxDrawdown <= 0 {
		maxDrawdown = 20.0
	}
	if varLimit <= 0 {
		varLimit = 5.0
	}
	if maxExposure <= 0 {
		maxExposure = 300.0 // Default 3x max leverage
	}

	return &RiskMonitor{
		MaxDrawdownLimit: maxDrawdown,
		VaRLimit:         varLimit,
		MaxExposure:      maxExposure,
	}
}

// Analyze analyzes the current portfolio risk
func (rm *RiskMonitor) Analyze(account *kernel.AccountInfo, positions []kernel.PositionInfo, marketDataMap map[string]*market.Data) *kernel.RiskState {
	state := &kernel.RiskState{
		Level:       RiskLevelSafe,
		VaR:         0,
		VaRPct:      0,
		MaxDrawdown: 0, // In a real system, this would track historical high water mark
		Utilization: account.MarginUsedPct,
		Exposure:    0,
		Message:     "",
	}

	// 1. Calculate Total Exposure and VaR
	totalExposure := 0.0
	totalVaR := 0.0

	// Z-score for 95% confidence interval (1.645)
	const zScore95 = 1.645

	for _, pos := range positions {
		value := math.Abs(pos.Quantity * pos.MarkPrice)
		totalExposure += value

		// Get volatility for this position
		volatility := 0.02 // Default 2% daily volatility
		if data, ok := marketDataMap[pos.Symbol]; ok {
			// Use ATR/Price as volatility estimate if available
			if data.LongerTermContext != nil && data.LongerTermContext.ATR14 > 0 && data.CurrentPrice > 0 {
				volatility = data.LongerTermContext.ATR14 / data.CurrentPrice
			} else if data.IntradaySeries != nil && data.IntradaySeries.ATR14 > 0 && data.CurrentPrice > 0 {
				volatility = data.IntradaySeries.ATR14 / data.CurrentPrice
			}
		}

		// Calculate position VaR = Value * Volatility * Z-Score
		// This assumes 1-day horizon
		posVaR := value * volatility * zScore95
		totalVaR += posVaR
	}

	state.Exposure = totalExposure
	state.VaR = totalVaR

	if account.TotalEquity > 0 {
		state.VaRPct = (state.VaR / account.TotalEquity) * 100
	}

	// 2. Determine Risk Level
	reasons := []string{}

	// Check VaR
	if state.VaRPct > rm.VaRLimit {
		state.Level = RiskLevelDanger
		reasons = append(reasons, fmt.Sprintf("VaR exceeds limit (%.1f%% > %.1f%%)", state.VaRPct, rm.VaRLimit))
	} else if state.VaRPct > rm.VaRLimit*0.7 {
		if state.Level != RiskLevelDanger {
			state.Level = RiskLevelCaution
		}
		reasons = append(reasons, fmt.Sprintf("VaR approaching limit (%.1f%%)", state.VaRPct))
	}

	// Check Exposure (Leverage)
	leverage := 0.0
	if account.TotalEquity > 0 {
		leverage = (totalExposure / account.TotalEquity) * 100
	}
	
	if leverage > rm.MaxExposure {
		state.Level = RiskLevelDanger
		reasons = append(reasons, fmt.Sprintf("Exposure exceeds limit (%.1f%% > %.1f%%)", leverage, rm.MaxExposure))
	} else if leverage > rm.MaxExposure*0.8 {
		if state.Level != RiskLevelDanger {
			state.Level = RiskLevelCaution
		}
		reasons = append(reasons, fmt.Sprintf("High leverage usage (%.1f%%)", leverage))
	}

	// Check Margin Utilization
	if account.MarginUsedPct > 80.0 {
		state.Level = RiskLevelDanger
		reasons = append(reasons, fmt.Sprintf("Margin usage critical (%.1f%%)", account.MarginUsedPct))
	} else if account.MarginUsedPct > 60.0 {
		if state.Level != RiskLevelDanger {
			state.Level = RiskLevelCaution
		}
		reasons = append(reasons, fmt.Sprintf("Margin usage high (%.1f%%)", account.MarginUsedPct))
	}

	if len(reasons) > 0 {
		state.Message = fmt.Sprintf("Risk Alerts: %v", reasons)
	} else {
		state.Message = "Portfolio risk is within safe limits."
	}

	return state
}
