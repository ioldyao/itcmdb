import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Tag, Space, Modal, Form, message, Popconfirm } from 'antd'
import { SearchOutlined, PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { Box as BoxIcon } from 'lucide-react'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useCMDBStore, CIInstance } from '@/stores/cmdbStore'
import { useRoleStore } from '@/stores/roleStore'
import { useTagStore } from '@/stores/tagStore'
import { useAuthStore } from '@/stores/authStore'

export default function CMDBApplications() {
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
    fetchInstanceRoles,
    assignInstanceRoles,
    fetchInstanceTags,
    assignInstanceTags,
    setFilters,
    resetFilters,
  } = useCMDBStore()

  const { ciRoles, fetchCIRoles } = useRoleStore()
  const { tags, fetchTags } = useTagStore()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  // 检查权限
  const canCreate = hasPermission('ci', 'create')
  const canUpdate = hasPermission('ci', 'update')
  const canDelete = hasPermission('ci', 'delete')

  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [appTypeFilter, setAppTypeFilter] = useState<string>('')
  const [envFilter, setEnvFilter] = useState<string>('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingInstance, setEditingInstance] = useState<CIInstance | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // 假设应用服务类型的 ci_type_id 是 3
    fetchInstances(3, page, pageSize)
    fetchCIRoles()
    fetchTags()
  }, [fetchInstances, fetchCIRoles, fetchTags, page, pageSize])

  // 状态映射
  const statusConfig = {
    active: { text: '运行中', color: 'green' },
    inactive: { text: '已停止', color: 'gray' },
    maintenance: { text: '维护中', color: 'orange' },
    decommissioned: { text: '已下线', color: 'red' },
  }

  // 应用类型映射
  const appTypeConfig = {
    web: { text: 'Web应用', color: 'blue' },
    database: { text: '数据库', color: 'green' },
    middleware: { text: '中间件', color: 'orange' },
    cache: { text: '缓存', color: 'red' },
    message_queue: { text: '消息队列', color: 'purple' },
    bigdata: { text: '大数据', color: 'cyan' },
    ai: { text: 'AI服务', color: 'magenta' },
  }

  // 环境映射
  const envConfig = {
    dev: { text: '开发', color: 'blue' },
    test: { text: '测试', color: 'orange' },
    staging: { text: '预发', color: 'purple' },
    prod: { text: '生产', color: 'red' },
  }

  const handleSearch = () => {
    setFilters({ name: searchText, status: statusFilter, app_type: appTypeFilter, environment: envFilter })
  }

  const handleReset = () => {
    setSearchText('')
    setStatusFilter('')
    setAppTypeFilter('')
    setEnvFilter('')
    resetFilters()
  }

  const handleCreate = () => {
    setEditingInstance(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = async (record: CIInstance) => {
    setEditingInstance(record)

    const [rolesData, tagsData] = await Promise.all([
      fetchInstanceRoles(record.id),
      fetchInstanceTags(record.id),
    ])

    const roleIds = rolesData.map((r: any) => r.role_id || r.id)
    const tagIds = tagsData.map((t: any) => t.tag_id || t.id)

    form.setFieldsValue({
      name: record.name,
      status: record.status,
      app_type: record.attributes?.app_type || '',
      version: record.attributes?.version || '',
      language: record.attributes?.language || '',
      port: record.attributes?.port || '',
      environment: record.attributes?.environment || '',
      deploy_type: record.attributes?.deploy_type || '',
      repository_url: record.attributes?.repository_url || '',
      roles: roleIds,
      tags: tagIds,
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
        app_type: values.app_type,
        version: values.version,
        language: values.language,
        port: values.port,
        environment: values.environment,
        deploy_type: values.deploy_type,
        repository_url: values.repository_url,
      }

      if (editingInstance) {
        await updateInstance(editingInstance.id, {
          name: values.name,
          status: values.status,
          attributes,
        })

        const roles = values.roles || []
        const tags = values.tags || []
        if (roles.length > 0) await assignInstanceRoles(editingInstance.id, roles)
        if (tags.length > 0) await assignInstanceTags(editingInstance.id, tags)

        message.success('更新成功')
      } else {
        const instance = await createInstance({
          ci_type_id: 3,
          name: values.name,
          status: values.status,
          attributes,
        })

        const roles = values.roles || []
        const tags = values.tags || []
        if (roles.length > 0) await assignInstanceRoles(instance.id, roles)
        if (tags.length > 0) await assignInstanceTags(instance.id, tags)

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
      title: '应用名称',
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
      title: '应用类型',
      dataIndex: 'attributes',
      key: 'app_type',
      render: (attr: Record<string, any>) => {
        const type = attr?.app_type
        const config = appTypeConfig[type as keyof typeof appTypeConfig]
        return <Tag color={config?.color}>{config?.text || type}</Tag>
      },
    },
    {
      title: '版本',
      dataIndex: 'attributes',
      key: 'version',
      render: (attr: Record<string, any>) => attr?.version || '-',
    },
    {
      title: '开发语言',
      dataIndex: 'attributes',
      key: 'language',
      render: (attr: Record<string, any>) => attr?.language || '-',
    },
    {
      title: '端口',
      dataIndex: 'attributes',
      key: 'port',
      render: (attr: Record<string, any>) => attr?.port || '-',
    },
    {
      title: '环境',
      dataIndex: 'attributes',
      key: 'environment',
      render: (attr: Record<string, any>) => {
        const env = attr?.environment
        const config = envConfig[env as keyof typeof envConfig]
        return <Tag color={config?.color}>{config?.text || env}</Tag>
      },
    },
    {
      title: '部署方式',
      dataIndex: 'attributes',
      key: 'deploy_type',
      render: (attr: Record<string, any>) => {
        const type = attr?.deploy_type
        if (type === 'container') return <Tag color="blue">容器</Tag>
        if (type === 'vm') return <Tag color="green">虚拟机</Tag>
        if (type === 'bare_metal') return <Tag color="orange">物理机</Tag>
        return type || '-'
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
          {canUpdate && (
            <Button
              type="link"
              size="small"
              icon={<EditOutlined size={14} />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
          )}
          {canDelete && (
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
          )}
          {!canUpdate && !canDelete && <span style={{ color: '#999' }}>无权限</span>}
        </Space>
      ),
    },
  ]

  const handleTableChange = (pagination: TablePaginationConfig) => {
    fetchInstances(3, pagination.current || 1, pagination.pageSize || 20)
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">应用服务管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理和监控所有应用服务资产</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex flex-wrap gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Input
          placeholder="搜索应用名称..."
          prefix={<SearchOutlined size={16} />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onPressEnter={handleSearch}
          className="w-64"
          allowClear
        />
        <Select
          placeholder="应用类型"
          value={appTypeFilter}
          onChange={setAppTypeFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="web">Web应用</Select.Option>
          <Select.Option value="database">数据库</Select.Option>
          <Select.Option value="middleware">中间件</Select.Option>
          <Select.Option value="cache">缓存</Select.Option>
          <Select.Option value="message_queue">消息队列</Select.Option>
          <Select.Option value="bigdata">大数据</Select.Option>
          <Select.Option value="ai">AI服务</Select.Option>
        </Select>
        <Select
          placeholder="环境"
          value={envFilter}
          onChange={setEnvFilter}
          className="w-32"
          allowClear
        >
          <Select.Option value="">全部</Select.Option>
          <Select.Option value="dev">开发</Select.Option>
          <Select.Option value="test">测试</Select.Option>
          <Select.Option value="staging">预发</Select.Option>
          <Select.Option value="prod">生产</Select.Option>
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
        <Button icon={<ReloadOutlined size={16} />} onClick={() => fetchInstances(3, page, pageSize)}>
          刷新
        </Button>
        {canCreate && (
          <Button type="primary" icon={<PlusOutlined size={16} />} onClick={handleCreate}>
            添加应用
          </Button>
        )}
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
        title={editingInstance ? '编辑应用服务' : '添加应用服务'}
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
            ci_type_id: 3,
            app_type: '',
            version: '',
            language: '',
            port: '',
            environment: 'dev',
            deploy_type: 'container',
            repository_url: '',
          }}
        >
          <Form.Item
            label="应用名称"
            name="name"
            rules={[{ required: true, message: '请输入应用名称' }]}
          >
            <Input placeholder="例如: user-service" prefix={<BoxIcon size={16} />} />
          </Form.Item>
          <Form.Item
            label="应用类型"
            name="app_type"
            rules={[{ required: true, message: '请选择应用类型' }]}
          >
            <Select placeholder="选择应用类型">
              <Select.Option value="web">Web应用</Select.Option>
              <Select.Option value="database">数据库</Select.Option>
              <Select.Option value="middleware">中间件</Select.Option>
              <Select.Option value="cache">缓存</Select.Option>
              <Select.Option value="message_queue">消息队列</Select.Option>
              <Select.Option value="bigdata">大数据</Select.Option>
              <Select.Option value="ai">AI服务</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="版本" name="version" rules={[{ required: true, message: '请输入版本' }]}>
            <Input placeholder="例如: v1.0.0" />
          </Form.Item>
          <Form.Item
            label="开发语言/框架"
            name="language"
            rules={[{ required: true, message: '请输入开发语言' }]}
          >
            <Select placeholder="选择开发语言" showSearch allowClear>
              <Select.Option value="Java">Java</Select.Option>
              <Select.Option value="Python">Python</Select.Option>
              <Select.Option value="Go">Go</Select.Option>
              <Select.Option value="Node.js">Node.js</Select.Option>
              <Select.Option value="PHP">PHP</Select.Option>
              <Select.Option value="Ruby">Ruby</Select.Option>
              <Select.Option value=".NET">.NET</Select.Option>
              <Select.Option value="C++">C++</Select.Option>
              <Select.Option value="Rust">Rust</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="端口" name="port" rules={[{ required: true, message: '请输入端口号' }]}>
            <Input type="number" placeholder="例如: 8081" min={1} max={65535} />
          </Form.Item>
          <Form.Item label="环境" name="environment" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="dev">开发</Select.Option>
              <Select.Option value="test">测试</Select.Option>
              <Select.Option value="staging">预发</Select.Option>
              <Select.Option value="prod">生产</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="部署方式" name="deploy_type" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="container">容器</Select.Option>
              <Select.Option value="vm">虚拟机</Select.Option>
              <Select.Option value="bare_metal">物理机</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="代码仓库" name="repository_url">
            <Input placeholder="例如: https://github.com/user/repo" />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="active">运行中</Select.Option>
              <Select.Option value="inactive">已停止</Select.Option>
              <Select.Option value="maintenance">维护中</Select.Option>
              <Select.Option value="decommissioned">已下线</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="角色" name="roles">
            <Select
              mode="multiple"
              placeholder="选择角色"
              options={ciRoles.map(role => ({
                label: role.display_name,
                value: role.id,
              }))}
            />
          </Form.Item>
          <Form.Item label="标签" name="tags">
            <Select
              mode="multiple"
              placeholder="选择标签"
              options={tags.map(tag => ({
                label: (
                  <span>
                    <span
                      style={{
                        display: 'inline-block',
                        width: 12,
                        height: 12,
                        backgroundColor: tag.color,
                        borderRadius: 2,
                        marginRight: 8,
                      }}
                    />
                    {tag.display_name}
                  </span>
                ),
                value: tag.id,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
