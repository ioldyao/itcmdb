import { Card, Row, Col } from 'antd'
import { Server, Network, Package, Box, Shield, Tags as TagsIcon } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export default function CMDBDefaultPage() {
  const navigate = useNavigate()

  const cmdbSections = [
    {
      key: 'servers',
      title: '服务器',
      description: '管理物理服务器和虚拟机',
      icon: <Server size={32} />,
      path: '/cmdb/servers',
    },
    {
      key: 'networks',
      title: '网络设备',
      description: '管理交换机、路由器等网络设备',
      icon: <Network size={32} />,
      path: '/cmdb/networks',
    },
    {
      key: 'applications',
      title: '应用系统',
      description: '管理应用程序和服务',
      icon: <Package size={32} />,
      path: '/cmdb/applications',
    },
    {
      key: 'containers',
      title: '容器实例',
      description: '管理 Docker 容器和 Kubernetes Pod',
      icon: <Box size={32} />,
      path: '/cmdb/containers',
    },
    {
      key: 'roles',
      title: '角色管理',
      description: '管理 CI 角色和分类',
      icon: <Shield size={32} />,
      path: '/cmdb/roles',
    },
    {
      key: 'tags',
      title: '标签管理',
      description: '管理资源标签和分组',
      icon: <TagsIcon size={32} />,
      path: '/cmdb/tags',
    },
  ]

  return (
    <div>
      <h2 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-6">配置管理数据库 (CMDB)</h2>
      <p className="text-gray-600 dark:text-text-secondary mb-6">
        选择要管理的资源类型
      </p>
      <Row gutter={[16, 16]}>
        {cmdbSections.map(section => (
          <Col key={section.key} xs={24} sm={12} md={8} lg={6}>
            <Card
              hoverable
              onClick={() => navigate(section.path)}
              className="h-full dark:bg-bg-secondary dark:border-white/8"
            >
              <div className="text-center">
                <div className="mb-4 text-brand-primary">
                  {section.icon}
                </div>
                <h3 className="mb-2 text-gray-900 dark:text-text-primary">{section.title}</h3>
                <p className="text-sm text-gray-600 dark:text-text-secondary">{section.description}</p>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  )
}
