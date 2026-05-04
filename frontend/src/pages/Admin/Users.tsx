import { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, Select, Space, message, Popconfirm, Tag } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import type { ColumnsType } from 'antd/es/table'

interface User {
  id: number
  username: string
  email: string
  full_name: string
  status: string
}

interface Role {
  id: number
  name: string
  description: string
}

export default function AdminUsers() {
  const { token } = useAuthStore()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [form] = Form.useForm()

  // 角色管理相关状态
  const [roleModalVisible, setRoleModalVisible] = useState(false)
  const [managingUser, setManagingUser] = useState<User | null>(null)
  const [allRoles, setAllRoles] = useState<Role[]>([])
  const [userRoleIds, setUserRoleIds] = useState<number[]>([])
  const [roleLoading, setRoleLoading] = useState(false)

  // 获取用户列表
  const fetchUsers = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/users', {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()

      if (data.code === 0) {
        setUsers(data.data || [])
      } else {
        message.error(data.message || '获取用户列表失败')
      }
    } catch (error) {
      message.error('获取用户列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUsers()
    fetchAllRoles()
  }, [])

  // 获取所有角色
  const fetchAllRoles = async () => {
    try {
      const response = await fetch('/api/v1/roles', {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()
      if (data.code === 0) {
        setAllRoles(data.data || [])
      }
    } catch (error) {
      console.error('获取角色列表失败:', error)
    }
  }

  // 获取用户的角色
  const fetchUserRoles = async (userId: number) => {
    setRoleLoading(true)
    try {
      const response = await fetch(`/api/v1/user-roles/user/${userId}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()
      if (data.code === 0) {
        const roles = data.data || []
        setUserRoleIds(roles.map((r: Role) => r.id))
      }
    } catch (error) {
      message.error('获取用户角色失败')
    } finally {
      setRoleLoading(false)
    }
  }

  // 打开角色管理模态框
  const handleManageRoles = (user: User) => {
    setManagingUser(user)
    setRoleModalVisible(true)
    fetchUserRoles(user.id)
  }

  // 保存用户角色
  const handleSaveRoles = async () => {
    if (!managingUser) return

    // 角色互斥检查
    const selectedRoleNames = allRoles
      .filter(role => userRoleIds.includes(role.id))
      .map(role => role.name.toLowerCase())

    // 检查是否同时选择了互斥的角色
    const hasAdmin = selectedRoleNames.includes('admin')
    const hasUser = selectedRoleNames.includes('user')
    const hasOperator = selectedRoleNames.includes('operator')

    if (hasAdmin && (hasUser || hasOperator)) {
      message.error('管理员角色不能与普通用户或运维人员角色同时分配')
      return
    }

    if (hasUser && hasOperator) {
      message.warning('建议不要同时分配普通用户和运维人员角色')
    }

    setRoleLoading(true)
    try {
      // 获取当前用户角色
      const response = await fetch(`/api/v1/user-roles/user/${managingUser.id}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()
      const currentRoles = data.code === 0 ? (data.data || []) : []
      const currentRoleIds = currentRoles.map((r: Role) => r.id)

      // 计算需要添加和删除的角色
      const toAdd = userRoleIds.filter(id => !currentRoleIds.includes(id))
      const toRemove = currentRoleIds.filter((id: number) => !userRoleIds.includes(id))

      // 添加新角色
      for (const roleId of toAdd) {
        await fetch('/api/v1/user-roles', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            user_id: managingUser.id,
            role_id: roleId,
          }),
        })
      }

      // 删除角色
      for (const roleId of toRemove) {
        await fetch('/api/v1/user-roles', {
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            user_id: managingUser.id,
            role_id: roleId,
          }),
        })
      }

      message.success('角色分配成功')
      setRoleModalVisible(false)
    } catch (error) {
      message.error('角色分配失败')
    } finally {
      setRoleLoading(false)
    }
  }

  // 打开新增用户模态框
  const handleAdd = () => {
    setEditingUser(null)
    form.resetFields()
    setModalVisible(true)
  }

  // 打开编辑用户模态框
  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue({
      username: user.username,
      email: user.email,
      full_name: user.full_name,
      status: user.status,
    })
    setModalVisible(true)
  }

  // 删除用户
  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/v1/users/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await response.json()

      if (data.code === 0) {
        message.success('删除成功')
        fetchUsers()
      } else {
        message.error(data.message || '删除失败')
      }
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      if (editingUser) {
        // 更新用户
        const response = await fetch(`/api/v1/users/${editingUser.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            full_name: values.full_name,
            status: values.status,
          }),
        })

        const data = await response.json()
        if (data.code === 0) {
          message.success('更新成功')
        } else {
          message.error(data.message || '更新失败')
          return
        }
      } else {
        // 创建用户
        const response = await fetch('/api/v1/auth/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            username: values.username,
            email: values.email,
            password: values.password,
            full_name: values.full_name,
          }),
        })

        const data = await response.json()
        if (data.code === 0) {
          message.success('创建成功')
        } else {
          message.error(data.message || '创建失败')
          return
        }
      }

      setModalVisible(false)
      form.resetFields()
      fetchUsers()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<User> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '姓名',
      dataIndex: 'full_name',
      key: 'full_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<UserOutlined />}
            onClick={() => handleManageRoles(record)}
          >
            管理角色
          </Button>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个用户吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">用户管理</h1>
          <p className="text-gray-600 dark:text-text-secondary">管理系统用户账户和状态</p>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加用户
        </Button>
      </div>

      <div className="bg-white dark:bg-bg-secondary rounded-lg border border-gray-200 dark:border-white/8">
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条用户`,
          }}
          locale={{
            emptyText: '暂无用户数据',
          }}
        />
      </div>

      <Modal
        title={editingUser ? '编辑用户' : '添加用户'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
        }}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="用户名"
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { pattern: /^[a-zA-Z0-9_]{3,20}$/, message: '用户名只能包含字母、数字、下划线，3-20个字符' },
            ]}
          >
            <Input placeholder="请输入用户名" disabled={!!editingUser} />
          </Form.Item>

          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="请输入邮箱" disabled={!!editingUser} />
          </Form.Item>

          <Form.Item
            label="姓名"
            name="full_name"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              label="密码"
              name="password"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' },
              ]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}

          <Form.Item
            label="状态"
            name="status"
            initialValue="active"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Select.Option value="active">启用</Select.Option>
              <Select.Option value="inactive">禁用</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingUser ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 角色管理模态框 */}
      <Modal
        title={`管理用户角色 - ${managingUser?.username || ''}`}
        open={roleModalVisible}
        onCancel={() => {
          setRoleModalVisible(false)
          setUserRoleIds([])
        }}
        onOk={handleSaveRoles}
        confirmLoading={roleLoading}
        width={600}
      >
        <div style={{ marginBottom: 16 }}>
          <p>为用户 <strong>{managingUser?.full_name}</strong> 分配角色：</p>
        </div>
        <Select
          mode="multiple"
          style={{ width: '100%' }}
          placeholder="选择角色"
          value={userRoleIds}
          onChange={setUserRoleIds}
          loading={roleLoading}
          options={allRoles.map(role => ({
            label: `${role.name} - ${role.description || ''}`,
            value: role.id,
          }))}
        />
        <div style={{ marginTop: 16, padding: 12, background: '#f0f0f0', borderRadius: 4 }}>
          <div style={{ color: '#666', fontSize: 12, marginBottom: 8 }}>
            已选择 {userRoleIds.length} 个角色
          </div>
          <div style={{ color: '#ff4d4f', fontSize: 12 }}>
            ⚠️ 注意：管理员角色不能与其他角色同时分配
          </div>
        </div>
      </Modal>
    </div>
  )
}
