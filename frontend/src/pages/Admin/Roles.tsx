import { Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

export default function AdminRoles() {
  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>角色权限管理</h2>
        <Button type="primary" icon={<PlusOutlined />}>添加角色</Button>
      </div>
      <Table columns={[]} dataSource={[]} rowKey="id" />
    </div>
  )
}
