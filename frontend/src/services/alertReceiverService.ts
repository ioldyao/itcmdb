import api from './api'

// 接收人类型
export interface AlertReceiver {
  id: number
  name: string
  type: 'wechat' | 'dingtalk' | 'feishu' | 'email' | 'sms'
  webhook_url: string
  at_mobiles: string[]
  at_user_ids: string[]
  secret: string
  config: Record<string, any>
  enabled: boolean
  created_at: string
  updated_at: string
}

// 接收组类型
export interface AlertReceiverGroup {
  id: number
  name: string
  description: string
  enabled: boolean
  created_at: string
  updated_at: string
  receivers?: AlertReceiver[]
}

// 接收人列表请求参数
export interface ReceiverListRequest {
  page?: number
  page_size?: number
  type?: string
  enabled?: boolean
}

// 接收人列表响应
export interface ReceiverListResponse {
  total: number
  receivers: AlertReceiver[]
}

// 接收组列表请求参数
export interface ReceiverGroupListRequest {
  page?: number
  page_size?: number
  enabled?: boolean
}

// 接收组列表响应
export interface ReceiverGroupListResponse {
  total: number
  groups: AlertReceiverGroup[]
}

// 创建接收人请求
export interface CreateReceiverRequest {
  name: string
  type: 'wechat' | 'dingtalk' | 'feishu' | 'email' | 'sms'
  webhook_url: string
  at_mobiles?: string[]
  at_user_ids?: string[]
  secret?: string
  config?: Record<string, any>
}

// 更新接收人请求
export interface UpdateReceiverRequest {
  name?: string
  webhook_url?: string
  at_mobiles?: string[]
  at_user_ids?: string[]
  secret?: string
  config?: Record<string, any>
  enabled?: boolean
}

// 创建接收组请求
export interface CreateReceiverGroupRequest {
  name: string
  description?: string
  receiver_ids: number[]
}

// 更新接收组请求
export interface UpdateReceiverGroupRequest {
  name?: string
  description?: string
  enabled?: boolean
  receiver_ids?: number[]
}

// API响应
export interface ApiResponse<T> {
  data?: T
  error?: string
  message?: string
}

export const alertReceiverService = {
  // 接收人相关
  getReceivers: async (params: ReceiverListRequest = {}): Promise<ReceiverListResponse> => {
    return api.get('/receivers', { params })
  },

  getReceiver: async (id: number): Promise<AlertReceiver> => {
    return api.get(`/receivers/${id}`)
  },

  createReceiver: async (data: CreateReceiverRequest): Promise<AlertReceiver> => {
    return api.post('/receivers', data)
  },

  updateReceiver: async (id: number, data: UpdateReceiverRequest): Promise<AlertReceiver> => {
    return api.put(`/receivers/${id}`, data)
  },

  deleteReceiver: async (id: number): Promise<{ message: string }> => {
    return api.delete(`/receivers/${id}`)
  },

  testReceiver: async (id: number): Promise<{ success: boolean; message: string }> => {
    return api.post(`/receivers/${id}/test`)
  },

  // 接收组相关
  getReceiverGroups: async (params: ReceiverGroupListRequest = {}): Promise<ReceiverGroupListResponse> => {
    return api.get('/receiver-groups', { params })
  },

  getReceiverGroup: async (id: number): Promise<AlertReceiverGroup> => {
    return api.get(`/receiver-groups/${id}`)
  },

  createReceiverGroup: async (data: CreateReceiverGroupRequest): Promise<AlertReceiverGroup> => {
    return api.post('/receiver-groups', data)
  },

  updateReceiverGroup: async (id: number, data: UpdateReceiverGroupRequest): Promise<AlertReceiverGroup> => {
    return api.put(`/receiver-groups/${id}`, data)
  },

  deleteReceiverGroup: async (id: number): Promise<{ message: string }> => {
    return api.delete(`/receiver-groups/${id}`)
  },
}
