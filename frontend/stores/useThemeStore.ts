import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Theme = 'light' | 'dark' | 'system'
type UIDensity = 'compact' | 'default' | 'spacious'

interface ThemeState {
  theme: Theme
  setTheme: (theme: Theme) => void
  uiDensity: UIDensity
  setUIDensity: (density: UIDensity) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: 'system',
      setTheme: (theme) => set({ theme }),
      uiDensity: 'default',
      setUIDensity: (uiDensity) => set({ uiDensity }),
    }),
    {
      name: 'wsapi-theme-storage',
    }
  )
)