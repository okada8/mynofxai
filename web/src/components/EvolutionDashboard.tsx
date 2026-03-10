import React from 'react'
import { Activity, TrendingUp, Zap, Clock } from 'lucide-react'

interface EvolutionDashboardProps {
  status: any
}

export function EvolutionDashboard({ status }: EvolutionDashboardProps) {
  if (!status) return null

  const {
    generation,
    best_fitness,
    population,
    progress,
    status: taskStatus,
  } = status
  const topCandidates = population ? population.slice(0, 5) : []

  return (
    <div className="h-full flex flex-col space-y-6 text-[#EAECEF]">
      {/* Header Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-[#0B0E11] p-4 rounded-lg border border-[#2B3139]">
          <div className="text-xs text-[#848E9C] mb-1">Generation</div>
          <div className="text-2xl font-mono text-[#F0B90B]">
            {generation || 0}
          </div>
        </div>
        <div className="bg-[#0B0E11] p-4 rounded-lg border border-[#2B3139]">
          <div className="text-xs text-[#848E9C] mb-1">Best Fitness</div>
          <div className="text-2xl font-mono text-[#0ECB81]">
            {(best_fitness || 0).toFixed(4)}
          </div>
        </div>
        <div className="bg-[#0B0E11] p-4 rounded-lg border border-[#2B3139]">
          <div className="text-xs text-[#848E9C] mb-1">Status</div>
          <div
            className={`text-2xl font-bold ${
              taskStatus === 'completed'
                ? 'text-[#0ECB81]'
                : taskStatus === 'failed'
                  ? 'text-[#F6465D]'
                  : 'text-[#3B82F6]'
            }`}
          >
            {taskStatus?.toUpperCase() || 'IDLE'}
          </div>
        </div>
        <div className="bg-[#0B0E11] p-4 rounded-lg border border-[#2B3139]">
          <div className="text-xs text-[#848E9C] mb-1">Progress</div>
          <div className="text-2xl font-mono text-[#EAECEF]">
            {(progress || 0).toFixed(1)}%
          </div>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="w-full bg-[#2B3139] rounded-full h-2.5 overflow-hidden">
        <div
          className="bg-[#F0B90B] h-2.5 rounded-full transition-all duration-500 ease-out"
          style={{ width: `${progress || 0}%` }}
        ></div>
      </div>

      {/* Top Candidates List */}
      <div className="flex-1 bg-[#0B0E11] rounded-lg border border-[#2B3139] overflow-hidden flex flex-col">
        <div className="p-4 border-b border-[#2B3139] bg-[#1E2329] flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-[#F0B90B]" />
          <h3 className="font-semibold text-sm">
            Top Candidates (Generation {generation})
          </h3>
        </div>

        <div className="overflow-y-auto flex-1 p-2">
          <table className="w-full text-sm text-left text-[#848E9C]">
            <thead className="text-xs text-[#5E6673] uppercase bg-[#1E2329]">
              <tr>
                <th className="px-4 py-3">Rank</th>
                <th className="px-4 py-3">Fitness</th>
                <th className="px-4 py-3">Genes (Params)</th>
              </tr>
            </thead>
            <tbody>
              {topCandidates.map((candidate: any, index: number) => (
                <tr
                  key={index}
                  className="border-b border-[#2B3139] hover:bg-[#1E2329]/50"
                >
                  <td className="px-4 py-3 font-mono text-[#EAECEF]">
                    #{index + 1}
                  </td>
                  <td className="px-4 py-3 font-mono text-[#0ECB81]">
                    {candidate.fitness?.toFixed(4) || 'N/A'}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-[#848E9C] truncate max-w-[200px]">
                    {JSON.stringify(candidate.genes)}
                  </td>
                </tr>
              ))}
              {topCandidates.length === 0 && (
                <tr>
                  <td
                    colSpan={3}
                    className="px-4 py-8 text-center text-[#5E6673]"
                  >
                    Waiting for first generation results...
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
