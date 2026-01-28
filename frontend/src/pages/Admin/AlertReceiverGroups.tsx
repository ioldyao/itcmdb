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
  Transfer,
  type TransferProps,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertReceiverService, AlertReceiverGroup, AlertReceiver } from '@/services/alertReceiverService'

export default function AlertReceiverGroups() {
  const [loading, setLoading] = useState(false)
  const [groups, setGroups] = useState<AlertReceiverGroup[]>([])
  const [allReceivers, setAllReceivers] = useState<AlertReceiver[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingGroup, setEditingGroup] = useState<AlertReceiverGroup | null>(null)
  const [selectedReceiverIds, setSelectedReceiverIds] = useState<number[]>([])
  const [form] = Form.useForm()

  useEffect(() => {
    fetchGroups()
    fetchAllReceivers()
  }, [page, pageSize])

  const fetchGroups = async () => {
    setLoading(true)
    try {
      const response = await alertReceiverService.getReceiverGroups({ page, page_size: pageSize })
      if (response.data) {
        setGroups(response.data.groups)
        setTotal(response.data.total)
      }
    } catch (error) {
      message.error('获取接收组列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchAllReceivers = async () => {
    try {
      const response = await alertReceiverService.getReceivers({ page: 1, page_size: 1000 })
      if (response.data) {
        setAllReceivers(response.data.receivers)
      }
    } catch (error) {
      console.error('获取接收人列表失败', error)
    }
  }

  const handleCreate = () => {
    setEditingGroup(null)
    setSelectedReceiverIds([])
    form.resetFields()
    setIsModalVisible(true)
  }

  const handleEdit = (record: AlertReceiverGroup) => {
    setEditingGroup(record)
    const receiverIds = record.receivers?.map(r => r.id) || []
    setSelectedReceiverIds(receiverIds)
    form.setFieldsValue({
      name: record.name,
      description: record.description,
      enabled: record.enabled,
    })
    setIsModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertReceiverService.deleteReceiverGroup(id)
      message.success('删除成功')
      fetchGroups()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data = {
        ...values,
        receiver_ids: selectedReceiverIds,
      }
      if (editingGroup) {
        await alertReceiverService.updateReceiverGroup(editingGroup.id, data)
        message.success('更新成功')
      } else {
        await alertReceiverService.createReceiverGroup(data)
        message.success('创建成功')
      }
      setIsModalVisible(false)
      fetchGroups()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleTransferChange: TransferProps['onChange'] = (newTargetKeys) => {
    setSelectedReceiverIds(newTargetKeys.map(id => Number(id)))
  }

  const columns: ColumnsType<AlertReceiverGroup> = [
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
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '接收人',
      dataIndex: 'receivers',
      width: 300,
      render: (receivers: AlertReceiver[]) => {
        if (!receivers || receivers.length === 0) {
          return <span style={{ color: '#999' }}>-</span>
        }
        return (
          <Space size={4} wrap>
            {receivers.map(r => (
              <Tag key={r.id} color="blue">{r.name}</Tag>
            ))}
          </Space>
        )
      },
    },
    {
      title: '接收人数量',
      dataIndex: 'receivers',
      width: 120,
      render: (receivers: AlertReceiver[]) => (
        <span>{receivers?.length || 0}</span>
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
      width: 150,
      fixed: 'right',
      render: (_: any, record: AlertReceiverGroup) => (
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
            title="确认删除此接收组？"
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

  const transferDataSource = allReceivers.map(receiver => ({
    key: receiver.id.toString(),
    title: receiver.name,
    description: `${receiver.type} - ${receiver.webhook_url ? '已配置' : '未配置'}`,
    chosen: selectedReceiverIds.includes(receiver.id),
  }))

  return (
    <div>
      <Card
        title="告警接收组配置"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新增接收组
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={groups}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1400 }}
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
        title={editingGroup ? '编辑接收组' : '新增接收组'}
        open={isModalVisible}
        onOk={handleSubmit}
        onCancel={() => setIsModalVisible(false)}
        width={800}
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
                label="组名称"
                name="name"
                rules={[{ required: true, message: '请输入组名称' }]}
              >
                <Input placeholder="请输入组名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="启用状态"
                name="enabled"
                initialValue={true}
                valuePropName="checked"
              >
                <Switch checkedChildren="启用" unCheckedChildren="禁用" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="描述"
            name="description"
          >
            <Input.TextArea
              rows={3}
              placeholder="请输入组描述"
            />
          </Form.Item>

          <Form.Item label="关联接收人">
            <Transfer
              dataSource={transferDataSource}
              titles={['可用接收人', '已选接收人']}
              targetKeys={selectedReceiverIds.map(id => id.toString())}
              onChange={handleTransferChange}
              render={item => item.title}
              listStyle={{
                width: 300,
                height: 400,
              }}
              showSearch
              filterOption={(inputValue, item) =>
                item.title?.indexOf(inputValue) !== -1 ||
                item.description?.indexOf(inputValue) !== -1
              }
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
