import { useState, useEffect } from 'react'
import { 
  Play, 
  Settings, 
  Activity, 
  BarChart3, 
  Check, 
  Loader2, 
  ChevronDown,
  ChevronRight,
  Calendar,
  DollarSign,
  Save,
  History
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from 'recharts'
import type { StrategyConfig, GeneDef, GAConfig, OptimizationBacktestConfig, OptimizationStatus, Strategy } from '../../types'
import { api } from '../../lib/api'
import { notify } from '../../lib/notify'

interface StrategyOptimizerProps {
  strategyId: string
  config: StrategyConfig
  onApplyParams: (params: Record<string, number>) => void
  onSaveAsNewStrategy?: (params: Record<string, number>, baseStrategyId: string) => void
  language: 'zh' | 'en'
}

export function StrategyOptimizer({ strategyId, config, onApplyParams, onSaveAsNewStrategy, language }: StrategyOptimizerProps) {
  const [genes, setGenes] = useState<GeneDef[]>([])
  const [gaConfig, setGaConfig] = useState<GAConfig>({
    population_size: 20,
    generations: 10,
    mutation_rate: 0.1,
    elite_size: 2,
    tournament_size: 3
  })
  const [backtestConfig, setBacktestConfig] = useState<OptimizationBacktestConfig>({
    symbols: ['BTCUSDT'],
    timeframes: ['15m'],
    start_time: Date.now() - 30 * 24 * 60 * 60 * 1000, // Last 30 days
    end_time: Date.now(),
    initial_balance: 10000
  })
  
  // Strategy selection state
  const [allStrategies, setAllStrategies] = useState<Strategy[]>([])
  const [currentStrategyId, setCurrentStrategyId] = useState(strategyId)

  // Backtest UI state
  const [selectedDuration, setSelectedDuration] = useState('30d')
  
  const [target, setTarget] = useState('profit')
  const [status, setStatus] = useState<OptimizationStatus | null>(null)
  const [isOptimizing, setIsOptimizing] = useState(false)
  const [expandedSections, setExpandedSections] = useState({
    params: true,
    settings: false,
    backtest: false,
    history: false
  })

  // Load all strategies on mount
  useEffect(() => {
    api.getStrategies().then(strategies => {
      setAllStrategies(strategies || [])
    }).catch(console.error)
  }, [])

  // Sync currentStrategyId if prop changes
  useEffect(() => {
    setCurrentStrategyId(strategyId)
  }, [strategyId])

  // Initialize genes helper function
  const initGenes = (cfg: StrategyConfig) => {
    const newGenes: GeneDef[] = []
    
    // RSI
    if (cfg.indicators.enable_rsi) {
      newGenes.push({
        name: 'RSI Period',
        path: 'indicators.rsi_periods.0',
        type: 0, // Int
        min: 7,
        max: 21,
        step: 1,
        enabled: true
      })
    }

    // EMA
    if (cfg.indicators.enable_ema) {
      newGenes.push({
        name: 'EMA Period',
        path: 'indicators.ema_periods.0',
        type: 0, // Int
        min: 5,
        max: 50,
        step: 1,
        enabled: true
      })
    }

    // Risk Control
    newGenes.push({
      name: 'Min Confidence',
      path: 'risk_control.min_confidence',
      type: 0, // Int
      min: 60,
      max: 95,
      step: 5,
      enabled: true
    })
    
    newGenes.push({
      name: 'Reward Ratio',
      path: 'risk_control.min_risk_reward_ratio',
      type: 1, // Float
      min: 1.5,
      max: 5.0,
      step: 0.1,
      enabled: true
    })

    // Macro Config (New in v3.0)
    if (cfg.macro_config) {
       // Could add macro params here if needed
    }

    // Grid Strategy Specifics
    if (cfg.strategy_type === 'grid_trading' && cfg.grid_config) {
      newGenes.push({
        name: 'Grid Count',
        path: 'grid_config.grid_count',
        type: 0, // Int
        min: 10,
        max: 100,
        step: 5,
        enabled: true
      })
      newGenes.push({
        name: 'Leverage',
        path: 'grid_config.leverage',
        type: 0, // Int
        min: 1,
        max: 20,
        step: 1,
        enabled: true
      })
    }

    setGenes(newGenes)
  }

  // Effect to re-init genes when config OR selected strategy changes
  useEffect(() => {
    if (currentStrategyId === strategyId) {
      initGenes(config)
    } else {
      const selected = allStrategies.find(s => s.id === currentStrategyId)
      if (selected) {
        initGenes(selected.config)
      }
    }
  }, [config, currentStrategyId, allStrategies, strategyId])

  const handleStart = async () => {
    const enabledGenes = genes.filter(g => g.enabled)
    if (enabledGenes.length === 0) {
      notify.error(language === 'zh' ? '请至少选择一个优化参数' : 'Please select at least one parameter to optimize')
      return
    }

    setIsOptimizing(true)
    setStatus(null) // Clear previous status
    
    // Check start_time and end_time validity
    if (isNaN(backtestConfig.start_time) || isNaN(backtestConfig.end_time)) {
       notify.error(language === 'zh' ? '无效的时间范围' : 'Invalid time range')
       return
    }

    try {
      const res = await api.startOptimization({
        strategy_id: currentStrategyId || strategyId, // Use selected strategy ID or prop fallback
        // Only send config if we are optimizing the current editing strategy
        // Otherwise, let backend fetch the strategy config from DB using strategy_id
        strategy_config: currentStrategyId === strategyId ? config : undefined,
        parameter_ranges: enabledGenes,
        optimization_target: target,
        ga_config: gaConfig,
        backtest_config: {
          ...backtestConfig,
          start_time: Math.floor(backtestConfig.start_time / 1000), // Convert to seconds for backend
          end_time: Math.floor(backtestConfig.end_time / 1000)      // Convert to seconds for backend
        }
      })
      
      // Start polling
      const pollInterval = setInterval(async () => {
        try {
          const s = await api.getOptimizationStatus(res.task_id)
          setStatus(s)
          
          if (s.status === 'completed' || s.status === 'failed' || s.status === 'cancelled') {
            clearInterval(pollInterval)
            setIsOptimizing(false)
            if (s.status === 'completed') {
              notify.success(language === 'zh' ? '优化完成' : 'Optimization completed')
            } else if (s.status === 'failed') {
              notify.error(language === 'zh' ? '优化失败: ' + s.error : 'Optimization failed: ' + s.error)
            }
          }
        } catch (e) {
          console.error(e)
        }
      }, 2000)
    } catch (e) {
      notify.error(language === 'zh' ? '启动优化失败' : 'Failed to start optimization')
      setIsOptimizing(false)
    }
  }

  const handleApplyParams = () => {
    // Check nested structure first (backend return) or flat structure (frontend helper)
    const bestParams = status?.progress?.best_genes || status?.best_individual?.parameters
    
    if (bestParams) {
      // Map genes back to parameters structure if needed, or use directly
      // The backend returns a flat map "path": value in best_genes
      onApplyParams(bestParams)
      notify.success(language === 'zh' ? '参数已应用' : 'Parameters applied')
      
      // If user selected a different strategy, we should probably warn or handle it
      // For now, we assume the parent handles applying to the *currently editing* config
    }
  }

  // --- Optimization History (Mock or Real) ---
  // In a real implementation, we would fetch this from an API
  // For now, let's just keep track of local completions or maybe backend supports it?
  // The backend DOES NOT persist history currently, only runtime memory.
  // So we will just show the "History" tab if status exists and has history.
  
  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections(prev => ({ ...prev, [section]: !prev[section] }))
  }

  const handleDurationChange = (duration: string) => {
    setSelectedDuration(duration)
    const now = Date.now()
    let days = 30
    switch(duration) {
      case '7d': days = 7; break;
      case '30d': days = 30; break;
      case '90d': days = 90; break;
      case '180d': days = 180; break;
    }
    setBacktestConfig(prev => ({
      ...prev,
      start_time: now - days * 24 * 60 * 60 * 1000,
      end_time: now
    }))
  }

  // Calculate chart data from history
  const chartData = status?.progress?.history?.map(h => ({
    gen: h.generation,
    fitness: h.best_fitness
  })) || []

  return (
    <div className="flex flex-col h-full bg-nofx-bg/50">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        
        {/* Progress & Chart Section */}
        {status && (
          <div className="p-4 rounded-lg bg-nofx-bg border border-nofx-gold/20 animate-fade-in">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-bold text-nofx-gold flex items-center gap-2">
                {status.status === 'running' && <Loader2 className="w-3 h-3 animate-spin" />}
                {status.status === 'running' ? (language === 'zh' ? '优化运行中...' : 'Optimizing...') : 
                 status.status === 'completed' ? (language === 'zh' ? '优化完成' : 'Completed') : 
                 (language === 'zh' ? '优化失败' : 'Failed')}
              </span>
              <span className="text-xs text-nofx-text-muted">
                Gen {status.progress?.generation || 0} / {gaConfig.generations}
              </span>
            </div>
            
            {/* Progress Bar */}
            <div className="w-full h-1.5 bg-nofx-bg-dark rounded-full overflow-hidden mb-4">
              <div 
                className="h-full bg-nofx-gold transition-all duration-500"
                style={{ width: `${((status.progress?.generation || 0) / gaConfig.generations) * 100}%` }}
              />
            </div>

            {/* Fitness Chart */}
            <div className="h-32 w-full mb-3">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                  <defs>
                    <linearGradient id="colorFitness" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#F0B90B" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#F0B90B" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#2B3139" />
                  <XAxis dataKey="gen" stroke="#848E9C" fontSize={10} tickLine={false} axisLine={false} />
                  <YAxis stroke="#848E9C" fontSize={10} tickLine={false} axisLine={false} domain={['auto', 'auto']} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0B0E11', border: '1px solid #2B3139', borderRadius: '4px' }}
                    itemStyle={{ color: '#F0B90B', fontSize: '12px' }}
                    labelStyle={{ color: '#848E9C', fontSize: '10px', marginBottom: '4px' }}
                  />
                  <Area type="monotone" dataKey="fitness" stroke="#F0B90B" fillOpacity={1} fill="url(#colorFitness)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>

            {status.progress?.best_fitness > 0 && (
              <div className="border-t border-nofx-gold/10 pt-2 mt-2 space-y-1">
                <div className="text-xs text-nofx-text flex justify-between items-center">
                  <span>Best Fitness ({target}):</span>
                  <span className="text-green-400 font-mono font-bold">{status.progress.best_fitness.toFixed(4)}</span>
                </div>
                {/* Always display key metrics */}
                <div className="grid grid-cols-3 gap-2 mt-2 pt-2 border-t border-nofx-gold/5">
                   <div className="flex flex-col items-center p-1.5 rounded bg-nofx-bg-lighter border border-nofx-gold/10">
                     <span className="text-[10px] text-nofx-text-muted mb-0.5">{language === 'zh' ? '总收益' : 'Profit'}</span>
                     <span className={`text-xs font-bold font-mono ${status.best_individual?.metrics?.total_return_pct && status.best_individual.metrics.total_return_pct >= 0 ? 'text-nofx-success' : 'text-nofx-danger'}`}>
                       {status.best_individual?.metrics?.total_return_pct ? `${status.best_individual.metrics.total_return_pct.toFixed(2)}%` : '-'}
                     </span>
                   </div>
                   <div className="flex flex-col items-center p-1.5 rounded bg-nofx-bg-lighter border border-nofx-gold/10">
                     <span className="text-[10px] text-nofx-text-muted mb-0.5">{language === 'zh' ? '夏普' : 'Sharpe'}</span>
                     <span className="text-xs font-bold font-mono text-nofx-gold">
                       {status.best_individual?.metrics?.sharpe_ratio ? status.best_individual.metrics.sharpe_ratio.toFixed(2) : '-'}
                     </span>
                   </div>
                   <div className="flex flex-col items-center p-1.5 rounded bg-nofx-bg-lighter border border-nofx-gold/10">
                     <span className="text-[10px] text-nofx-text-muted mb-0.5">{language === 'zh' ? '回撤' : 'MaxDD'}</span>
                     <span className="text-xs font-bold font-mono text-nofx-danger">
                       {status.best_individual?.metrics?.max_drawdown_pct ? `${status.best_individual.metrics.max_drawdown_pct.toFixed(2)}%` : '-'}
                     </span>
                   </div>
                </div>
              </div>
            )}

            {/* Completion Actions */}
            {status.status === 'completed' && status.progress?.best_genes && (
              <div className="mt-4 p-3 rounded bg-green-500/10 border border-green-500/20">
                <div className="text-xs font-bold text-green-500 mb-2">
                  {language === 'zh' ? '找到最佳参数:' : 'Best Parameters Found:'}
                </div>
                <div className="grid grid-cols-2 gap-2 mb-3">
                  {Object.entries(status.progress.best_genes).map(([key, val]) => (
                    <div key={key} className="text-xs flex justify-between">
                      <span className="text-nofx-text-muted truncate mr-2" title={key}>{key.split('.').pop()}:</span>
                      <span className="text-nofx-text font-mono">{Number(val).toFixed(4)}</span>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={handleApplyParams}
                    className="flex-1 py-2 rounded text-xs font-bold bg-green-500 hover:bg-green-600 text-white transition-colors flex items-center justify-center gap-1"
                  >
                    <Check className="w-3 h-3" />
                    {language === 'zh' ? '应用' : 'Apply'}
                  </button>
                  {onSaveAsNewStrategy && (
                    <button
                      onClick={() => {
                        const bestParams = status?.progress?.best_genes || status?.best_individual?.parameters
                        if (bestParams) {
                          onSaveAsNewStrategy(bestParams, currentStrategyId)
                        }
                      }}
                      className="flex-1 py-2 rounded text-xs font-bold bg-nofx-gold text-black hover:bg-yellow-500 transition-colors flex items-center justify-center gap-1"
                    >
                      <Save className="w-3 h-3" />
                      {language === 'zh' ? '另存为新策略' : 'Save as New'}
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        {/* 1. Parameters Selection */}
        <div className="rounded-lg border border-nofx-gold/20 overflow-hidden mb-4">
          {/* Strategy Selection */}
          <div className="p-3 bg-nofx-bg border-b border-nofx-gold/20">
             <label className="text-xs text-nofx-text-muted block mb-1">
               {language === 'zh' ? '基准策略' : 'Base Strategy'}
             </label>
             <select 
               value={currentStrategyId}
               onChange={e => setCurrentStrategyId(e.target.value)}
               className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
             >
               {allStrategies.map(s => (
                 <option key={s.id} value={s.id}>{s.name}</option>
               ))}
             </select>
          </div>
          
          <button 
            onClick={() => toggleSection('params')}
            className="w-full flex items-center justify-between p-3 bg-nofx-bg-lighter hover:bg-white/5 transition-colors"
          >
            <div className="flex items-center gap-2">
              <Settings className="w-4 h-4 text-nofx-gold" />
              <span className="text-sm font-medium text-nofx-text">
                {language === 'zh' ? '优化参数' : 'Parameters'}
              </span>
            </div>
            {expandedSections.params ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
          
          {expandedSections.params && (
            <div className="p-3 bg-nofx-bg space-y-3">
              {genes.map((gene, idx) => (
                <div key={idx} className="flex items-center gap-2 text-xs">
                  <input 
                    type="checkbox" 
                    checked={gene.enabled}
                    onChange={e => {
                      const newGenes = [...genes]
                      newGenes[idx].enabled = e.target.checked
                      setGenes(newGenes)
                    }}
                    className="accent-nofx-gold"
                  />
                  <span className="flex-1 text-nofx-text truncate" title={gene.path}>{gene.name}</span>
                  <input 
                    type="number" 
                    value={gene.min}
                    onChange={e => {
                      const newGenes = [...genes]
                      newGenes[idx].min = parseFloat(e.target.value)
                      setGenes(newGenes)
                    }}
                    className="w-16 px-1 py-0.5 bg-black/30 border border-nofx-gold/10 rounded text-center text-nofx-text"
                    placeholder="Min"
                  />
                  <span className="text-nofx-text-muted">-</span>
                  <input 
                    type="number" 
                    value={gene.max}
                    onChange={e => {
                      const newGenes = [...genes]
                      newGenes[idx].max = parseFloat(e.target.value)
                      setGenes(newGenes)
                    }}
                    className="w-16 px-1 py-0.5 bg-black/30 border border-nofx-gold/10 rounded text-center text-nofx-text"
                    placeholder="Max"
                  />
                </div>
              ))}
              {genes.length === 0 && (
                <div className="text-center py-4 text-xs text-nofx-text-muted">
                  {language === 'zh' ? '没有可优化的参数' : 'No optimizable parameters found'}
                </div>
              )}
            </div>
          )}
        </div>

        {/* 2. GA Settings */}
        <div className="rounded-lg border border-nofx-gold/20 overflow-hidden">
          <button 
            onClick={() => toggleSection('settings')}
            className="w-full flex items-center justify-between p-3 bg-nofx-bg-lighter hover:bg-white/5 transition-colors"
          >
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-blue-400" />
              <span className="text-sm font-medium text-nofx-text">
                {language === 'zh' ? '算法设置' : 'Algorithm Settings'}
              </span>
            </div>
            {expandedSections.settings ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
          
          {expandedSections.settings && (
            <div className="p-3 bg-nofx-bg space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-nofx-text-muted block mb-1">
                    {language === 'zh' ? '目标' : 'Target'}
                  </label>
                  <select 
                    value={target}
                    onChange={e => setTarget(e.target.value)}
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  >
                    <option value="profit">{language === 'zh' ? '总收益' : 'Total Profit'}</option>
                    <option value="sharpe">{language === 'zh' ? '夏普比率' : 'Sharpe Ratio'}</option>
                    <option value="drawdown">{language === 'zh' ? '最小回撤' : 'Min Drawdown'}</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs text-nofx-text-muted block mb-1">
                    {language === 'zh' ? '种群' : 'Population'}
                  </label>
                  <input 
                    type="number"
                    value={gaConfig.population_size}
                    onChange={e => setGaConfig({...gaConfig, population_size: parseInt(e.target.value)})}
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  />
                </div>
                <div>
                  <label className="text-xs text-nofx-text-muted block mb-1">
                    {language === 'zh' ? '代数' : 'Generations'}
                  </label>
                  <input 
                    type="number"
                    value={gaConfig.generations}
                    onChange={e => setGaConfig({...gaConfig, generations: parseInt(e.target.value)})}
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  />
                </div>
                <div>
                  <label className="text-xs text-nofx-text-muted block mb-1">
                    {language === 'zh' ? '变异率' : 'Mutation Rate'}
                  </label>
                  <input 
                    type="number"
                    step="0.01"
                    value={gaConfig.mutation_rate}
                    onChange={e => setGaConfig({...gaConfig, mutation_rate: parseFloat(e.target.value)})}
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  />
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 3. Backtest Configuration */}
        <div className="rounded-lg border border-nofx-gold/20 overflow-hidden">
          <button 
            onClick={() => toggleSection('backtest')}
            className="w-full flex items-center justify-between p-3 bg-nofx-bg-lighter hover:bg-white/5 transition-colors"
          >
            <div className="flex items-center gap-2">
              <BarChart3 className="w-4 h-4 text-green-400" />
              <span className="text-sm font-medium text-nofx-text">
                {language === 'zh' ? '回测环境' : 'Backtest Environment'}
              </span>
            </div>
            {expandedSections.backtest ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
          
          {expandedSections.backtest && (
            <div className="p-3 bg-nofx-bg space-y-3">
              <div>
                <label className="text-xs text-nofx-text-muted block mb-1">
                  {language === 'zh' ? '交易对' : 'Symbol'}
                </label>
                <input 
                  type="text"
                  value={backtestConfig.symbols[0]}
                  onChange={e => setBacktestConfig({...backtestConfig, symbols: [e.target.value.toUpperCase()]})}
                  className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  placeholder="BTCUSDT"
                />
              </div>

              <div>
                <label className="text-xs text-nofx-text-muted block mb-1">
                  {language === 'zh' ? '时间周期' : 'Timeframe'}
                </label>
                <div className="flex gap-2">
                  {['5m', '15m', '1h', '4h', '1d'].map(tf => (
                    <button
                      key={tf}
                      onClick={() => setBacktestConfig({...backtestConfig, timeframes: [tf]})}
                      className={`px-2 py-1 rounded text-xs transition-colors ${
                        backtestConfig.timeframes[0] === tf 
                          ? 'bg-nofx-gold text-black font-bold' 
                          : 'bg-black/30 text-nofx-text hover:bg-white/10'
                      }`}
                    >
                      {tf}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="text-xs text-nofx-text-muted block mb-1">
                  {language === 'zh' ? '回测时长' : 'Duration'}
                </label>
                <div className="flex gap-2">
                  {['7d', '30d', '90d', '180d'].map(d => (
                    <button
                      key={d}
                      onClick={() => handleDurationChange(d)}
                      className={`px-2 py-1 rounded text-xs transition-colors ${
                        selectedDuration === d 
                          ? 'bg-nofx-gold text-black font-bold' 
                          : 'bg-black/30 text-nofx-text hover:bg-white/10'
                      }`}
                    >
                      {d}
                    </button>
                  ))}
                </div>
                <div className="text-[10px] text-nofx-text-muted mt-1 flex items-center gap-1">
                  <Calendar className="w-3 h-3" />
                  {new Date(backtestConfig.start_time).toLocaleDateString()} - {new Date(backtestConfig.end_time).toLocaleDateString()}
                </div>
              </div>

              <div>
                <label className="text-xs text-nofx-text-muted block mb-1">
                  {language === 'zh' ? '初始资金' : 'Initial Balance'}
                </label>
                <div className="relative">
                  <DollarSign className="w-3 h-3 absolute left-2 top-1/2 -translate-y-1/2 text-nofx-text-muted" />
                  <input 
                    type="number"
                    value={backtestConfig.initial_balance}
                    onChange={e => setBacktestConfig({...backtestConfig, initial_balance: parseFloat(e.target.value)})}
                    className="w-full pl-6 pr-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text outline-none focus:border-nofx-gold/50"
                  />
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 4. Optimization History (Parameters) */}
        {status?.progress?.best_genes && (
          <div className="rounded-lg border border-nofx-gold/20 overflow-hidden">
            <button 
              onClick={() => toggleSection('history')}
              className="w-full flex items-center justify-between p-3 bg-nofx-bg-lighter hover:bg-white/5 transition-colors"
            >
              <div className="flex items-center gap-2">
                <History className="w-4 h-4 text-purple-400" />
                <span className="text-sm font-medium text-nofx-text">
                  {language === 'zh' ? '优化结果参数' : 'Optimized Parameters'}
                </span>
              </div>
              {expandedSections.history ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
            </button>
            
            {expandedSections.history && (
              <div className="p-3 bg-nofx-bg space-y-2">
                <div className="text-xs text-nofx-text-muted mb-2">
                  {language === 'zh' 
                    ? '以下是当前找到的最佳参数组合：' 
                    : 'The best parameter combination found so far:'}
                </div>
                <div className="grid grid-cols-1 gap-2">
                  {Object.entries(status.progress.best_genes).map(([key, val]) => (
                    <div key={key} className="flex justify-between items-center p-2 rounded bg-black/20 border border-nofx-gold/5">
                      <span className="text-xs text-nofx-text truncate max-w-[180px]" title={key}>
                        {key.split('.').pop()}
                      </span>
                      <div className="flex items-center gap-2">
                        {/* Show change indicator if we could compare with original */}
                        <span className="text-xs font-mono font-bold text-nofx-gold">
                          {Number(val).toFixed(4)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

      </div>

      {/* Action Footer */}
      <div className="p-4 border-t border-nofx-gold/20 bg-nofx-bg">
        <button
          onClick={handleStart}
          disabled={isOptimizing || genes.length === 0}
          className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg text-sm font-bold transition-all text-black shadow-lg shadow-nofx-gold/20 bg-gradient-to-r from-nofx-gold to-yellow-500 hover:from-yellow-400 hover:to-yellow-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isOptimizing ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              {language === 'zh' ? '优化中...' : 'Optimizing...'}
            </>
          ) : (
            <>
              <Play className="w-4 h-4" />
              {language === 'zh' ? '开始优化' : 'Start Optimization'}
            </>
          )}
        </button>
      </div>
    </div>
  )
}
