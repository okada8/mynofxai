package backtest

import (
	"fmt"
	"testing"
)

func BenchmarkIndicatorCache_Get(b *testing.B) {
	cache := NewIndicatorCache()
	key := "rsi_14_BTCUSDT"
	
	// Pre-fill cache
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i)
	}
	cache.Put(key, data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val, ok := cache.Get(key)
			if !ok || len(val) != 10000 {
				panic("cache miss or invalid data")
			}
		}
	})
}

func BenchmarkIndicatorCache_Put(b *testing.B) {
	cache := NewIndicatorCache()
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Put(key, data)
	}
}

// Simulate a scenario where multiple strategies access the same indicator
func BenchmarkSharedIndicatorAccess(b *testing.B) {
	cache := NewIndicatorCache()
	key := "rsi_14_BTCUSDT"
	
	// Pre-fill cache
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i)
	}
	cache.Put(key, data)

	b.ResetTimer()
	// Simulate 100 concurrent strategies accessing the same data
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Get from cache
			val, ok := cache.Get(key)
			if !ok {
				panic("cache miss")
			}
			// Simulate using the data (read access)
			_ = val[len(val)-1]
		}
	})
}

// Simulate calculation cost vs cache retrieval
func BenchmarkCalculationVsCache(b *testing.B) {
	cache := NewIndicatorCache()
	key := "rsi_14_BTCUSDT"
	data := make([]float64, 1000)
	cache.Put(key, data)

	b.Run("CacheHit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(key)
		}
	})

	b.Run("Calculation_Simulated", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate O(N) calculation
			res := make([]float64, 1000)
			for j := 1; j < 1000; j++ {
				res[j] = res[j-1] + 1.0
			}
		}
	})
}
