import { Handle, Position, NodeProps } from '@xyflow/react'
import { SplitCellsOutlined } from '@ant-design/icons'

export default function ParallelGatewayNode({ data }: NodeProps) {
  return (
    <div className="workflow-node workflow-node-gateway-parallel">
      <Handle
        type="target"
        position={Position.Top}
        className="workflow-node-handle"
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <SplitCellsOutlined />
        </div>
        <div className="workflow-node-label">
          {(data.label as string) || '并行网关'}
        </div>
      </div>

      <div className="workflow-node-gateway-outputs">
        <Handle
          type="source"
          position={Position.Bottom}
          id="out-1"
          className="workflow-node-handle"
          style={{ left: '20%' }}
        />
        <Handle
          type="source"
          position={Position.Bottom}
          id="out-2"
          className="workflow-node-handle"
          style={{ left: '80%' }}
        />
      </div>
    </div>
  )
}
