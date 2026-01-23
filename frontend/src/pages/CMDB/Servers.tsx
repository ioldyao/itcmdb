import { Table, Button, Space } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useCMDBStore } from '@/stores/cmdbStore'

interface Server {
  id: number
  name: string
  ip: string
  os: string
  status: string
  cpu: string
  memory: string
}

export default function CMDBServers() {
  const { instances, loading } = useCMDBStore()

  const columns: ColumnsType<Server> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '名称', dataIndex: 'name' },
    { title: 'IP地址', dataIndex: 'ip' },
    { title: '操作系统', dataIndex: 'os' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => (
        <span style={{ color: status === 'online' ? 'green' : 'red' }}>
          {status === 'online' ? '在线' : '离线'}
        </span>
      ),
    },
    { title: 'CPU', dataIndex: 'cpu' },
    { title: '内存', dataIndex: 'memory' },
    {
      title: '操作',
      render: () => (
        <Space>
          <Button type="link" size="small">查看</Button>
          <Button type="link" size="small">编辑</Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>服务器管理</h2>
        <Button type="primary" icon={<PlusOutlined />}>添加服务器</Button>
      </div>
      <Table columns={columns} dataSource={[]} loading={loading} rowKey="id" />
    </div>
  )
}
