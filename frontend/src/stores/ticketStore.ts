import { create } from 'zustand'

interface Ticket {
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

interface Workflow {
  id: number
  name: string
  states: string[]
  transitions: any[]
}

interface TicketFilters {
  status?: string
  priority?: string
  assignee?: number
}

interface TicketState {
  tickets: Ticket[]
  workflows: Workflow[]
  filters: TicketFilters
  loading: boolean
  setTickets: (tickets: Ticket[]) => void
  setWorkflows: (workflows: Workflow[]) => void
  setFilters: (filters: TicketFilters) => void
  setLoading: (loading: boolean) => void
}

export const useTicketStore = create<TicketState>((set) => ({
  tickets: [],
  workflows: [],
  filters: {},
  loading: false,
  setTickets: (tickets) => set({ tickets }),
  setWorkflows: (workflows) => set({ workflows }),
  setFilters: (filters) => set({ filters }),
  setLoading: (loading) => set({ loading }),
}))
