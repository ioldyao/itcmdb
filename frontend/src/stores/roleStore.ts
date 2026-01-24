import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// ==================== 类型定义 ====================

export interface CIRole {
  id: number
  name: string
  display_name: string
  description: string
  color: string
  icon: string
  priority: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface OwnerRole {
  id: number
  name: string
  display_name: string
  description: string
  level: number
  responsibilities: Record<string, any>
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CIInstanceRole {
  id: number
  ci_id: number
  role_type: 'ci_role' | 'owner_role'
  role_id: number
  user_id?: number
  assigned_at: string
  assigned_by?: number
  // 关联数据
  ci_instance?: {
    id: number
    name: string
  }
  user?: {
    id: number
    username: string
    full_name: string
  }
}

export interface RolePermission {
  id: number
  role_name: string
  permissions: string[]
  description: string
  created_at: string
  updated_at: string
}

// ==================== 请求类型 ====================

interface CreateCIRoleRequest {
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  priority?: number
}

interface UpdateCIRoleRequest {
  display_name?: string
  description?: string
  color?: string
  icon?: string
  priority?: number
  is_active?: boolean
}

interface CreateOwnerRoleRequest {
  name: string
  display_name: string
  description?: string
  level?: number
  responsibilities?: Record<string, any>
}

interface UpdateOwnerRoleRequest {
  display_name?: string
  description?: string
  level?: number
  responsibilities?: Record<string, any>
  is_active?: boolean
}

interface AssignRoleRequest {
  role_type: 'ci_role' | 'owner_role'
  role_id: number
  user_id?: number
}

// ==================== Store ====================

interface RoleState {
  ciRoles: CIRole[]
  ownerRoles: OwnerRole[]
  instanceRoles: CIInstanceRole[]
  selectedCIID: number | null
  loading: boolean

  // CI角色操作
  fetchCIRoles: () => Promise<void>
  createCIRole: (data: CreateCIRoleRequest) => Promise<CIRole>
  updateCIRole: (id: number, data: UpdateCIRoleRequest) => Promise<void>
  deleteCIRole: (id: number) => Promise<void>

  // 负责人角色操作
  fetchOwnerRoles: () => Promise<void>
  createOwnerRole: (data: CreateOwnerRoleRequest) => Promise<OwnerRole>
  updateOwnerRole: (id: number, data: UpdateOwnerRoleRequest) => Promise<void>
  deleteOwnerRole: (id: number) => Promise<void>

  // CI实例角色关联
  fetchInstanceRoles: (ciID: number) => Promise<void>
  assignRole: (ciID: number, data: AssignRoleRequest) => Promise<void>
  removeRole: (ciID: number, data: AssignRoleRequest) => Promise<void>
}

const getAuthToken = () => {
  const authState = JSON.parse(localStorage.getItem('auth-storage') || '{}')
  return authState.state?.token || authState.token
}

export const useRoleStore = create<RoleState>()(
  persist(
    (set, get) => ({
      ciRoles: [],
      ownerRoles: [],
      instanceRoles: [],
      selectedCIID: null,
      loading: false,

      // ==================== CI角色操作 ====================

      fetchCIRoles: async () => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles/ci', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ ciRoles: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch CI roles:', error)
        } finally {
          set({ loading: false })
        }
      },

      createCIRole: async (data: CreateCIRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles/ci', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchCIRoles()
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create CI role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      updateCIRole: async (id: number, data: UpdateCIRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/ci/${id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchCIRoles()
        } catch (error) {
          console.error('Failed to update CI role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      deleteCIRole: async (id: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/ci/${id}`, {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchCIRoles()
        } catch (error) {
          console.error('Failed to delete CI role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // ==================== 负责人角色操作 ====================

      fetchOwnerRoles: async () => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles/owner', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ ownerRoles: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch owner roles:', error)
        } finally {
          set({ loading: false })
        }
      },

      createOwnerRole: async (data: CreateOwnerRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles/owner', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchOwnerRoles()
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create owner role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      updateOwnerRole: async (id: number, data: UpdateOwnerRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/owner/${id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchOwnerRoles()
        } catch (error) {
          console.error('Failed to update owner role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      deleteOwnerRole: async (id: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/owner/${id}`, {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchOwnerRoles()
        } catch (error) {
          console.error('Failed to delete owner role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // ==================== CI实例角色关联 ====================

      fetchInstanceRoles: async (ciID: number) => {
        set({ loading: true, selectedCIID: ciID })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/roles`, {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ instanceRoles: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch instance roles:', error)
        } finally {
          set({ loading: false })
        }
      },

      assignRole: async (ciID: number, data: AssignRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/roles`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchInstanceRoles(ciID)
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to assign role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      removeRole: async (ciID: number, data: AssignRoleRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/roles`, {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchInstanceRoles(ciID)
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to remove role:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },
    }),
    {
      name: 'role-storage',
      partialize: (state) => Object.fromEntries(Object.entries(state).filter(([key]) => key !== 'loading')),
    }
  )
)
