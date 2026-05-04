import { useState } from 'react'
import { Badge, Collapse, Checkbox, Space } from 'antd'
import {
  UserOutlined,
  StarOutlined,
  BellOutlined,
  StopOutlined,
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
      className={`
        flex items-center justify-between px-4 py-2 cursor-pointer
        transition-colors duration-150
        ${active
          ? 'bg-gray-100 dark:bg-white/10'
          : 'hover:bg-gray-50 dark:hover:bg-white/5'
        }
      `}
    >
      <Space size={8}>
        {icon}
        <span className="text-sm">{label}</span>
      </Space>
      <span className="text-sm text-gray-400 dark:text-text-tertiary">{count}</span>
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
  const colorClass = color === '#ff4d4f' ? 'text-red-500'
    : color === '#fa8c16' ? 'text-orange-500'
    : color === '#faad14' ? 'text-yellow-500'
    : ''

  return (
    <div className="flex items-center justify-between px-4 py-1.5 cursor-pointer hover:bg-gray-50 dark:hover:bg-white/5 transition-colors">
      <label className="flex items-center gap-2 cursor-pointer flex-1">
        <Checkbox value={value} checked={checked} onChange={onChange} />
        {color && <div className="w-1 h-4 rounded-sm" style={{ background: color }} />}
        <span className={`text-sm ${colorClass}`}>{label}</span>
      </label>
      <span className="text-sm text-gray-400 dark:text-text-tertiary">{count}</span>
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

  const stats = statistics?.stats || { total: 0, firing: 0, acknowledged: 0, resolved: 0, closed: 0 }
  const severityStats = statistics?.severity_stats || []

  const handleSidebarItemClick = (key: string, value: any) => {
    setActiveItem(key)
    if (onFilterChange) {
      onFilterChange({ [key]: value })
    }
  }

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
    <div className="w-[260px] h-full border-r border-gray-200 dark:border-white/8 bg-white dark:bg-bg-secondary overflow-y-auto flex flex-col">
      {/* 告警部分 */}
      <div className="border-b border-gray-200 dark:border-white/8">
        <Collapse
          defaultActiveKey={['alerts']}
          bordered={false}
          expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
          className="bg-transparent dark:bg-transparent"
        >
          <Panel
            header={
              <div className="flex items-center justify-between">
                <Space>
                  <DownOutlined className="text-xs text-gray-400 dark:text-text-tertiary" />
                  <span className="font-medium text-gray-900 dark:text-text-primary">告警</span>
                </Space>
                <Badge count={stats.total} className="dark:bg-blue-600" />
              </div>
            }
            key="alerts"
            className="!border-none !p-0"
          >
            <SidebarItem
              icon={<UserOutlined className="text-blue-500 text-sm" />}
              label="我负责的"
              count={0}
              active={activeItem === 'assigned'}
              onClick={() => handleSidebarItemClick('assigned', true)}
            />
            <SidebarItem
              icon={<StarOutlined className="text-green-500 text-sm" />}
              label="我关注的"
              count={0}
              active={activeItem === 'watched'}
              onClick={() => handleSidebarItemClick('watched', true)}
            />
            <SidebarItem
              icon={<BellOutlined className="text-sm" />}
              label="我收到的"
              count={0}
              active={activeItem === 'received'}
              onClick={() => handleSidebarItemClick('received', true)}
            />
            <SidebarItem
              icon={<AlertOutlined className="text-red-500 text-sm" />}
              label="未恢复"
              count={stats.firing}
              active={activeItem === 'firing'}
              onClick={() => handleSidebarItemClick('status', 'firing')}
            />
            <SidebarItem
              icon={<StopOutlined className="text-sm" />}
              label="未恢复(已屏蔽)"
              count={0}
              active={activeItem === 'suppressed'}
              onClick={() => handleSidebarItemClick('suppressed', true)}
            />
            <SidebarItem
              icon={<CheckCircleOutlined className="text-green-500 text-sm" />}
              label="已恢复"
              count={stats.resolved}
              active={activeItem === 'resolved'}
              onClick={() => handleSidebarItemClick('status', 'resolved')}
            />
          </Panel>
        </Collapse>
      </div>

      {/* 处理记录 */}
      <div className="border-b border-gray-200 dark:border-white/8">
        <Collapse
          bordered={false}
          expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
          className="bg-transparent dark:bg-transparent"
        >
          <Panel
            header={
              <div className="flex items-center justify-between">
                <Space>
                  <RightOutlined className="text-xs text-gray-400 dark:text-text-tertiary" />
                  <span className="font-medium text-gray-900 dark:text-text-primary">处理记录</span>
                </Space>
                <span className="text-sm text-gray-400 dark:text-text-tertiary">0</span>
              </div>
            }
            key="history"
            className="!border-none"
          >
            <div className="px-4 py-2 text-sm text-gray-400 dark:text-text-tertiary">
              查看历史处理记录
            </div>
          </Panel>
        </Collapse>
      </div>

      {/* 高级筛选 */}
      <div className="px-4 py-3 border-b border-gray-200 dark:border-white/8">
        <span className="text-sm font-medium text-gray-600 dark:text-text-secondary">高级筛选</span>
      </div>

      {/* 级别筛选 */}
      <Collapse
        defaultActiveKey={['severity']}
        bordered={false}
        expandIcon={({ isActive }) => (isActive ? <DownOutlined /> : <RightOutlined />)}
        className="bg-transparent dark:bg-transparent"
      >
        <Panel header="级别" key="severity" className="!border-none !px-4">
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
                onChange={() => {
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

        <Panel header="处理阶段" key="handling" className="!border-none !px-4">
          <FilterCheckbox label="已通知" count={0} value="notified" />
          <FilterCheckbox label="已确认" count={0} value="acknowledged" />
          <FilterCheckbox label="已屏蔽" count={0} value="suppressed" />
          <FilterCheckbox label="已流控" count={0} value="throttled" />
        </Panel>

        <Panel header="数据类型" key="dataType" className="!border-none !px-4">
          <FilterCheckbox label="时序数据" count={0} value="metric" />
          <FilterCheckbox label="事件" count={0} value="event" />
          <FilterCheckbox label="日志" count={0} value="log" />
        </Panel>

        <Panel header="分类" key="category" className="!border-none !px-4">
          <div className="py-2 text-sm text-gray-400 dark:text-text-tertiary">加载中...</div>
        </Panel>
      </Collapse>
    </div>
  )
}
