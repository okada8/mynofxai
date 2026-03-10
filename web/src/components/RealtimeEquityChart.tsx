import React, { useState, useEffect } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  AreaChart,
  Area,
} from 'recharts'
import useSWR from 'swr'
import { api } from '../lib/api'
import { useLanguage } from '../contexts/LanguageContext'
import { useAuth } from '../contexts/AuthContext'
import { t } from '../i18n/translations'
import {
  AlertTriangle,
  BarChart3,
  DollarSign,
  Percent,
  TrendingUp as ArrowUp,
  TrendingDown as ArrowDown,
  Activity,
  Zap,
} from 'lucide-react'

interface EquityPoint {
  timestamp: string
  total_equity: number
  pnl: number
  pnl_pct: number
  cycle_number: number
}

interface RealtimeEquityChartProps {
  traderId?: string
  embedded?: boolean
}

export function RealtimeEquityChart({
  traderId,
  embedded = false,
}: RealtimeEquityChartProps) {
  const { language } = useLanguage()
  const { user, token } = useAuth()
  const [displayMode, setDisplayMode] = useState<'dollar' | 'percent'>('dollar')
  const [isLive, setIsLive] = useState(false)

  // Use a faster refresh interval for real-time feel
  const {
    data: history,
    error,
    isLoading,
    isValidating,
  } = useSWR<EquityPoint[]>(
    user && token && traderId ? `equity-history-${traderId}` : null,
    () => api.getEquityHistory(traderId),
    {
      refreshInterval: 2000, // 2 seconds refresh
      revalidateOnFocus: true,
      dedupingInterval: 1000,
      onSuccess: () => setIsLive(true),
      onError: () => setIsLive(false),
    }
  )

  const { data: account } = useSWR(
    user && token && traderId ? `account-${traderId}` : null,
    () => api.getAccount(traderId),
    {
      refreshInterval: 2000, // Sync with equity chart
      revalidateOnFocus: true,
      dedupingInterval: 1000,
    }
  )

  // Blink effect for the live indicator
  const [blink, setBlink] = useState(true)
  useEffect(() => {
    const interval = setInterval(() => {
      setBlink((b) => !b)
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  // Loading state
  if (isLoading && !history) {
    return (
      <div className={embedded ? 'p-6 h-full' : 'binance-card p-6'}>
        {!embedded && (
          <h3 className="text-lg font-semibold mb-6 text-[#EAECEF]">
            {t('accountEquityCurve', language)}
          </h3>
        )}
        <div className="animate-pulse h-full flex flex-col justify-center">
          <div className="skeleton h-64 w-full rounded bg-white/5"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div
        className={
          embedded
            ? 'p-6 h-full flex items-center justify-center'
            : 'binance-card p-6'
        }
      >
        <div className="flex items-center gap-3 p-4 rounded bg-nofx-red/10 border border-nofx-red/20">
          <AlertTriangle className="w-6 h-6 text-nofx-red" />
          <div>
            <div className="font-semibold text-nofx-red">
              {t('loadingError', language)}
            </div>
            <div className="text-sm text-nofx-text-muted">{error.message}</div>
          </div>
        </div>
      </div>
    )
  }

  // Filter valid data
  const validHistory = history?.filter((point) => point.total_equity > 1) || []

  if (!validHistory || validHistory.length === 0) {
    return (
      <div
        className={
          embedded
            ? 'p-6 h-full flex flex-col justify-center'
            : 'binance-card p-6'
        }
      >
        {!embedded && (
          <h3 className="text-lg font-semibold mb-6 text-[#EAECEF]">
            {t('accountEquityCurve', language)}
          </h3>
        )}
        <div className="text-center py-16 text-[#848E9C]">
          <div className="mb-4 flex justify-center opacity-50">
            <BarChart3 className="w-16 h-16" />
          </div>
          <div className="text-lg font-semibold mb-2">
            {t('noHistoricalData', language)}
          </div>
          <div className="text-sm">{t('dataWillAppear', language)}</div>
        </div>
      </div>
    )
  }

  // Limit display points for performance but keep enough for trend
  const MAX_DISPLAY_POINTS = 500
  const displayHistory =
    validHistory.length > MAX_DISPLAY_POINTS
      ? validHistory.slice(-MAX_DISPLAY_POINTS)
      : validHistory

  // Calculate initial balance
  const initialBalance =
    account?.initial_balance ||
    (validHistory[0]
      ? validHistory[0].total_equity - validHistory[0].pnl
      : undefined) ||
    1000

  // Transform data
  const chartData = displayHistory.map((point) => {
    const pnl = point.total_equity - initialBalance
    const pnlPct = ((pnl / initialBalance) * 100).toFixed(2)
    return {
      time: new Date(point.timestamp).toLocaleTimeString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      }),
      value: displayMode === 'dollar' ? point.total_equity : parseFloat(pnlPct),
      cycle: point.cycle_number,
      raw_equity: point.total_equity,
      raw_pnl: pnl,
      raw_pnl_pct: parseFloat(pnlPct),
      timestamp: new Date(point.timestamp).getTime(),
    }
  })

  const currentValue = chartData[chartData.length - 1]
  const isProfit = currentValue.raw_pnl >= 0

  // Y-Axis domain calculation
  const calculateYDomain = () => {
    const values = chartData.map((d) => d.value)

    if (displayMode === 'percent') {
      const minVal = Math.min(...values)
      const maxVal = Math.max(...values)
      const range = Math.max(Math.abs(maxVal), Math.abs(minVal))
      const padding = Math.max(range * 0.2, 1)
      return [Math.floor(minVal - padding), Math.ceil(maxVal + padding)]
    } else {
      const minVal = Math.min(...values, initialBalance)
      const maxVal = Math.max(...values, initialBalance)
      const range = maxVal - minVal
      const padding = Math.max(range * 0.15, initialBalance * 0.01)
      return [Math.floor(minVal - padding), Math.ceil(maxVal + padding)]
    }
  }

  // Custom Tooltip
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload
      return (
        <div className="rounded-lg p-3 shadow-xl backdrop-blur-md bg-[#1E2329]/90 border border-white/10 ring-1 ring-black/50">
          <div className="flex items-center justify-between gap-4 mb-2 border-b border-white/10 pb-2">
            <div className="text-xs text-[#848E9C] font-mono">{label}</div>
            <div className="text-xs text-nofx-gold font-mono flex items-center gap-1">
              <Activity className="w-3 h-3" /> Cycle #{data.cycle}
            </div>
          </div>
          <div className="font-bold mono text-lg mb-1 text-[#EAECEF]">
            {data.raw_equity.toFixed(2)}{' '}
            <span className="text-xs text-[#848E9C]">USDT</span>
          </div>
          <div
            className="text-sm mono font-bold flex items-center gap-1"
            style={{ color: data.raw_pnl >= 0 ? '#0ECB81' : '#F6465D' }}
          >
            {data.raw_pnl >= 0 ? (
              <ArrowUp className="w-3 h-3" />
            ) : (
              <ArrowDown className="w-3 h-3" />
            )}
            {data.raw_pnl >= 0 ? '+' : ''}
            {data.raw_pnl.toFixed(2)} ({data.raw_pnl_pct >= 0 ? '+' : ''}
            {data.raw_pnl_pct}%)
          </div>
        </div>
      )
    }
    return null
  }

  return (
    <div
      className={
        embedded
          ? 'p-3 sm:p-5 h-full flex flex-col'
          : 'binance-card p-3 sm:p-5 animate-fade-in'
      }
    >
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between mb-4 shrink-0">
        <div className="flex-1">
          {!embedded && (
            <h3 className="text-base sm:text-lg font-bold mb-2 text-[#EAECEF]">
              {t('accountEquityCurve', language)}
            </h3>
          )}
          <div className="flex flex-col sm:flex-row sm:items-baseline gap-2 sm:gap-4">
            <span className="text-2xl sm:text-3xl font-bold mono text-[#EAECEF] flex items-center gap-2">
              {account?.total_equity.toFixed(2) || '0.00'}
              <span className="text-base sm:text-lg text-[#848E9C]">USDT</span>
            </span>
            <div className="flex items-center gap-2 flex-wrap">
              <span
                className="text-sm sm:text-lg font-bold mono px-2 sm:px-3 py-0.5 rounded flex items-center gap-1 transition-colors duration-300"
                style={{
                  color: isProfit ? '#0ECB81' : '#F6465D',
                  background: isProfit
                    ? 'rgba(14, 203, 129, 0.1)'
                    : 'rgba(246, 70, 93, 0.1)',
                  border: `1px solid ${isProfit ? 'rgba(14, 203, 129, 0.2)' : 'rgba(246, 70, 93, 0.2)'}`,
                }}
              >
                {isProfit ? (
                  <ArrowUp className="w-4 h-4" />
                ) : (
                  <ArrowDown className="w-4 h-4" />
                )}
                {isProfit ? '+' : ''}
                {currentValue.raw_pnl_pct}%
              </span>
            </div>

            {/* Live Indicator */}
            <div
              className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full border transition-all duration-300 ${
                isValidating
                  ? 'bg-nofx-gold/10 border-nofx-gold/30 text-nofx-gold'
                  : 'bg-white/5 border-white/10 text-nofx-text-muted'
              }`}
            >
              <div
                className={`w-1.5 h-1.5 rounded-full ${blink ? 'bg-current shadow-[0_0_8px_currentColor]' : 'bg-current/30'}`}
              />
              <span className="text-[10px] font-bold uppercase tracking-wider">
                {isValidating ? 'LIVE' : 'SYNC'}
              </span>
            </div>
          </div>
        </div>

        {/* Display Mode Toggle */}
        <div className="flex gap-0.5 sm:gap-1 rounded p-0.5 sm:p-1 self-start sm:self-auto bg-[#0B0E11] border border-[#2B3139]">
          <button
            onClick={() => setDisplayMode('dollar')}
            className={`px-3 sm:px-4 py-1.5 sm:py-2 rounded text-xs sm:text-sm font-bold transition-all flex items-center gap-1 ${
              displayMode === 'dollar'
                ? 'bg-nofx-gold text-black shadow-[0_2px_8px_rgba(240,185,11,0.4)]'
                : 'bg-transparent text-[#848E9C] hover:text-white'
            }`}
          >
            <DollarSign className="w-4 h-4" /> USDT
          </button>
          <button
            onClick={() => setDisplayMode('percent')}
            className={`px-3 sm:px-4 py-1.5 sm:py-2 rounded text-xs sm:text-sm font-bold transition-all flex items-center gap-1 ${
              displayMode === 'percent'
                ? 'bg-nofx-gold text-black shadow-[0_2px_8px_rgba(240,185,11,0.4)]'
                : 'bg-transparent text-[#848E9C] hover:text-white'
            }`}
          >
            <Percent className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Chart */}
      <div className="flex-1 min-h-0 w-full relative rounded-lg overflow-hidden">
        {/* NOFX Watermark */}
        <div className="absolute top-4 right-4 text-xl font-bold text-nofx-gold/10 font-mono pointer-events-none z-10 select-none">
          NOFX REALTIME
        </div>

        <ResponsiveContainer width="100%" height="100%">
          <AreaChart
            data={chartData}
            margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
          >
            <defs>
              <linearGradient id="colorGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#F0B90B" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#F0B90B" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="#2B3139"
              vertical={false}
            />
            <XAxis
              dataKey="time"
              stroke="#5E6673"
              tick={{ fill: '#848E9C', fontSize: 10, fontFamily: 'monospace' }}
              tickLine={false}
              axisLine={false}
              minTickGap={30}
            />
            <YAxis
              stroke="#5E6673"
              tick={{ fill: '#848E9C', fontSize: 10, fontFamily: 'monospace' }}
              tickLine={false}
              axisLine={false}
              domain={calculateYDomain()}
              tickFormatter={(value) =>
                displayMode === 'dollar' ? `${value.toFixed(0)}` : `${value}%`
              }
              width={40}
            />
            <Tooltip
              content={<CustomTooltip />}
              cursor={{
                stroke: '#F0B90B',
                strokeWidth: 1,
                strokeDasharray: '4 4',
              }}
            />
            <ReferenceLine
              y={displayMode === 'dollar' ? initialBalance : 0}
              stroke="#474D57"
              strokeDasharray="3 3"
              label={{
                value: displayMode === 'dollar' ? 'Init' : '0%',
                fill: '#848E9C',
                fontSize: 10,
                position: 'right',
              }}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#F0B90B"
              strokeWidth={2}
              fill="url(#colorGradient)"
              animationDuration={500}
              isAnimationActive={true}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* Footer Stats */}
      <div className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-2 sm:gap-3 pt-3 border-t border-[#2B3139] shrink-0">
        <StatItem
          label={t('initialBalance', language)}
          value={`${initialBalance.toFixed(2)} USDT`}
        />
        <StatItem
          label={t('currentEquity', language)}
          value={`${currentValue.raw_equity.toFixed(2)} USDT`}
          highlight
        />
        <StatItem
          label={t('historicalCycles', language)}
          value={`${validHistory.length} ${t('cycles', language)}`}
        />
        <StatItem
          label="Refresh Rate"
          value="2.0s"
          icon={<Zap className="w-3 h-3 text-nofx-gold" />}
        />
      </div>
    </div>
  )
}

function StatItem({
  label,
  value,
  highlight,
  icon,
}: {
  label: string
  value: string
  highlight?: boolean
  icon?: React.ReactNode
}) {
  return (
    <div className="p-2 rounded bg-nofx-gold/5 hover:bg-nofx-gold/10 transition-colors">
      <div className="text-[10px] mb-1 uppercase tracking-wider text-[#848E9C] flex items-center gap-1">
        {label}
        {icon}
      </div>
      <div
        className={`text-xs sm:text-sm font-bold mono ${highlight ? 'text-[#EAECEF]' : 'text-[#848E9C]'}`}
      >
        {value}
      </div>
    </div>
  )
}
