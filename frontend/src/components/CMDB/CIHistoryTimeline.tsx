import { useEffect, useState } from 'react'
import { Timeline, Spin, Empty, Tag, Card } from 'antd'
import { ClockCircleOutlined, EditOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { useCMDBStore, CIHistory } from '@/stores/cmdbStore'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

interface CIHistoryTimelineProps {
  ciId: number
  limit?: number
}

export default function CIHistoryTimeline({ ciId, limit = 50 }: CIHistoryTimelineProps) {
  const { fetchHistory } = useCMDBStore()
  const [history, setHistory] = useState<CIHistory[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadHistory()
  }, [ciId])

  const loadHistory = async () => {
    setLoading(true)
    try {
      const data = await fetchHistory(ciId, limit)
      setHistory(data)
    } catch (error) {
      console.error('Failed to load history:', error)
    } finally {
      setLoading(false)
    }
  }

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'create':
        return <PlusOutlined style={{ color: '#52c41a' }} />
      case 'update':
        return <EditOutlined style={{ color: '#1890ff' }} />
      case 'delete':
        return <DeleteOutlined style={{ color: '#ff4d4f' }} />
      default:
        return <ClockCircleOutlined />
    }
  }

  const getActionColor = (action: string) => {
    switch (action) {
      case 'create':
        return 'green'
      case 'update':
        return 'blue'
      case 'delete':
        return 'red'
      default:
        return 'default'
    }
  }

  const getActionText = (action: string) => {
    switch (action) {
      case 'create':
        return '创建'
      case 'update':
        return '更新'
      case 'delete':
        return '删除'
      default:
        return action
    }
  }

  const renderHistoryItem = (item: CIHistory) => {
    const hasFieldChange = item.field_name && (item.old_value || item.new_value)

    return (
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Tag color={getActionColor(item.action)}>{getActionText(item.action)}</Tag>
          <span className="text-gray-500 text-sm">
            {dayjs(item.changed_at).format('YYYY-MM-DD HH:mm:ss')}
          </span>
          <span className="text-gray-400 text-xs">
            ({dayjs(item.changed_at).fromNow()})
          </span>
        </div>

        {hasFieldChange && (
          <Card size="small" className="mt-2 bg-gray-50 dark:bg-gray-800">
            <div className="space-y-1">
              <div className="text-sm font-medium text-gray-700 dark:text-gray-300">
                字段: {item.field_name}
              </div>
              {item.old_value && (
                <div className="text-sm">
                  <span className="text-gray-500">旧值: </span>
                  <span className="text-red-600 dark:text-red-400 line-through">
                    {item.old_value}
                  </span>
                </div>
              )}
              {item.new_value && (
                <div className="text-sm">
                  <span className="text-gray-500">新值: </span>
                  <span className="text-green-600 dark:text-green-400 font-medium">
                    {item.new_value}
                  </span>
                </div>
              )}
            </div>
          </Card>
        )}
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center py-8">
        <Spin tip="加载历史记录..." />
      </div>
    )
  }

  if (!history || history.length === 0) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="暂无变更历史"
        className="py-8"
      />
    )
  }

  return (
    <div className="py-4">
      <Timeline
        items={history.map((item) => ({
          dot: getActionIcon(item.action),
          children: renderHistoryItem(item),
        }))}
      />
    </div>
  )
}
