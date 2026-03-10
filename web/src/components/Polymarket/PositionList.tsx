import React from 'react'
import { PolymarketPosition } from '../../services/polymarket'
import { Loader2, TrendingUp } from 'lucide-react'

interface PositionListProps {
  positions: PolymarketPosition[]
  loading: boolean
}

export const PositionList: React.FC<PositionListProps> = ({
  positions,
  loading,
}) => {
  return (
    <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
      <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
        <TrendingUp className="w-5 h-5 text-green-400" /> My Positions
      </h2>

      {loading ? (
        <div className="flex justify-center py-8 text-gray-500">
          <Loader2 className="w-6 h-6 animate-spin" />
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
                    Size: <span className="text-gray-300">{pos.size}</span>
                  </span>
                  <span>
                    Avg Price:{' '}
                    <span className="text-gray-300">
                      ${pos.price?.toFixed(2) ?? '-'}
                    </span>
                  </span>
                </div>
                <div
                  className={pos.size > 0 ? 'text-green-500' : 'text-gray-500'}
                >
                  {pos.size > 0 ? 'LONG' : 'FLAT'}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
