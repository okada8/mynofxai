import {
  calculateSMA,
  calculateEMA,
  calculateBollingerBands,
  Kline,
} from '../utils/indicators'

// Define message types
type WorkerMessage = {
  type: 'CALCULATE_INDICATORS'
  payload: {
    klineData: Kline[]
    indicators: any[] // IndicatorConfig[]
  }
}

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
  const { type, payload } = e.data

  if (type === 'CALCULATE_INDICATORS') {
    const { klineData, indicators } = payload
    const results: Record<string, any> = {}

    try {
      indicators.forEach((indicator) => {
        if (!indicator.enabled) return

        if (indicator.id.startsWith('ma')) {
          results[indicator.id] = calculateSMA(
            klineData,
            indicator.params.period
          )
        } else if (indicator.id.startsWith('ema')) {
          results[indicator.id] = calculateEMA(
            klineData,
            indicator.params.period
          )
        } else if (indicator.id === 'bb') {
          results[indicator.id] = calculateBollingerBands(klineData)
        }
      })

      self.postMessage({ type: 'INDICATORS_CALCULATED', payload: results })
    } catch (error) {
      self.postMessage({
        type: 'ERROR',
        payload:
          error instanceof Error ? error.message : 'Unknown worker error',
      })
    }
  }
}
