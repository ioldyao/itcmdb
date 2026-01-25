import { useEffect, useState } from 'react'
import { Timeline, Spin, Empty, Tag, Card, Collapse } from 'antd'
import { ClockCircleOutlined, EditOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { useCMDBStore, CIHistory } from '@/stores/cmdbStore'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

// 忽略的字段和子字段配置（与后端保持一致）
const IGNORED_FIELDS = new Set(['last_hardware_report'])
const IGNORED_SUB_FIELDS = {
  optical_modules_info: new Set(['temperature']),
}

// JSON diff 比较函数
function compareJSON(oldJSON: string, newJSON: string): Array<{ path: string; oldValue: any; newValue: any }> {
  try {
    const oldObj = JSON.parse(oldJSON)
    const newObj = JSON.parse(newJSON)

    const changes: Array<{ path: string; oldValue: any; newValue: any }> = []

    // 递归比较对象
    function compare(prefix: string, oldVal: any, newVal: any) {
      // 检查当前路径是否应该被忽略
      const currentPath = prefix || ''
      if (IGNORED_FIELDS.has(currentPath)) {
        return // 跳过被忽略的字段
      }

      // 如果类型不同，记录变化
      if (typeof oldVal !== typeof newVal) {
        changes.push({ path: prefix, oldValue: oldVal, newValue: newVal })
        return
      }

      // 如果是对象，递归比较
      if (typeof oldVal === 'object' && oldVal !== null && newVal !== null) {
        const oldKeys = Object.keys(oldVal)
        const newKeys = Object.keys(newVal)

        // 检查新增的键
        for (const key of newKeys) {
          if (!(key in oldVal)) {
            changes.push({ path: prefix ? `${prefix}.${key}` : key, oldValue: undefined, newValue: newVal[key] })
          }
        }

        // 检查删除的键
        for (const key of oldKeys) {
          if (!(key in newVal)) {
            changes.push({ path: prefix ? `${prefix}.${key}` : key, oldValue: oldVal[key], newValue: undefined })
          }
        }

        // 检查修改的键
        for (const key of oldKeys) {
          if (key in newVal) {
            const fullPath = prefix ? `${prefix}.${key}` : key

            // 检查是否需要忽略该字段的子字段
            if (IGNORED_SUB_FIELDS[fullPath]) {
              // 如果是数组，比较时排除指定的子字段
              if (Array.isArray(oldVal[key]) && Array.isArray(newVal[key])) {
                const oldArr = oldVal[key] as any[]
                const newArr = newVal[key] as any[]

                // 比较数组长度
                if (oldArr.length !== newArr.length) {
                  changes.push({ path: fullPath, oldValue: oldVal[key], newValue: newVal[key] })
                } else {
                  // 逐个比较元素，排除指定的子字段
                  for (let i = 0; i < oldArr.length; i++) {
                    const oldItem = oldArr[i]
                    const newItem = newArr[i]
                    let hasDifference = false

                    // 比较除了被忽略子字段外的所有字段
                    for (const itemKey of Object.keys({ ...oldItem, ...newItem })) {
                      if (IGNORED_SUB_FIELDS[fullPath].has(itemKey)) {
                        continue // 跳过被忽略的子字段
                      }

                      const oldJSON = JSON.stringify(oldItem[itemKey])
                      const newJSON = JSON.stringify(newItem[itemKey])
                      if (oldJSON !== newJSON) {
                        hasDifference = true
                        break
                      }
                    }

                    if (hasDifference) {
                      changes.push({ path: `${fullPath}[${i}]`, oldValue: oldItem, newValue: newItem })
                    }
                  }
                }
              } else {
                // 不是数组，继续递归比较
                compare(fullPath, oldVal[key], newVal[key])
              }
            } else {
              const oldJSON = JSON.stringify(oldVal[key])
              const newJSON = JSON.stringify(newVal[key])
              if (oldJSON !== newJSON) {
                compare(fullPath, oldVal[key], newVal[key])
              }
            }
          }
        }
      } else if (oldVal !== newVal) {
        // 基本类型值不同
        changes.push({ path: prefix, oldValue: oldVal, newValue: newVal })
      }
    }

    compare('', oldObj, newObj)
    return changes
  } catch (error) {
    console.error('JSON comparison failed:', error)
    return []
  }
}

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

    // 检查是否是 attributes 字段且有 JSON 数据
    const isAttributesWithJSON = item.field_name === 'attributes' &&
      item.old_value &&
      item.new_value &&
      item.old_value !== '-' &&
      item.new_value !== '-' &&
      (item.old_value.startsWith('{') || item.old_value.startsWith('['))

    // 解析变化
    let changes: Array<{ path: string; oldValue: any; newValue: any }> = []
    if (isAttributesWithJSON) {
      try {
        changes = compareJSON(item.old_value, item.new_value)
      } catch (e) {
        console.error('Failed to parse JSON:', e)
      }
    }

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

              {/* attributes 字段的 JSON diff 显示 */}
              {isAttributesWithJSON ? (
                <div>
                  <div className="text-xs text-gray-500 mb-2">
                    共 {changes.length} 处变更
                  </div>
                  <Collapse
                    size="small"
                    items={[
                      {
                        key: '1',
                        label: '查看详细变更',
                        children: (
                          <div className="space-y-2 max-h-96 overflow-y-auto">
                            {changes.map((change, idx) => (
                              <div key={idx} className="text-sm border-l-2 border-blue-500 pl-2 py-1">
                                <div className="font-medium text-gray-700 dark:text-gray-300 mb-1">
                                  {change.path}
                                </div>
                                <div className="flex gap-4">
                                  {change.oldValue !== undefined && (
                                    <div className="flex-1">
                                      <span className="text-xs text-gray-500">旧值: </span>
                                      <span className="text-red-600 dark:text-red-400 line-through break-all">
                                        {typeof change.oldValue === 'object'
                                          ? JSON.stringify(change.oldValue, null, 2)
                                          : String(change.oldValue)}
                                      </span>
                                    </div>
                                  )}
                                  {change.newValue !== undefined && (
                                    <div className="flex-1">
                                      <span className="text-xs text-gray-500">新值: </span>
                                      <span className="text-green-600 dark:text-green-400 break-all">
                                        {typeof change.newValue === 'object'
                                          ? JSON.stringify(change.newValue, null, 2)
                                          : String(change.newValue)}
                                      </span>
                                    </div>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        ),
                      },
                    ]}
                  />
                </div>
              ) : (
                <>
                  {/* 其他字段的正常显示 */}
                  {item.old_value && item.old_value !== '-' && (
                    <div className="text-sm">
                      <span className="text-gray-500">旧值: </span>
                      <span className="text-red-600 dark:text-red-400 line-through">
                        {item.old_value}
                      </span>
                    </div>
                  )}
                  {item.new_value && item.new_value !== '-' && (
                    <div className="text-sm">
                      <span className="text-gray-500">新值: </span>
                      <span className="text-green-600 dark:text-green-400 font-medium">
                        {item.new_value}
                      </span>
                    </div>
                  )}
                </>
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
