import { useEffect, useState } from 'react'
import { Table, Tag, Card, DatePicker, Select, Button, Empty } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useAlertStore } from '@/stores/alertStore'
import dayjs from 'dayjs'

interface HistoryRecord {
  id: number
  alertId: number
  alertTitle: string
  severity: string
  action: string
  oldStatus: string
  newStatus: string
  operatedBy: string
  operatedAt: string
  notes?: string
}

const { RangePicker } = DatePicker

const actionMap: Record<string, { text: string; color: string }> = {
  triggered: { text: '触发', color: 'red' },
  acknowledged: { text: '确认', color: 'orange' },
  resolved: { text: '恢复', color: 'green' },
  closed: { text: '关闭', color: 'default' },
  updated: { text: '更新', color: 'blue' },
}

const severityMap: Record<string, { text: string; color: string }> = {
  critical: { text: '致命', color: 'red' },
  high: { text: '高', color: 'orange' },
  medium: { text: '中', color: 'gold' },
  low: { text: '低', color: 'blue' },
}

const statusMap: Record<string, { text: string; color: string }> = {
  firing: { text: '未恢复', color: 'red' },
  acknowledged: { text: '已确认', color: 'orange' },
  resolved: { text: '已恢复', color: 'green' },
  closed: { text: '已关闭', color: 'default' },
}

export default function AlertHistory() {
  const { fetchAlerts } = useAlertStore()
  const [loading, setLoading] = useState(false)
  const [historyRecords, setHistoryRecords] = useState<HistoryRecord[]>([])
  const [actionFilter, setActionFilter] = useState<string>('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null]>([null, null])

  useEffect(() => {
    loadHistory()
  }, [])

  const loadHistory = async () => {
    setLoading(true)
    try {
      // Fetch all alerts to build history from their state changes
      await fetchAlerts({ page: 1, page_size: 100 })

      // Build history records from alerts
      const records: HistoryRecord[] = []
      const currentAlerts = useAlertStore.getState().alerts

      currentAlerts.forEach((alert) => {
        // Add the initial trigger record
        records.push({
          id: alert.id * 100 + 1,
          alertId: alert.id,
          alertTitle: alert.title,
          severity: alert.severity,
          action: 'triggered',
          oldStatus: '',
          newStatus: 'firing',
          operatedBy: '系统',
          operatedAt: alert.first_triggered,
        })

        // Add acknowledged record if applicable
        if (alert.acknowledged_at) {
          records.push({
            id: alert.id * 100 + 2,
            alertId: alert.id,
            alertTitle: alert.title,
            severity: alert.severity,
            action: 'acknowledged',
            oldStatus: 'firing',
            newStatus: 'acknowledged',
            operatedBy: alert.handler ? String(alert.handler) : '未知',
            operatedAt: alert.acknowledged_at,
          })
        }

        // Add resolved record if applicable
        if (alert.recovered_at) {
          records.push({
            id: alert.id * 100 + 3,
            alertId: alert.id,
            alertTitle: alert.title,
            severity: alert.severity,
            action: 'resolved',
            oldStatus: 'acknowledged',
            newStatus: 'resolved',
            operatedBy: '系统',
            operatedAt: alert.recovered_at,
          })
        }

        // Add closed record if applicable
        if (alert.closed_at) {
          records.push({
            id: alert.id * 100 + 4,
            alertId: alert.id,
            alertTitle: alert.title,
            severity: alert.severity,
            action: 'closed',
            oldStatus: alert.status === 'closed' ? 'acknowledged' : alert.status,
            newStatus: 'closed',
            operatedBy: alert.handler ? String(alert.handler) : '未知',
            operatedAt: alert.closed_at,
          })
        }
      })

      // Sort by operatedAt descending
      records.sort((a, b) => new Date(b.operatedAt).getTime() - new Date(a.operatedAt).getTime())
      setHistoryRecords(records)
    } catch (error) {
      console.error('Failed to load history:', error)
    } finally {
      setLoading(false)
    }
  }

  // Filter records
  const filteredRecords = historyRecords.filter((record) => {
    if (actionFilter && record.action !== actionFilter) return false
    if (dateRange[0] && dateRange[1]) {
      const recordTime = dayjs(record.operatedAt)
      if (recordTime.isBefore(dateRange[0]) || recordTime.isAfter(dateRange[1])) return false
    }
    return true
  })

  const columns: ColumnsType<HistoryRecord> = [
    {
      title: '时间',
      dataIndex: 'operatedAt',
      width: 180,
      render: (time: string) => (
        <span className="text-sm">{dayjs(time).format('YYYY-MM-DD HH:mm:ss')}</span>
      ),
    },
    {
      title: '告警',
      dataIndex: 'alertTitle',
      ellipsis: true,
      render: (title: string, record) => (
        <div>
          <span className="text-brand-primary cursor-pointer hover:underline">#{record.alertId}</span>
          <span className="ml-2 text-gray-900 dark:text-text-primary">{title}</span>
        </div>
      ),
    },
    {
      title: '级别',
      dataIndex: 'severity',
      width: 80,
      render: (severity: string) => {
        const s = severityMap[severity] || { text: severity, color: 'default' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 100,
      render: (action: string) => {
        const a = actionMap[action] || { text: action, color: 'default' }
        return <Tag color={a.color}>{a.text}</Tag>
      },
    },
    {
      title: '状态变更',
      width: 150,
      render: (_, record) => (
        <span className="text-sm">
          {record.oldStatus && (
            <>
              <Tag color={statusMap[record.oldStatus]?.color}>{statusMap[record.oldStatus]?.text || record.oldStatus}</Tag>
              <span className="mx-1">→</span>
            </>
          )}
          <Tag color={statusMap[record.newStatus]?.color}>{statusMap[record.newStatus]?.text || record.newStatus}</Tag>
        </span>
      ),
    },
    {
      title: '操作人',
      dataIndex: 'operatedBy',
      width: 100,
      render: (user: string) => <span className="text-sm">{user}</span>,
    },
  ]

  return (
    <div className="p-6">
      <Card className="dark:bg-bg-secondary dark:border-white/8">
        {/* 页面标题 */}
        <div className="mb-6">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">告警历史</h2>
          <p className="text-gray-600 dark:text-text-secondary text-sm">查看所有告警的状态变更记录</p>
        </div>

        {/* 筛选栏 */}
        <div className="mb-4 flex flex-wrap gap-4 items-center">
          <Select
            placeholder="操作类型"
            value={actionFilter}
            onChange={setActionFilter}
            className="w-32"
            allowClear
          >
            <Select.Option value="">全部</Select.Option>
            <Select.Option value="triggered">触发</Select.Option>
            <Select.Option value="acknowledged">确认</Select.Option>
            <Select.Option value="resolved">恢复</Select.Option>
            <Select.Option value="closed">关闭</Select.Option>
          </Select>
          <RangePicker
            value={dateRange as any}
            onChange={(dates) => setDateRange(dates as any)}
            showTime
            placeholder={['开始时间', '结束时间']}
          />
          <Button icon={<ReloadOutlined />} onClick={loadHistory} loading={loading}>
            刷新
          </Button>
        </div>

        {/* 表格 */}
        <Table
          columns={columns}
          dataSource={filteredRecords}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          locale={{
            emptyText: <Empty description="暂无历史记录" />,
          }}
        />
      </Card>
    </div>
  )
}
