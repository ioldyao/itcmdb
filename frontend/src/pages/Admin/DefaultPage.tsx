import { Card, Row, Col } from 'antd'
import { Users, Shield, FileText } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

export default function AdminDefaultPage() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  const adminSections = [
    {
      key: 'users',
      title: '用户管理',
      description: '管理系统用户、查看用户信息',
      icon: <Users size={32} />,
      path: '/admin/users',
      permission: { resource: 'user', action: 'view' }
    },
    {
      key: 'roles',
      title: '角色管理',
      description: '管理角色和权限配置',
      icon: <Shield size={32} />,
      path: '/admin/roles',
      permission: { resource: 'role', action: 'view' }
    },
    {
      key: 'audit',
      title: '审计日志',
      description: '查看系统操作审计记录',
      icon: <FileText size={32} />,
      path: '/admin/audit',
      permission: null // 所有用户都能访问
    },
  ]

  // 过滤出用户有权限访问的部分
  const availableSections = adminSections.filter(section => {
    if (!section.permission) return true
    return hasPermission(section.permission.resource, section.permission.action)
  })

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>系统管理</h2>
      <Row gutter={[16, 16]}>
        {availableSections.map(section => (
          <Col key={section.key} xs={24} sm={12} md={8}>
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
      {availableSections.length === 0 && (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <p style={{ color: '#999' }}>您没有权限访问任何系统管理功能</p>
          </div>
        </Card>
      )}
    </div>
  )
}
