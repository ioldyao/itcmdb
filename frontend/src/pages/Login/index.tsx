import { useState } from 'react'
import { Form, Input, Button, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { Server } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { authService } from '@/services/auth'
import './Login.css'

export default function Login() {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true)
    try {
      const response = await authService.login(values) as any
      if (response.code === 0) {
        setAuth(response.data.user, response.data.token, response.data.permissions)
        message.success('登录成功')
        navigate('/dashboard')
      } else {
        message.error(response.message || '登录失败')
      }
    } catch (error: any) {
      const msg = error?.response?.data?.message
      if (msg === 'Locked') {
        message.error({ content: '账户已被锁定，请15分钟后重试', duration: 5 })
      } else if (msg === 'Unauthorized') {
        message.error({ content: '用户名或密码错误', duration: 5 })
      } else {
        message.error({ content: msg || '登录失败，请稍后重试', duration: 5 })
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <div className="login-wrapper">
        {/* 左侧品牌区域 */}
        <div className="login-brand">
          <div className="brand-content">
            <div className="brand-logo">
              <Server size={40} />
            </div>
            <h1 className="brand-title">ITCMDB</h1>
            <p className="brand-subtitle">智能运维管理平台</p>
            <div className="brand-features">
              <div className="feature-item">
                <div className="feature-dot" />
                <span>CMDB 资产管理</span>
              </div>
              <div className="feature-item">
                <div className="feature-dot" />
                <span>智能告警监控</span>
              </div>
              <div className="feature-item">
                <div className="feature-dot" />
                <span>工单流程管理</span>
              </div>
              <div className="feature-item">
                <div className="feature-dot" />
                <span>RBAC 权限控制</span>
              </div>
            </div>
          </div>
        </div>

        {/* 右侧登录表单 */}
        <div className="login-form-container">
          <div className="login-form-wrapper">
            <div className="form-header">
              <h2>欢迎登录</h2>
              <p>请输入您的账户信息</p>
            </div>

            <Form onFinish={onFinish} size="large" className="login-form">
              <Form.Item
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="用户名"
                  className="login-input"
                />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                  className="login-input"
                />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  block
                  loading={loading}
                  className="login-button"
                >
                  登录
                </Button>
              </Form.Item>
            </Form>

            <div className="form-footer">
              <span className="version-text">ITCMDB v1.0.0</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
