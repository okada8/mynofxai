import { Bot, Users } from 'lucide-react'
import type { MultiAgentConfig, AIModel } from '../../types'

interface MultiAgentEditorProps {
  config?: MultiAgentConfig
  aiModels: AIModel[]
  onChange: (config: MultiAgentConfig) => void
  disabled?: boolean
  language: string
}

export function MultiAgentEditor({
  config,
  aiModels,
  onChange,
  disabled,
  language,
}: MultiAgentEditorProps) {
  // Default values
  const safeConfig: MultiAgentConfig = config || {
    enabled: false,
    mode: 'debate',
    agents: [
      { role: 'risk_manager', model_id: '', enabled: true },
      { role: 'alpha_hunter', model_id: '', enabled: true },
      { role: 'analyst', model_id: '', enabled: true },
      { role: 'executor', model_id: '', enabled: true },
    ],
  }

  const t = (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      multiAgentSystem: { zh: '多代理系统', en: 'Multi-Agent System' },
      multiAgentDesc: { zh: '配置多个 AI Agent 协同工作，分工处理不同任务', en: 'Configure multiple AI Agents to collaborate on different tasks' },
      enableSystem: { zh: '启用多代理系统', en: 'Enable Multi-Agent System' },
      collaborationMode: { zh: '协作模式', en: 'Collaboration Mode' },
      debate: { zh: '辩论模式 (Debate)', en: 'Debate Mode' },
      consensus: { zh: '共识投票 (Consensus)', en: 'Consensus Voting' },
      hierarchical: { zh: '层级管理 (Hierarchical)', en: 'Hierarchical Management' },
      agentRoles: { zh: '代理角色配置', en: 'Agent Roles Configuration' },
      role: { zh: '角色', en: 'Role' },
      model: { zh: '模型', en: 'Model' },
      status: { zh: '状态', en: 'Status' },
      risk_manager: { zh: '风控官 (Risk Manager)', en: 'Risk Manager' },
      alpha_hunter: { zh: 'Alpha 猎手 (Alpha Hunter)', en: 'Alpha Hunter' },
      analyst: { zh: '分析师 (Analyst)', en: 'Analyst' },
      executor: { zh: '执行官 (Executor)', en: 'Executor' },
      selectModel: { zh: '选择模型', en: 'Select Model' },
    }
    return translations[key]?.[language] || key
  }

  const updateField = <K extends keyof MultiAgentConfig>(
    key: K,
    value: MultiAgentConfig[K]
  ) => {
    if (!disabled) {
      onChange({ ...safeConfig, [key]: value })
    }
  }

  const updateAgent = (index: number, field: string, value: any) => {
    if (disabled) return
    const newAgents = [...safeConfig.agents]
    newAgents[index] = { ...newAgents[index], [field]: value }
    onChange({ ...safeConfig, agents: newAgents })
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Users className="w-5 h-5" style={{ color: '#a855f7' }} />
          <h3 className="font-medium" style={{ color: '#EAECEF' }}>
            {t('multiAgentSystem')}
          </h3>
        </div>
        <p className="text-xs mb-4" style={{ color: '#848E9C' }}>
          {t('multiAgentDesc')}
        </p>

        {/* Enable Toggle */}
        <div className="mb-4 p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
          <label className="flex items-center gap-2 text-sm font-medium cursor-pointer" style={{ color: '#EAECEF' }}>
            <input
              type="checkbox"
              checked={safeConfig.enabled}
              onChange={(e) => updateField('enabled', e.target.checked)}
              disabled={disabled}
              className="w-4 h-4 rounded border-gray-600 bg-transparent text-purple-500 focus:ring-offset-0 focus:ring-0"
            />
            {t('enableSystem')}
          </label>
        </div>

        {safeConfig.enabled && (
          <div className="space-y-4">
            {/* Mode Selection */}
            <div className="p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
              <label className="block text-xs mb-2" style={{ color: '#848E9C' }}>
                {t('collaborationMode')}
              </label>
              <div className="flex gap-2">
                {(['debate', 'consensus', 'hierarchical'] as const).map((mode) => (
                  <button
                    key={mode}
                    onClick={() => updateField('mode', mode)}
                    disabled={disabled}
                    className={`flex-1 py-2 px-3 rounded text-xs transition-colors ${
                      safeConfig.mode === mode
                        ? 'bg-purple-600 text-white'
                        : 'bg-white/5 text-gray-400 hover:bg-white/10'
                    }`}
                  >
                    {t(mode)}
                  </button>
                ))}
              </div>
            </div>

            {/* Agent Configuration */}
            <div className="space-y-2">
              <h4 className="text-xs font-medium mb-2" style={{ color: '#EAECEF' }}>
                {t('agentRoles')}
              </h4>
              {safeConfig.agents.map((agent, index) => (
                <div
                  key={agent.role}
                  className="p-3 rounded-lg flex items-center justify-between gap-4"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139' }}
                >
                  <div className="flex items-center gap-2 min-w-[120px]">
                    <Bot className="w-4 h-4 text-purple-400" />
                    <span className="text-sm" style={{ color: '#EAECEF' }}>
                      {t(agent.role)}
                    </span>
                  </div>

                  <select
                    value={agent.model_id}
                    onChange={(e) => updateAgent(index, 'model_id', e.target.value)}
                    disabled={disabled || !agent.enabled}
                    className="flex-1 px-2 py-1.5 rounded text-xs bg-[#1E2329] border border-[#2B3139] text-[#EAECEF] outline-none focus:border-purple-500"
                  >
                    <option value="">{t('selectModel')}</option>
                    {aiModels.map((model) => (
                      <option key={model.id} value={model.id}>
                        {model.name}
                      </option>
                    ))}
                  </select>

                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={agent.enabled}
                      onChange={(e) => updateAgent(index, 'enabled', e.target.checked)}
                      disabled={disabled}
                      className="w-4 h-4 rounded border-gray-600 bg-transparent text-purple-500 focus:ring-offset-0 focus:ring-0"
                    />
                    <span className="text-xs text-gray-400">{agent.enabled ? 'ON' : 'OFF'}</span>
                  </label>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
