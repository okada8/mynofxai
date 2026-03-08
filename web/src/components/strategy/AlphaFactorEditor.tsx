import { BarChart3 } from 'lucide-react'
import type { AlphaFactorConfig } from '../../types'

interface AlphaFactorEditorProps {
  config?: AlphaFactorConfig
  onChange: (config: AlphaFactorConfig) => void
  disabled?: boolean
  language: string
}

export function AlphaFactorEditor({
  config,
  onChange,
  disabled,
  language,
}: AlphaFactorEditorProps) {
  // Default values if config is undefined
  const safeConfig: AlphaFactorConfig = config || {
    enable_liquidation_clusters: false,
    liquidation_cluster_weight: 1.0,
    enable_netflow_ranking: false,
    netflow_ranking_weight: 1.0,
    enable_whale_activity: false,
    whale_activity_weight: 1.0,
    enable_sentiment_analysis: false,
    sentiment_weight: 1.0,
  }

  const t = (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      alphaFactors: { zh: 'Alpha 因子', en: 'Alpha Factors' },
      alphaFactorsDesc: { zh: '配置影响交易决策的 Alpha 信号源及其权重', en: 'Configure Alpha signal sources and weights for trading decisions' },
      liquidationClusters: { zh: '清算集群 (Liquidation)', en: 'Liquidation Clusters' },
      liquidationDesc: { zh: '基于 CoinAnk 清算热力图，识别多空博弈关键位', en: 'Identify key levels based on CoinAnk liquidation heatmap' },
      netflowRanking: { zh: '资金流向 (NetFlow)', en: 'NetFlow Ranking' },
      netflowDesc: { zh: '基于 NofxOS 机构/散户资金流向排行', en: 'Institutional/Retail fund flow ranking from NofxOS' },
      whaleActivity: { zh: '鲸鱼活动 (Whale)', en: 'Whale Activity' },
      whaleDesc: { zh: '大额链上转账与交易所充提监控', en: 'Large on-chain transfers and exchange flow monitoring' },
      sentiment: { zh: '市场情绪 (Sentiment)', en: 'Market Sentiment' },
      sentimentDesc: { zh: '社交媒体与新闻情绪分析', en: 'Social media and news sentiment analysis' },
      weight: { zh: '权重', en: 'Weight' },
    }
    return translations[key]?.[language] || key
  }

  const updateField = <K extends keyof AlphaFactorConfig>(
    key: K,
    value: AlphaFactorConfig[K]
  ) => {
    if (!disabled) {
      onChange({ ...safeConfig, [key]: value })
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2 mb-2">
          <BarChart3 className="w-5 h-5" style={{ color: '#F0B90B' }} />
          <h3 className="font-medium" style={{ color: '#EAECEF' }}>
            {t('alphaFactors')}
          </h3>
        </div>
        <p className="text-xs mb-4" style={{ color: '#848E9C' }}>
          {t('alphaFactorsDesc')}
        </p>

        <div className="space-y-4">
          {/* Liquidation Clusters */}
          <div className="p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
            <div className="flex items-center justify-between mb-2">
              <div>
                <label className="flex items-center gap-2 text-sm font-medium" style={{ color: '#EAECEF' }}>
                  <input
                    type="checkbox"
                    checked={safeConfig.enable_liquidation_clusters}
                    onChange={(e) => updateField('enable_liquidation_clusters', e.target.checked)}
                    disabled={disabled}
                    className="w-4 h-4 rounded border-gray-600 bg-transparent text-yellow-500 focus:ring-offset-0 focus:ring-0"
                  />
                  {t('liquidationClusters')}
                </label>
                <p className="text-xs mt-1" style={{ color: '#848E9C' }}>
                  {t('liquidationDesc')}
                </p>
              </div>
              {safeConfig.enable_liquidation_clusters && (
                <div className="flex items-center gap-2">
                  <span className="text-xs" style={{ color: '#848E9C' }}>{t('weight')}:</span>
                  <input
                    type="number"
                    value={safeConfig.liquidation_cluster_weight}
                    onChange={(e) => updateField('liquidation_cluster_weight', parseFloat(e.target.value) || 0)}
                    disabled={disabled}
                    step={0.1}
                    min={0}
                    max={10}
                    className="w-16 px-2 py-1 rounded text-xs text-center"
                    style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              )}
            </div>
          </div>

          {/* NetFlow Ranking */}
          <div className="p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
            <div className="flex items-center justify-between mb-2">
              <div>
                <label className="flex items-center gap-2 text-sm font-medium" style={{ color: '#EAECEF' }}>
                  <input
                    type="checkbox"
                    checked={safeConfig.enable_netflow_ranking}
                    onChange={(e) => updateField('enable_netflow_ranking', e.target.checked)}
                    disabled={disabled}
                    className="w-4 h-4 rounded border-gray-600 bg-transparent text-yellow-500 focus:ring-offset-0 focus:ring-0"
                  />
                  {t('netflowRanking')}
                </label>
                <p className="text-xs mt-1" style={{ color: '#848E9C' }}>
                  {t('netflowDesc')}
                </p>
              </div>
              {safeConfig.enable_netflow_ranking && (
                <div className="flex items-center gap-2">
                  <span className="text-xs" style={{ color: '#848E9C' }}>{t('weight')}:</span>
                  <input
                    type="number"
                    value={safeConfig.netflow_ranking_weight}
                    onChange={(e) => updateField('netflow_ranking_weight', parseFloat(e.target.value) || 0)}
                    disabled={disabled}
                    step={0.1}
                    min={0}
                    max={10}
                    className="w-16 px-2 py-1 rounded text-xs text-center"
                    style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Whale Activity */}
          <div className="p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
            <div className="flex items-center justify-between mb-2">
              <div>
                <label className="flex items-center gap-2 text-sm font-medium" style={{ color: '#EAECEF' }}>
                  <input
                    type="checkbox"
                    checked={safeConfig.enable_whale_activity}
                    onChange={(e) => updateField('enable_whale_activity', e.target.checked)}
                    disabled={disabled}
                    className="w-4 h-4 rounded border-gray-600 bg-transparent text-yellow-500 focus:ring-offset-0 focus:ring-0"
                  />
                  {t('whaleActivity')}
                </label>
                <p className="text-xs mt-1" style={{ color: '#848E9C' }}>
                  {t('whaleDesc')}
                </p>
              </div>
              {safeConfig.enable_whale_activity && (
                <div className="flex items-center gap-2">
                  <span className="text-xs" style={{ color: '#848E9C' }}>{t('weight')}:</span>
                  <input
                    type="number"
                    value={safeConfig.whale_activity_weight}
                    onChange={(e) => updateField('whale_activity_weight', parseFloat(e.target.value) || 0)}
                    disabled={disabled}
                    step={0.1}
                    min={0}
                    max={10}
                    className="w-16 px-2 py-1 rounded text-xs text-center"
                    style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Sentiment Analysis */}
          <div className="p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
            <div className="flex items-center justify-between mb-2">
              <div>
                <label className="flex items-center gap-2 text-sm font-medium" style={{ color: '#EAECEF' }}>
                  <input
                    type="checkbox"
                    checked={safeConfig.enable_sentiment_analysis}
                    onChange={(e) => updateField('enable_sentiment_analysis', e.target.checked)}
                    disabled={disabled}
                    className="w-4 h-4 rounded border-gray-600 bg-transparent text-yellow-500 focus:ring-offset-0 focus:ring-0"
                  />
                  {t('sentiment')}
                </label>
                <p className="text-xs mt-1" style={{ color: '#848E9C' }}>
                  {t('sentimentDesc')}
                </p>
              </div>
              {safeConfig.enable_sentiment_analysis && (
                <div className="flex items-center gap-2">
                  <span className="text-xs" style={{ color: '#848E9C' }}>{t('weight')}:</span>
                  <input
                    type="number"
                    value={safeConfig.sentiment_weight}
                    onChange={(e) => updateField('sentiment_weight', parseFloat(e.target.value) || 0)}
                    disabled={disabled}
                    step={0.1}
                    min={0}
                    max={10}
                    className="w-16 px-2 py-1 rounded text-xs text-center"
                    style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
