import { useEffect } from 'react'
import { useStore } from '@tanstack/react-store'
import { useNavigate, useLocation } from '@tanstack/react-router'
import { LayoutDashboard, AppWindow, Users } from 'lucide-react'
import { TooltipProvider } from '#/components/ui/tooltip'
import { Sidebar } from '#/layout/sidebar'
import type { MenuItem } from '#/layout/sidebar'
import { Header } from '#/layout/header'
import { Tabbar } from '#/layout/tabbar'
import { authStore, clearAuth, setUser } from '#/stores/auth'
import { layoutStore, SIDEBAR_WIDTH, SIDEBAR_COLLAPSED_WIDTH } from '#/stores/layout'
import { addTab, resetTabs } from '#/stores/tabbar'
import { iamClients } from '#/api'

const MENU_MAIN: MenuItem[] = [
  { label: '概览', href: '/dashboard', icon: LayoutDashboard },
  {
    label: '应用管理',
    icon: AppWindow,
    children: [{ label: '应用列表', href: '/applications' }],
  },
  {
    label: '用户管理',
    icon: Users,
    children: [{ label: '用户列表', href: '/users' }],
  },
]

const ROUTE_TITLES: Record<string, string> = {
  '/dashboard': '概览',
  '/applications': '应用列表',
  '/users': '用户列表',
  '/settings/profile': '个人设置',
  '/settings/security': '安全设置',
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const user = useStore(authStore, (s) => s.user)
  const collapsed = useStore(layoutStore, (s) => s.sidebarCollapsed)
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    resetTabs({ path: '/dashboard', title: '概览' })
  }, [])

  useEffect(() => {
    if (user || !authStore.state.accessToken) return
    iamClients.user
      .CurrentUserInfo({})
      .then((info) => {
        setUser({
          id: info.id ?? '',
          name: info.username ?? info.email ?? '',
          email: info.email ?? '',
          role: info.role ?? '',
        })
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

  function handleLogout() {
    clearAuth()
    void navigate({ to: '/login' as string })
  }

  const sidebarWidth = collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_WIDTH

  return (
    <TooltipProvider>
      <div className="min-h-dvh bg-background-deep">
        <Sidebar
          title="Servora IAM"
          titleHref="/dashboard"
          menuGroups={[MENU_MAIN, []]}
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
