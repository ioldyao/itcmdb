import { useState } from 'react'
import { Form, Input, Select, Button, Card, message } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { ticketService } from '@/services/ticket'

export default function TicketCreate() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const handleSubmit = async (values: any) => {
    setLoading(true)
    try {
      const res = await ticketService.createTicket(values)
      const data = res as any
      if (data.code === 0) {
        message.success('工单创建成功')
        navigate('/tickets')
      } else {
        message.error(data.message || '创建失败')
      }
    } catch (error) {
      message.error('创建失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8 max-w-3xl mx-auto">
      {/* 页面头部 */}
      <div className="mb-6 flex items-center gap-4">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
          返回
        </Button>
        <div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">创建工单</h1>
          <p className="text-gray-600 dark:text-text-secondary">提交新的工单请求</p>
        </div>
      </div>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ priority: 'medium' }}
        >
          <Form.Item
            label="工单标题"
            name="title"
            rules={[{ required: true, message: '请输入工单标题' }]}
          >
            <Input placeholder="请输入工单标题" />
          </Form.Item>

          <Form.Item
            label="工单类型"
            name="type"
            rules={[{ required: true, message: '请选择工单类型' }]}
          >
            <Select placeholder="请选择工单类型">
              <Select.Option value="incident">故障工单</Select.Option>
              <Select.Option value="request">服务请求</Select.Option>
              <Select.Option value="change">变更申请</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="优先级"
            name="priority"
            rules={[{ required: true, message: '请选择优先级' }]}
          >
            <Select placeholder="请选择优先级">
              <Select.Option value="low">低</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="critical">紧急</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="详细描述"
            name="description"
            rules={[{ required: true, message: '请输入详细描述' }]}
          >
            <Input.TextArea rows={6} placeholder="请详细描述问题内容" />
          </Form.Item>

          <Form.Item>
            <div className="flex gap-2">
              <Button type="primary" htmlType="submit" loading={loading}>
                提交工单
              </Button>
              <Button onClick={() => navigate('/tickets')}>
                取消
              </Button>
            </div>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
