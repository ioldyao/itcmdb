import api from './api'

// ============================================
// 类型定义
// ============================================

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface AlertNotificationTemplate {
  id: number
  name: string
  description?: string
  template_type: 'dingtalk' | 'feishu' | 'wechat' | 'email'
  template_content: string
  is_default: boolean
  created_by?: number
  updated_by?: number
  created_at: string
  updated_at: string
}

export interface TemplateListRequest {
  page?: number
  page_size?: number
  template_type?: string
}

export interface TemplateListResponse {
  total: number
  templates: AlertNotificationTemplate[]
}

// ============================================
// API方法
// ============================================

export const alertTemplateService = {
  // 获取通知模板列表
  getTemplates: async (params: TemplateListRequest = {}): Promise<ApiResponse<TemplateListResponse>> => {
    return api.get('/alert-notification-templates', { params })
  },

  // 获取通知模板详情
  getTemplate: async (id: number): Promise<ApiResponse<AlertNotificationTemplate>> => {
    return api.get(`/alert-notification-templates/${id}`)
  },

  // 创建通知模板
  createTemplate: async (data: Partial<AlertNotificationTemplate>): Promise<ApiResponse<AlertNotificationTemplate>> => {
    return api.post('/alert-notification-templates', data)
  },

  // 更新通知模板
  updateTemplate: async (id: number, data: Partial<AlertNotificationTemplate>): Promise<ApiResponse<AlertNotificationTemplate>> => {
    return api.put(`/alert-notification-templates/${id}`, data)
  },

  // 删除通知模板
  deleteTemplate: async (id: number): Promise<ApiResponse<void>> => {
    return api.delete(`/alert-notification-templates/${id}`)
  },

  // 设置为默认模板
  setDefaultTemplate: async (id: number): Promise<ApiResponse<void>> => {
    return api.post(`/alert-notification-templates/${id}/set-default`)
  },

  // 预览模板
  previewTemplate: async (data: {
    template_content: string
    template_type: string
    sample_data?: Record<string, any>
  }): Promise<ApiResponse<{ preview: string }>> => {
    return api.post('/alert-notification-templates/preview', data)
  },
}
