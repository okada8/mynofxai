export const BASE_URL = '/api/nofxos'

export interface AI500Coin {
  pair: string
  score: number
  start_price: number
  max_price: number
  increase_percent: number
  start_time: number
}

export interface OIRankingItem {
  symbol: string
  oi_delta: number
  oi_delta_percent: number
  price: number
  price_delta_percent: number
  current_oi: number
  rank: number
}

export interface FundFlowItem {
  symbol: string
  amount: number
  price: number
  rank: number
}

export interface QueryRankItem {
  rank: number
  symbol: string
  query_count: number
  future_flow: number
}

export interface HeatmapItem {
  rank: number
  symbol: string
  bid_volume: number
  ask_volume: number
  delta: number
}

export interface FundingRateItem {
  rank: number
  symbol: string
  funding_rate: number
  mark_price: number
  index_price: number
  next_funding_time: number
}

export interface AI300Item {
  rank: number
  symbol: string
  future_flow: number
  spot_flow: number
  level: string
}

export interface AI500Response {
  coins: AI500Coin[]
  desc?: string
}

export interface AI300Response {
  coins: AI300Item[]
  desc?: string
}

export interface QueryRankResponse {
  rankings: QueryRankItem[]
  desc?: string
}

export interface FundingRateResponse {
  rates: FundingRateItem[]
  desc?: string
}

export interface PriceRankingItem {
  pair: string
  symbol: string
  price: number
  price_delta: number
  future_flow: number
  spot_flow: number
  oi: number
  oi_delta: number
  oi_delta_value: number
}

export interface PriceRankingResponse {
  top: PriceRankingItem[]
  low: PriceRankingItem[]
}

// Helper to handle API responses
const handleResponse = async (response: Response) => {
  if (!response.ok) {
    throw new Error(`API Error ${response.status}`)
  }
  return response.json()
}

export const nofxosApi = {
  /**
   * Get AI500 Index (Verified)
   */
  getAI500: async (limit = 15): Promise<AI500Response> => {
    try {
      const res = await fetch(`${BASE_URL}/api/ai500/list?limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data) {
        return {
          coins: json.data.coins || [],
          desc: json.data.desc
        }
      }
      return { coins: [] }
    } catch (error) {
      console.error('Failed to fetch AI500:', error)
      return { coins: [] }
    }
  },

  /**
   * Get Query Rank (Verified)
   */
  getQueryRank: async (limit = 8): Promise<QueryRankResponse> => {
    try {
      const res = await fetch(`${BASE_URL}/api/query-rank/list?limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data) {
        return {
          rankings: json.data.rankings || [],
          desc: json.data.desc
        }
      }
      return { rankings: [] }
    } catch (error) {
      console.error('Failed to fetch Query Rank:', error)
      return { rankings: [] }
    }
  },

  /**
   * Get Open Interest Top Ranking (Verified)
   */
  getOITopRanking: async (duration = '1h', limit = 20): Promise<OIRankingItem[]> => {
    try {
      const res = await fetch(`${BASE_URL}/api/oi/top-ranking?duration=${duration}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.positions) {
        return json.data.positions
      }
      return []
    } catch (error) {
      console.error('Failed to fetch OI Top Ranking:', error)
      return []
    }
  },

  /**
   * Get Open Interest Low Ranking (Verified)
   */
  getOILowRanking: async (duration = '1h', limit = 20): Promise<OIRankingItem[]> => {
    try {
      const res = await fetch(`${BASE_URL}/api/oi/low-ranking?duration=${duration}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.positions) {
        return json.data.positions
      }
      return []
    } catch (error) {
      console.error('Failed to fetch OI Low Ranking:', error)
      return []
    }
  },

  /**
   * Get Netflow Top Ranking (Verified)
   */
  getNetflowTopRanking: async (duration = '30m', limit = 20): Promise<FundFlowItem[]> => {
    try {
      const res = await fetch(`${BASE_URL}/api/netflow/top-ranking?duration=${duration}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.netflows) {
        return json.data.netflows
      }
      return []
    } catch (error) {
      console.error('Failed to fetch Netflow Top Ranking:', error)
      return []
    }
  },

  /**
   * Get Netflow Low Ranking (Verified)
   */
  getNetflowLowRanking: async (duration = '30m', limit = 20): Promise<FundFlowItem[]> => {
    try {
      const res = await fetch(`${BASE_URL}/api/netflow/low-ranking?duration=${duration}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.netflows) {
        return json.data.netflows
      }
      return []
    } catch (error) {
      console.error('Failed to fetch Netflow Low Ranking:', error)
      return []
    }
  },

  /**
   * Get Market Heatmap (Depth)
   */
  getHeatmap: async (tradeType: 'future' | 'spot', limit = 20): Promise<HeatmapItem[]> => {
    try {
      const res = await fetch(`${BASE_URL}/api/heatmap/list?trade=${tradeType}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.heatmaps) {
        return json.data.heatmaps
      }
      return []
    } catch (error) {
      console.error(`Failed to fetch ${tradeType} Heatmap:`, error)
      return []
    }
  },

  /**
   * Get Funding Rate Top Ranking (Verified)
   */
  getFundingRateTop: async (limit = 20): Promise<FundingRateResponse> => {
    try {
      const res = await fetch(`${BASE_URL}/api/funding-rate/top?limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data) {
        return {
          rates: json.data.rates || [],
          desc: json.data.desc
        }
      }
      return { rates: [] }
    } catch (error) {
      console.error('Failed to fetch Funding Rate Top:', error)
      return { rates: [] }
    }
  },

  /**
   * Get Funding Rate Low Ranking (Verified)
   */
  getFundingRateLow: async (limit = 20): Promise<FundingRateResponse> => {
    try {
      const res = await fetch(`${BASE_URL}/api/funding-rate/low?limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data) {
        return {
          rates: json.data.rates || [],
          desc: json.data.desc
        }
      }
      return { rates: [] }
    } catch (error) {
      console.error('Failed to fetch Funding Rate Low:', error)
      return { rates: [] }
    }
  },

  /**
   * Get AI300 List (Verified)
   */
  getAI300: async (limit = 15): Promise<AI300Response> => {
    try {
      const res = await fetch(`${BASE_URL}/api/ai300/list?limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data) {
        return {
          coins: json.data.coins || [],
          desc: json.data.desc
        }
      }
      return { coins: [] }
    } catch (error) {
      console.error('Failed to fetch AI300:', error)
      return { coins: [] }
    }
  },

  /**
   * Get Price Ranking (Verified)
   */
  getPriceRanking: async (duration = '1h', limit = 20): Promise<PriceRankingResponse> => {
    try {
      const res = await fetch(`${BASE_URL}/api/price/ranking?duration=${duration}&limit=${limit}`)
      const json = await handleResponse(res)
      if (json.success && json.data && json.data.data && json.data.data[duration]) {
        return {
          top: json.data.data[duration].top || [],
          low: json.data.data[duration].low || []
        }
      }
      return { top: [], low: [] }
    } catch (error) {
      console.error('Failed to fetch Price Ranking:', error)
      return { top: [], low: [] }
    }
  }
}
