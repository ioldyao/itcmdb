import { useEffect, useState } from 'react'
import { Table, Button, Tag, Space, Modal, Form, Input, InputNumber, Select, message, Popconfirm, Tabs } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { Shield, Users } from 'lucide-react'
import type { ColumnsType } from 'antd/es/table'
import { useRoleStore, CIRole, OwnerRole } from '@/stores/roleStore'

export default function CIRolesPage() {
  const {
    ciRoles,
    ownerRoles,
    loading,
    fetchCIRoles,
    fetchOwnerRoles,
    createCIRole,
    updateCIRole,
    deleteCIRole,
    createOwnerRole,
    updateOwnerRole,
    deleteOwnerRole,
  } = useRoleStore()

  const [activeTab, setActiveTab] = useState<'ci' | 'owner'>('ci')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<CIRole | OwnerRole | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchCIRoles()
    fetchOwnerRoles()
  }, [fetchCIRoles, fetchOwnerRoles])

  // CI角色表格列
  const ciRoleColumns: ColumnsType<CIRole> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    {
      title: '图标',
      dataIndex: 'icon',
      key: 'icon',
      width: 60,
      render: () => <Shield size={20} />,
    },
    {
      title: '角色名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <code>{text}</code>,
    },
    { title: '显示名称', dataIndex: 'display_name', key: 'display_name' },
    { title: '描述', dataIndex: 'description', key: 'description', render: (text) => text || '-' },
    {
      title: '颜色',
      dataIndex: 'color',
      key: 'color',
      width: 100,
      render: (color: string) => (
        <span
          style={{
            display: 'inline-block',
            width: 20,
            height: 20,
            backgroundColor: color || '#ccc',
            borderRadius: 4,
          }}
        />
      ),
    },
    { title: '优先级', dataIndex: 'priority', width: 80, key: 'priority' },
    {
      title: '状态',
      dataIndex: 'is_active',
      width: 100,
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'default'}>{isActive ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: CIRole) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteCIRole(record.id)}
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

  // 负责人角色表格列
  const ownerRoleColumns: ColumnsType<OwnerRole> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    {
      title: '图标',
      key: 'icon',
      width: 60,
      render: () => <Users size={20} />,
    },
    {
      title: '角色名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <code>{text}</code>,
    },
    { title: '显示名称', dataIndex: 'display_name', key: 'display_name' },
    { title: '描述', dataIndex: 'description', key: 'description', render: (text) => text || '-' },
    { title: '级别', dataIndex: 'level', width: 80, key: 'level' },
    {
      title: '状态',
      dataIndex: 'is_active',
      width: 100,
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'default'}>{isActive ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: OwnerRole) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteOwnerRole(record.id)}
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

  const handleCreate = () => {
    setEditingRole(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = (record: CIRole | OwnerRole) => {
    setEditingRole(record)
    if (activeTab === 'ci') {
      const ciRole = record as CIRole
      form.setFieldsValue({
        name: ciRole.name,
        display_name: ciRole.display_name,
        description: ciRole.description,
        color: ciRole.color,
        icon: ciRole.icon,
        priority: ciRole.priority,
      })
    } else {
      const ownerRole = record as OwnerRole
      form.setFieldsValue({
        name: ownerRole.name,
        display_name: ownerRole.display_name,
        description: ownerRole.description,
        level: ownerRole.level,
      })
    }
    setIsModalOpen(true)
  }

  const handleDeleteCIRole = async (id: number) => {
    try {
      await deleteCIRole(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleDeleteOwnerRole = async (id: number) => {
    try {
      await deleteOwnerRole(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (activeTab === 'ci') {
        if (editingRole) {
          await updateCIRole((editingRole as CIRole).id, values)
          message.success('更新成功')
        } else {
          await createCIRole(values)
          message.success('创建成功')
        }
      } else {
        if (editingRole) {
          await updateOwnerRole((editingRole as OwnerRole).id, values)
          message.success('更新成功')
        } else {
          await createOwnerRole(values)
          message.success('创建成功')
        }
      }

      setIsModalOpen(false)
      form.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">角色管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理CI技术角色和负责人角色</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Button icon={<ReloadOutlined />} onClick={() => { fetchCIRoles(); fetchOwnerRoles() }}>
          刷新
        </Button>
        <div className="flex-1" />
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          添加角色
        </Button>
      </div>

      {/* 角色表格 */}
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as 'ci' | 'owner')}
        items={[
          {
            key: 'ci',
            label: `CI技术角色 (${ciRoles.length})`,
            children: (
              <Table
                columns={ciRoleColumns}
                dataSource={ciRoles}
                rowKey="id"
                loading={loading}
                pagination={false}
                className="bg-white dark:bg-bg-secondary"
              />
            ),
          },
          {
            key: 'owner',
            label: `负责人角色 (${ownerRoles.length})`,
            children: (
              <Table
                columns={ownerRoleColumns}
                dataSource={ownerRoles}
                rowKey="id"
                loading={loading}
                pagination={false}
                className="bg-white dark:bg-bg-secondary"
              />
            ),
          },
        ]}
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingRole ? '编辑角色' : '添加角色'}
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
            priority: 0,
            level: 0,
          }}
        >
          <Form.Item
            label="角色名称"
            name="name"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="例如: primary_db" />
          </Form.Item>

          <Form.Item
            label="显示名称"
            name="display_name"
            rules={[{ required: true, message: '请输入显示名称' }]}
          >
            <Input placeholder="例如: 主数据库" />
          </Form.Item>

          <Form.Item label="描述" name="description">
            <Input.TextArea placeholder="角色描述" rows={3} />
          </Form.Item>

          {activeTab === 'ci' && (
            <>
              <Form.Item label="颜色" name="color" rules={[{ required: true, message: '请选择颜色' }]}>
                <Input type="color" className="w-32" />
              </Form.Item>

              <Form.Item label="图标" name="icon">
                <Select placeholder="选择图标">
                  <Select.Option value="database">数据库</Select.Option>
                  <Select.Option value="server">服务器</Select.Option>
                  <Select.Option value="workflow">流程</Select.Option>
                  <Select.Option value="cpu">CPU</Select.Option>
                  <Select.Option value="hard-drive">存储</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item label="优先级" name="priority">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </>
          )}

          {activeTab === 'owner' && (
            <Form.Item label="级别" name="level">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
