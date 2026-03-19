import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { iamClients } from '#/api'
import { Page } from '#/components/page'
import { DataTable } from '#/components/data-table'
import { FormDrawer } from '#/components/form-drawer'
import { Button } from '#/components/ui/button'
import { Label } from '#/components/ui/label'
import { Input } from '#/components/ui/input'
import { UserPlus } from 'lucide-react'
import { toast } from '#/lib/toast'

export const Route = createFileRoute('/_app/users/')({
  component: UserListPage,
})

type User = NonNullable<Awaited<ReturnType<typeof iamClients.user.ListUsers>>['users']>[number]

const columns: ColumnDef<User, unknown>[] = [
  {
    accessorKey: 'username',
    header: '用户名',
    cell: ({ row }) => (
      <Link
        to="/users/$userId"
        params={{ userId: row.original.id ?? '' }}
        className="font-medium text-foreground hover:underline"
      >
        {row.getValue('username')}
      </Link>
    ),
  },
  {
    accessorKey: 'email',
    header: '邮箱',
    cell: ({ row }) => <span className="text-muted-foreground">{row.getValue('email')}</span>,
  },
  {
    accessorKey: 'role',
    header: '角色',
    cell: ({ row }) => <span className="text-muted-foreground">{row.getValue('role')}</span>,
  },
  {
    accessorKey: 'status',
    header: '状态',
    cell: ({ row }) => <span className="text-muted-foreground">{row.getValue('status')}</span>,
  },
]

function UserListPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['users', 'list', page, pageSize],
    queryFn: () => iamClients.user.ListUsers({ pagination: { page: { page, pageSize } } }),
  })

  const users = data?.users ?? []
  const total = data?.pagination?.page?.total ?? 0

  const [createOpen, setCreateOpen] = useState(false)
  const [createLoading, setCreateLoading] = useState(false)
  const [form, setForm] = useState({ username: '', email: '', password: '' })

  function resetForm() {
    setForm({ username: '', email: '', password: '' })
  }

  async function handleCreate() {
    if (!form.username || !form.email || !form.password) return
    setCreateLoading(true)
    try {
      await iamClients.user.CreateUser({
        username: form.username,
        email: form.email,
        password: form.password,
      })
      setCreateOpen(false)
      resetForm()
      void queryClient.invalidateQueries({ queryKey: ['users', 'list'] })
      toast.success('用户已创建')
    } finally {
      setCreateLoading(false)
    }
  }

  return (
    <Page
      title="用户"
      extra={
        <Button onClick={() => setCreateOpen(true)}>
          <UserPlus className="size-4" />
          创建用户
        </Button>
      }
    >
      <DataTable
        columns={columns}
        data={users}
        isLoading={isLoading}
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <FormDrawer
        open={createOpen}
        onOpenChange={(open) => {
          setCreateOpen(open)
          if (!open) resetForm()
        }}
        title="创建用户"
        loading={createLoading}
        onSubmit={handleCreate}
        submitLabel="创建"
      >
        <div className="space-y-2">
          <Label htmlFor="create-user-username">用户名</Label>
          <Input
            id="create-user-username"
            placeholder="输入用户名"
            value={form.username}
            onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="create-user-email">邮箱</Label>
          <Input
            id="create-user-email"
            type="email"
            placeholder="user@example.com"
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="create-user-password">密码</Label>
          <Input
            id="create-user-password"
            type="password"
            placeholder="至少 6 个字符"
            value={form.password}
            onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
          />
        </div>
      </FormDrawer>
    </Page>
  )
}
