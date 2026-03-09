package backtest

import (
	"fmt"
	"math"
	"sort"
)

// RiskLevel represents the severity of risk
type RiskLevel string

const (
	RiskLevelGreen  RiskLevel = "green"  // VaR < 1% -> Normal trading
	RiskLevelYellow RiskLevel = "yellow" // 1% <= VaR < 2% -> Reduce positions
	RiskLevelRed    RiskLevel = "red"    // VaR >= 2% -> Pause trading
)

// RiskMonitor monitors real-time portfolio risk (VaR)
type RiskMonitor struct {
	Positions         []PositionSnapshot
	HistoricalReturns []float64 // Daily portfolio returns (percentage, e.g. 0.01 for 1%)
	ConfidenceLevel   float64   // e.g. 0.95 for 95%
	TimeHorizon       int       // e.g. 1 for 1 day
}

// NewRiskMonitor creates a new RiskMonitor
func NewRiskMonitor(confidence float64, timeHorizon int) *RiskMonitor {
	if confidence <= 0 || confidence >= 1 {
		confidence = 0.95 // Default to 95%
	}
	if timeHorizon <= 0 {
		timeHorizon = 1 // Default to 1 day
	}
	return &RiskMonitor{
		Positions:         make([]PositionSnapshot, 0),
		HistoricalReturns: make([]float64, 0),
		ConfidenceLevel:   confidence,
		TimeHorizon:       timeHorizon,
	}
}

// UpdatePositions updates the current positions for monitoring
func (rm *RiskMonitor) UpdatePositions(positions []PositionSnapshot) {
	rm.Positions = positions
}

// AddReturn adds a new historical return observation (e.g. daily PnL %)
func (rm *RiskMonitor) AddReturn(ret float64) {
	rm.HistoricalReturns = append(rm.HistoricalReturns, ret)
	// Optional: Keep window size fixed (e.g. last 252 days)
	if len(rm.HistoricalReturns) > 252 {
		rm.HistoricalReturns = rm.HistoricalReturns[1:]
	}
}

// CalculateVaR calculates the Value at Risk (VaR) percentage using Historical Simulation
// Returns a positive float representing the potential loss percentage (e.g. 0.02 for 2%)
func (rm *RiskMonitor) CalculateVaR() float64 {
	if len(rm.HistoricalReturns) < 10 {
		// Not enough data to be statistically significant
		return 0.0
	}

	// 1. Sort historical returns (ascending) to find the tail
	sorted := make([]float64, len(rm.HistoricalReturns))
	copy(sorted, rm.HistoricalReturns)
	sort.Float64s(sorted)

	// 2. Find the percentile index
	// For 95% confidence, we look at the worst 5% of returns
	percentile := 1.0 - rm.ConfidenceLevel
	index := int(math.Floor(percentile * float64(len(sorted))))

	// Clamp index
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	// 3. Get the return at that percentile
	// If the return is positive (profit), VaR is 0 (no loss risk at this confidence)
	varReturn := sorted[index]
	if varReturn > 0 {
		return 0.0
	}

	// 4. Scale by Time Horizon (Square Root of Time Rule)
	// VaR(T days) ≈ VaR(1 day) * sqrt(T)
	// We assume HistoricalReturns are daily returns.
	// We take absolute value to represent "Value at Risk" as a positive loss magnitude.
	scaledVaR := math.Abs(varReturn) * math.Sqrt(float64(rm.TimeHorizon))

	return scaledVaR
}

// CheckRiskLevel determines the current risk warning level based on VaR
func (rm *RiskMonitor) CheckRiskLevel() (RiskLevel, float64) {
	varPct := rm.CalculateVaR()

	// Risk Thresholds
	if varPct >= 0.02 { // VaR >= 2%
		return RiskLevelRed, varPct
	} else if varPct >= 0.01 { // 1% <= VaR < 2%
		return RiskLevelYellow, varPct
	}
	
	return RiskLevelGreen, varPct // VaR < 1%
}

// GetAlertMessage returns a human-readable alert if risk is elevated
func (rm *RiskMonitor) GetAlertMessage() string {
	level, val := rm.CheckRiskLevel()
	switch level {
	case RiskLevelRed:
		return "⚠️ CRITICAL RISK ALERT: Portfolio VaR is " + formatPct(val) + " (>= 2%). Pausing trading recommended."
	case RiskLevelYellow:
		return "⚠️ RISK WARNING: Portfolio VaR is " + formatPct(val) + " (>= 1%). Position reduction recommended."
	default:
		return "✅ Risk Normal: VaR is " + formatPct(val)
	}
}

func formatPct(val float64) string {
	return fmt.Sprintf("%.2f%%", val*100)
}
