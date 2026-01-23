import { useParams } from 'react-router-dom'
import { Card, Descriptions, Button } from 'antd'

export default function TicketDetail() {
  const { id } = useParams()

  return (
    <div>
      <h2>工单详情 #{id}</h2>
      <Card>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="标题">服务器CPU异常</Descriptions.Item>
          <Descriptions.Item label="状态">处理中</Descriptions.Item>
          <Descriptions.Item label="优先级">高</Descriptions.Item>
          <Descriptions.Item label="创建时间">2026-01-24 10:00</Descriptions.Item>
          <Descriptions.Item label="处理人">张三</Descriptions.Item>
          <Descriptions.Item label="提交人">李四</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            服务器CPU使用率持续超过90%
          </Descriptions.Item>
        </Descriptions>
        <div style={{ marginTop: 16 }}>
          <Button>确认</Button>
          <Button>关闭</Button>
        </div>
      </Card>
    </div>
  )
}
