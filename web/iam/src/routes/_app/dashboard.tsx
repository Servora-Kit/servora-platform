import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Users, AppWindow } from 'lucide-react'
import { iamClients } from '#/api'
import { Page } from '#/components/page'
import { KpiCard } from '#/components/kpi-card'

export const Route = createFileRoute('/_app/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const userCount = useQuery({
    queryKey: ['dashboard', 'user-count'],
    queryFn: () => iamClients.user.ListUsers({ pagination: { page: { page: 1, pageSize: 1 } } }),
    staleTime: 60_000,
  })

  const appCount = useQuery({
    queryKey: ['dashboard', 'app-count'],
    queryFn: () =>
      iamClients.application.ListApplications({
        pagination: { page: { page: 1, pageSize: 1 } },
      }),
    staleTime: 60_000,
  })

  const getTotal = (data: unknown): number | undefined => {
    const d = data as { pagination?: { page?: { total?: number } } } | undefined
    return d?.pagination?.page?.total
  }

  return (
    <Page title="概览" contentClass="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <KpiCard
          title="用户"
          value={getTotal(userCount.data)}
          icon={Users}
          href="/users"
          isLoading={userCount.isLoading}
        />
        <KpiCard
          title="应用"
          value={getTotal(appCount.data)}
          icon={AppWindow}
          href="/applications"
          isLoading={appCount.isLoading}
        />
      </div>
    </Page>
  )
}
