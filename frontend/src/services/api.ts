import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

let isRefreshing = false
let failedQueue: { resolve: (value: unknown) => void; reject: (reason?: unknown) => void }[] = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return api(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      const currentToken = useAuthStore.getState().token
      if (!currentToken) {
        isRefreshing = false
        useAuthStore.getState().logout()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const response = await fetch('/api/v1/auth/refresh', {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${currentToken}`,
          },
        })
        const data = await response.json()

        if (data.code === 0 && data.data?.token) {
          const newToken = data.data.token
          const currentUser = useAuthStore.getState().user

          // Parse new permissions from JWT
          try {
            const payload = JSON.parse(atob(newToken.split('.')[1]))
            const permissions = payload.permissions || []
            useAuthStore.getState().setAuth(currentUser!, newToken, permissions)
          } catch {
            useAuthStore.getState().setAuth(currentUser!, newToken, [])
          }

          processQueue(null, newToken)
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          return api(originalRequest)
        } else {
          processQueue(error, null)
          useAuthStore.getState().logout()
          window.location.href = '/login'
          return Promise.reject(error)
        }
      } catch (refreshError) {
        processQueue(refreshError, null)
        useAuthStore.getState().logout()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export default api
