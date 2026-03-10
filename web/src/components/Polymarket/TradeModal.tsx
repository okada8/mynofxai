import React, { useState } from 'react'
import { PolymarketMarket } from '../../services/polymarket'
import { polymarketService } from '../../services/polymarket'
import { Loader2 } from 'lucide-react'

interface TradeModalProps {
  isOpen: boolean
  onClose: () => void
  market: PolymarketMarket
  traderId: string
  onSuccess: () => void
}

export const TradeModal: React.FC<TradeModalProps> = ({
  isOpen,
  onClose,
  market,
  traderId,
  onSuccess,
}) => {
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY') // In Polymarket context, typically BUY Yes/No
  const [outcome, setOutcome] = useState<'YES' | 'NO'>('YES')
  const [quantity, setQuantity] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      // In this simple integration, we map YES/NO to specific logic or token IDs
      // For now, let's assume the API handles "YES" or "NO" string as symbol or side logic
      // Ideally, we'd pass the specific outcome Token ID.
      // Let's assume the backend expects symbol to be the Market ID + outcome?
      // Or simply "YES" / "NO" and the backend resolves it for the given market?
      // Based on previous backend code:
      // handler.CreateOrder expects { symbol, side, quantity }
      // And backend does: OpenLong(symbol) -> buy outcome token.

      // We need to pass the correct Token ID for the outcome.
      // The market object has outcomes array.
      // outcome[0] is typically YES, outcome[1] is NO (or vice versa, need to check).
      // Let's assume outcomes[0] = Yes, outcomes[1] = No for binary.

      const outcomeIndex = outcome === 'YES' ? 0 : 1
      const outcomeTokenId =
        market.outcomes[outcomeIndex]?.id || (outcome === 'YES' ? 'YES' : 'NO') // Fallback if ID missing

      // Note: The backend implementation of OpenLong(symbol) expects a symbol/tokenID.
      // If we pass the market ID, the backend might not know which outcome.
      // Let's pass the specific outcome ID if available, otherwise construct a logical one.

      await polymarketService.createOrder({
        trader_id: traderId,
        symbol: outcomeTokenId, // Passing the token ID of the outcome
        side: 'BUY', // We are buying shares of the outcome
        quantity: parseFloat(quantity),
      })

      onSuccess()
      onClose()
    } catch (err: any) {
      console.error('Trade failed:', err)
      setError(err.response?.data?.error || 'Failed to place order')
    } finally {
      setLoading(false)
    }
  }

  const currentPrice = outcome === 'YES' ? market.yesPrice : market.noPrice

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="bg-gray-900 border border-gray-800 rounded-xl w-full max-w-md p-6 relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-white"
        >
          ✕
        </button>

        <h2 className="text-xl font-bold mb-4">Trade Market</h2>
        <p className="text-gray-400 text-sm mb-6 line-clamp-2">
          {market.question}
        </p>

        <div className="flex gap-4 mb-6">
          <button
            onClick={() => setOutcome('YES')}
            className={`flex-1 py-3 rounded-lg border font-medium transition-colors ${
              outcome === 'YES'
                ? 'bg-green-500/20 border-green-500 text-green-400'
                : 'bg-gray-800 border-gray-700 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Buy YES
            <div className="text-sm opacity-80">
              {(market.yesPrice * 100).toFixed(1)}%
            </div>
          </button>
          <button
            onClick={() => setOutcome('NO')}
            className={`flex-1 py-3 rounded-lg border font-medium transition-colors ${
              outcome === 'NO'
                ? 'bg-red-500/20 border-red-500 text-red-400'
                : 'bg-gray-800 border-gray-700 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Buy NO
            <div className="text-sm opacity-80">
              {(market.noPrice * 100).toFixed(1)}%
            </div>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-1">
              Amount (USDC)
            </label>
            <div className="relative">
              <input
                type="number"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                placeholder="0.00"
                min="0"
                step="0.01"
                required
                className="w-full bg-gray-950 border border-gray-800 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-purple-500 transition-colors"
              />
              <span className="absolute right-4 top-3 text-gray-500">USDC</span>
            </div>
          </div>

          <div className="bg-gray-800/50 rounded-lg p-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-400">Price per share</span>
              <span className="text-white">${currentPrice.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-400">Est. Shares</span>
              <span className="text-white">
                {quantity && !isNaN(parseFloat(quantity))
                  ? (parseFloat(quantity) / currentPrice).toFixed(2)
                  : '0.00'}
              </span>
            </div>
            <div className="flex justify-between border-t border-gray-700 pt-2 mt-2">
              <span className="text-gray-400">Total Cost</span>
              <span className="font-bold text-white">
                ${quantity || '0.00'}
              </span>
            </div>
          </div>

          {error && (
            <div className="p-3 bg-red-900/20 border border-red-800 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !quantity}
            className="w-full bg-purple-600 hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-3 rounded-lg transition-colors flex items-center justify-center gap-2"
          >
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            {loading ? 'Placing Order...' : 'Confirm Trade'}
          </button>
        </form>
      </div>
    </div>
  )
}
