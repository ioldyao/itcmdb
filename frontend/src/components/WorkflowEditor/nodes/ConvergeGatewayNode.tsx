import { Handle, Position, NodeProps } from '@xyflow/react'
import { MergeCellsOutlined } from '@ant-design/icons'

export default function ConvergeGatewayNode({ data }: NodeProps) {
  return (
    <div className="workflow-node workflow-node-gateway-converge">
      <Handle
        type="target"
        position={Position.Top}
        id="in-1"
        className="workflow-node-handle"
        style={{ left: '20%' }}
      />
      <Handle
        type="target"
        position={Position.Top}
        id="in-2"
        className="workflow-node-handle"
        style={{ left: '80%' }}
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <MergeCellsOutlined />
        </div>
        <div className="workflow-node-label">
          {(data.label as string) || '聚合网关'}
        </div>
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="workflow-node-handle"
      />
    </div>
  )
}
