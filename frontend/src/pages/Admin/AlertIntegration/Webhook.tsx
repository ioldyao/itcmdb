import { useEffect, useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  message,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Tag,
  Popconfirm,
  Row,
  Col,
  Typography,
  Divider,
  Alert,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

const { Title, Paragraph } = Typography

interface WebhookConfig {
  id: number
  name: string
  type: 'alertmanager' | 'prometheus' | 'victoriametrics'
  url: string
  enabled: boolean
  description?: string
  created_at: string
  updated_at: string
}

export default function AlertIntegrationWebhook() {
  const [loading, setLoading] = useState(false)
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<WebhookConfig | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchWebhooks()
  }, [])

  const fetchWebhooks = async () => {
    setLoading(true)
    try {
      // TODO: 实际API调用
      // const response = await webhookService.getWebhooks()
      // 暂时使用模拟数据
      setWebhooks([
        {
          id: 1,
          name: 'Alertmanager 主集群',
          type: 'alertmanager',
          url: 'http://alertmanager:9093/api/v1/alerts',
          enabled: true,
          description: '主集群告警接收',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
      ])
    } catch (error) {
      message.error('获取Webhook配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingWebhook(null)
    form.resetFields()
    setIsModalVisible(true)
  }

  const handleEdit = (record: WebhookConfig) => {
    setEditingWebhook(record)
    form.setFieldsValue(record)
    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      // TODO: 实际API调用
      message.success('删除成功')
      fetchWebhooks()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      // TODO: 实际API调用
      message.success(editingWebhook ? '更新成功' : '创建成功')
      setIsModalVisible(false)
      fetchWebhooks()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<WebhookConfig> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      width: 200,
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 150,
      render: (type: string) => {
        const typeMap: Record<string, { text: string; color: string }> = {
          alertmanager: { text: 'Alertmanager', color: 'orange' },
          prometheus: { text: 'Prometheus', color: 'blue' },
          victoriametrics: { text: 'VictoriaMetrics', color: 'green' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'URL',
      dataIndex: 'url',
      ellipsis: true,
      render: (url: string) => (
        <span style={{ fontSize: 12 }}>{url}</span>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      render: (desc: string) => desc || '-',
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 100,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: WebhookConfig) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除此Webhook配置？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card>
        <Title level={4}>
          <WebhookOutlined style={{ marginRight: 8 }} />
          Alertmanager WebHook 集成
        </Title>
        <Paragraph>
          启动 Alertmanager WebHook 集成接收告警，可配置多个 Webhook 地址接收告警信息。
          告警将通过标准 Webhook 协议推送到配置的接收端点。
        </Paragraph>

        <Alert
          message="功能说明"
          description={
            <div>
              <p>• 支持配置多个 Alertmanager 实例的 Webhook 地址</p>
              <p>• 自动将接收到的告警转换为统一的告警格式</p>
              <p>• 支持告警去重、聚合等高级功能</p>
              <p>• 实时显示 Webhook 接收状态和错误日志</p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#1890ff' }}>{webhooks.length}</div>
                <div style={{ color: '#999', marginTop: 8 }}>Webhook总数</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#52c41a' }}>
                  {webhooks.filter(w => w.enabled).length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>已启用</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#faad14' }}>
                  {webhooks.filter(w => !w.enabled).length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>已禁用</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#13c2c2' }}>99.9%</div>
                <div style={{ color: '#999', marginTop: 8 }}>可用率</div>
              </div>
            </Card>
          </Col>
        </Row>

        <Divider />

        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={5} style={{ margin: 0 }}>Webhook 配置列表</Title>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新增 Webhook
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={webhooks}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title={editingWebhook ? '编辑 Webhook' : '新增 Webhook'}
        open={isModalVisible}
        onOk={handleSubmit}
        onCancel={() => setIsModalVisible(false)}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          preserve={false}
        >
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="请输入Webhook名称" />
          </Form.Item>

          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="请选择Webhook类型">
              <Select.Option value="alertmanager">Alertmanager</Select.Option>
              <Select.Option value="prometheus">Prometheus</Select.Option>
              <Select.Option value="victoriametrics">VictoriaMetrics</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Webhook URL"
            name="url"
            rules={[
              { required: true, message: '请输入Webhook URL' },
              { type: 'url', message: '请输入有效的URL' }
            ]}
          >
            <Input.TextArea
              rows={3}
              placeholder="例如: http://alertmanager:9093/api/v1/alerts"
            />
          </Form.Item>

          <Form.Item
            label="描述"
            name="description"
          >
            <Input.TextArea
              rows={3}
              placeholder="请输入描述信息（可选）"
            />
          </Form.Item>

          <Form.Item
            label="启用状态"
            name="enabled"
            initialValue={true}
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
