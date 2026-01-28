import { Handle, Position, NodeProps } from '@xyflow/react'
import { BranchesOutlined } from '@ant-design/icons'

export default function ExclusiveGatewayNode({ data }: NodeProps) {
  return (
    <div className="workflow-node workflow-node-gateway-exclusive">
      <Handle
        type="target"
        position={Position.Top}
        className="workflow-node-handle"
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <BranchesOutlined />
        </div>
        <div className="workflow-node-label">
          {(data.label as string) || '排他网关'}
        </div>
      </div>

      <div className="workflow-node-gateway-outputs">
        <Handle
          type="source"
          position={Position.Bottom}
          id="out-1"
          className="workflow-node-handle"
        />
      </div>
    </div>
  )
}
