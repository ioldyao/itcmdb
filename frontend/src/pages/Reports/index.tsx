import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Spin, Button } from 'antd'
import { Server, FileText, Bell } from 'lucide-react'
import { ReloadOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAuthStore } from '@/stores/authStore'
import { useNavigate } from 'react-router-dom'

export default function Reports() {
  const { token } = useAuthStore()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState({
    ciCount: 0,
    ticketCount: 0,
    alertCount: 0,
  })

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    setLoading(true)
    try {
      const headers = { Authorization: `Bearer ${token}` }

      const [ciRes, ticketRes, alertRes] = await Promise.allSettled([
        fetch('/api/v1/ci/instances?pageSize=1', { headers }).then(r => r.json()),
        fetch('/api/v1/reports/tickets/stats', { headers }).then(r => r.json()),
        fetch('/api/v1/alerts/statistics', { headers }).then(r => r.json()),
      ])

      const ciCount = ciRes.status === 'fulfilled' ? (ciRes.value?.data?.total || ciRes.value?.total || 0) : 0
      const ticketCount = ticketRes.status === 'fulfilled' ? (ticketRes.value?.data?.total || 0) : 0
      const alertCount = alertRes.status === 'fulfilled' ? (alertRes.value?.data?.stats?.total || 0) : 0

      setStats({ ciCount, ticketCount, alertCount })
    } catch (error) {
      console.error('Failed to fetch report stats:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">报表中心</h1>
          <p className="text-gray-600 dark:text-text-secondary">查看系统运营数据和统计报表</p>
        </div>
        <Button icon={<ReloadOutlined />} onClick={fetchStats} loading={loading}>
          刷新
        </Button>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <Spin size="large" tip="加载中..." />
        </div>
      ) : (
        <>
          {/* 统计概览 */}
          <Row gutter={[16, 16]} className="mb-6">
            <Col xs={24} sm={8}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
              >
                <Card hoverable onClick={() => navigate('/cmdb/servers')}>
                  <Statistic
                    title="资产总数"
                    value={stats.ciCount}
                    prefix={<Server size={20} className="text-blue-500" />}
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Card>
              </motion.div>
            </Col>
            <Col xs={24} sm={8}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
              >
                <Card hoverable onClick={() => navigate('/tickets')}>
                  <Statistic
                    title="工单总数"
                    value={stats.ticketCount}
                    prefix={<FileText size={20} className="text-orange-500" />}
                    valueStyle={{ color: '#fa8c16' }}
                  />
                </Card>
              </motion.div>
            </Col>
            <Col xs={24} sm={8}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                <Card hoverable onClick={() => navigate('/alerts')}>
                  <Statistic
                    title="告警总数"
                    value={stats.alertCount}
                    prefix={<Bell size={20} className="text-red-500" />}
                    valueStyle={{ color: '#ff4d4f' }}
                  />
                </Card>
              </motion.div>
            </Col>
          </Row>

          {/* 报表入口 */}
          <h2 className="text-xl font-semibold text-gray-900 dark:text-text-primary mb-4">报表列表</h2>
          <Row gutter={[16, 16]}>
            {[
              { title: 'CMDB资产报表', desc: '查看资产分布、状态统计', path: '/cmdb/servers', icon: Server, color: 'text-blue-500' },
              { title: '工单统计报表', desc: '工单处理效率、趋势分析', path: '/tickets', icon: FileText, color: 'text-orange-500' },
              { title: '告警趋势报表', desc: '告警频率、恢复率分析', path: '/alerts', icon: Bell, color: 'text-red-500' },
            ].map((item, index) => (
              <Col xs={24} sm={8} key={item.title}>
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.4 + index * 0.1 }}
                >
                  <Card
                    hoverable
                    onClick={() => navigate(item.path)}
                    className="h-full"
                  >
                    <div className="flex items-center gap-4">
                      <div className={`w-12 h-12 rounded-lg bg-gray-100 dark:bg-white/5 flex items-center justify-center`}>
                        <item.icon size={24} className={item.color} />
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900 dark:text-text-primary">{item.title}</h3>
                        <p className="text-sm text-gray-500 dark:text-text-secondary mt-1">{item.desc}</p>
                      </div>
                    </div>
                  </Card>
                </motion.div>
              </Col>
            ))}
          </Row>
        </>
      )}
    </div>
  )
}
