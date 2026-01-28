import api from './api'

export interface Workflow {
  id: number
  name: string
  description: string
  direction: 'inbound' | 'outbound'
  type: 'alertmanager' | 'prometheus' | 'victoriametrics' | 'workflow'
  pipeline: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateWorkflowRequest {
  name: string
  description?: string
  direction: 'inbound' | 'outbound'
  type: string
  pipeline: string
  enabled?: boolean
}

export interface UpdateWorkflowRequest {
  name?: string
  description?: string
  direction?: 'inbound' | 'outbound'
  type?: string
  pipeline?: string
  enabled?: boolean
}

export const workflowService = {
  getWorkflows: async (params: { page?: number; page_size?: number } = {}) => {
    return api.get('/workflows', { params })
  },

  getWorkflow: async (id: number) => {
    return api.get(`/workflows/${id}`)
  },

  createWorkflow: async (data: CreateWorkflowRequest) => {
    return api.post('/workflows', data)
  },

  updateWorkflow: async (id: number, data: UpdateWorkflowRequest) => {
    return api.put(`/workflows/${id}`, data)
  },

  deleteWorkflow: async (id: number) => {
    return api.delete(`/workflows/${id}`)
  },

  executeWorkflow: async (data: { pipeline: any; data?: any }) => {
    return api.post('/workflows/execute', data)
  },
}
