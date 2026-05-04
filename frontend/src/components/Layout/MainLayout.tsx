import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Avatar, Dropdown, Badge, Input } from 'antd'
import {
  LayoutDashboard,
  Server,
  FileText,
  Bell,
  BarChart3,
  Settings,
  User,
  LogOut,
  Search,
  Sun,
  Moon,
  Monitor,
  ArrowLeft,
  Webhook,
  SlidersHorizontal,
  Users,
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuthStore } from '@/stores/authStore'
import { useThemeStore } from '@/stores/themeStore'
import type { MenuProps } from 'antd'
import { useState, useEffect, useRef } from 'react'

const menuItems = [
  { key: '/dashboard', label: '仪表板', icon: LayoutDashboard },
  { key: '/cmdb', label: 'CMDB', icon: Server, hasSubNav: true, permission: { resource: 'ci', action: 'view' } },
  { key: '/tickets', label: '工单', icon: FileText, permission: { resource: 'ticket', action: 'view' } },
  { key: '/alerts', label: '告警', icon: Bell, hasSubNav: true, permission: { resource: 'alert', action: 'view' } },
  { key: '/reports', label: '报表', icon: BarChart3, permission: { resource: 'report', action: 'view' } },
  { key: '/admin', label: '系统', icon: Settings },
]

const cmdbSubMenuItems = [
  { key: '/cmdb', label: 'CMDB', icon: Server },
  { key: '/cmdb/victoriametrics', label: 'VictoriaMetrics配置', icon: Monitor, permission: { resource: 'monitoring', action: 'view' } },
]

const alertSubMenuItems = [
  { key: '/alerts', label: '告警', icon: Bell },
  { key: '/alerts/rules', label: '规则配置', icon: SlidersHorizontal, permission: { resource: 'alert_rule', action: 'view' } },
  { key: '/alerts/integration/webhook', label: 'Webhook', icon: Webhook, permission: { resource: 'webhook', action: 'view' } },
  { key: '/alerts/receivers', label: '告警接收', icon: Users, permission: { resource: 'alert_receiver', action: 'view' } },
]

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout, token, hasPermission } = useAuthStore()
  const { theme, setTheme } = useThemeStore()
  const [showSubNav, setShowSubNav] = useState(false)
  const [currentSubNav, setCurrentSubNav] = useState<'cmdb' | 'alerts' | null>(null)
  const userCollapsedSubNav = useRef(false)

  // 检查当前是否在子导航模块下
  useEffect(() => {
    const isInAlerts = location.pathname.startsWith('/alerts')
    const isInCMDB = location.pathname.startsWith('/cmdb')

    // 用户手动收起了子导航，且路径没有变化（仍在同模块内），不自动展开
    if (userCollapsedSubNav.current) {
      // 但如果路径变了（比如从 /cmdb/servers 跳到 /cmdb/tags），重新展开
      const prevModule = currentSubNav === 'cmdb' ? '/cmdb' : currentSubNav === 'alerts' ? '/alerts' : ''
      const newModule = isInCMDB ? '/cmdb' : isInAlerts ? '/alerts' : ''
      if (prevModule !== newModule) {
        userCollapsedSubNav.current = false
      } else {
        return
      }
    }

    if (isInAlerts) {
      setCurrentSubNav('alerts')
      setShowSubNav(true)
    } else if (isInCMDB) {
      setCurrentSubNav('cmdb')
      setShowSubNav(true)
    } else {
      setCurrentSubNav(null)
      setShowSubNav(false)
    }
  }, [location.pathname])

  const handleLogout = async () => {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      })
    } catch (error) {
      console.error('Logout failed:', error)
    }
    logout()
    navigate('/login')
  }

  // 处理 CMDB 菜单点击
  const handleCMDBClick = () => {
    userCollapsedSubNav.current = false
    if (location.pathname.startsWith('/cmdb')) {
      setShowSubNav(true)
      setCurrentSubNav('cmdb')
    } else {
      navigate('/cmdb')
    }
  }

  // 处理告警菜单点击
  const handleAlertClick = () => {
    userCollapsedSubNav.current = false
    if (location.pathname.startsWith('/alerts')) {
      setShowSubNav(true)
      setCurrentSubNav('alerts')
    } else {
      navigate('/alerts')
    }
  }

  // 返回主导航（只切换导航层级，不跳转页面）
  const handleBackToMain = () => {
    userCollapsedSubNav.current = true
    setShowSubNav(false)
    setCurrentSubNav(null)
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <User size={16} />,
      label: '个人中心',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogOut size={16} />,
      label: '退出登录',
      danger: true,
      onClick: handleLogout,
    },
  ]

  const themeMenuItems: MenuProps['items'] = [
    {
      key: 'light',
      icon: <Sun size={16} />,
      label: '浅色主题',
      onClick: () => setTheme('light'),
    },
    {
      key: 'dark',
      icon: <Moon size={16} />,
      label: '深色主题',
      onClick: () => setTheme('dark'),
    },
    {
      key: 'system',
      icon: <Monitor size={16} />,
      label: '跟随系统',
      onClick: () => setTheme('system'),
    },
  ]

  // 判断当前路径是否匹配菜单项
  const isActive = (path: string) => {
    if (path === '/dashboard') {
      return location.pathname === '/' || location.pathname === '/dashboard'
    }
    return location.pathname.startsWith(path)
  }

  // 判断子导航项是否激活
  const isSubNavActive = (path: string) => {
    if (path === '/alerts') {
      return location.pathname === '/alerts'
    }
    return location.pathname.startsWith(path)
  }

  // 获取当前主题图标
  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor

  return (
    <div className="min-h-screen bg-white dark:bg-bg-primary transition-colors">
      {/* 顶部导航栏 */}
      <header className="h-16 bg-white dark:bg-bg-secondary border-b border-gray-200 dark:border-white/8 sticky top-0 z-50 backdrop-blur-md">
        <div className="h-full px-6 flex items-center justify-between">
          {/* Logo + 菜单 */}
          <div className="flex items-center gap-8">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-primary to-brand-active flex items-center justify-center">
                <Server size={20} className="text-white" />
              </div>
              <span className="text-xl font-semibold text-gray-900 dark:text-text-primary">ITCMDB</span>
            </div>

            {/* 横向菜单 - 主导航和子导航切换 */}
            <nav className="relative h-10 min-w-[600px]">
              {/* 主导航 */}
              <AnimatePresence>
                {!showSubNav && (
                  <motion.div
                    key="main-nav"
                    initial={{ opacity: 0, x: -20, zIndex: 1 }}
                    animate={{ opacity: 1, x: 0, zIndex: 1 }}
                    exit={{ opacity: 0, x: -20, zIndex: 0 }}
                    transition={{ duration: 0.2, ease: [0.4, 0, 0.2, 1] }}
                    className="absolute inset-0 flex items-center gap-1"
                  >
                    {menuItems.filter((item) => !item.permission || hasPermission(item.permission.resource, item.permission.action)).map((item) => {
                      const Icon = item.icon
                      const active = isActive(item.key)

                      return (
                        <button
                          key={item.key}
                          onClick={() => {
                            if (item.hasSubNav) {
                              if (item.key === '/cmdb') {
                                handleCMDBClick()
                              } else if (item.key === '/alerts') {
                                handleAlertClick()
                              }
                            } else {
                              navigate(item.key)
                            }
                          }}
                          className={`
                            relative px-4 py-2 rounded-lg flex items-center gap-2
                            transition-all duration-200
                            ${active
                              ? 'text-brand-primary bg-brand-primary/10'
                              : 'text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary hover:bg-gray-100 dark:hover:bg-white/5'
                            }
                          `}
                        >
                          <Icon size={18} />
                          <span className="text-sm font-medium">{item.label}</span>

                          {/* 活动指示器 */}
                          {active && (
                            <motion.div
                              layoutId="activeTab"
                              className="absolute bottom-0 left-0 right-0 h-0.5 bg-brand-primary"
                              initial={false}
                              transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                            />
                          )}
                        </button>
                      )
                    })}
                  </motion.div>
                )}
              </AnimatePresence>

              {/* 子导航 - CMDB 或告警模块 */}
              <AnimatePresence>
                {showSubNav && (
                  <motion.div
                    key={`sub-nav-${currentSubNav}`}
                    initial={{ opacity: 0, x: 20, zIndex: 1 }}
                    animate={{ opacity: 1, x: 0, zIndex: 1 }}
                    exit={{ opacity: 0, x: 20, zIndex: 0 }}
                    transition={{ duration: 0.2, ease: [0.4, 0, 0.2, 1] }}
                    className="absolute inset-0 flex items-center gap-1"
                  >
                    {/* 返回按钮 */}
                    <button
                      onClick={handleBackToMain}
                      className="px-3 py-2 rounded-lg flex items-center gap-2 text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary hover:bg-gray-100 dark:hover:bg-white/5 transition-all duration-200"
                    >
                      <ArrowLeft size={18} />
                    </button>

                    {/* 子菜单项 */}
                    {(currentSubNav === 'cmdb' ? cmdbSubMenuItems : alertSubMenuItems).filter((item) => !item.permission || hasPermission(item.permission.resource, item.permission.action)).map((item) => {
                      const Icon = item.icon
                      const active = isSubNavActive(item.key)

                      return (
                        <button
                          key={item.key}
                          onClick={() => navigate(item.key)}
                          className={`
                            relative px-4 py-2 rounded-lg flex items-center gap-2
                            transition-all duration-200
                            ${active
                              ? 'text-brand-primary bg-brand-primary/10'
                              : 'text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary hover:bg-gray-100 dark:hover:bg-white/5'
                            }
                          `}
                        >
                          <Icon size={18} />
                          <span className="text-sm font-medium">{item.label}</span>

                          {/* 活动指示器 */}
                          {active && (
                            <motion.div
                              layoutId="activeSubTab"
                              className="absolute bottom-0 left-0 right-0 h-0.5 bg-brand-primary"
                              initial={false}
                              transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                            />
                          )}
                        </button>
                      )
                    })}
                  </motion.div>
                )}
              </AnimatePresence>
            </nav>
          </div>

          {/* 右侧工具栏 */}
          <div className="flex items-center gap-4">
            {/* 搜索框 */}
            <div className="relative">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-text-tertiary" />
              <Input
                placeholder="搜索... (Cmd+K)"
                className="w-64 h-9 pl-9 bg-gray-50 dark:bg-white/5 border-gray-200 dark:border-white/10 text-gray-900 dark:text-text-primary placeholder:text-gray-400 dark:placeholder:text-text-tertiary hover:bg-gray-100 dark:hover:bg-white/8 focus:bg-white dark:focus:bg-white/8"
              />
            </div>

            {/* 主题切换 */}
            <Dropdown menu={{ items: themeMenuItems, selectedKeys: [theme] }} placement="bottomRight" trigger={['click']}>
              <button className="p-2 rounded-lg text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary hover:bg-gray-100 dark:hover:bg-white/5 transition-colors">
                <ThemeIcon size={20} />
              </button>
            </Dropdown>

            {/* 通知 */}
            <button className="relative p-2 rounded-lg text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary hover:bg-gray-100 dark:hover:bg-white/5 transition-colors">
              <Badge count={3} size="small">
                <Bell size={20} />
              </Badge>
            </button>

            {/* 用户菜单 */}
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
              <button className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-white/5 transition-colors">
                <Avatar size={32} icon={<User size={16} />} className="bg-brand-primary" />
                <span className="text-sm text-gray-900 dark:text-text-primary">{user?.fullName || user?.username || '用户'}</span>
              </button>
            </Dropdown>
          </div>
        </div>
      </header>

      {/* 内容区域 */}
      <main className="min-h-[calc(100vh-64px)]">
        <Outlet />
      </main>
    </div>
  )
}
