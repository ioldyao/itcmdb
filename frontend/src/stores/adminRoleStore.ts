import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// 系统角色接口
interface Role {
  id: number
  name: string
  description: string
  created_at: string
  updated_at: string
}

// 权限接口
interface Permission {
  id: number
  resource: string
  action: string
}

interface AdminRoleState {
  roles: Role[]
  permissions: Permission[]
  loading: boolean
  error: string | null

  // 角色操作
  fetchRoles: () => Promise<void>
  createRole: (name: string, description: string) => Promise<void>
  updateRole: (id: number, name: string, description: string) => Promise<void>
  deleteRole: (id: number) => Promise<void>
  getRolePermissions: (roleId: number) => Promise<Permission[]>

  // 权限操作
  fetchPermissions: () => Promise<void>
  createPermission: (resource: string, action: string) => Promise<void>
  deletePermission: (id: number) => Promise<void>

  // 关联操作
  assignPermissionToRole: (roleId: number, permissionId: number) => Promise<void>
  removePermissionFromRole: (roleId: number, permissionId: number) => Promise<void>
  assignRoleToUser: (userId: number, roleId: number) => Promise<void>
  removeRoleFromUser: (userId: number, roleId: number) => Promise<void>
}

const getAuthToken = () => {
  const authState = JSON.parse(localStorage.getItem('auth-storage') || '{}')
  return authState.state?.token || authState.token
}

export const useAdminRoleStore = create<AdminRoleState>()(
  persist(
    (set, get) => ({
      roles: [],
      permissions: [],
      loading: false,
      error: null,

      // 获取所有角色
      fetchRoles: async () => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ roles: result.data, loading: false })
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to fetch roles:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        }
      },

      // 创建角色
      createRole: async (name: string, description: string) => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/roles', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ name, description }),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchRoles()
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to create role:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // 更新角色
      updateRole: async (id: number, name: string, description: string) => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/${id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ name, description }),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchRoles()
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to update role:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // 删除角色
      deleteRole: async (id: number) => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/${id}`, {
            method: 'DELETE',
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchRoles()
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to delete role:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // 获取角色权限
      getRolePermissions: async (roleId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/roles/${roleId}/permissions`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result = await response.json()
          if (result.code === 0) {
            return result.data
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to get role permissions:', error)
          throw error
        }
      },

      // 获取所有权限
      fetchPermissions: async () => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/permissions', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ permissions: result.data, loading: false })
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to fetch permissions:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        }
      },

      // 创建权限
      createPermission: async (resource: string, action: string) => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/permissions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ resource, action }),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchPermissions()
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to create permission:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // 删除权限
      deletePermission: async (id: number) => {
        set({ loading: true, error: null })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/permissions/${id}`, {
            method: 'DELETE',
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchPermissions()
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to delete permission:', error)
          set({ error: (error as Error).message, loading: false })
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // 为角色分配权限
      assignPermissionToRole: async (roleId: number, permissionId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/role-permissions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ role_id: roleId, permission_id: permissionId }),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to assign permission:', error)
          throw error
        }
      },

      // 移除角色权限
      removePermissionFromRole: async (roleId: number, permissionId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/role-permissions', {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ role_id: roleId, permission_id: permissionId }),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to remove permission:', error)
          throw error
        }
      },

      // 为用户分配角色
      assignRoleToUser: async (userId: number, roleId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/user-roles', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ user_id: userId, role_id: roleId }),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to assign role:', error)
          throw error
        }
      },

      // 移除用户角色
      removeRoleFromUser: async (userId: number, roleId: number) => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/user-roles', {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ user_id: userId, role_id: roleId }),
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to remove role:', error)
          throw error
        }
      },
    }),
    {
      name: 'admin-role-storage',
      partialize: (state) => Object.fromEntries(Object.entries(state).filter(([key]) => key !== 'loading')),
    }
  )
)
