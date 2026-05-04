import { create } from 'zustand'
import { alertService, type AlertInstance, type AlertRule } from '@/services/alertService'

interface AlertState {
  // 状态
  alerts: AlertInstance[]
  rules: AlertRule[]
  statistics: any
  loading: boolean
  error: string | null

  // 分页
  total: number
  page: number
  pageSize: number

  // 筛选
  filters: {
    status?: string[]
    severity?: string[]
    category?: string
    handler?: number
    handlingStatus?: string
    objectType?: string
    searchKeyword?: string
    startTime?: string
    endTime?: string
  }

  // Actions
  fetchAlerts: (params?: any) => Promise<void>
  fetchAlertById: (id: number) => Promise<AlertInstance | null>
  acknowledgeAlert: (id: number, data: { handler: number; notes?: string }) => Promise<void>
  closeAlert: (id: number, data: { handler: number; notes?: string }) => Promise<void>
  batchAcknowledge: (alertIds: number[], data: { handler: number; notes?: string }) => Promise<void>
  batchClose: (alertIds: number[], data: { handler: number; notes?: string }) => Promise<void>
  fetchStatistics: () => Promise<void>
  setFilters: (filters: Partial<AlertState['filters']>) => void
  clearFilters: () => void

  // 规则相关
  fetchRules: (params?: any) => Promise<void>
  createRule: (data: Partial<AlertRule>) => Promise<void>
  updateRule: (id: number, data: Partial<AlertRule>) => Promise<void>
  deleteRule: (id: number) => Promise<void>
  enableRule: (id: number) => Promise<void>
  disableRule: (id: number) => Promise<void>
}

export const useAlertStore = create<AlertState>((set, get) => ({
  // 初始状态
  alerts: [],
  rules: [],
  statistics: null,
  loading: false,
  error: null,
  total: 0,
  page: 1,
  pageSize: 20,
  filters: {},

  // 获取告警列表
  fetchAlerts: async (params = {}) => {
    set({ loading: true, error: null })
    try {
      const { filters, page, pageSize } = get()
      const apiParams: Record<string, any> = {
        page,
        page_size: pageSize,
        ...params,
      }
      // 映射 camelCase 筛选参数到 snake_case
      if (filters.status) apiParams.status = filters.status
      if (filters.severity) apiParams.severity = filters.severity
      if (filters.category) apiParams.category = filters.category
      if (filters.handler) apiParams.handler = filters.handler
      if (filters.handlingStatus) apiParams.handling_status = filters.handlingStatus
      if (filters.objectType) apiParams.object_type = filters.objectType
      if (filters.searchKeyword) apiParams.search_keyword = filters.searchKeyword
      if (filters.startTime) apiParams.start_time = filters.startTime
      if (filters.endTime) apiParams.end_time = filters.endTime

      const response = await alertService.getAlerts(apiParams)

      if (response.code === 0 && response.data) {
        set({
          alerts: response.data.alerts,
          total: response.data.total,
          loading: false,
        })
      } else {
        set({ error: response.message || '获取告警列表失败', loading: false })
      }
    } catch (error: any) {
      set({ error: error.message || '获取告警列表失败', loading: false })
    }
  },

  // 获取告警详情
  fetchAlertById: async (id: number) => {
    try {
      const response = await alertService.getAlertById(id)
      if (response.code === 0 && response.data) {
        return response.data
      }
      return null
    } catch (error) {
      console.error('获取告警详情失败:', error)
      return null
    }
  },

  // 确认告警
  acknowledgeAlert: async (id: number, data: { handler: number; notes?: string }) => {
    try {
      await alertService.acknowledgeAlert(id, data)
      // 刷新列表
      get().fetchAlerts()
    } catch (error: any) {
      set({ error: error.message || '确认告警失败' })
      throw error
    }
  },

  // 关闭告警
  closeAlert: async (id: number, data: { handler: number; notes?: string }) => {
    try {
      await alertService.closeAlert(id, data)
      // 刷新列表
      get().fetchAlerts()
    } catch (error: any) {
      set({ error: error.message || '关闭告警失败' })
      throw error
    }
  },

  // 批量确认
  batchAcknowledge: async (alertIds: number[], data: { handler: number; notes?: string }) => {
    try {
      await alertService.batchAcknowledgeAlerts({ alert_ids: alertIds, ...data })
      // 刷新列表
      get().fetchAlerts()
    } catch (error: any) {
      set({ error: error.message || '批量确认失败' })
      throw error
    }
  },

  // 批量关闭
  batchClose: async (alertIds: number[], data: { handler: number; notes?: string }) => {
    try {
      await alertService.batchCloseAlerts({ alert_ids: alertIds, ...data })
      // 刷新列表
      get().fetchAlerts()
    } catch (error: any) {
      set({ error: error.message || '批量关闭失败' })
      throw error
    }
  },

  // 获取统计数据
  fetchStatistics: async () => {
    try {
      const response = await alertService.getAlertStatistics()
      if (response.code === 0 && response.data) {
        set({ statistics: response.data })
      }
    } catch (error: any) {
      console.error('获取统计数据失败:', error)
    }
  },

  // 设置筛选条件
  setFilters: (filters: Partial<AlertState['filters']>) => {
    set({ filters: { ...get().filters, ...filters }, page: 1 })
  },

  // 清除筛选条件
  clearFilters: () => {
    set({ filters: {}, page: 1 })
  },

  // 获取规则列表
  fetchRules: async (params = {}) => {
    set({ loading: true, error: null })
    try {
      const response = await alertService.getRules(params)
      if (response.code === 0 && response.data) {
        set({
          rules: response.data.rules,
          total: response.data.total,
          loading: false,
        })
      } else {
        set({ error: response.message || '获取规则列表失败', loading: false })
      }
    } catch (error: any) {
      set({ error: error.message || '获取规则列表失败', loading: false })
    }
  },

  // 创建规则
  createRule: async (data: Partial<AlertRule>) => {
    try {
      await alertService.createRule(data)
      // 刷新列表
      get().fetchRules()
    } catch (error: any) {
      set({ error: error.message || '创建规则失败' })
      throw error
    }
  },

  // 更新规则
  updateRule: async (id: number, data: Partial<AlertRule>) => {
    try {
      await alertService.updateRule(id, data)
      // 刷新列表
      get().fetchRules()
    } catch (error: any) {
      set({ error: error.message || '更新规则失败' })
      throw error
    }
  },

  // 删除规则
  deleteRule: async (id: number) => {
    try {
      await alertService.deleteRule(id)
      // 刷新列表
      get().fetchRules()
    } catch (error: any) {
      set({ error: error.message || '删除规则失败' })
      throw error
    }
  },

  // 启用规则
  enableRule: async (id: number) => {
    try {
      await alertService.enableRule(id)
      // 刷新列表
      get().fetchRules()
    } catch (error: any) {
      set({ error: error.message || '启用规则失败' })
      throw error
    }
  },

  // 禁用规则
  disableRule: async (id: number) => {
    try {
      await alertService.disableRule(id)
      // 刷新列表
      get().fetchRules()
    } catch (error: any) {
      set({ error: error.message || '禁用规则失败' })
      throw error
    }
  },
}))
