import { Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

export default function CMDBApplications() {
  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>应用服务管理</h2>
        <Button type="primary" icon={<PlusOutlined />}>添加应用</Button>
      </div>
      <Table columns={[]} dataSource={[]} rowKey="id" />
    </div>
  )
}
