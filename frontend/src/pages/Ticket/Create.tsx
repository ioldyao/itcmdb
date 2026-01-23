import { Form, Input, Select, Button, Card } from 'antd'

export default function TicketCreate() {
  return (
    <div>
      <h2>创建工单</h2>
      <Card>
        <Form layout="vertical">
          <Form.Item label="工单标题" name="title" rules={[{ required: true }]}>
            <Input placeholder="请输入工单标题" />
          </Form.Item>
          <Form.Item label="工单类型" name="type">
            <Select placeholder="请选择工单类型">
              <Select.Option value="incident">故障工单</Select.Option>
              <Select.Option value="request">服务请求</Select.Option>
              <Select.Option value="change">变更申请</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="优先级" name="priority">
            <Select placeholder="请选择优先级">
              <Select.Option value="low">低</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="critical">紧急</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="详细描述" name="description">
            <Input.TextArea rows={6} placeholder="请详细描述问题内容" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">提交工单</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
