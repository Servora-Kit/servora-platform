import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { iamClients } from '#/api'
import { Page } from '#/components/page'
import { DataTable } from '#/components/data-table'
import { FormDrawer } from '#/components/form-drawer'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Plus } from 'lucide-react'
import { toast } from '#/lib/toast'

export const Route = createFileRoute('/_app/applications/')({
  component: ApplicationListPage,
})

type App = NonNullable<Awaited<ReturnType<typeof iamClients.application.ListApplications>>['applications']>[number]

const columns: ColumnDef<App, unknown>[] = [
  {
    accessorKey: 'name',
    header: '名称',
    cell: ({ row }) => (
      <Link
        to="/applications/$appId"
        params={{ appId: row.original.id ?? '' }}
        className="font-medium text-foreground hover:underline"
      >
        {row.getValue('name')}
      </Link>
    ),
  },
  {
    accessorKey: 'clientId',
    header: 'Client ID',
    cell: ({ row }) => (
      <span className="font-mono text-xs text-muted-foreground">{row.getValue('clientId')}</span>
    ),
  },
  {
    accessorKey: 'createdAt',
    header: '创建时间',
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.original.createdAt ? new Date(row.original.createdAt).toLocaleDateString('zh-CN') : '-'}
      </span>
    ),
  },
]

function ApplicationListPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const { data, isLoading } = useQuery({
    queryKey: ['applications', 'list', page, pageSize],
    queryFn: () => iamClients.application.ListApplications({ pagination: { page: { page, pageSize } } }),
  })

  const apps = data?.applications ?? []
  const total = data?.pagination?.page?.total ?? 0

  return (
    <Page
      title="应用"
      description="管理 OIDC 客户端应用。"
      extra={<CreateAppButton onCreated={() => void queryClient.invalidateQueries({ queryKey: ['applications'] })} />}
    >
      <DataTable
        columns={columns}
        data={apps}
        isLoading={isLoading}
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />
    </Page>
  )
}

function CreateAppButton({ onCreated }: { onCreated: () => void }) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [redirectUris, setRedirectUris] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit() {
    setLoading(true)
    try {
      await iamClients.application.CreateApplication({
        data: {
          id: '',
          clientId: '',
          name,
          redirectUris: redirectUris.split('\n').map((u) => u.trim()).filter(Boolean),
          scopes: [],
          grantTypes: [],
          applicationType: '',
          accessTokenType: '',
          type: '',
          idTokenLifetime: 0,
        },
      })
      setOpen(false)
      setName('')
      setRedirectUris('')
      onCreated()
      toast.success('应用注册成功')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="size-4" />
        注册应用
      </Button>
      <FormDrawer open={open} onOpenChange={setOpen} title="注册应用" loading={loading} onSubmit={handleSubmit} submitLabel="创建">
        <div className="space-y-2">
          <Label>名称</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <Label>Redirect URIs（每行一个）</Label>
          <textarea
            className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            value={redirectUris}
            onChange={(e) => setRedirectUris(e.target.value)}
          />
        </div>
      </FormDrawer>
    </>
  )
}
