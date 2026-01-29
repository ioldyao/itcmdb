import api from './api'

// 类型定义
export interface InboundWebhook {
  id: number
  name: string
  webhook_url: string
  source_type: 'alertmanager' | 'prometheus' | 'victoriametrics' | 'custom'
  enabled: boolean
  description?: string
  last_received?: string
  created_at: string
  updated_at: string
}

export interface OutboundWebhook {
  id: number
  name: string
  target_type: 'alertmanager' | 'receiver'
  receiver_id?: number
  endpoint_url?: string
  enabled: boolean
  description?: string
  last_sent?: string
  created_at: string
  updated_at: string
  receiver?: {
    id: number
    name: string
    type: string
    webhook_url: string
  }
}

export interface CreateInboundWebhookRequest {
  name: string
  source_type: 'alertmanager' | 'prometheus' | 'victoriametrics' | 'custom'
  description?: string
}

export interface UpdateInboundWebhookRequest {
  name?: string
  enabled?: boolean
  description?: string
}

export interface CreateOutboundWebhookRequest {
  name: string
  target_type: 'alertmanager' | 'receiver'
  receiver_id?: number
  endpoint_url?: string
  description?: string
}

export interface UpdateOutboundWebhookRequest {
  name?: string
  receiver_id?: number
  endpoint_url?: string
  enabled?: boolean
  description?: string
}

export interface InboundWebhookListResponse {
  total: number
  webhooks: InboundWebhook[]
}

export interface OutboundWebhookListResponse {
  total: number
  webhooks: OutboundWebhook[]
}

const API_BASE = '/api/v1'

// Inbound Webhook Service
export const inboundWebhookService = {
  // 获取接收Webhook列表
  getWebhooks: async (params?: { page?: number; page_size?: number; source_type?: string; enabled?: boolean }) => {
    const response = await api.get<InboundWebhookListResponse>(`/webhooks/inbound`, { params })
    return response
  },

  // 获取接收Webhook详情
  getWebhook: async (id: number) => {
    const response = await api.get<InboundWebhook>(`/`webhooks/inbound/${id}`)
    return response
  },

  // 创建接收Webhook
  createWebhook: async (data: CreateInboundWebhookRequest) => {
    const response = await api.post<InboundWebhook>(`/`webhooks/inbound`, data)
    return response
  },

  // 更新接收Webhook
  updateWebhook: async (id: number, data: UpdateInboundWebhookRequest) => {
    const response = await api.put<InboundWebhook>(`/`webhooks/inbound/${id}`, data)
    return response
  },

  // 删除接收Webhook
  deleteWebhook: async (id: number) => {
    const response = await api.delete<{ message: string }>(`/`webhooks/inbound/${id}`)
    return response
  },
}

// Outbound Webhook Service
export const outboundWebhookService = {
  // 获取推送Webhook列表
  getWebhooks: async (params?: { page?: number; page_size?: number; target_type?: string; enabled?: boolean }) => {
    const response = await api.get<OutboundWebhookListResponse>(`/`webhooks/outbound`, { params })
    return response
  },

  // 获取推送Webhook详情
  getWebhook: async (id: number) => {
    const response = await api.get<OutboundWebhook>(`/`webhooks/outbound/${id}`)
    return response
  },

  // 创建推送Webhook
  createWebhook: async (data: CreateOutboundWebhookRequest) => {
    const response = await api.post<OutboundWebhook>(`/`webhooks/outbound`, data)
    return response
  },

  // 更新推送Webhook
  updateWebhook: async (id: number, data: UpdateOutboundWebhookRequest) => {
    const response = await api.put<OutboundWebhook>(`/`webhooks/outbound/${id}`, data)
    return response
  },

  // 删除推送Webhook
  deleteWebhook: async (id: number) => {
    const response = await api.delete<{ message: string }>(`/`webhooks/outbound/${id}`)
    return response
  },

  // 测试推送Webhook
  testWebhook: async (id: number) => {
    const response = await api.post<{ message: string }>(`/`webhooks/outbound/${id}/test`)
    return response
  },
}
