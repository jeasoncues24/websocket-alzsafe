import { create } from 'zustand'

interface AppState {
  activeNav: string
  setActiveNav: (nav: string) => void
  sidebarOpen: boolean
  setSidebarOpen: (open: boolean) => void
  refreshInterval: number
  setRefreshInterval: (interval: number) => void
  itemsPerPage: number
  setItemsPerPage: (count: number) => void
  allowedModules: string[]
  setAllowedModules: (modules: string[]) => void
  user: {
    id: number
    username: string
    email: string
    role_id: number | null
    is_root: boolean
    activo: boolean
  } | null
  setUser: (user: {
    id: number
    username: string
    email: string
    role_id: number | null
    is_root: boolean
    activo: boolean
  } | null) => void
}

export const useAppStore = create<AppState>((set) => ({
  activeNav: 'dashboard',
  setActiveNav: (activeNav) => set({ activeNav }),
  sidebarOpen: true,
  setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
  refreshInterval: 30000,
  setRefreshInterval: (refreshInterval) => set({ refreshInterval }),
  itemsPerPage: 20,
  setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),
  allowedModules: [],
  setAllowedModules: (allowedModules) => set({ allowedModules }),
  user: null,
  setUser: (user) => set({ user }),
}))