import React from 'react'
import { OptimizationPanel } from '../components/strategy/OptimizationPanel'
import { DeepVoidBackground } from '../components/DeepVoidBackground'

export function OptimizationPage() {
  return (
    <DeepVoidBackground className="min-h-screen pt-20 px-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold text-[#EAECEF] mb-6 font-mono">
          STRATEGY_OPTIMIZATION_LAB
        </h1>
        <OptimizationPanel />
      </div>
    </DeepVoidBackground>
  )
}
