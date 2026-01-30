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
  InputNumber,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertRoutingService, AlertRoutingRule } from '@/services/alertRoutingService'
import { alertReceiverService, AlertReceiverGroup } from '@/services/alertReceiverService'
import { alertTemplateService, AlertNotificationTemplate } from '@/services/alertTemplateService'

const { TextArea } = Input

export default function RoutingRulesTab() {
  const [loading, setLoading] = useState(false)
  const [rules, setRules] = useState<AlertRoutingRule[]>([])
  const [receiverGroups, setReceiverGroups] = useState<AlertReceiverGroup[]>([])
  const [templates, setTemplates] = useState<AlertNotificationTemplate[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRoutingRule | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchRules()
    fetchReceiverGroups()
    fetchTemplates()
  }, [page, pageSize])

  const fetchRules = async () => {
    setLoading(true)
    try {
      const response = await alertRoutingService.getRoutingRules({ page, page_size: pageSize })
      setRules(response.data.rules)
      setTotal(response.data.total)
    } catch (error) {
      message.error('获取路由规则失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchReceiverGroups = async () => {
    try {
      const response = await alertReceiverService.getReceiverGroups({})
      setReceiverGroups(response.groups)
    } catch (error) {
      console.error('获取接收组失败', error)
    }
  }

  const fetchTemplates = async () => {
    try {
      const response = await alertTemplateService.getTemplates({})
      setTemplates(response.data.templates)
    } catch (error) {
      console.error('获取通知模板失败', error)
    }
  }

  const handleCreate = () => {
    setEditingRule(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: AlertRoutingRule) => {
    setEditingRule(record)
    form.setFieldsValue({
      ...record,
      matchers: JSON.stringify(record.matchers, null, 2),
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertRoutingService.deleteRoutingRule(id)
      message.success('删除成功')
      fetchRules()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data = {
        ...values,
        matchers: JSON.parse(values.matchers),
      }
      if (editingRule) {
        await alertRoutingService.updateRoutingRule(editingRule.id, data)
        message.success('更新成功')
      } else {
        await alertRoutingService.createRoutingRule(data)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchRules()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<AlertRoutingRule> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '规则名称',
      dataIndex: 'name',
      width: 200,
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '匹配类型',
      dataIndex: 'match_type',
      width: 100,
      render: (type: string) => (
        <Tag color={type === 'match' ? 'blue' : 'purple'}>
          {type === 'match' ? '完全匹配' : '正则匹配'}
        </Tag>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 80,
      sorter: (a, b) => a.priority - b.priority,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 80,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      width: 180,
      render: (_, record) => (
        <Space>
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
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h3>路由规则</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          创建路由规则
        </Button>
      </div>

      <Table
        loading={loading}
        columns={columns}
        dataSource={rules}
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
        title={editingRule ? '编辑路由规则' : '创建路由规则'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入规则名称' }]}>
            <Input placeholder="请输入规则名称" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="请输入描述" />
          </Form.Item>

          <Form.Item
            name="matchers"
            label="匹配条件 (JSON格式)"
            rules={[
              { required: true, message: '请输入匹配条件' },
              {
                validator: (_, value) => {
                  try {
                    JSON.parse(value)
                    return Promise.resolve()
                  } catch {
                    return Promise.reject('请输入有效的JSON格式')
                  }
                },
              },
            ]}
          >
            <TextArea
              rows={4}
              placeholder='例如: {"severity": "critical", "env": "production"}'
            />
          </Form.Item>

          <Form.Item name="match_type" label="匹配类型" rules={[{ required: true }]} initialValue="match">
            <Select>
              <Select.Option value="match">完全匹配</Select.Option>
              <Select.Option value="match_re">正则匹配</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="receiver_group_id" label="接收组" rules={[{ required: true, message: '请选择接收组' }]}>
            <Select placeholder="请选择接收组">
              {receiverGroups.map((group) => (
                <Select.Option key={group.id} value={group.id}>
                  {group.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="template_id" label="通知模板（可选）" tooltip="指定此路由规则使用的通知模板，优先级最高">
            <Select placeholder="请选择通知模板（不选则使用接收人默认模板）" allowClear>
              {templates.map((template) => (
                <Select.Option key={template.id} value={template.id}>
                  {template.name} ({template.template_type})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="priority" label="优先级" rules={[{ required: true }]} initialValue={0}>
            <InputNumber style={{ width: '100%' }} placeholder="数字越小优先级越高" />
          </Form.Item>

          <Form.Item name="continue" label="继续匹配" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>

          <Form.Item name="enabled" label="启用状态" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
