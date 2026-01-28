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
  Tabs,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
  DownloadOutlined,
  UploadOutlined,
  CopyOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { inboundWebhookService, outboundWebhookService, InboundWebhook, OutboundWebhook } from '@/services/webhookService'
import { alertReceiverService, AlertReceiver } from '@/services/alertReceiverService'

const { Title, Paragraph, Text } = Typography

export default function AlertIntegrationWebhook() {
  const [activeTab, setActiveTab] = useState('inbound')

  // 渲染接收外部告警标签页内容
  const renderInboundTab = () => <InboundWebhooks />

  // 渲染ITCMDB推送标签页内容
  const renderOutboundTab = () => <OutboundWebhooks />

  return (
    <div>
      <Card>
        <Title level={4}>
          <ApiOutlined style={{ marginRight: 8 }} />
          告警 Webhook 集成
        </Title>
        <Paragraph>
          配置告警的 Webhook 接收和推送。支持接收外部监控系统的告警，也可以将ITCMDB告警推送到外部系统。
        </Paragraph>

        <Alert
          message="功能说明"
          description={
            <div>
              <p><strong>接收外部告警：</strong>生成接收地址，外部系统（如Alertmanager）可向ITCMDB推送告警</p>
              <p><strong>ITCMDB推送：</strong>将ITCMDB告警推送到外部系统，支持Alertmanager、钉钉、企业微信等</p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'inbound',
              label: (
                <span>
                  <DownloadOutlined />
                  接收外部告警
                </span>
              ),
              children: renderInboundTab(),
            },
            {
              key: 'outbound',
              label: (
                <span>
                  <UploadOutlined />
                  ITCMDB推送
                </span>
              ),
              children: renderOutboundTab(),
            },
          ]}
        />
      </Card>
    </div>
  )
}

// ==================== 接收外部告警组件 ====================

function InboundWebhooks() {
  const [loading, setLoading] = useState(false)
  const [webhooks, setWebhooks] = useState<InboundWebhook[]>([])
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<InboundWebhook | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchInboundWebhooks()
  }, [])

  const fetchInboundWebhooks = async () => {
    setLoading(true)
    try {
      const response = await inboundWebhookService.getWebhooks()
      setWebhooks(response?.webhooks || [])
    } catch (error) {
      message.error('获取接收地址失败')
      setWebhooks([])
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingWebhook(null)
    form.resetFields()
    setIsModalVisible(true)
  }

  const handleEdit = (record: InboundWebhook) => {
    setEditingWebhook(record)
    form.setFieldsValue({
      name: record.name,
      source_type: record.source_type,
      enabled: record.enabled,
      description: record.description,
    })
    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await inboundWebhookService.deleteWebhook(id)
      message.success('删除成功')
      fetchInboundWebhooks()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleCopyUrl = (url: string) => {
    navigator.clipboard.writeText(url)
    message.success('URL已复制到剪贴板')
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingWebhook) {
        await inboundWebhookService.updateWebhook(editingWebhook.id, values)
        message.success('更新成功')
      } else {
        await inboundWebhookService.createWebhook(values)
        message.success('创建成功')
      }
      setIsModalVisible(false)
      fetchInboundWebhooks()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<InboundWebhook> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      width: 180,
    },
    {
      title: '来源类型',
      dataIndex: 'source_type',
      width: 150,
      render: (type: string) => {
        const typeMap: Record<string, { text: string; color: string }> = {
          alertmanager: { text: 'Alertmanager', color: 'orange' },
          prometheus: { text: 'Prometheus', color: 'blue' },
          victoriametrics: { text: 'VictoriaMetrics', color: 'green' },
          custom: { text: '自定义', color: 'purple' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Webhook接收地址',
      dataIndex: 'webhook_url',
      width: 350,
      render: (url: string) => (
        <Space size="small">
          <Text ellipsis style={{ maxWidth: 280, fontSize: 12 }}>
            {url}
          </Text>
          <Button
            type="link"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => handleCopyUrl(url)}
          >
            复制
          </Button>
        </Space>
      ),
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
      title: '最后接收',
      dataIndex: 'last_received',
      width: 180,
      render: (time: string) => time ? new Date(time).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_: any, record: InboundWebhook) => (
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
            title="确认删除此接收地址？"
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
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 32, fontWeight: 600, color: '#1890ff' }}>{webhooks.length}</div>
              <div style={{ color: '#999', marginTop: 8 }}>接收地址总数</div>
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
              <div style={{ fontSize: 32, fontWeight: 600, color: '#13c2c2' }}>
                {webhooks.filter(w => w.last_received &&
                  new Date(w.last_received!) > new Date(Date.now() - 3600000)).length
                }
              </div>
              <div style={{ color: '#999', marginTop: 8 }}>活跃（1小时内）</div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 32, fontWeight: 600, color: '#722ed1' }}>
                {webhooks.reduce((acc, w) => acc + (w.last_received ? 1 : 0), 0)}
              </div>
              <div style={{ color: '#999', marginTop: 8 }}>有接收记录</div>
            </div>
          </Card>
        </Col>
      </Row>

      <Divider />

      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={5} style={{ margin: 0 }}>接收地址列表</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新增接收地址
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={webhooks}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1400 }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
      />

      <Modal
        title={editingWebhook ? '编辑接收地址' : '新增接收地址'}
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
            <Input placeholder="请输入接收地址名称" />
          </Form.Item>

          <Form.Item
            label="来源类型"
            name="source_type"
            rules={[{ required: true, message: '请选择来源类型' }]}
          >
            <Select placeholder="请选择告警来源类型">
              <Select.Option value="alertmanager">Alertmanager</Select.Option>
              <Select.Option value="prometheus">Prometheus</Select.Option>
              <Select.Option value="victoriametrics">VictoriaMetrics</Select.Option>
              <Select.Option value="custom">自定义</Select.Option>
            </Select>
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

          <Alert
            message="Webhook URL将自动生成"
            description="创建成功后，系统将自动生成唯一的Webhook接收地址"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

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

// ==================== ITCMDB推送组件 ====================

function OutboundWebhooks() {
  const [loading, setLoading] = useState(false)
  const [webhooks, setWebhooks] = useState<OutboundWebhook[]>([])
  const [receivers, setReceivers] = useState<AlertReceiver[]>([])
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<OutboundWebhook | null>(null)
  const [targetType, setTargetType] = useState<'alertmanager' | 'receiver'>('alertmanager')
  const [form] = Form.useForm()

  useEffect(() => {
    fetchOutboundWebhooks()
    fetchAllReceivers()
  }, [])

  const fetchAllReceivers = async () => {
    try {
      const response = await alertReceiverService.getReceivers({ page: 1, page_size: 1000 })
      if (response?.receivers) {
        setReceivers(response.receivers)
      }
    } catch (error) {
      console.error('获取接收人列表失败', error)
    }
  }

  useEffect(() => {
    fetchOutboundWebhooks()
  }, [])

  const fetchOutboundWebhooks = async () => {
    setLoading(true)
    try {
      const response = await outboundWebhookService.getWebhooks()
      setWebhooks(response?.webhooks || [])
    } catch (error) {
      message.error('获取推送配置失败')
      setWebhooks([])
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingWebhook(null)
    setTargetType('alertmanager')
    form.resetFields()
    setIsModalVisible(true)
  }

  const handleEdit = (record: OutboundWebhook) => {
    setEditingWebhook(record)
    setTargetType(record.target_type)
    form.setFieldsValue({
      name: record.name,
      target_type: record.target_type,
      receiver_id: record.receiver_id,
      endpoint_url: record.endpoint_url,
      enabled: record.enabled,
      description: record.description,
    })
    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await outboundWebhookService.deleteWebhook(id)
      message.success('删除成功')
      fetchOutboundWebhooks()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleTest = async (id: number) => {
    try {
      await outboundWebhookService.testWebhook(id)
      message.success('测试消息发送成功')
    } catch (error) {
      message.error('测试消息发送失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingWebhook) {
        await outboundWebhookService.updateWebhook(editingWebhook.id, values)
        message.success('更新成功')
      } else {
        await outboundWebhookService.createWebhook(values)
        message.success('创建成功')
      }
      setIsModalVisible(false)
      fetchOutboundWebhooks()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<OutboundWebhook> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      width: 180,
    },
    {
      title: '目标类型',
      dataIndex: 'target_type',
      width: 150,
      render: (type: string) => {
        const typeMap: Record<string, { text: string; color: string }> = {
          alertmanager: { text: 'Alertmanager', color: 'orange' },
          receiver: { text: '告警接收人', color: 'blue' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '推送目标',
      width: 300,
      render: (_: any, record: OutboundWebhook) => {
        if (record.target_type === 'receiver' && record.receiver) {
          const typeMap: Record<string, { text: string; color: string }> = {
            wechat: { text: '企业微信', color: 'green' },
            dingtalk: { text: '钉钉', color: 'blue' },
            feishu: { text: '飞书', color: 'cyan' },
            email: { text: '邮件', color: 'purple' },
            sms: { text: '短信', color: 'magenta' },
          }
          const config = typeMap[record.receiver.type] || { text: record.receiver.type, color: 'default' }
          return (
            <Space>
              <Tag color={config.color}>{config.text}</Tag>
              <Text>{record.receiver.name}</Text>
            </Space>
          )
        }
        return (
          <Text ellipsis style={{ maxWidth: 250, fontSize: 12 }}>
            {record.endpoint_url}
          </Text>
        )
      },
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
      title: '最后推送',
      dataIndex: 'last_sent',
      width: 180,
      render: (time: string) => time ? new Date(time).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: any, record: OutboundWebhook) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            onClick={() => handleTest(record.id)}
          >
            测试
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除此推送配置？"
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
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 32, fontWeight: 600, color: '#1890ff' }}>{webhooks.length}</div>
              <div style={{ color: '#999', marginTop: 8 }}>推送目标总数</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 32, fontWeight: 600, color: '#52c41a' }}>
                {webhooks.filter(w => w.enabled).length}
              </div>
              <div style={{ color: '#999', marginTop: 8 }}>已启用</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 32, fontWeight: 600, color: '#13c2c2' }}>
                {webhooks.filter(w => w.last_sent &&
                  new Date(w.last_sent!) > new Date(Date.now() - 3600000)).length
                }
              </div>
              <div style={{ color: '#999', marginTop: 8 }}>活跃（1小时内）</div>
            </div>
          </Card>
        </Col>
      </Row>

      <Divider />

      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={5} style={{ margin: 0 }}>推送目标列表</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新增推送目标
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={webhooks}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1400 }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
      />

      <Modal
        title={editingWebhook ? '编辑推送目标' : '新增推送目标'}
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
            <Input placeholder="请输入推送目标名称" />
          </Form.Item>

          <Form.Item
            label="目标类型"
            name="target_type"
            rules={[{ required: true, message: '请选择目标类型' }]}
          >
            <Select
              placeholder="请选择推送目标类型"
              onChange={(value: 'alertmanager' | 'receiver') => setTargetType(value)}
            >
              <Select.Option value="alertmanager">Alertmanager</Select.Option>
              <Select.Option value="receiver">告警接收人</Select.Option>
            </Select>
          </Form.Item>

          {targetType === 'receiver' ? (
            <Form.Item
              label="选择接收人"
              name="receiver_id"
              rules={[{ required: true, message: '请选择告警接收人' }]}
            >
              <Select placeholder="请选择已配置的告警接收人">
                {receivers.map(recv => {
                  const typeMap: Record<string, { text: string; color: string }> = {
                    wechat: { text: '企业微信', color: 'green' },
                    dingtalk: { text: '钉钉', color: 'blue' },
                    feishu: { text: '飞书', color: 'cyan' },
                    email: { text: '邮件', color: 'purple' },
                    sms: { text: '短信', color: 'magenta' },
                  }
                  const config = typeMap[recv.type] || { text: recv.type, color: 'default' }
                  return (
                    <Select.Option key={recv.id} value={recv.id}>
                      <Space>
                        <Tag color={config.color}>{config.text}</Tag>
                        <span>{recv.name}</span>
                      </Space>
                    </Select.Option>
                  )
                })}
              </Select>
            </Form.Item>
          ) : (
            <Form.Item
              label="Alertmanager URL"
              name="endpoint_url"
              rules={[{ required: true, message: '请输入Alertmanager URL' }]}
            >
              <Input.TextArea
                rows={3}
                placeholder="请输入Alertmanager的Webhook URL，例如: http://alertmanager:9093/api/v1/alerts"
              />
            </Form.Item>
          )}

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
