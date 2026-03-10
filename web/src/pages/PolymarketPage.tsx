import React, { useState, useEffect } from 'react'
import useSWR from 'swr'
import { useTradersConfigStore } from '../stores'
import { api } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { TraderInfo } from '../types'
import { MarketCard } from '../components/Polymarket/MarketCard'
import { TradeModal } from '../components/Polymarket/TradeModal'
import { PositionList } from '../components/Polymarket/PositionList'
import {
  polymarketService,
  PolymarketMarket,
  PolymarketBalance,
  PolymarketPosition,
} from '../services/polymarket'
import { Wallet, RefreshCw, AlertTriangle, TrendingUp } from 'lucide-react'
import { toast } from 'sonner'

const PolymarketPage: React.FC = () => {
  const { user, token } = useAuth()
  const { allExchanges, loadConfigs } = useTradersConfigStore()
  const { data: traders } = useSWR<TraderInfo[]>(
    user && token ? 'traders' : null,
    api.getTraders
  )

  const [markets, setMarkets] = useState<PolymarketMarket[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [balance, setBalance] = useState<PolymarketBalance | null>(null)
  const [positions, setPositions] = useState<PolymarketPosition[]>([])
  const [positionsLoading, setPositionsLoading] = useState(false)

  // Modal state
  const [selectedMarket, setSelectedMarket] = useState<PolymarketMarket | null>(
    null
  )
  const [isTradeModalOpen, setIsTradeModalOpen] = useState(false)

  // Load configs on mount
  useEffect(() => {
    loadConfigs(user, token)
  }, [user, token, loadConfigs])

  // Find configured Polymarket trader
  const polymarketTrader = traders?.find((t) => {
    const exchange = allExchanges.find((e) => e.id === t.exchange_id)
    return exchange?.exchange_type === 'polymarket'
  })

  const polymarketExchange = polymarketTrader
    ? allExchanges.find((e) => e.id === polymarketTrader.exchange_id)
    : undefined
  
  const isConfigured = !!polymarketTrader

  useEffect(() => {
    loadMarkets()
    if (isConfigured && polymarketTrader) {
      loadBalance(polymarketTrader.trader_id)
      loadPositions(polymarketTrader.trader_id)
    }
  }, [isConfigured, polymarketTrader])

  const loadMarkets = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await polymarketService.getMarkets(20)
      setMarkets(data)
    } catch (err) {
      console.error('Failed to load markets:', err)
      setError('Failed to load markets. Please check your connection.')
    } finally {
      setLoading(false)
    }
  }

  const loadBalance = async (traderId: string) => {
    try {
      const data = await polymarketService.getBalance(traderId)
      setBalance(data)
    } catch (err) {
      console.error('Failed to load balance:', err)
    }
  }

  const loadPositions = async (traderId: string) => {
    setPositionsLoading(true)
    try {
      const data = await polymarketService.getPositions(traderId)
      setPositions(data)
    } catch (err) {
      console.error('Failed to load positions:', err)
    } finally {
      setPositionsLoading(false)
    }
  }

  const handleTrade = (market: PolymarketMarket) => {
    if (!isConfigured) {
      toast.error('Please configure a Polymarket trader first')
      return
    }
    setSelectedMarket(market)
    setIsTradeModalOpen(true)
  }

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
              Please add a Polymarket trader in the Trader Configuration to
              start trading.
            </p>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Sidebar - Account Info & Positions */}
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
            <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <Wallet className="w-5 h-5 text-purple-400" /> Account Info
            </h2>
            <div className="space-y-4">
              <div>
                <div className="text-sm text-gray-400 mb-1">Wallet Address</div>
                <div className="font-mono text-sm bg-gray-900/50 p-2 rounded truncate text-gray-300">
                  {polymarketExchange?.polymarketWalletAddr || 'Not Configured'}
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
                  ${balance?.collateral_value?.toFixed(2) || '0.00'}{' '}
                  <span className="text-sm font-normal text-gray-500">
                    USDC
                  </span>
                </div>
                {balance && (
                  <div className="text-xs text-gray-500 mt-1">
                    Cash: ${balance.cash?.toFixed(2)}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Positions List */}
          {isConfigured && (
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-green-400" /> My Positions
              </h2>

              {positionsLoading ? (
                <div className="flex justify-center py-8 text-gray-500">
                  <RefreshCw className="w-6 h-6 animate-spin" />
                </div>
              ) : positions.length === 0 ? (
                <div className="text-center py-8 text-gray-500 text-sm">
                  No active positions found.
                </div>
              ) : (
                <div className="space-y-3">
                  {positions.map((pos, idx) => (
                    <div
                      key={pos.asset_id || idx}
                      className="bg-gray-900/50 rounded-lg p-3 border border-gray-800 hover:border-gray-700 transition-colors"
                    >
                      <div className="flex justify-between items-start mb-2">
                        <div className="font-medium text-gray-200 text-sm break-all pr-2">
                          {pos.symbol || `Asset ${pos.asset_id.slice(0, 8)}...`}
                        </div>
                        <div className="text-right whitespace-nowrap">
                          <div className="text-sm font-bold text-white">
                            ${pos.value_usd?.toFixed(2) ?? '0.00'}
                          </div>
                        </div>
                      </div>

                      <div className="flex justify-between items-center text-xs text-gray-400">
                        <div className="flex gap-3">
                          <span>
                            Size:{' '}
                            <span className="text-gray-300">{pos.size}</span>
                          </span>
                          <span>
                            Avg:{' '}
                            <span className="text-gray-300">
                              ${pos.price?.toFixed(2) ?? '-'}
                            </span>
                          </span>
                        </div>
                        <div
                          className={
                            pos.size > 0 ? 'text-green-500' : 'text-gray-500'
                          }
                        >
                          {pos.size > 0 ? 'LONG' : 'FLAT'}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Main Content - Market List */}
        <div className="lg:col-span-2">
          <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-blue-400" /> Active Markets
              </h2>
              <button
                onClick={loadMarkets}
                disabled={loading}
                className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors disabled:opacity-50"
              >
                <RefreshCw
                  className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`}
                />
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
                <button
                  onClick={loadMarkets}
                  className="mt-2 text-sm underline"
                >
                  Try Again
                </button>
              </div>
            ) : markets.length > 0 ? (
              <div className="grid gap-4">
                {markets.map((market) => (
                  <MarketCard
                    key={market.id}
                    market={market}
                    onTrade={() => handleTrade(market)}
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

      {/* Trade Modal */}
      {selectedMarket && polymarketTrader && (
        <TradeModal
          isOpen={isTradeModalOpen}
          onClose={() => setIsTradeModalOpen(false)}
          market={selectedMarket}
          traderId={polymarketTrader.trader_id}
          onSuccess={() => {
            toast.success('Order placed successfully!')
            loadBalance(polymarketTrader.trader_id)
            loadPositions(polymarketTrader.trader_id)
          }}
        />
      )}
    </div>
  )
}

export default PolymarketPage
