import { Table, Button, Input, Select, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'

interface Ticket {
  id: number
  title: string
  status: string
  priority: string
  createdAt: string
}

export default function TicketList() {
  const columns: ColumnsType<Ticket> = []

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>工单列表</h2>
        <Button type="primary" icon={<PlusOutlined />}>创建工单</Button>
      </div>
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="搜索工单..." prefix={<SearchOutlined />} style={{ width: 200 }} />
        <Select placeholder="状态" style={{ width: 120 }} />
        <Select placeholder="优先级" style={{ width: 120 }} />
      </Space>
      <Table columns={columns} dataSource={[]} rowKey="id" />
    </div>
  )
}
