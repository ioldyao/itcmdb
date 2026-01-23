export interface User {
  id: number
  username: string
  email: string
  fullName: string
}

export interface Role {
  id: number
  name: string
  description: string
}

export interface Permission {
  id: number
  resource: string
  action: string
  description: string
}

export interface CIInstance {
  id: number
  name: string
  type: string
  status: string
  attributes: Record<string, any>
  createdAt: string
  updatedAt: string
}

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
