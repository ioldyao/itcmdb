import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Server, FileText, Bell, Settings as SettingsIcon, Activity, TrendingUp } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

interface StatCardProps {
  title: string
  value: string | number
  trend?: string
  trendUp?: boolean
  icon: typeof Server
  color: string
  path: string
  loading?: boolean
}

const StatCard = ({ title, value, trend, trendUp, icon: Icon, color, path, loading }: StatCardProps) => {
  const navigate = useNavigate()

  return (
    <motion.div
      whileHover={{ scale: 1.02, y: -4 }}
      whileTap={{ scale: 0.98 }}
      onClick={() => navigate(path)}
      className={`
        relative cursor-pointer
        bg-white dark:bg-bg-secondary rounded-xl p-6
        border border-gray-200 dark:border-white/8
        shadow-sm dark:shadow-card hover:shadow-md dark:hover:shadow-card-hover
        transition-all duration-300
        overflow-hidden
        group
      `}
    >
      <div className={`absolute inset-0 bg-gradient-to-br ${color} opacity-0 group-hover:opacity-10 transition-opacity`} />

      <div className={`w-12 h-12 rounded-lg bg-gradient-to-br ${color} flex items-center justify-center mb-4`}>
        <Icon size={24} className="text-white" />
      </div>

      <h3 className="text-gray-600 dark:text-text-secondary text-sm mb-2">{title}</h3>

      <div className="flex items-end justify-between">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-3xl font-semibold text-gray-900 dark:text-text-primary"
        >
          {loading ? (
            <div className="w-16 h-8 bg-gray-200 dark:bg-white/10 rounded animate-pulse" />
          ) : (
            value
          )}
        </motion.div>

        {trend && !loading && (
          <div className={`flex items-center gap-1 text-sm ${trendUp ? 'text-green-500' : 'text-red-500'}`}>
            <TrendingUp size={16} className={trendUp ? '' : 'rotate-180'} />
            <span>{trend}</span>
          </div>
        )}
      </div>

      <div className="mt-4 flex items-center gap-2 text-brand-primary text-sm opacity-0 group-hover:opacity-100 transition-opacity">
        <span>查看详情</span>
        <span>→</span>
      </div>
    </motion.div>
  )
}

export default function Dashboard() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState({
    ciCount: 0,
    ticketCount: 0,
    alertFiring: 0,
    alertTotal: 0,
  })

  useEffect(() => {
    fetchDashboardData()
  }, [])

  const fetchDashboardData = async () => {
    setLoading(true)
    try {
      const headers = { Authorization: `Bearer ${token}` }

      // Parallel fetch all stats
      const [ciRes, ticketRes, alertRes] = await Promise.allSettled([
        fetch('/api/v1/ci/instances?pageSize=1', { headers }).then(r => r.json()),
        fetch('/api/v1/tickets?pageSize=1', { headers }).then(r => r.json()),
        fetch('/api/v1/alerts/statistics', { headers }).then(r => r.json()),
      ])

      const ciCount = ciRes.status === 'fulfilled' ? (ciRes.value?.data?.total || ciRes.value?.total || 0) : 0
      const ticketCount = ticketRes.status === 'fulfilled' ? (ticketRes.value?.data?.total || ticketRes.value?.total || 0) : 0

      let alertFiring = 0
      let alertTotal = 0
      if (alertRes.status === 'fulfilled' && alertRes.value?.data) {
        alertFiring = alertRes.value.data.stats?.firing || 0
        alertTotal = alertRes.value.data.stats?.total || 0
      }

      setStats({ ciCount, ticketCount, alertFiring, alertTotal })
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const cardStats = [
    {
      title: 'CMDB 资产总数',
      value: stats.ciCount,
      icon: Server,
      color: 'from-blue-500 to-blue-600',
      path: '/cmdb/servers',
    },
    {
      title: '工单总数',
      value: stats.ticketCount,
      icon: FileText,
      color: 'from-orange-500 to-orange-600',
      path: '/tickets',
    },
    {
      title: '活跃告警',
      value: stats.alertFiring,
      trend: stats.alertFiring > 0 ? `${stats.alertFiring} 未恢复` : undefined,
      trendUp: false,
      icon: Bell,
      color: 'from-red-500 to-red-600',
      path: '/alerts',
    },
    {
      title: '告警总数',
      value: stats.alertTotal,
      icon: Activity,
      color: 'from-green-500 to-green-600',
      path: '/alerts',
    },
  ]

  return (
    <div className="p-8">
      {/* 欢迎区域 */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="mb-8"
      >
        <h1 className="text-3xl font-semibold text-gray-900 dark:text-text-primary mb-2">欢迎回来</h1>
        <p className="text-gray-600 dark:text-text-secondary">
          今日：{new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })}
        </p>
      </motion.div>

      {/* 统计卡片 */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.1 }}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        {cardStats.map((stat, index) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
          >
            <StatCard {...stat} loading={loading} />
          </motion.div>
        ))}
      </motion.div>

      {/* 快速操作 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
        className="mt-8"
      >
        <h2 className="text-xl font-semibold text-gray-900 dark:text-text-primary mb-4">快速操作</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: '添加资产', path: '/cmdb/servers', icon: Server },
            { label: '创建工单', path: '/tickets', icon: FileText },
            { label: '查看告警', path: '/alerts', icon: Bell },
            { label: '系统管理', path: '/admin', icon: SettingsIcon },
          ].map((action, index) => (
            <motion.button
              key={action.label}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => navigate(action.path)}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.4 + index * 0.05 }}
              className="flex items-center gap-3 p-4 rounded-lg bg-white dark:bg-bg-secondary border border-gray-200 dark:border-white/8 hover:border-brand-primary/50 hover:bg-gray-50 dark:hover:bg-white/5 transition-all"
            >
              <action.icon size={20} className="text-brand-primary" />
              <span className="text-gray-900 dark:text-text-primary">{action.label}</span>
            </motion.button>
          ))}
        </div>
      </motion.div>
    </div>
  )
}
