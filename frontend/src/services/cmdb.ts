import api from './api'

export interface CIInstance {
  id: number
  name: string
  type: string
  status: string
  attributes: Record<string, any>
  createdAt: string
  updatedAt: string
}

export interface CIType {
  id: number
  name: string
  icon: string
  description: string
}

export const cmdbService = {
  getTypes: () => api.get<CIType[]>('/ci/types'),
  getInstances: (params?: any) => api.get<CIInstance[]>('/ci/instances', { params }),
  getInstance: (id: number) => api.get<CIInstance>(`/ci/instances/${id}`),
  createInstance: (data: any) => api.post('/ci/instances', data),
  updateInstance: (id: number, data: any) => api.put(`/ci/instances/${id}`, data),
  deleteInstance: (id: number) => api.delete(`/ci/instances/${id}`),
  getRelations: (ciId: number) => api.get(`/ci/relations?ci_id=${ciId}`),
}
