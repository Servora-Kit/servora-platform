import { createFileRoute, redirect } from '@tanstack/react-router'
import { isAuthenticated, isSuperAdmin } from '#/stores/auth'

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (!isAuthenticated()) {
      throw redirect({ to: '/login' as string })
    }
    if (isSuperAdmin()) {
      throw redirect({ to: '/tenants' as string })
    }
    throw redirect({ to: '/dashboard' as string })
  },
})
