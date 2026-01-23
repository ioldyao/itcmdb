import { Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

export default function CMDBContainers() {
  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>容器/K8s管理</h2>
        <Button type="primary" icon={<PlusOutlined />}>添加集群</Button>
      </div>
      <Table columns={[]} dataSource={[]} rowKey="id" />
    </div>
  )
}
