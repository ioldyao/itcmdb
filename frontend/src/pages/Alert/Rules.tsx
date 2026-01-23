import { Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

export default function AlertRules() {
  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>告警规则</h2>
        <Button type="primary" icon={<PlusOutlined />}>创建规则</Button>
      </div>
      <Table columns={[]} dataSource={[]} rowKey="id" />
    </div>
  )
}
