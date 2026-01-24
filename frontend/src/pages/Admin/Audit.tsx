import { useState, useEffect } from 'react'
import { Table, Card, Form, Select, DatePicker, Input, Button, Space, Tag, Statistic, Row, Col } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import dayjs from 'dayjs'

interface AuditLog {
  id: number
  user_id?: number
  action: string
  resource: string
  resource_id?: number
  details: Record<string, any>
  ip_address: string
  user_agent: string
  status: string
  error_msg?: string
  created_at: string
}

interface AuditStats {
  total_logs: number
  by_action: Record<string, number>
  by_resource: Record<string, number>
}

interface PaginationParams {
  current: number
  pageSize: number
}

export default function AdminAudit() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [stats, setStats] = useState<AuditStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [pagination, setPagination] = useState<PaginationParams>({
    current: 1,
    pageSize: 20,
  })
  const [filters, setFilters] = useState<Record<string, any>>({})
  const [form] = Form.useForm()

  // 获取审计日志列表
  const fetchAuditLogs = async (params?: PaginationParams) => {
    setLoading(true)
    try {
      const searchParams = new URLSearchParams({
        page: String(params?.current || pagination.current),
        page_size: String(params?.pageSize || pagination.pageSize),
        ...filters,
      })

      const response = await fetch(`/api/v1/audit?${searchParams}`)
      const data = await response.json()

      if (data.code === 0) {
        // 确保每条日志都有唯一的 ID，如果没有则使用索引
        const logsWithKeys = (data.data.logs || []).map((log: AuditLog, index: number) => ({
          ...log,
          id: log.id || `temp-${Date.now()}-${index}`,
        }))
        setLogs(logsWithKeys)
        setTotal(data.data.pagination?.total || 0)
        if (params) {
          setPagination(params)
        }
      } else {
        console.error('Failed to fetch audit logs:', data.message)
      }
    } catch (error) {
      console.error('Error fetching audit logs:', error)
    } finally {
      setLoading(false)
    }
  }

  // 获取审计统计信息
  const fetchAuditStats = async () => {
    try {
      const response = await fetch('/api/v1/audit/stats')
      const data = await response.json()

      if (data.code === 0) {
        setStats(data.data)
      }
    } catch (error) {
      console.error('Error fetching audit stats:', error)
    }
  }

  useEffect(() => {
    fetchAuditLogs()
    fetchAuditStats()
  }, [])

  // 处理筛选条件变化
  const handleFilterChange = (_changedFields: any, allFields: any) => {
    const newFilters: Record<string, any> = {}

    if (allFields.action) {
      newFilters.action = allFields.action
    }
    if (allFields.resource) {
      newFilters.resource = allFields.resource
    }
    if (allFields.user_id) {
      newFilters.user_id = allFields.user_id
    }
    if (allFields.date_range) {
      newFilters.start_time = allFields.date_range[0]?.format('YYYY-MM-DD HH:mm:ss')
      newFilters.end_time = allFields.date_range[1]?.format('YYYY-MM-DD HH:mm:ss')
    }

    setFilters(newFilters)
  }

  // 执行搜索
  const handleSearch = () => {
    fetchAuditLogs({ current: 1, pageSize: pagination.pageSize })
  }

  // 重置筛选
  const handleReset = () => {
    form.resetFields()
    setFilters({})
    fetchAuditLogs({ current: 1, pageSize: pagination.pageSize })
  }

  // 处理分页变化
  const handleTableChange = (newPagination: TablePaginationConfig) => {
    fetchAuditLogs({
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 20,
    })
  }

  // 格式化详情显示
  const formatDetails = (details: Record<string, any>): string => {
    if (!details) return '-'
    const entries = Object.entries(details)
    const preview = entries.slice(0, 2).map(([key, value]) => `${key}: ${JSON.stringify(value)}`).join(', ')
    return entries.length > 2 ? `${preview}...` : preview
  }

  const columns: ColumnsType<AuditLog> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id: number | string) => String(id).replace(/^temp-\d+-\d+$/, '-'),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
      render: (id?: number) => id || '-',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (action: string) => {
        const colors: Record<string, string> = {
          create: 'green',
          update: 'blue',
          delete: 'red',
        }
        return <Tag color={colors[action] || 'default'}>{action}</Tag>
      },
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      width: 150,
    },
    {
      title: '资源ID',
      dataIndex: 'resource_id',
      key: 'resource_id',
      width: 100,
      render: (id?: number) => id || '-',
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      ellipsis: true,
      render: (details: Record<string, any>) => (
        <span title={JSON.stringify(details)}>{formatDetails(details)}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>{status}</Tag>
      ),
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 150,
      ellipsis: true,
    },
  ]

  return (
    <div>
      <h2>审计日志</h2>

      {/* 统计信息卡片 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={8}>
            <Card>
              <Statistic title="总记录数" value={stats.total_logs || 0} />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="操作类型"
                value={Object.keys(stats.by_action || {}).length}
                suffix="种"
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="资源类型"
                value={Object.keys(stats.by_resource || {}).length}
                suffix="种"
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 筛选条件 */}
      <Card style={{ marginBottom: 16 }}>
        <Form
          form={form}
          layout="inline"
          onFieldsChange={handleFilterChange}
        >
          <Form.Item name="action" label="操作类型">
            <Select
              style={{ width: 150 }}
              placeholder="选择操作"
              allowClear
            >
              <Select.Option value="create">创建</Select.Option>
              <Select.Option value="update">更新</Select.Option>
              <Select.Option value="delete">删除</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="resource" label="资源类型">
            <Select
              style={{ width: 150 }}
              placeholder="选择资源"
              allowClear
            >
              <Select.Option value="roles">角色</Select.Option>
              <Select.Option value="users">用户</Select.Option>
              <Select.Option value="ci_instances">CI实例</Select.Option>
              <Select.Option value="tickets">工单</Select.Option>
              <Select.Option value="alerts">告警</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="user_id" label="用户ID">
            <Input
              style={{ width: 120 }}
              placeholder="输入用户ID"
              allowClear
            />
          </Form.Item>

          <Form.Item name="date_range" label="时间范围">
            <DatePicker.RangePicker
              showTime
              format="YYYY-MM-DD HH:mm:ss"
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button onClick={handleReset}>
                重置
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => fetchAuditLogs()}
              >
                刷新
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* 审计日志表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}
