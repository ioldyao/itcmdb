import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Spin, Alert, Progress } from 'antd'
import { Activity, Cpu, HardDrive, Network } from 'lucide-react'

interface MonitoringDataProps {
  ciId: number
}

interface ContainerStats {
  container_id: string
  cpu_usage_percent: number
  memory_usage_mb: number
  memory_limit_mb: number
  network_rx_bytes: number
  network_tx_bytes: number
  disk_usage_mb: number
  uptime_seconds: number
  timestamp: number
}

export default function ContainerMonitoring({ ciId }: MonitoringDataProps) {
  const [stats, setStats] = useState<ContainerStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!ciId) {
      setLoading(false)
      return
    }

    fetchMonitoringData()
    // 每30秒刷新一次监控数据
    const interval = setInterval(fetchMonitoringData, 30000)
    return () => clearInterval(interval)
  }, [ciId])

  const fetchMonitoringData = async () => {
    try {
      setLoading(true)
      setError(null)

      const token = localStorage.getItem('token')
      const response = await fetch(`/api/v1/monitoring/containers/${ciId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      if (result.code === 200 && result.data) {
        setStats(result.data)
      } else {
        throw new Error(result.message || '获取监控数据失败')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '无法获取监控数据')
      console.error('Failed to fetch monitoring data:', err)
    } finally {
      setLoading(false)
    }
  }

  if (!ciId) {
    return (
      <Alert
        message="监控未配置"
        description="请配置容器ID和cAdvisor端点以启用监控"
        type="info"
        showIcon
      />
    )
  }

  if (loading && !stats) {
    return (
      <div className="flex justify-center items-center p-8">
        <Spin size="large" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert
        message="监控数据获取失败"
        description={error}
        type="error"
        showIcon
      />
    )
  }

  if (!stats) {
    return null
  }

  const memoryUsagePercent = (stats.memory_usage_mb / stats.memory_limit_mb) * 100
  const uptimeHours = Math.floor(stats.uptime_seconds / 3600)
  const uptimeDays = Math.floor(uptimeHours / 24)

  return (
    <div className="space-y-4">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="CPU 使用率"
              value={stats.cpu_usage_percent.toFixed(2)}
              suffix="%"
              prefix={<Cpu size={20} className="text-blue-500" />}
            />
            <Progress
              percent={stats.cpu_usage_percent}
              strokeColor={stats.cpu_usage_percent > 80 ? '#ff4d4f' : '#1890ff'}
              showInfo={false}
              size="small"
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="内存使用"
              value={stats.memory_usage_mb.toFixed(0)}
              suffix={`MB / ${stats.memory_limit_mb} MB`}
              prefix={<HardDrive size={20} className="text-green-500" />}
            />
            <Progress
              percent={memoryUsagePercent}
              strokeColor={memoryUsagePercent > 80 ? '#ff4d4f' : '#52c41a'}
              showInfo={false}
              size="small"
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="网络接收"
              value={(stats.network_rx_bytes / 1024 / 1024).toFixed(2)}
              suffix="MB"
              prefix={<Network size={20} className="text-purple-500" />}
            />
            <div className="text-xs text-gray-500 mt-2">
              发送: {(stats.network_tx_bytes / 1024 / 1024).toFixed(2)} MB
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="运行时间"
              value={uptimeDays > 0 ? uptimeDays : uptimeHours}
              suffix={uptimeDays > 0 ? '天' : '小时'}
              prefix={<Activity size={20} className="text-orange-500" />}
            />
            <div className="text-xs text-gray-500 mt-2">
              磁盘: {stats.disk_usage_mb.toFixed(0)} MB
            </div>
          </Card>
        </Col>
      </Row>

      <div className="text-xs text-gray-400 text-right">
        容器ID: {stats.container_id} | 自动刷新: 30秒 | 更新时间: {new Date(stats.timestamp * 1000).toLocaleString()}
      </div>
    </div>
  )
}
