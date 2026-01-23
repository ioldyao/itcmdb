import api from './api'

export interface Alert {
  id: number
  title: string
  description: string
  severity: string
  status: string
  affectedCIId: number
  triggeredAt: string
  acknowledgedAt?: string
  closedAt?: string
}

export const alertService = {
  getAlerts: (params?: any) => api.get<Alert[]>('/alerts', { params }),
  acknowledge: (id: number) => api.post(`/alerts/${id}/ack`),
  close: (id: number) => api.post(`/alerts/${id}/close`),
  getRules: () => api.get('/rules'),
  createRule: (data: any) => api.post('/rules', data),
}
