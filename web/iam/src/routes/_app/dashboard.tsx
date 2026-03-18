import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { Building2, Users, AppWindow } from 'lucide-react'
import { iamClients } from '#/api'
import { Page } from '#/components/page'
import { KpiCard } from '#/components/kpi-card'
import { QuickActions } from '#/components/quick-actions'
import { DataTable } from '#/components/data-table'

export const Route = createFileRoute('/_app/dashboard')({
  component: DashboardPage,
})

interface RecentOrg {
  id?: string
  name?: string
  displayName?: string
  createdAt?: string
}

const recentOrgColumns: ColumnDef<RecentOrg, unknown>[] = [
  {
    accessorKey: 'name',
    header: '名称',
    cell: ({ row }) => (
      <span className="font-medium text-foreground">
        {row.original.displayName || row.original.name}
      </span>
    ),
  },
  {
    accessorKey: 'createdAt',
    header: '创建时间',
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.original.createdAt
          ? new Date(row.original.createdAt).toLocaleDateString('zh-CN')
          : '-'}
      </span>
    ),
  },
]

function DashboardPage() {
  const userCount = useQuery({
    queryKey: ['dashboard', 'user-count'],
    queryFn: () =>
      iamClients.user.ListUsers({ pagination: { page: { page: 1, pageSize: 1 } } }),
    staleTime: 60_000,
  })

  const orgCount = useQuery({
    queryKey: ['dashboard', 'org-count'],
    queryFn: () =>
      iamClients.organization.ListOrganizations({
        pagination: { page: { page: 1, pageSize: 1 } },
      }),
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

  const recentOrgs = useQuery({
    queryKey: ['dashboard', 'recent-orgs'],
    queryFn: () =>
      iamClients.organization.ListOrganizations({
        pagination: { page: { page: 1, pageSize: 5 } },
      }),
    staleTime: 60_000,
  })

  const getTotal = (data: unknown): number | undefined => {
    const d = data as { pagination?: { page?: { total?: number } } } | undefined
    return d?.pagination?.page?.total
  }

  return (
    <Page title="概览" contentClass="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <KpiCard
          title="用户"
          value={getTotal(userCount.data)}
          icon={Users}
          href="/users"
          isLoading={userCount.isLoading}
        />
        <KpiCard
          title="组织"
          value={getTotal(orgCount.data)}
          icon={Building2}
          href="/organizations"
          isLoading={orgCount.isLoading}
        />
        <KpiCard
          title="应用"
          value={getTotal(appCount.data)}
          icon={AppWindow}
          href="/applications"
          isLoading={appCount.isLoading}
        />
      </div>

      <QuickActions />

      <DataTable
        title="最近创建的组织"
        columns={recentOrgColumns}
        data={(recentOrgs.data?.organizations ?? []) as RecentOrg[]}
        isLoading={recentOrgs.isLoading}
      />
    </Page>
  )
}
