import { useEffect, useState } from 'react'
import { Card, Form, Input, Button, message, Divider, Space, Alert } from 'antd'
import { SaveOutlined, ReloadOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'

interface SystemConfig {
  id: number
  category: string
  key: string
  value: string
  description: string
  is_encrypted: boolean
  is_active: boolean
}

export default function VictoriaMetrics() {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [healthStatus, setHealthStatus] = useState<'success' | 'error' | null>(null)
  const token = useAuthStore((state) => state.token)

  useEffect(() => {
    loadConfigs()
  }, [])

  const loadConfigs = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/configs/category/monitoring', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        if (response.status === 403) {
          message.error('需要管理员权限才能访问配置管理')
        } else if (response.status === 401) {
          message.error('认证失败，请重新登录')
        } else {
          message.error('加载配置失败')
        }
        return
      }

      const result = await response.json()

      if (result.code === 0 && result.data) {
        const configs: SystemConfig[] = result.data
        const formData: Record<string, string> = {}

        configs.forEach((config) => {
          formData[config.key] = config.value
        })

        form.setFieldsValue(formData)
      } else {
        message.error(result.message || '加载配置失败')
      }
    } catch (error) {
      console.error('Failed to load configs:', error)
      message.error('加载配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const configs = [
        {
          category: 'monitoring',
          key: 'victoriametrics_endpoint',
          value: values.victoriametrics_endpoint || '',
          description: 'VictoriaMetrics API 端点',
          is_encrypted: false,
        },
        {
          category: 'monitoring',
          key: 'victoriametrics_username',
          value: values.victoriametrics_username || '',
          description: 'VictoriaMetrics 用户名',
          is_encrypted: false,
        },
        {
          category: 'monitoring',
          key: 'victoriametrics_password',
          value: values.victoriametrics_password || '',
          description: 'VictoriaMetrics 密码',
          is_encrypted: true,
        },
        {
          category: 'monitoring',
          key: 'victoriametrics_sync_interval',
          value: values.victoriametrics_sync_interval || '5m',
          description: '容器同步间隔',
          is_encrypted: false,
        },
      ]

      const response = await fetch('/api/v1/configs/batch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ configs }),
      })

      const result = await response.json()

      if (result.code === 0) {
        message.success('配置保存成功，重启服务后生效')
      } else {
        message.error(result.message || '保存配置失败')
      }
    } catch (error) {
      console.error('Failed to save configs:', error)
      message.error('保存配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleTestConnection = async () => {
    setTesting(true)
    setHealthStatus(null)

    try {
      const response = await fetch('/api/v1/monitoring/victoriametrics/health', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      const result = await response.json()

      if (result.code === 0) {
        setHealthStatus('success')
        message.success('连接测试成功')
      } else {
        setHealthStatus('error')
        message.error('连接测试失败: ' + result.message)
      }
    } catch (error) {
      setHealthStatus('error')
      message.error('连接测试失败')
    } finally {
      setTesting(false)
    }
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">
          VictoriaMetrics 配置
        </h1>
        <p className="text-gray-600 dark:text-text-secondary">
          配置 VictoriaMetrics 连接信息，用于容器监控和自动同步
        </p>
      </div>

      <Card>
        <Alert
          message="配置说明"
          description="修改配置后需要重启 CMDB 服务才能生效。配置将优先于环境变量使用。"
          type="info"
          showIcon
          className="mb-6"
        />

        <Form form={form} layout="vertical" onFinish={handleSave}>
          <Form.Item
            label="VictoriaMetrics 端点"
            name="victoriametrics_endpoint"
            rules={[
              { required: true, message: '请输入 VictoriaMetrics 端点' },
              { type: 'url', message: '请输入有效的 URL' },
            ]}
            extra="例如: https://10.120.43.230:8109"
          >
            <Input placeholder="https://victoriametrics.example.com:8109" />
          </Form.Item>

          <Form.Item
            label="用户名"
            name="victoriametrics_username"
            extra="如果启用了基本认证，请填写用户名"
          >
            <Input placeholder="admin" />
          </Form.Item>

          <Form.Item
            label="密码"
            name="victoriametrics_password"
            extra="密码将加密存储在数据库中"
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>

          <Form.Item
            label="同步间隔"
            name="victoriametrics_sync_interval"
            rules={[{ required: true, message: '请输入同步间隔' }]}
            extra="容器自动同步的时间间隔，例如: 5m, 10m, 1h"
          >
            <Input placeholder="5m" />
          </Form.Item>

          <Divider />

          <Space>
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={loading}>
              保存配置
            </Button>
            <Button icon={<ReloadOutlined />} onClick={loadConfigs} disabled={loading}>
              重新加载
            </Button>
            <Button
              icon={<CheckCircleOutlined />}
              onClick={handleTestConnection}
              loading={testing}
              disabled={loading}
            >
              测试连接
            </Button>
          </Space>

          {healthStatus && (
            <Alert
              message={healthStatus === 'success' ? '连接成功' : '连接失败'}
              type={healthStatus}
              showIcon
              className="mt-4"
            />
          )}
        </Form>
      </Card>
    </div>
  )
}
