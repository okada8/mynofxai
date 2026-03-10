import React from 'react'
import { RiskState } from '../../types'
import {
  AlertTriangle,
  CheckCircle,
  ShieldAlert,
  TrendingDown,
  Percent,
  Activity,
} from 'lucide-react'
import { Language } from '../../i18n/translations'

interface RiskMonitorPanelProps {
  riskState: RiskState | undefined
  totalEquity: number
  language: Language
}

export const RiskMonitorPanel = React.memo(function RiskMonitorPanel({
  riskState,
  totalEquity,
  language,
}: RiskMonitorPanelProps) {
  if (!riskState) return null

  const {
    level,
    var: valueAtRisk,
    var_pct,
    max_drawdown,
    exposure,
    utilization,
    message,
  } = riskState

  // Calculate leverage ratio
  const leverageRatio = totalEquity > 0 ? exposure / totalEquity : 0

  // Determine color based on risk level
  const getLevelColor = (level: string) => {
    switch (level?.toLowerCase()) {
      case 'safe':
        return 'text-nofx-green border-nofx-green/30 bg-nofx-green/10'
      case 'caution':
        return 'text-nofx-gold border-nofx-gold/30 bg-nofx-gold/10'
      case 'danger':
        return 'text-nofx-red border-nofx-red/30 bg-nofx-red/10'
      default:
        return 'text-gray-400 border-gray-400/30 bg-gray-400/10'
    }
  }

  const getLevelIcon = (level: string) => {
    switch (level?.toLowerCase()) {
      case 'safe':
        return <CheckCircle className="w-5 h-5" />
      case 'caution':
        return <AlertTriangle className="w-5 h-5" />
      case 'danger':
        return <ShieldAlert className="w-5 h-5" />
      default:
        return <Activity className="w-5 h-5" />
    }
  }

  const levelColorClass = getLevelColor(level)
  const levelIcon = getLevelIcon(level)

  // Helper for labels
  const t = (key: string) => {
    const en: Record<string, string> = {
      title: 'Risk Monitor',
      var: 'Value at Risk (VaR)',
      maxDrawdown: 'Max Drawdown',
      exposure: 'Leverage',
      utilization: 'Margin Usage',
      status: 'Status',
      safe: 'SAFE',
      caution: 'CAUTION',
      danger: 'DANGER',
      riskMsg: 'RISK_MSG',
    }
    const zh: Record<string, string> = {
      title: '风控监控',
      var: '风险价值 (VaR)',
      maxDrawdown: '最大回撤',
      exposure: '杠杆倍数',
      utilization: '保证金占用',
      status: '状态',
      safe: '安全',
      caution: '警告',
      danger: '危险',
      riskMsg: '风险提示',
    }
    return language === 'zh' ? zh[key] || key : en[key] || key
  }

  // Format helpers
  const formatPct = (val: number) => `${val.toFixed(2)}%`
  const formatCurrency = (val: number) =>
    `$${val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`

  return (
    <div className="nofx-glass p-6 rounded-lg border border-white/5 animate-slide-in relative overflow-hidden group">
      {/* Background Glow */}
      <div
        className={`absolute top-0 right-0 p-20 opacity-5 blur-3xl rounded-full transition-colors duration-500 ${level?.toLowerCase() === 'safe' ? 'bg-nofx-green' : level?.toLowerCase() === 'caution' ? 'bg-nofx-gold' : 'bg-nofx-red'}`}
      />

      <div className="flex items-center justify-between mb-6 relative z-10">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg ${levelColorClass}`}>
            <ShieldAlert className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-nofx-text-main">
              {t('title')}
            </h2>
            <div className="flex items-center gap-2 text-xs font-mono mt-1">
              <span className="text-nofx-text-muted">
                SYSTEM_RISK_CHECK::ACTIVE
              </span>
              <span
                className={`w-1.5 h-1.5 rounded-full animate-pulse ${level?.toLowerCase() === 'safe' ? 'bg-nofx-green' : level?.toLowerCase() === 'caution' ? 'bg-nofx-gold' : 'bg-nofx-red'}`}
              />
            </div>
          </div>
        </div>

        <div
          className={`flex items-center gap-2 px-4 py-2 rounded-lg border ${levelColorClass}`}
        >
          {levelIcon}
          <span className="font-bold uppercase tracking-wider">
            {t(level?.toLowerCase() || 'unknown')}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 relative z-10">
        {/* VaR Metric */}
        <RiskMetric
          label={t('var')}
          value={formatPct(var_pct)}
          subValue={formatCurrency(valueAtRisk)}
          icon={<Activity className="w-4 h-4" />}
          color={
            var_pct > 5
              ? 'text-nofx-red'
              : var_pct > 2
                ? 'text-nofx-gold'
                : 'text-nofx-green'
          }
        />

        {/* Max Drawdown */}
        <RiskMetric
          label={t('maxDrawdown')}
          value={formatPct(Math.abs(max_drawdown))}
          icon={<TrendingDown className="w-4 h-4" />}
          color={
            Math.abs(max_drawdown) > 20
              ? 'text-nofx-red'
              : Math.abs(max_drawdown) > 10
                ? 'text-nofx-gold'
                : 'text-nofx-text-main'
          }
        />

        {/* Leverage */}
        <RiskMetric
          label={t('exposure')}
          value={`x${leverageRatio.toFixed(2)}`}
          subValue={formatCurrency(exposure)}
          icon={<Percent className="w-4 h-4" />}
          color={
            leverageRatio > 3
              ? 'text-nofx-red'
              : leverageRatio > 1.5
                ? 'text-nofx-gold'
                : 'text-nofx-text-main'
          }
        />

        {/* Margin Utilization */}
        <RiskMetric
          label={t('utilization')}
          value={formatPct(utilization)}
          icon={<Percent className="w-4 h-4" />}
          color={
            utilization > 80
              ? 'text-nofx-red'
              : utilization > 50
                ? 'text-nofx-gold'
                : 'text-nofx-green'
          }
        />
      </div>

      {/* Message Area */}
      {message && message !== 'Portfolio risk is within safe limits.' && (
        <div className="mt-6 p-4 rounded bg-black/20 border border-white/5 font-mono text-sm text-nofx-text-muted flex items-start gap-3 relative z-10">
          <div className="mt-0.5 text-nofx-gold">⚠️</div>
          <div>
            <span className="opacity-50 mr-2">{t('riskMsg')}:</span>
            <span className="text-nofx-text-main">{message}</span>
          </div>
        </div>
      )}
    </div>
  )
})

function RiskMetric({
  label,
  value,
  subValue,
  icon,
  color,
}: {
  label: string
  value: string
  subValue?: string
  icon: React.ReactNode
  color: string
}) {
  return (
    <div className="bg-black/20 rounded-lg p-4 border border-white/5 hover:border-white/10 transition-colors">
      <div className="flex items-center gap-2 text-nofx-text-muted text-xs uppercase tracking-wider mb-2 opacity-70">
        {icon}
        <span>{label}</span>
      </div>
      <div className={`text-2xl font-bold font-mono ${color}`}>{value}</div>
      {subValue && (
        <div className="text-xs text-nofx-text-muted mt-1 font-mono opacity-50">
          {subValue}
        </div>
      )}
    </div>
  )
}
