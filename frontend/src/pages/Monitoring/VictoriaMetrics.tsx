import { useEffect, useState } from 'react'
import {
  Card,
  Form,
  Input,
  Button,
  message,
  Divider,
  Space,
  Alert,
  Table,
  Modal,
  Switch,
  Tag,
  Popconfirm,
  Tabs,
  Select,
  Row,
  Col
} from 'antd'
import {
  SaveOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
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

interface DataSource {
  name: string
  id: string
  endpoint: string
  username: string
  password: string
  enabled: boolean
  container_prefix?: string[]
  labels?: Record<string, string>
}

export default function VictoriaMetrics() {
  const [form] = Form.useForm()
  const [datasourceForm] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [healthStatus, setHealthStatus] = useState<Record<string, 'success' | 'error'>>({})
  const [datasources, setDatasources] = useState<DataSource[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [editingDatasource, setEditingDatasource] = useState<DataSource | null>(null)
  const [activeTab, setActiveTab] = useState('multi')
  const token = useAuthStore((state) => state.token)

  useEffect(() => {
    loadConfigs()
    loadDatasources()
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
      }
    } catch (error) {
      console.error('Failed to load configs:', error)
      message.error('加载配置失败')
    } finally {
      setLoading(false)
    }
  }

  const loadDatasources = async () => {
    try {
      const response = await fetch('/api/v1/configs/category/monitoring', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) return

      const result = await response.json()
      if (result.code === 0 && result.data) {
        const configs: SystemConfig[] = result.data
        const datasourceConfig = configs.find(c => c.key === 'victoriametrics_datasources')

        if (datasourceConfig?.value) {
          try {
            const data = JSON.parse(datasourceConfig.value)
            setDatasources(Array.isArray(data) ? data : [])
          } catch (error) {
            console.error('Failed to parse datasources:', error)
          }
        }
      }
    } catch (error) {
      console.error('Failed to load datasources:', error)
    }
  }

  const handleSaveSingle = async () => {
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
        message.success('单数据源配置保存成功，重启服务后生效')
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

  const handleSaveMulti = async () => {
    try {
      setLoading(true)

      // 清空单数据源配置（避免冲突）
      const clearConfigs = [
        { category: 'monitoring', key: 'victoriametrics_endpoint', value: '', description: '', is_encrypted: false },
        { category: 'monitoring', key: 'victoriametrics_username', value: '', description: '', is_encrypted: false },
        { category: 'monitoring', key: 'victoriametrics_password', value: '', description: '', is_encrypted: true },
      ]

      // 保存多数据源配置
      const datasourceConfig = {
        category: 'monitoring',
        key: 'victoriametrics_datasources',
        value: JSON.stringify(datasources, null, 2),
        description: 'VictoriaMetrics 多数据源配置',
        is_encrypted: true,
      }

      const response = await fetch('/api/v1/configs/batch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ configs: [...clearConfigs, datasourceConfig] }),
      })

      const result = await response.json()

      if (result.code === 0) {
        message.success('多数据源配置保存成功')
        await loadDatasources()
      } else {
        message.error(result.message || '保存配置失败')
      }
    } catch (error) {
      console.error('Failed to save multi-source configs:', error)
      message.error('保存配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAddDatasource = () => {
    setEditingDatasource(null)
    datasourceForm.resetFields()
    datasourceForm.setFieldsValue({
      enabled: true,
      container_prefix: [],
      labels: {}
    })
    setModalVisible(true)
  }

  const handleEditDatasource = (datasource: DataSource) => {
    setEditingDatasource(datasource)
    datasourceForm.setFieldsValue({
      ...datasource,
      labels: datasource.labels || {},
      container_prefix: datasource.container_prefix || []
    })
    setModalVisible(true)
  }

  const handleDeleteDatasource = async (id: string) => {
    try {
      const newDatasources = datasources.filter(ds => ds.id !== id)
      setDatasources(newDatasources)

      // 自动保存
      const response = await fetch('/api/v1/configs/batch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          configs: [{
            category: 'monitoring',
            key: 'victoriametrics_datasources',
            value: JSON.stringify(newDatasources, null, 2),
            description: 'VictoriaMetrics 多数据源配置',
            is_encrypted: true,
          }]
        }),
      })

      const result = await response.json()
      if (result.code === 0) {
        message.success('删除成功')
      } else {
        message.error(result.message || '删除失败')
        await loadDatasources() // 恢复
      }
    } catch (error) {
      console.error('Failed to delete datasource:', error)
      message.error('删除失败')
      await loadDatasources() // 恢复
    }
  }

  const handleSaveDatasource = async () => {
    try {
      const values = await datasourceForm.validateFields()

      let newDatasources: DataSource[]

      if (editingDatasource) {
        // 编辑模式
        newDatasources = datasources.map(ds =>
          ds.id === editingDatasource.id ? { ...values } : ds
        )
      } else {
        // 新增模式
        newDatasources = [...datasources, values]
      }

      setDatasources(newDatasources)
      setModalVisible(false)

      // 自动保存到数据库
      const response = await fetch('/api/v1/configs/batch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          configs: [{
            category: 'monitoring',
            key: 'victoriametrics_datasources',
            value: JSON.stringify(newDatasources, null, 2),
            description: 'VictoriaMetrics 多数据源配置',
            is_encrypted: true,
          }]
        }),
      })

      const result = await response.json()
      if (result.code === 0) {
        message.success(editingDatasource ? '更新成功' : '添加成功')
      } else {
        message.error(result.message || '保存失败')
        await loadDatasources()
      }
    } catch (error) {
      console.error('Failed to save datasource:', error)
      message.error('保存失败')
    }
  }

  const handleTestConnection = async (datasource?: DataSource) => {
    const key = datasource?.id || 'default'

    setTesting(true)

    try {
      const response = await fetch('/api/v1/monitoring/victoriametrics/health', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      const result = await response.json()

      if (result.code === 0) {
        setHealthStatus(prev => ({ ...prev, [key]: 'success' }))
        message.success(datasource ? `${datasource.name} 连接成功` : '连接测试成功')
      } else {
        setHealthStatus(prev => ({ ...prev, [key]: 'error' }))
        message.error(datasource ? `${datasource.name} 连接失败: ${result.message}` : `连接测试失败: ${result.message}`)
      }
    } catch (error) {
      setHealthStatus(prev => ({ ...prev, [key]: 'error' }))
      message.error('连接测试失败')
    } finally {
      setTesting(false)
    }
  }

  const columns = [
    {
      title: '数据源名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <strong>{text}</strong>
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
      )
    },
    {
      title: 'Endpoint',
      dataIndex: 'endpoint',
      key: 'endpoint',
      ellipsis: true,
      render: (text: string) => <code className="text-xs">{text}</code>
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: DataSource) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<CheckCircleOutlined />}
            onClick={() => handleTestConnection(record)}
          >
            测试
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditDatasource(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteDatasource(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">
          VictoriaMetrics 配置
        </h1>
        <p className="text-gray-600 dark:text-text-secondary">
          配置多个VictoriaMetrics数据源，用于容器监控和自动同步
        </p>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'multi',
            label: <span><DatabaseOutlined /> 多数据源配置</span>,
            children: (
              <Card>
                <Alert
                  message="多数据源配置"
                  description="配置多个VictoriaMetrics数据源后，系统会自动从所有启用的数据源同步容器。修改配置后需要重启 CMDB 服务才能生效。"
                  type="info"
                  showIcon
                  className="mb-4"
                />

                <div className="mb-4 flex justify-end">
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={handleAddDatasource}
                  >
                    添加数据源
                  </Button>
                </div>

                <Table
                  columns={columns}
                  dataSource={datasources}
                  rowKey="id"
                  pagination={false}
                  size="middle"
                />

                <Divider className="my-4" />

                <div className="flex justify-between items-center">
                  <Space>
                    <span className="text-sm text-gray-500">
                      共 {datasources.length} 个数据源，其中 {datasources.filter(d => d.enabled).length} 个已启用
                    </span>
                  </Space>
                  <Space>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={loadDatasources}
                      disabled={loading}
                    >
                      刷新
                    </Button>
                    <Button
                      type="primary"
                      icon={<SaveOutlined />}
                      onClick={handleSaveMulti}
                      loading={loading}
                    >
                      保存配置
                    </Button>
                  </Space>
                </div>
              </Card>
            )
          },
          {
            key: 'single',
            label: <span><ApiOutlined /> 单数据源配置</span>,
            children: (
              <Card>
                <Alert
                  message="单数据源配置"
                  description="配置单个VictoriaMetrics实例。推荐使用多数据源配置以获得更好的灵活性和容错能力。"
                  type="warning"
                  showIcon
                  className="mb-6"
                />

                <Form form={form} layout="vertical" onFinish={handleSaveSingle}>
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
                      onClick={() => handleTestConnection()}
                      loading={testing}
                      disabled={loading}
                    >
                      测试连接
                    </Button>
                  </Space>

                  {healthStatus.default && (
                    <Alert
                      message={healthStatus.default === 'success' ? '连接成功' : '连接失败'}
                      type={healthStatus.default}
                      showIcon
                      className="mt-4"
                    />
                  )}
                </Form>
              </Card>
            )
          }
        ]}
      />

      {/* 添加/编辑数据源对话框 */}
      <Modal
        title={editingDatasource ? '编辑数据源' : '添加数据源'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => datasourceForm.submit()}
        width={700}
        destroyOnClose
      >
        <Form
          form={datasourceForm}
          layout="vertical"
          onFinish={handleSaveDatasource}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="数据源名称"
                name="name"
                rules={[{ required: true, message: '请输入数据源名称' }]}
              >
                <Input placeholder="例如: 主数据中心" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="数据源ID"
                name="id"
                rules={[
                  { required: true, message: '请输入数据源ID' },
                  { pattern: /^[a-zA-Z0-9-_]+$/, message: '只能包含字母、数字、中划线和下划线' }
                ]}
              >
                <Input placeholder="例如: primary-dc" disabled={!!editingDatasource} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Endpoint"
            name="endpoint"
            rules={[
              { required: true, message: '请输入 VictoriaMetrics endpoint' },
              { type: 'url', message: '请输入有效的 URL' }
            ]}
          >
            <Input placeholder="https://victoriametrics.example.com:8429" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="用户名"
                name="username"
              >
                <Input placeholder="admin" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="密码"
                name="password"
              >
                <Input.Password placeholder="请输入密码" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="容器名前缀过滤"
            name="container_prefix"
            tooltip="只同步匹配这些前缀的容器，留空表示同步所有容器"
          >
            <Select
              mode="tags"
              placeholder="输入前缀后按回车添加，例如: prod-, staging-"
              style={{ width: '100%' }}
            >
            </Select>
          </Form.Item>

          <Form.Item
            label="自动标签"
            name="labels"
            tooltip="这些标签会自动添加到从该数据源同步的所有容器上"
          >
            <Input.TextArea
              rows={3}
              placeholder='JSON格式，例如: {"location": "主数据中心", "environment": "production"}'
            />
          </Form.Item>

          <Form.Item
            label="状态"
            name="enabled"
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
