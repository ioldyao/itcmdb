import { Handle, Position, NodeProps } from '@xyflow/react'
import { StopOutlined } from '@ant-design/icons'

export default function EndEventNode({ data }: NodeProps) {
  return (
    <div className="workflow-node workflow-node-end">
      <Handle
        type="target"
        position={Position.Top}
        className="workflow-node-handle"
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <StopOutlined />
        </div>
        <div className="workflow-node-label">
          {(data.label as string) || '结束'}
        </div>
      </div>

      <div className="workflow-node-type">结束事件</div>
    </div>
  )
}
