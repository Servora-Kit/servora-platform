import { useEffect } from 'react'
import { createFileRoute, Outlet, redirect, useNavigate } from '@tanstack/react-router'
import { PlatformShell } from '#/layout/platform-shell'
import { LoginExpiredDialog } from '#/components/login-expired-dialog'
import { isAuthenticated, isSuperAdmin } from '#/stores/auth'

export const Route = createFileRoute('/_platform')({
  beforeLoad: ({ location }) => {
    if (typeof window === 'undefined') return
    if (!isAuthenticated()) {
      throw redirect({ to: '/login' as string, search: { redirect: location.pathname } })
    }
    if (!isSuperAdmin()) {
      throw redirect({ to: '/dashboard' as string })
    }
  },
  component: PlatformLayout,
})

function PlatformLayout() {
  const navigate = useNavigate()

  useEffect(() => {
    if (!isAuthenticated()) {
      void navigate({ to: '/login' as string, search: { redirect: window.location.pathname } })
    }
  }, [navigate])

  if (typeof window !== 'undefined' && !isAuthenticated()) {
    return null
  }

  return (
    <PlatformShell>
      <Outlet />
      <LoginExpiredDialog />
    </PlatformShell>
  )
}
