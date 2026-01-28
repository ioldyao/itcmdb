import { Handle, Position, NodeProps } from '@xyflow/react'
import { SettingOutlined } from '@ant-design/icons'

export default function ServiceActivityNode({ data }: NodeProps) {
  const label: string = String(data?.label || '活动节点')
  const componentCode: string | null = data?.component && typeof data.component === 'object'
    ? String((data.component as any)?.code || 'custom')
    : null

  return (
    <div className="workflow-node workflow-node-activity">
      <Handle
        type="target"
        position={Position.Top}
        className="workflow-node-handle"
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <SettingOutlined />
        </div>
        <div className="workflow-node-label">
          {label}
        </div>
        {componentCode ? (
          <div className="workflow-node-component">
            {componentCode}
          </div>
        ) : null}
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="workflow-node-handle"
      />
    </div>
  )
}
