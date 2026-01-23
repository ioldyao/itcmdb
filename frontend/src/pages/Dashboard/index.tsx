import { motion } from 'framer-motion'
import { Server, FileText, Bell, PieChart, Settings as SettingsIcon, Activity, TrendingUp } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

interface StatCardProps {
  title: string
  value: string | number
  trend?: string
  trendUp?: boolean
  icon: typeof Server
  color: string
  path: string
}

const StatCard = ({ title, value, trend, trendUp, icon: Icon, color, path }: StatCardProps) => {
  const navigate = useNavigate()

  return (
    <motion.div
      whileHover={{ scale: 1.02, y: -4 }}
      whileTap={{ scale: 0.98 }}
      onClick={() => navigate(path)}
      className={`
        relative cursor-pointer
        bg-white dark:bg-gradient-card rounded-xl p-6
        border border-gray-200 dark:border-white/8
        shadow-card-light dark:shadow-card hover:shadow-card-hover-light dark:hover:shadow-card-hover
        transition-all duration-300
        overflow-hidden
        group
      `}
    >
      {/* 背景渐变效果 */}
      <div className={`absolute inset-0 bg-gradient-to-br ${color} opacity-0 group-hover:opacity-10 transition-opacity`} />

      {/* 图标 */}
      <div className={`w-12 h-12 rounded-lg bg-gradient-to-br ${color} flex items-center justify-center mb-4`}>
        <Icon size={24} className="text-white" />
      </div>

      {/* 标题 */}
      <h3 className="text-gray-600 dark:text-text-secondary text-sm mb-2">{title}</h3>

      {/* 数值 */}
      <div className="flex items-end justify-between">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-3xl font-semibold text-gray-900 dark:text-text-primary"
        >
          {value}
        </motion.div>

        {/* 趋势 */}
        {trend && (
          <div className={`flex items-center gap-1 text-sm ${trendUp ? 'text-green-400' : 'text-red-400'}`}>
            <TrendingUp size={16} className={trendUp ? '' : 'rotate-180'} />
            <span>{trend}</span>
          </div>
        )}
      </div>

      {/* 查看详情 */}
      <div className="mt-4 flex items-center gap-2 text-brand-primary text-sm opacity-0 group-hover:opacity-100 transition-opacity">
        <span>查看详情</span>
        <span>→</span>
      </div>
    </motion.div>
  )
}

export default function Dashboard() {
  const navigate = useNavigate()

  const stats = [
    {
      title: 'CMDB 资产总数',
      value: '1,234',
      trend: '+12%',
      trendUp: true,
      icon: Server,
      color: 'from-blue-500 to-blue-600',
      path: '/cmdb/servers',
    },
    {
      title: '待处理工单',
      value: 45,
      trend: '8 紧急',
      trendUp: false,
      icon: FileText,
      color: 'from-orange-500 to-orange-600',
      path: '/tickets',
    },
    {
      title: '活跃告警',
      value: 12,
      trend: '3 严重',
      trendUp: false,
      icon: Bell,
      color: 'from-red-500 to-red-600',
      path: '/alerts',
    },
    {
      title: '资产健康度',
      value: '92%',
      trend: '+5%',
      trendUp: true,
      icon: Activity,
      color: 'from-green-500 to-green-600',
      path: '/reports',
    },
    {
      title: '本月工单',
      value: 156,
      trend: '-8%',
      trendUp: false,
      icon: PieChart,
      color: 'from-purple-500 to-purple-600',
      path: '/reports',
    },
    {
      title: '系统配置',
      value: '正常',
      icon: SettingsIcon,
      color: 'from-gray-500 to-gray-600',
      path: '/admin',
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

      {/* 卡片工作台 */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.1 }}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      >
        {stats.map((stat, index) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
          >
            <StatCard {...stat} />
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
            { label: '创建工单', path: '/tickets/create', icon: FileText },
            { label: '查看告警', path: '/alerts', icon: Bell },
            { label: '生成报表', path: '/reports', icon: PieChart },
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
