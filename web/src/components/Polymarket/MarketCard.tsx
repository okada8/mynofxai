import React from 'react'
import { PolymarketMarket } from '../../services/polymarket'

interface MarketCardProps {
  market: PolymarketMarket
  onTrade: () => void
}

export function MarketCard({ market, onTrade }: MarketCardProps) {
  const outcomes = market.outcomes || []

  return (
    <div className="rounded-lg p-4 border border-gray-800 bg-gray-900/50 hover:border-purple-500/50 transition-all">
      <div className="flex justify-between items-start mb-3">
        <div>
          <h3 className="text-lg font-medium text-white line-clamp-2 mb-1">
            {market.question}
          </h3>
          <div className="flex gap-2 text-xs text-gray-400">
            <span>Liquidity: ${market.liquidity?.toLocaleString() || '0'}</span>
            <span>•</span>
            <span>Vol: ${market.volume24h?.toLocaleString() || '0'}</span>
            <span>•</span>
            <span>
              Ends:{' '}
              {market.endDate
                ? new Date(market.endDate).toLocaleDateString()
                : 'N/A'}
            </span>
          </div>
        </div>
        <button
          onClick={onTrade}
          className="px-3 py-1 text-sm bg-purple-600 hover:bg-purple-700 text-white rounded transition-colors"
        >
          Trade
        </button>
      </div>

      <div className="space-y-2">
        {outcomes.slice(0, 2).map((outcome: any, idx: number) => {
          const price = parseFloat(outcome.price || '0')
          const percentage = Math.round(price * 100)

          return (
            <div
              key={outcome.id || idx}
              className="relative h-8 bg-gray-800 rounded overflow-hidden flex items-center px-3"
            >
              <div
                className={`absolute left-0 top-0 bottom-0 opacity-20 ${idx === 0 ? 'bg-green-500' : 'bg-red-500'}`}
                style={{ width: `${percentage}%` }}
              />
              <div className="relative flex justify-between w-full text-sm">
                <span className="font-medium text-gray-200">
                  {idx === 0 ? 'Yes' : 'No'}
                </span>
                <span className={idx === 0 ? 'text-green-400' : 'text-red-400'}>
                  {percentage}% ({price.toFixed(2)}¢)
                </span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
