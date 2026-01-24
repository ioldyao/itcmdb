import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Tabs, Descriptions, Tag, Button, Spin, message } from 'antd'
import { ArrowLeftOutlined, EditOutlined } from '@ant-design/icons'
import { useCMDBStore, CIInstance } from '@/stores/cmdbStore'
import CIHistoryTimeline from '@/components/CMDB/CIHistoryTimeline'
import dayjs from 'dayjs'

export default function CIDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { fetchInstance } = useCMDBStore()
  const [instance, setInstance] = useState<CIInstance | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (id) {
      loadInstance(parseInt(id))
    }
  }, [id])

  const loadInstance = async (ciId: number) => {
    setLoading(true)
    try {
      const data = await fetchInstance(ciId)
      setInstance(data)
    } catch (error) {
      message.error('加载CI实例失败')
      console.error('Failed to load CI instance:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusConfig = (status: string) => {
    const configs: Record<string, { text: string; color: string }> = {
      active: { text: '在线', color: 'green' },
      inactive: { text: '离线', color: 'gray' },
      maintenance: { text: '维护', color: 'orange' },
      decommissioned: { text: '退役', color: 'red' },
    }
    return configs[status] || { text: status, color: 'default' }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (!instance) {
    return (
      <div className="p-8">
        <div className="text-center text-gray-500">未找到CI实例</div>
      </div>
    )
  }

  const statusConfig = getStatusConfig(instance.status)

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(-1)}
          >
            返回
          </Button>
          <div>
            <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary">
              {instance.name}
            </h1>
            <p className="text-gray-600 dark:text-text-secondary mt-1">
              {instance.ci_type?.display_name || instance.ci_type?.name}
            </p>
          </div>
        </div>
        <Button
          type="primary"
          icon={<EditOutlined />}
          onClick={() => navigate(`/cmdb/instances/${instance.id}/edit`)}
        >
          编辑
        </Button>
      </div>

      {/* 内容区域 */}
      <Card>
        <Tabs
          defaultActiveKey="basic"
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="ID">{instance.id}</Descriptions.Item>
                  <Descriptions.Item label="名称">{instance.name}</Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={statusConfig.color}>{statusConfig.text}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="类型">
                    {instance.ci_type?.display_name || instance.ci_type?.name}
                  </Descriptions.Item>
                  <Descriptions.Item label="创建时间">
                    {dayjs(instance.created_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>
                  <Descriptions.Item label="更新时间">
                    {dayjs(instance.updated_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>

                  {/* 动态属性 */}
                  {instance.attributes && Object.entries(instance.attributes).map(([key, value]) => (
                    <Descriptions.Item label={key} key={key}>
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </Descriptions.Item>
                  ))}
                </Descriptions>
              ),
            },
            {
              key: 'history',
              label: '变更历史',
              children: <CIHistoryTimeline ciId={instance.id} />,
            },
            {
              key: 'relations',
              label: '关系图谱',
              children: (
                <div className="text-center text-gray-500 py-8">
                  关系图谱功能开发中...
                </div>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}
