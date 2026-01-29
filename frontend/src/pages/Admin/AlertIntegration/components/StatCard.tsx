import { Card } from 'antd'

interface StatCardProps {
  value: number
  label: string
  color: string
}

export default function StatCard({ value, label, color }: StatCardProps) {
  return (
    <Card>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontSize: 32, fontWeight: 600, color }}>{value}</div>
        <div style={{ color: '#999', marginTop: 8 }}>{label}</div>
      </div>
    </Card>
  )
}
