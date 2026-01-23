import api from './api'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: {
    id: number
    username: string
    email: string
    fullName: string
  }
  permissions: string[]
}

export const authService = {
  login: (data: LoginRequest) => api.post<any, LoginResponse>('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  refreshToken: () => api.post('/auth/refresh'),
  getMe: () => api.get('/users/me'),
}
