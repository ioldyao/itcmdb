import { useEffect, useState } from 'react'
import {
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
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  StarOutlined,
  StarFilled,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertTemplateService, AlertNotificationTemplate } from '@/services/alertTemplateService'

const { TextArea } = Input

export default function NotificationTemplatesTab() {
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<AlertNotificationTemplate[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [previewVisible, setPreviewVisible] = useState(false)
  const [previewContent, setPreviewContent] = useState('')
  const [editingTemplate, setEditingTemplate] = useState<AlertNotificationTemplate | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchTemplates()
  }, [page, pageSize])

  const fetchTemplates = async () => {
    setLoading(true)
    try {
      const response = await alertTemplateService.getTemplates({ page, page_size: pageSize })
      setTemplates(response.data.templates)
      setTotal(response.data.total)
    } catch (error) {
      message.error('获取通知模板失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingTemplate(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: AlertNotificationTemplate) => {
    setEditingTemplate(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertTemplateService.deleteTemplate(id)
      message.success('删除成功')
      fetchTemplates()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSetDefault = async (id: number) => {
    try {
      await alertTemplateService.setDefaultTemplate(id)
      message.success('设置默认模板成功')
      fetchTemplates()
    } catch (error) {
      message.error('设置失败')
    }
  }

  const handlePreview = async (record: AlertNotificationTemplate) => {
    try {
      const response = await alertTemplateService.previewTemplate({
        template_content: record.template_content,
        template_type: record.template_type,
      })
      setPreviewContent(response.data.preview)
      setPreviewVisible(true)
    } catch (error) {
      message.error('预览失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingTemplate) {
        await alertTemplateService.updateTemplate(editingTemplate.id, values)
        message.success('更新成功')
      } else {
        await alertTemplateService.createTemplate(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchTemplates()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<AlertNotificationTemplate> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '模板名称',
      dataIndex: 'name',
      width: 200,
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '模板类型',
      dataIndex: 'template_type',
      width: 120,
      render: (type: string) => {
        const typeMap: Record<string, { label: string; color: string }> = {
          dingtalk: { label: '钉钉', color: 'blue' },
          feishu: { label: '飞书', color: 'green' },
          wechat: { label: '企业微信', color: 'cyan' },
          email: { label: '邮件', color: 'orange' },
        }
        const config = typeMap[type] || { label: type, color: 'default' }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '默认模板',
      dataIndex: 'is_default',
      width: 100,
      render: (isDefault: boolean) =>
        isDefault ? <StarFilled className="text-yellow-500" /> : null,
    },
    {
      title: '操作',
      width: 250,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handlePreview(record)}
          >
            预览
          </Button>
          {!record.is_default && (
            <Button
              type="link"
              size="small"
              icon={<StarOutlined />}
              onClick={() => handleSetDefault(record.id)}
            >
              设为默认
            </Button>
          )}
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除吗？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="dark:text-text-primary">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-medium text-gray-900 dark:text-text-primary">通知模板</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          创建模板
        </Button>
      </div>

      <Table
        loading={loading}
        columns={columns}
        dataSource={templates}
        rowKey="id"
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps || 10)
          },
        }}
      />

      <Modal
        title={editingTemplate ? '编辑通知模板' : '创建通知模板'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="模板名称" rules={[{ required: true, message: '请输入模板名称' }]}>
            <Input placeholder="请输入模板名称" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="请输入描述" />
          </Form.Item>

          <Form.Item name="template_type" label="模板类型" rules={[{ required: true, message: '请选择模板类型' }]}>
            <Select placeholder="请选择模板类型">
              <Select.Option value="dingtalk">钉钉</Select.Option>
              <Select.Option value="feishu">飞书</Select.Option>
              <Select.Option value="wechat">企业微信</Select.Option>
              <Select.Option value="email">邮件</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="template_content"
            label="模板内容"
            rules={[{ required: true, message: '请输入模板内容' }]}
            extra="支持Go模板语法，可用变量: .AlertID, .Title, .Content, .Severity, .Status, .Instance, .Labels, .Timestamp"
          >
            <TextArea
              rows={10}
              placeholder="请输入模板内容"
              className="font-mono"
            />
          </Form.Item>

          <Form.Item name="is_default" label="设为默认模板" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="模板预览"
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={[
          <Button key="close" onClick={() => setPreviewVisible(false)}>
            关闭
          </Button>,
        ]}
        width={700}
      >
        <pre className="whitespace-pre-wrap break-words text-sm">{previewContent}</pre>
      </Modal>
    </div>
  )
}
