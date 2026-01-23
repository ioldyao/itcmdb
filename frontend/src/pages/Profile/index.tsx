import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, message, Tabs, Descriptions } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'

export default function Profile() {
  const { user, token } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [passwordLoading, setPasswordLoading] = useState(false)
  const [form] = Form.useForm()

  // 获取用户详细信息
  useEffect(() => {
    if (!token || !user) return

    const fetchUserInfo = async () => {
      try {
        const response = await fetch('/api/v1/users/me', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })
        const data = await response.json()
        if (data.code === 0) {
          form.setFieldsValue({
            username: data.data.username,
            email: data.data.email,
            fullName: data.data.fullName,
          })
        }
      } catch (error) {
        message.error('获取用户信息失败')
      }
    }

    fetchUserInfo()
  }, [token, user, form])

  // 更新用户信息
  const handleUpdateInfo = async (values: any) => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/users/me', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          full_name: values.fullName,
        }),
      })

      const data = await response.json()
      if (data.code === 0) {
        message.success('个人信息更新成功')
        // 更新本地 store
        useAuthStore.setState({
          user: {
            ...user!,
            fullName: values.fullName,
          },
        })
      } else {
        message.error(data.message || '更新失败')
      }
    } catch (error) {
      message.error('更新失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  // 修改密码
  const handleChangePassword = async (values: any) => {
    setPasswordLoading(true)
    try {
      const response = await fetch('/api/v1/users/me', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          password: values.newPassword,
        }),
      })

      const data = await response.json()
      if (data.code === 0) {
        message.success('密码修改成功，请重新登录')
        // 登出
        useAuthStore.getState().logout()
        window.location.href = '/login'
      } else {
        message.error(data.message || '修改失败')
      }
    } catch (error) {
      message.error('修改失败，请重试')
    } finally {
      setPasswordLoading(false)
    }
  }

  if (!user) {
    return null
  }

  const tabItems = [
    {
      key: 'info',
      label: '基本信息',
      children: (
        <>
          <Descriptions bordered column={1}>
            <Descriptions.Item label="用户名">{user.username}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{user.email}</Descriptions.Item>
            <Descriptions.Item label="姓名">{user.fullName}</Descriptions.Item>
            <Descriptions.Item label="用户ID">{user.id}</Descriptions.Item>
          </Descriptions>

          <div style={{ marginTop: 24 }}>
            <h3>修改个人信息</h3>
            <Form
              form={form}
              layout="vertical"
              onFinish={handleUpdateInfo}
              initialValues={{
                username: user.username,
                email: user.email,
                fullName: user.fullName,
              }}
            >
              <Form.Item label="用户名" name="username">
                <Input disabled prefix={<UserOutlined />} />
              </Form.Item>

              <Form.Item label="邮箱" name="email">
                <Input disabled prefix={<UserOutlined />} />
              </Form.Item>

              <Form.Item
                label="姓名"
                name="fullName"
                rules={[{ required: true, message: '请输入姓名' }]}
              >
                <Input prefix={<UserOutlined />} placeholder="请输入姓名" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  更新信息
                </Button>
              </Form.Item>
            </Form>
          </div>
        </>
      ),
    },
    {
      key: 'password',
      label: '修改密码',
      children: (
        <Form
          layout="vertical"
          onFinish={handleChangePassword}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            label="新密码"
            name="newPassword"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码" />
          </Form.Item>

          <Form.Item
            label="确认密码"
            name="confirmPassword"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请再次输入新密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={passwordLoading} danger>
              修改密码
            </Button>
          </Form.Item>
        </Form>
      ),
    },
  ]

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Card title="个人中心" bordered={false}>
        <Tabs defaultActiveKey="info" items={tabItems} />
      </Card>
    </div>
  )
}
