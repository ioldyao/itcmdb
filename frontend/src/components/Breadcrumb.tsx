import { Link, useLocation } from 'react-router-dom'
import { ChevronRight, Home } from 'lucide-react'

const breadcrumbMap: Record<string, string> = {
  '/dashboard': '首页',
  '/cmdb': 'CMDB',
  '/cmdb/servers': '服务器',
  '/cmdb/networks': '网络设备',
  '/cmdb/applications': '应用系统',
  '/cmdb/containers': '容器实例',
  '/cmdb/roles': '角色管理',
  '/cmdb/tags': '标签管理',
  '/cmdb/victoriametrics': 'VictoriaMetrics',
  '/tickets': '工单',
  '/tickets/create': '创建工单',
  '/alerts': '告警',
  '/alerts/rules': '规则配置',
  '/alerts/history': '历史告警',
  '/alerts/integration/webhook': 'Webhook',
  '/alerts/receivers': '告警接收',
  '/admin': '系统管理',
  '/admin/users': '用户管理',
  '/admin/roles': '角色管理',
  '/admin/audit': '审计日志',
  '/reports': '报表中心',
  '/profile': '个人中心',
}

export default function Breadcrumb() {
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  // Build breadcrumb items
  const items: { label: string; path?: string }[] = []

  // Always start with home
  items.push({ label: '首页', path: '/dashboard' })

  // Add each path segment
  let accumulatedPath = ''
  pathSegments.forEach((segment) => {
    accumulatedPath += '/' + segment
    const label = breadcrumbMap[accumulatedPath]
    if (label) {
      items.push({ label, path: accumulatedPath })
    }
  })

  // Make the last item non-clickable
  if (items.length > 1) {
    items[items.length - 1].path = undefined
  }

  if (items.length <= 1) return null

  return (
    <nav className="flex items-center gap-1 text-sm text-gray-500 dark:text-text-secondary mb-4">
      {items.map((item, index) => (
        <span key={item.path || item.label} className="flex items-center gap-1">
          {index > 0 && <ChevronRight size={14} className="text-gray-400" />}
          {item.path ? (
            <Link
              to={item.path}
              className="hover:text-brand-primary transition-colors"
            >
              {index === 0 && <Home size={14} className="inline mr-1" />}
              {item.label}
            </Link>
          ) : (
            <span className="text-gray-900 dark:text-text-primary font-medium">
              {item.label}
            </span>
          )}
        </span>
      ))}
    </nav>
  )
}
