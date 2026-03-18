import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-store'
import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { iamClients } from '#/api'
import { scopeStore } from '#/stores/scope'
import { authStore } from '#/stores/auth'
import { Page } from '#/components/page'
import { DataTable } from '#/components/data-table'
import { FormDrawer } from '#/components/form-drawer'
import { ConfirmDialog } from '#/components/confirm-dialog'
import { Button } from '#/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Label } from '#/components/ui/label'
import { Badge } from '#/components/ui/badge'
import { UserPlus, Trash2, ArrowRightLeft } from 'lucide-react'
import { toast } from '#/lib/toast'

export const Route = createFileRoute('/_app/organizations/$orgId/members')({
  component: OrgMembersPage,
})

// owner 只能通过专用的所有权转让操作获得，不在下拉选项中显示
const ASSIGNABLE_ROLES = ['admin', 'member', 'viewer']
const PAGE_SIZE = 50

interface Member {
  userId?: string
  userName?: string
  userEmail?: string
  role?: string
}

function roleBadgeVariant(role: string): 'default' | 'secondary' | 'outline' {
  if (role === 'owner') return 'default'
  if (role === 'admin') return 'secondary'
  return 'outline'
}

function OrgMembersPage() {
  const { orgId } = Route.useParams()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['org-members', orgId, page],
    queryFn: () =>
      iamClients.organization.ListMembers({
        organizationId: orgId,
        pagination: { page: { page, pageSize: PAGE_SIZE } },
      }),
  })

  const members = data?.members ?? []
  const total = data?.pagination?.page?.total ?? 0

  const currentTenantId = useStore(scopeStore, (s) => s.currentTenantId)
  const currentUserId = useStore(authStore, (s) => s.user?.id)

  const isCurrentUserOwner = members.some(
    (m) => m.userId === currentUserId && m.role === 'owner',
  )

  const [roleChange, setRoleChange] = useState<{
    userId: string
    name: string
    oldRole: string
    newRole: string
  } | null>(null)
  const [removeTarget, setRemoveTarget] = useState<{ userId: string; name: string } | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [inviteUserId, setInviteUserId] = useState<string | undefined>(undefined)
  const [inviteRole, setInviteRole] = useState('member')
  const [inviteLoading, setInviteLoading] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferTargetId, setTransferTargetId] = useState<string | undefined>(undefined)
  const [transferLoading, setTransferLoading] = useState(false)

  const { data: tenantMembersData } = useQuery({
    queryKey: ['tenant-members', 'list-for-invite', currentTenantId],
    queryFn: () =>
      iamClients.tenant.ListMembers({
        tenantId: currentTenantId!,
        pagination: { page: { page: 1, pageSize: 100 } },
      }),
    enabled: !!currentTenantId,
    staleTime: 60_000,
  })
  const userOptions = tenantMembersData?.members ?? []

  // 可被转让所有权的目标：必须是 admin
  const adminMembers = members.filter(
    (m) => m.role === 'admin' && m.userId !== currentUserId,
  )

  function invalidate() {
    void queryClient.invalidateQueries({ queryKey: ['org-members', orgId] })
  }

  async function handleRoleConfirm() {
    if (!roleChange) return
    const change = roleChange
    setRoleChange(null)
    await toast.promise(
      iamClients.organization
        .UpdateMemberRole({
          organizationId: orgId,
          userId: change.userId,
          role: change.newRole,
        })
        .then(() => invalidate()),
      { loading: '更新角色...', success: `已将 ${change.name} 的角色改为 ${change.newRole}` },
    )
  }

  async function handleRemoveConfirm() {
    if (!removeTarget) return
    const target = removeTarget
    setRemoveTarget(null)
    await toast.promise(
      iamClients.organization
        .RemoveMember({
          organizationId: orgId,
          userId: target.userId,
        })
        .then(() => invalidate()),
      { loading: '移除中...', success: `已移除成员「${target.name}」` },
    )
  }

  async function handleInvite() {
    if (!inviteUserId) return
    setInviteLoading(true)
    try {
      await iamClients.organization.AddMember({
        organizationId: orgId,
        userId: inviteUserId,
        role: inviteRole,
      })
      setInviteOpen(false)
      setInviteUserId(undefined)
      setInviteRole('member')
      invalidate()
      toast.success('成员已添加')
    } finally {
      setInviteLoading(false)
    }
  }

  async function handleTransferOwnership() {
    if (!transferTargetId) return
    setTransferLoading(true)
    try {
      await iamClients.organization.TransferOwnership({
        organizationId: orgId,
        newOwnerUserId: transferTargetId,
      })
      setTransferOpen(false)
      setTransferTargetId(undefined)
      invalidate()
      toast.success('所有权已转让')
    } finally {
      setTransferLoading(false)
    }
  }

  const columns: ColumnDef<Member, unknown>[] = [
    {
      accessorKey: 'userName',
      header: '用户',
      cell: ({ row }) => (
        <span className="font-medium text-foreground">{row.original.userName ?? '-'}</span>
      ),
    },
    {
      accessorKey: 'userEmail',
      header: '邮箱',
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.userEmail ?? '-'}</span>
      ),
    },
    {
      accessorKey: 'role',
      header: '角色',
      cell: ({ row }) => {
        const m = row.original
        const isOwner = m.role === 'owner'

        if (isOwner) {
          return <Badge variant={roleBadgeVariant('owner')}>owner</Badge>
        }

        return (
          <Select
            value={m.role ?? 'member'}
            onValueChange={(v) =>
              setRoleChange({
                userId: m.userId ?? '',
                name: m.userName ?? '',
                oldRole: m.role ?? '',
                newRole: v,
              })
            }
          >
            <SelectTrigger className="h-7 w-28 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ASSIGNABLE_ROLES.map((r) => (
                <SelectItem key={r} value={r}>
                  {r}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )
      },
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const m = row.original
        const isOwner = m.role === 'owner'

        if (isOwner) {
          return <span className="text-xs text-muted-foreground">—</span>
        }

        return (
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() =>
              setRemoveTarget({ userId: m.userId ?? '', name: m.userName ?? '' })
            }
          >
            <Trash2 className="size-3.5 text-destructive" />
          </Button>
        )
      },
    },
  ]

  return (
    <Page
      title="成员管理"
      description="管理组织成员。"
      extra={
        <div className="flex gap-2">
          {isCurrentUserOwner && (
            <Button
              variant="outline"
              onClick={() => setTransferOpen(true)}
              disabled={adminMembers.length === 0}
              title={adminMembers.length === 0 ? '需要先有 admin 才能转让所有权' : ''}
            >
              <ArrowRightLeft className="size-4" />
              转让所有权
            </Button>
          )}
          <Button onClick={() => setInviteOpen(true)}>
            <UserPlus className="size-4" />
            添加成员
          </Button>
        </div>
      }
    >
      <DataTable
        columns={columns}
        data={members}
        isLoading={isLoading}
        page={page}
        pageSize={PAGE_SIZE}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={() => {}}
      />

      <FormDrawer
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        title="添加成员"
        loading={inviteLoading}
        onSubmit={handleInvite}
        submitLabel="添加"
      >
        <div className="space-y-2">
          <Label>选择用户</Label>
          <Select value={inviteUserId} onValueChange={setInviteUserId}>
            <SelectTrigger>
              <SelectValue placeholder="从租户成员中选择" />
            </SelectTrigger>
            <SelectContent>
              {userOptions.length === 0 && (
                <div className="py-4 text-center text-xs text-muted-foreground">暂无可选用户</div>
              )}
              {userOptions.map((u) => (
                <SelectItem key={u.userId} value={u.userId ?? ''}>
                  <span className="font-medium">{u.userName || u.userId}</span>
                  {u.userEmail && (
                    <span className="ml-2 text-xs text-muted-foreground">{u.userEmail}</span>
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>角色</Label>
          <Select value={inviteRole} onValueChange={setInviteRole}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ASSIGNABLE_ROLES.map((r) => (
                <SelectItem key={r} value={r}>
                  {r}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </FormDrawer>

      <FormDrawer
        open={transferOpen}
        onOpenChange={setTransferOpen}
        title="转让组织所有权"
        loading={transferLoading}
        onSubmit={handleTransferOwnership}
        submitLabel="确认转让"
      >
        <p className="text-sm text-muted-foreground">
          转让后你将降为 admin，新的 owner 将获得完整控制权。此操作立即生效且不可撤销。
        </p>
        <div className="space-y-2">
          <Label>选择新的 owner（必须是现有 admin）</Label>
          <Select value={transferTargetId} onValueChange={setTransferTargetId}>
            <SelectTrigger>
              <SelectValue placeholder="选择目标 admin" />
            </SelectTrigger>
            <SelectContent>
              {adminMembers.map((m) => (
                <SelectItem key={m.userId} value={m.userId ?? ''}>
                  <span className="font-medium">{m.userName || m.userId}</span>
                  {m.userEmail && (
                    <span className="ml-2 text-xs text-muted-foreground">{m.userEmail}</span>
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </FormDrawer>

      <ConfirmDialog
        open={!!roleChange}
        onOpenChange={(open) => {
          if (!open) setRoleChange(null)
        }}
        title="确认角色变更"
        description={
          roleChange
            ? `确认将 ${roleChange.name} 的角色从 ${roleChange.oldRole} 改为 ${roleChange.newRole}？`
            : ''
        }
        onConfirm={handleRoleConfirm}
      />

      <ConfirmDialog
        open={!!removeTarget}
        onOpenChange={(open) => {
          if (!open) setRemoveTarget(null)
        }}
        title="移除成员"
        description={
          removeTarget ? `确认将 ${removeTarget.name} 从组织中移除？此操作不可撤销。` : ''
        }
        onConfirm={handleRemoveConfirm}
        destructive
        confirmLabel="移除"
      />
    </Page>
  )
}
