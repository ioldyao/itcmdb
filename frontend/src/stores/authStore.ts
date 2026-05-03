import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  username: string
  email: string
  fullName: string
}

interface AuthState {
  user: User | null
  token: string | null
  permissions: string[]
  isAuthenticated: boolean
  _hasHydrated: boolean
  setHasHydrated: (state: boolean) => void
  setAuth: (user: User, token: string, permissions: string[]) => void
  logout: () => void
  hasPermission: (resource: string, action: string) => boolean
  refreshPermissions: () => Promise<void>
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      permissions: [],
      isAuthenticated: false,
      _hasHydrated: false,
      setHasHydrated: (state) => {
        set({ _hasHydrated: state })
      },
      setAuth: (user, token, permissions) =>
        set({ user, token, permissions, isAuthenticated: true }),
      logout: () =>
        set({ user: null, token: null, permissions: [], isAuthenticated: false }),
      hasPermission: (resource, action) => {
        const { permissions } = get()
        const requiredPermission = `${resource}:${action}`
        for (const perm of permissions) {
          if (perm === requiredPermission || perm === `${resource}:*` || perm === '*:*') {
            return true
          }
        }
        return false
      },
      refreshPermissions: async () => {
        const { token, user } = get()
        if (!token || !user) return

        try {
          const response = await fetch('/api/v1/users/me/permissions', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const data = await response.json()

          if (data.code === 0) {
            set({ permissions: data.data || [] })
          }
        } catch (error) {
          console.error('Failed to refresh permissions:', error)
        }
      },
    }),
    {
      name: 'auth-storage',
      onRehydrateStorage: () => {
        return (state) => {
          state?.setHasHydrated(true)
        }
      },
    }
  )
)
