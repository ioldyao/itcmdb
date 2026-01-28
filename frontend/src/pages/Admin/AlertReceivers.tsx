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
  Card,
  Row,
  Col,
  Divider,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TestOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertReceiverService, AlertReceiver } from '@/services/alertReceiverService'

export default function AlertReceivers() {
  const [loading, setLoading] = useState(false)
  const [receivers, setReceivers] = useState<AlertReceiver[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingReceiver, setEditingReceiver] = useState<AlertReceiver | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchReceivers()
  }, [page, pageSize])

  const fetchReceivers = async () => {
    setLoading(true)
    try {
      const response = await alertReceiverService.getReceivers({ page, page_size: pageSize })
      if (response.data) {
        setReceivers(response.data.receivers)
        setTotal(response.data.total)
      }
    } catch (error) {
      message.error('获取接收人列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingReceiver(null)
    form.resetFields()
    setIsModalVisible(true)
  }

  const handleEdit = (record: AlertReceiver) => {
    setEditingReceiver(record)
    form.setFieldsValue(record)
    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertReceiverService.deleteReceiver(id)
      message.success('删除成功')
      fetchReceivers()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleTest = async (id: number) => {
    try {
      const response = await alertReceiverService.testReceiver(id)
      if (response.data?.success) {
        message.success('测试消息发送成功')
      } else {
        message.error(response.data?.error || '测试消息发送失败')
      }
    } catch (error: any) {
      message.error(error.response?.data?.error || '测试消息发送失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingReceiver) {
        await alertReceiverService.updateReceiver(editingReceiver.id, values)
        message.success('更新成功')
      } else {
        await alertReceiverService.createReceiver(values)
        message.success('创建成功')
      }
      setIsModalVisible(false)
      fetchReceivers()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const receiverTypeOptions = [
    { label: '企业微信', value: 'wechat' },
    { label: '钉钉', value: 'dingtalk' },
    { label: '飞书', value: 'feishu' },
    { label: '邮件', value: 'email' },
    { label: '短信', value: 'sms' },
  ]

  const columns: ColumnsType<AlertReceiver> = [
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
      width: 120,
      render: (type: string) => {
        const typeMap: Record<string, { text: string; color: string }> = {
          wechat: { text: '企业微信', color: 'green' },
          dingtalk: { text: '钉钉', color: 'blue' },
          feishu: { text: '飞书', color: 'cyan' },
          email: { text: '邮件', color: 'orange' },
          sms: { text: '短信', color: 'purple' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Webhook URL',
      dataIndex: 'webhook_url',
      ellipsis: true,
      render: (url: string) => (
        <span style={{ fontSize: 12, color: '#999' }}>
          {url ? url.substring(0, 50) + (url.length > 50 ? '...' : '') : '-'}
        </span>
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
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: any, record: AlertReceiver) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<TestOutlined />}
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
            title="确认删除此接收人？"
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
      <Card
        title="告警接收人配置"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新增接收人
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={receivers}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1200 }}
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
        title={editingReceiver ? '编辑接收人' : '新增接收人'}
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
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="名称"
                name="name"
                rules={[{ required: true, message: '请输入名称' }]}
              >
                <Input placeholder="请输入接收人名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="类型"
                name="type"
                rules={[{ required: true, message: '请选择类型' }]}
              >
                <Select placeholder="请选择接收人类型">
                  {receiverTypeOptions.map(option => (
                    <Select.Option key={option.value} value={option.value}>
                      {option.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Webhook URL"
            name="webhook_url"
            rules={[{ required: true, message: '请输入Webhook URL' }]}
          >
            <Input.TextArea
              rows={3}
              placeholder="请输入Webhook地址"
            />
          </Form.Item>

          <Form.Item
            label="签名密钥（可选）"
            name="secret"
          >
            <Input.Password placeholder="钉钉/企业微信签名密钥" />
          </Form.Item>

          <Form.Item
            label="@手机号列表"
            name="at_mobiles"
          >
            <Select
              mode="tags"
              placeholder="输入手机号后回车"
              tokenSeparators={[',', ' ']}
            />
          </Form.Item>

          <Form.Item
            label="@用户ID列表"
            name="at_user_ids"
          >
            <Select
              mode="tags"
              placeholder="输入用户ID后回车"
              tokenSeparators={[',', ' ']}
            />
          </Form.Item>

          <Divider />

          {!editingReceiver && (
            <Form.Item
              label="启用状态"
              name="enabled"
              initialValue={true}
              valuePropName="checked"
            >
              <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            </Form.Item>
          )}

          {editingReceiver && (
            <Form.Item
              label="启用状态"
              name="enabled"
              valuePropName="checked"
            >
              <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
