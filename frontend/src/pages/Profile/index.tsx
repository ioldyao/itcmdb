import { useState, useEffect, useMemo } from 'react'
import { Form, Input, Button, message, Tabs, Tag, Empty, Spin, Tooltip } from 'antd'
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SaveOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { User, Shield, Key, Settings } from 'lucide-react'
import { useAuthStore } from '@/stores/authStore'

const passwordRules = [
  { label: '至少8个字符', test: (p: string) => p.length >= 8 },
  { label: '包含大写字母', test: (p: string) => /[A-Z]/.test(p) },
  { label: '包含小写字母', test: (p: string) => /[a-z]/.test(p) },
  { label: '包含数字', test: (p: string) => /[0-9]/.test(p) },
  { label: '包含特殊字符', test: (p: string) => /[^A-Za-z0-9]/.test(p) },
]

const getPasswordStrength = (password: string) => {
  if (!password) return { percent: 0, color: '#d9d9d9', text: '' }
  const passed = passwordRules.filter(r => r.test(password)).length
  if (passed <= 2) return { percent: 25, color: '#ff4d4f', text: '弱' }
  if (passed === 3) return { percent: 50, color: '#faad14', text: '中' }
  if (passed === 4) return { percent: 75, color: '#52c41a', text: '强' }
  return { percent: 100, color: '#13c2c2', text: '非常强' }
}

const resourceMap: Record<string, string> = {
  user: '用户', role: '角色', permission: '权限', ci: 'CI资产',
  ticket: '工单', alert: '告警', alert_rule: '告警规则',
  alert_receiver: '接收人', alert_receiver_group: '接收组',
  webhook: 'Webhook', routing: '路由规则', template: '通知模板',
  platform: '平台', audit_log: '审计日志',
}
const actionMap: Record<string, string> = {
  create: '创建', read: '读取', view: '查看', update: '更新',
  delete: '删除', manage: '管理', '*': '全部',
}

interface UserRole {
  id: number
  name: string
  description: string
}

export default function Profile() {
  const { user, token, logout, permissions } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [passwordLoading, setPasswordLoading] = useState(false)
  const [form] = Form.useForm()
  const [passwordForm] = Form.useForm()
  const [activeTab, setActiveTab] = useState('info')
  const [newPassword, setNewPassword] = useState('')
  const [showPasswordSuccess, setShowPasswordSuccess] = useState(false)
  const [roles, setRoles] = useState<UserRole[]>([])
  const [rolesLoading, setRolesLoading] = useState(false)

  useEffect(() => {
    if (!token || !user) return
    const fetchData = async () => {
      try {
        const userRes = await fetch('/api/v1/users/me', {
          headers: { Authorization: `Bearer ${token}` },
        })
        const userData = await userRes.json()
        if (userData.code === 0) {
          form.setFieldsValue({ fullName: userData.data.fullName })
        }

        setRolesLoading(true)
        const rolesRes = await fetch(`/api/v1/user-roles/user/${user.id}`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        const rolesData = await rolesRes.json()
        if (rolesData.code === 0) {
          setRoles(rolesData.data || [])
        }
      } catch (error) {
        console.error('Failed to fetch profile data:', error)
      } finally {
        setRolesLoading(false)
      }
    }
    fetchData()
  }, [token, user, form])

  const handleUpdateInfo = async (values: any) => {
    setLoading(true)
    try {
      const res = await fetch('/api/v1/users/me', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ full_name: values.fullName }),
      })
      const data = await res.json()
      if (data.code === 0) {
        message.success('姓名更新成功')
        useAuthStore.setState({ user: { ...user!, fullName: values.fullName } })
      } else {
        message.error(data.message || '更新失败')
      }
    } catch {
      message.error('更新失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async (values: any) => {
    setPasswordLoading(true)
    try {
      const res = await fetch('/api/v1/users/me', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ old_password: values.oldPassword, password: values.newPassword }),
      })
      const data = await res.json()
      if (data.code === 0) {
        setShowPasswordSuccess(true)
      } else {
        message.error(data.message || '修改失败')
      }
    } catch {
      message.error('修改失败')
    } finally {
      setPasswordLoading(false)
    }
  }

  const handleLogout = () => {
    logout()
    window.location.href = '/login'
  }

  const avatarInitial = useMemo(() => {
    return (user?.fullName || user?.username || '?').charAt(0).toUpperCase()
  }, [user])

  const strength = useMemo(() => getPasswordStrength(newPassword), [newPassword])

  const parsedPermissions = useMemo(() => {
    return permissions.map(perm => {
      const [resource, action] = perm.split(':')
      return {
        raw: perm,
        resource: resourceMap[resource] || resource,
        action: actionMap[action] || action,
        actionRaw: action,
      }
    })
  }, [permissions])

  if (!user) return null

  if (showPasswordSuccess) {
    return (
      <div className="max-w-md mx-auto mt-20">
        <div className="bg-white dark:bg-bg-secondary rounded-xl border border-gray-200 dark:border-white/8 p-10 text-center">
          <div className="w-16 h-16 mx-auto mb-5 rounded-full bg-green-50 dark:bg-green-500/10 flex items-center justify-center">
            <CheckCircleOutlined className="text-3xl text-green-500" />
          </div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-text-primary mb-2">密码修改成功</h2>
          <p className="text-gray-500 dark:text-text-secondary mb-6 text-sm">为了安全，请使用新密码重新登录</p>
          <Button type="primary" size="large" icon={<LogoutOutlined />} onClick={handleLogout} block>
            重新登录
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-5xl mx-auto px-4">
      {/* 用户信息头部 */}
      <div className="bg-white dark:bg-bg-secondary rounded-xl border border-gray-200 dark:border-white/8 p-6 mb-6">
        <div className="flex items-center gap-5">
          <div className="w-16 h-16 rounded-full bg-gray-100 dark:bg-white/10 flex items-center justify-center flex-shrink-0">
            <span className="text-2xl font-semibold text-gray-600 dark:text-text-secondary">{avatarInitial}</span>
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-xl font-semibold text-gray-900 dark:text-text-primary">
                {user.fullName || user.username}
              </h1>
              <Tag color="green" className="m-0">活跃</Tag>
            </div>
            <p className="text-sm text-gray-500 dark:text-text-secondary">@{user.username}</p>
          </div>
        </div>

        <div className="flex items-center gap-6 mt-5 pt-5 border-t border-gray-100 dark:border-white/8 text-sm">
          <div className="flex items-center gap-2 text-gray-600 dark:text-text-secondary">
            <MailOutlined className="text-gray-400" />
            <span>{user.email}</span>
          </div>
          <div className="flex items-center gap-2 text-gray-600 dark:text-text-secondary">
            <User size={14} className="text-gray-400" />
            <span>ID: {user.id}</span>
          </div>
          <div className="flex items-center gap-2 text-gray-600 dark:text-text-secondary">
            <Shield size={14} className="text-gray-400" />
            <span>{rolesLoading ? '加载中...' : roles.length > 0 ? roles.map(r => r.name).join(', ') : '未分配角色'}</span>
          </div>
          <div className="flex items-center gap-2 text-gray-600 dark:text-text-secondary">
            <Key size={14} className="text-gray-400" />
            <span>{permissions.length} 项权限</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左侧：角色 + 权限 */}
        <div className="lg:col-span-1 space-y-6">
          {/* 角色 */}
          <div className="bg-white dark:bg-bg-secondary rounded-xl border border-gray-200 dark:border-white/8 p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium text-gray-900 dark:text-text-primary">我的角色</h3>
              <span className="text-xs text-gray-400">{roles.length}</span>
            </div>
            {rolesLoading ? (
              <div className="py-6 text-center"><Spin size="small" /></div>
            ) : roles.length > 0 ? (
              <div className="space-y-2">
                {roles.map(role => (
                  <div key={role.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-white/5">
                    <div className="w-8 h-8 rounded-lg bg-gray-200 dark:bg-white/10 flex items-center justify-center">
                      <Shield size={14} className="text-gray-500 dark:text-text-secondary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 dark:text-text-primary">{role.name}</p>
                      <p className="text-xs text-gray-500 dark:text-text-secondary truncate">
                        {role.description || '无描述'}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <Empty description="暂无角色" image={Empty.PRESENTED_IMAGE_SIMPLE} className="py-4" />
            )}
          </div>

          {/* 权限 */}
          <div className="bg-white dark:bg-bg-secondary rounded-xl border border-gray-200 dark:border-white/8 p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium text-gray-900 dark:text-text-primary">我的权限</h3>
              <span className="text-xs text-gray-400">{permissions.length}</span>
            </div>
            {permissions.length > 0 ? (
              <div className="flex flex-wrap gap-1.5">
                {parsedPermissions.map((perm) => (
                  <Tooltip key={perm.raw} title={perm.raw}>
                    <Tag
                      className="m-0 cursor-default"
                      color={
                        perm.actionRaw === 'delete' ? 'red' :
                        perm.actionRaw === 'create' ? 'green' :
                        perm.actionRaw === 'update' ? 'blue' :
                        perm.actionRaw === '*' ? 'purple' : 'default'
                      }
                    >
                      {perm.resource}:{perm.action}
                    </Tag>
                  </Tooltip>
                ))}
              </div>
            ) : (
              <Empty description="暂无权限" image={Empty.PRESENTED_IMAGE_SIMPLE} className="py-4" />
            )}
          </div>
        </div>

        {/* 右侧：设置 */}
        <div className="lg:col-span-2">
          <div className="bg-white dark:bg-bg-secondary rounded-xl border border-gray-200 dark:border-white/8">
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              className="px-6 pt-2"
              items={[
                {
                  key: 'info',
                  label: (
                    <span className="flex items-center gap-2">
                      <Settings size={14} />
                      账户设置
                    </span>
                  ),
                  children: (
                    <div className="pb-6 max-w-md">
                      <h3 className="text-sm font-medium text-gray-900 dark:text-text-primary mb-4">修改姓名</h3>
                      <Form form={form} layout="vertical" onFinish={handleUpdateInfo}>
                        <Form.Item label={<span className="dark:text-text-secondary">用户名</span>}>
                          <Input value={user.username} disabled prefix={<UserOutlined />} />
                        </Form.Item>
                        <Form.Item label={<span className="dark:text-text-secondary">邮箱</span>}>
                          <Input value={user.email} disabled prefix={<MailOutlined />} />
                        </Form.Item>
                        <Form.Item
                          label={<span className="dark:text-text-secondary">姓名</span>}
                          name="fullName"
                          rules={[{ required: true, message: '请输入姓名' }]}
                        >
                          <Input prefix={<UserOutlined />} placeholder="请输入姓名" />
                        </Form.Item>
                        <Form.Item className="mb-0">
                          <Button type="primary" htmlType="submit" loading={loading} icon={<SaveOutlined />}>
                            保存修改
                          </Button>
                        </Form.Item>
                      </Form>
                    </div>
                  ),
                },
                {
                  key: 'password',
                  label: (
                    <span className="flex items-center gap-2">
                      <LockOutlined />
                      修改密码
                    </span>
                  ),
                  children: (
                    <div className="pb-6 max-w-md">
                      <h3 className="text-sm font-medium text-gray-900 dark:text-text-primary mb-4">安全设置</h3>
                      <Form form={passwordForm} layout="vertical" onFinish={handleChangePassword}>
                        <Form.Item
                          label={<span className="dark:text-text-secondary">当前密码</span>}
                          name="oldPassword"
                          rules={[{ required: true, message: '请输入当前密码' }]}
                        >
                          <Input.Password prefix={<LockOutlined />} placeholder="请输入当前密码" />
                        </Form.Item>

                        <Form.Item
                          label={<span className="dark:text-text-secondary">新密码</span>}
                          name="newPassword"
                          rules={[
                            { required: true, message: '请输入新密码' },
                            { min: 8, message: '密码至少8个字符' },
                            {
                              validator: (_, value) => {
                                if (!value) return Promise.resolve()
                                if (!/[A-Z]/.test(value)) return Promise.reject('需包含大写字母')
                                if (!/[a-z]/.test(value)) return Promise.reject('需包含小写字母')
                                if (!/[0-9]/.test(value)) return Promise.reject('需包含数字')
                                if (!/[^A-Za-z0-9]/.test(value)) return Promise.reject('需包含特殊字符')
                                return Promise.resolve()
                              },
                            },
                          ]}
                        >
                          <Input.Password
                            prefix={<LockOutlined />}
                            placeholder="请输入新密码"
                            onChange={(e) => setNewPassword(e.target.value)}
                          />
                        </Form.Item>

                        {newPassword && (
                          <div className="mb-4 p-3 bg-gray-50 dark:bg-white/5 rounded-lg">
                            <div className="flex items-center justify-between mb-2">
                              <span className="text-xs text-gray-500 dark:text-text-secondary">密码强度</span>
                              <span className="text-xs font-medium" style={{ color: strength.color }}>{strength.text}</span>
                            </div>
                            <div className="h-1.5 bg-gray-200 dark:bg-white/10 rounded-full overflow-hidden mb-3">
                              <div
                                className="h-full rounded-full transition-all duration-300"
                                style={{ width: `${strength.percent}%`, backgroundColor: strength.color }}
                              />
                            </div>
                            <div className="space-y-1">
                              {passwordRules.map((rule) => {
                                const passed = rule.test(newPassword)
                                return (
                                  <div key={rule.label} className="flex items-center gap-2 text-xs">
                                    {passed ? (
                                      <CheckCircleOutlined className="text-green-500" />
                                    ) : (
                                      <CloseCircleOutlined className="text-gray-300 dark:text-text-tertiary" />
                                    )}
                                    <span className={passed ? 'text-green-600 dark:text-green-400' : 'text-gray-500 dark:text-text-secondary'}>
                                      {rule.label}
                                    </span>
                                  </div>
                                )
                              })}
                            </div>
                          </div>
                        )}

                        <Form.Item
                          label={<span className="dark:text-text-secondary">确认密码</span>}
                          name="confirmPassword"
                          dependencies={['newPassword']}
                          rules={[
                            { required: true, message: '请确认新密码' },
                            ({ getFieldValue }) => ({
                              validator(_, value) {
                                if (!value || getFieldValue('newPassword') === value) return Promise.resolve()
                                return Promise.reject(new Error('两次输入的密码不一致'))
                              },
                            }),
                          ]}
                        >
                          <Input.Password prefix={<LockOutlined />} placeholder="请再次输入新密码" />
                        </Form.Item>

                        <Form.Item className="mb-0">
                          <Button type="primary" htmlType="submit" loading={passwordLoading} danger icon={<LockOutlined />}>
                            修改密码
                          </Button>
                        </Form.Item>
                      </Form>
                    </div>
                  ),
                },
              ]}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
