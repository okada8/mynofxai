import { useState, useEffect, useCallback } from 'react'
import { 
  Play, 
  Settings, 
  Activity, 
  BarChart3, 
  Check, 
  Loader2, 
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  RefreshCw
} from 'lucide-react'
import type { StrategyConfig, GeneDef, GAConfig, OptimizationBacktestConfig, OptimizationStatus } from '../../types'
import { api } from '../../lib/api'
import { notify } from '../../lib/notify'

interface StrategyOptimizerProps {
  strategyId: string
  config: StrategyConfig
  onApplyParams: (params: Record<string, number>) => void
  language: 'zh' | 'en'
}

export function StrategyOptimizer({ strategyId, config, onApplyParams, language }: StrategyOptimizerProps) {
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
  const [target, setTarget] = useState('profit')
  const [status, setStatus] = useState<OptimizationStatus | null>(null)
  const [isOptimizing, setIsOptimizing] = useState(false)
  const [expandedSections, setExpandedSections] = useState({
    params: true,
    settings: false,
    backtest: false
  })

  // Initialize genes based on strategy config
  useEffect(() => {
    const newGenes: GeneDef[] = []
    
    // RSI
    if (config.indicators.enable_rsi) {
      newGenes.push({
        name: 'RSI Period',
        path: 'indicators.rsi_periods.0',
        type: 'int',
        min: 7,
        max: 21,
        step: 1,
        enabled: true
      })
    }

    // EMA
    if (config.indicators.enable_ema) {
      newGenes.push({
        name: 'EMA Period',
        path: 'indicators.ema_periods.0',
        type: 'int',
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
      type: 'int',
      min: 60,
      max: 95,
      step: 5,
      enabled: true
    })
    
    newGenes.push({
      name: 'Reward Ratio',
      path: 'risk_control.min_risk_reward_ratio',
      type: 'float',
      min: 1.5,
      max: 5.0,
      step: 0.1,
      enabled: true
    })

    // Grid Strategy Specifics
    if (config.strategy_type === 'grid_trading' && config.grid_config) {
      newGenes.push({
        name: 'Grid Count',
        path: 'grid_config.grid_count',
        type: 'int',
        min: 10,
        max: 100,
        step: 5,
        enabled: true
      })
      newGenes.push({
        name: 'Leverage',
        path: 'grid_config.leverage',
        type: 'int',
        min: 1,
        max: 20,
        step: 1,
        enabled: true
      })
    }

    setGenes(newGenes)
  }, [config.strategy_type, config.indicators, config.grid_config])

  const handleStart = async () => {
    const enabledGenes = genes.filter(g => g.enabled)
    if (enabledGenes.length === 0) {
      notify.error(language === 'zh' ? '请至少选择一个优化参数' : 'Please select at least one parameter to optimize')
      return
    }

    setIsOptimizing(true)
    try {
      const res = await api.startOptimization({
        strategy_id: strategyId,
        strategy_config: config,
        parameter_ranges: enabledGenes,
        optimization_target: target,
        ga_config: gaConfig,
        backtest_config: backtestConfig
      })
      
      // Start polling
      const pollInterval = setInterval(async () => {
        try {
          const s = await api.getOptimizationStatus(res.task_id)
          setStatus(s)
          if (s.status === 'completed' || s.status === 'failed') {
            clearInterval(pollInterval)
            setIsOptimizing(false)
            if (s.status === 'completed') {
              notify.success(language === 'zh' ? '优化完成' : 'Optimization completed')
            } else {
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
    if (status?.best_individual?.parameters) {
      onApplyParams(status.best_individual.parameters)
    }
  }

  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections(prev => ({ ...prev, [section]: !prev[section] }))
  }

  return (
    <div className="flex flex-col h-full bg-nofx-bg/50">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        
        {/* Progress Section */}
        {status && (
          <div className="p-4 rounded-lg bg-nofx-bg border border-nofx-gold/20 animate-fade-in">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-bold text-nofx-gold">
                {status.status === 'running' ? (language === 'zh' ? '优化运行中...' : 'Optimizing...') : 
                 status.status === 'completed' ? (language === 'zh' ? '优化完成' : 'Completed') : 
                 (language === 'zh' ? '优化失败' : 'Failed')}
              </span>
              <span className="text-xs text-nofx-text-muted">
                Gen {status.current_generation} / {status.total_generations}
              </span>
            </div>
            <div className="w-full h-2 bg-nofx-bg-dark rounded-full overflow-hidden mb-2">
              <div 
                className="h-full bg-nofx-gold transition-all duration-500"
                style={{ width: `${status.progress_pct}%` }}
              />
            </div>
            {status.best_fitness > 0 && (
              <div className="text-xs text-nofx-text">
                Best Fitness: <span className="text-green-400">{status.best_fitness.toFixed(4)}</span>
              </div>
            )}

            {/* Completion Actions */}
            {status.status === 'completed' && status.best_individual && (
              <div className="mt-4 p-3 rounded bg-green-500/10 border border-green-500/20">
                <div className="text-xs font-bold text-green-500 mb-2">
                  {language === 'zh' ? '找到最佳参数:' : 'Best Parameters Found:'}
                </div>
                <div className="grid grid-cols-2 gap-2 mb-3">
                  {Object.entries(status.best_individual.parameters).map(([key, val]) => (
                    <div key={key} className="text-xs flex justify-between">
                      <span className="text-nofx-text-muted">{key.split('.').pop()}:</span>
                      <span className="text-nofx-text font-mono">{val}</span>
                    </div>
                  ))}
                </div>
                <button
                  onClick={handleApplyParams}
                  className="w-full py-2 rounded text-xs font-bold bg-green-500 hover:bg-green-600 text-white transition-colors flex items-center justify-center gap-1"
                >
                  <Check className="w-3 h-3" />
                  {language === 'zh' ? '应用到当前策略' : 'Apply to Current Strategy'}
                </button>
              </div>
            )}
          </div>
        )}

        {/* 1. Parameters Selection */}
        <div className="rounded-lg border border-nofx-gold/20 overflow-hidden">
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
                    className="w-16 px-1 py-0.5 bg-black/30 border border-nofx-gold/10 rounded text-center"
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
                    className="w-16 px-1 py-0.5 bg-black/30 border border-nofx-gold/10 rounded text-center"
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
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text"
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
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text"
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
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text"
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
                    className="w-full px-2 py-1.5 bg-black/30 border border-nofx-gold/10 rounded text-xs text-nofx-text"
                  />
                </div>
              </div>
            </div>
          )}
        </div>

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
