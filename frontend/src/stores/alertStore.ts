import { create } from 'zustand'

interface Alert {
  id: number
  title: string
  description: string
  severity: string
  status: string
  affectedCIId: number
  triggeredAt: string
  acknowledgedAt?: string
  acknowledgedBy?: number
  closedAt?: string
}

interface AlertStats {
  total: number
  critical: number
  high: number
  medium: number
  low: number
}

interface AlertState {
  alerts: Alert[]
  stats: AlertStats
  loading: boolean
  setAlerts: (alerts: Alert[]) => void
  setStats: (stats: AlertStats) => void
  setLoading: (loading: boolean) => void
}

export const useAlertStore = create<AlertState>((set) => ({
  alerts: [],
  stats: {
    total: 0,
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
  },
  loading: false,
  setAlerts: (alerts) => set({ alerts }),
  setStats: (stats) => set({ stats }),
  setLoading: (loading) => set({ loading }),
}))
