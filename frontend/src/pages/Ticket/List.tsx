import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Tag, Space, message, Empty } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { ColumnsType } from 'antd/es/table'
import { ticketService, Ticket } from '@/services/ticket'
import dayjs from 'dayjs'

const statusMap: Record<string, { text: string; color: string }> = {
  open: { text: '待处理', color: 'blue' },
  in_progress: { text: '处理中', color: 'orange' },
  resolved: { text: '已解决', color: 'green' },
  closed: { text: '已关闭', color: 'default' },
}

const priorityMap: Record<string, { text: string; color: string }> = {
  low: { text: '低', color: 'blue' },
  medium: { text: '中', color: 'orange' },
  high: { text: '高', color: 'red' },
  critical: { text: '紧急', color: 'magenta' },
}

export default function TicketList() {
  const navigate = useNavigate()
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [priorityFilter, setPriorityFilter] = useState<string>('')
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const fetchTickets = async () => {
    setLoading(true)
    try {
      const params: any = { page, pageSize }
      if (keyword) params.keyword = keyword
      if (statusFilter) params.status = statusFilter
      if (priorityFilter) params.priority = priorityFilter

      const res = await ticketService.getTickets(params)
      const data = res as any
      if (data.code === 0) {
        setTickets(data.data || [])
        setTotal(data.total || 0)
      }
    } catch (error) {
      message.error('获取工单列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTickets()
  }, [page, pageSize])

  const handleSearch = () => {
    setPage(1)
    fetchTickets()
  }

  const handleReset = () => {
    setKeyword('')
    setStatusFilter('')
    setPriorityFilter('')
    setPage(1)
    setTimeout(fetchTickets, 0)
  }

  const columns: ColumnsType<Ticket> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      render: (id: string) => <span className="text-brand-primary">{id}</span>,
    },
    {
      title: '标题',
      dataIndex: 'title',
      ellipsis: true,
      render: (text: string, record) => (
        <a
          onClick={() => navigate(`/tickets/${record.id}`)}
          className="text-brand-primary hover:underline cursor-pointer"
        >
          {text}
        </a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => {
        const s = statusMap[status] || { text: status, color: 'default' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      render: (priority: string) => {
        const p = priorityMap[priority] || { text: priority, color: 'default' }
        return <Tag color={p.color}>{p.text}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      render: (time: string) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      width: 180,
      render: (time: string) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => navigate(`/tickets/${record.id}`)}>
            详情
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">工单管理</h1>
          <p className="text-gray-600 dark:text-text-secondary">管理和跟踪所有工单</p>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tickets')}>
          创建工单
        </Button>
      </div>

      {/* 筛选栏 */}
      <div className="mb-6 flex flex-wrap gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Input
          placeholder="搜索工单标题..."
          prefix={<SearchOutlined size={16} />}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onPressEnter={handleSearch}
          className="w-64"
          allowClear
        />
        <Select
          placeholder="状态"
          value={statusFilter}
          onChange={setStatusFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="open">待处理</Select.Option>
          <Select.Option value="in_progress">处理中</Select.Option>
          <Select.Option value="resolved">已解决</Select.Option>
          <Select.Option value="closed">已关闭</Select.Option>
        </Select>
        <Select
          placeholder="优先级"
          value={priorityFilter}
          onChange={setPriorityFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="low">低</Select.Option>
          <Select.Option value="medium">中</Select.Option>
          <Select.Option value="high">高</Select.Option>
          <Select.Option value="critical">紧急</Select.Option>
        </Select>
        <Button onClick={handleSearch} icon={<SearchOutlined size={16} />}>
          搜索
        </Button>
        <Button onClick={handleReset}>重置</Button>
        <div className="flex-1" />
        <Button icon={<ReloadOutlined size={16} />} onClick={fetchTickets}>
          刷新
        </Button>
      </div>

      {/* 表格 */}
      <div className="bg-white dark:bg-bg-secondary rounded-lg border border-gray-200 dark:border-white/8">
        <Table
          columns={columns}
          dataSource={tickets}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条工单`,
            pageSizeOptions: ['10', '20', '50'],
            onChange: (newPage, newPageSize) => {
              setPage(newPage)
              setPageSize(newPageSize)
            },
          }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无工单数据"
              >
                <Button type="primary" onClick={() => navigate('/tickets')}>
                  创建第一个工单
                </Button>
              </Empty>
            ),
          }}
        />
      </div>
    </div>
  )
}
