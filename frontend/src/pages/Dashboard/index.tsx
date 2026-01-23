import { Row, Col, Card, Statistic } from 'antd'
import {
  ServerOutlined,
  CustomerServiceOutlined,
  AlertOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'

export default function Dashboard() {
  return (
    <div>
      <h2>仪表板</h2>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic
              title="服务器总数"
              value={156}
              prefix={<ServerOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="进行中工单"
              value={23}
              prefix={<CustomerServiceOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃告警"
              value={7}
              prefix={<AlertOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="本月已解决"
              value={89}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
