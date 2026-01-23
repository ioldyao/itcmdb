import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Tag, Space, Modal, Form, message, Popconfirm } from 'antd'
import { SearchOutlined, PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, Network as NetworkIcon } from 'lucide-react'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useCMDBStore, CIInstance } from '@/stores/cmdbStore'

export default function CMDBNetworks() {
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
  const [deviceTypeFilter, setDeviceTypeFilter] = useState<string>('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingInstance, setEditingInstance] = useState<CIInstance | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // 假设网络设备类型的 ci_type_id 是 2
    fetchInstances(2, page, pageSize)
  }, [fetchInstances, page, pageSize])

  // 状态映射
  const statusConfig = {
    active: { text: '在线', color: 'green' },
    inactive: { text: '离线', color: 'gray' },
    maintenance: { text: '维护', color: 'orange' },
    decommissioned: { text: '退役', color: 'red' },
  }

  // 设备类型映射
  const deviceTypeConfig = {
    router: { text: '路由器', color: 'blue' },
    switch: { text: '交换机', color: 'cyan' },
    firewall: { text: '防火墙', color: 'red' },
    load_balancer: { text: '负载均衡', color: 'purple' },
    wireless: { text: '无线设备', color: 'orange' },
  }

  const handleSearch = () => {
    setFilters({ name: searchText, status: statusFilter, device_type: deviceTypeFilter })
  }

  const handleReset = () => {
    setSearchText('')
    setStatusFilter('')
    setDeviceTypeFilter('')
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
      device_type: record.attributes?.device_type || '',
      management_ip: record.attributes?.management_ip || '',
      vendor: record.attributes?.vendor || '',
      model: record.attributes?.model || '',
      port_count: record.attributes?.port_count || '',
      firmware_version: record.attributes?.firmware_version || '',
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
        device_type: values.device_type,
        management_ip: values.management_ip,
        vendor: values.vendor,
        model: values.model,
        port_count: values.port_count,
        firmware_version: values.firmware_version,
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
          ci_type_id: 2, // 网络设备类型
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
      title: '设备名称',
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
      title: '设备类型',
      dataIndex: 'attributes',
      key: 'device_type',
      render: (attr: Record<string, any>) => {
        const type = attr?.device_type
        const config = deviceTypeConfig[type as keyof typeof deviceTypeConfig]
        return <Tag color={config?.color}>{config?.text || type}</Tag>
      },
    },
    {
      title: '管理IP',
      dataIndex: 'attributes',
      key: 'management_ip',
      render: (attr: Record<string, any>) => attr?.management_ip || '-',
    },
    {
      title: '厂商',
      dataIndex: 'attributes',
      key: 'vendor',
      render: (attr: Record<string, any>) => attr?.vendor || '-',
    },
    {
      title: '型号',
      dataIndex: 'attributes',
      key: 'model',
      render: (attr: Record<string, any>) => attr?.model || '-',
    },
    {
      title: '端口数',
      dataIndex: 'attributes',
      key: 'port_count',
      render: (attr: Record<string, any>) => attr?.port_count || '-',
    },
    {
      title: '固件版本',
      dataIndex: 'attributes',
      key: 'firmware_version',
      render: (attr: Record<string, any>) => attr?.firmware_version || '-',
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
    fetchInstances(2, pagination.current || 1, pagination.pageSize || 20)
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">网络设备管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理和监控所有网络设备资产</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex flex-wrap gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Input
          placeholder="搜索设备名称..."
          prefix={<SearchOutlined size={16} />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onPressEnter={handleSearch}
          className="w-64"
          allowClear
        />
        <Select
          placeholder="设备类型"
          value={deviceTypeFilter}
          onChange={setDeviceTypeFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="router">路由器</Select.Option>
          <Select.Option value="switch">交换机</Select.Option>
          <Select.Option value="firewall">防火墙</Select.Option>
          <Select.Option value="load_balancer">负载均衡</Select.Option>
          <Select.Option value="wireless">无线设备</Select.Option>
        </Select>
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
        <Button icon={<ReloadOutlined size={16} />} onClick={() => fetchInstances(2, page, pageSize)}>
          刷新
        </Button>
        <Button type="primary" icon={<PlusOutlined size={16} />} onClick={handleCreate}>
          添加设备
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
        title={editingInstance ? '编辑网络设备' : '添加网络设备'}
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
            ci_type_id: 2,
            device_type: '',
            management_ip: '',
            vendor: '',
            model: '',
            port_count: '',
            firmware_version: '',
          }}
        >
          <Form.Item
            label="设备名称"
            name="name"
            rules={[{ required: true, message: '请输入设备名称' }]}
          >
            <Input placeholder="例如: core-switch-01" prefix={<NetworkIcon size={16} />} />
          </Form.Item>
          <Form.Item
            label="设备类型"
            name="device_type"
            rules={[{ required: true, message: '请选择设备类型' }]}
          >
            <Select placeholder="选择设备类型">
              <Select.Option value="router">路由器</Select.Option>
              <Select.Option value="switch">交换机</Select.Option>
              <Select.Option value="firewall">防火墙</Select.Option>
              <Select.Option value="load_balancer">负载均衡</Select.Option>
              <Select.Option value="wireless">无线设备</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            label="管理IP"
            name="management_ip"
            rules={[
              { required: true, message: '请输入管理IP' },
              { pattern: /^(?:[0-9]{1,3}\\.){3}[0-9]$/, message: '请输入有效的IP地址' },
            ]}
          >
            <Input placeholder="例如: 192.168.1.1" />
          </Form.Item>
          <Form.Item label="厂商" name="vendor" rules={[{ required: true, message: '请输入厂商' }]}>
            <Select placeholder="选择厂商" showSearch allowClear>
              <Select.Option value="Cisco">Cisco</Select.Option>
              <Select.Option value="Huawei">Huawei</Select.Option>
              <Select.Option value="H3C">H3C</Select.Option>
              <Select.Option value="Juniper">Juniper</Select.Option>
              <Select.Option value="Arista">Arista</Select.Option>
              <Select.Option value="Fortinet">Fortinet</Select.Option>
              <Select.Option value="Ruijie">Ruijie</Select.Option>
              <Select.Option value="Other">其他</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="型号" name="model" rules={[{ required: true, message: '请输入型号' }]}>
            <Input placeholder="例如: WS-C3850-48T" />
          </Form.Item>
          <Form.Item
            label="端口数量"
            name="port_count"
            rules={[{ required: true, message: '请输入端口数量' }]}
          >
            <Input type="number" placeholder="例如: 48" min={1} />
          </Form.Item>
          <Form.Item label="固件版本" name="firmware_version">
            <Input placeholder="例如: 16.9.5" />
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
