import axios from 'axios'

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
    const response = await axios.get<InboundWebhookListResponse>(`${API_BASE}/webhooks/inbound`, { params })
    return response.data
  },

  // 获取接收Webhook详情
  getWebhook: async (id: number) => {
    const response = await axios.get<InboundWebhook>(`${API_BASE}/webhooks/inbound/${id}`)
    return response.data
  },

  // 创建接收Webhook
  createWebhook: async (data: CreateInboundWebhookRequest) => {
    const response = await axios.post<InboundWebhook>(`${API_BASE}/webhooks/inbound`, data)
    return response.data
  },

  // 更新接收Webhook
  updateWebhook: async (id: number, data: UpdateInboundWebhookRequest) => {
    const response = await axios.put<InboundWebhook>(`${API_BASE}/webhooks/inbound/${id}`, data)
    return response.data
  },

  // 删除接收Webhook
  deleteWebhook: async (id: number) => {
    const response = await axios.delete<{ message: string }>(`${API_BASE}/webhooks/inbound/${id}`)
    return response.data
  },
}

// Outbound Webhook Service
export const outboundWebhookService = {
  // 获取推送Webhook列表
  getWebhooks: async (params?: { page?: number; page_size?: number; target_type?: string; enabled?: boolean }) => {
    const response = await axios.get<OutboundWebhookListResponse>(`${API_BASE}/webhooks/outbound`, { params })
    return response.data
  },

  // 获取推送Webhook详情
  getWebhook: async (id: number) => {
    const response = await axios.get<OutboundWebhook>(`${API_BASE}/webhooks/outbound/${id}`)
    return response.data
  },

  // 创建推送Webhook
  createWebhook: async (data: CreateOutboundWebhookRequest) => {
    const response = await axios.post<OutboundWebhook>(`${API_BASE}/webhooks/outbound`, data)
    return response.data
  },

  // 更新推送Webhook
  updateWebhook: async (id: number, data: UpdateOutboundWebhookRequest) => {
    const response = await axios.put<OutboundWebhook>(`${API_BASE}/webhooks/outbound/${id}`, data)
    return response.data
  },

  // 删除推送Webhook
  deleteWebhook: async (id: number) => {
    const response = await axios.delete<{ message: string }>(`${API_BASE}/webhooks/outbound/${id}`)
    return response.data
  },

  // 测试推送Webhook
  testWebhook: async (id: number) => {
    const response = await axios.post<{ message: string }>(`${API_BASE}/webhooks/outbound/${id}/test`)
    return response.data
  },
}
