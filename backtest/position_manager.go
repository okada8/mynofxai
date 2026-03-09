package backtest

import "math"

// DynamicPositionManager handles position sizing logic
type DynamicPositionManager struct {
	TotalCapital    float64
	MaxRiskPerTrade float64 // Max risk per trade (e.g., 0.01 for 1%)
	MaxPositionPct  float64 // Max position size as percentage of capital (e.g., 0.05 for 5%)
	LiquidityAware  bool    // Whether to consider liquidity
}

// NewDynamicPositionManager creates a new position manager
func NewDynamicPositionManager(capital, riskPerTrade, maxPosPct float64, liquidityAware bool) *DynamicPositionManager {
	if riskPerTrade <= 0 {
		riskPerTrade = 0.01 // Default 1%
	}
	if maxPosPct <= 0 {
		maxPosPct = 0.05 // Default 5%
	}
	
	return &DynamicPositionManager{
		TotalCapital:    capital,
		MaxRiskPerTrade: riskPerTrade,
		MaxPositionPct:  maxPosPct,
		LiquidityAware:  liquidityAware,
	}
}

// CalculatePositionSize computes the recommended position size (in Quote currency value)
func (dpm *DynamicPositionManager) CalculatePositionSize(
	symbol string,
	dailyVolume float64,
	volatility float64,
	confidence float64,
) float64 {
	// Base size calculation
	// Strategy: Risk % of Capital per trade
	// Position Size = (Capital * Risk%) / StopLoss%
	// But here we use a simplified model: Base allocation scaled by factors
	
	// Start with a base allocation of TotalCapital * MaxRiskPerTrade * 5 (assuming 20% stop loss effectively)
	// Or more simply: 
	baseSize := dpm.TotalCapital * dpm.MaxPositionPct * 0.5 // Start with 50% of max allowed position

	
	// Liquidity Adjustment: Based on 24h volume
	if dpm.LiquidityAware && dailyVolume > 0 {
		// Small cap (< 10M daily volume) -> reduce by 50%
		if dailyVolume < 10_000_000 {
			baseSize *= 0.5
		} else if dailyVolume < 50_000_000 {
			// Mid cap (< 50M) -> reduce by 20%
			baseSize *= 0.8
		}
	}
	
	// Volatility Adjustment: High volatility (ATR > 5%) -> reduce size
	// Assuming volatility is passed as a percentage (e.g., 0.05 for 5%)
	if volatility > 0.05 {
		baseSize *= 0.7
	} else if volatility > 0.10 {
		baseSize *= 0.5
	}
	
	// AI Confidence Adjustment: Lower confidence -> reduce size
	// Assuming confidence is 0-100 or 0-1
	normConfidence := confidence
	if confidence > 1.0 {
		normConfidence = confidence / 100.0
	}
	
	if normConfidence < 0.7 {
		baseSize *= 0.5
	} else if normConfidence > 0.9 {
		// High confidence boost (up to 1.2x, but capped by max)
		baseSize *= 1.2
	}
	
	// Cap at MaxPositionPct
	maxSize := dpm.TotalCapital * dpm.MaxPositionPct
	return math.Min(baseSize, maxSize)
}

// UpdateCapital updates the total capital (e.g., after trade close)
func (dpm *DynamicPositionManager) UpdateCapital(newCapital float64) {
	dpm.TotalCapital = newCapital
}
