import axios from 'axios'

const API_BASE_URL = '/api/polymarket'

const getHeaders = () => {
  const token = localStorage.getItem('auth_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export interface PolymarketMarket {
  id: string
  question: string
  slug: string
  yesPrice: number
  noPrice: number
  liquidity: number
  volume24h: number
  endDate: string
  outcomes: any[]
}

export interface PolymarketPosition {
  asset_id: string
  symbol: string
  price: number
  size: number
  value_usd: number
}

export interface PolymarketBalance {
  collateral_value: number
  cash: number
}

export const polymarketService = {
  getMarkets: async (limit: number = 20): Promise<PolymarketMarket[]> => {
    const response = await axios.get(`${API_BASE_URL}/markets`, {
      params: { limit },
      headers: getHeaders(),
    })
    return response.data
  },

  getPositions: async (traderId: string): Promise<PolymarketPosition[]> => {
    const response = await axios.get(`${API_BASE_URL}/positions`, {
      params: { trader_id: traderId },
      headers: getHeaders(),
    })
    return response.data
  },

  getBalance: async (traderId: string): Promise<PolymarketBalance> => {
    const response = await axios.get(`${API_BASE_URL}/balance`, {
      params: { trader_id: traderId },
      headers: getHeaders(),
    })
    return response.data
  },

  createOrder: async (order: {
    trader_id: string
    symbol: string
    side: 'BUY' | 'SELL'
    quantity: number
    leverage?: number
  }) => {
    const response = await axios.post(`${API_BASE_URL}/orders`, order, {
      headers: getHeaders(),
    })
    return response.data
  },

  cancelOrder: async (orderId: string, traderId: string, symbol?: string) => {
    const response = await axios.delete(`${API_BASE_URL}/orders/${orderId}`, {
      params: { trader_id: traderId, symbol },
      headers: getHeaders(),
    })
    return response.data
  },
}
