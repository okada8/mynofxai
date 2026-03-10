package kernel

import (
	"fmt"
	"math"
	"time"

	"nofx/store"
)

// PolymarketStrategyEngine implements Polymarket-specific strategies
type PolymarketStrategyEngine struct {
	config *store.PolymarketStrategyConfig
}

// NewPolymarketStrategyEngine creates a new engine
func NewPolymarketStrategyEngine(config *store.PolymarketStrategyConfig) *PolymarketStrategyEngine {
	return &PolymarketStrategyEngine{
		config: config,
	}
}

// MarketData represents the market data structure for Polymarket
type MarketData struct {
	ID           string
	Question     string
	EndDate      time.Time
	YesPrice     float64
	NoPrice      float64
	Liquidity    float64
	Volume24h    float64
	DaysToExpiry float64
	OrderBook    *OrderBook // Optional, for market making
}

// OrderBook represents simple order book data
type OrderBook struct {
	Bids []OrderLevel
	Asks []OrderLevel
}

type OrderLevel struct {
	Price float64
	Size  float64
}

// PolymarketDecision represents a trading decision
type PolymarketDecision struct {
	Action     string  // "BUY_YES", "BUY_NO", "SELL_YES", "SELL_NO", "HOLD"
	Confidence float64 // 0.0 - 1.0
	Reasoning  string
	Orders     []OrderRequest // For market making (limit orders)
}

type OrderRequest struct {
	Side  string // "BUY", "SELL"
	Price float64
	Size  float64
}

// Execute runs the enabled sub-strategies and combines their results
func (e *PolymarketStrategyEngine) Execute(market *MarketData) *PolymarketDecision {
	if market.Liquidity < e.config.MinLiquidity {
		return &PolymarketDecision{Action: "HOLD", Reasoning: "Insufficient liquidity"}
	}

	if market.DaysToExpiry > float64(e.config.MaxDaysToExpiry) || market.DaysToExpiry < float64(e.config.MinDaysToExpiry) {
		return &PolymarketDecision{Action: "HOLD", Reasoning: "Outside expiry window"}
	}

	var weightedConfidence float64
	var totalWeight float64
	var reasonings []string
	var marketMakingOrders []OrderRequest

	actionScores := map[string]float64{
		"BUY_YES": 0,
		"BUY_NO":  0,
		"HOLD":    0,
	}

	for _, sub := range e.config.SubStrategies {
		if !sub.Enabled {
			continue
		}

		var action string
		var conf float64
		var reasoning string

		switch sub.Name {
		case "probability_arbitrage":
			action, conf, reasoning = e.probabilityArbitrageStrategy(market)
		case "time_decay":
			action, conf, reasoning = e.timeDecayStrategy(market)
		case "market_making":
			if market.OrderBook != nil {
				orders := e.marketMakingStrategy(market.OrderBook)
				if len(orders) > 0 {
					marketMakingOrders = append(marketMakingOrders, orders...)
					reasoning = fmt.Sprintf("Market making: placing %d orders", len(orders))
					// Market making is independent, doesn't vote on directional bias usually
					// unless we want to direct it. For now treating as neutral/separate.
				}
			}
		}

		if action != "" {
			weightedConfidence += conf * sub.Weight
			totalWeight += sub.Weight
			actionScores[action] += conf * sub.Weight
			if reasoning != "" {
				reasonings = append(reasonings, fmt.Sprintf("[%s]: %s", sub.Name, reasoning))
			}
		}
	}

	// Determine winner
	bestAction := "HOLD"
	bestScore := 0.0
	for act, score := range actionScores {
		if score > bestScore {
			bestScore = score
			bestAction = act
		}
	}

	// Normalize confidence
	finalConfidence := 0.0
	if totalWeight > 0 {
		finalConfidence = bestScore / totalWeight
	}

	return &PolymarketDecision{
		Action:     bestAction,
		Confidence: finalConfidence,
		Reasoning:  fmt.Sprintf("Combined strategy: %s", reasonings),
		Orders:     marketMakingOrders,
	}
}

// 1. Probability Arbitrage Strategy
func (e *PolymarketStrategyEngine) probabilityArbitrageStrategy(market *MarketData) (string, float64, string) {
	impliedProb := market.YesPrice // Assuming YesPrice is normalized 0-1
	if market.YesPrice+market.NoPrice > 0 {
		impliedProb = market.YesPrice / (market.YesPrice + market.NoPrice)
	}

	// Calculate fair probability (Placeholder for external model/data)
	// In a real system, this would call an AI model or oracle
	fairProb := e.calculateFairProbability(market.Question)

	diff := fairProb - impliedProb
	threshold := e.config.ProbabilityThreshold

	if diff > threshold {
		// Fair > Implied => Undervalued => Buy YES
		return "BUY_YES", math.Abs(diff), fmt.Sprintf("Fair (%.2f) > Implied (%.2f)", fairProb, impliedProb)
	} else if diff < -threshold {
		// Fair < Implied => Overvalued => Buy NO (or Sell YES)
		return "BUY_NO", math.Abs(diff), fmt.Sprintf("Fair (%.2f) < Implied (%.2f)", fairProb, impliedProb)
	}

	return "HOLD", 0, "Fair price within threshold"
}

// 2. Time Decay Strategy
func (e *PolymarketStrategyEngine) timeDecayStrategy(market *MarketData) (string, float64, string) {
	if market.DaysToExpiry <= 1 {
		// Last day: aggressive
		if market.YesPrice < 0.1 {
			return "BUY_YES", 0.3, "High odds snipe (<0.1)"
		} else if market.YesPrice > 0.9 {
			return "SELL_YES", 0.3, "Profit taking (>0.9)"
		}
	} else if market.DaysToExpiry <= 7 {
		return "HOLD", 0.1, "Last week caution"
	}

	return "HOLD", 0, "Normal time window"
}

// 3. Market Making Strategy
func (e *PolymarketStrategyEngine) marketMakingStrategy(ob *OrderBook) []OrderRequest {
	if len(ob.Bids) == 0 || len(ob.Asks) == 0 {
		return nil
	}

	bestBid := ob.Bids[0].Price
	bestAsk := ob.Asks[0].Price
	spread := bestAsk - bestBid

	if spread > 0.02 { // > 2% spread
		midPrice := (bestBid + bestAsk) / 2
		size := 10.0 // Default size, should be dynamic based on config

		return []OrderRequest{
			{Side: "BUY", Price: midPrice - 0.005, Size: size},
			{Side: "SELL", Price: midPrice + 0.005, Size: size},
		}
	}

	return nil
}

// Helper: Mock fair probability calculation
func (e *PolymarketStrategyEngine) calculateFairProbability(question string) float64 {
	// TODO: Integrate with LLM to analyze question semantics and news
	return 0.5 // Default neutral
}
