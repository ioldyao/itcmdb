import api from './api'

export interface Ticket {
  id: string
  title: string
  description: string
  status: string
  priority: string
  assigneeId: number
  requesterId: number
  createdAt: string
  updatedAt: string
}

export const ticketService = {
  getTickets: (params?: any) => api.get<Ticket[]>('/tickets', { params }),
  getTicket: (id: string) => api.get<Ticket>(`/tickets/${id}`),
  createTicket: (data: any) => api.post('/tickets', data),
  updateStatus: (id: string, status: string) => api.put(`/tickets/${id}/status`, { status }),
  addComment: (id: string, content: string) => api.post(`/tickets/${id}/comments`, { content }),
  getWorkflows: () => api.get('/workflows'),
}
