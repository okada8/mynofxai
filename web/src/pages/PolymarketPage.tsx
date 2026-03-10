import React, { useState, useEffect } from 'react'
import { useStore } from '../stores'
import { MarketCard } from '../components/Polymarket/MarketCard'
import { Wallet, RefreshCw, AlertTriangle } from 'lucide-react'

const PolymarketPage: React.FC = () => {
  const { traderStore } = useStore()
  const [markets, setMarkets] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadMarkets()
  }, [])

  const loadMarkets = async () => {
    setLoading(true)
    setError(null)
    try {
      // Mock API call for now, replace with real backend endpoint
      // const response = await fetch('/api/polymarket/markets?limit=20')
      // const data = await response.json()
      
      // Mock data
      await new Promise(resolve => setTimeout(resolve, 1000))
      const mockData = [
        {
          id: '1',
          question: 'Will Bitcoin hit $100k in 2024?',
          liquidity: '500000',
          volume24h: '120000',
          outcomes: [
            { id: '1-yes', price: '0.45' },
            { id: '1-no', price: '0.55' }
          ]
        },
        {
          id: '2',
          question: 'Will Ethereum upgrade execute successfully in Q2?',
          liquidity: '300000',
          volume24h: '80000',
          outcomes: [
            { id: '2-yes', price: '0.80' },
            { id: '2-no', price: '0.20' }
          ]
        }
      ]
      
      setMarkets(mockData)
    } catch (err) {
      console.error('Failed to load markets:', err)
      setError('Failed to load markets. Please check your connection.')
    } finally {
      setLoading(false)
    }
  }

  const isConfigured = !!traderStore.traders.find(t => t.exchange === 'polymarket')

  return (
    <div className="container mx-auto px-4 py-8 text-gray-100">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2 flex items-center gap-3">
          <span className="text-4xl">🔮</span> Polymarket Prediction Market
        </h1>
        <p className="text-gray-400">
          Trade on real-world events. Use your judgment to profit from outcomes.
        </p>
      </div>

      {!isConfigured && (
        <div className="bg-yellow-900/20 border border-yellow-700/50 rounded-lg p-4 mb-8 flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-yellow-500 mt-0.5" />
          <div>
            <h3 className="font-semibold text-yellow-500">Not Configured</h3>
            <p className="text-sm text-gray-300">
              Please add a Polymarket trader in the Trader Configuration to start trading.
            </p>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Sidebar - Account Info */}
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
            <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <Wallet className="w-5 h-5 text-purple-400" /> Account Info
            </h2>
            <div className="space-y-4">
              <div>
                <div className="text-sm text-gray-400 mb-1">Wallet Address</div>
                <div className="font-mono text-sm bg-gray-900/50 p-2 rounded truncate text-gray-300">
                  {traderStore.traders.find(t => t.exchange === 'polymarket')?.polymarketWalletAddr || 'Not Configured'}
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-400 mb-1">Network</div>
                <div className="text-sm font-medium text-green-400 flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                  Polygon Mainnet
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-400 mb-1">Balance</div>
                <div className="text-2xl font-bold text-white">
                  $0.00 <span className="text-sm font-normal text-gray-500">USDC</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content - Market List */}
        <div className="lg:col-span-2">
          <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold">Active Markets</h2>
              <button
                onClick={loadMarkets}
                disabled={loading}
                className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors disabled:opacity-50"
              >
                <RefreshCw className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
              </button>
            </div>

            {loading ? (
              <div className="flex flex-col items-center justify-center py-12 text-gray-500">
                <RefreshCw className="w-8 h-8 animate-spin mb-2" />
                <p>Loading markets...</p>
              </div>
            ) : error ? (
              <div className="text-center py-12 text-red-400">
                <p>{error}</p>
                <button onClick={loadMarkets} className="mt-2 text-sm underline">Try Again</button>
              </div>
            ) : markets.length > 0 ? (
              <div className="grid gap-4">
                {markets.map((market) => (
                  <MarketCard
                    key={market.id}
                    market={market}
                    onTrade={(id) => console.log('Trade market:', id)}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-12 text-gray-500">
                No active markets found.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default PolymarketPage
