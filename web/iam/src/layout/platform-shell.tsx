import { useEffect } from 'react'
import { useStore } from '@tanstack/react-store'
import { useNavigate, useLocation } from '@tanstack/react-router'
import { Building } from 'lucide-react'
import { TooltipProvider } from '#/components/ui/tooltip'
import { Sidebar } from '#/layout/sidebar'
import type { MenuItem } from '#/layout/sidebar'
import { Header } from '#/layout/header'
import { Tabbar } from '#/layout/tabbar'
import { authStore, clearAuth, setUser } from '#/stores/auth'
import { layoutStore, SIDEBAR_WIDTH, SIDEBAR_COLLAPSED_WIDTH } from '#/stores/layout'
import { addTab, resetTabs } from '#/stores/tabbar'
import { iamClients } from '#/api'

const MENU_ITEMS: MenuItem[] = [
  {
    label: '租户管理',
    icon: Building,
    children: [
      { label: '租户列表', href: '/tenants' },
    ],
  },
]

const ROUTE_TITLES: Record<string, string> = {
  '/tenants': '租户列表',
}

export function PlatformShell({ children }: { children: React.ReactNode }) {
  const user = useStore(authStore, (s) => s.user)
  const collapsed = useStore(layoutStore, (s) => s.sidebarCollapsed)
  const navigate = useNavigate()
  const location = useLocation()

  // Reset tabbar when entering the platform shell to prevent app-shell tabs from leaking in.
  useEffect(() => {
    resetTabs({ path: '/tenants', title: '租户列表' })
  }, [])

  function handleLogout() {
    clearAuth()
    void navigate({ to: '/login' as string })
  }

  // Fetch current user info on mount if not yet loaded
  useEffect(() => {
    if (user || !authStore.state.accessToken) return
    iamClients.user
      .CurrentUserInfo({})
      .then((info) => {
        setUser({ id: info.id ?? '', name: info.name ?? '', email: info.email ?? '', role: info.role ?? '' })
      })
      .catch(() => {})
  }, [user])

  useEffect(() => {
    const path = location.pathname
    const baseMatch = Object.keys(ROUTE_TITLES).find(
      (k) => path === k || path.startsWith(`${k}/`),
    )
    if (baseMatch) {
      addTab({ path: baseMatch, title: ROUTE_TITLES[baseMatch] })
    } else {
      addTab({ path, title: path.split('/').filter(Boolean).pop() ?? path })
    }
  }, [location.pathname])

  const sidebarWidth = collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_WIDTH

  return (
    <TooltipProvider>
      <div className="min-h-dvh bg-background-deep">
        <Sidebar
          title="Servora 平台"
          titleHref="/tenants"
          menuGroups={[MENU_ITEMS]}
        />

        <div
          className="flex min-h-dvh flex-col transition-all duration-200"
          style={{ marginLeft: sidebarWidth }}
        >
          <Header user={user} onLogout={handleLogout} />
          <Tabbar />
          <main className="flex-1 p-4">{children}</main>
        </div>
      </div>
    </TooltipProvider>
  )
}
