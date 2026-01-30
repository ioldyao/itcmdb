import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom'
import { Layout, Menu, Breadcrumb } from 'antd'
import {
  Users,
  Shield,
  FileText,
  Database,
} from 'lucide-react'
import { useState, useMemo } from 'react'
import { useAuthStore } from '@/stores/authStore'

const { Sider, Content } = Layout

// 定义菜单项及其所需权限
const allMenuItems = [
  {
    key: '/admin/users',
    label: '用户管理',
    icon: <Users size={16} />,
    permission: { resource: 'user', action: 'view' }
  },
  {
    key: '/admin/roles',
    label: '角色管理',
    icon: <Shield size={16} />,
    permission: { resource: 'role', action: 'view' }
  },
  {
    key: '/admin/audit',
    label: '审计日志',
    icon: <FileText size={16} />,
    permission: null
  },
]

// 面包屑配置
const breadcrumbMap: Record<string, { title: string; icon?: React.ReactNode; path?: string }> = {
  '/admin': { title: '系统管理', icon: <Database size={14} /> },
  '/admin/users': { title: '用户管理', icon: <Users size={14} /> },
  '/admin/roles': { title: '角色管理', icon: <Shield size={14} /> },
  '/admin/audit': { title: '审计日志', icon: <FileText size={14} /> },
}

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)
  const hasPermission = useAuthStore((state) => state.hasPermission)

  // 根据权限过滤菜单项
  const menuItems = useMemo(() => {
    const filteredItems: any[] = []

    allMenuItems.forEach(item => {
      // 如果菜单项不需要权限，直接显示
      if (!item.permission) {
        filteredItems.push(item)
        return
      }
      // 检查用户是否有所需权限
      if (hasPermission(item.permission.resource, item.permission.action)) {
        filteredItems.push(item)
      }
    })

    return filteredItems
  }, [hasPermission])

  // 展平菜单项用于查找选中项
  const flatMenuItems = useMemo(() => {
    return allMenuItems
  }, [])

  // 获取当前选中的菜单项
  const selectedKey = useMemo(() => {
    const currentPath = location.pathname
    // 精确匹配
    let found = flatMenuItems.find(item => item.key === currentPath)
    if (found) return found.key

    // 模糊匹配（对于子路由）
    found = flatMenuItems.find(item => currentPath.startsWith(item.key + '/'))
    return found?.key || flatMenuItems[0]?.key
  }, [location.pathname, flatMenuItems])

  // 生成面包屑
  const breadcrumbItems = useMemo(() => {
    const path = location.pathname
    const segments = path.split('/').filter(Boolean)

    const items = [
      {
        title: <Link to="/dashboard">首页</Link>
      }
    ]

    // 累积路径
    let accumulatedPath = ''
    segments.forEach((segment, index) => {
      accumulatedPath += '/' + segment
      const config = breadcrumbMap[accumulatedPath]

      if (config) {
        const isLast = index === segments.length - 1
        items.push({
          title: isLast
            ? <span>{config.title}</span>
            : <Link to={accumulatedPath}>{config.title}</Link>
        })
      }
    })

    return items
  }, [location.pathname])

  // 菜单点击处理
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  return (
    <Layout style={{ minHeight: 'calc(100vh - 64px)' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        width={220}
        style={{
          background: '#fff',
          borderRight: '1px solid rgba(0,0,0,0.06)',
        }}
        trigger={null}
      >
        <div style={{ padding: '16px 8px' }}>
          <Menu
            mode="inline"
            selectedKeys={[selectedKey]}
            onClick={handleMenuClick}
            style={{ border: 'none' }}
            items={menuItems}
          />
        </div>
      </Sider>
      <Content style={{ padding: '24px', background: '#fff' }}>
        {/* 面包屑 */}
        <Breadcrumb style={{ marginBottom: 16 }} items={breadcrumbItems} />

        {/* 内容区域 */}
        <Outlet />
      </Content>
    </Layout>
  )
}
