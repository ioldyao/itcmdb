import { useCallback, useEffect, useState } from 'react'
import {
  ReactFlow,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import StartEventNode from './nodes/StartEventNode'
import ServiceActivityNode from './nodes/ServiceActivityNode'
import ExclusiveGatewayNode from './nodes/ExclusiveGatewayNode'
import ParallelGatewayNode from './nodes/ParallelGatewayNode'
import ConvergeGatewayNode from './nodes/ConvergeGatewayNode'
import EndEventNode from './nodes/EndEventNode'

import {
  convertPipelineToFlow,
  convertFlowToPipeline,
  autoLayoutNodes
} from './utils/converter'

import './WorkflowEditor.css'

const initialNodes: Node[] = []
const initialEdges: Edge[] = []

export interface BambooPipeline {
  id: string
  start_event: any
  end_event: any
  activities: Record<string, any>
  gateways: Record<string, any>
  flows: Record<string, any>
  data: any
}

interface WorkflowEditorProps {
  pipeline?: BambooPipeline
  onChange?: (pipeline: BambooPipeline) => void
  onSave?: (pipeline: BambooPipeline) => void
  readonly?: boolean
}

// 节点类型定义
const NODE_TYPES = [
  { type: 'startEvent', label: '开始事件', color: '#52c41a' },
  { type: 'serviceActivity', label: '活动节点', color: '#1890ff' },
  { type: 'exclusiveGateway', label: '排他网关', color: '#faad14' },
  { type: 'parallelGateway', label: '并行网关', color: '#13c2c2' },
  { type: 'convergeGateway', label: '聚合网关', color: '#722ed1' },
  { type: 'endEvent', label: '结束事件', color: '#f5222d' },
]

export default function WorkflowEditor({
  pipeline,
  onChange,
  onSave,
  readonly = false
}: WorkflowEditorProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const [selectedNode, setSelectedNode] = useState<string | null>(null)

  // 转换 Pipeline 到 Flow
  useEffect(() => {
    if (pipeline) {
      const { nodes: flowNodes, edges: flowEdges } = convertPipelineToFlow(pipeline)
      setNodes(flowNodes as any)
      setEdges(flowEdges as any)
    }
  }, [pipeline])

  const onConnect = useCallback(
    (params: Connection) => {
      if (!readonly) {
        setEdges((eds) => addEdge(params, eds as any))
      }
    },
    [setEdges, readonly]
  )

  // 添加节点到画布
  const handleAddNode = useCallback((nodeType: string) => {
    const id = `${nodeType}-${Date.now()}`
    const newNode: Node = {
      id,
      type: nodeType,
      position: { x: Math.random() * 500, y: Math.random() * 500 },
      data: { label: NODE_TYPES.find(t => t.type === nodeType)?.label || nodeType },
    }

    setNodes((nds) => [...nds, newNode] as any)
  }, [setNodes])

  // 删除选中的节点
  const handleDeleteNode = useCallback(() => {
    if (selectedNode) {
      setNodes((nds) => nds.filter((n) => n.id !== selectedNode) as any)
      setEdges((eds) => eds.filter((e) => e.source !== selectedNode && e.target !== selectedNode) as any)
      setSelectedNode(null)
    }
  }, [selectedNode, setNodes, setEdges])

  // 删除选中的边
  const handleDeleteEdge = useCallback(() => {
    // ReactFlow 的选中边会存储在 edges 的 selected 属性中
    setEdges((eds) => eds.filter((e) => !e.selected) as any)
  }, [setEdges])

  // 清空画布
  const handleClear = useCallback(() => {
    if (confirm('确定要清空所有节点吗？')) {
      setNodes([] as any)
      setEdges([] as any)
    }
  }, [setNodes, setEdges])

  // 自动布局
  const handleAutoLayout = useCallback(() => {
    const layoutedNodes = autoLayoutNodes(nodes as any[], edges as any[])
    setNodes(layoutedNodes)
  }, [nodes, edges, setNodes])

  // 保存工作流
  const handleSave = useCallback(() => {
    const pipelineData = convertFlowToPipeline(nodes as Node[], edges as Edge[])
    onSave?.(pipelineData)
  }, [nodes, edges, onSave])

  // 监听变化
  useEffect(() => {
    if (!readonly && onChange) {
      const pipelineData = convertFlowToPipeline(nodes as Node[], edges as Edge[])
      onChange(pipelineData)
    }
  }, [nodes, edges, readonly, onChange])

  // 键盘删除事件
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.key === 'Delete' || e.key === 'Backspace') && !readonly) {
        handleDeleteNode()
        handleDeleteEdge()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleDeleteNode, handleDeleteEdge, readonly])

  return (
    <div className="workflow-editor">
      {/* 节点工具栏 */}
      {!readonly && (
        <div className="workflow-node-palette">
          <div className="palette-title">节点工具箱</div>
          <div className="palette-content">
            {NODE_TYPES.map((nodeType) => (
              <div
                key={nodeType.type}
                className="palette-node"
                onClick={() => handleAddNode(nodeType.type)}
                style={{ '--node-color': nodeType.color } as any}
              >
                <div className="palette-node-icon" />
                <div className="palette-node-label">{nodeType.label}</div>
              </div>
            ))}
          </div>
          <div className="palette-actions">
            <button className="palette-btn" onClick={handleClear}>
              清空画布
            </button>
            <button className="palette-btn" onClick={handleAutoLayout}>
              自动布局
            </button>
            <button className="palette-btn palette-btn-primary" onClick={handleSave}>
              保存工作流
            </button>
          </div>
        </div>
      )}

      {/* 主画布 */}
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={readonly ? undefined : onNodesChange}
        onEdgesChange={readonly ? undefined : onEdgesChange}
        onConnect={onConnect}
        onSelectionChange={({ nodes: selectedNodes }) => {
          if (selectedNodes && selectedNodes.length > 0) {
            setSelectedNode(selectedNodes[0].id)
          } else {
            setSelectedNode(null)
          }
        }}
        deleteKeyCode="Delete"
        nodeTypes={{
          startEvent: StartEventNode,
          serviceActivity: ServiceActivityNode,
          exclusiveGateway: ExclusiveGatewayNode,
          parallelGateway: ParallelGatewayNode,
          convergeGateway: ConvergeGatewayNode,
          endEvent: EndEventNode,
        }}
        fitView
        attributionPosition="bottom-left"
        nodesDraggable={!readonly}
        nodesConnectable={!readonly}
        elementsSelectable={!readonly}
        zoomOnScroll={!readonly}
        panOnScroll={!readonly}
      >
        <Controls />
        <MiniMap />
        <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
      </ReactFlow>

      {/* 提示信息 */}
      {!readonly && nodes.length === 0 && (
        <div className="workflow-editor-hint">
          <h3>开始设计工作流</h3>
          <p>从左侧工具栏点击节点类型添加到画布</p>
          <p>拖拽节点可以移动位置</p>
          <p>从节点的连接点拖拽到另一个节点创建连线</p>
          <p>选中节点或边后按 Delete 键删除</p>
        </div>
      )}
    </div>
  )
}
