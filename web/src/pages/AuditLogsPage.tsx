import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import {
  Activity,
  Search,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Shield,
} from 'lucide-react'
import { api } from '../lib/api'
import { Container } from '../components/Container'

interface AuditLog {
  id: number
  user_id: string
  email: string
  action: string
  resource: string
  resource_name: string
  resource_type: string
  details: string
  ip_address: string
  country?: string
  user_agent: string
  status: string
  created_at: string
}

const getFlagEmoji = (countryCode: string) => {
  if (!countryCode) return ''
  if (countryCode === 'Local' || countryCode === 'LAN') return '🖥️'
  if (countryCode.length > 2) return '🌐' // Fallback for other non-ISO codes

  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [limit] = useState(20)
  const [offset, setOffset] = useState(0)

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const res = await api.getAuditLogs(limit, offset)
      setLogs(res.data)
      setTotal(res.total)
    } catch (error) {
      console.error('Failed to fetch audit logs:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
  }, [offset])

  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit) + 1

  return (
    <Container>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-3 rounded-xl bg-nofx-gold/10 border border-nofx-gold/20">
              <Shield className="w-6 h-6 text-nofx-gold" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">审计日志</h1>
              <p className="text-nofx-text-muted">系统操作安全审计记录</p>
            </div>
          </div>
          <button
            onClick={fetchLogs}
            className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-nofx-text-muted hover:text-white hover:border-zinc-700 transition-colors"
          >
            <RefreshCw className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {/* Table */}
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden backdrop-blur-sm">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-zinc-800 bg-black/20">
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    时间
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    操作者
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    动作
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    资源
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    IP/地区
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    状态
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-nofx-text-muted uppercase tracking-wider">
                    详情
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800">
                {loading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-6 py-12 text-center text-nofx-text-muted"
                    >
                      加载中...
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-6 py-12 text-center text-nofx-text-muted"
                    >
                      暂无审计日志
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <motion.tr
                      key={log.id}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="group hover:bg-white/5 transition-colors"
                    >
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-nofx-text-muted font-mono">
                        {new Date(log.created_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-white">
                        {log.email || log.user_id || 'System'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 py-1 rounded text-xs font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                          {log.action}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-nofx-text-muted font-mono">
                        <div className="flex flex-col">
                          <span className="text-white">
                            {log.resource_name || log.resource}
                          </span>
                          {log.resource_type && (
                            <span className="text-xs text-nofx-text-muted mt-0.5 px-1.5 py-0.5 bg-white/5 rounded w-fit">
                              {log.resource_type}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-nofx-text-muted">
                        <div className="flex flex-col">
                          <span className="font-mono text-white">
                            {log.ip_address || '-'}
                          </span>
                          {log.country && (
                            <span className="text-xs text-nofx-text-muted mt-0.5 flex items-center gap-1">
                              <span className="inline-block w-4 text-center">
                                {getFlagEmoji(log.country)}
                              </span>
                              {log.country}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 py-1 rounded text-xs font-bold border ${
                            log.status === 'SUCCESS'
                              ? 'bg-green-500/10 text-green-400 border-green-500/20'
                              : 'bg-red-500/10 text-red-400 border-red-500/20'
                          }`}
                        >
                          {log.status}
                        </span>
                      </td>
                      <td
                        className="px-6 py-4 text-sm text-nofx-text-muted max-w-xs truncate"
                        title={log.details}
                      >
                        {log.details}
                      </td>
                    </motion.tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="px-6 py-4 border-t border-zinc-800 flex items-center justify-between">
            <div className="text-sm text-nofx-text-muted">
              显示 {offset + 1} 到 {Math.min(offset + limit, total)} 条，共{' '}
              {total} 条
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setOffset(Math.max(0, offset - limit))}
                disabled={offset === 0}
                className="p-2 rounded-lg hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                <ChevronLeft className="w-5 h-5 text-nofx-text-muted" />
              </button>
              <span className="text-sm text-white font-medium">
                {currentPage} / {totalPages || 1}
              </span>
              <button
                onClick={() => setOffset(offset + limit)}
                disabled={offset + limit >= total}
                className="p-2 rounded-lg hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                <ChevronRight className="w-5 h-5 text-nofx-text-muted" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </Container>
  )
}
