import { useState } from 'react'
import useSWR from 'swr'
import { useLanguage } from '../contexts/LanguageContext'
import { TradingViewChart } from '../components/TradingViewChart'
import {
  nofxosApi,
  type AI500Coin,
  type OIRankingItem,
  type FundFlowItem,
  type QueryRankItem,
  type HeatmapItem,
  type FundingRateItem,
  type AI300Item,
  type AI500Response,
  type AI300Response,
  type QueryRankResponse,
  type FundingRateResponse,
  type PriceRankingResponse,
} from '../services/nofxos'
import {
  Search,
  LayoutDashboard,
  LineChart,
  TrendingUp,
  TrendingDown,
  ArrowUpRight,
  ArrowDownRight,
  RefreshCcw,
  Activity,
  DollarSign,
  Layers,
  Percent,
  Search as SearchIcon,
  Brain,
} from 'lucide-react'

// Format helper
const formatCompactNumber = (num: number) => {
  if (num === undefined || num === null) return '-'
  return new Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 1,
  }).format(num)
}

const formatPrice = (num: number) => {
  if (num === undefined || num === null) return '-'
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 6,
  }).format(num)
}

const PriceChange = ({ value }: { value: number }) => {
  if (!value) return <span className="text-gray-500">-</span>
  const isPositive = value > 0
  return (
    <span
      className={`flex items-center text-xs font-medium ${isPositive ? 'text-[#0ECB81]' : 'text-[#F6465D]'}`}
    >
      {isPositive ? (
        <ArrowUpRight className="w-3 h-3" />
      ) : (
        <ArrowDownRight className="w-3 h-3" />
      )}
      {Math.abs(value).toFixed(2)}%
    </span>
  )
}

// 热门币种列表 (for Chart Mode)
const MARKET_PAIRS = [
  { symbol: 'BTCUSDT', name: 'Bitcoin' },
  { symbol: 'ETHUSDT', name: 'Ethereum' },
  { symbol: 'SOLUSDT', name: 'Solana' },
  { symbol: 'BNBUSDT', name: 'Binance Coin' },
  { symbol: 'XRPUSDT', name: 'XRP' },
  { symbol: 'DOGEUSDT', name: 'Dogecoin' },
  { symbol: 'ADAUSDT', name: 'Cardano' },
  { symbol: 'AVAXUSDT', name: 'Avalanche' },
  { symbol: 'DOTUSDT', name: 'Polkadot' },
  { symbol: 'LINKUSDT', name: 'Chainlink' },
]

type TerminalTab = 'netflow' | 'oi' | 'depth' | 'price' | 'rates' | 'ai_model'

export function DataPage() {
  const { language } = useLanguage()
  const [viewMode, setViewMode] = useState<'terminal' | 'chart'>('terminal')
  const [activeTab, setActiveTab] = useState<TerminalTab>('ai_model')
  const [duration, setDuration] = useState('1h')
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT')
  const [searchTerm, setSearchTerm] = useState('')

  // --- Terminal Data Fetching ---

  // AI500 & Query Rank (AI Model Tab) - No duration support usually
  const {
    data: ai500Data,
    error: ai500Error,
    mutate: refreshAi500,
  } = useSWR<AI500Response>(
    viewMode === 'terminal' && activeTab === 'ai_model' ? 'ai500' : null,
    () => nofxosApi.getAI500()
  )
  const {
    data: queryRankData,
    error: queryError,
    mutate: refreshQuery,
  } = useSWR<QueryRankResponse>(
    viewMode === 'terminal' && activeTab === 'ai_model' ? 'query-rank' : null,
    () => nofxosApi.getQueryRank()
  )

  // OI (OI Tab)
  const {
    data: oiTop,
    error: topError,
    mutate: refreshTop,
  } = useSWR<OIRankingItem[]>(
    viewMode === 'terminal' && activeTab === 'oi' ? `oi-top-${duration}` : null,
    () => nofxosApi.getOITopRanking(duration, 20)
  )
  const {
    data: oiLow,
    error: lowError,
    mutate: refreshLow,
  } = useSWR<OIRankingItem[]>(
    viewMode === 'terminal' && activeTab === 'oi' ? `oi-low-${duration}` : null,
    () => nofxosApi.getOILowRanking(duration, 20)
  )

  // Netflow (Netflow Tab)
  const {
    data: netflowTop,
    error: netflowTopError,
    mutate: refreshNetflowTop,
  } = useSWR<FundFlowItem[]>(
    viewMode === 'terminal' && activeTab === 'netflow'
      ? `netflow-top-${duration}`
      : null,
    () => nofxosApi.getNetflowTopRanking(duration, 20)
  )
  const {
    data: netflowLow,
    error: netflowLowError,
    mutate: refreshNetflowLow,
  } = useSWR<FundFlowItem[]>(
    viewMode === 'terminal' && activeTab === 'netflow'
      ? `netflow-low-${duration}`
      : null,
    () => nofxosApi.getNetflowLowRanking(duration, 20)
  )

  // Heatmap (Depth Tab)
  const {
    data: heatmapFuture,
    error: heatmapFutureError,
    mutate: refreshHeatmapFuture,
  } = useSWR<HeatmapItem[]>(
    viewMode === 'terminal' && activeTab === 'depth' ? 'heatmap-future' : null,
    () => nofxosApi.getHeatmap('future', 20)
  )
  const {
    data: heatmapSpot,
    error: heatmapSpotError,
    mutate: refreshHeatmapSpot,
  } = useSWR<HeatmapItem[]>(
    viewMode === 'terminal' && activeTab === 'depth' ? 'heatmap-spot' : null,
    () => nofxosApi.getHeatmap('spot', 20)
  )

  // AI300 (AI Model Tab)
  const {
    data: ai300Data,
    error: ai300Error,
    mutate: refreshAi300,
  } = useSWR<AI300Response>(
    viewMode === 'terminal' && activeTab === 'ai_model' ? 'ai300' : null,
    () => nofxosApi.getAI300()
  )

  // Funding Rates (Rates Tab)
  const {
    data: fundingTopData,
    error: fundingTopError,
    mutate: refreshFundingTop,
  } = useSWR<FundingRateResponse>(
    viewMode === 'terminal' && activeTab === 'rates' ? 'funding-top' : null,
    () => nofxosApi.getFundingRateTop(20)
  )
  const {
    data: fundingLowData,
    error: fundingLowError,
    mutate: refreshFundingLow,
  } = useSWR<FundingRateResponse>(
    viewMode === 'terminal' && activeTab === 'rates' ? 'funding-low' : null,
    () => nofxosApi.getFundingRateLow(20)
  )

  // Price Ranking (Price Tab)
  const {
    data: priceRanking,
    error: priceRankingError,
    mutate: refreshPriceRanking,
  } = useSWR<PriceRankingResponse>(
    viewMode === 'terminal' && activeTab === 'price'
      ? `price-ranking-${duration}`
      : null,
    () => nofxosApi.getPriceRanking(duration, 20)
  )

  const handleRefresh = () => {
    if (activeTab === 'ai_model') {
      refreshAi500()
      refreshQuery()
      refreshAi300()
    }
    if (activeTab === 'oi') {
      refreshTop()
      refreshLow()
    }
    if (activeTab === 'netflow') {
      refreshNetflowTop()
      refreshNetflowLow()
    }
    if (activeTab === 'depth') {
      refreshHeatmapFuture()
      refreshHeatmapSpot()
    }
    if (activeTab === 'rates') {
      refreshFundingTop()
      refreshFundingLow()
    }
    if (activeTab === 'price') {
      refreshPriceRanking()
    }
  }

  // Filter for Chart Mode
  const filteredPairs = MARKET_PAIRS.filter(
    (pair) =>
      pair.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pair.name.toLowerCase().includes(searchTerm.toLowerCase())
  )

  // Generic Table Renderer
  const renderRankingTable = (
    title: string,
    subtitle: string,
    icon: React.ReactNode,
    data: any[] | undefined,
    error: any,
    columns: {
      header: string
      align?: 'left' | 'right'
      render: (item: any, index: number) => React.ReactNode
    }[],
    isLowRanking = false
  ) => (
    <div className="bg-[#1E2329] border border-[#2B3139] rounded-xl overflow-hidden flex flex-col h-[600px]">
      <div className="p-4 border-b border-[#2B3139] flex items-center justify-between bg-gradient-to-r from-[#1E2329] to-[#2B3139]/30">
        <div className="flex items-center gap-2">
          <div
            className={`w-8 h-8 rounded-lg flex items-center justify-center border ${isLowRanking ? 'bg-red-500/10 border-red-500/20' : 'bg-green-500/10 border-green-500/20'}`}
          >
            {icon}
          </div>
          <div>
            <h3 className="font-bold text-white text-sm">{title}</h3>
            <p className="text-[10px] text-gray-400">
              {subtitle} ({duration})
            </p>
          </div>
        </div>
      </div>
      <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700">
        {error ? (
          <div className="h-full flex items-center justify-center text-red-400 text-xs">
            Failed to load API data
          </div>
        ) : !data ? (
          <div className="h-full flex items-center justify-center text-gray-500 text-xs">
            Loading...
          </div>
        ) : data.length === 0 ? (
          <div className="h-full flex items-center justify-center text-gray-500 text-xs">
            No data available
          </div>
        ) : (
          <table className="w-full text-sm text-left">
            <thead className="text-xs text-gray-500 bg-[#15181D] sticky top-0 z-10">
              <tr>
                {columns.map((col, i) => (
                  <th
                    key={i}
                    className={`px-4 py-3 font-medium ${col.align === 'right' ? 'text-right' : ''}`}
                  >
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-[#2B3139]">
              {data.map((item, index) => (
                <tr
                  key={index}
                  className="hover:bg-[#2B3139]/50 transition-colors group"
                >
                  {columns.map((col, i) => (
                    <td
                      key={i}
                      className={`px-4 py-3 ${col.align === 'right' ? 'text-right' : ''}`}
                    >
                      {col.render(item, index)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )

  const renderPlaceholder = (title: string, icon: React.ReactNode) => (
    <div className="h-[400px] flex flex-col items-center justify-center text-gray-500">
      <div className="w-16 h-16 rounded-full bg-[#1E2329] flex items-center justify-center mb-4 border border-[#2B3139]">
        {icon}
      </div>
      <h3 className="text-lg font-bold text-gray-300 mb-2">{title}</h3>
      <p className="text-sm">This module is currently under development.</p>
    </div>
  )

  return (
    <div className="flex flex-col h-[calc(100dvh-64px)] bg-[#0B0E11] text-white overflow-hidden">
      {/* Top Navigation Bar */}
      <div className="flex-none flex flex-col border-b border-[#2B3139] bg-[#1E2329]">
        <div className="flex items-center justify-between px-6 py-3 border-b border-[#2B3139]/50">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setViewMode('terminal')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-bold transition-all ${
                viewMode === 'terminal'
                  ? 'bg-[#F0B90B] text-black shadow-lg shadow-yellow-500/20'
                  : 'text-gray-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <LayoutDashboard className="w-4 h-4" />
              NOFX Terminal
            </button>
            <button
              onClick={() => setViewMode('chart')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-bold transition-all ${
                viewMode === 'chart'
                  ? 'bg-[#F0B90B] text-black shadow-lg shadow-yellow-500/20'
                  : 'text-gray-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <LineChart className="w-4 h-4" />
              TradingView
            </button>
          </div>

          <div className="flex items-center gap-4">
            {viewMode === 'terminal' && (
              <button
                onClick={handleRefresh}
                className="p-2 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 transition-colors"
                title="Refresh Data"
              >
                <RefreshCcw className="w-4 h-4" />
              </button>
            )}
            <div className="text-xs text-gray-500 font-mono hidden sm:block">
              {viewMode === 'terminal'
                ? 'Powered by NOFX OS API'
                : 'Market Data from Binance'}
            </div>
          </div>
        </div>

        {/* Secondary Navigation (Tabs) for Terminal */}
        {viewMode === 'terminal' && (
          <div className="px-6 py-2 border-b border-[#2B3139]/30">
            <nav className="flex items-center justify-between w-full">
              <div className="flex items-center gap-2 overflow-x-auto scrollbar-none">
                {[
                  { id: 'netflow', label: 'NETFLOW', icon: DollarSign },
                  { id: 'oi', label: 'OI', icon: TrendingUp },
                  { id: 'depth', label: 'DEPTH', icon: Layers },
                  { id: 'price', label: 'PRICE', icon: Activity },
                  { id: 'rates', label: 'RATES', icon: Percent },
                  { id: 'ai_model', label: 'AI MODEL', icon: Brain },
                ].map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id as TerminalTab)}
                    className={`flex items-center gap-2 px-4 py-1.5 rounded text-sm font-bold transition-all whitespace-nowrap border ${
                      activeTab === tab.id
                        ? 'text-[#F0B90B] bg-[#F0B90B]/10 border-[#F0B90B] shadow-[0_0_10px_rgba(240,185,11,0.2)]'
                        : 'text-gray-500 border-transparent hover:text-gray-300 hover:bg-white/5'
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              {/* Duration Selector (Only for Netflow, OI, Price) */}
              {(activeTab === 'netflow' ||
                activeTab === 'oi' ||
                activeTab === 'price') && (
                <div className="flex items-center gap-1 bg-[#0B0E11] p-1 rounded-lg border border-[#2B3139] ml-4 flex-none">
                  {['5m', '15m', '30m', '1h', '4h', '8h', '24h'].map((d) => (
                    <button
                      key={d}
                      onClick={() => setDuration(d)}
                      className={`px-3 py-1 text-xs font-bold rounded transition-all ${
                        duration === d
                          ? 'text-[#F0B90B] bg-[#F0B90B]/20'
                          : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
                      }`}
                    >
                      {d.toUpperCase()}
                    </button>
                  ))}
                </div>
              )}
            </nav>
          </div>
        )}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 relative overflow-hidden bg-[#0B0E11]">
        {viewMode === 'terminal' ? (
          <div className="h-full overflow-y-auto p-6 scrollbar-thin scrollbar-thumb-gray-800">
            {/* --- NETFLOW TAB --- */}
            {activeTab === 'netflow' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-[1920px] mx-auto">
                {renderRankingTable(
                  'Netflow Inflow',
                  'Top Institutional Inflow',
                  <DollarSign className="w-4 h-4 text-green-500" />,
                  netflowTop,
                  netflowTopError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Inflow',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-mono">
                          +${formatCompactNumber(item.amount)}
                        </span>
                      ),
                    },
                    {
                      header: 'Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-400 font-mono">
                          ${formatPrice(item.price)}
                        </span>
                      ),
                    },
                  ]
                )}
                {renderRankingTable(
                  'Netflow Outflow',
                  'Top Institutional Outflow',
                  <DollarSign className="w-4 h-4 text-red-500" />,
                  netflowLow,
                  netflowLowError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Outflow',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-mono">
                          -${formatCompactNumber(Math.abs(item.amount))}
                        </span>
                      ),
                    },
                    {
                      header: 'Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-400 font-mono">
                          ${formatPrice(item.price)}
                        </span>
                      ),
                    },
                  ],
                  true
                )}
              </div>
            )}

            {/* --- OI TAB --- */}
            {activeTab === 'oi' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-[1920px] mx-auto">
                {renderRankingTable(
                  'OI Surge Ranking',
                  'Open Interest Change',
                  <TrendingUp className="w-4 h-4 text-green-500" />,
                  oiTop,
                  topError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'OI Change',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-300 font-mono">
                          {formatCompactNumber(item.oi_delta)}
                        </span>
                      ),
                    },
                    {
                      header: 'OI Delta %',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-bold font-mono">
                          +{item.oi_delta_percent?.toFixed(2)}%
                        </span>
                      ),
                    },
                    {
                      header: 'Price %',
                      align: 'right',
                      render: (item) => (
                        <div className="flex justify-end">
                          <PriceChange value={item.price_delta_percent} />
                        </div>
                      ),
                    },
                  ]
                )}
                {renderRankingTable(
                  'OI Plunge Ranking',
                  'Open Interest Change',
                  <TrendingDown className="w-4 h-4 text-red-500" />,
                  oiLow,
                  lowError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'OI Change',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-300 font-mono">
                          {formatCompactNumber(item.oi_delta)}
                        </span>
                      ),
                    },
                    {
                      header: 'OI Delta %',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-bold font-mono">
                          {item.oi_delta_percent?.toFixed(2)}%
                        </span>
                      ),
                    },
                    {
                      header: 'Price %',
                      align: 'right',
                      render: (item) => (
                        <div className="flex justify-end">
                          <PriceChange value={item.price_delta_percent} />
                        </div>
                      ),
                    },
                  ],
                  true
                )}
              </div>
            )}

            {/* --- AI MODEL TAB --- */}
            {activeTab === 'ai_model' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6 max-w-[1920px] mx-auto">
                {/* AI500 Index */}
                {renderRankingTable(
                  'AI500 Index',
                  ai500Data?.desc || 'High Potential Coins (Score > 70)',
                  <Activity className="w-4 h-4 text-blue-500" />,
                  ai500Data?.coins,
                  ai500Error,
                  [
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.pair?.replace('USDT', '') || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Start Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-300 font-mono">
                          ${formatPrice(item.start_price)}
                        </span>
                      ),
                    },
                    {
                      header: 'Gain %',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-bold font-mono ${item.increase_percent >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {item.increase_percent > 0 ? '+' : ''}
                          {item.increase_percent?.toFixed(2) || '0.00'}%
                        </span>
                      ),
                    },
                    {
                      header: 'Score',
                      align: 'right',
                      render: (item) => (
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                          {item.score?.toFixed(1) || '-'}
                        </span>
                      ),
                    },
                  ]
                )}
                {/* AI300 Index */}
                {renderRankingTable(
                  'AI300 Index',
                  ai300Data?.desc || 'Smart Money Flow Analysis',
                  <Activity className="w-4 h-4 text-purple-500" />,
                  ai300Data?.coins,
                  ai300Error,
                  [
                    {
                      header: 'Rank',
                      render: (item) => (
                        <span className="text-gray-500">
                          {item.rank || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Future Flow',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-mono ${item.future_flow >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {formatCompactNumber(item.future_flow)}
                        </span>
                      ),
                    },
                    {
                      header: 'Level',
                      align: 'right',
                      render: (item) => (
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-bold bg-purple-500/10 text-purple-400 border border-purple-500/20">
                          {item.level || '-'}
                        </span>
                      ),
                    },
                  ]
                )}
                {/* Query Rank */}
                {renderRankingTable(
                  'Search Trending',
                  queryRankData?.desc || 'Most Queried Coins Today',
                  <SearchIcon className="w-4 h-4 text-yellow-500" />,
                  queryRankData?.rankings,
                  queryError,
                  [
                    {
                      header: 'Rank',
                      render: (item) => (
                        <span className="text-gray-500">
                          {item.rank || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Queries',
                      align: 'right',
                      render: (item) => (
                        <span className="text-yellow-400 font-bold">
                          {item.query_count || '0'}
                        </span>
                      ),
                    },
                    {
                      header: 'Future Flow',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-mono ${item.future_flow >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {formatCompactNumber(item.future_flow)}
                        </span>
                      ),
                    },
                  ]
                )}
              </div>
            )}

            {activeTab === 'price' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-[1920px] mx-auto">
                {renderRankingTable(
                  'Top Gainers',
                  'Price Increase Ranking',
                  <TrendingUp className="w-4 h-4 text-green-500" />,
                  priceRanking?.top,
                  priceRankingError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol?.replace('USDT', '') || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-300 font-mono">
                          ${formatPrice(item.price)}
                        </span>
                      ),
                    },
                    {
                      header: 'Change',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-bold font-mono">
                          +{(item.price_delta * 100).toFixed(2)}%
                        </span>
                      ),
                    },
                    {
                      header: 'OI Delta',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-mono ${item.oi_delta >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {formatCompactNumber(item.oi_delta)}
                        </span>
                      ),
                    },
                  ]
                )}
                {renderRankingTable(
                  'Top Losers',
                  'Price Decrease Ranking',
                  <TrendingDown className="w-4 h-4 text-red-500" />,
                  priceRanking?.low,
                  priceRankingError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol?.replace('USDT', '') || '-'}
                        </span>
                      ),
                    },
                    {
                      header: 'Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-300 font-mono">
                          ${formatPrice(item.price)}
                        </span>
                      ),
                    },
                    {
                      header: 'Change',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-bold font-mono">
                          {(item.price_delta * 100).toFixed(2)}%
                        </span>
                      ),
                    },
                    {
                      header: 'OI Delta',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-mono ${item.oi_delta >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {formatCompactNumber(item.oi_delta)}
                        </span>
                      ),
                    },
                  ],
                  true
                )}
              </div>
            )}

            {activeTab === 'depth' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-[1920px] mx-auto">
                {renderRankingTable(
                  'Future Market Depth',
                  'Bid/Ask Volume Delta',
                  <Layers className="w-4 h-4 text-blue-500" />,
                  heatmapFuture,
                  heatmapFutureError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Bid Vol',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-mono">
                          {formatCompactNumber(item.bid_volume)}
                        </span>
                      ),
                    },
                    {
                      header: 'Ask Vol',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-mono">
                          {formatCompactNumber(item.ask_volume)}
                        </span>
                      ),
                    },
                    {
                      header: 'Delta',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-bold font-mono ${item.delta >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {item.delta > 0 ? '+' : ''}
                          {formatCompactNumber(item.delta)}
                        </span>
                      ),
                    },
                  ]
                )}
                {renderRankingTable(
                  'Spot Market Depth',
                  'Bid/Ask Volume Delta',
                  <Layers className="w-4 h-4 text-purple-500" />,
                  heatmapSpot,
                  heatmapSpotError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Bid Vol',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-mono">
                          {formatCompactNumber(item.bid_volume)}
                        </span>
                      ),
                    },
                    {
                      header: 'Ask Vol',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-mono">
                          {formatCompactNumber(item.ask_volume)}
                        </span>
                      ),
                    },
                    {
                      header: 'Delta',
                      align: 'right',
                      render: (item) => (
                        <span
                          className={`font-bold font-mono ${item.delta >= 0 ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {item.delta > 0 ? '+' : ''}
                          {formatCompactNumber(item.delta)}
                        </span>
                      ),
                    },
                  ]
                )}
              </div>
            )}
            {activeTab === 'rates' && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-[1920px] mx-auto">
                {renderRankingTable(
                  'Top Funding Rates',
                  fundingTopData?.desc || 'Long Crowded (High Rate)',
                  <Percent className="w-4 h-4 text-green-500" />,
                  fundingTopData?.rates,
                  fundingTopError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Rate',
                      align: 'right',
                      render: (item) => (
                        <span className="text-green-400 font-mono font-bold">
                          {item.funding_rate.toFixed(4)}%
                        </span>
                      ),
                    },
                    {
                      header: 'Index Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-400 font-mono">
                          ${formatPrice(item.index_price)}
                        </span>
                      ),
                    },
                  ]
                )}
                {renderRankingTable(
                  'Low Funding Rates',
                  fundingLowData?.desc || 'Short Crowded (Negative Rate)',
                  <Percent className="w-4 h-4 text-red-500" />,
                  fundingLowData?.rates,
                  fundingLowError,
                  [
                    {
                      header: 'Rank',
                      render: (_, i) => (
                        <span className="text-gray-500">{i + 1}</span>
                      ),
                    },
                    {
                      header: 'Token',
                      render: (item) => (
                        <span className="font-bold text-white">
                          {item.symbol.replace('USDT', '')}
                        </span>
                      ),
                    },
                    {
                      header: 'Rate',
                      align: 'right',
                      render: (item) => (
                        <span className="text-red-400 font-mono font-bold">
                          {item.funding_rate.toFixed(4)}%
                        </span>
                      ),
                    },
                    {
                      header: 'Index Price',
                      align: 'right',
                      render: (item) => (
                        <span className="text-gray-400 font-mono">
                          ${formatPrice(item.index_price)}
                        </span>
                      ),
                    },
                  ],
                  true
                )}
              </div>
            )}

            {/* Footer Info */}
            <div className="mt-6 text-center pb-6">
              <p className="text-xs text-gray-600">
                Data provided by NofxOS API (Native Integration).
              </p>
            </div>
          </div>
        ) : (
          <div className="flex h-full w-full">
            {/* Chart Mode Sidebar */}
            <div className="w-72 border-r border-[#2B3139] flex flex-col bg-[#1E2329] flex-none">
              <div className="p-4 border-b border-[#2B3139]">
                <div className="relative">
                  <input
                    type="text"
                    placeholder={
                      language === 'zh' ? '搜索币种...' : 'Search coin...'
                    }
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full bg-[#0B0E11] border border-[#2B3139] rounded-lg py-2 pl-9 pr-4 text-sm focus:outline-none focus:border-[#F0B90B] text-gray-300 placeholder-gray-600 transition-colors"
                  />
                  <Search className="absolute left-3 top-2.5 w-4 h-4 text-gray-500" />
                </div>
              </div>
              <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent">
                {filteredPairs.map((pair) => (
                  <button
                    key={pair.symbol}
                    onClick={() => setSelectedSymbol(pair.symbol)}
                    className={`w-full px-4 py-3 flex items-center gap-3 hover:bg-[#2B3139] transition-all border-l-4 ${
                      selectedSymbol === pair.symbol
                        ? 'bg-[#2B3139] border-[#F0B90B]'
                        : 'border-transparent'
                    }`}
                  >
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs ${
                        selectedSymbol === pair.symbol
                          ? 'bg-[#F0B90B] text-black'
                          : 'bg-gray-800 text-gray-400'
                      }`}
                    >
                      {pair.symbol.substring(0, 1)}
                    </div>
                    <div className="flex flex-col items-start">
                      <span
                        className={`font-bold text-sm ${selectedSymbol === pair.symbol ? 'text-[#F0B90B]' : 'text-gray-200'}`}
                      >
                        {pair.symbol.replace('USDT', '')}
                      </span>
                      <span className="text-xs text-gray-500">{pair.name}</span>
                    </div>
                  </button>
                ))}
              </div>
            </div>
            <div className="flex-1 flex flex-col bg-[#0B0E11] relative overflow-hidden">
              <TradingViewChart
                defaultSymbol={selectedSymbol}
                defaultExchange="BINANCE"
                height="100%"
                showToolbar={true}
                embedded={true}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
