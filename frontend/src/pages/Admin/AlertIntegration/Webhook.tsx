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
  ApiOutlined,
  CopyOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import WorkflowEditor, { BambooPipeline } from '@/components/WorkflowEditor'
import { webhookService, WebhookConfig } from '@/services/webhookService'
import { workflowService } from '@/services/workflowService'

const { Title, Paragraph } = Typography
const { TextArea } = Input

export default function AlertIntegrationWebhook() {
  const [loading, setLoading] = useState(false)
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<WebhookConfig | null>(null)
  const [workflowModalVisible, setWorkflowModalVisible] = useState(false)
  const [currentPipeline, setCurrentPipeline] = useState<BambooPipeline | undefined>()
  const [form] = Form.useForm()

  useEffect(() => {
    fetchWebhooks()
  }, [page, pageSize])

  const fetchWebhooks = async () => {
    setLoading(true)
    try {
      const response: any = await webhookService.getWebhooks({ page, page_size: pageSize })
      setWebhooks(response.webhooks || [])
      setTotal(response.total || 0)
    } catch (error) {
      message.error('获取Webhook配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingWebhook(null)
    form.resetFields()
    setCurrentPipeline(undefined)
    setIsModalVisible(true)
  }

  const handleEdit = (record: WebhookConfig) => {
    setEditingWebhook(record)
    form.setFieldsValue(record)

    // 如果有 workflow 配置，解析它
    if (record.workflow?.pipeline) {
      try {
        const pipeline = JSON.parse(record.workflow.pipeline)
        setCurrentPipeline(pipeline)
      } catch (error) {
        console.error('Failed to parse pipeline:', error)
      }
    }

    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await webhookService.deleteWebhook(id)
      message.success('删除成功')
      fetchWebhooks()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      // 如果是 inbound，不需要 webhook_url
      if (values.direction === 'inbound') {
        delete values.webhook_url
      } else {
        if (!values.webhook_url) {
          message.error('请输入 Webhook URL')
          return
        }
      }

      // 如果配置了工作流，先创建/更新 Workflow
      let workflowId = values.workflow_id
      if (currentPipeline) {
        const workflowData = {
          name: values.name + ' - 工作流',
          description: values.description || '',
          direction: values.direction,
          type: values.type,
          pipeline: JSON.stringify(currentPipeline),
          enabled: true,
        }

        if (editingWebhook && editingWebhook.workflow_id) {
          // 更新现有工作流
          await workflowService.updateWorkflow(editingWebhook.workflow_id, workflowData)
          workflowId = editingWebhook.workflow_id
        } else {
          // 创建新工作流
          const workflow: any = await workflowService.createWorkflow(workflowData)
          workflowId = workflow.id
        }
      }

      // 准备 Webhook 数据
      const webhookData: any = {
        name: values.name,
        direction: values.direction,
        type: values.type,
        webhook_url: values.webhook_url,
        enabled: values.enabled,
        description: values.description,
      }

      // 如果有 workflow_id，添加到数据中
      if (workflowId) {
        webhookData.workflow_id = workflowId
      }

      if (editingWebhook) {
        await webhookService.updateWebhook(editingWebhook.id, webhookData)
        message.success('更新成功')
      } else {
        // 创建时必须有 workflow_id（如果配置了工作流）
        if (currentPipeline && !workflowId) {
          message.error('工作流创建失败')
          return
        }
        await webhookService.createWebhook(webhookData)
        message.success('创建成功')
      }

      setIsModalVisible(false)
      fetchWebhooks()
    } catch (error) {
      console.error('Submit error:', error)
      message.error('操作失败')
    }
  }

  const generateInboundUrl = (token: string) => {
    // TODO: 从配置读取基础 URL
    const baseUrl = window.location.origin
    return `${baseUrl}/api/v1/webhooks/${token}`
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
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
      title: '方向',
      dataIndex: 'direction',
      width: 100,
      render: (direction: string) => {
        if (direction === 'inbound') {
          return <Tag color="green">接收</Tag>
        } else {
          return <Tag color="blue">推送</Tag>
        }
      },
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
          workflow: { text: '自定义工作流', color: 'purple' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'URL / Token',
      dataIndex: 'webhook_url',
      render: (_: any, record: WebhookConfig) => {
        if (record.direction === 'inbound') {
          const url = generateInboundUrl(record.webhook_token || '')
          return (
            <div>
              <div style={{ fontSize: 11, color: '#999', marginBottom: 4 }}>
                接收地址:
              </div>
              <code style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                {url}
              </code>
              <Button
                size="small"
                type="link"
                icon={<CopyOutlined />}
                onClick={() => copyToClipboard(url)}
              >
                复制
              </Button>
            </div>
          )
        } else {
          return (
            <code style={{ fontSize: 12 }}>
              {record.webhook_url?.substring(0, 50)}
              {record.webhook_url && record.webhook_url.length > 50 ? '...' : ''}
            </code>
          )
        }
      },
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
          <ApiOutlined style={{ marginRight: 8 }} />
          Webhook 集成配置
        </Title>
        <Paragraph>
          配置 Webhook 集成，支持接收外部告警（Inbound）和推送到外部系统（Outbound）。
          可与工作流引擎配合实现复杂的告警处理逻辑。
        </Paragraph>

        <Alert
          message="功能说明"
          description={
            <div>
              <p>• <strong>Inbound（接收）</strong>：自动生成 Webhook URL，接收 Alertmanager/Prometheus/VictoriaMetrics 的告警</p>
              <p>• <strong>Outbound（推送）</strong>：将内部告警推送到外部系统</p>
              <p>• <strong>工作流支持</strong>：可配置工作流实现告警过滤、转换、路由等复杂逻辑</p>
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
                <div style={{ fontSize: 32, fontWeight: 600, color: '#1890ff' }}>
                  {webhooks.length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>Webhook总数</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#52c41a' }}>
                  {webhooks.filter((w) => w.enabled).length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>已启用</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#faad14' }}>
                  {webhooks.filter((w) => !w.enabled).length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>已禁用</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 32, fontWeight: 600, color: '#13c2c2' }}>
                  {webhooks.filter((w) => w.direction === 'inbound').length}
                </div>
                <div style={{ color: '#999', marginTop: 8 }}>Inbound</div>
              </div>
            </Card>
          </Col>
        </Row>

        <Divider />

        <div
          style={{
            marginBottom: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Title level={5} style={{ margin: 0 }}>
            Webhook 配置列表
          </Title>
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
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (newPage, newPageSize) => {
              setPage(newPage)
              setPageSize(newPageSize || 10)
            },
          }}
        />
      </Card>

      <Modal
        title={editingWebhook ? '编辑 Webhook' : '新增 Webhook'}
        open={isModalVisible}
        onOk={handleSubmit}
        onCancel={() => setIsModalVisible(false)}
        width={800}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="名称"
                name="name"
                rules={[{ required: true, message: '请输入名称' }]}
              >
                <Input placeholder="请输入Webhook名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="方向"
                name="direction"
                rules={[{ required: true, message: '请选择方向' }]}
              >
                <Select
                  placeholder="选择方向"
                  onChange={() => {
                    form.setFieldsValue({ webhook_url: undefined })
                  }}
                >
                  <Select.Option value="inbound">Inbound（接收外部告警）</Select.Option>
                  <Select.Option value="outbound">Outbound（推送到外部系统）</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="请选择Webhook类型">
              <Select.Option value="alertmanager">Alertmanager</Select.Option>
              <Select.Option value="prometheus">Prometheus</Select.Option>
              <Select.Option value="victoriametrics">VictoriaMetrics</Select.Option>
              <Select.Option value="workflow">自定义工作流</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.direction !== curr.direction}>
            {({ getFieldValue }) => {
              const direction = getFieldValue('direction') as string

              if (direction === 'inbound') {
                return (
                  <Form.Item label="接收 URL">
                    <Input
                      disabled
                      placeholder="保存后自动生成"
                      value={editingWebhook?.webhook_token ? generateInboundUrl(editingWebhook.webhook_token) : '保存后自动生成'}
                    />
                  </Form.Item>
                )
              } else {
                return (
                  <Form.Item
                    label="目标 URL"
                    name="webhook_url"
                    rules={[{ required: true, message: '请输入Webhook URL' }]}
                  >
                    <TextArea
                      rows={2}
                      placeholder="例如: http://alertmanager:9093/api/v1/alerts"
                    />
                  </Form.Item>
                )
              }
            }}
          </Form.Item>

          <Form.Item label="描述" name="description">
            <TextArea rows={2} placeholder="请输入描述信息（可选）" />
          </Form.Item>

          <Form.Item
            label="启用状态"
            name="enabled"
            initialValue={true}
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <Form.Item label="工作流配置">
            <Button
              onClick={() => setWorkflowModalVisible(true)}
              icon={<SettingOutlined />}
            >
              {currentPipeline ? '编辑工作流' : '配置工作流'}
            </Button>
            {currentPipeline && (
              <Tag color="purple" style={{ marginLeft: 8 }}>
                已配置工作流
              </Tag>
            )}
          </Form.Item>
        </Form>
      </Modal>

      {/* 工作流编辑器 Modal */}
      <Modal
        title="配置工作流"
        open={workflowModalVisible}
        onCancel={() => setWorkflowModalVisible(false)}
        onOk={() => {
          // 保存工作流配置到表单
          if (currentPipeline) {
            form.setFieldsValue({
              workflow: {
                pipeline: JSON.stringify(currentPipeline),
                enabled: true,
              },
            })
          }
          setWorkflowModalVisible(false)
        }}
        width={1200}
        destroyOnClose
        okText="保存工作流"
      >
        <WorkflowEditor
          pipeline={currentPipeline}
          onChange={(pipeline) => setCurrentPipeline(pipeline)}
          onSave={(pipeline) => {
            setCurrentPipeline(pipeline)
            form.setFieldsValue({
              workflow: {
                pipeline: JSON.stringify(pipeline),
                enabled: true,
              },
            })
          }}
        />
      </Modal>
    </div>
  )
}
