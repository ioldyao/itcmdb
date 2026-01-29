import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Tabs, Descriptions, Tag, Button, Spin, message, Space, Divider } from 'antd'
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import { useAlertStore } from '@/stores/alertStore'
import { useAuthStore } from '@/stores/authStore'
import type { AlertInstance } from '@/services/alertService'
import dayjs from 'dayjs'

export default function AlertDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { fetchAlertById } = useAlertStore()
  const { user } = useAuthStore()
  const [alert, setAlert] = useState<AlertInstance | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (id) {
      loadAlert(parseInt(id))
    }
  }, [id])

  const loadAlert = async (alertId: number) => {
    setLoading(true)
    try {
      const data = await fetchAlertById(alertId)
      setAlert(data)
    } catch (error: any) {
      message.error('加载告警详情失败')
      console.error('Failed to load alert:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleAcknowledge = async () => {
    if (!alert) return
    try {
      await useAlertStore.getState().acknowledgeAlert(alert.id, { handler: user?.id || 1, notes: '' })
      message.success('确认成功')
      loadAlert(alert.id)
    } catch (error: any) {
      message.error(error.message || '确认失败')
    }
  }

  const handleClose = async () => {
    if (!alert) return
    try {
      await useAlertStore.getState().closeAlert(alert.id, { handler: user?.id || 1, notes: '' })
      message.success('关闭成功')
      loadAlert(alert.id)
    } catch (error: any) {
      message.error(error.message || '关闭失败')
    }
  }

  if (loading) {
    return (
      <div style={{ padding: 100, textAlign: 'center' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (!alert) {
    return (
      <div style={{ padding: 100, textAlign: 'center' }}>
        <h2>告警不存在</h2>
        <Button type="primary" onClick={() => navigate('/alerts')}>
          返回告警列表
        </Button>
      </div>
    )
  }

  const severityColors: Record<string, string> = {
    critical: 'red',
    high: 'orange',
    medium: 'gold',
    low: 'blue',
  }

  const severityLabels: Record<string, string> = {
    critical: '紧急',
    high: '高',
    medium: '中',
    low: '低',
  }

  const statusMap: Record<string, { text: string; color: string }> = {
    firing: { text: '未恢复', color: 'red' },
    acknowledged: { text: '已确认', color: 'orange' },
    resolved: { text: '已恢复', color: 'green' },
    closed: { text: '已关闭', color: 'default' },
  }

  const statusInfo = statusMap[alert.status] || { text: alert.status, color: 'default' }

  return (
    <div style={{ padding: 24 }}>
      {/* 头部 */}
      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/alerts')}>
              返回列表
            </Button>
            <Tag color={severityColors[alert.severity]} style={{ fontSize: 14 }}>
              {severityLabels[alert.severity]}
            </Tag>
            <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
          </Space>
          <Space>
            {alert.status === 'firing' && (
              <>
                <Button
                  type="primary"
                  icon={<CheckCircleOutlined />}
                  onClick={handleAcknowledge}
                >
                  确认告警
                </Button>
                <Button
                  danger
                  icon={<CloseCircleOutlined />}
                  onClick={handleClose}
                >
                  关闭告警
                </Button>
              </>
            )}
            {alert.status === 'acknowledged' && (
              <Button danger icon={<CloseCircleOutlined />} onClick={handleClose}>
                关闭告警
              </Button>
            )}
          </Space>
        </Space>
      </Card>

      {/* 详情内容 */}
      <Card style={{ marginTop: 16 }}>
        <Tabs
          defaultActiveKey="info"
          items={[
            {
              key: 'info',
              label: '基本信息',
              children: (
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="告警ID">{alert.id}</Descriptions.Item>
                  <Descriptions.Item label="告警标识">{alert.alert_id}</Descriptions.Item>
                  <Descriptions.Item label="告警名称">{alert.title}</Descriptions.Item>
                  <Descriptions.Item label="分类">{alert.category || '-'}</Descriptions.Item>
                  <Descriptions.Item label="严重级别">
                    <Tag color={severityColors[alert.severity]}>
                      {severityLabels[alert.severity]}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="空间/对象">{alert.object_type || '-'}</Descriptions.Item>
                  <Descriptions.Item label="触发次数">{alert.count}</Descriptions.Item>
                  <Descriptions.Item label="首次触发">
                    {dayjs(alert.first_triggered).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>
                  <Descriptions.Item label="最后触发">
                    {dayjs(alert.last_triggered).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>
                  {alert.recovered_at && (
                    <Descriptions.Item label="恢复时间">
                      {dayjs(alert.recovered_at).format('YYYY-MM-DD HH:mm:ss')}
                    </Descriptions.Item>
                  )}
                  {alert.closed_at && (
                    <Descriptions.Item label="关闭时间">
                      {dayjs(alert.closed_at).format('YYYY-MM-DD HH:mm:ss')}
                    </Descriptions.Item>
                  )}
                  {alert.acknowledged_at && (
                    <Descriptions.Item label="确认时间">
                      {dayjs(alert.acknowledged_at).format('YYYY-MM-DD HH:mm:ss')}
                    </Descriptions.Item>
                  )}
                  {alert.handler && (
                    <Descriptions.Item label="处理人">{alert.handler}</Descriptions.Item>
                  )}
                  {alert.handling_notes && (
                    <Descriptions.Item label="处理备注" span={2}>
                      {alert.handling_notes || '-'}
                    </Descriptions.Item>
                  )}
                  <Descriptions.Item label="描述" span={2}>
                    {alert.description || '-'}
                  </Descriptions.Item>
                </Descriptions>
              ),
            },
            {
              key: 'labels',
              label: '标签信息',
              children: (
                <Descriptions column={2} bordered>
                  {alert.tags && Object.keys(alert.tags).length > 0 ? (
                    Object.entries(alert.tags).map(([key, value]) => (
                      <Descriptions.Item key={key} label={key}>
                        <Tag>{value}</Tag>
                      </Descriptions.Item>
                    ))
                  ) : (
                    <Descriptions.Item span={2}>-</Descriptions.Item>
                  )}
                </Descriptions>
              ),
            },
            {
              key: 'triggers',
              label: '触发条件',
              children: (
                <Descriptions column={2} bordered>
                  {alert.trigger_conditions && Object.keys(alert.trigger_conditions).length > 0 ? (
                    Object.entries(alert.trigger_conditions).map(([key, value]) => (
                      <Descriptions.Item key={key} label={key}>
                        {String(value)}
                      </Descriptions.Item>
                    ))
                  ) : (
                    <Descriptions.Item span={2}>-</Descriptions.Item>
                  )}
                </Descriptions>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}
