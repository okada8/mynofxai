import React, { useState, useEffect, useRef } from 'react'
import { Play, Square, Settings, Activity, TrendingUp, AlertTriangle } from 'lucide-react'
import { api } from '../../lib/api'
import { toast } from 'sonner'
import { Strategy } from '../../types'
import { EvolutionDashboard } from '../EvolutionDashboard'

interface OptimizationPanelProps {
  strategyId?: string
}

export function OptimizationPanel({ strategyId }: OptimizationPanelProps) {
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [selectedStrategyId, setSelectedStrategyId] = useState<string>(strategyId || '')
  const [isRunning, setIsRunning] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  const [status, setStatus] = useState<any>(null)
  const [config, setConfig] = useState({
    generations: 50,
    population_size: 100,
    mutation_rate: 0.1,
    crossover_rate: 0.7,
    training_period_days: 30,
  })

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    loadStrategies()
    // Check for existing task in local storage
    const savedTaskId = localStorage.getItem('optimization_task_id')
    if (savedTaskId) {
      setTaskId(savedTaskId)
      setIsRunning(true)
      startPolling(savedTaskId)
    }
    return () => stopPolling()
  }, [])

  const loadStrategies = async () => {
    try {
      // Assuming getStrategies endpoint exists or we fetch traders to get strategies
      // For now, let's use a placeholder or try to fetch strategies if the endpoint exists
      // If not, we might need to fetch traders and extract strategies
      const strats = await api.getStrategies().catch(() => [])
      setStrategies(strats)
      if (!selectedStrategyId && strats.length > 0) {
        setSelectedStrategyId(strats[0].id)
      }
    } catch (err) {
      console.error('Failed to load strategies', err)
    }
  }

  const startPolling = (id: string) => {
    stopPolling()
    pollIntervalRef.current = setInterval(async () => {
      try {
        const data = await api.getOptimizationStatus(id)
        setStatus(data)
        if (data.status === 'completed' || data.status === 'failed' || data.status === 'stopped') {
          setIsRunning(false)
          stopPolling()
          if (data.status === 'completed') {
            toast.success('Optimization completed!')
          } else if (data.status === 'failed') {
            toast.error(`Optimization failed: ${data.error}`)
          }
        }
      } catch (err: any) {
        console.error('Poll error:', err)
        // Handle 404 - Task not found (server restarted or task expired)
        if (err.message && err.message.includes('404')) {
          stopPolling()
          setIsRunning(false)
          setTaskId(null)
          setStatus(null)
          localStorage.removeItem('optimization_task_id')
          toast.error('Optimization task not found. It may have been stopped or expired.')
        }
      }
    }, 1000)
  }

  const stopPolling = () => {
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current)
      pollIntervalRef.current = null
    }
  }

  const handleStart = async () => {
    if (!selectedStrategyId) {
      toast.error('Please select a strategy first')
      return
    }

    try {
      setIsRunning(true)
      const res = await api.startOptimization({
        strategy_id: selectedStrategyId,
        config: config
      })
      setTaskId(res.task_id)
      localStorage.setItem('optimization_task_id', res.task_id)
      startPolling(res.task_id)
      toast.success('Optimization started')
    } catch (err: any) {
      setIsRunning(false)
      toast.error(err.message || 'Failed to start optimization')
    }
  }

  const handleStop = async () => {
    if (!taskId) return
    try {
      await api.stopOptimization(taskId)
      setIsRunning(false)
      stopPolling()
      toast.success('Optimization stopped')
    } catch (err: any) {
      toast.error(err.message || 'Failed to stop optimization')
    }
  }

  return (
    <div className="space-y-6 p-6 bg-[#0B0E11] text-[#EAECEF]">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold flex items-center gap-2">
          <Activity className="w-6 h-6 text-[#F0B90B]" />
          Strategy Optimization (GA)
        </h2>
        <div className="flex gap-2">
          {!isRunning ? (
            <button
              onClick={handleStart}
              className="px-4 py-2 bg-[#0ECB81] hover:bg-[#0ECB81]/80 text-black font-bold rounded flex items-center gap-2"
            >
              <Play className="w-4 h-4" /> Start
            </button>
          ) : (
            <button
              onClick={handleStop}
              className="px-4 py-2 bg-[#F6465D] hover:bg-[#F6465D]/80 text-white font-bold rounded flex items-center gap-2"
            >
              <Square className="w-4 h-4" /> Stop
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Configuration Panel */}
        <div className="md:col-span-1 space-y-4 bg-[#1E2329] p-4 rounded-lg border border-[#2B3139]">
          <h3 className="font-semibold flex items-center gap-2 mb-4">
            <Settings className="w-4 h-4 text-[#848E9C]" /> Configuration
          </h3>
          
          <div className="space-y-2">
            <label className="text-sm text-[#848E9C]">Target Strategy</label>
            <select
              value={selectedStrategyId}
              onChange={(e) => setSelectedStrategyId(e.target.value)}
              disabled={isRunning}
              className="w-full bg-[#0B0E11] border border-[#2B3139] rounded p-2 text-sm text-[#EAECEF]"
            >
              <option value="">Select Strategy...</option>
              {strategies.map(s => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>

          <div className="space-y-2">
            <label className="text-sm text-[#848E9C]">Generations</label>
            <input
              type="number"
              value={config.generations}
              onChange={(e) => setConfig({ ...config, generations: parseInt(e.target.value) })}
              disabled={isRunning}
              className="w-full bg-[#0B0E11] border border-[#2B3139] rounded p-2 text-sm text-[#EAECEF]"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm text-[#848E9C]">Population Size</label>
            <input
              type="number"
              value={config.population_size}
              onChange={(e) => setConfig({ ...config, population_size: parseInt(e.target.value) })}
              disabled={isRunning}
              className="w-full bg-[#0B0E11] border border-[#2B3139] rounded p-2 text-sm text-[#EAECEF]"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm text-[#848E9C]">Mutation Rate (0-1)</label>
            <input
              type="number"
              step="0.01"
              value={config.mutation_rate}
              onChange={(e) => setConfig({ ...config, mutation_rate: parseFloat(e.target.value) })}
              disabled={isRunning}
              className="w-full bg-[#0B0E11] border border-[#2B3139] rounded p-2 text-sm text-[#EAECEF]"
            />
          </div>
        </div>

        {/* Dashboard / Status Panel */}
        <div className="md:col-span-2 bg-[#1E2329] p-4 rounded-lg border border-[#2B3139] min-h-[400px]">
          {status ? (
            <EvolutionDashboard status={status} />
          ) : (
            <div className="h-full flex flex-col items-center justify-center text-[#848E9C] opacity-50">
              <Activity className="w-16 h-16 mb-4" />
              <p>Ready to optimize strategy parameters using Genetic Algorithms</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
