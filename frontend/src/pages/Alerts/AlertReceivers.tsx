import { Tabs } from 'antd'
import { UserPlus, Users } from 'lucide-react'
import AdminAlertReceivers from '@/pages/Admin/AlertReceivers'
import AdminAlertReceiverGroups from '@/pages/Admin/AlertReceiverGroups'

export default function AlertReceivers() {
  const items = [
    {
      key: 'receivers',
      label: (
        <span className="flex items-center gap-2">
          <UserPlus size={16} />
          告警接收人
        </span>
      ),
      children: <AdminAlertReceivers />,
    },
    {
      key: 'groups',
      label: (
        <span className="flex items-center gap-2">
          <Users size={16} />
          告警接收组
        </span>
      ),
      children: <AdminAlertReceiverGroups />,
    },
  ]

  return (
    <div className="p-6">
      <Tabs items={items} />
    </div>
  )
}
