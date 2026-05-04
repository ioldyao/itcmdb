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
  PlayCircleOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertService, AlertRule } from '@/services/alertService'

const { TextArea } = Input

const severityMap: Record<string, { text: string; color: string }> = {
  critical: { text: '致命', color: 'red' },
  high: { text: '高', color: 'orange' },
  medium: { text: '中', color: 'gold' },
  low: { text: '低', color: 'blue' },
}

export default function AlertRulesTab() {
  const [loading, setLoading] = useState(false)
  const [rules, setRules] = useState<AlertRule[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchRules()
  }, [page, pageSize])

  const fetchRules = async () => {
    setLoading(true)
    try {
      const response = await alertService.getRules({ page, page_size: pageSize })
      setRules(response.data.rules)
      setTotal(response.data.total)
    } catch (error) {
      message.error('获取告警规则失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingRule(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: AlertRule) => {
    setEditingRule(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertService.deleteRule(id)
      message.success('删除成功')
      fetchRules()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleToggleEnabled = async (record: AlertRule) => {
    try {
      if (record.enabled) {
        await alertService.disableRule(record.id)
        message.success('已禁用')
      } else {
        await alertService.enableRule(record.id)
        message.success('已启用')
      }
      fetchRules()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRule) {
        await alertService.updateRule(editingRule.id, values)
        message.success('更新成功')
      } else {
        await alertService.createRule(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchRules()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<AlertRule> = [
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
      title: '严重级别',
      dataIndex: 'severity',
      width: 100,
      render: (severity: string) => {
        const s = severityMap[severity] || { text: severity, color: 'default' }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '阈值',
      width: 150,
      render: (_, record) => `${record.threshold_operator} ${record.threshold_value}`,
    },
    {
      title: '持续时间',
      dataIndex: 'duration',
      width: 100,
      render: (duration: number) => `${duration}s`,
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
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={record.enabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            onClick={() => handleToggleEnabled(record)}
          >
            {record.enabled ? '禁用' : '启用'}
          </Button>
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
        <h3 className="text-base font-medium text-gray-900 dark:text-text-primary">告警规则</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          创建规则
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
        title={editingRule ? '编辑告警规则' : '创建告警规则'}
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

          <Form.Item name="metric_query" label="监控指标查询" rules={[{ required: true, message: '请输入监控指标查询' }]}>
            <TextArea rows={3} placeholder="例如: node_cpu_usage > 80" />
          </Form.Item>

          <Form.Item label="触发条件">
            <Space.Compact className="w-full">
              <Form.Item name="threshold_operator" noStyle rules={[{ required: true }]}>
                <Select className="w-[30%]" placeholder="运算符">
                  <Select.Option value=">">{'>'}</Select.Option>
                  <Select.Option value="<">{'<'}</Select.Option>
                  <Select.Option value=">=">{'>='}</Select.Option>
                  <Select.Option value="<=">{'<='}</Select.Option>
                  <Select.Option value="==">{'=='}</Select.Option>
                  <Select.Option value="!=">{'!='}</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item name="threshold_value" noStyle rules={[{ required: true }]}>
                <InputNumber className="w-[70%]" placeholder="阈值" />
              </Form.Item>
            </Space.Compact>
          </Form.Item>

          <Form.Item name="duration" label="持续时间(秒)" rules={[{ required: true, message: '请输入持续时间' }]}>
            <InputNumber className="w-full" placeholder="300" min={0} />
          </Form.Item>

          <Form.Item name="severity" label="严重级别" rules={[{ required: true, message: '请选择严重级别' }]}>
            <Select placeholder="请选择严重级别">
              <Select.Option value="critical">致命</Select.Option>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="enabled" label="启用状态" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
