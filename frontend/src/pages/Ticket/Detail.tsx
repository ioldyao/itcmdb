import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Descriptions, Tag, Button, Spin, message, Space, Input, Empty } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { ticketService, Ticket } from '@/services/ticket'
import dayjs from 'dayjs'

const statusMap: Record<string, { text: string; color: string }> = {
  open: { text: '待处理', color: 'blue' },
  in_progress: { text: '处理中', color: 'orange' },
  resolved: { text: '已解决', color: 'green' },
  closed: { text: '已关闭', color: 'default' },
}

const priorityMap: Record<string, { text: string; color: string }> = {
  low: { text: '低', color: 'blue' },
  medium: { text: '中', color: 'orange' },
  high: { text: '高', color: 'red' },
  critical: { text: '紧急', color: 'magenta' },
}

const typeMap: Record<string, string> = {
  incident: '故障工单',
  request: '服务请求',
  change: '变更申请',
}

export default function TicketDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const [commentLoading, setCommentLoading] = useState(false)

  useEffect(() => {
    if (id) loadTicket(id)
  }, [id])

  const loadTicket = async (ticketId: string) => {
    setLoading(true)
    try {
      const res = await ticketService.getTicket(ticketId)
      const data = res as any
      if (data.code === 0 && data.data) {
        setTicket(data.data)
      } else {
        message.error('工单不存在')
        navigate('/tickets')
      }
    } catch (error) {
      message.error('加载工单失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAddComment = async () => {
    if (!comment.trim() || !id) return
    setCommentLoading(true)
    try {
      await ticketService.addComment(id, comment)
      message.success('评论添加成功')
      setComment('')
      loadTicket(id)
    } catch (error) {
      message.error('添加评论失败')
    } finally {
      setCommentLoading(false)
    }
  }

  const handleStatusChange = async (status: string) => {
    if (!id) return
    try {
      await ticketService.updateStatus(id, status)
      message.success('状态更新成功')
      loadTicket(id)
    } catch (error) {
      message.error('更新状态失败')
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (!ticket) {
    return (
      <div className="p-8">
        <Empty description="工单不存在">
          <Button type="primary" onClick={() => navigate('/tickets')}>
            返回工单列表
          </Button>
        </Empty>
      </div>
    )
  }

  const statusInfo = statusMap[ticket.status] || { text: ticket.status, color: 'default' }
  const priorityInfo = priorityMap[ticket.priority] || { text: ticket.priority, color: 'default' }

  return (
    <div className="p-8">
      {/* 头部 */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
            返回
          </Button>
          <div>
            <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary">
              {ticket.title}
            </h1>
            <div className="flex items-center gap-2 mt-1">
              <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
              <Tag color={priorityInfo.color}>{priorityInfo.text}</Tag>
            </div>
          </div>
        </div>
        <Space>
          {ticket.status === 'open' && (
            <Button type="primary" onClick={() => handleStatusChange('in_progress')}>
              开始处理
            </Button>
          )}
          {ticket.status === 'in_progress' && (
            <Button type="primary" onClick={() => handleStatusChange('resolved')}>
              标记已解决
            </Button>
          )}
          {ticket.status !== 'closed' && (
            <Button onClick={() => handleStatusChange('closed')}>
              关闭工单
            </Button>
          )}
        </Space>
      </div>

      {/* 详情 */}
      <Card>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="工单ID">{ticket.id}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="优先级">
            <Tag color={priorityInfo.color}>{priorityInfo.text}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="类型">{typeMap[ticket.description] || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {ticket.createdAt ? dayjs(ticket.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {ticket.updatedAt ? dayjs(ticket.updatedAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            {ticket.description || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 评论区 */}
      <Card title="处理记录" className="mt-4">
        <div className="mb-4">
          <Input.TextArea
            rows={3}
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="添加评论..."
          />
          <Button
            type="primary"
            className="mt-2"
            loading={commentLoading}
            onClick={handleAddComment}
            disabled={!comment.trim()}
          >
            提交评论
          </Button>
        </div>
        <Empty description="暂无处理记录" />
      </Card>
    </div>
  )
}
