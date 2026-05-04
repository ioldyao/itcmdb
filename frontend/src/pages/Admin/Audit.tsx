import { useState, useEffect } from 'react'
import { Table, Card, Form, Select, DatePicker, Input, Button, Space, Tag, Tooltip } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import {
  PlusCircle, Edit, Trash2, LogIn, LogOut, CheckCircle, StopCircle,
  Play, Power, Eye, XCircle, FlaskConical, Send, MessageSquare, RefreshCw,
  User, Shield, Server, FileText, Bell, BellRing, UserCheck, Users,
  Globe, GitBranch, Layout, Mail, Monitor, ClipboardList, BarChart3, Layers,
  CheckCircle2, XOctagon, CircleDot,
} from 'lucide-react'
import { motion } from 'framer-motion'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { useAuthStore } from '@/stores/authStore'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

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

// Action type config: icon, color, label
const ACTION_CONFIG: Record<string, { icon: typeof PlusCircle; color: string; antColor: string; label: string }> = {
  create:            { icon: PlusCircle,     color: 'text-green-500',   antColor: 'green',   label: '创建' },
  update:            { icon: Edit,           color: 'text-blue-500',    antColor: 'blue',    label: '更新' },
  delete:            { icon: Trash2,         color: 'text-red-500',     antColor: 'red',     label: '删除' },
  login:             { icon: LogIn,          color: 'text-cyan-500',    antColor: 'cyan',    label: '登录' },
  logout:            { icon: LogOut,         color: 'text-orange-500',  antColor: 'orange',  label: '登出' },
  enable:            { icon: CheckCircle,    color: 'text-green-500',   antColor: 'green',   label: '启用' },
  disable:           { icon: StopCircle,     color: 'text-gray-500',    antColor: 'default', label: '禁用' },
  platform_start:    { icon: Play,           color: 'text-green-500',   antColor: 'green',   label: '平台启动' },
  platform_stop:     { icon: Power,          color: 'text-red-500',     antColor: 'red',     label: '平台停止' },
  acknowledge:       { icon: Eye,            color: 'text-blue-500',    antColor: 'blue',    label: '确认' },
  close:             { icon: XCircle,        color: 'text-gray-500',    antColor: 'default', label: '关闭' },
  batch_acknowledge: { icon: Eye,            color: 'text-blue-500',    antColor: 'blue',    label: '批量确认' },
  batch_close:       { icon: XCircle,        color: 'text-gray-500',    antColor: 'default', label: '批量关闭' },
  test:              { icon: FlaskConical,   color: 'text-purple-500',  antColor: 'purple',  label: '测试' },
  send:              { icon: Send,           color: 'text-blue-500',    antColor: 'blue',    label: '发送' },
  add_comment:       { icon: MessageSquare,  color: 'text-teal-500',    antColor: 'cyan',    label: '添加评论' },
  update_status:     { icon: RefreshCw,      color: 'text-blue-500',    antColor: 'blue',    label: '更新状态' },
}

// Resource type config: icon, label
const RESOURCE_CONFIG: Record<string, { icon: typeof User; label: string }> = {
  user:                { icon: User,         label: '用户' },
  roles:               { icon: Shield,       label: '角色' },
  ci_instances:        { icon: Server,       label: 'CI资产' },
  ticket:              { icon: FileText,     label: '工单' },
  alert:               { icon: Bell,         label: '告警' },
  alert_rule:          { icon: BellRing,     label: '告警规则' },
  alert_receiver:      { icon: UserCheck,    label: '接收人' },
  alert_receiver_group:{ icon: Users,        label: '接收组' },
  webhook:             { icon: Globe,        label: 'Webhook' },
  routing:             { icon: GitBranch,    label: '路由规则' },
  template:            { icon: Layout,       label: '通知模板' },
  notification:        { icon: Mail,         label: '通知' },
  platform:            { icon: Monitor,      label: '平台' },
  audit_logs:          { icon: ClipboardList,label: '审计日志' },
}

const StatCard = ({ title, value, icon: Icon, color }: {
  title: string
  value: string | number
  icon: typeof ClipboardList
  color: string
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    className={`
      relative bg-white dark:bg-gradient-card rounded-xl p-5
      border border-gray-200 dark:border-white/8
      shadow-card-light dark:shadow-card
      overflow-hidden group
    `}
  >
    <div className={`absolute inset-0 bg-gradient-to-br ${color} opacity-0 group-hover:opacity-10 transition-opacity`} />
    <div className={`w-10 h-10 rounded-lg bg-gradient-to-br ${color} flex items-center justify-center mb-3`}>
      <Icon size={20} className="text-white" />
    </div>
    <h3 className="text-gray-500 dark:text-text-secondary text-xs mb-1">{title}</h3>
    <div className="text-2xl font-semibold text-gray-900 dark:text-text-primary">{value}</div>
  </motion.div>
)

export default function AdminAudit() {
  const { token, _hasHydrated } = useAuthStore()
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [stats, setStats] = useState<AuditStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [pagination, setPagination] = useState<PaginationParams>({ current: 1, pageSize: 20 })
  const [filters, setFilters] = useState<Record<string, any>>({})
  const [form] = Form.useForm()

  if (!_hasHydrated) {
    return <div className="p-6 text-center text-gray-500 dark:text-text-secondary">加载中...</div>
  }

  const fetchAuditLogs = async (params?: PaginationParams) => {
    setLoading(true)
    try {
      const searchParams = new URLSearchParams({
        page: String(params?.current || pagination.current),
        page_size: String(params?.pageSize || pagination.pageSize),
        ...filters,
      })
      const response = await fetch(`/api/v1/audit?${searchParams}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()
      if (data.code === 0) {
        const logsWithKeys = (data.data.logs || []).map((log: AuditLog, index: number) => ({
          ...log,
          id: log.id || `temp-${Date.now()}-${index}`,
        }))
        setLogs(logsWithKeys)
        setTotal(data.data.pagination?.total || 0)
        if (params) setPagination(params)
      }
    } catch (error) {
      console.error('Error fetching audit logs:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchAuditStats = async () => {
    try {
      const response = await fetch('/api/v1/audit/stats', {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()
      if (data.code === 0) setStats(data.data)
    } catch (error) {
      console.error('Error fetching audit stats:', error)
    }
  }

  useEffect(() => {
    fetchAuditLogs()
    fetchAuditStats()
  }, [])

  const handleFilterChange = (_changedFields: any, allFields: any) => {
    const newFilters: Record<string, any> = {}
    if (allFields.action) newFilters.action = allFields.action
    if (allFields.resource) newFilters.resource = allFields.resource
    if (allFields.user_id) newFilters.user_id = allFields.user_id
    if (allFields.date_range) {
      newFilters.start_time = allFields.date_range[0]?.format('YYYY-MM-DD HH:mm:ss')
      newFilters.end_time = allFields.date_range[1]?.format('YYYY-MM-DD HH:mm:ss')
    }
    setFilters(newFilters)
  }

  const handleSearch = () => {
    fetchAuditLogs({ current: 1, pageSize: pagination.pageSize })
  }

  const handleReset = () => {
    form.resetFields()
    setFilters({})
    fetchAuditLogs({ current: 1, pageSize: pagination.pageSize })
  }

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    fetchAuditLogs({ current: newPagination.current || 1, pageSize: newPagination.pageSize || 20 })
  }

  // Render action with icon and colored tag, with service name for platform events
  const renderAction = (action: string, record: AuditLog) => {
    const config = ACTION_CONFIG[action]
    const serviceName = record.details?.service
    if (!config) return <Tag>{action}</Tag>
    const Icon = config.icon
    return (
      <div className="flex flex-col gap-0.5">
        <Tag color={config.antColor} className="flex items-center gap-1 !m-0 w-fit">
          <Icon size={13} />
          <span>{config.label}</span>
        </Tag>
        {serviceName && (
          <span className="text-xs text-gray-500 dark:text-text-tertiary pl-0.5">{serviceName}</span>
        )}
      </div>
    )
  }

  // Render resource with icon and label
  const renderResource = (resource: string) => {
    const config = RESOURCE_CONFIG[resource]
    if (!config) return <span className="text-gray-600 dark:text-text-secondary">{resource}</span>
    const Icon = config.icon
    return (
      <span className="inline-flex items-center gap-1.5 text-gray-700 dark:text-text-secondary">
        <Icon size={14} className="text-gray-400 dark:text-text-tertiary" />
        <span>{config.label}</span>
      </span>
    )
  }

  // Render status with icon
  const renderStatus = (status: string) => {
    if (status === 'success') {
      return (
        <span className="inline-flex items-center gap-1 text-green-600 dark:text-green-400">
          <CheckCircle2 size={14} />
          <span className="text-xs">成功</span>
        </span>
      )
    }
    return (
      <span className="inline-flex items-center gap-1 text-red-500 dark:text-red-400">
        <XOctagon size={14} />
        <span className="text-xs">失败</span>
      </span>
    )
  }

  // Render details as key-value chips
  const renderDetails = (details: Record<string, any>) => {
    if (!details || Object.keys(details).length === 0) {
      return <span className="text-gray-400 dark:text-text-tertiary">-</span>
    }
    const entries = Object.entries(details)
    const visible = entries.slice(0, 2)
    const remaining = entries.length - 2

    return (
      <div className="flex flex-wrap items-center gap-1">
        {visible.map(([key, value]) => (
          <Tooltip key={key} title={`${key}: ${JSON.stringify(value)}`}>
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-gray-100 dark:bg-white/5 text-xs text-gray-600 dark:text-text-secondary max-w-[180px] truncate">
              <CircleDot size={10} className="text-gray-400 dark:text-text-tertiary flex-shrink-0" />
              <span className="font-medium text-gray-700 dark:text-text-primary">{key}:</span>
              <span className="truncate">{typeof value === 'object' ? JSON.stringify(value) : String(value)}</span>
            </span>
          </Tooltip>
        ))}
        {remaining > 0 && (
          <Tooltip title={entries.slice(2).map(([k, v]) => `${k}: ${JSON.stringify(v)}`).join('\n')}>
            <span className="text-xs text-blue-500 dark:text-blue-400 cursor-pointer">+{remaining}</span>
          </Tooltip>
        )}
      </div>
    )
  }

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 200,
      render: (time: string) => (
        <div className="flex flex-col">
          <span className="text-gray-900 dark:text-text-primary text-sm">{dayjs(time).format('MM-DD HH:mm:ss')}</span>
          <span className="text-gray-400 dark:text-text-tertiary text-xs">{dayjs(time).fromNow()}</span>
        </div>
      ),
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 140,
      render: (action: string, record: AuditLog) => renderAction(action, record),
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: renderResource,
    },
    {
      title: '资源ID',
      dataIndex: 'resource_id',
      key: 'resource_id',
      width: 80,
      render: (id?: number) => id ? (
        <span className="inline-flex items-center justify-center w-7 h-7 rounded-full bg-gray-100 dark:bg-white/5 text-xs font-medium text-gray-700 dark:text-text-primary">
          {id}
        </span>
      ) : <span className="text-gray-400 dark:text-text-tertiary">-</span>,
    },
    {
      title: '用户',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
      render: (id?: number) => id ? (
        <span className="inline-flex items-center gap-1 text-gray-600 dark:text-text-secondary text-sm">
          <User size={13} className="text-gray-400 dark:text-text-tertiary" />
          #{id}
        </span>
      ) : <span className="text-gray-400 dark:text-text-tertiary text-xs">系统</span>,
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      ellipsis: true,
      render: renderDetails,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: renderStatus,
    },
    {
      title: 'IP',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
      render: (ip: string) => (
        <span className="font-mono text-xs text-gray-600 dark:text-text-secondary">{ip || '-'}</span>
      ),
    },
  ]

  return (
    <div className="p-6">
      {/* Header */}
      <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-1">审计日志</h1>
        <p className="text-gray-500 dark:text-text-secondary text-sm">系统操作记录与安全审计</p>
      </motion.div>

      {/* Stat Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <StatCard title="总记录数" value={stats.total_logs || 0} icon={ClipboardList} color="from-blue-500 to-blue-600" />
          <StatCard title="操作类型" value={`${Object.keys(stats.by_action || {}).length} 种`} icon={BarChart3} color="from-purple-500 to-purple-600" />
          <StatCard title="资源类型" value={`${Object.keys(stats.by_resource || {}).length} 种`} icon={Layers} color="from-green-500 to-green-600" />
        </div>
      )}

      {/* Filters */}
      <Card className="mb-6 dark:bg-bg-secondary dark:border-white/8">
        <Form form={form} layout="inline" onFieldsChange={handleFilterChange} className="flex flex-wrap gap-y-3">
          <Form.Item name="action" label={<span className="text-gray-600 dark:text-text-secondary">操作</span>}>
            <Select style={{ width: 150 }} placeholder="选择操作" allowClear>
              {Object.entries(ACTION_CONFIG).map(([key, cfg]) => (
                <Select.Option key={key} value={key}>
                  <span className="flex items-center gap-1.5">
                    <cfg.icon size={13} />
                    {cfg.label}
                  </span>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="resource" label={<span className="text-gray-600 dark:text-text-secondary">资源</span>}>
            <Select style={{ width: 150 }} placeholder="选择资源" allowClear>
              {Object.entries(RESOURCE_CONFIG).map(([key, cfg]) => (
                <Select.Option key={key} value={key}>
                  <span className="flex items-center gap-1.5">
                    <cfg.icon size={13} />
                    {cfg.label}
                  </span>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="user_id" label={<span className="text-gray-600 dark:text-text-secondary">用户</span>}>
            <Input style={{ width: 120 }} placeholder="用户ID" allowClear />
          </Form.Item>

          <Form.Item name="date_range" label={<span className="text-gray-600 dark:text-text-secondary">时间</span>}>
            <DatePicker.RangePicker showTime format="YYYY-MM-DD HH:mm:ss" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>搜索</Button>
              <Button onClick={handleReset}>重置</Button>
              <Button icon={<ReloadOutlined />} onClick={() => fetchAuditLogs()}>刷新</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* Table */}
      <Card className="dark:bg-bg-secondary dark:border-white/8">
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
            showTotal: (t) => `共 ${t} 条记录`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1100 }}
        />
      </Card>
    </div>
  )
}
