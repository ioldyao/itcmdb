import api from './api'

// ============================================
// 类型定义
// ============================================

export interface AlertRoutingRule {
  id: number
  name: string
  description?: string
  matchers: Record<string, string>
  match_type: 'match' | 'match_re'
  receiver_group_id?: number
  continue: boolean
  priority: number
  enabled: boolean
  created_by?: number
  updated_by?: number
  created_at: string
  updated_at: string
}

export interface RoutingRuleListRequest {
  page?: number
  page_size?: number
  enabled?: boolean
}

export interface RoutingRuleListResponse {
  total: number
  rules: AlertRoutingRule[]
}

// ============================================
// API方法
// ============================================

export const alertRoutingService = {
  // 获取路由规则列表
  getRoutingRules: async (params: RoutingRuleListRequest = {}): Promise<{ data: RoutingRuleListResponse }> => {
    return api.get('/alert-routing-rules', { params })
  },

  // 获取路由规则详情
  getRoutingRule: async (id: number): Promise<{ data: AlertRoutingRule }> => {
    return api.get(`/alert-routing-rules/${id}`)
  },

  // 创建路由规则
  createRoutingRule: async (data: Partial<AlertRoutingRule>): Promise<{ data: AlertRoutingRule }> => {
    return api.post('/alert-routing-rules', data)
  },

  // 更新路由规则
  updateRoutingRule: async (id: number, data: Partial<AlertRoutingRule>): Promise<{ data: AlertRoutingRule }> => {
    return api.put(`/alert-routing-rules/${id}`, data)
  },

  // 删除路由规则
  deleteRoutingRule: async (id: number): Promise<{ data: void }> => {
    return api.delete(`/alert-routing-rules/${id}`)
  },

  // 启用路由规则
  enableRoutingRule: async (id: number): Promise<{ data: void }> => {
    return api.post(`/alert-routing-rules/${id}/enable`)
  },

  // 禁用路由规则
  disableRoutingRule: async (id: number): Promise<{ data: void }> => {
    return api.post(`/alert-routing-rules/${id}/disable`)
  },
}
