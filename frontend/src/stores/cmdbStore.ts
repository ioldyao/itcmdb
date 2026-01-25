import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { message } from 'antd'
import { useAuthStore } from './authStore'

// CI 类型
export interface CIType {
  id: number
  name: string
  display_name: string
  icon: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
  attributes?: CIAttribute[]
}

// CI 属性
export interface CIAttribute {
  id: number
  ci_type_id: number
  name: string
  display_name: string
  type: string
  options: Record<string, any>
  is_required: boolean
  is_unique: boolean
  default_value: string
  sort_order: number
}

// CI 实例
export interface CIInstance {
  id: number
  ci_type_id: number
  name: string
  status: 'active' | 'inactive' | 'maintenance' | 'decommissioned'
  attributes: Record<string, any>
  tags: Record<string, any>
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
  ci_type?: CIType
}

// CI 关系
export interface CIRelation {
  id: number
  parent_id: number
  child_id: number
  relation_type: string
  description: string
  created_by: number
  created_at: string
  updated_at: string
  parent?: CIInstance
  child?: CIInstance
}

// CI 变更历史
export interface CIHistory {
  id: number
  ci_id: number
  changed_by: number
  action: string // create, update, delete
  field_name: string
  old_value: string
  new_value: string
  changed_at: string
}

// 分页响应
export interface PaginationResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// API 响应
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

interface CMDBState {
  ciTypes: CIType[]
  instances: CIInstance[]
  selectedInstance: CIInstance | null
  total: number
  page: number
  pageSize: number
  filters: Record<string, any>
  currentCITypeID: number  // 当前过滤的CI类型ID
  loading: boolean

  // Actions
  fetchCITypes: () => Promise<void>
  fetchInstances: (ciTypeID?: number, page?: number, pageSize?: number) => Promise<void>
  fetchInstance: (id: number) => Promise<CIInstance>
  createInstance: (data: CreateCIInstanceRequest) => Promise<CIInstance>
  updateInstance: (id: number, data: UpdateCIInstanceRequest) => Promise<void>
  deleteInstance: (id: number) => Promise<void>
  fetchHistory: (ciId: number, limit?: number) => Promise<CIHistory[]>
  fetchRelations: (ciId: number) => Promise<CIRelation[]>
  createRelation: (data: CreateCIRelationRequest) => Promise<CIRelation>
  setFilters: (filters: Record<string, any>) => void
  resetFilters: () => void
}

interface CreateCIInstanceRequest {
  ci_type_id: number
  name: string
  status?: string
  attributes?: Record<string, any>
  tags?: Record<string, any>
}

interface UpdateCIInstanceRequest {
  name?: string
  status?: string
  attributes?: Record<string, any>
  tags?: Record<string, any>
}

interface CreateCIRelationRequest {
  parent_id: number
  child_id: number
  relation_type: string
  description?: string
}

// Helper function to get auth token
const getAuthToken = () => {
  const authState = useAuthStore.getState()
  const token = authState?.token
  console.log('[CMDB] Getting token:', token ? `${token.substring(0, 20)}...` : 'undefined')
  console.log('[CMDB] Auth state:', { isAuthenticated: authState?.isAuthenticated, user: authState?.user })
  return token
}

export const useCMDBStore = create<CMDBState>()(
  persist(
    (set, get) => ({
      ciTypes: [],
      instances: [],
      selectedInstance: null,
      total: 0,
      page: 1,
      pageSize: 20,
      filters: {},
      currentCITypeID: 0,
      loading: false,

      fetchCITypes: async () => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/ci/types', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result: ApiResponse<CIType[]> = await response.json()
          if (result.code === 0) {
            set({ ciTypes: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch CI types:', error)
        } finally {
          set({ loading: false })
        }
      },

      fetchInstances: async (ciTypeID = 0, page = 1, pageSize = 20) => {
        set({ loading: true, currentCITypeID: ciTypeID })
        try {
          const token = getAuthToken()
          if (!token) {
            console.error('No auth token found')
            message.error('请先登录')
            return
          }

          const params = new URLSearchParams()
          if (ciTypeID > 0) params.append('ci_type_id', ciTypeID.toString())
          params.append('page', page.toString())
          params.append('page_size', pageSize.toString())

          const filters = get().filters
          if (filters.status) params.append('status', filters.status)
          if (filters.name) params.append('name', filters.name)

          const headers: Record<string, string> = {
            'Authorization': `Bearer ${token}`,
          }

          console.log('[CMDB] Fetching with headers:', {
            'Authorization': headers.Authorization.substring(0, 50) + '...'
          })

          const response = await fetch(`/api/v1/ci/instances?${params}`, {
            headers,
          })

          console.log('[CMDB] Response status:', response.status)
          console.log('[CMDB] Response headers:', Object.fromEntries(response.headers.entries()))

          if (response.status === 401) {
            message.error('登录已过期，请重新登录')
            // 清除认证状态
            useAuthStore.getState().logout()
            return
          }

          const result: ApiResponse<PaginationResponse<CIInstance>> = await response.json()
          if (result.code === 0) {
            set({
              instances: result.data.items,
              total: result.data.total,
              page: result.data.page,
              pageSize: result.data.page_size,
            })
          } else {
            message.error(result.message || '加载数据失败')
          }
        } catch (error) {
          console.error('Failed to fetch CI instances:', error)
          message.error('网络错误，请稍后重试')
        } finally {
          set({ loading: false })
        }
      },

      fetchInstance: async (id: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${id}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result: ApiResponse<CIInstance> = await response.json()
          if (result.code === 0) {
            set({ selectedInstance: result.data })
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to fetch CI instance:', error)
          throw error
        }
      },

      createInstance: async (data: CreateCIInstanceRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/ci/instances', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result: ApiResponse<CIInstance> = await response.json()
          if (result.code === 0) {
            // 使用当前保存的 CI 类型 ID 来刷新列表
            const state = get()
            await get().fetchInstances(state.currentCITypeID, state.page, state.pageSize)
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create CI instance:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      updateInstance: async (id: number, data: UpdateCIInstanceRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result: ApiResponse<CIInstance> = await response.json()
          if (result.code === 0) {
            // 使用当前保存的 CI 类型 ID 来刷新列表
            const state = get()
            await get().fetchInstances(state.currentCITypeID, state.page, state.pageSize)
            await get().fetchInstance(id)
            return
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to update CI instance:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      deleteInstance: async (id: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${id}`, {
            method: 'DELETE',
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result: ApiResponse<null> = await response.json()
          if (result.code === 0) {
            // 使用当前保存的 CI 类型 ID 来刷新列表
            const state = get()
            await get().fetchInstances(state.currentCITypeID, state.page, state.pageSize)
            return
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to delete CI instance:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      fetchHistory: async (ciId: number, limit = 50) => {
        try {
          const token = getAuthToken()
          const params = new URLSearchParams()
          if (limit) params.append('limit', limit.toString())

          const response = await fetch(`/api/v1/ci/instances/${ciId}/history?${params}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result: ApiResponse<CIHistory[]> = await response.json()
          if (result.code === 0) {
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to fetch CI history:', error)
          throw error
        }
      },

      fetchRelations: async (ciId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/relations?ci_id=${ciId}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result: ApiResponse<CIRelation[]> = await response.json()
          if (result.code === 0) {
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to fetch CI relations:', error)
          throw error
        }
      },

      createRelation: async (data: CreateCIRelationRequest) => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/ci/relations', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result: ApiResponse<CIRelation> = await response.json()
          if (result.code === 0) {
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create CI relation:', error)
          throw error
        }
      },

      setFilters: (filters) => set({ filters, page: 1 }),

      resetFilters: () => set({ filters: {}, page: 1 }),
    }),
    {
      name: 'cmdb-storage',
      partialize: (state) =>
        Object.fromEntries(
          Object.entries(state).filter(([key]) =>
            !['instances', 'selectedInstance', 'loading', 'currentCITypeID'].includes(key)
          )
        ),
    }
  )
)
