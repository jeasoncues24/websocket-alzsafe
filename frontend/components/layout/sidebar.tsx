'use client'

import { usePathname, useRouter } from 'next/navigation'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Building2,
  MessageSquare,
  Wifi,
  Send,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { useAppStore } from '@/stores/useAppStore'
import { Button } from '@/components/ui/button'
import { useTheme } from 'next-themes'
import { Moon, Sun, LogOut } from 'lucide-react'

const navItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
  { id: 'companies', label: 'Empresas', icon: Building2, href: '/companies' },
  { id: 'messages', label: 'Mensajes', icon: MessageSquare, href: '/messages' },
  { id: 'sessions', label: 'Sesiones', icon: Wifi, href: '/sessions' },
  { id: 'broadcasts', label: 'Broadcasts', icon: Send, href: '/broadcasts' },
  { id: 'settings', label: 'Settings', icon: Settings, href: '/settings' },
]

export function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()
  const { sidebarOpen, setSidebarOpen, activeNav, setActiveNav } = useAppStore()
  const { theme, setTheme, resolvedTheme } = useTheme()

  const currentTheme = theme || resolvedTheme

  const handleNavClick = (item: typeof navItems[0]) => {
    setActiveNav(item.id)
    router.push(item.href)
  }

  return (
    <div
      className={cn(
        'flex flex-col h-screen bg-background border-r transition-all duration-300',
        sidebarOpen ? 'w-64' : 'w-16'
      )}
    >
      <div className="flex items-center justify-between p-4 border-b">
        {sidebarOpen && (
          <span className="font-semibold text-lg">WhatsApp API</span>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setSidebarOpen(!sidebarOpen)}
        >
          {sidebarOpen ? <ChevronLeft className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </Button>
      </div>

      <nav className="flex-1 p-2 space-y-1">
        {navItems.map((item) => {
          const isActive = pathname.startsWith(item.href)
          return (
            <button
              key={item.id}
              onClick={() => handleNavClick(item)}
              className={cn(
                'flex items-center w-full p-3 rounded-lg transition-colors',
                isActive
                  ? 'bg-primary/10 text-primary'
                  : 'hover:bg-muted',
                !sidebarOpen && 'justify-center'
              )}
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {sidebarOpen && (
                <span className="ml-3">{item.label}</span>
              )}
            </button>
          )
        })}
      </nav>

      <div className="p-4 border-t flex justify-between">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(currentTheme === 'dark' ? 'light' : 'dark')}
        >
          {currentTheme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => {
            localStorage.removeItem("admin_token")
            window.location.href = "/login"
          }}
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}