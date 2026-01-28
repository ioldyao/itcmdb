import { useState } from 'react'
import { Card, Badge, Collapse, Checkbox, Space, Divider } from 'antd'
import {
  UserOutlined,
  StarOutlined,
  BellOutlined,
  BellSlashOutlined,
  CheckCircleOutlined,
  AlertOutlined,
  RightOutlined,
  DownOutlined,
} from '@ant-design/icons'
import type { CheckboxChangeEvent } from 'antd/es/checkbox'
import { useAlertStore } from '@/stores/alertStore'

const { Panel } = Collapse

interface SidebarItemProps {
  icon: React.ReactNode
  label: string
  count: number
  active?: boolean
  onClick?: () => void
}

function SidebarItem({ icon, label, count, active = false, onClick }: SidebarItemProps) {
  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '8px 16px',
        cursor: 'pointer',
        background: active ? '#f0f0f0' : 'transparent',
      }}
      onMouseEnter={(e) => {
        if (!active) {
          e.currentTarget.style.background = '#fafafa'
        }
      }}
      onMouseLeave={(e) => {
        if (!active) {
          e.currentTarget.style.background = 'transparent'
        }
      }}
    >
      <Space size={8}>
        {icon}
        <span style={{ fontSize: 14 }}>{label}</span>
      </Space>
      <span style={{ fontSize: 14, color: '#999' }}>{count}</span>
    </div>
  )
}

interface FilterCheckboxProps {
  label: string
  count: number
  value: string
  checked?: boolean
  onChange?: (e: CheckboxChangeEvent) => void
  color?: string
}

function FilterCheckbox({ label, count, value, checked, onChange, color }: FilterCheckboxProps) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '6px 16px',
        cursor: 'pointer',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = '#fafafa'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent'
      }}
    >
      <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', flex: 1 }}>
        <Checkbox value={value} checked={checked} onChange={onChange} />
        {color && <div style={{ width: 4, height: 16, background: color, borderRadius: 2 }} />}
        <span
          style={{
            fontSize: 14,
            ...(color === '#ff4d4f' ? { color: '#ff4d4f' } : {}),
            ...(color === '#fa8c16' ? { color: '#fa8c16' } : {}),
            ...(color === '#faad14' ? { color: '#faad14' } : {}),
          }}
        >
          {label}
        </span>
      </label>
      <span style={{ fontSize: 14, color: '#999' }}>{count}</span>
    </div>
  )
}

interface AlertSidebarProps {
  onFilterChange?: (filters: any) => void
  collapsed?: boolean
}

export default function AlertSidebar({ onFilterChange, collapsed = false }: AlertSidebarProps) {
  const { statistics, filters, setFilters } = useAlertStore()
  const [activeItem, setActiveItem] = useState<string>('all')

  // 计算各分类数量（实际应从API获取）
  const stats = statistics?.stats || { total: 0, firing: 0, acknowledged: 0, resolved: 0, closed: 0 }
  const severityStats = statistics?.severity_stats || []

  // 处理边栏项点击
  const handleSidebarItemClick = (key: string, value: any) => {
    setActiveItem(key)
    if (onFilterChange) {
      onFilterChange({ [key]: value })
    }
  }

  // 处理复选框变化
  const handleCheckboxChange = (type: string, values: string[]) => {
    const newFilters = { ...filters }
    if (values.length === 0) {
      delete newFilters[type as keyof typeof newFilters]
    } else {
      ;(newFilters as any)[type] = values
    }
    setFilters(newFilters)
    if (onFilterChange) {
      onFilterChange(newFilters)
    }
  }

  if (collapsed) {
    return null
  }

  return (
    <div
      style={{
        width: 260,
        height: '100%',
        borderRight: '1px solid #f0f0f0',
        background: '#fff',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 告警部分 */}
      <div style={{ borderBottom: '1px solid #f0f0f0' }}>
        <Collapse
          defaultActiveKey={['alerts']}
          bordered={false}
          expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
          style={{ background: 'transparent' }}
        >
          <Panel
            header={
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Space>
                  <DownOutlined style={{ fontSize: 12, color: '#999' }} />
                  <span style={{ fontWeight: 500 }}>告警</span>
                </Space>
                <Badge count={stats.total} style={{ background: '#1890ff' }} />
              </div>
            }
            key="alerts"
            style={{ border: 'none', padding: 0 }}
          >
            <SidebarItem
              icon={<UserOutlined style={{ color: '#1890ff', fontSize: 14 }} />}
              label="我负责的"
              count={0}
              active={activeItem === 'assigned'}
              onClick={() => handleSidebarItemClick('assigned', true)}
            />
            <SidebarItem
              icon={<StarOutlined style={{ color: '#52c41a', fontSize: 14 }} />}
              label="我关注的"
              count={0}
              active={activeItem === 'watched'}
              onClick={() => handleSidebarItemClick('watched', true)}
            />
            <SidebarItem
              icon={<BellOutlined style={{ fontSize: 14 }} />}
              label="我收到的"
              count={0}
              active={activeItem === 'received'}
              onClick={() => handleSidebarItemClick('received', true)}
            />
            <SidebarItem
              icon={<AlertOutlined style={{ color: '#ff4d4f', fontSize: 14 }} />}
              label="未恢复"
              count={stats.firing}
              active={activeItem === 'firing'}
              onClick={() => handleSidebarItemClick('status', 'firing')}
            />
            <SidebarItem
              icon={<BellSlashOutlined style={{ fontSize: 14 }} />}
              label="未恢复(已屏蔽)"
              count={0}
              active={activeItem === 'suppressed'}
              onClick={() => handleSidebarItemClick('suppressed', true)}
            />
            <SidebarItem
              icon={<CheckCircleOutlined style={{ color: '#52c41a', fontSize: 14 }} />}
              label="已恢复"
              count={stats.resolved}
              active={activeItem === 'resolved'}
              onClick={() => handleSidebarItemClick('status', 'resolved')}
            />
          </Panel>
        </Collapse>
      </div>

      {/* 处理记录 */}
      <div style={{ borderBottom: '1px solid #f0f0f0' }}>
        <Collapse
          bordered={false}
          expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
          style={{ background: 'transparent' }}
        >
          <Panel
            header={
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Space>
                  <RightOutlined style={{ fontSize: 12, color: '#999' }} />
                  <span style={{ fontWeight: 500 }}>处理记录</span>
                </Space>
                <span style={{ fontSize: 14, color: '#999' }}>0</span>
              </div>
            }
            key="history"
            style={{ border: 'none' }}
          >
            <div style={{ padding: '8px 16px', fontSize: 14, color: '#999' }}>
              查看历史处理记录
            </div>
          </Panel>
        </Collapse>
      </div>

      {/* 高级筛选 */}
      <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
        <span style={{ fontSize: 14, fontWeight: 500, color: '#666' }}>高级筛选</span>
      </div>

      {/* 级别筛选 */}
      <Collapse
        defaultActiveKey={['severity']}
        bordered={false}
        expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
        style={{ background: 'transparent' }}
      >
        <Panel header="级别" key="severity" style={{ border: 'none', padding: '0 16px' }}>
          {severityStats.map((stat: any) => {
            const config: Record<string, { label: string; color: string }> = {
              critical: { label: '致命', color: '#ff4d4f' },
              high: { label: '预警', color: '#fa8c16' },
              medium: { label: '提醒', color: '#faad14' },
              low: { label: '低', color: '#1890ff' },
            }
            const cfg = config[stat.severity] || { label: stat.severity, color: '#999' }
            const checked = (filters.severity as string[])?.includes(stat.severity)
            return (
              <FilterCheckbox
                key={stat.severity}
                label={cfg.label}
                count={stat.count}
                value={stat.severity}
                checked={checked}
                onChange={(e) => {
                  const newValues = checked
                    ? (filters.severity as string[]).filter((s) => s !== stat.severity)
                    : [...(filters.severity as string[] || []), stat.severity]
                  handleCheckboxChange('severity', newValues)
                }}
                color={cfg.color}
              />
            )
          })}
        </Panel>

        <Panel header="处理阶段" key="handling" style={{ border: 'none', padding: '0 16px' }}>
          <FilterCheckbox label="已通知" count={0} value="notified" />
          <FilterCheckbox label="已确认" count={0} value="acknowledged" />
          <FilterCheckbox label="已屏蔽" count={0} value="suppressed" />
          <FilterCheckbox label="已流控" count={0} value="throttled" />
        </Panel>

        <Panel header="数据类型" key="dataType" style={{ border: 'none', padding: '0 16px' }}>
          <FilterCheckbox label="时序数据" count={0} value="metric" />
          <FilterCheckbox label="事件" count={0} value="event" />
          <FilterCheckbox label="日志" count={0} value="log" />
        </Panel>

        <Panel header="分类" key="category" style={{ border: 'none', padding: '0 16px' }}>
          <div style={{ padding: '8px 0', fontSize: 14, color: '#999' }}>加载中...</div>
        </Panel>
      </Collapse>
    </div>
  )
}
