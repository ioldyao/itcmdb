import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Tag, Space, Modal, Form, message, Popconfirm } from 'antd'
import { SearchOutlined, PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { Container as ContainerIcon } from 'lucide-react'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useCMDBStore, CIInstance } from '@/stores/cmdbStore'

export default function CMDBContainers() {
  const {
    instances,
    total,
    page,
    pageSize,
    loading,
    fetchInstances,
    createInstance,
    updateInstance,
    deleteInstance,
    setFilters,
    resetFilters,
  } = useCMDBStore()

  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingInstance, setEditingInstance] = useState<CIInstance | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // 假设容器类型的 ci_type_id 是 4
    fetchInstances(4, page, pageSize)
  }, [fetchInstances, page, pageSize])

  // 状态映射
  const statusConfig = {
    active: { text: '运行中', color: 'green' },
    inactive: { text: '已停止', color: 'gray' },
    maintenance: { text: '维护中', color: 'orange' },
    decommissioned: { text: '已下线', color: 'red' },
  }

  const handleSearch = () => {
    setFilters({ name: searchText, status: statusFilter, container_type: typeFilter })
  }

  const handleReset = () => {
    setSearchText('')
    setStatusFilter('')
    setTypeFilter('')
    resetFilters()
  }

  const handleCreate = () => {
    setEditingInstance(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = (record: CIInstance) => {
    setEditingInstance(record)
    form.setFieldsValue({
      name: record.name,
      status: record.status,
      container_type: record.attributes?.container_type || '',
      cluster_name: record.attributes?.cluster_name || '',
      namespace: record.attributes?.namespace || '',
      image_name: record.attributes?.image_name || '',
      image_tag: record.attributes?.image_tag || '',
      cpu_limit: record.attributes?.cpu_limit || '',
      memory_limit: record.attributes?.memory_limit || '',
      node_count: record.attributes?.node_count || '',
      container_id: record.attributes?.container_id || '',
      cadvisor_endpoint: record.attributes?.cadvisor_endpoint || '',
      host_id: record.attributes?.host_id || '',
    })
    setIsModalOpen(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteInstance(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const attributes = {
        container_type: values.container_type,
        cluster_name: values.cluster_name,
        namespace: values.namespace,
        image_name: values.image_name,
        image_tag: values.image_tag,
        cpu_limit: values.cpu_limit,
        memory_limit: values.memory_limit,
        node_count: values.node_count,
        // cAdvisor 监控相关
        container_id: values.container_id || '',
        cadvisor_endpoint: values.cadvisor_endpoint || '',
        // 宿主机关联
        host_id: values.host_id || '',
      }

      if (editingInstance) {
        await updateInstance(editingInstance.id, {
          name: values.name,
          status: values.status,
          attributes,
        })
        message.success('更新成功')
      } else {
        await createInstance({
          ci_type_id: 4, // 容器类型
          name: values.name,
          status: values.status,
          attributes,
        })
        message.success('创建成功')
      }
      setIsModalOpen(false)
      form.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<CIInstance> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      key: 'id',
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: CIInstance) => (
        <a
          onClick={() => handleEdit(record)}
          className="text-brand-primary hover:underline cursor-pointer"
        >
          {text}
        </a>
      ),
    },
    {
      title: '容器类型',
      dataIndex: 'ci_type',
      key: 'ci_type',
      render: (ciType: any) => {
        if (!ciType) return '-'
        return <Tag color="blue">{ciType.display_name || ciType.name}</Tag>
      },
    },
    {
      title: '集群名称',
      dataIndex: 'attributes',
      key: 'cluster_name',
      render: (attr: Record<string, any>) => attr?.cluster_name || '-',
    },
    {
      title: '命名空间',
      dataIndex: 'attributes',
      key: 'namespace',
      render: (attr: Record<string, any>) => attr?.namespace || '-',
    },
    {
      title: '镜像',
      dataIndex: 'attributes',
      key: 'image',
      render: (attr: Record<string, any>) => {
        // 优先使用自动发现的 container_image
        const containerImage = attr?.container_image
        if (containerImage) return containerImage

        // 兼容手动创建的容器
        const image = attr?.image_name
        const tag = attr?.image_tag
        if (image && tag) return `${image}:${tag}`
        return image || '-'
      },
    },
    {
      title: 'CPU使用率',
      dataIndex: 'attributes',
      key: 'cpu_usage',
      render: (attr: Record<string, any>) => {
        const cpu = attr?.cpu_usage_percent
        if (cpu !== undefined && cpu !== null) {
          return `${cpu.toFixed(2)}%`
        }
        return '-'
      },
    },
    {
      title: '内存使用',
      dataIndex: 'attributes',
      key: 'memory_usage',
      render: (attr: Record<string, any>) => {
        const usage = attr?.memory_usage_mb
        const limit = attr?.memory_limit_mb
        if (usage !== undefined && usage !== null) {
          if (limit) {
            return `${usage.toFixed(0)}MB / ${limit.toFixed(0)}MB`
          }
          return `${usage.toFixed(0)}MB`
        }
        return '-'
      },
    },
    {
      title: '运行时间',
      dataIndex: 'attributes',
      key: 'uptime',
      render: (attr: Record<string, any>) => {
        const uptime = attr?.uptime_seconds
        if (uptime !== undefined && uptime !== null) {
          const hours = Math.floor(uptime / 3600)
          const minutes = Math.floor((uptime % 3600) / 60)
          if (hours > 24) {
            const days = Math.floor(hours / 24)
            return `${days}天${hours % 24}小时`
          }
          return `${hours}小时${minutes}分钟`
        }
        return '-'
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: keyof typeof statusConfig) => {
        const config = statusConfig[status]
        return <Tag color={config?.color}>{config?.text}</Tag>
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: CIInstance) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined size={14} />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined size={14} />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const handleTableChange = (pagination: TablePaginationConfig) => {
    fetchInstances(4, pagination.current || 1, pagination.pageSize || 20)
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">容器/K8s管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理和监控所有容器和Kubernetes集群</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex flex-wrap gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Input
          placeholder="搜索名称..."
          prefix={<SearchOutlined size={16} />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onPressEnter={handleSearch}
          className="w-64"
          allowClear
        />
        <Select
          placeholder="容器类型"
          value={typeFilter}
          onChange={setTypeFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="docker">Docker</Select.Option>
          <Select.Option value="kubernetes">Kubernetes</Select.Option>
          <Select.Option value="openshift">OpenShift</Select.Option>
          <Select.Option value="ecs">ECS</Select.Option>
        </Select>
        <Select
          placeholder="状态"
          value={statusFilter}
          onChange={setStatusFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="active">运行中</Select.Option>
          <Select.Option value="inactive">已停止</Select.Option>
          <Select.Option value="maintenance">维护中</Select.Option>
          <Select.Option value="decommissioned">已下线</Select.Option>
        </Select>
        <Button onClick={handleSearch} icon={<SearchOutlined size={16} />}>
          搜索
        </Button>
        <Button onClick={handleReset}>重置</Button>
        <div className="flex-1" />
        <Button icon={<ReloadOutlined size={16} />} onClick={() => fetchInstances(4, page, pageSize)}>
          刷新
        </Button>
        <Button type="primary" icon={<PlusOutlined size={16} />} onClick={handleCreate}>
          添加容器
        </Button>
      </div>

      {/* 表格 */}
      <Table
        columns={columns}
        dataSource={instances}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          pageSizeOptions: ['10', '20', '50', '100'],
        }}
        onChange={handleTableChange}
        className="bg-white dark:bg-bg-secondary"
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingInstance ? '编辑容器/K8s' : '添加容器/K8s'}
        open={isModalOpen}
        onCancel={() => {
          setIsModalOpen(false)
          form.resetFields()
        }}
        onOk={handleSubmit}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            status: 'active',
            ci_type_id: 4,
            container_type: 'kubernetes',
            cluster_name: '',
            namespace: 'default',
            image_name: '',
            image_tag: 'latest',
            cpu_limit: '',
            memory_limit: '',
            node_count: '',
            container_id: '',
            cadvisor_endpoint: '',
            host_id: '',
          }}
        >
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="例如: prod-cluster" prefix={<ContainerIcon size={16} />} />
          </Form.Item>
          <Form.Item
            label="容器类型"
            name="container_type"
            rules={[{ required: true, message: '请选择容器类型' }]}
          >
            <Select placeholder="选择容器类型">
              <Select.Option value="docker">Docker</Select.Option>
              <Select.Option value="kubernetes">Kubernetes</Select.Option>
              <Select.Option value="openshift">OpenShift</Select.Option>
              <Select.Option value="ecs">ECS</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="集群名称" name="cluster_name">
            <Input placeholder="例如: production-cluster" />
          </Form.Item>
          <Form.Item label="命名空间" name="namespace">
            <Input placeholder="例如: default" />
          </Form.Item>
          <Form.Item label="镜像名称" name="image_name">
            <Input placeholder="例如: nginx" />
          </Form.Item>
          <Form.Item label="镜像标签" name="image_tag">
            <Input placeholder="例如: latest" />
          </Form.Item>
          <Form.Item label="CPU限制" name="cpu_limit">
            <Input placeholder="例如: 2 cores" />
          </Form.Item>
          <Form.Item label="内存限制" name="memory_limit">
            <Input placeholder="例如: 4Gi" />
          </Form.Item>
          <Form.Item label="节点数量" name="node_count">
            <Input type="number" placeholder="例如: 3" min={1} />
          </Form.Item>

          {/* cAdvisor 监控配置 */}
          <Form.Item label="容器ID" name="container_id">
            <Input placeholder="Docker容器ID，例如: abc123def456" />
          </Form.Item>
          <Form.Item label="cAdvisor端点" name="cadvisor_endpoint">
            <Input placeholder="例如: http://192.168.1.100:8080" />
          </Form.Item>

          {/* 宿主机关联 */}
          <Form.Item label="宿主机ID" name="host_id">
            <Input type="number" placeholder="关联的宿主机CI实例ID" />
          </Form.Item>

          <Form.Item label="状态" name="status" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="active">运行中</Select.Option>
              <Select.Option value="inactive">已停止</Select.Option>
              <Select.Option value="maintenance">维护中</Select.Option>
              <Select.Option value="decommissioned">已下线</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
