import { useState } from 'react'
import { Badge, Collapse, Checkbox, Space } from 'antd'
import {
  UserOutlined,
  CheckCircleOutlined,
  AlertOutlined,
  RightOutlined,
  DownOutlined,
} from '@ant-design/icons'
import type { CheckboxChangeEvent } from 'antd/es/checkbox'
import { useAlertStore } from '@/stores/alertStore'
import { useAuthStore } from '@/stores/authStore'

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
  collapsed?: boolean
}

export default function AlertSidebar({ collapsed = false }: AlertSidebarProps) {
  const { statistics, filters, setFilters } = useAlertStore()
  const { user } = useAuthStore()
  const [activeItem, setActiveItem] = useState<string>('all')

  const stats = statistics?.stats || { total: 0, firing: 0, acknowledged: 0, resolved: 0, closed: 0 }
  const severityStats = statistics?.severity_stats || []

  // 点击侧边栏快捷项
  const handleSidebarItemClick = (key: string, value: any) => {
    // 切换同一个项时取消筛选
    if (activeItem === key) {
      setActiveItem('all')
      const newFilters = { ...filters }
      delete (newFilters as any)[key]
      setFilters(newFilters)
      return
    }
    setActiveItem(key)
    const newFilters: Record<string, any> = { ...filters }
    // 清除快捷筛选项
    delete newFilters.handler
    delete newFilters.handlingStatus
    delete newFilters.objectType
    delete newFilters.status

    if (key === 'handler') {
      newFilters.handler = value
    } else if (key === 'status') {
      newFilters.status = [value]
    } else if (key === 'handlingStatus') {
      newFilters.handlingStatus = value
    } else if (key === 'objectType') {
      newFilters.objectType = value
    }
    setFilters(newFilters)
  }

  // 处理复选框筛选
  const handleCheckboxChange = (type: string, values: string[]) => {
    const newFilters = { ...filters }
    if (values.length === 0) {
      delete (newFilters as any)[type]
    } else {
      ;(newFilters as any)[type] = values
    }
    setFilters(newFilters)
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
              count={stats.firing}
              active={activeItem === 'handler'}
              onClick={() => user?.id && handleSidebarItemClick('handler', user.id)}
            />
            <SidebarItem
              icon={<AlertOutlined className="text-red-500 text-sm" />}
              label="未恢复"
              count={stats.firing}
              active={activeItem === 'status' && filters.status?.[0] === 'firing'}
              onClick={() => handleSidebarItemClick('status', 'firing')}
            />
            <SidebarItem
              icon={<CheckCircleOutlined className="text-green-500 text-sm" />}
              label="已恢复"
              count={stats.resolved}
              active={activeItem === 'status' && filters.status?.[0] === 'resolved'}
              onClick={() => handleSidebarItemClick('status', 'resolved')}
            />
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
          {severityStats.length > 0 ? severityStats.map((stat: any) => {
            const config: Record<string, { label: string; color: string }> = {
              critical: { label: '致命', color: '#ff4d4f' },
              high: { label: '高', color: '#fa8c16' },
              medium: { label: '中', color: '#faad14' },
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
                  const current = (filters.severity as string[]) || []
                  const newValues = checked
                    ? current.filter((s) => s !== stat.severity)
                    : [...current, stat.severity]
                  handleCheckboxChange('severity', newValues)
                }}
                color={cfg.color}
              />
            )
          }) : (
            <>
              <FilterCheckbox
                label="致命" count={0} value="critical"
                checked={(filters.severity as string[])?.includes('critical')}
                onChange={() => {
                  const current = (filters.severity as string[]) || []
                  const newValues = current.includes('critical')
                    ? current.filter((s) => s !== 'critical')
                    : [...current, 'critical']
                  handleCheckboxChange('severity', newValues)
                }}
                color="#ff4d4f"
              />
              <FilterCheckbox
                label="高" count={0} value="high"
                checked={(filters.severity as string[])?.includes('high')}
                onChange={() => {
                  const current = (filters.severity as string[]) || []
                  const newValues = current.includes('high')
                    ? current.filter((s) => s !== 'high')
                    : [...current, 'high']
                  handleCheckboxChange('severity', newValues)
                }}
                color="#fa8c16"
              />
              <FilterCheckbox
                label="中" count={0} value="medium"
                checked={(filters.severity as string[])?.includes('medium')}
                onChange={() => {
                  const current = (filters.severity as string[]) || []
                  const newValues = current.includes('medium')
                    ? current.filter((s) => s !== 'medium')
                    : [...current, 'medium']
                  handleCheckboxChange('severity', newValues)
                }}
                color="#faad14"
              />
              <FilterCheckbox
                label="低" count={0} value="low"
                checked={(filters.severity as string[])?.includes('low')}
                onChange={() => {
                  const current = (filters.severity as string[]) || []
                  const newValues = current.includes('low')
                    ? current.filter((s) => s !== 'low')
                    : [...current, 'low']
                  handleCheckboxChange('severity', newValues)
                }}
                color="#1890ff"
              />
            </>
          )}
        </Panel>

        <Panel header="处理阶段" key="handling" className="!border-none !px-4">
          {[
            { label: '未处理', value: '' },
            { label: '已通知', value: 'notified' },
            { label: '已确认', value: 'acknowledged' },
            { label: '已屏蔽', value: 'suppressed' },
            { label: '已流控', value: 'throttled' },
          ].map((item) => (
            <FilterCheckbox
              key={item.value}
              label={item.label}
              count={0}
              value={item.value}
              checked={filters.handlingStatus === item.value || (!filters.handlingStatus && item.value === '')}
              onChange={() => {
                if (item.value === '') {
                  const newFilters = { ...filters }
                  delete newFilters.handlingStatus
                  setFilters(newFilters)
                } else {
                  setFilters({ handlingStatus: item.value })
                }
              }}
            />
          ))}
        </Panel>

        <Panel header="数据类型" key="dataType" className="!border-none !px-4">
          {[
            { label: '监控指标', value: 'metric' },
            { label: '事件', value: 'event' },
            { label: '日志', value: 'log' },
            { label: '外部告警', value: 'external' },
          ].map((item) => (
            <FilterCheckbox
              key={item.value}
              label={item.label}
              count={0}
              value={item.value}
              checked={filters.objectType === item.value}
              onChange={() => {
                if (filters.objectType === item.value) {
                  const newFilters = { ...filters }
                  delete newFilters.objectType
                  setFilters(newFilters)
                } else {
                  setFilters({ objectType: item.value })
                }
              }}
            />
          ))}
        </Panel>

        <Panel header="分类" key="category" className="!border-none !px-4">
          {['基础设施', '应用服务', '网络', '安全', '数据库'].map((cat) => (
            <FilterCheckbox
              key={cat}
              label={cat}
              count={0}
              value={cat}
              checked={filters.category === cat}
              onChange={() => {
                if (filters.category === cat) {
                  const newFilters = { ...filters }
                  delete newFilters.category
                  setFilters(newFilters)
                } else {
                  setFilters({ category: cat })
                }
              }}
            />
          ))}
        </Panel>
      </Collapse>
    </div>
  )
}
