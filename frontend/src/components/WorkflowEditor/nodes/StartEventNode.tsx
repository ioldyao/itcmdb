import { Handle, Position, NodeProps } from '@xyflow/react'
import { PlayCircleOutlined } from '@ant-design/icons'

export default function StartEventNode({ data }: NodeProps) {
  return (
    <div className="workflow-node workflow-node-start">
      <Handle
        type="source"
        position={Position.Bottom}
        className="workflow-node-handle"
      />

      <div className="workflow-node-content">
        <div className="workflow-node-icon">
          <PlayCircleOutlined />
        </div>
        <div className="workflow-node-label">
          {(data.label as string) || '开始'}
        </div>
      </div>

      <div className="workflow-node-type">开始事件</div>
    </div>
  )
}
