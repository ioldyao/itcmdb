import { Card, Row, Col } from 'antd'
import { Server, Network, Package, Box, Shield, Tags as TagsIcon, Activity } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

export default function CMDBDefaultPage() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  const cmdbSections = [
    {
      key: 'servers',
      title: '服务器',
      description: '管理物理服务器和虚拟机',
      icon: <Server size={32} />,
      path: '/cmdb/servers',
      permission: null
    },
    {
      key: 'networks',
      title: '网络设备',
      description: '管理交换机、路由器等网络设备',
      icon: <Network size={32} />,
      path: '/cmdb/networks',
      permission: null
    },
    {
      key: 'applications',
      title: '应用系统',
      description: '管理应用程序和服务',
      icon: <Package size={32} />,
      path: '/cmdb/applications',
      permission: null
    },
    {
      key: 'containers',
      title: '容器实例',
      description: '管理 Docker 容器和 Kubernetes Pod',
      icon: <Box size={32} />,
      path: '/cmdb/containers',
      permission: null
    },
    {
      key: 'roles',
      title: '角色管理',
      description: '管理 CI 角色和分类',
      icon: <Shield size={32} />,
      path: '/cmdb/roles',
      permission: null
    },
    {
      key: 'tags',
      title: '标签管理',
      description: '管理资源标签和分组',
      icon: <TagsIcon size={32} />,
      path: '/cmdb/tags',
      permission: null
    },
    {
      key: 'victoriametrics',
      title: 'VictoriaMetrics',
      description: '监控指标配置和管理',
      icon: <Activity size={32} />,
      path: '/cmdb/victoriametrics',
      permission: { resource: 'config', action: 'view' }
    },
  ]

  // 过滤出用户有权限访问的部分
  const availableSections = cmdbSections.filter(section => {
    if (!section.permission) return true
    return hasPermission(section.permission.resource, section.permission.action)
  })

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>配置管理数据库 (CMDB)</h2>
      <p style={{ marginBottom: 24, color: '#666' }}>
        选择要管理的资源类型
      </p>
      <Row gutter={[16, 16]}>
        {availableSections.map(section => (
          <Col key={section.key} xs={24} sm={12} md={8} lg={6}>
            <Card
              hoverable
              onClick={() => navigate(section.path)}
              style={{ height: '100%' }}
            >
              <div style={{ textAlign: 'center' }}>
                <div style={{ marginBottom: 16, color: '#1890ff' }}>
                  {section.icon}
                </div>
                <h3 style={{ marginBottom: 8 }}>{section.title}</h3>
                <p style={{ color: '#666', fontSize: 14 }}>{section.description}</p>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  )
}
