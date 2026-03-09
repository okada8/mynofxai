package backtest

import (
	"math"
	"time"
)

// OrderChunk represents a portion of a larger order
type OrderChunk struct {
	Size      float64
	Type      string // "LIMIT", "MARKET"
	TimeDelay time.Duration
}

// OrderSplitter handles splitting large orders into smaller chunks
type OrderSplitter struct {
	TotalSize  float64
	MinChunk   float64 // Min chunk size (e.g. 5000 USD)
	MaxChunk   float64 // Max chunk size (e.g. 20000 USD)
	TimeWindow int     // Execution time window in seconds
}

// SplitOrder generates a list of order chunks based on the configuration
func (os *OrderSplitter) SplitOrder() []OrderChunk {
	chunks := []OrderChunk{}
	remaining := os.TotalSize

	// If total size is small, execute immediately
	if remaining <= os.MinChunk {
		return []OrderChunk{{Size: remaining, Type: "MARKET", TimeDelay: 0}}
	}

	// Basic TWAP strategy logic
	currentDelay := 0
	
	// Calculate approximate step delay
	// Assuming we split into max chunks
	numChunksEst := int(math.Ceil(remaining / os.MaxChunk))
	delayStep := 0
	if numChunksEst > 1 && os.TimeWindow > 0 {
		delayStep = os.TimeWindow / numChunksEst
	}

	for remaining > 0 {
		chunkSize := math.Min(remaining, os.MaxChunk)
		
		// Avoid tiny tail chunks: merge if remaining is too small
		if remaining-chunkSize < os.MinChunk && remaining-chunkSize > 0 {
			chunkSize = remaining
		}
		
		// If current chunk is too small (and not the only one), merge it? 
		// Logic above handles the tail. For the head, we use MaxChunk.
		
		chunks = append(chunks, OrderChunk{
			Size:      chunkSize,
			Type:      "LIMIT", // Default to LIMIT for smart execution
			TimeDelay: time.Duration(currentDelay) * time.Second,
		})
		
		remaining -= chunkSize
		currentDelay += delayStep
	}

	return chunks
}

// TransactionCostAnalysis (TCA) holds execution performance metrics
type TCA struct {
	OrderSize    float64
	AvgFillPrice float64
	VWAPPrice    float64 // Market VWAP during execution
	Slippage     float64 // Realized slippage vs arrival price
	MarketImpact float64 // Estimated price move caused by order
	TimingCost   float64 // Cost of delayed execution
}

// AnalyzeExecution computes TCA metrics
func AnalyzeExecution(orderSize, avgFillPrice, arrivalPrice, vwapPrice, marketPriceAfter float64) TCA {
	// Slippage: (Fill - Arrival) / Arrival (for Buy)
	slippage := (avgFillPrice - arrivalPrice) / arrivalPrice
	if orderSize < 0 { // Sell
		slippage = (arrivalPrice - avgFillPrice) / arrivalPrice
	}

	return TCA{
		OrderSize:    orderSize,
		AvgFillPrice: avgFillPrice,
		VWAPPrice:    vwapPrice,
		Slippage:     slippage,
		// Simplified impact model
		MarketImpact: math.Abs(marketPriceAfter-arrivalPrice) / arrivalPrice,
		TimingCost:   math.Abs(avgFillPrice-vwapPrice) / vwapPrice,
	}
}

// Exchange represents a trading venue
type Exchange struct {
	Name           string
	FeeRate        float64
	LiquidityScore float64 // 0-1 score
}

// SmartOrderRouter determines the best venue for execution
type SmartOrderRouter struct {
	Exchanges []Exchange
	// Simplified depth map: Exchange -> Symbol -> Available Liquidity (USD value at 1% depth)
	OrderBookDepth map[string]map[string]float64
}

// NewSmartOrderRouter creates a router with default exchanges
func NewSmartOrderRouter() *SmartOrderRouter {
	return &SmartOrderRouter{
		Exchanges: []Exchange{
			{Name: "Binance", FeeRate: 0.0004, LiquidityScore: 1.0},
			{Name: "OKX", FeeRate: 0.0005, LiquidityScore: 0.8},
			{Name: "Bybit", FeeRate: 0.00055, LiquidityScore: 0.7},
		},
		OrderBookDepth: make(map[string]map[string]float64),
	}
}

// RouteOrder finds the best exchange for a given order
func (sor *SmartOrderRouter) RouteOrder(symbol string, size float64) string {
	bestExchange := ""
	minCost := math.MaxFloat64

	for _, ex := range sor.Exchanges {
		// Get depth for this exchange and symbol
		var depth float64 = 1000000 // Default fallback depth
		if exDepths, ok := sor.OrderBookDepth[ex.Name]; ok {
			if d, ok := exDepths[symbol]; ok {
				depth = d
			}
		}

		// Estimate cost: Fee + Slippage
		// Simple slippage model: (Size / Depth) * ImpactFactor
		// If Size > Depth, slippage explodes
		impactFactor := 0.01 // 1% price move for full depth consumption
		estimatedSlippage := (size / depth) * impactFactor
		
		totalCost := ex.FeeRate + estimatedSlippage

		if totalCost < minCost {
			minCost = totalCost
			bestExchange = ex.Name
		}
	}

	if bestExchange == "" && len(sor.Exchanges) > 0 {
		return sor.Exchanges[0].Name // Default to first
	}

	return bestExchange
}
