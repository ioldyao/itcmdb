import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  Users,
  Shield,
  FileText,
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
    permission: null  // 所有用户都能看到，但内容根据权限过滤
  },
]

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)
  const hasPermission = useAuthStore((state) => state.hasPermission)

  // 根据权限过滤菜单项
  const menuItems = useMemo(() => {
    return allMenuItems.filter(item => {
      // 如果菜单项不需要权限，直接显示
      if (!item.permission) return true
      // 检查用户是否有所需权限
      return hasPermission(item.permission.resource, item.permission.action)
    })
  }, [hasPermission])

  // 获取当前选中的菜单项
  const selectedKey = menuItems.find(item => location.pathname.startsWith(item.key))?.key || menuItems[0]?.key

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
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            inlineIndent={12}
            style={{ border: 'none' }}
          />
        </div>
      </Sider>
      <Content style={{ padding: '24px', background: '#fff' }}>
        <Outlet />
      </Content>
    </Layout>
  )
}
