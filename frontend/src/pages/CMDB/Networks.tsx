import { Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

export default function CMDBNetworks() {
  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>网络设备管理</h2>
        <Button type="primary" icon={<PlusOutlined />}>添加设备</Button>
      </div>
      <Table columns={[]} dataSource={[]} rowKey="id" />
    </div>
  )
}
