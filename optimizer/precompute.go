package optimizer

import (
	"fmt"
	"nofx/backtest"
	"nofx/market"
)

// PrecomputeIndicatorsForPopulation computes indicators for the entire population
// to leverage caching and avoid redundant calculations.
func (o *Optimizer) PrecomputeIndicatorsForPopulation(
	pop []Chromosome,
	baseData *market.Data,
	cache *backtest.IndicatorCache,
) {
	if baseData == nil || cache == nil {
		return
	}

	// 1. Extract all needed indicator parameters from genes
	rsiPeriods := make(map[int]bool)
	
	// Default periods we always want if not specified
	rsiPeriods[7] = true
	rsiPeriods[14] = true

	for _, chrom := range pop {
		if p, ok := chrom.Genes["rsi_period"]; ok {
			rsiPeriods[int(p)] = true
		}
	}
	
	// 2. Prepare Kline data
	var klines []market.Kline
	
	if baseData.TimeframeData != nil {
		// Use the first available timeframe data as base
		for _, tfData := range baseData.TimeframeData {
			klines = convertKlineBarsToKlines(tfData.Klines)
			break 
		}
	}
	
	// Fallback to IntradaySeries if TimeframeData is missing but IntradaySeries has Klines (it doesn't usually)
	// If klines is empty, we can't compute.
	if len(klines) == 0 {
		return
	}

	// 3. Batch compute and cache
	for period := range rsiPeriods {
		if period <= 0 {
			continue
		}
		key := fmt.Sprintf("rsi_%d_%s", period, baseData.Symbol)
		
		if _, ok := cache.Get(key); !ok {
			values := calculateRSISeries(klines, period)
			cache.Put(key, values)
		}
	}

	// Compute MACD (Standard 12, 26, 9)
	// In a real scenario, we would extract these from genes too
	macdKey := fmt.Sprintf("macd_12_26_9_%s", baseData.Symbol)
	if _, ok := cache.Get(macdKey); !ok {
		macd, signal, hist := calculateMACDSeries(klines, 12, 26, 9)
		cache.Put(macdKey, macd)
		cache.Put(macdKey+"_signal", signal)
		cache.Put(macdKey+"_hist", hist)
	}
}

func convertKlineBarsToKlines(bars []market.KlineBar) []market.Kline {
	res := make([]market.Kline, len(bars))
	for i, b := range bars {
		res[i] = market.Kline{
			OpenTime:  b.Time,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    b.Volume,
			CloseTime: b.Time + 60000, // Approx 1m
		}
	}
	return res
}

// calculateRSISeries calculates RSI for the entire series
func calculateRSISeries(klines []market.Kline, period int) []float64 {
	if len(klines) <= period {
		return make([]float64, len(klines))
	}
	
	res := make([]float64, len(klines))
	
	gains := 0.0
	losses := 0.0

	// Calculate initial average gain/loss
	for i := 1; i <= period; i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// First RSI
	if avgLoss == 0 {
		res[period] = 100
	} else {
		rs := avgGain / avgLoss
		res[period] = 100 - (100 / (1 + rs))
	}

	// Subsequent RSI calculations
	for i := period + 1; i < len(klines); i++ {
		change := klines[i].Close - klines[i-1].Close
		
		// Calculate current gain/loss
		currentGain := 0.0
		currentLoss := 0.0
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		// Smoothed averages
		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)

		if avgLoss == 0 {
			res[i] = 100
		} else {
			rs := avgGain / avgLoss
			res[i] = 100 - (100 / (1 + rs))
		}
	}
	
	return res
}

// calculateEMASeries calculates EMA series from Klines
func calculateEMASeries(klines []market.Kline, period int) []float64 {
	if len(klines) < period {
		return make([]float64, len(klines))
	}
	res := make([]float64, len(klines))
	
	// SMA as first EMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += klines[i].Close
	}
	ema := sum / float64(period)
	res[period-1] = ema 
	
	multiplier := 2.0 / float64(period+1)
	for i := period; i < len(klines); i++ {
		ema = (klines[i].Close-ema)*multiplier + ema
		res[i] = ema
	}
	return res
}

// calculateMACDSeries calculates MACD, Signal, and Histogram series
func calculateMACDSeries(klines []market.Kline, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, hist []float64) {
	if len(klines) < slowPeriod {
		return nil, nil, nil
	}
	
	emaFast := calculateEMASeries(klines, fastPeriod)
	emaSlow := calculateEMASeries(klines, slowPeriod)
	
	macd = make([]float64, len(klines))
	for i := 0; i < len(klines); i++ {
		if i >= slowPeriod-1 {
			 macd[i] = emaFast[i] - emaSlow[i]
		}
	}
	
	signal = calculateEMAOnSeries(macd, signalPeriod)
	
	hist = make([]float64, len(klines))
	for i := 0; i < len(klines); i++ {
		// Valid only after signal becomes valid
		// Signal valid after (slowPeriod-1) + (signalPeriod-1) roughly
		// We just check if signal[i] is computed (non-zero or index check)
		// Simpler: just subtract
		hist[i] = macd[i] - signal[i]
	}
	
	return macd, signal, hist
}

// calculateEMAOnSeries calculates EMA on a float64 slice
func calculateEMAOnSeries(values []float64, period int) []float64 {
	res := make([]float64, len(values))
	
	// Find start index (first non-zero value as proxy for valid data)
	// This is a simplification. Ideally pass start index.
	start := 0
	for i, v := range values {
		if v != 0 {
			start = i
			break
		}
	}
	
	// If the array is all zeros or not enough data after start
	if start == len(values) || len(values)-start < period {
		return res
	}
	
	sum := 0.0
	for i := start; i < start+period; i++ {
		sum += values[i]
	}
	ema := sum / float64(period)
	res[start+period-1] = ema
	
	multiplier := 2.0 / float64(period+1)
	for i := start + period; i < len(values); i++ {
		ema = (values[i]-ema)*multiplier + ema
		res[i] = ema
	}
	return res
}
