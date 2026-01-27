import { useEffect, useState } from 'react'
import {
  Table,
  Tag,
  Button,
  Space,
  Input,
  Select,
  DatePicker,
  message,
  Row,
  Col,
  Statistic,
  Card,
  Popconfirm,
  Tooltip,
} from 'antd'
import {
  ReloadOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useAlertStore } from '@/stores/alertStore'
import type { AlertInstance } from '@/services/alertService'
import { useAuthStore } from '@/stores/authStore'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

export default function AlertList() {
  const {
    alerts,
    loading,
    total,
    page,
    pageSize,
    statistics,
    fetchAlerts,
    fetchStatistics,
    acknowledgeAlert,
    closeAlert,
    setFilters,
    clearFilters,
  } = useAlertStore()

  const { user } = useAuthStore()
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string[]>([])
  const [severityFilter, setSeverityFilter] = useState<string[]>([])

  useEffect(() => {
    fetchAlerts()
    fetchStatistics()
  }, [])

  useEffect(() => {
    fetchAlerts()
  }, [page, pageSize]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    // 当filters变化时，重置到第一页并重新获取
    if (filters && Object.keys(filters).length > 0) {
      fetchAlerts()
    }
  }, [filters]) // eslint-disable-line react-hooks/exhaustive-deps

  // 搜索处理
  const handleSearch = () => {
    setFilters({
      searchKeyword: keyword || undefined,
      status: statusFilter.length > 0 ? statusFilter : undefined,
      severity: severityFilter.length > 0 ? severityFilter : undefined,
    })
  }

  // 重置筛选
  const handleReset = () => {
    setKeyword('')
    setStatusFilter([])
    setSeverityFilter([])
    clearFilters()
  }

  // 时间范围变化
  const handleDateRangeChange = (dates: any) => {
    if (dates && dates[0] && dates[1]) {
      setFilters({
        startTime: dates[0].toISOString(),
        endTime: dates[1].toISOString(),
      })
    } else {
      setFilters({
        startTime: undefined,
        endTime: undefined,
      })
    }
  }

  // 确认告警
  const handleAcknowledge = async (id: number) => {
    try {
      await acknowledgeAlert(id, { handler: user?.id || 1 })
      message.success('确认成功')
    } catch (error: any) {
      message.error(error.message || '确认失败')
    }
  }

  // 关闭告警
  const handleClose = async (id: number) => {
    try {
      await closeAlert(id, { handler: user?.id || 1 })
      message.success('关闭成功')
    } catch (error: any) {
      message.error(error.message || '关闭失败')
    }
  }

  // 批量确认
  const handleBatchAcknowledge = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要确认的告警')
      return
    }
    try {
      await acknowledgeAlert(
        selectedRowKeys[0] as number,
        { handler: user?.id || 1 }
      )
      message.success('批量确认成功')
      setSelectedRowKeys([])
    } catch (error: any) {
      message.error(error.message || '批量确认失败')
    }
  }

  // 批量关闭
  const handleBatchClose = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要关闭的告警')
      return
    }
    try {
      await closeAlert(
        selectedRowKeys[0] as number,
        { handler: user?.id || 1 }
      )
      message.success('批量关闭成功')
      setSelectedRowKeys([])
    } catch (error: any) {
      message.error(error.message || '批量关闭失败')
    }
  }

  const columns: ColumnsType<AlertInstance> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      fixed: 'left',
    },
    {
      title: '告警标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
      fixed: 'left',
    },
    {
      title: '分类',
      dataIndex: 'category',
      width: 120,
      render: (category: string) => category || '-',
    },
    {
      title: '严重级别',
      dataIndex: 'severity',
      width: 100,
      render: (severity: string) => {
        const colors: Record<string, string> = {
          critical: 'red',
          high: 'orange',
          medium: 'gold',
          low: 'blue',
        }
        const labels: Record<string, string> = {
          critical: '紧急',
          high: '高',
          medium: '中',
          low: '低',
        }
        return <Tag color={colors[severity]}>{labels[severity]}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => {
        const map: Record<string, { text: string; color: string }> = {
          firing: { text: '未恢复', color: 'red' },
          acknowledged: { text: '已确认', color: 'orange' },
          resolved: { text: '已恢复', color: 'green' },
          closed: { text: '已关闭', color: 'default' },
        }
        const s = map[status] || { text: status, color: 'default' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '触发次数',
      dataIndex: 'count',
      width: 100,
    },
    {
      title: '当前值/阈值',
      width: 150,
      render: (_: any, record: AlertInstance) => {
        if (record.metrics) {
          return (
            <Tooltip title={`偏差: ${record.metrics.deviation?.toFixed(2)}%`}>
              <span>
                {record.metrics.current_value?.toFixed(2)} / {record.metrics.threshold_value}
              </span>
            </Tooltip>
          )
        }
        return '-'
      },
    },
    {
      title: '首次触发',
      dataIndex: 'first_triggered',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '最后触发',
      dataIndex: 'last_triggered',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_: any, record: AlertInstance) => (
        <Space size="small">
          {record.status === 'firing' && (
            <>
              <Button
                type="link"
                size="small"
                icon={<CheckCircleOutlined />}
                onClick={() => handleAcknowledge(record.id)}
              >
                确认
              </Button>
              <Popconfirm
                title="确认关闭此告警？"
                onConfirm={() => handleClose(record.id)}
                okText="确定"
                cancelText="取消"
              >
                <Button type="link" size="small" danger icon={<CloseCircleOutlined />}>
                  关闭
                </Button>
              </Popconfirm>
            </>
          )}
          {record.status === 'acknowledged' && (
            <Popconfirm
              title="确认关闭此告警？"
              onConfirm={() => handleClose(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<CloseCircleOutlined />}>
                关闭
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 24, fontWeight: 600, marginBottom: 8, color: '#1890ff' }}>告警中心</h2>
        <p style={{ color: '#666', margin: 0 }}>
          实时监控和管理系统告警，支持告警查看、确认、关闭等操作
        </p>
      </div>

      {/* 统计卡片 */}
      {statistics && (
        <>
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={4}>
              <Card>
                <Statistic
                  title="总告警数"
                  value={statistics.stats?.total || 0}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col span={5}>
              <Card>
                <Statistic
                  title="未恢复"
                  value={statistics.stats?.firing || 0}
                  valueStyle={{ color: '#ff4d4f' }}
                  suffix="条"
                />
              </Card>
            </Col>
            <Col span={5}>
              <Card>
                <Statistic
                  title="已确认"
                  value={statistics.stats?.acknowledged || 0}
                  valueStyle={{ color: '#faad14' }}
                  suffix="条"
                />
              </Card>
            </Col>
            <Col span={5}>
              <Card>
                <Statistic
                  title="已恢复"
                  value={statistics.stats?.resolved || 0}
                  valueStyle={{ color: '#52c41a' }}
                  suffix="条"
                />
              </Card>
            </Col>
            <Col span={5}>
              <Card>
                <Statistic
                  title="已关闭"
                  value={statistics.stats?.closed || 0}
                  valueStyle={{ color: '#8c8c8c' }}
                  suffix="条"
                />
              </Card>
            </Col>
          </Row>

          {/* 严重程度统计 */}
          {statistics.severity_stats && statistics.severity_stats.length > 0 && (
            <Row gutter={16} style={{ marginBottom: 16 }}>
              {statistics.severity_stats.map((stat: any) => {
                const config: Record<string, { label: string; color: string; bg: string }> = {
                  critical: { label: '紧急', color: '#fff', bg: '#ff4d4f' },
                  high: { label: '高', color: '#fff', bg: '#fa8c16' },
                  medium: { label: '中', color: '#fff', bg: '#faad14' },
                  low: { label: '低', color: '#fff', bg: '#1890ff' },
                }
                const cfg = config[stat.severity] || { label: stat.severity, color: '#fff', bg: '#8c8c8c' }
                return (
                  <Col span={6} key={stat.severity}>
                    <Card
                      size="small"
                      style={{
                        background: cfg.bg,
                        border: 'none',
                      }}
                    >
                      <div style={{ color: cfg.color }}>
                        <div style={{ fontSize: 14, opacity: 0.9 }}>{cfg.label}告警</div>
                        <div style={{ fontSize: 24, fontWeight: 600, marginTop: 4 }}>
                          {stat.count} <span style={{ fontSize: 14, marginLeft: 4 }}>条</span>
                        </div>
                      </div>
                    </Card>
                  </Col>
                )
              })}
            </Row>
          )}
        </>
      )}

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Input
              placeholder="搜索告警标题"
              prefix={<SearchOutlined />}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={handleSearch}
              allowClear
            />
          </Col>
          <Col span={4}>
            <Select
              mode="multiple"
              placeholder="状态筛选"
              style={{ width: '100%' }}
              value={statusFilter}
              onChange={setStatusFilter}
              allowClear
            >
              <Select.Option value="firing">未恢复</Select.Option>
              <Select.Option value="acknowledged">已确认</Select.Option>
              <Select.Option value="resolved">已恢复</Select.Option>
              <Select.Option value="closed">已关闭</Select.Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              mode="multiple"
              placeholder="严重程度"
              style={{ width: '100%' }}
              value={severityFilter}
              onChange={setSeverityFilter}
              allowClear
            >
              <Select.Option value="critical">紧急</Select.Option>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Col>
          <Col span={6}>
            <RangePicker
              style={{ width: '100%' }}
              onChange={handleDateRangeChange}
              showTime
              placeholder={['开始时间', '结束时间']}
            />
          </Col>
          <Col span={4}>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => fetchAlerts()}>
                刷新
              </Button>
              {(keyword || statusFilter.length > 0 || severityFilter.length > 0) && (
                <Button icon={<ClearOutlined />} onClick={handleReset}>
                  重置
                </Button>
              )}
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Input
              placeholder="搜索告警标题"
              prefix={<SearchOutlined />}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={handleSearch}
              allowClear
            />
          </Col>
          <Col span={4}>
            <Select
              mode="multiple"
              placeholder="状态"
              style={{ width: '100%' }}
              value={statusFilter}
              onChange={setStatusFilter}
              allowClear
            >
              <Select.Option value="firing">未恢复</Select.Option>
              <Select.Option value="acknowledged">已确认</Select.Option>
              <Select.Option value="resolved">已恢复</Select.Option>
              <Select.Option value="closed">已关闭</Select.Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              mode="multiple"
              placeholder="严重程度"
              style={{ width: '100%' }}
              value={severityFilter}
              onChange={setSeverityFilter}
              allowClear
            >
              <Select.Option value="critical">紧急</Select.Option>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Col>
          <Col span={6}>
            <RangePicker
              style={{ width: '100%' }}
              onChange={handleDateRangeChange}
              showTime
            />
          </Col>
          <Col span={4}>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => fetchAlerts()}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 批量操作栏 */}
      {selectedRowKeys.length > 0 && (
        <Card style={{ marginBottom: 16 }}>
          <Space>
            <span>已选择 {selectedRowKeys.length} 项</span>
            <Button
              type="primary"
              icon={<CheckCircleOutlined />}
              onClick={handleBatchAcknowledge}
            >
              批量确认
            </Button>
            <Popconfirm
              title={`确认关闭选中的 ${selectedRowKeys.length} 个告警？`}
              onConfirm={handleBatchClose}
              okText="确定"
              cancelText="取消"
            >
              <Button danger icon={<CloseCircleOutlined />}>
                批量关闭
              </Button>
            </Popconfirm>
            <Button icon={<ClearOutlined />} onClick={() => setSelectedRowKeys([])}>
              取消选择
            </Button>
          </Space>
        </Card>
      )}

      {/* 告警列表 */}
      <Table
        columns={columns}
        dataSource={alerts}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1500 }}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (newPage, newPageSize) => {
            useAlertStore.getState().page = newPage
            useAlertStore.getState().pageSize = newPageSize || 20
          },
        }}
        rowSelection={{
          selectedRowKeys,
          onChange: setSelectedRowKeys,
          getCheckboxProps: (record) => ({
            disabled: record.status === 'closed',
          }),
        }}
      />
    </div>
  )
}
