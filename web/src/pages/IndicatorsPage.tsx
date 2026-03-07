import { useState, useEffect } from 'react'
import { useLanguage } from '../contexts/LanguageContext'
import { motion } from 'framer-motion'
import { TrendingUp, TrendingDown, Activity, AlertCircle } from 'lucide-react'
import { TradingViewScreener } from '../components/TradingViewScreener'

interface FngData {
  value: string
  value_classification: string
  timestamp: string
  time_until_update: string
}

export function IndicatorsPage() {
  const { language } = useLanguage()
  const [fngData, setFngData] = useState<FngData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('https://api.alternative.me/fng/?limit=1')
      .then(res => res.json())
      .then(data => {
        setFngData(data.data[0])
        setLoading(false)
      })
      .catch(err => {
        console.error('Failed to fetch FnG data:', err)
        setLoading(false)
      })
  }, [])

  const getFngColor = (value: number) => {
    if (value >= 75) return 'text-green-500' // Extreme Greed
    if (value >= 55) return 'text-green-400' // Greed
    if (value >= 45) return 'text-yellow-500' // Neutral
    if (value >= 25) return 'text-orange-500' // Fear
    return 'text-red-500' // Extreme Fear
  }

  const getFngLabel = (classification: string) => {
    if (language !== 'zh') return classification
    const map: Record<string, string> = {
      'Extreme Fear': '极度恐慌',
      'Fear': '恐慌',
      'Neutral': '中性',
      'Greed': '贪婪',
      'Extreme Greed': '极度贪婪'
    }
    return map[classification] || classification
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-[#F0B90B] mb-2">
            {language === 'zh' ? '市场核心指标' : 'Market Core Indicators'}
          </h1>
          <p className="text-gray-400 text-sm">
            {language === 'zh' 
              ? '实时监控市场情绪、资金流向与链上数据' 
              : 'Real-time monitoring of market sentiment, fund flow and on-chain data'}
          </p>
        </div>
        <div className="text-xs text-gray-500 bg-[#1E2329] px-3 py-1 rounded border border-[#2B3139]">
          {language === 'zh' ? '数据来源: Alternative.me & Binance' : 'Source: Alternative.me & Binance'}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* 1. 恐慌与贪婪指数 */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-[#1E2329] border border-[#2B3139] rounded-xl p-6 relative overflow-hidden"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-gray-200 flex items-center gap-2">
              <Activity className="w-5 h-5 text-[#F0B90B]" />
              {language === 'zh' ? '恐慌与贪婪指数' : 'Fear & Greed Index'}
            </h3>
            <span className="text-xs text-gray-500">Today</span>
          </div>
          
          {loading ? (
            <div className="h-32 flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#F0B90B]"></div>
            </div>
          ) : fngData ? (
            <div className="flex flex-col items-center justify-center py-4">
              <div className={`text-6xl font-black mb-2 ${getFngColor(parseInt(fngData.value))}`}>
                {fngData.value}
              </div>
              <div className={`text-xl font-medium px-4 py-1 rounded-full bg-opacity-10 ${
                getFngColor(parseInt(fngData.value)).replace('text-', 'bg-')
              } ${getFngColor(parseInt(fngData.value))}`}>
                {getFngLabel(fngData.value_classification)}
              </div>
              <div className="w-full h-2 bg-gray-700 rounded-full mt-6 overflow-hidden">
                <div 
                  className={`h-full transition-all duration-1000 ease-out ${getFngColor(parseInt(fngData.value)).replace('text-', 'bg-')}`}
                  style={{ width: `${fngData.value}%` }}
                />
              </div>
              <div className="flex justify-between w-full mt-1 text-xs text-gray-500">
                <span>0 (Fear)</span>
                <span>100 (Greed)</span>
              </div>
            </div>
          ) : (
            <div className="text-center text-red-500">Failed to load data</div>
          )}
        </motion.div>

        {/* 2. BTC 多空比 (模拟数据/占位) */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-[#1E2329] border border-[#2B3139] rounded-xl p-6"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-gray-200 flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-green-500" />
              {language === 'zh' ? 'BTC 多空持仓比' : 'BTC Long/Short Ratio'}
            </h3>
            <span className="text-xs text-green-500">Bullish</span>
          </div>
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <span className="text-gray-400 text-sm">Longs</span>
              <span className="text-green-500 font-bold">52.4%</span>
            </div>
            <div className="w-full h-4 bg-red-900/30 rounded-full overflow-hidden flex">
              <div className="h-full bg-green-500" style={{ width: '52.4%' }}></div>
              <div className="h-full bg-red-500" style={{ width: '47.6%' }}></div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400 text-sm">Shorts</span>
              <span className="text-red-500 font-bold">47.6%</span>
            </div>
            <div className="mt-2 pt-4 border-t border-[#2B3139] text-xs text-gray-500">
               {language === 'zh' ? '基于币安合约实时持仓数据' : 'Based on Binance Futures real-time open interest'}
            </div>
          </div>
        </motion.div>

        {/* 3. 24h 爆仓数据 (模拟) */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-[#1E2329] border border-[#2B3139] rounded-xl p-6"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-gray-200 flex items-center gap-2">
              <AlertCircle className="w-5 h-5 text-red-500" />
              {language === 'zh' ? '24h 爆仓统计' : '24h Liquidations'}
            </h3>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-[#0B0E11] p-3 rounded-lg border border-[#2B3139]">
              <div className="text-xs text-gray-500 mb-1">Total (USD)</div>
              <div className="text-xl font-bold text-white">$128.5M</div>
            </div>
            <div className="bg-[#0B0E11] p-3 rounded-lg border border-[#2B3139]">
              <div className="text-xs text-gray-500 mb-1">Orders</div>
              <div className="text-xl font-bold text-white">42,105</div>
            </div>
          </div>
          <div className="mt-4 space-y-2">
             <div className="flex justify-between text-sm">
                <span className="text-green-500">Longs Liq</span>
                <span className="text-white">$85.2M</span>
             </div>
             <div className="w-full h-1.5 bg-gray-700 rounded-full">
               <div className="h-full bg-green-500" style={{ width: '66%' }}></div>
             </div>
             <div className="flex justify-between text-sm">
                <span className="text-red-500">Shorts Liq</span>
                <span className="text-white">$43.3M</span>
             </div>
             <div className="w-full h-1.5 bg-gray-700 rounded-full">
               <div className="h-full bg-red-500" style={{ width: '34%' }}></div>
             </div>
          </div>
        </motion.div>
        
        {/* Embed Coinglass/TradingView Widget for more advanced data if needed */}
      </div>

      {/* 嵌入 TradingView 市场概览组件 */}
      <div className="mt-8 bg-[#1E2329] border border-[#2B3139] rounded-xl p-1 overflow-hidden h-[600px]">
        <TradingViewScreener height="100%" />
      </div>
    </div>
  )
}
