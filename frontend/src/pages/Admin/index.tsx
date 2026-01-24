import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  Users,
  Shield,
  FileText,
} from 'lucide-react'
import { useState } from 'react'

const { Sider, Content } = Layout

const menuItems = [
  { key: '/admin/users', label: '用户管理', icon: <Users size={16} /> },
  { key: '/admin/roles', label: '角色管理', icon: <Shield size={16} /> },
  { key: '/admin/audit', label: '审计日志', icon: <FileText size={16} /> },
]

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)

  // 获取当前选中的菜单项
  const selectedKey = menuItems.find(item => location.pathname.startsWith(item.key))?.key || '/admin/users'

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
