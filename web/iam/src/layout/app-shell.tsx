import { useEffect } from 'react'
import { useStore } from '@tanstack/react-store'
import { useQuery } from '@tanstack/react-query'
import { useNavigate, useLocation } from '@tanstack/react-router'
import {
  LayoutDashboard,
  Building2,
  AppWindow,
  Users,
  SlidersHorizontal,
} from 'lucide-react'
import { TooltipProvider } from '#/components/ui/tooltip'
import { Sidebar } from '#/layout/sidebar'
import type { MenuItem } from '#/layout/sidebar'
import { Header } from '#/layout/header'
import { Tabbar } from '#/layout/tabbar'
import { authStore, clearAuth, setUser } from '#/stores/auth'
import {
  scopeStore,
  setCurrentTenantId,
  setCurrentOrganizationId,
  clearScope,
} from '#/stores/scope'
import { layoutStore, SIDEBAR_WIDTH, SIDEBAR_COLLAPSED_WIDTH } from '#/stores/layout'
import { addTab, resetTabs } from '#/stores/tabbar'
import { iamClients } from '#/api'

const MENU_MAIN: MenuItem[] = [
  { label: '概览', href: '/dashboard', icon: LayoutDashboard },
  {
    label: '组织管理',
    icon: Building2,
    children: [
      { label: '组织列表', href: '/organizations' },
    ],
  },
  {
    label: '应用管理',
    icon: AppWindow,
    children: [
      { label: '应用列表', href: '/applications' },
    ],
  },
  {
    label: '用户管理',
    icon: Users,
    children: [
      { label: '用户列表', href: '/users' },
    ],
  },
]

const MENU_SETTINGS: MenuItem[] = [
  {
    label: '系统设置',
    icon: SlidersHorizontal,
    children: [
      { label: '个人设置', href: '/settings/profile' },
    ],
  },
]

const ROUTE_TITLES: Record<string, string> = {
  '/dashboard': '概览',
  '/organizations': '组织列表',
  '/applications': '应用列表',
  '/users': '用户列表',
  '/settings/profile': '个人设置',
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const user = useStore(authStore, (s) => s.user)
  const collapsed = useStore(layoutStore, (s) => s.sidebarCollapsed)
  const currentTenantId = useStore(scopeStore, (s) => s.currentTenantId)
  const currentOrgId = useStore(scopeStore, (s) => s.currentOrganizationId)
  const navigate = useNavigate()
  const location = useLocation()

  // Reset tabbar when entering the app shell to prevent platform-shell tabs from leaking in.
  useEffect(() => {
    resetTabs({ path: '/dashboard', title: '概览' })
  }, [])

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

  // Auto-fetch tenant on mount
  useEffect(() => {
    if (currentTenantId || !authStore.state.accessToken) return
    iamClients.tenant
      .ListTenants({ pagination: { page: { page: 1, pageSize: 100 } } })
      .then((res) => {
        const firstId = res.tenants?.[0]?.id
        if (firstId) setCurrentTenantId(firstId)
      })
      .catch(() => {})
  }, [currentTenantId])

  // Org query
  const { data: orgs } = useQuery({
    queryKey: ['organizations', 'list-for-switcher', currentTenantId],
    queryFn: () =>
      iamClients.organization.ListOrganizations({
        pagination: { page: { page: 1, pageSize: 100 } },
      }),
    enabled: !!currentTenantId,
    staleTime: 60_000,
  })

  const orgItems =
    orgs?.organizations?.map((o) => ({ id: o.id ?? '', name: o.displayName || o.name || '' })) ?? []

  useEffect(() => {
    if (currentOrgId || orgItems.length === 0) return
    setCurrentOrganizationId(orgItems[0].id)
  }, [currentOrgId, orgItems])

  function handleLogout() {
    clearAuth()
    clearScope()
    void navigate({ to: '/login' as string })
  }

  // Sync route to tabbar — always use the base path as the tab key so that
  // sub-routes (e.g. /projects/xxx/members) don't create duplicate tabs.
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
          title="Servora IAM"
          titleHref="/dashboard"
          menuGroups={[MENU_MAIN, MENU_SETTINGS]}
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
