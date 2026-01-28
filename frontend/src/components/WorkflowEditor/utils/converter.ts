import { Node, Edge } from '@xyflow/react'
import { BambooPipeline } from '../index'

/**
 * 生成 UUID
 */
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

/**
 * 将 Bamboo-Engine Pipeline 转换为 React Flow 格式
 */
export function convertPipelineToFlow(pipeline: BambooPipeline): {
  nodes: Node[]
  edges: Edge[]
} {
  const nodes: Node[] = []
  const edges: Edge[] = []

  // 转换开始事件
  if (pipeline.start_event) {
    nodes.push({
      id: pipeline.start_event.id,
      type: 'startEvent',
      position: { x: 0, y: 0 },
      data: {
        label: pipeline.start_event.name || '开始',
      },
    })
  }

  // 转换活动节点
  Object.values(pipeline.activities || {}).forEach((activity: any) => {
    nodes.push({
      id: activity.id,
      type: 'serviceActivity',
      position: { x: 0, y: 0 },
      data: {
        label: activity.name,
        component: activity.component,
        error_ignorable: activity.error_ignore,
        retryable: activity.retryable,
        skippable: activity.skippable,
        timeout: activity.timeout,
      },
    })
  })

  // 转换网关节点
  Object.values(pipeline.gateways || {}).forEach((gateway: any) => {
    const gatewayType = gateway.type.toLowerCase()
    let nodeType = 'exclusiveGateway'

    if (gatewayType.includes('parallel')) {
      nodeType = 'parallelGateway'
    } else if (gatewayType.includes('converge')) {
      nodeType = 'convergeGateway'
    }

    nodes.push({
      id: gateway.id,
      type: nodeType,
      position: { x: 0, y: 0 },
      data: {
        label: gateway.name,
        conditions: gateway.conditions,
      },
    })
  })

  // 转换结束事件
  if (pipeline.end_event) {
    nodes.push({
      id: pipeline.end_event.id,
      type: 'endEvent',
      position: { x: 0, y: 0 },
      data: {
        label: pipeline.end_event.name || '结束',
      },
    })
  }

  // 转换连线
  Object.values(pipeline.flows || {}).forEach((flow: any) => {
    edges.push({
      id: flow.id,
      source: flow.source,
      target: flow.target,
      animated: true,
      type: 'smoothstep',
      label: flow.condition ? flow.condition : undefined,
    })
  })

  // 自动布局
  const layoutedNodes = autoLayoutNodes(nodes, edges)

  return {
    nodes: layoutedNodes,
    edges,
  }
}

/**
 * 将 React Flow 格式转换为 Bamboo-Engine Pipeline
 */
export function convertFlowToPipeline(
  nodes: Node[],
  edges: Edge[]
): BambooPipeline {
  const pipeline: BambooPipeline = {
    id: generateUUID(),
    start_event: {} as any,
    end_event: {} as any,
    activities: {},
    gateways: {},
    flows: {},
    data: {
      inputs: {},
      outputs: [],
    },
  }

  // 先创建所有 flow
  edges.forEach((edge) => {
    const flowId = edge.id || generateUUID()
    pipeline.flows[flowId] = {
      id: flowId,
      source: edge.source,
      target: edge.target,
      is_default: false,
    }
  })

  // 遍历节点
  nodes.forEach((node) => {
    if (node.type === 'startEvent') {
      const outgoingFlowId = getOutgoingFlowId(edges, node.id)
      pipeline.start_event = {
        id: node.id,
        type: 'EmptyStartEvent',
        name: node.data.label || '开始',
        incoming: '',
        outgoing: outgoingFlowId,
      }
    } else if (node.type === 'endEvent') {
      const incomingFlowIds = getIncomingFlowIds(edges, node.id)
      pipeline.end_event = {
        id: node.id,
        type: 'EmptyEndEvent',
        name: node.data.label || '结束',
        incoming: incomingFlowIds,
        outgoing: '',
      }
    } else if (node.type === 'serviceActivity') {
      const incomingFlowIds = getIncomingFlowIds(edges, node.id)
      const outgoingFlowId = getOutgoingFlowId(edges, node.id)

      pipeline.activities[node.id] = {
        id: node.id,
        type: 'ServiceActivity',
        name: node.data.label || '活动节点',
        component: node.data.component || {
          code: 'custom_component',
          inputs: node.data.inputs || {},
        },
        incoming: incomingFlowIds,
        outgoing: outgoingFlowId,
        error_ignore: node.data.error_ignorable || false,
        optional: false,
        retryable: node.data.retryable !== undefined ? node.data.retryable : true,
        skippable: node.data.skippable !== undefined ? node.data.skippable : true,
        timeout: node.data.timeout || null,
      }
    } else if (node.type && node.type.includes('Gateway')) {
      const gatewayType =
        node.type === 'parallelGateway'
          ? 'ParallelGateway'
          : node.type === 'convergeGateway'
          ? 'ConvergeGateway'
          : 'ExclusiveGateway'

      const outgoingFlowIds = getOutgoingFlowIds(edges, node.id)

      pipeline.gateways[node.id] = {
        id: node.id,
        type: gatewayType,
        name: node.data.label || gatewayType,
        outgoing: outgoingFlowIds,
        conditions: node.data.conditions || {},
      }
    }
  })

  return pipeline
}

/**
 * 自动布局节点（简单的分层布局）
 */
export function autoLayoutNodes(
  nodes: Node[],
  edges: Edge[]
): Node[] {
  // 构建节点层级
  const levels = new Map<string, number>()
  const visited = new Set<string>()

  // 找到开始节点
  const startNode = nodes.find((n) => n.type === 'startEvent')
  if (!startNode) {
    return nodes.map((n) => ({ ...n, position: { x: Math.random() * 500, y: Math.random() * 500 } }))
  }

  // BFS 计算节点层级
  const queue: { id: string; level: number }[] = [
    { id: startNode.id, level: 0 },
  ]

  while (queue.length > 0) {
    const { id, level } = queue.shift()!

    if (visited.has(id)) {
      continue
    }

    visited.add(id)
    levels.set(id, level)

    // 找到所有从该节点出发的边
    const outgoingEdges = edges.filter((e) => e.source === id)
    outgoingEdges.forEach((edge) => {
      if (!visited.has(edge.target)) {
        queue.push({ id: edge.target, level: level + 1 })
      }
    })
  }

  // 根据层级计算位置
  const levelGroups = new Map<number, string[]>()
  levels.forEach((level, id) => {
    if (!levelGroups.has(level)) {
      levelGroups.set(level, [])
    }
    levelGroups.get(level)!.push(id)
  })

  // 计算节点位置
  const positionedNodes = nodes.map((node) => {
    const level = levels.get(node.id) ?? 0
    const levelNodes = levelGroups.get(level) || []
    const indexInLevel = levelNodes.indexOf(node.id)

    const x = level * 300
    const y = (indexInLevel - (levelNodes.length - 1) / 2) * 120 + 250

    return {
      ...node,
      position: { x, y },
    }
  })

  return positionedNodes
}

/**
 * 获取节点的所有输入 flow ID
 */
function getIncomingFlowIds(edges: Edge[], nodeId: string): string[] {
  return edges
    .filter((e) => e.target === nodeId)
    .map((e) => e.id)
}

/**
 * 获取节点的单个输出 flow ID
 */
function getOutgoingFlowId(edges: Edge[], nodeId: string): string {
  const flow = edges.find((e) => e.source === nodeId)
  return flow?.id || ''
}

/**
 * 获取节点的所有输出 flow ID
 */
function getOutgoingFlowIds(edges: Edge[], nodeId: string): string[] {
  return edges
    .filter((e) => e.source === nodeId)
    .map((e) => e.id)
}
