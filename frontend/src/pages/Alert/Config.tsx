import { useEffect, useState } from 'react'
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  AutoComplete,
  InputNumber,
  Switch,
  Tag,
  Space,
  message,
  Popconfirm,
  Tabs,
  Empty,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  EnvironmentOutlined,
  BranchesOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertService } from '@/services/alertService'
import Breadcrumb from '@/components/Breadcrumb'

interface SpaceItem {
  id: number
  name: string
  description: string
  roles: { id: number; name: string }[]
  created_at: string
}

interface RouteItem {
  id: number
  field_name: string
  field_value: string
  space_id: number
  space_name: string
  priority: number
  enabled: boolean
}

interface RoleItem {
  id: number
  name: string
}

export default function AlertConfig() {
  return (
    <div className="p-6">
      <Breadcrumb />
      <Card className="dark:bg-bg-secondary dark:border-white/8">
        <div className="mb-6">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">告警配置</h2>
          <p className="text-gray-600 dark:text-text-secondary text-sm">管理告警空间和路由规则，将告警自动分配到对应团队</p>
        </div>

        <Tabs
          defaultActiveKey="spaces"
          items={[
            { key: 'spaces', label: '空间管理', icon: <EnvironmentOutlined />, children: <SpaceManager /> },
            { key: 'routes', label: '路由规则', icon: <BranchesOutlined />, children: <RouteManager /> },
          ]}
        />
      </Card>
    </div>
  )
}

// ============================================
// 空间管理
// ============================================

function SpaceManager() {
  const [spaces, setSpaces] = useState<SpaceItem[]>([])
  const [roles, setRoles] = useState<RoleItem[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingSpace, setEditingSpace] = useState<SpaceItem | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchSpaces()
    fetchRoles()
  }, [])

  const fetchSpaces = async () => {
    setLoading(true)
    try {
      const res = await alertService.getSpaces()
      if (res.code === 0) setSpaces(res.data || [])
    } catch {
      message.error('获取空间列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchRoles = async () => {
    try {
      const res = await alertService.getRoles()
      if (res.code === 0) setRoles(res.data || [])
    } catch {
      message.error('获取角色列表失败，请确认服务已部署')
    }
  }

  const handleCreate = () => {
    setEditingSpace(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (space: SpaceItem) => {
    setEditingSpace(space)
    form.setFieldsValue({
      name: space.name,
      description: space.description,
      role_ids: space.roles.map((r) => r.id),
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertService.deleteSpace(id)
      message.success('删除成功')
      fetchSpaces()
    } catch {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingSpace) {
        await alertService.updateSpace(editingSpace.id, values)
        message.success('更新成功')
      } else {
        await alertService.createSpace(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchSpaces()
    } catch {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<SpaceItem> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
    },
    {
      title: '空间名称',
      dataIndex: 'name',
      width: 200,
      render: (name: string) => <span className="font-medium">{name}</span>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      render: (desc: string) => desc || <span className="text-gray-400">-</span>,
    },
    {
      title: '关联角色',
      dataIndex: 'roles',
      width: 300,
      render: (roles: { id: number; name: string }[]) =>
        roles.length > 0 ? (
          <Space wrap>
            {roles.map((r) => (
              <Tag key={r.id} color="blue">{r.name}</Tag>
            ))}
          </Space>
        ) : (
          <span className="text-gray-400 text-sm">未关联角色</span>
        ),
    },
    {
      title: '操作',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="删除空间将同时删除关联的路由规则，确定？" onConfirm={() => handleDelete(record.id)}>
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
      <div className="mb-4 flex justify-end">
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          添加空间
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={spaces}
        rowKey="id"
        loading={loading}
        pagination={false}
        locale={{ emptyText: <Empty description="暂无空间" /> }}
      />

      <Modal
        title={editingSpace ? '编辑空间' : '添加空间'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={520}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="空间名称" rules={[{ required: true, message: '请输入空间名称' }]}>
            <Input placeholder="如：HPC运维、系统组、网络组" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="空间用途说明" />
          </Form.Item>
          <Form.Item name="role_ids" label="关联角色">
            <Select
              mode="multiple"
              placeholder="选择该空间可接收告警的角色"
              optionFilterProp="label"
              options={roles.map((r) => ({ label: r.name, value: r.id }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

// ============================================
// 路由规则管理
// ============================================

function RouteManager() {
  const [routes, setRoutes] = useState<RouteItem[]>([])
  const [spaces, setSpaces] = useState<SpaceItem[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRoute, setEditingRoute] = useState<RouteItem | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchRoutes()
    fetchSpaces()
  }, [])

  const fetchRoutes = async () => {
    setLoading(true)
    try {
      const res = await alertService.getSpaceRoutes()
      if (res.code === 0) setRoutes(res.data || [])
    } catch {
      message.error('获取路由规则失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchSpaces = async () => {
    try {
      const res = await alertService.getSpaces()
      if (res.code === 0) setSpaces(res.data || [])
    } catch {
      // 静默
    }
  }

  const handleCreate = () => {
    setEditingRoute(null)
    form.resetFields()
    form.setFieldsValue({ priority: 0, enabled: true })
    setModalVisible(true)
  }

  const handleEdit = (route: RouteItem) => {
    setEditingRoute(route)
    form.setFieldsValue(route)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await alertService.deleteSpaceRoute(id)
      message.success('删除成功')
      fetchRoutes()
    } catch {
      message.error('删除失败')
    }
  }

  const handleToggleEnabled = async (route: RouteItem) => {
    try {
      await alertService.updateSpaceRoute(route.id, { enabled: !route.enabled })
      fetchRoutes()
    } catch {
      message.error('操作失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRoute) {
        await alertService.updateSpaceRoute(editingRoute.id, values)
        message.success('更新成功')
      } else {
        await alertService.createSpaceRoute(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchRoutes()
    } catch {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<RouteItem> = [
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 80,
      sorter: (a, b) => a.priority - b.priority,
    },
    {
      title: '匹配字段',
      dataIndex: 'field_name',
      width: 150,
      render: (name: string) => <Tag>{name}</Tag>,
    },
    {
      title: '匹配值',
      dataIndex: 'field_value',
      width: 200,
      render: (val: string) => <span className="font-mono text-sm">{val}</span>,
    },
    {
      title: '推送到空间',
      dataIndex: 'space_name',
      width: 160,
      render: (name: string) => (
        <Tag icon={<EnvironmentOutlined />} color="processing">{name}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 80,
      render: (enabled: boolean, record) => (
        <Switch checked={enabled} size="small" onChange={() => handleToggleEnabled(record)} />
      ),
    },
    {
      title: '操作',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
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
      <div className="mb-4 flex items-center justify-between">
        <p className="text-sm text-gray-500 dark:text-text-secondary">
          按告警标签字段匹配，命中规则的告警将路由到对应空间，并通知该空间关联角色下的所有用户
        </p>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          添加规则
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={routes}
        rowKey="id"
        loading={loading}
        pagination={false}
        locale={{ emptyText: <Empty description="暂无路由规则" /> }}
      />

      <Modal
        title={editingRoute ? '编辑路由规则' : '添加路由规则'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={520}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="field_name"
            label="匹配字段"
            rules={[{ required: true, message: '请输入匹配字段' }]}
            extra="告警标签中的字段名，如 project、alertgroup、env"
          >
            <AutoComplete
              placeholder="输入字段名，支持自定义"
              options={[
                { label: 'project (项目)', value: 'project' },
                { label: 'alertgroup (告警组)', value: 'alertgroup' },
                { label: 'alertname (告警名)', value: 'alertname' },
                { label: 'env (环境)', value: 'env' },
                { label: 'instance (实例)', value: 'instance' },
                { label: 'job (任务)', value: 'job' },
                { label: 'severity (级别)', value: 'severity' },
              ]}
              filterOption={(inputValue, option) =>
                (option?.label ?? '').toLowerCase().includes(inputValue.toLowerCase())
              }
            />
          </Form.Item>
          <Form.Item
            name="field_value"
            label="匹配值"
            rules={[{ required: true, message: '请输入匹配值' }]}
            extra="精确匹配告警标签中该字段的值"
          >
            <Input placeholder="如：HPC1、系统资源告警、production" />
          </Form.Item>
          <Form.Item
            name="space_id"
            label="推送到空间"
            rules={[{ required: true, message: '请选择空间' }]}
          >
            <Select
              placeholder="选择目标空间"
              options={spaces.map((s) => ({ label: s.name, value: s.id }))}
            />
          </Form.Item>
          <Form.Item name="priority" label="优先级">
            <InputNumber min={0} max={999} className="w-full" placeholder="数字越小优先级越高" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
