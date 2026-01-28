import api from './api'

export interface WebhookConfig {
  id: number
  name: string
  direction: 'inbound' | 'outbound'
  type: 'alertmanager' | 'prometheus' | 'victoriametrics' | 'workflow'
  webhook_url?: string
  webhook_token?: string
  enabled: boolean
  description?: string
  created_at: string
  updated_at: string
  workflow_id?: number
  workflow?: any
}

export interface CreateWebhookRequest {
  name: string
  direction: 'inbound' | 'outbound'
  type: 'alertmanager' | 'prometheus' | 'victoriametrics' | 'workflow'
  webhook_url?: string
  workflow_id?: number
  workflow?: {
    pipeline: string
    enabled: boolean
  }
  enabled: boolean
  description?: string
}

export const webhookService = {
  getWebhooks: async (params: { page?: number; page_size?: number } = {}) => {
    return api.get('/webhooks', { params })
  },

  getWebhook: async (id: number) => {
    return api.get(`/webhooks/${id}`)
  },

  createWebhook: async (data: CreateWebhookRequest) => {
    return api.post('/webhooks', data)
  },

  updateWebhook: async (id: number, data: Partial<CreateWebhookRequest>) => {
    return api.put(`/webhooks/${id}`, data)
  },

  deleteWebhook: async (id: number) => {
    return api.delete(`/webhooks/${id}`)
  },
}
