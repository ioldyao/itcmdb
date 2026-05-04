import api from './api'

// ============================================
// 类型定义
// ============================================

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface AlertInstance {
  id: number
  alert_id: string
  rule_id?: number
  title: string
  description?: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  status: 'firing' | 'acknowledged' | 'resolved' | 'closed'
  category?: string
  tags?: Record<string, any>
  object_type?: string
  target_info?: Record<string, any>
  affected_ci_id?: number
  trigger_conditions?: Record<string, any>
  metrics?: {
    current_value: number
    threshold_value: number
    deviation?: number
  }
  fingerprint: string
  first_triggered: string
  last_triggered: string
  recovered_at?: string
  closed_at?: string
  count: number
  handler?: number
  handling_status?: string
  handling_notes?: string
  acknowledged_at?: string
  notification_sent: boolean
  notification_channels?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface AlertRule {
  id: number
  name: string
  description?: string
  metric_query: string
  threshold_operator: '>' | '<' | '>=' | '<=' | '==' | '!='
  threshold_value: number
  duration: number
  severity: 'critical' | 'high' | 'medium' | 'low'
  enabled: boolean
  ci_type_id?: number
  notification_channels?: Record<string, any>
  silenced_until?: string
  created_by?: number
  updated_by?: number
  created_at: string
  updated_at: string
}

export interface AlertHistory {
  id: number
  alert_id: number
  event_type: 'triggered' | 'updated' | 'acknowledged' | 'resolved' | 'closed'
  old_status?: string
  new_status?: string
  operated_by?: number
  operated_at: string
  message?: string
  details?: Record<string, any>
}

export interface AlertListRequest {
  page?: number
  page_size?: number
  status?: string[]
  severity?: string[]
  category?: string
  search_keyword?: string
  start_time?: string
  end_time?: string
  sort_field?: string
  sort_order?: 'asc' | 'desc'
}

export interface AlertListResponse {
  total: number
  alerts: AlertInstance[]
}

// ============================================
// API方法
// ============================================

export const alertService = {
  // 获取告警列表
  getAlerts: async (params: AlertListRequest = {}): Promise<ApiResponse<AlertListResponse>> => {
    return api.get('/alerts', { params })
  },

  // 获取告警详情
  getAlertById: async (id: number): Promise<ApiResponse<AlertInstance>> => {
    return api.get(`/alerts/${id}`)
  },

  // 确认告警
  acknowledgeAlert: async (id: number, data: { handler: number; notes?: string }): Promise<ApiResponse<void>> => {
    return api.post(`/alerts/${id}/ack`, data)
  },

  // 关闭告警
  closeAlert: async (id: number, data: { handler: number; notes?: string }): Promise<ApiResponse<void>> => {
    return api.post(`/alerts/${id}/close`, data)
  },

  // 批量确认告警
  batchAcknowledgeAlerts: async (data: {
    alert_ids: number[]
    handler: number
    notes?: string
  }): Promise<ApiResponse<{ affected_rows: number }>> => {
    return api.post('/alerts/batch/ack', data)
  },

  // 批量关闭告警
  batchCloseAlerts: async (data: {
    alert_ids: number[]
    handler: number
    notes?: string
  }): Promise<ApiResponse<{ affected_rows: number }>> => {
    return api.post('/alerts/batch/close', data)
  },

  // 获取告警历史
  getAlertHistory: async (id: number): Promise<ApiResponse<AlertHistory[]>> => {
    return api.get(`/alerts/${id}/history`)
  },

  // 获取告警统计
  getAlertStatistics: async (): Promise<ApiResponse<{
    stats: {
      total: number
      firing: number
      acknowledged: number
      resolved: number
      closed: number
    }
    severity_stats: Array<{ severity: string; count: number }>
  }>> => {
    return api.get('/alerts/statistics')
  },

  // 获取告警分析数据
  getAlertAnalytics: async (params: {
    start_time: string
    end_time: string
    group_by?: string[]
  }): Promise<ApiResponse<any>> => {
    return api.get('/alerts/analytics', { params })
  },

  // ============================================
  // 告警静默管理
  // ============================================

  getSilences: async (params?: { active?: string }): Promise<ApiResponse<any[]>> => {
    return api.get('/silences', { params })
  },

  createSilence: async (data: {
    name: string
    comment?: string
    matchers: Record<string, any>
    starts_at: string
    ends_at: string
  }): Promise<ApiResponse<any>> => {
    return api.post('/silences', data)
  },

  deleteSilence: async (id: number): Promise<ApiResponse<void>> => {
    return api.delete(`/silences/${id}`)
  },

  // ============================================
  // 空间管理
  // ============================================

  getSpaces: async (): Promise<ApiResponse<any[]>> => {
    return api.get('/spaces')
  },

  createSpace: async (data: { name: string; description?: string; role_ids?: number[] }): Promise<ApiResponse<any>> => {
    return api.post('/spaces', data)
  },

  updateSpace: async (id: number, data: { name?: string; description?: string; role_ids?: number[] }): Promise<ApiResponse<any>> => {
    return api.put(`/spaces/${id}`, data)
  },

  deleteSpace: async (id: number): Promise<ApiResponse<void>> => {
    return api.delete(`/spaces/${id}`)
  },

  // ============================================
  // 角色（从 auth-service 获取）
  // ============================================

  getRoles: async (): Promise<ApiResponse<any[]>> => {
    return api.get('/auth/roles')
  },

  // ============================================
  // 空间路由规则
  // ============================================

  getSpaceRoutes: async (): Promise<ApiResponse<any[]>> => {
    return api.get('/space-routes')
  },

  createSpaceRoute: async (data: { field_name: string; field_value: string; space_id: number; priority?: number }): Promise<ApiResponse<any>> => {
    return api.post('/space-routes', data)
  },

  updateSpaceRoute: async (id: number, data: { field_name?: string; field_value?: string; space_id?: number; priority?: number; enabled?: boolean }): Promise<ApiResponse<any>> => {
    return api.put(`/space-routes/${id}`, data)
  },

  deleteSpaceRoute: async (id: number): Promise<ApiResponse<void>> => {
    return api.delete(`/space-routes/${id}`)
  },

  // ============================================
  // 告警规则管理
  // ============================================

  // 获取规则列表
  getRules: async (params?: {
    page?: number
    page_size?: number
    severity?: string
    enabled?: string
  }): Promise<ApiResponse<{ total: number; rules: AlertRule[] }>> => {
    return api.get('/rules', { params })
  },

  // 获取规则详情
  getRuleById: async (id: number): Promise<ApiResponse<AlertRule>> => {
    return api.get(`/rules/${id}`)
  },

  // 创建规则
  createRule: async (data: Partial<AlertRule>): Promise<ApiResponse<AlertRule>> => {
    return api.post('/rules', data)
  },

  // 更新规则
  updateRule: async (id: number, data: Partial<AlertRule>): Promise<ApiResponse<AlertRule>> => {
    return api.put(`/rules/${id}`, data)
  },

  // 删除规则
  deleteRule: async (id: number): Promise<ApiResponse<void>> => {
    return api.delete(`/rules/${id}`)
  },

  // 启用规则
  enableRule: async (id: number): Promise<ApiResponse<void>> => {
    return api.post(`/rules/${id}/enable`)
  },

  // 禁用规则
  disableRule: async (id: number): Promise<ApiResponse<void>> => {
    return api.post(`/rules/${id}/disable`)
  },

  // 测试规则
  testRule: async (data: {
    metric_query: string
    threshold_operator: string
    threshold_value: number
  }): Promise<ApiResponse<any>> => {
    return api.post('/rules/test', data)
  },
}
