import { create } from 'zustand'

interface CIInstance {
  id: number
  name: string
  type: string
  status: string
  attributes: Record<string, any>
  createdAt: string
  updatedAt: string
}

interface CIType {
  id: number
  name: string
  icon: string
  description: string
}

interface CMDBState {
  instances: CIInstance[]
  types: CIType[]
  selectedInstance: CIInstance | null
  loading: boolean
  setInstances: (instances: CIInstance[]) => void
  setTypes: (types: CIType[]) => void
  setSelectedInstance: (instance: CIInstance | null) => void
  setLoading: (loading: boolean) => void
}

export const useCMDBStore = create<CMDBState>((set) => ({
  instances: [],
  types: [],
  selectedInstance: null,
  loading: false,
  setInstances: (instances) => set({ instances }),
  setTypes: (types) => set({ types }),
  setSelectedInstance: (selectedInstance) => set({ selectedInstance }),
  setLoading: (loading) => set({ loading }),
}))
