import { useEffect, useState } from 'react'
import {
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Tabs,
  Transfer,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { Shield } from 'lucide-react'
import type { ColumnsType } from 'antd/es/table'
import { useAdminRoleStore } from '@/stores/adminRoleStore'

// 角色接口
interface Role {
  id: number
  name: string
  description: string
  created_at: string
  updated_at: string
}

// 权限接口
interface Permission {
  id: number
  resource: string
  action: string
}

export default function AdminRoles() {
  const {
    roles,
    permissions,
    loading,
    fetchRoles,
    fetchPermissions,
    createRole,
    updateRole,
    deleteRole,
    createPermission,
    deletePermission,
    getRolePermissions,
    assignPermissionToRole,
    removePermissionFromRole,
  } = useAdminRoleStore()

  const [activeTab, setActiveTab] = useState('roles')
  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false)
  const [isPermissionModalOpen, setIsPermissionModalOpen] = useState(false)
  const [isPermissionAssignModalOpen, setIsPermissionAssignModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [selectedRoleForPermissions, setSelectedRoleForPermissions] = useState<Role | null>(null)
  const [selectedPermissionIds, setSelectedPermissionIds] = useState<number[]>([])
  const [roleForm] = Form.useForm()
  const [permissionForm] = Form.useForm()

  useEffect(() => {
    fetchRoles()
    fetchPermissions()
  }, [fetchRoles, fetchPermissions])

  // 角色表格列
  const roleColumns: ColumnsType<Role> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    {
      title: '角色名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <code>{text}</code>,
    },
    { title: '描述', dataIndex: 'description', key: 'description', render: (text) => text || '-' },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 250,
      render: (_: any, record: Role) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditRole(record)}>
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => handleAssignPermissions(record)}
          >
            权限
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteRole(record.id)}
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

  // 权限表格列
  const permissionColumns: ColumnsType<Permission> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    { title: '资源', dataIndex: 'resource', key: 'resource' },
    { title: '操作', dataIndex: 'action', key: 'action' },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: Permission) => (
        <Popconfirm
          title="确认删除"
          description="删除后无法恢复"
          onConfirm={() => handleDeletePermission(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ]

  const handleCreateRole = () => {
    setEditingRole(null)
    roleForm.resetFields()
    setIsRoleModalOpen(true)
  }

  const handleEditRole = (record: Role) => {
    setEditingRole(record)
    roleForm.setFieldsValue({
      name: record.name,
      description: record.description,
    })
    setIsRoleModalOpen(true)
  }

  const handleDeleteRole = async (id: number) => {
    try {
      await deleteRole(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleRoleSubmit = async () => {
    try {
      const values = await roleForm.validateFields()
      if (editingRole) {
        await updateRole(editingRole.id, values.name, values.description)
        message.success('更新成功')
      } else {
        await createRole(values.name, values.description)
        message.success('创建成功')
      }
      setIsRoleModalOpen(false)
      roleForm.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleCreatePermission = () => {
    permissionForm.resetFields()
    setIsPermissionModalOpen(true)
  }

  const handleDeletePermission = async (id: number) => {
    try {
      await deletePermission(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handlePermissionSubmit = async () => {
    try {
      const values = await permissionForm.validateFields()
      await createPermission(values.resource, values.action)
      message.success('创建成功')
      setIsPermissionModalOpen(false)
      permissionForm.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleAssignPermissions = async (role: Role) => {
    setSelectedRoleForPermissions(role)
    try {
      const rolePermissions = await getRolePermissions(role.id)
      setSelectedPermissionIds(rolePermissions.map((p: Permission) => p.id))
      setIsPermissionAssignModalOpen(true)
    } catch (error) {
      message.error('获取角色权限失败')
    }
  }

  const handlePermissionAssignSubmit = async () => {
    if (!selectedRoleForPermissions) return

    try {
      // 获取当前角色权限
      const currentPermissions = await getRolePermissions(selectedRoleForPermissions.id)
      const currentIds = currentPermissions.map((p: Permission) => p.id)

      // 添加新权限
      for (const permissionId of selectedPermissionIds) {
        if (!currentIds.includes(permissionId)) {
          await assignPermissionToRole(selectedRoleForPermissions.id, permissionId)
        }
      }

      // 移除旧权限
      for (const permissionId of currentIds) {
        if (!selectedPermissionIds.includes(permissionId)) {
          await removePermissionFromRole(selectedRoleForPermissions.id, permissionId)
        }
      }

      message.success('权限分配成功')
      setIsPermissionAssignModalOpen(false)
    } catch (error) {
      message.error('权限分配失败')
    }
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">角色权限管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理系统用户角色和权限</p>
      </div>

      {/* Tabs */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'roles',
            label: '角色管理',
            children: (
              <div className="bg-white dark:bg-bg-secondary p-6 rounded-lg border border-gray-200 dark:border-white/8">
                <div className="mb-6 flex justify-between items-center">
                  <div className="flex items-center gap-2">
                    <Shield size={20} />
                    <span className="text-lg font-medium">系统角色</span>
                  </div>
                  <Space>
                    <Button icon={<ReloadOutlined />} onClick={() => fetchRoles()}>
                      刷新
                    </Button>
                    <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateRole}>
                      添加角色
                    </Button>
                  </Space>
                </div>

                <Table
                  columns={roleColumns}
                  dataSource={roles}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    pageSize: 20,
                    showSizeChanger: true,
                    showTotal: (total) => `共 ${total} 条`,
                  }}
                />
              </div>
            ),
          },
          {
            key: 'permissions',
            label: '权限管理',
            children: (
              <div className="bg-white dark:bg-bg-secondary p-6 rounded-lg border border-gray-200 dark:border-white/8">
                <div className="mb-6 flex justify-between items-center">
                  <div className="flex items-center gap-2">
                    <Shield size={20} />
                    <span className="text-lg font-medium">权限列表</span>
                  </div>
                  <Space>
                    <Button icon={<ReloadOutlined />} onClick={() => fetchPermissions()}>
                      刷新
                    </Button>
                    <Button type="primary" icon={<PlusOutlined />} onClick={handleCreatePermission}>
                      添加权限
                    </Button>
                  </Space>
                </div>

                <Table
                  columns={permissionColumns}
                  dataSource={permissions}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    pageSize: 20,
                    showSizeChanger: true,
                    showTotal: (total) => `共 ${total} 条`,
                  }}
                />
              </div>
            ),
          },
        ]}
      />

      {/* 角色编辑弹窗 */}
      <Modal
        title={editingRole ? '编辑角色' : '添加角色'}
        open={isRoleModalOpen}
        onCancel={() => {
          setIsRoleModalOpen(false)
          roleForm.resetFields()
        }}
        onOk={handleRoleSubmit}
        width={600}
      >
        <Form form={roleForm} layout="vertical">
          <Form.Item
            label="角色名称"
            name="name"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="例如: admin" />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea placeholder="角色描述" rows={4} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限创建弹窗 */}
      <Modal
        title="添加权限"
        open={isPermissionModalOpen}
        onCancel={() => {
          setIsPermissionModalOpen(false)
          permissionForm.resetFields()
        }}
        onOk={handlePermissionSubmit}
        width={600}
      >
        <Form form={permissionForm} layout="vertical">
          <Form.Item
            label="资源"
            name="resource"
            rules={[{ required: true, message: '请输入资源' }]}
          >
            <Input placeholder="例如: users, servers" />
          </Form.Item>
          <Form.Item
            label="操作"
            name="action"
            rules={[{ required: true, message: '请输入操作' }]}
          >
            <Select placeholder="选择操作">
              <Select.Option value="create">创建</Select.Option>
              <Select.Option value="read">读取</Select.Option>
              <Select.Option value="update">更新</Select.Option>
              <Select.Option value="delete">删除</Select.Option>
              <Select.Option value="manage">管理</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限分配弹窗 */}
      <Modal
        title={`分配权限 - ${selectedRoleForPermissions?.name || ''}`}
        open={isPermissionAssignModalOpen}
        onCancel={() => {
          setIsPermissionAssignModalOpen(false)
          setSelectedPermissionIds([])
        }}
        onOk={handlePermissionAssignSubmit}
        width={800}
      >
        <Transfer
          dataSource={permissions}
          titles={['可用权限', '已分配权限']}
          targetKeys={selectedPermissionIds}
          onChange={(targetKeys) => setSelectedPermissionIds(targetKeys as number[])}
          render={(item) => `${item.resource}:${item.action}`}
          listStyle={{
            width: 300,
            height: 400,
          }}
          showSearch
          filterOption={(inputValue, item) =>
            item.resource.toLowerCase().includes(inputValue.toLowerCase()) ||
            item.action.toLowerCase().includes(inputValue.toLowerCase())
          }
        />
      </Modal>
    </div>
  )
}
