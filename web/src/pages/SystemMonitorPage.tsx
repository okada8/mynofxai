import { useState, useEffect, useRef } from 'react'
import {
  Cpu,
  Server,
  Clock,
  Terminal as TerminalIcon,
  Container,
  PlayCircle,
  StopCircle,
  PauseCircle,
} from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import { api } from '../lib/api' // Keep api import for other stats

// Add AttachAddon from xterm-addon-attach if available, or implement simple WebSocket handler
// For now, we'll implement a custom simple WebSocket handler to avoid another dependency issue

// Define interface locally to avoid import issues
import { SystemStats } from '../types'

export function SystemMonitorPage() {
  console.log('[SystemMonitorPage] Rendered')
  const { language } = useLanguage()
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.getSystemStats()
        setStats(data)
        setError(null)
      } catch (err) {
        console.error('Failed to fetch system stats:', err)
        setError('Failed to fetch system stats')
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
    const interval = setInterval(fetchStats, 3000) // Update every 3 seconds

    return () => clearInterval(interval)
  }, [])

  const formatBytes = (bytes: number, decimals = 2) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
  }

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / (3600 * 24))
    const hours = Math.floor((seconds % (3600 * 24)) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)

    const parts = []
    if (days > 0) parts.push(`${days}d`)
    if (hours > 0) parts.push(`${hours}h`)
    if (minutes > 0) parts.push(`${minutes}m`)
    parts.push(`${secs}s`)

    return parts.join(' ')
  }

  // Helper to format labels based on language
  const labels = {
    title: language === 'zh' ? '系统监控' : 'System Monitor',
    os: language === 'zh' ? '操作系统' : 'OS',
    arch: language === 'zh' ? '架构' : 'Architecture',
    cpu: language === 'zh' ? 'CPU核心数' : 'CPU Cores',
    goroutines: language === 'zh' ? 'Go协程数' : 'Go Routines',
    memory: language === 'zh' ? '内存使用' : 'Memory Usage',
    uptime: language === 'zh' ? '运行时间' : 'Uptime',
    hostname: language === 'zh' ? '主机名' : 'Hostname',
    kernel: language === 'zh' ? '内核版本' : 'Kernel',
    platform: language === 'zh' ? '平台' : 'Platform',
    storage: language === 'zh' ? '存储' : 'Storage',
    containers: language === 'zh' ? 'Docker 容器' : 'Docker Containers',
    status: language === 'zh' ? '状态' : 'Status',
    image: language === 'zh' ? '镜像' : 'Image',
    created: language === 'zh' ? '创建时间' : 'Created',
    terminal: language === 'zh' ? '终端' : 'Terminal',
  }

  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermInstance = useRef<Terminal | null>(null)

  useEffect(() => {
    if (!terminalRef.current || xtermInstance.current) return

    const term = new Terminal({
      theme: {
        background: '#000000',
        foreground: '#00ff00',
        cursor: '#00ff00',
      },
      fontSize: 14,
      fontFamily: 'monospace',
      cursorBlink: true,
      rows: 12,
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)

    term.open(terminalRef.current)
    fitAddon.fit()

    xtermInstance.current = term

    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.hostname}:8080/api/ws/terminal` // Adjust port if needed, assuming backend is on 8080

    let socket: WebSocket | null = null

    try {
      socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        term.writeln('\r\n\x1b[1;32m✓ Connected to backend terminal\x1b[0m\r\n')
      }

      socket.onmessage = (event) => {
        term.write(event.data)
      }

      socket.onclose = () => {
        term.writeln('\r\n\x1b[1;31m✗ Connection closed\x1b[0m\r\n')
      }

      socket.onerror = (error) => {
        term.writeln('\r\n\x1b[1;31m✗ Connection error\x1b[0m\r\n')
        console.error('WebSocket error:', error)
      }
    } catch (e) {
      term.writeln(`\r\n\x1b[1;31m✗ Failed to connect: ${e}\x1b[0m\r\n`)
    }

    // Handle user input
    term.onData((data) => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(data)
      } else {
        // Local echo fallback if not connected
        const code = data.charCodeAt(0)
        if (code === 13) {
          // Enter
          term.writeln('')
          term.write('\x1b[1;32m$ \x1b[0m')
        } else if (code === 127) {
          // Backspace
          term.write('\b \b')
        } else {
          term.write(data)
        }
      }
    })

    const handleResize = () => {
      fitAddon.fit()
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      term.dispose()
      xtermInstance.current = null
    }
  }, [])

  const getContainerIcon = (state: string) => {
    switch (state) {
      case 'running':
        return <PlayCircle className="w-5 h-5 text-green-500" />
      case 'exited':
        return <StopCircle className="w-5 h-5 text-red-500" />
      case 'paused':
        return <PauseCircle className="w-5 h-5 text-yellow-500" />
      default:
        return <Container className="w-5 h-5 text-gray-500" />
    }
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl text-white">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center gap-3 mb-8">
          <div className="p-3 rounded-lg bg-nofx-gold/10 text-nofx-gold">
            <TerminalIcon className="w-8 h-8" />
          </div>
          <div>
            <h1 className="text-3xl font-bold">{labels.title}</h1>
            <div className="flex items-center gap-2 text-sm text-gray-500 mt-1">
              <Clock className="w-3 h-3" />
              <span>Last updated: {new Date().toLocaleTimeString()}</span>
            </div>
          </div>
        </div>

        {loading && !stats ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-nofx-gold"></div>
          </div>
        ) : error ? (
          <div className="text-center text-red-500 py-10 bg-[#1E1E1E] rounded-xl border border-red-500/20">
            {error}
          </div>
        ) : stats ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* System Info Card */}
            <div className="bg-[#1E1E1E] rounded-xl p-6 border border-white/5 space-y-6">
              <div className="flex items-center gap-2 mb-4 text-nofx-gold">
                <Server className="w-6 h-6" />
                <h3 className="text-xl font-bold">
                  {language === 'zh' ? '系统信息' : 'System Info'}
                </h3>
              </div>

              <div className="grid gap-4">
                <div className="flex justify-between p-3 bg-black/20 rounded-lg">
                  <span className="text-gray-400">{labels.hostname}</span>
                  <span className="font-mono font-bold">
                    {stats.host_name || '-'}
                  </span>
                </div>
                <div className="flex justify-between p-3 bg-black/20 rounded-lg">
                  <span className="text-gray-400">{labels.os}</span>
                  <span className="font-mono">
                    {stats.os} {stats.platform}
                  </span>
                </div>
                <div className="flex justify-between p-3 bg-black/20 rounded-lg">
                  <span className="text-gray-400">{labels.kernel}</span>
                  <span className="font-mono text-sm" title={stats.kernel}>
                    {stats.kernel}
                  </span>
                </div>
                <div className="flex justify-between p-3 bg-black/20 rounded-lg">
                  <span className="text-gray-400">{labels.arch}</span>
                  <span className="font-mono">{stats.arch}</span>
                </div>
                <div className="flex justify-between p-3 bg-black/20 rounded-lg">
                  <span className="text-gray-400">{labels.uptime}</span>
                  <span className="font-mono">
                    {formatUptime(stats.uptime)}
                  </span>
                </div>
              </div>
            </div>

            {/* Resources Card */}
            <div className="bg-[#1E1E1E] rounded-xl p-6 border border-white/5 space-y-6">
              <div className="flex items-center gap-2 mb-4 text-nofx-gold">
                <Cpu className="w-6 h-6" />
                <h3 className="text-xl font-bold">
                  {language === 'zh' ? '资源使用' : 'Resources'}
                </h3>
              </div>

              <div className="space-y-8">
                {/* CPU / Cores */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="bg-black/20 p-4 rounded-lg text-center">
                    <div className="text-gray-400 text-sm mb-1">
                      {labels.cpu}
                    </div>
                    <div className="text-2xl font-mono font-bold">
                      {stats.num_cpu}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      Load: {stats.cpu_load?.toFixed(1)}%
                    </div>
                    {stats.cpu_temp > 0 && (
                      <div className="text-xs text-gray-500">
                        Temp: {stats.cpu_temp.toFixed(1)}°C
                      </div>
                    )}
                  </div>
                  <div className="bg-black/20 p-4 rounded-lg text-center">
                    <div className="text-gray-400 text-sm mb-1">
                      {labels.goroutines}
                    </div>
                    <div className="text-2xl font-mono font-bold">
                      {stats.go_routines}
                    </div>
                  </div>
                </div>

                {/* Memory */}
                <div className="bg-black/20 p-4 rounded-lg">
                  <div className="flex justify-between mb-2">
                    <span className="text-gray-400">{labels.memory}</span>
                    <span className="font-mono">
                      {formatBytes(stats.memory_used)} /{' '}
                      {formatBytes(stats.memory_total)}
                    </span>
                  </div>
                  {/* Memory Bar */}
                  <div className="h-4 bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        stats.memory_usage > 90
                          ? 'bg-red-500'
                          : stats.memory_usage > 70
                            ? 'bg-yellow-500'
                            : 'bg-green-500'
                      }`}
                      style={{ width: `${Math.min(stats.memory_usage, 100)}%` }}
                    />
                  </div>
                  <div className="text-right mt-1 text-sm text-gray-500 font-mono">
                    {stats.memory_usage?.toFixed(1) || '0.0'}%
                  </div>
                </div>

                {/* Storage */}
                <div className="bg-black/20 p-4 rounded-lg">
                  <div className="flex justify-between mb-2">
                    <span className="text-gray-400">{labels.storage}</span>
                    <span className="font-mono">
                      {formatBytes(stats.disk_used)} /{' '}
                      {formatBytes(stats.disk_total)}
                    </span>
                  </div>
                  {/* Storage Bar */}
                  <div className="h-4 bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        stats.disk_usage > 90
                          ? 'bg-red-500'
                          : stats.disk_usage > 70
                            ? 'bg-yellow-500'
                            : 'bg-green-500'
                      }`}
                      style={{ width: `${Math.min(stats.disk_usage, 100)}%` }}
                    />
                  </div>
                  <div className="text-right mt-1 text-sm text-gray-500 font-mono">
                    {stats.disk_usage?.toFixed(1) || '0.0'}%
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : null}

        {/* Docker Containers Section */}
        {stats && stats.containers && stats.containers.length > 0 && (
          <div className="bg-[#1E1E1E] rounded-xl p-6 border border-white/5 space-y-6">
            <div className="flex items-center gap-2 mb-4 text-nofx-gold">
              <Container className="w-6 h-6" />
              <h3 className="text-xl font-bold">{labels.containers}</h3>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-gray-700 text-gray-400">
                    <th className="pb-3 pl-2">{labels.status}</th>
                    <th className="pb-3">Name</th>
                    <th className="pb-3">{labels.image}</th>
                    <th className="pb-3">{labels.status}</th>
                    <th className="pb-3">{labels.created}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {stats.containers.map((container) => (
                    <tr
                      key={container.id}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="py-3 pl-2">
                        {getContainerIcon(container.state)}
                      </td>
                      <td className="py-3 font-mono font-bold text-white">
                        {container.name}
                        <div className="text-xs text-gray-500 font-normal">
                          {container.id}
                        </div>
                      </td>
                      <td
                        className="py-3 text-gray-300 font-mono text-xs truncate max-w-[200px]"
                        title={container.image}
                      >
                        {container.image}
                      </td>
                      <td className="py-3">
                        <span
                          className={`px-2 py-1 rounded-full text-xs font-medium ${
                            container.state === 'running'
                              ? 'bg-green-500/20 text-green-400'
                              : container.state === 'exited'
                                ? 'bg-red-500/20 text-red-400'
                                : 'bg-gray-500/20 text-gray-400'
                          }`}
                        >
                          {container.state}
                        </span>
                        <div className="text-xs text-gray-500 mt-0.5">
                          {container.status}
                        </div>
                      </td>
                      <td className="py-3 text-gray-400">
                        {new Date(container.created * 1000).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Terminal Section */}
        <div className="bg-[#1E1E1E] rounded-xl p-6 border border-white/5 space-y-6">
          <div className="flex items-center gap-2 mb-4 text-nofx-gold">
            <TerminalIcon className="w-6 h-6" />
            <h3 className="text-xl font-bold">{labels.terminal}</h3>
          </div>
          <div className="h-64 bg-black rounded-lg overflow-hidden p-2 border border-gray-800">
            <div ref={terminalRef} style={{ width: '100%', height: '100%' }} />
          </div>
        </div>
      </div>
    </div>
  )
}
