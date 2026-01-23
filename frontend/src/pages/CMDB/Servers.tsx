import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Tag, Space, Modal, Form, message, Popconfirm } from 'antd'
import { SearchOutlined, PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, Server as ServerIcon } from 'lucide-react'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useCMDBStore, CIInstance } from '@/stores/cmdbStore'

export default function CMDBServers() {
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
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingInstance, setEditingInstance] = useState<CIInstance | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // 假设服务器类型的 ci_type_id 是 1
    fetchInstances(1, page, pageSize)
  }, [fetchInstances, page, pageSize])

  // 状态映射
  const statusConfig = {
    active: { text: '在线', color: 'green' },
    inactive: { text: '离线', color: 'gray' },
    maintenance: { text: '维护', color: 'orange' },
    decommissioned: { text: '退役', color: 'red' },
  }

  const handleSearch = () => {
    setFilters({ name: searchText, status: statusFilter })
  }

  const handleReset = () => {
    setSearchText('')
    setStatusFilter('')
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
      ip_address: record.attributes?.ip_address || '',
      os: record.attributes?.os || '',
      cpu_cores: record.attributes?.cpu_cores || '',
      memory_gb: record.attributes?.memory_gb || '',
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
        ip_address: values.ip_address,
        os: values.os,
        cpu_cores: values.cpu_cores,
        memory_gb: values.memory_gb,
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
          ci_type_id: 1, // 服务器类型
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
      title: 'IP地址',
      dataIndex: 'attributes',
      key: 'ip_address',
      render: (attr: Record<string, any>) => attr?.ip_address || '-',
    },
    {
      title: '操作系统',
      dataIndex: 'attributes',
      key: 'os',
      render: (attr: Record<string, any>) => attr?.os || '-',
    },
    {
      title: 'CPU核数',
      dataIndex: 'attributes',
      key: 'cpu_cores',
      render: (attr: Record<string, any>) => attr?.cpu_cores || '-',
    },
    {
      title: '内存',
      dataIndex: 'attributes',
      key: 'memory_gb',
      render: (attr: Record<string, any>) => attr?.memory_gb ? `${attr.memory_gb} GB` : '-',
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

  const handleTableChange: TablePaginationConfig['onChange'] = (pagination) => {
    fetchInstances(1, pagination.current || 1, pagination.pageSize || 20)
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">服务器管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理和监控所有服务器资产</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex flex-wrap gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Input
          placeholder="搜索服务器名称..."
          prefix={<SearchOutlined size={16} />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onPressEnter={handleSearch}
          className="w-64"
          allowClear
        />
        <Select
          placeholder="状态"
          value={statusFilter}
          onChange={setStatusFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="active">在线</Select.Option>
          <Select.Option value="inactive">离线</Select.Option>
          <Select.Option value="maintenance">维护</Select.Option>
          <Select.Option value="decommissioned">退役</Select.Option>
        </Select>
        <Button onClick={handleSearch} icon={<SearchOutlined size={16} />}>
          搜索
        </Button>
        <Button onClick={handleReset}>重置</Button>
        <div className="flex-1" />
        <Button icon={<ReloadOutlined size={16} />} onClick={() => fetchInstances(1, page, pageSize)}>
          刷新
        </Button>
        <Button type="primary" icon={<PlusOutlined size={16} />} onClick={handleCreate}>
          添加服务器
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
        title={editingInstance ? '编辑服务器' : '添加服务器'}
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
          ci_type_id: 1,
          ip_address: '',
            os: '',
            cpu_cores: '',
            memory_gb: '',
          }}
        >
          <Form.Item
            label="服务器名称"
            name="name"
            rules={[{ required: true, message: '请输入服务器名称' }]}
          >
            <Input placeholder="例如: web-server-01" prefix={<ServerIcon size={16} />} />
          </Form.Item>
          <Form.Item
            label="IP地址"
            name="ip_address"
            rules={[
              { required: true, message: '请输入IP地址' },
              { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]$/, message: '请输入有效的IP地址' },
            ]}
          >
            <Input placeholder="例如: 192.168.1.100" />
          </Form.Item>
          <Form.Item label="操作系统" name="os" rules={[{ required: true, message: '请输入操作系统' }]}>
            <Input placeholder="例如: CentOS 7.9" />
          </Form.Item>
          <Form.Item
            label="CPU核数"
            name="cpu_cores"
            rules={[{ required: true, message: '请输入CPU核数' }]}
          >
            <Input type="number" placeholder="例如: 4" min={1} />
          </Form.Item>
          <Form.Item
            label="内存 (GB)"
            name="memory_gb"
            rules={[{ required: true, message: '请输入内存大小' }]}
          >
            <Input type="number" placeholder="例如: 16" min={1} />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="active">在线</Select.Option>
              <Select.Option value="inactive">离线</Select.Option>
              <Select.Option value="maintenance">维护</Select.Option>
              <Select.Option value="decommissioned">退役</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
