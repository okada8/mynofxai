import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useLanguage } from '../contexts/LanguageContext'
import {
  Plus,
  Trash2,
  Play,
  Edit2,
  Clock,
  X,
  Activity
} from 'lucide-react'
import type { Task, TraderInfo } from '../types'
import { notify } from '../lib/notify'
import { DeepVoidBackground } from '../components/DeepVoidBackground'

const API_BASE = import.meta.env.VITE_API_BASE || ''

export function TasksPage() {
  const { token } = useAuth()
  const { language } = useLanguage()
  const [tasks, setTasks] = useState<Task[]>([])
  const [traders, setTraders] = useState<TraderInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  
  // Form state
  const [formData, setFormData] = useState({
    name: '',
    type: 'report',
    trader_id: '',
    cron_expression: '@hourly',
    enabled: true,
    params: '{}'
  })

  const fetchTasks = async () => {
    setIsLoading(true)
    try {
      const res = await fetch(`${API_BASE}/api/tasks`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!res.ok) throw new Error('Failed to fetch tasks')
      const data = await res.json()
      setTasks(data)
    } catch (err) {
      console.error(err)
      notify.error(language === 'zh' ? '加载任务失败' : 'Failed to load tasks')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchTraders = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/my-traders`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!res.ok) throw new Error('Failed to fetch traders')
      const data = await res.json()
      setTraders(data)
    } catch (err) {
      console.error('Failed to fetch traders:', err)
    }
  }

  useEffect(() => {
    if (token) {
      fetchTasks()
      fetchTraders()
    }
  }, [token])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const url = editingTask 
        ? `${API_BASE}/api/tasks/${editingTask.id}`
        : `${API_BASE}/api/tasks`
      
      const method = editingTask ? 'PUT' : 'POST'
      
      // Convert form data to API format
      const apiData = {
        name: formData.name,
        type: formData.type,
        trader_id: formData.trader_id,
        cron_expression: formData.cron_expression,
        enabled: formData.enabled,
        params: formData.params
      }
      
      console.log('Sending data:', apiData) // Debug log

      const res = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify(apiData)
      })

      if (!res.ok) throw new Error('Operation failed')
      
      notify.success(language === 'zh' ? '保存成功' : 'Saved successfully')
      
      // Close modal first
      setIsModalOpen(false)
      
      // Delay fetch to ensure DB write is committed and query returns new data
      setTimeout(() => {
        fetchTasks()
      }, 500)
    } catch (err) {
      notify.error(language === 'zh' ? '保存失败' : 'Failed to save')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(language === 'zh' ? '确定删除吗？' : 'Are you sure?')) return
    
    try {
      const res = await fetch(`${API_BASE}/api/tasks/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!res.ok) throw new Error('Delete failed')
      notify.success(language === 'zh' ? '删除成功' : 'Deleted successfully')
      fetchTasks()
    } catch (err) {
      notify.error(language === 'zh' ? '删除失败' : 'Failed to delete')
    }
  }

  const handleRun = async (id: string) => {
    try {
      const res = await fetch(`${API_BASE}/api/tasks/${id}/run`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!res.ok) throw new Error('Run failed')
      notify.success(language === 'zh' ? '已触发执行' : 'Task triggered')
      // Refresh to update last run time if needed, though usually not immediate
      setTimeout(fetchTasks, 1000)
    } catch (err) {
      notify.error(language === 'zh' ? '执行失败' : 'Failed to run')
    }
  }

  const openModal = (task?: Task) => {
    if (task) {
      setEditingTask(task)
      setFormData({
        name: task.name,
        type: task.type,
        trader_id: task.trader_id,
        cron_expression: task.cron_expression,
        enabled: task.enabled,
        params: task.params
      })
    } else {
      setEditingTask(null)
      setFormData({
        name: '',
        type: 'report',
        trader_id: traders.length > 0 ? traders[0].trader_id : '',
        cron_expression: '@hourly',
        enabled: true,
        params: '{}'
      })
    }
    setIsModalOpen(true)
  }

  return (
    <DeepVoidBackground className="py-8">
      <div className="container mx-auto px-4 max-w-7xl">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-nofx-gold to-yellow-200">
              {language === 'zh' ? '定时任务管理' : 'Scheduled Tasks'}
            </h1>
            <p className="text-nofx-text-muted mt-2">
              {language === 'zh' ? '管理系统中的自动运行任务' : 'Manage automated tasks in the system'}
            </p>
          </div>
          <button
            onClick={() => openModal()}
            className="flex items-center gap-2 bg-nofx-gold text-black px-4 py-2 rounded-lg font-bold hover:bg-yellow-400 transition-colors"
          >
            <Plus className="w-4 h-4" />
            {language === 'zh' ? '新建任务' : 'New Task'}
          </button>
        </div>

        {/* Task List */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {tasks.map(task => (
            <div key={task.id} className="bg-[#1E2329]/80 backdrop-blur border border-white/5 rounded-xl p-6 hover:border-nofx-gold/30 transition-all group">
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${task.enabled ? 'bg-nofx-gold/20 text-nofx-gold' : 'bg-zinc-800 text-zinc-500'}`}>
                    {task.type === 'report' ? <Activity className="w-5 h-5" /> : <Clock className="w-5 h-5" />}
                  </div>
                  <div>
                    <h3 className="font-bold text-white">{task.name}</h3>
                    <div className="flex items-center gap-2 text-xs text-nofx-text-muted">
                      <span className="bg-white/10 px-1.5 py-0.5 rounded">{task.type}</span>
                      <span>{task.cron_expression}</span>
                    </div>
                  </div>
                </div>
                <div className="relative">
                  <div className={`w-2 h-2 rounded-full ${task.enabled ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                </div>
              </div>

              <div className="space-y-2 mb-6 text-sm text-nofx-text-muted">
                <div className="flex justify-between">
                  <span>Last Run:</span>
                  <span className="font-mono text-white">
                    {task.last_run_time > 0 ? new Date(task.last_run_time).toLocaleString() : '-'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Next Run:</span>
                  <span className="font-mono text-nofx-gold">
                    {task.next_run_time > 0 ? new Date(task.next_run_time).toLocaleString() : '-'}
                  </span>
                </div>
                {task.trader_id && (
                  <div className="flex justify-between">
                    <span>Trader:</span>
                    <span className="font-mono text-xs truncate max-w-[150px]">
                      {traders.find(t => t.trader_id === task.trader_id)?.trader_name || task.trader_id}
                    </span>
                  </div>
                )}
              </div>

              <div className="flex gap-2 pt-4 border-t border-white/5">
                <button
                  onClick={() => handleRun(task.id)}
                  className="flex-1 flex items-center justify-center gap-2 py-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs font-bold transition-colors"
                  title="Run Now"
                >
                  <Play className="w-3 h-3" />
                  {language === 'zh' ? '立即运行' : 'Run'}
                </button>
                <button
                  onClick={() => openModal(task)}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-nofx-text-muted hover:text-white transition-colors"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleDelete(task.id)}
                  className="p-2 rounded-lg bg-white/5 hover:bg-red-500/20 text-nofx-text-muted hover:text-red-500 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
          
          {tasks.length === 0 && !isLoading && (
            <div className="col-span-full flex flex-col items-center justify-center py-12 text-nofx-text-muted border border-dashed border-white/10 rounded-xl">
              <Clock className="w-12 h-12 mb-4 opacity-50" />
              <p>{language === 'zh' ? '暂无定时任务' : 'No scheduled tasks'}</p>
            </div>
          )}
        </div>

        {/* Modal */}
        {isModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
            <div className="bg-[#1E2329] border border-nofx-gold/20 rounded-2xl w-full max-w-lg p-6 shadow-2xl">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold">
                  {editingTask ? (language === 'zh' ? '编辑任务' : 'Edit Task') : (language === 'zh' ? '新建任务' : 'New Task')}
                </h2>
                <button onClick={() => setIsModalOpen(false)} className="text-gray-400 hover:text-white">
                  <X className="w-6 h-6" />
                </button>
              </div>

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-1">
                    {language === 'zh' ? '任务名称' : 'Name'}
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={e => setFormData({...formData, name: e.target.value})}
                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 focus:border-nofx-gold outline-none"
                    required
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm text-gray-400 mb-1">
                      {language === 'zh' ? '任务类型' : 'Type'}
                    </label>
                    <select
                      value={formData.type}
                      onChange={e => setFormData({...formData, type: e.target.value})}
                      className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 focus:border-nofx-gold outline-none"
                    >
                      <option value="report">Report (持仓报告)</option>
                      <option value="sync">Sync (数据同步)</option>
                      <option value="custom">Custom (自定义)</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm text-gray-400 mb-1">
                      {language === 'zh' ? 'Cron 表达式' : 'Cron Expression'}
                    </label>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={formData.cron_expression}
                        onChange={e => setFormData({...formData, cron_expression: e.target.value})}
                        placeholder="* * * * *"
                        className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 focus:border-nofx-gold outline-none"
                        required
                      />
                      <div className="relative group">
                         <div className="w-10 h-10 flex items-center justify-center bg-white/5 rounded-lg border border-white/10 cursor-help">
                           <span className="text-nofx-gold text-lg">?</span>
                         </div>
                         <div className="absolute right-0 bottom-full mb-2 w-64 p-4 bg-[#1E2329] border border-white/10 rounded-lg shadow-xl text-xs text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50">
                           <p className="font-bold text-white mb-2">{language === 'zh' ? 'Cron 格式说明:' : 'Cron Format:'}</p>
                           <ul className="space-y-1 font-mono">
                             <li>* * * * *</li>
                             <li>│ │ │ │ │</li>
                             <li>│ │ │ │ └── {language === 'zh' ? '周 (0-6)' : 'Week (0-6)'}</li>
                             <li>│ │ │ └──── {language === 'zh' ? '月 (1-12)' : 'Month (1-12)'}</li>
                             <li>│ │ └────── {language === 'zh' ? '日 (1-31)' : 'Day (1-31)'}</li>
                             <li>│ └──────── {language === 'zh' ? '时 (0-23)' : 'Hour (0-23)'}</li>
                             <li>└────────── {language === 'zh' ? '分 (0-59)' : 'Minute (0-59)'}</li>
                           </ul>
                           <div className="mt-2 pt-2 border-t border-white/10">
                             <p className="text-nofx-gold">@hourly = 0 * * * *</p>
                             <p className="text-nofx-gold">@daily = 0 0 * * *</p>
                           </div>
                         </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div>
                  <label className="block text-sm text-gray-400 mb-1">
                    {language === 'zh' ? '交易员 (可选)' : 'Trader (Optional)'}
                  </label>
                  <select
                    value={formData.trader_id}
                    onChange={e => setFormData({...formData, trader_id: e.target.value})}
                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 focus:border-nofx-gold outline-none"
                  >
                    <option value="">{language === 'zh' ? '-- 不选择 --' : '-- None --'}</option>
                    {traders.map(t => (
                      <option key={t.trader_id} value={t.trader_id}>
                        {t.trader_name} ({t.ai_model})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm text-gray-400 mb-1">
                    {language === 'zh' ? '脚本路径' : 'Script Path'}
                  </label>
                  <input
                    type="text"
                    value={formData.params}
                    onChange={e => setFormData({...formData, params: e.target.value})}
                    placeholder="/path/to/script.sh"
                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 focus:border-nofx-gold outline-none font-mono text-sm"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    {language === 'zh' 
                      ? '当任务类型为 Custom 时生效，填写服务器上的绝对路径' 
                      : 'Required for Custom task type. Absolute path on server.'}
                  </p>
                </div>

                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="enabled"
                    checked={formData.enabled}
                    onChange={e => setFormData({...formData, enabled: e.target.checked})}
                    className="w-4 h-4 rounded border-gray-600 text-nofx-gold focus:ring-nofx-gold bg-black/40"
                  />
                  <label htmlFor="enabled" className="text-sm">
                    {language === 'zh' ? '启用' : 'Enabled'}
                  </label>
                </div>

                <div className="flex gap-4 pt-4">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="flex-1 py-3 rounded-lg border border-white/10 hover:bg-white/5 transition-colors font-bold"
                  >
                    {language === 'zh' ? '取消' : 'Cancel'}
                  </button>
                  <button
                    type="submit"
                    className="flex-1 py-3 rounded-lg bg-nofx-gold text-black hover:bg-yellow-400 transition-colors font-bold"
                  >
                    {language === 'zh' ? '保存' : 'Save'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </DeepVoidBackground>
  )
}

