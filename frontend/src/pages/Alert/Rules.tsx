import { Tabs } from 'antd'
import { Bell, Route, FileText } from 'lucide-react'
import AlertRulesTab from './AlertRulesTab'
import RoutingRulesTab from './RoutingRulesTab'
import NotificationTemplatesTab from './NotificationTemplatesTab'

export default function AlertRules() {
  const items = [
    {
      key: 'rules',
      label: (
        <span className="flex items-center gap-2">
          <Bell size={16} />
          告警规则
        </span>
      ),
      children: <AlertRulesTab />,
    },
    {
      key: 'routing',
      label: (
        <span className="flex items-center gap-2">
          <Route size={16} />
          路由规则
        </span>
      ),
      children: <RoutingRulesTab />,
    },
    {
      key: 'templates',
      label: (
        <span className="flex items-center gap-2">
          <FileText size={16} />
          通知模板
        </span>
      ),
      children: <NotificationTemplatesTab />,
    },
  ]

  return (
    <div className="p-6">
      <h2 className="text-2xl font-semibold mb-4">规则配置</h2>
      <Tabs items={items} />
    </div>
  )
}
