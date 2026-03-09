package backtest

import "math"

// DynamicLeverageAdjuster handles dynamic leverage adjustment logic.
type DynamicLeverageAdjuster struct {
	MaxLeverage            int     // Maximum allowed leverage
	VolatilityThreshold    float64 // Volatility threshold (e.g., 0.08 for 8% ATR)
	DrawdownThreshold      float64 // Drawdown threshold (e.g., 0.05 for 5%)
	HighRiskLeverageCap    int     // Cap for high risk scenarios (e.g. 1x or 2x)
	DrawdownLeverageFactor float64 // Factor to reduce leverage by during drawdown (e.g. 0.5)
}

// NewDynamicLeverageAdjuster creates a new adjuster.
func NewDynamicLeverageAdjuster(maxLev int) *DynamicLeverageAdjuster {
	if maxLev <= 0 {
		maxLev = 1
	}
	return &DynamicLeverageAdjuster{
		MaxLeverage:            maxLev,
		VolatilityThreshold:    0.08, // Default 8% ATR
		DrawdownThreshold:      0.05, // Default 5% drawdown
		HighRiskLeverageCap:    maxLev / 2,
		DrawdownLeverageFactor: 0.5,
	}
}

// AdjustLeverage calculates the recommended leverage based on market conditions and portfolio state.
func (dla *DynamicLeverageAdjuster) AdjustLeverage(
	marketVolatility float64, // ATR percentage (e.g. 0.05)
	portfolioDrawdown float64, // Current drawdown percentage (positive value, e.g. 0.10)
	currentLeverage int, // Current leverage setting
) int {
	targetLeverage := currentLeverage
	if targetLeverage <= 0 {
		targetLeverage = dla.MaxLeverage
	}
	
	// 1. Market Volatility Check
	// High volatility -> Reduce leverage significantly
	if marketVolatility > dla.VolatilityThreshold {
		// Cap leverage in high volatility (e.g. max 2x or 3x)
		cap := dla.HighRiskLeverageCap
		if cap < 1 { cap = 1 }
		
		if targetLeverage > cap {
			targetLeverage = cap
		}
	} else if marketVolatility > dla.VolatilityThreshold * 0.75 {
		// Moderate-high volatility -> reduce slightly
		targetLeverage = int(math.Round(float64(targetLeverage) * 0.75))
	}

	// 2. Portfolio Drawdown Check
	// Deep drawdown -> De-risk to preserve capital
	if portfolioDrawdown > dla.DrawdownThreshold {
		// Reduce leverage by factor (e.g. halve it)
		targetLeverage = int(math.Round(float64(targetLeverage) * dla.DrawdownLeverageFactor))
	} else if portfolioDrawdown > dla.DrawdownThreshold * 0.5 {
		// Approaching drawdown limit -> reduce slightly (e.g. 0.8x)
		targetLeverage = int(math.Round(float64(targetLeverage) * 0.8))
	}

	// Safety bounds
	if targetLeverage < 1 {
		targetLeverage = 1
	}
	if targetLeverage > dla.MaxLeverage {
		targetLeverage = dla.MaxLeverage
	}

	return targetLeverage
}
