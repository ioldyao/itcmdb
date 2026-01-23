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
  setAuth: (user: User, token: string, permissions: string[]) => void
  logout: () => void
  hasPermission: (resource: string, action: string) => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      permissions: [],
      isAuthenticated: false,
      setAuth: (user, token, permissions) =>
        set({ user, token, permissions, isAuthenticated: true }),
      logout: () =>
        set({ user: null, token: null, permissions: [], isAuthenticated: false }),
      hasPermission: (resource, action) => {
        const { permissions } = get()
        const requiredPermission = `${resource}:${action}`
        return permissions.includes(requiredPermission) || permissions.includes('*:*')
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)
