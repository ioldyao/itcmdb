import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  Server,
  Network,
  Package,
  Box,
  Shield,
  Tags as TagsIcon,
} from 'lucide-react'
import { useState } from 'react'

const { Sider, Content } = Layout

// 定义菜单项及其所需权限
const allMenuItems = [
  { key: '/cmdb/servers', label: '服务器', icon: <Server size={16} />, permission: null },
  { key: '/cmdb/networks', label: '网络设备', icon: <Network size={16} />, permission: null },
  { key: '/cmdb/applications', label: '应用系统', icon: <Package size={16} />, permission: null },
  { key: '/cmdb/containers', label: '容器实例', icon: <Box size={16} />, permission: null },
  { key: '/cmdb/roles', label: '角色管理', icon: <Shield size={16} />, permission: null },
  { key: '/cmdb/tags', label: '标签管理', icon: <TagsIcon size={16} />, permission: null },
]

export default function CMDBLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)

  // CMDB菜单不需要权限过滤
  const menuItems = allMenuItems

  // 获取当前选中的菜单项
  const selectedKey = menuItems.find(item => location.pathname.startsWith(item.key))?.key || '/cmdb/servers'

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
