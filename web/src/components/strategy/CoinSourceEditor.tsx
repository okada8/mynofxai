import { useState } from 'react'
import {
  Plus,
  X,
  Database,
  TrendingUp,
  TrendingDown,
  List,
  Ban,
  Zap,
  Shuffle,
  Scan,
  Activity,
  BarChart2,
  Globe,
  Layers,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react'
import type { CoinSourceConfig } from '../../types'

interface CoinSourceEditorProps {
  config: CoinSourceConfig
  onChange: (config: CoinSourceConfig) => void
  disabled?: boolean
  language: string
}

export function CoinSourceEditor({
  config,
  onChange,
  disabled,
  language,
}: CoinSourceEditorProps) {
  const [newCoin, setNewCoin] = useState('')
  const [newExcludedCoin, setNewExcludedCoin] = useState('')

  const t = (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      sourceType: { zh: '数据来源类型', en: 'Source Type' },
      static: { zh: '静态列表', en: 'Static List' },
      ai500: { zh: 'AI500 数据源', en: 'AI500 Data Provider' },
      oi_top: { zh: 'OI 持仓增加', en: 'OI Increase' },
      oi_low: { zh: 'OI 持仓减少', en: 'OI Decrease' },
      screener: { zh: '视觉筛选器', en: 'Visual Screener' },
      mixed: { zh: '混合模式', en: 'Mixed Mode' },
      dynamic_macro: { zh: '动态宏观', en: 'Dynamic Macro' },
      staticCoins: { zh: '自定义币种', en: 'Custom Coins' },
      addCoin: { zh: '添加币种', en: 'Add Coin' },
      useAI500: { zh: '启用 AI500 数据源', en: 'Enable AI500 Data Provider' },
      ai500Limit: { zh: '数量上限', en: 'Limit' },
      useOITop: { zh: '启用 OI 持仓增加榜', en: 'Enable OI Increase' },
      oiTopLimit: { zh: '数量上限', en: 'Limit' },
      useOILow: { zh: '启用 OI 持仓减少榜', en: 'Enable OI Decrease' },
      oiLowLimit: { zh: '数量上限', en: 'Limit' },
      useScreener: { zh: '启用视觉筛选器', en: 'Enable Visual Screener' },
      screenerLimit: { zh: '数量上限', en: 'Limit' },
      screenerDuration: { zh: '筛选周期', en: 'Duration' },
      screenerSortBy: { zh: '排序指标', en: 'Sort By' },
      useGainersLosers: { zh: '启用涨跌幅榜', en: 'Enable Gainers/Losers' },
      gainersTop: { zh: '涨幅榜前N', en: 'Top Gainers' },
      losersTop: { zh: '跌幅榜前N', en: 'Top Losers' },
      macroScreening: { zh: '宏观筛选配置', en: 'Macro Screening Config' },
      enableMacroFilter: { zh: '启用宏观过滤', en: 'Enable Macro Filter' },
      maxSectorExposure: { zh: '最大板块敞口', en: 'Max Sector Exposure' },
      sectorAllocation: { zh: '板块配置', en: 'Sector Allocation' },
      oi: { zh: '持仓变化', en: 'OI Change' },
      price: { zh: '价格变化', en: 'Price Change' },
      vol: { zh: '成交量变化', en: 'Volume Change' },
      staticDesc: {
        zh: '手动指定交易币种列表',
        en: 'Manually specify trading coins',
      },
      ai500Desc: {
        zh: '使用 AI500 智能筛选的热门币种',
        en: 'Use AI500 smart-filtered popular coins',
      },
      oiTopDesc: {
        zh: '持仓增加榜，适合做多',
        en: 'OI increase ranking, for long',
      },
      oi_lowDesc: {
        zh: '持仓减少榜，适合做空',
        en: 'OI decrease ranking, for short',
      },
      screenerDesc: {
        zh: '基于 CoinAnk 视觉筛选器',
        en: 'Based on CoinAnk Screener',
      },
      mixedDesc: {
        zh: '组合多种数据源',
        en: 'Combine multiple sources',
      },
      dynamic_macroDesc: {
        zh: '宏观驱动的动态币种池',
        en: 'Macro-driven dynamic coin pool',
      },
      mixedConfig: {
        zh: '组合数据源配置',
        en: 'Combined Sources Configuration',
      },
      mixedSummary: { zh: '已选组合', en: 'Selected Sources' },
      maxCoins: { zh: '最多', en: 'Up to' },
      coins: { zh: '个币种', en: 'coins' },
      dataSourceConfig: { zh: '数据源配置', en: 'Data Source Configuration' },
      excludedCoins: { zh: '排除币种', en: 'Excluded Coins' },
      excludedCoinsDesc: {
        zh: '这些币种将从所有数据源中排除，不会被交易',
        en: 'These coins will be excluded from all sources and will not be traded',
      },
      addExcludedCoin: { zh: '添加排除', en: 'Add Excluded' },
      onlyBinanceSymbols: { zh: '仅限币安交易对', en: 'Only Binance Symbols' },
      nofxosNote: {
        zh: '使用 NofxOS API Key（在指标配置中设置）',
        en: 'Uses NofxOS API Key (set in Indicators config)',
      },
    }
    return translations[key]?.[language] || key
  }

  const sourceTypes = [
    { value: 'static', icon: List, color: '#848E9C' },
    { value: 'ai500', icon: Database, color: '#F0B90B' },
    { value: 'oi_top', icon: TrendingUp, color: '#0ECB81' },
    { value: 'oi_low', icon: TrendingDown, color: '#F6465D' },
    { value: 'screener', icon: Scan, color: '#8B5CF6' },
    { value: 'mixed', icon: Shuffle, color: '#60a5fa' },
    { value: 'dynamic_macro', icon: Globe, color: '#F59E0B' },
  ] as const

  // Calculate mixed mode summary
  const getMixedSummary = () => {
    const sources: string[] = []
    let totalLimit = 0

    if (config.use_ai500) {
      sources.push(`AI500(${config.ai500_limit || 10})`)
      totalLimit += config.ai500_limit || 10
    }
    if (config.use_oi_top) {
      sources.push(
        `${language === 'zh' ? 'OI增' : 'OI↑'}(${config.oi_top_limit || 10})`
      )
      totalLimit += config.oi_top_limit || 10
    }
    if (config.use_oi_low) {
      sources.push(
        `${language === 'zh' ? 'OI减' : 'OI↓'}(${config.oi_low_limit || 10})`
      )
      totalLimit += config.oi_low_limit || 10
    }
    if (config.use_screener) {
      sources.push(
        `${language === 'zh' ? '筛选器' : 'Screener'}(${config.screener_limit || 10}/${config.screener_duration || '1h'}/${config.screener_sort_by || 'oi'})`
      )
      totalLimit += config.screener_limit || 10
    }
    if (config.use_gainers_losers) {
      sources.push(
        `${language === 'zh' ? '涨跌榜' : 'G/L'}(+${config.gainers_top || 4}/-${config.losers_top || 4})`
      )
      totalLimit += (config.gainers_top || 4) + (config.losers_top || 4)
    }
    if ((config.static_coins || []).length > 0) {
      sources.push(
        `${language === 'zh' ? '自定义' : 'Custom'}(${config.static_coins?.length || 0})`
      )
      totalLimit += config.static_coins?.length || 0
    }

    return { sources, totalLimit }
  }

  // xyz dex assets (stocks, forex, commodities) - should NOT get USDT suffix
  const xyzDexAssets = new Set([
    // Stocks
    'TSLA',
    'NVDA',
    'AAPL',
    'MSFT',
    'META',
    'AMZN',
    'GOOGL',
    'AMD',
    'COIN',
    'NFLX',
    'PLTR',
    'HOOD',
    'INTC',
    'MSTR',
    'TSM',
    'ORCL',
    'MU',
    'RIVN',
    'COST',
    'LLY',
    'CRCL',
    'SKHX',
    'SNDK',
    // Forex
    'EUR',
    'JPY',
    // Commodities
    'GOLD',
    'SILVER',
    // Index
    'XYZ100',
  ])

  const isXyzDexAsset = (symbol: string): boolean => {
    const base = symbol
      .toUpperCase()
      .replace(/^XYZ:/, '')
      .replace(/USDT$|USD$|-USDC$/, '')
    return xyzDexAssets.has(base)
  }

  const handleAddCoin = () => {
    if (!newCoin.trim()) return
    const symbol = newCoin.toUpperCase().trim()

    // For xyz dex assets (stocks, forex, commodities), use xyz: prefix without USDT
    let formattedSymbol: string
    if (isXyzDexAsset(symbol)) {
      // Remove xyz: prefix (case-insensitive) and any USD suffixes
      const base = symbol
        .replace(/^xyz:/i, '')
        .replace(/USDT$|USD$|-USDC$/i, '')
      formattedSymbol = `xyz:${base}`
    } else {
      formattedSymbol = symbol.endsWith('USDT') ? symbol : `${symbol}USDT`
    }

    const currentCoins = config.static_coins || []
    if (!currentCoins.includes(formattedSymbol)) {
      onChange({
        ...config,
        static_coins: [...currentCoins, formattedSymbol],
      })
    }
    setNewCoin('')
  }

  const handleRemoveCoin = (coin: string) => {
    onChange({
      ...config,
      static_coins: (config.static_coins || []).filter((c) => c !== coin),
    })
  }

  const handleAddExcludedCoin = () => {
    if (!newExcludedCoin.trim()) return
    const symbol = newExcludedCoin.toUpperCase().trim()

    // For xyz dex assets, use xyz: prefix without USDT
    let formattedSymbol: string
    if (isXyzDexAsset(symbol)) {
      const base = symbol
        .replace(/^xyz:/i, '')
        .replace(/USDT$|USD$|-USDC$/i, '')
      formattedSymbol = `xyz:${base}`
    } else {
      formattedSymbol = symbol.endsWith('USDT') ? symbol : `${symbol}USDT`
    }

    const currentExcluded = config.excluded_coins || []
    if (!currentExcluded.includes(formattedSymbol)) {
      onChange({
        ...config,
        excluded_coins: [...currentExcluded, formattedSymbol],
      })
    }
    setNewExcludedCoin('')
  }

  const handleRemoveExcludedCoin = (coin: string) => {
    onChange({
      ...config,
      excluded_coins: (config.excluded_coins || []).filter((c) => c !== coin),
    })
  }

  // NofxOS badge component
  const NofxOSBadge = () => (
    <span className="text-[9px] px-1.5 py-0.5 rounded font-medium bg-purple-500/20 text-purple-400 border border-purple-500/30">
      NofxOS
    </span>
  )

  return (
    <div className="space-y-6">
      {/* Source Type Selector */}
      <div>
        <label className="block text-sm font-medium mb-3 text-nofx-text">
          {t('sourceType')}
        </label>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
          {sourceTypes.map(({ value, icon: Icon, color }) => (
            <button
              key={value}
              onClick={() =>
                !disabled &&
                onChange({
                  ...config,
                  source_type: value as CoinSourceConfig['source_type'],
                })
              }
              disabled={disabled}
              className={`p-4 rounded-lg border transition-all ${
                config.source_type === value
                  ? 'ring-2 ring-nofx-gold bg-nofx-gold/10'
                  : 'hover:bg-white/5 bg-nofx-bg'
              } border-nofx-gold/20`}
            >
              <Icon className="w-6 h-6 mx-auto mb-2" style={{ color }} />
              <div className="text-sm font-medium text-nofx-text">
                {t(value)}
              </div>
              <div className="text-xs mt-1 text-nofx-text-muted">
                {t(`${value}Desc`)}
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Static Coins - only for static mode */}
      {config.source_type === 'static' && (
        <div>
          <label className="block text-sm font-medium mb-3 text-nofx-text">
            {t('staticCoins')}
          </label>
          <div className="flex flex-wrap gap-2 mb-3">
            {(config.static_coins || []).map((coin) => (
              <span
                key={coin}
                className="flex items-center gap-1 px-3 py-1.5 rounded-full text-sm bg-nofx-bg-lighter text-nofx-text"
              >
                {coin}
                {!disabled && (
                  <button
                    onClick={() => handleRemoveCoin(coin)}
                    className="ml-1 hover:text-red-400 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
          {!disabled && (
            <div className="flex gap-2">
              <input
                type="text"
                value={newCoin}
                onChange={(e) => setNewCoin(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddCoin()}
                placeholder="BTC, ETH, SOL..."
                className="flex-1 px-4 py-2 rounded-lg bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
              />
              <button
                onClick={handleAddCoin}
                className="px-4 py-2 rounded-lg flex items-center gap-2 transition-colors bg-nofx-gold text-black hover:bg-yellow-500"
              >
                <Plus className="w-4 h-4" />
                {t('addCoin')}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Global Filter Options */}
      <div className="flex items-center gap-2">
        <label className="flex items-center gap-2 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={config.only_binance_symbols ?? true}
            onChange={(e) =>
              !disabled &&
              onChange({ ...config, only_binance_symbols: e.target.checked })
            }
            disabled={disabled}
            className="w-4 h-4 rounded accent-nofx-gold"
          />
          <span className="text-sm font-medium text-nofx-text">
            {t('onlyBinanceSymbols')}
          </span>
        </label>
      </div>

      {/* Excluded Coins */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <Ban className="w-4 h-4 text-nofx-danger" />
          <label className="text-sm font-medium text-nofx-text">
            {t('excludedCoins')}
          </label>
        </div>
        <p className="text-xs mb-3 text-nofx-text-muted">
          {t('excludedCoinsDesc')}
        </p>
        <div className="flex flex-wrap gap-2 mb-3">
          {(config.excluded_coins || []).map((coin) => (
            <span
              key={coin}
              className="flex items-center gap-1 px-3 py-1.5 rounded-full text-sm bg-nofx-danger/15 text-nofx-danger"
            >
              {coin}
              {!disabled && (
                <button
                  onClick={() => handleRemoveExcludedCoin(coin)}
                  className="ml-1 hover:text-white transition-colors"
                >
                  <X className="w-3 h-3" />
                </button>
              )}
            </span>
          ))}
          {(config.excluded_coins || []).length === 0 && (
            <span className="text-xs italic text-nofx-text-muted">
              {language === 'zh' ? '无' : 'None'}
            </span>
          )}
        </div>
        {!disabled && (
          <div className="flex gap-2">
            <input
              type="text"
              value={newExcludedCoin}
              onChange={(e) => setNewExcludedCoin(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleAddExcludedCoin()}
              placeholder="BTC, ETH, DOGE..."
              className="flex-1 px-4 py-2 rounded-lg text-sm bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
            />
            <button
              onClick={handleAddExcludedCoin}
              className="px-4 py-2 rounded-lg flex items-center gap-2 transition-colors text-sm bg-nofx-danger text-white hover:bg-red-600"
            >
              <Ban className="w-4 h-4" />
              {t('addExcludedCoin')}
            </button>
          </div>
        )}
      </div>

      {/* AI500 Options - only for ai500 mode */}
      {config.source_type === 'ai500' && (
        <div className="p-4 rounded-lg bg-nofx-gold/5 border border-nofx-gold/20">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Zap className="w-4 h-4 text-nofx-gold" />
              <span className="text-sm font-medium text-nofx-text">
                AI500 {t('dataSourceConfig')}
              </span>
              <NofxOSBadge />
            </div>
          </div>

          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.use_ai500}
                onChange={(e) =>
                  !disabled &&
                  onChange({ ...config, use_ai500: e.target.checked })
                }
                disabled={disabled}
                className="w-5 h-5 rounded accent-nofx-gold"
              />
              <span className="text-nofx-text">{t('useAI500')}</span>
            </label>

            {config.use_ai500 && (
              <div className="flex items-center gap-3 pl-8">
                <span className="text-sm text-nofx-text-muted">
                  {t('ai500Limit')}:
                </span>
                <select
                  value={config.ai500_limit || 10}
                  onChange={(e) =>
                    !disabled &&
                    onChange({
                      ...config,
                      ai500_limit: parseInt(e.target.value) || 10,
                    })
                  }
                  disabled={disabled}
                  className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                >
                  {[5, 10, 15, 20, 30, 50].map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <p className="text-xs pl-8 text-nofx-text-muted">
              {t('nofxosNote')}
            </p>
          </div>
        </div>
      )}

      {/* OI Top Options - only for oi_top mode */}
      {config.source_type === 'oi_top' && (
        <div className="p-4 rounded-lg bg-nofx-success/5 border border-nofx-success/20">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <TrendingUp className="w-4 h-4 text-nofx-success" />
              <span className="text-sm font-medium text-nofx-text">
                OI {language === 'zh' ? '持仓增加榜' : 'Increase'}{' '}
                {t('dataSourceConfig')}
              </span>
              <NofxOSBadge />
            </div>
          </div>

          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.use_oi_top}
                onChange={(e) =>
                  !disabled &&
                  onChange({ ...config, use_oi_top: e.target.checked })
                }
                disabled={disabled}
                className="w-5 h-5 rounded accent-nofx-success"
              />
              <span className="text-nofx-text">{t('useOITop')}</span>
            </label>

            {config.use_oi_top && (
              <div className="flex items-center gap-3 pl-8">
                <span className="text-sm text-nofx-text-muted">
                  {t('oiTopLimit')}:
                </span>
                <select
                  value={config.oi_top_limit || 10}
                  onChange={(e) =>
                    !disabled &&
                    onChange({
                      ...config,
                      oi_top_limit: parseInt(e.target.value) || 10,
                    })
                  }
                  disabled={disabled}
                  className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                >
                  {[5, 10, 15, 20, 30, 50].map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <p className="text-xs pl-8 text-nofx-text-muted">
              {t('nofxosNote')}
            </p>
          </div>
        </div>
      )}

      {/* OI Low Options - only for oi_low mode */}
      {config.source_type === 'oi_low' && (
        <div className="p-4 rounded-lg bg-nofx-danger/5 border border-nofx-danger/20">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <TrendingDown className="w-4 h-4 text-nofx-danger" />
              <span className="text-sm font-medium text-nofx-text">
                OI {language === 'zh' ? '持仓减少榜' : 'Decrease'}{' '}
                {t('dataSourceConfig')}
              </span>
              <NofxOSBadge />
            </div>
          </div>

          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.use_oi_low}
                onChange={(e) =>
                  !disabled &&
                  onChange({ ...config, use_oi_low: e.target.checked })
                }
                disabled={disabled}
                className="w-5 h-5 rounded accent-red-500"
              />
              <span className="text-nofx-text">{t('useOILow')}</span>
            </label>

            {config.use_oi_low && (
              <div className="flex items-center gap-3 pl-8">
                <span className="text-sm text-nofx-text-muted">
                  {t('oiLowLimit')}:
                </span>
                <select
                  value={config.oi_low_limit || 10}
                  onChange={(e) =>
                    !disabled &&
                    onChange({
                      ...config,
                      oi_low_limit: parseInt(e.target.value) || 10,
                    })
                  }
                  disabled={disabled}
                  className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                >
                  {[5, 10, 15, 20, 30, 50].map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <p className="text-xs pl-8 text-nofx-text-muted">
              {t('nofxosNote')}
            </p>
          </div>
        </div>
      )}

      {/* Screener Options - only for screener mode */}
      {config.source_type === 'screener' && (
        <div className="p-4 rounded-lg bg-purple-500/5 border border-purple-500/20">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Scan className="w-4 h-4 text-purple-500" />
              <span className="text-sm font-medium text-nofx-text">
                {t('screener')} {t('dataSourceConfig')}
              </span>
              <NofxOSBadge />
            </div>
          </div>

          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.use_screener}
                onChange={(e) =>
                  !disabled &&
                  onChange({ ...config, use_screener: e.target.checked })
                }
                disabled={disabled}
                className="w-5 h-5 rounded accent-purple-500"
              />
              <span className="text-nofx-text">{t('useScreener')}</span>
            </label>

            {config.use_screener && (
              <>
                <div className="flex items-center gap-3 pl-8">
                  <span className="text-sm text-nofx-text-muted">
                    {t('screenerLimit')}:
                  </span>
                  <select
                    value={config.screener_limit || 10}
                    onChange={(e) =>
                      !disabled &&
                      onChange({
                        ...config,
                        screener_limit: parseInt(e.target.value) || 10,
                      })
                    }
                    disabled={disabled}
                    className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                  >
                    {[5, 10, 15, 20, 30, 50].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex items-center gap-3 pl-8">
                  <span className="text-sm text-nofx-text-muted">
                    {t('screenerDuration')}:
                  </span>
                  <select
                    value={config.screener_duration || '1h'}
                    onChange={(e) =>
                      !disabled &&
                      onChange({ ...config, screener_duration: e.target.value })
                    }
                    disabled={disabled}
                    className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                  >
                    {['5m', '15m', '30m', '1h', '4h'].map((d) => (
                      <option key={d} value={d}>
                        {d}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex items-center gap-3 pl-8">
                  <span className="text-sm text-nofx-text-muted">
                    {t('screenerSortBy')}:
                  </span>
                  <select
                    value={config.screener_sort_by || 'oi'}
                    onChange={(e) =>
                      !disabled &&
                      onChange({ ...config, screener_sort_by: e.target.value })
                    }
                    disabled={disabled}
                    className="px-3 py-1.5 rounded bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                  >
                    <option value="oi">{t('oi')}</option>
                    <option value="price">{t('price')}</option>
                    <option value="vol">{t('vol')}</option>
                  </select>
                </div>

                <p className="text-xs pl-8 text-nofx-text-muted">
                  {t('nofxosNote')}
                </p>
              </>
            )}
          </div>
        </div>
      )}

      {/* Mixed Mode & Dynamic Macro - Unified Card Selector */}
      {(config.source_type === 'mixed' ||
        config.source_type === 'dynamic_macro') && (
        <div
          className={`p-4 rounded-lg border ${
            config.source_type === 'dynamic_macro'
              ? 'bg-amber-500/5 border-amber-500/20'
              : 'bg-blue-500/5 border-blue-500/20'
          }`}
        >
          <div className="flex items-center gap-2 mb-4">
            {config.source_type === 'dynamic_macro' ? (
              <Globe className="w-4 h-4 text-amber-500" />
            ) : (
              <Shuffle className="w-4 h-4 text-blue-400" />
            )}
            <span className="text-sm font-medium text-nofx-text">
              {config.source_type === 'dynamic_macro'
                ? t('dynamic_macro')
                : t('mixedConfig')}
            </span>
          </div>

          {/* Macro Screening Section (Only for Dynamic Macro) */}
          {config.source_type === 'dynamic_macro' && (
            <div className="mb-4 p-3 rounded-lg border bg-nofx-bg border-amber-500/20">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Layers className="w-4 h-4 text-amber-500" />
                  <span className="text-sm font-medium text-nofx-text">
                    {t('macroScreening')}
                  </span>
                </div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={
                      config.macro_screening?.enable_macro_filter ?? false
                    }
                    onChange={(e) =>
                      !disabled &&
                      onChange({
                        ...config,
                        macro_screening: {
                          ...(config.macro_screening || {
                            max_sector_exposure: 0.4,
                            sector_allocation: {},
                          }),
                          enable_macro_filter: e.target.checked,
                        },
                      })
                    }
                    disabled={disabled}
                    className="w-4 h-4 rounded accent-amber-500"
                  />
                  <span className="text-xs text-nofx-text-muted">
                    {t('enableMacroFilter')}
                  </span>
                </label>
              </div>
              {config.macro_screening?.enable_macro_filter && (
                <div className="flex items-center gap-3 pl-6">
                  <span className="text-xs text-nofx-text-muted">
                    {t('maxSectorExposure')}:
                  </span>
                  <div className="flex items-center gap-1">
                    <input
                      type="number"
                      step="0.05"
                      min="0.1"
                      max="1.0"
                      value={config.macro_screening?.max_sector_exposure || 0.4}
                      onChange={(e) =>
                        !disabled &&
                        onChange({
                          ...config,
                          macro_screening: {
                            ...config.macro_screening!,
                            max_sector_exposure: parseFloat(e.target.value),
                          },
                        })
                      }
                      disabled={disabled}
                      className="w-16 px-2 py-1 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    />
                    <span className="text-xs text-nofx-text-muted">
                      (
                      {Math.round(
                        (config.macro_screening?.max_sector_exposure || 0.4) *
                          100
                      )}
                      %)
                    </span>
                  </div>
                </div>
              )}
            </div>
          )}

          <div className="grid grid-cols-2 gap-3 mb-4">
            {/* AI500 Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                config.use_ai500
                  ? 'bg-nofx-gold/10 border-nofx-gold/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-nofx-gold/30'
              }`}
              onClick={() =>
                !disabled &&
                onChange({ ...config, use_ai500: !config.use_ai500 })
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  checked={config.use_ai500}
                  onChange={(e) =>
                    !disabled &&
                    onChange({ ...config, use_ai500: e.target.checked })
                  }
                  disabled={disabled}
                  className="w-4 h-4 rounded accent-nofx-gold"
                  onClick={(e) => e.stopPropagation()}
                />
                <Database className="w-4 h-4 text-nofx-gold" />
                <span className="text-sm font-medium text-nofx-text">
                  AI500
                </span>
                <NofxOSBadge />
              </div>
              {config.use_ai500 && (
                <div className="flex items-center gap-2 mt-2 pl-6">
                  <span className="text-xs text-nofx-text-muted">Limit:</span>
                  <select
                    value={config.ai500_limit || 10}
                    onChange={(e) => {
                      e.stopPropagation()
                      !disabled &&
                        onChange({
                          ...config,
                          ai500_limit: parseInt(e.target.value) || 10,
                        })
                    }}
                    disabled={disabled}
                    className="px-2 py-1 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {[5, 10, 15, 20, 30, 50].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {/* OI Top Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                config.use_oi_top
                  ? 'bg-nofx-success/10 border-nofx-success/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-nofx-success/30'
              }`}
              onClick={() =>
                !disabled &&
                onChange({ ...config, use_oi_top: !config.use_oi_top })
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  checked={config.use_oi_top}
                  onChange={(e) =>
                    !disabled &&
                    onChange({ ...config, use_oi_top: e.target.checked })
                  }
                  disabled={disabled}
                  className="w-4 h-4 rounded accent-nofx-success"
                  onClick={(e) => e.stopPropagation()}
                />
                <TrendingUp className="w-4 h-4 text-nofx-success" />
                <span className="text-sm font-medium text-nofx-text">
                  {language === 'zh' ? 'OI 增加' : 'OI Increase'}
                </span>
              </div>
              <p className="text-xs text-nofx-text-muted pl-6 mb-1">
                {language === 'zh' ? '适合做多' : 'For long'}
              </p>
              {config.use_oi_top && (
                <div className="flex items-center gap-2 mt-2 pl-6">
                  <span className="text-xs text-nofx-text-muted">Limit:</span>
                  <select
                    value={config.oi_top_limit || 10}
                    onChange={(e) => {
                      e.stopPropagation()
                      !disabled &&
                        onChange({
                          ...config,
                          oi_top_limit: parseInt(e.target.value) || 10,
                        })
                    }}
                    disabled={disabled}
                    className="px-2 py-1 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {[5, 10, 15, 20, 30, 50].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {/* OI Low Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                config.use_oi_low
                  ? 'bg-nofx-danger/10 border-nofx-danger/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-nofx-danger/30'
              }`}
              onClick={() =>
                !disabled &&
                onChange({ ...config, use_oi_low: !config.use_oi_low })
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  checked={config.use_oi_low}
                  onChange={(e) =>
                    !disabled &&
                    onChange({ ...config, use_oi_low: e.target.checked })
                  }
                  disabled={disabled}
                  className="w-4 h-4 rounded accent-red-500"
                  onClick={(e) => e.stopPropagation()}
                />
                <TrendingDown className="w-4 h-4 text-nofx-danger" />
                <span className="text-sm font-medium text-nofx-text">
                  {language === 'zh' ? 'OI 减少' : 'OI Decrease'}
                </span>
              </div>
              <p className="text-xs text-nofx-text-muted pl-6 mb-1">
                {language === 'zh' ? '适合做空' : 'For short'}
              </p>
              {config.use_oi_low && (
                <div className="flex items-center gap-2 mt-2 pl-6">
                  <span className="text-xs text-nofx-text-muted">Limit:</span>
                  <select
                    value={config.oi_low_limit || 10}
                    onChange={(e) => {
                      e.stopPropagation()
                      !disabled &&
                        onChange({
                          ...config,
                          oi_low_limit: parseInt(e.target.value) || 10,
                        })
                    }}
                    disabled={disabled}
                    className="px-2 py-1 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {[5, 10, 15, 20, 30, 50].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {/* Gainers/Losers Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                config.use_gainers_losers
                  ? 'bg-emerald-500/10 border-emerald-500/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-emerald-500/30'
              }`}
              onClick={() =>
                !disabled &&
                onChange({
                  ...config,
                  use_gainers_losers: !config.use_gainers_losers,
                })
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  checked={config.use_gainers_losers}
                  onChange={(e) =>
                    !disabled &&
                    onChange({
                      ...config,
                      use_gainers_losers: e.target.checked,
                    })
                  }
                  disabled={disabled}
                  className="w-4 h-4 rounded accent-emerald-500"
                  onClick={(e) => e.stopPropagation()}
                />
                <Activity className="w-4 h-4 text-emerald-500" />
                <span className="text-sm font-medium text-nofx-text">
                  {t('useGainersLosers')}
                </span>
              </div>
              {config.use_gainers_losers && (
                <div className="flex items-center gap-2 mt-2 pl-6">
                  <div className="flex flex-col gap-1 w-full">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1">
                        <ArrowUpRight className="w-3 h-3 text-nofx-success" />
                        <span className="text-xs text-nofx-text-muted">
                          {t('gainersTop')}:
                        </span>
                      </div>
                      <select
                        value={config.gainers_top || 4}
                        onChange={(e) => {
                          e.stopPropagation()
                          !disabled &&
                            onChange({
                              ...config,
                              gainers_top: parseInt(e.target.value) || 4,
                            })
                        }}
                        disabled={disabled}
                        className="px-2 py-0.5 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                        onClick={(e) => e.stopPropagation()}
                      >
                        {[2, 4, 6, 8, 10].map((n) => (
                          <option key={n} value={n}>
                            {n}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1">
                        <ArrowDownRight className="w-3 h-3 text-nofx-danger" />
                        <span className="text-xs text-nofx-text-muted">
                          {t('losersTop')}:
                        </span>
                      </div>
                      <select
                        value={config.losers_top || 4}
                        onChange={(e) => {
                          e.stopPropagation()
                          !disabled &&
                            onChange({
                              ...config,
                              losers_top: parseInt(e.target.value) || 4,
                            })
                        }}
                        disabled={disabled}
                        className="px-2 py-0.5 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                        onClick={(e) => e.stopPropagation()}
                      >
                        {[2, 4, 6, 8, 10].map((n) => (
                          <option key={n} value={n}>
                            {n}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Screener Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                config.use_screener
                  ? 'bg-purple-500/10 border-purple-500/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-purple-500/30'
              }`}
              onClick={() =>
                !disabled &&
                onChange({ ...config, use_screener: !config.use_screener })
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  checked={config.use_screener}
                  onChange={(e) =>
                    !disabled &&
                    onChange({ ...config, use_screener: e.target.checked })
                  }
                  disabled={disabled}
                  className="w-4 h-4 rounded accent-purple-500"
                  onClick={(e) => e.stopPropagation()}
                />
                <Scan className="w-4 h-4 text-purple-500" />
                <span className="text-sm font-medium text-nofx-text">
                  {t('screener')}
                </span>
                <NofxOSBadge />
              </div>
              {config.use_screener && (
                <div className="pl-6 space-y-1 mt-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-nofx-text-muted">Lim:</span>
                    <select
                      value={config.screener_limit || 10}
                      onChange={(e) => {
                        e.stopPropagation()
                        !disabled &&
                          onChange({
                            ...config,
                            screener_limit: parseInt(e.target.value) || 10,
                          })
                      }}
                      disabled={disabled}
                      className="px-1 py-0.5 rounded text-[10px] bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {[5, 10, 15, 20].map((n) => (
                        <option key={n} value={n}>
                          {n}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-nofx-text-muted">Dur:</span>
                    <select
                      value={config.screener_duration || '1h'}
                      onChange={(e) => {
                        e.stopPropagation()
                        !disabled &&
                          onChange({
                            ...config,
                            screener_duration: e.target.value,
                          })
                      }}
                      disabled={disabled}
                      className="px-1 py-0.5 rounded text-[10px] bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {['15m', '1h', '4h'].map((d) => (
                        <option key={d} value={d}>
                          {d}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-nofx-text-muted">Sort:</span>
                    <select
                      value={config.screener_sort_by || 'oi'}
                      onChange={(e) => {
                        e.stopPropagation()
                        !disabled &&
                          onChange({
                            ...config,
                            screener_sort_by: e.target.value,
                          })
                      }}
                      disabled={disabled}
                      className="px-1 py-0.5 rounded text-[10px] bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <option value="oi">{t('oi')}</option>
                      <option value="price">{t('price')}</option>
                      <option value="vol">{t('vol')}</option>
                    </select>
                  </div>
                </div>
              )}
            </div>

            {/* Static/Custom Card */}
            <div
              className={`p-3 rounded-lg border transition-all cursor-pointer ${
                (config.static_coins || []).length > 0
                  ? 'bg-gray-500/10 border-gray-500/50'
                  : 'bg-nofx-bg border-nofx-border hover:border-gray-500/30'
              }`}
            >
              <div className="flex items-center gap-2 mb-2">
                <List className="w-4 h-4 text-gray-400" />
                <span className="text-sm font-medium text-nofx-text">
                  {language === 'zh' ? '自定义' : 'Custom'}
                </span>
                {(config.static_coins || []).length > 0 && (
                  <span className="text-xs px-1.5 py-0.5 rounded bg-gray-500/20 text-gray-400">
                    {config.static_coins?.length}
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-1 mt-2">
                {(config.static_coins || []).slice(0, 3).map((coin) => (
                  <span
                    key={coin}
                    className="flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-nofx-bg-lighter text-nofx-text"
                  >
                    {coin}
                    {!disabled && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleRemoveCoin(coin)
                        }}
                        className="hover:text-red-400 transition-colors"
                      >
                        <X className="w-2.5 h-2.5" />
                      </button>
                    )}
                  </span>
                ))}
                {(config.static_coins || []).length > 3 && (
                  <span className="text-xs text-nofx-text-muted">
                    +{(config.static_coins?.length || 0) - 3}
                  </span>
                )}
              </div>
              {!disabled && (
                <div className="flex gap-1 mt-2">
                  <input
                    type="text"
                    value={newCoin}
                    onChange={(e) => setNewCoin(e.target.value)}
                    onKeyDown={(e) => {
                      e.stopPropagation()
                      if (e.key === 'Enter') handleAddCoin()
                    }}
                    onClick={(e) => e.stopPropagation()}
                    placeholder="BTC, ETH..."
                    className="flex-1 px-2 py-1 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                  />
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleAddCoin()
                    }}
                    className="px-2 py-1 rounded text-xs bg-nofx-gold text-black hover:bg-yellow-500"
                  >
                    <Plus className="w-3 h-3" />
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Summary */}
          {(() => {
            const { sources, totalLimit } = getMixedSummary()
            if (sources.length === 0) return null
            return (
              <div className="p-2 rounded bg-nofx-bg border border-nofx-border">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-nofx-text-muted">
                    {t('mixedSummary')}:
                  </span>
                  <span className="text-nofx-text font-medium">
                    {sources.join(' + ')}
                  </span>
                </div>
                <div className="text-xs text-nofx-text-muted mt-1">
                  {t('maxCoins')} {totalLimit} {t('coins')}
                </div>
              </div>
            )
          })()}
        </div>
      )}
    </div>
  )
}
