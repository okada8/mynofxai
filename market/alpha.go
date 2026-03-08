package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nofx/config"
	"nofx/logger"
	"nofx/provider/coinank"
	"nofx/provider/coinank/coinank_enum"
	"nofx/provider/nofxos"
)

// AlphaProvider interface for alpha factor providers
type AlphaProvider interface {
	Name() string
	Fetch(symbol string) (*AlphaData, error)
}

// AlphaData standardized alpha factor data
type AlphaData struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`

	// Liquidation Clusters
	LiquidationLongUSD  float64 `json:"liquidation_long_usd"`
	LiquidationShortUSD float64 `json:"liquidation_short_usd"`
	LiquidationCluster  bool    `json:"liquidation_cluster"` // true if significant cluster detected

	// Exchange Flow
	NetFlowUSD         float64 `json:"net_flow_usd"`
	SignificantInflow  bool    `json:"significant_inflow"`
	SignificantOutflow bool    `json:"significant_outflow"`
	InstitutionNetFlow float64 `json:"institution_net_flow"`

	// Whale Activity (placeholder for future implementation)
	WhaleBuyVolume  float64 `json:"whale_buy_volume"`
	WhaleSellVolume float64 `json:"whale_sell_volume"`
}

// AlphaManager manages alpha factor data fetching and caching
type AlphaManager struct {
	providers map[string]AlphaProvider
	cache     sync.Map // map[string]*AlphaData
	cacheTTL  time.Duration
	coinank   *coinank.CoinankClient
}

// NewAlphaManager creates a new alpha manager
func NewAlphaManager() *AlphaManager {
	cfg := config.Get()
	return &AlphaManager{
		providers: make(map[string]AlphaProvider),
		cacheTTL:  15 * time.Minute, // Alpha factors update slower than price
		coinank:   coinank.NewCoinankClient(coinank.MainApiUrl, cfg.CoinAnkAPIKey),
	}
}

// GetAlphaData fetches alpha data for a symbol (with caching)
func (m *AlphaManager) GetAlphaData(symbol string) (*AlphaData, error) {
	// Check cache
	if val, ok := m.cache.Load(symbol); ok {
		data := val.(*AlphaData)
		if time.Since(data.Timestamp) < m.cacheTTL {
			return data, nil
		}
	}

	// Fetch fresh data
	// Note: Currently we aggregate data from multiple internal sources rather than external providers directly
	// In the future, this can be refactored to use registered providers
	data, err := m.fetchAggregatedAlpha(symbol)
	if err != nil {
		return nil, err
	}

	// Update cache
	m.cache.Store(symbol, data)
	return data, nil
}

// fetchAggregatedAlpha aggregates alpha data from various sources
func (m *AlphaManager) fetchAggregatedAlpha(symbol string) (*AlphaData, error) {
	data := &AlphaData{
		Symbol:    symbol,
		Timestamp: time.Now(),
	}

	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	// 1. Fetch Liquidation Data (CoinAnk)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.enrichLiquidationData(symbol, data); err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("liquidation data error: %w", err))
			mu.Unlock()
		}
	}()

	// 2. Fetch NetFlow Data (NofxOS)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.enrichNetFlowData(symbol, data); err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("netflow data error: %w", err))
			mu.Unlock()
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		// Log errors but return partial data if available
		logger.Warnf("Alpha data fetch partial errors for %s: %v", symbol, errs)
	}

	return data, nil
}

// enrichLiquidationData fetches liquidation data
func (m *AlphaManager) enrichLiquidationData(symbol string, data *AlphaData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use CoinAnk API to get recent liquidations
	// Note: Using a wider time window (e.g., 4h) to detect clusters
	klines, err := m.coinank.LiquidationHistory(ctx, coinank_enum.Binance, symbol, coinank_enum.Hour1, time.Now().UnixMilli(), 24)
	if err != nil {
		return err
	}

	var totalLong, totalShort float64
	for _, k := range klines {
		totalLong += k.LongTurnover
		totalShort += k.ShortTurnover
	}

	data.LiquidationLongUSD = totalLong
	data.LiquidationShortUSD = totalShort

	// Simple clustering detection: if liquidation > threshold (e.g. 10M USD for BTC, scaled for others)
	// This threshold should ideally be dynamic based on symbol volume
	threshold := 1000000.0 // Default 1M USD
	if symbol == "BTCUSDT" || symbol == "ETHUSDT" {
		threshold = 10000000.0 // 10M for majors
	}

	if totalLong > threshold || totalShort > threshold {
		data.LiquidationCluster = true
	}

	return nil
}

// enrichNetFlowData fetches netflow data
func (m *AlphaManager) enrichNetFlowData(symbol string, data *AlphaData) error {
	// Using NofxOS provider
	// Note: Currently NofxOS provider focuses on rankings, specific symbol lookup might need expansion
	// For now, we simulate fetching specific symbol data or use existing ranking cache if available
	
	// Check if symbol is in top rankings (a simple heuristic for "significant flow")
	rankings, err := nofxos.DefaultClient().GetNetFlowRanking("1h", 100)
	if err != nil {
		return err
	}

	// Helper to check rankings
	checkRankings := func(list []nofxos.NetFlowPosition) {
		for _, item := range list {
			if item.Symbol == symbol {
				data.NetFlowUSD = item.Amount
				data.InstitutionNetFlow = item.Amount // Assuming ranking reflects institutional flow primarily

				// Significant flow logic
				if item.Amount > 100000 { // > 100k inflow
					data.SignificantInflow = true
				} else if item.Amount < -100000 { // > 100k outflow
					data.SignificantOutflow = true
				}
			}
		}
	}

	checkRankings(rankings.InstitutionFutureTop)
	checkRankings(rankings.InstitutionFutureLow)
	checkRankings(rankings.PersonalFutureTop)
	checkRankings(rankings.PersonalFutureLow)

	return nil
}
