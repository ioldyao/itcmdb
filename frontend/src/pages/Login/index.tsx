import { Form, Input, Button, Card, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { authService } from '@/services/auth'
import './Login.css'

export default function Login() {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      const response = await authService.login(values)
      setAuth(response.user, response.token, response.permissions)
      message.success('登录成功')
      navigate('/dashboard')
    } catch (error) {
      message.error('登录失败，请检查用户名和密码')
    }
  }

  return (
    <div className="login-container">
      <Card className="login-card" title="ITCMDB 运维管理平台">
        <Form onFinish={onFinish} size="large">
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
