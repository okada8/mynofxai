package optimizer

import (
	"math"
	"nofx/backtest"
	"nofx/market"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPrecision_CacheVsRealtime verifies that cached indicators match real-time calculations
func TestPrecision_CacheVsRealtime(t *testing.T) {
	// 1. Setup Data
	// Create a simple price series: 10, 11, 12, ... 110
	klines := make([]market.Kline, 100)
	now := time.Now().Unix() * 1000
	for i := 0; i < 100; i++ {
		klines[i] = market.Kline{
			OpenTime:  now + int64(i*60000),
			CloseTime: now + int64(i*60000) + 60000,
			Open:      float64(10 + i),
			High:      float64(10 + i + 1),
			Low:       float64(10 + i - 1),
			Close:     float64(10 + i), // Linear uptrend
			Volume:    1000,
		}
	}

	baseData := &market.Data{
		Symbol: "BTCUSDT",
		// We need to mock TimeframeData or convertKlineBarsToKlines usage
		// In PrecomputeIndicatorsForPopulation, it uses baseData.TimeframeData
		TimeframeData: map[string]*market.TimeframeSeriesData{
			"1m": {
				Klines: convertKlinesToBars(klines),
			},
		},
	}

	// 2. Real-time Calculation (Direct)
	// We access the private function calculateRSISeries via a helper or by being in the same package
	realtimeRSI := calculateRSISeries(klines, 14)

	// 3. Cached Calculation via Precompute
	cache := backtest.NewIndicatorCache()
	pop := []Chromosome{
		{Genes: map[string]float64{"rsi_period": 14}},
	}
	
	optimizer := &Optimizer{}
	optimizer.PrecomputeIndicatorsForPopulation(pop, baseData, cache)

	// 4. Verification
	cachedRSI, found := cache.Get("rsi_14_BTCUSDT")
	assert.True(t, found, "RSI should be in cache")
	assert.Equal(t, len(realtimeRSI), len(cachedRSI), "Length mismatch")

	// Check precision
	epsilon := 1e-9
	for i := 0; i < len(realtimeRSI); i++ {
		if math.IsNaN(realtimeRSI[i]) {
			// Skip NaN checks if any, though RSI usually 0 or 100 or valid
			continue
		}
		diff := math.Abs(realtimeRSI[i] - cachedRSI[i])
		if diff > epsilon {
			t.Errorf("Precision error at index %d: expected %v, got %v (diff %v)", 
				i, realtimeRSI[i], cachedRSI[i], diff)
		}
	}
}

// Helper to convert back for the test structure setup
func convertKlinesToBars(klines []market.Kline) []market.KlineBar {
	res := make([]market.KlineBar, len(klines))
	for i, k := range klines {
		res[i] = market.KlineBar{
			Time:   k.OpenTime,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
		}
	}
	return res
}

func TestCache_Consistency_TTL(t *testing.T) {
	cache := backtest.NewIndicatorCache()
	cache.SetTTL(100 * time.Millisecond) // Short TTL

	key := "test_key"
	data := []float64{1.0, 2.0, 3.0}

	cache.Put(key, data)

	// Immediate check
	val, ok := cache.Get(key)
	assert.True(t, ok)
	assert.Equal(t, data, val)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Check again
	val, ok = cache.Get(key)
	assert.False(t, ok, "Cache should have expired")
	assert.Nil(t, val)
}
