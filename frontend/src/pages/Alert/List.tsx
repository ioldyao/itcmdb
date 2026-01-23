import { Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

interface Alert {
  id: number
  title: string
  severity: string
  status: string
  triggeredAt: string
}

export default function AlertList() {
  const columns: ColumnsType<Alert> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '告警标题', dataIndex: 'title' },
    {
      title: '严重级别',
      dataIndex: 'severity',
      render: (severity: string) => {
        const colors: Record<string, string> = {
          critical: 'red',
          high: 'orange',
          medium: 'gold',
          low: 'blue',
        }
        return <Tag color={colors[severity]}>{severity.toUpperCase()}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => {
        const map: Record<string, { text: string; color: string }> = {
          active: { text: '活跃', color: 'red' },
          acknowledged: { text: '已确认', color: 'orange' },
          closed: { text: '已关闭', color: 'default' },
        }
        const s = map[status] || { text: status, color: 'default' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    { title: '触发时间', dataIndex: 'triggeredAt' },
    {
      title: '操作',
      render: () => (
        <>
          <a style={{ marginRight: 8 }}>确认</a>
          <a>关闭</a>
        </>
      ),
    },
  ]

  return (
    <div>
      <h2>告警列表</h2>
      <Table columns={columns} dataSource={[]} rowKey="id" />
    </div>
  )
}
