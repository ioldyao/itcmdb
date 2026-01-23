export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

export interface ApiError {
  code: number
  message: string
  details?: any
}

export interface PaginationParams {
  page: number
  pageSize: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}
