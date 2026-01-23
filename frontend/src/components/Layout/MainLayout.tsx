import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Dropdown, Button } from 'antd'
import {
  DashboardOutlined,
  CloudServerOutlined,
  CustomerServiceOutlined,
  AlertOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'

const { Header, Sider, Content } = Layout

const menuItems: MenuProps['items'] = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '仪表板',
  },
  {
    key: '/cmdb',
    icon: <CloudServerOutlined />,
    label: 'CMDB',
    children: [
      { key: '/cmdb/servers', label: '服务器' },
      { key: '/cmdb/networks', label: '网络设备' },
      { key: '/cmdb/applications', label: '应用服务' },
      { key: '/cmdb/containers', label: '容器/K8s' },
    ],
  },
  {
    key: '/tickets',
    icon: <CustomerServiceOutlined />,
    label: '工单管理',
  },
  {
    key: '/alerts',
    icon: <AlertOutlined />,
    label: '告警管理',
  },
  {
    key: '/admin',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: '/admin/users', label: '用户管理' },
      { key: '/admin/roles', label: '角色权限' },
      { key: '/admin/audit', label: '审计日志' },
    ],
  },
]

const userMenuItems: MenuProps['items'] = [
  {
    key: 'profile',
    icon: <UserOutlined />,
    label: '个人信息',
  },
  {
    type: 'divider',
  },
  {
    key: 'logout',
    icon: <LogoutOutlined />,
    label: '退出登录',
    danger: true,
  },
]

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleUserMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key === 'logout') {
      // TODO: Implement logout
      navigate('/login')
    } else if (key === 'profile') {
      // TODO: Navigate to profile page
    }
  }

  const selectedKeys = [location.pathname]
  const openKey = location.pathname.split('/')[1]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" width={240}>
        <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontSize: 20, fontWeight: 'bold' }}>
          ITCMDB
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={[openKey]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #f0f0f0' }}>
          <div />
          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight">
            <Button type="text" icon={<Avatar size="small" icon={<UserOutlined />} />}>
              管理员
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ padding: 24 }}>
          <div style={{ background: '#fff', padding: 24, borderRadius: 8, minHeight: 'calc(100vh - 112px)' }}>
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
