import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// ==================== 类型定义 ====================

export interface TagCategory {
  id: number
  name: string
  display_name: string
  description: string
  color: string
  icon: string
  sort_order: number
  is_system: boolean
  created_at: string
  updated_at: string
  tags?: Tag[]
}

export interface Tag {
  id: number
  category_id?: number
  name: string
  display_name: string
  color: string
  description: string
  usage_count: number
  is_active: boolean
  created_at: string
  updated_at: string
  category?: TagCategory
}

export interface CITag {
  id: number
  ci_id: number
  tag_id: number
  tagged_by?: number
  tagged_at: string
  tag?: Tag
  ci_instance?: {
    id: number
    name: string
  }
}

export interface TagStat {
  id: number
  name: string
  display_name: string
  color: string
  category_id?: number
  category_name: string
  usage_count: number
}

// ==================== 请求类型 ====================

interface CreateTagCategoryRequest {
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
}

interface UpdateTagCategoryRequest {
  display_name?: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
}

interface CreateTagRequest {
  category_id?: number
  name: string
  display_name: string
  color?: string
  description?: string
}

interface UpdateTagRequest {
  display_name?: string
  color?: string
  description?: string
  is_active?: boolean
}

// ==================== Store ====================

interface TagState {
  categories: TagCategory[]
  tags: Tag[]
  ciTags: CITag[]
  selectedCIID: number | null
  stats: TagStat[]
  loading: boolean

  // 分类操作
  fetchCategories: () => Promise<void>
  createCategory: (data: CreateTagCategoryRequest) => Promise<TagCategory>
  updateCategory: (id: number, data: UpdateTagCategoryRequest) => Promise<void>
  deleteCategory: (id: number) => Promise<void>

  // 标签操作
  fetchTags: (categoryId?: number) => Promise<void>
  createTag: (data: CreateTagRequest) => Promise<Tag>
  updateTag: (id: number, data: UpdateTagRequest) => Promise<void>
  deleteTag: (id: number) => Promise<void>
  fetchStats: () => Promise<void>

  // CI实例标签操作
  fetchCITags: (ciID: number) => Promise<void>
  assignTag: (ciID: number, tagID: number) => Promise<void>
  removeTag: (ciID: number, tagID: number) => Promise<void>

  // 批量操作
  batchAssignTags: (ciIDs: number[], tagID: number) => Promise<void>
  batchRemoveTags: (ciIDs: number[], tagID: number) => Promise<void>
}

const getAuthToken = () => {
  const authState = JSON.parse(localStorage.getItem('auth-storage') || '{}')
  return authState.state?.token || authState.token
}

export const useTagStore = create<TagState>()(
  persist(
    (set, get) => ({
      categories: [],
      tags: [],
      ciTags: [],
      selectedCIID: null,
      stats: [],
      loading: false,

      // ==================== 分类操作 ====================

      fetchCategories: async () => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags/categories', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ categories: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch tag categories:', error)
        } finally {
          set({ loading: false })
        }
      },

      createCategory: async (data: CreateTagCategoryRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags/categories', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchCategories()
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create tag category:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      updateCategory: async (id: number, data: UpdateTagCategoryRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/tags/categories/${id}`, {
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
          await get().fetchCategories()
        } catch (error) {
          console.error('Failed to update tag category:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      deleteCategory: async (id: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/tags/categories/${id}`, {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchCategories()
        } catch (error) {
          console.error('Failed to delete tag category:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // ==================== 标签操作 ====================

      fetchTags: async (categoryId?: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const url = categoryId ? `/api/v1/tags?category_id=${categoryId}` : '/api/v1/tags'
          const response = await fetch(url, {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ tags: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch tags:', error)
        } finally {
          set({ loading: false })
        }
      },

      createTag: async (data: CreateTagRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(data),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchTags()
            return result.data
          }
          throw new Error(result.message)
        } catch (error) {
          console.error('Failed to create tag:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      updateTag: async (id: number, data: UpdateTagRequest) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/tags/${id}`, {
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
          await get().fetchTags()
        } catch (error) {
          console.error('Failed to update tag:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      deleteTag: async (id: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/tags/${id}`, {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code !== 0) {
            throw new Error(result.message)
          }
          await get().fetchTags()
        } catch (error) {
          console.error('Failed to delete tag:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      fetchStats: async () => {
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags/stats', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ stats: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch tag stats:', error)
        }
      },

      // ==================== CI实例标签操作 ====================

      fetchCITags: async (ciID: number) => {
        set({ loading: true, selectedCIID: ciID })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/tags`, {
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            set({ ciTags: result.data })
          }
        } catch (error) {
          console.error('Failed to fetch CI tags:', error)
        } finally {
          set({ loading: false })
        }
      },

      assignTag: async (ciID: number, tagID: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/tags`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ tag_id: tagID }),
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchCITags(ciID)
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to assign tag:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      removeTag: async (ciID: number, tagID: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch(`/api/v1/ci/instances/${ciID}/tags/${tagID}`, {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${token}` },
          })
          const result = await response.json()
          if (result.code === 0) {
            await get().fetchCITags(ciID)
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to remove tag:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      // ==================== 批量操作 ====================

      batchAssignTags: async (ciIDs: number[], tagID: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags/batch/assign', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ ci_ids: ciIDs, tag_id: tagID }),
          })
          const result = await response.json()
          if (result.code === 0) {
            // 刷新当前选中的CI标签
            const selectedId = get().selectedCIID
            if (selectedId) {
              await get().fetchCITags(selectedId)
            }
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to batch assign tags:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },

      batchRemoveTags: async (ciIDs: number[], tagID: number) => {
        set({ loading: true })
        try {
          const token = getAuthToken()
          const response = await fetch('/api/v1/tags/batch/remove', {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({ ci_ids: ciIDs, tag_id: tagID }),
          })
          const result = await response.json()
          if (result.code === 0) {
            const selectedId = get().selectedCIID
            if (selectedId) {
              await get().fetchCITags(selectedId)
            }
          } else {
            throw new Error(result.message)
          }
        } catch (error) {
          console.error('Failed to batch remove tags:', error)
          throw error
        } finally {
          set({ loading: false })
        }
      },
    }),
    {
      name: 'tag-storage',
      partialize: (state) => Object.fromEntries(Object.entries(state).filter(([key]) => key !== 'loading')),
    }
  )
)
