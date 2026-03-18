import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { iamClients } from '#/api'
import { toast } from '#/lib/toast'
import { DataState } from '#/components/data-state'
import { Page } from '#/components/page'
import { ConfirmDialog } from '#/components/confirm-dialog'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Separator } from '#/components/ui/separator'
import { Copy } from 'lucide-react'

export const Route = createFileRoute('/_app/applications/$appId')({
  component: AppDetailPage,
})

function AppDetailPage() {
  const { appId } = Route.useParams()
  const navigate = useNavigate()

  const { data: app, isLoading, isError, refetch } = useQuery({
    queryKey: ['application', appId],
    queryFn: () => iamClients.application.GetApplication({ id: appId }),
  })

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [regenOpen, setRegenOpen] = useState(false)
  const [newSecret, setNewSecret] = useState<string | null>(null)

  async function handleDelete() {
    await toast.promise(
      iamClients.application.DeleteApplication({ id: appId }).then(() => navigate({ to: '/applications' })),
      { loading: '删除中...', success: '应用已删除' },
    )
  }

  async function handleRegenSecret() {
    await toast.promise(
      iamClients.application.RegenerateClientSecret({ id: appId }).then((res) => {
        setNewSecret(res.clientSecret ?? null)
      }),
      { loading: '生成中...', success: '新 Secret 已生成，请及时复制保存' },
    )
  }

  function copyToClipboard(text: string) {
    void navigator.clipboard.writeText(text)
    toast.success('已复制到剪贴板')
  }

  return (
    <DataState isLoading={isLoading} isError={isError} isEmpty={!app} onRetry={() => void refetch()}>
      <Page title={app?.application?.name}>
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">应用配置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center gap-2">
                <span className="w-32 text-muted-foreground">Client ID</span>
                <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{app?.application?.clientId}</code>
                <Button variant="ghost" size="icon-xs" onClick={() => copyToClipboard(app?.application?.clientId ?? '')}>
                  <Copy className="size-3" />
                </Button>
              </div>
              <div className="flex gap-2">
                <span className="w-32 text-muted-foreground">Redirect URIs</span>
                <span className="text-foreground">{app?.application?.redirectUris?.join(', ') || '-'}</span>
              </div>
              <div className="flex gap-2">
                <span className="w-32 text-muted-foreground">Scopes</span>
                <span className="text-foreground">{app?.application?.scopes?.join(', ') || '-'}</span>
              </div>
              <div className="flex gap-2">
                <span className="w-32 text-muted-foreground">Grant Types</span>
                <span className="text-foreground">{app?.application?.grantTypes?.join(', ') || '-'}</span>
              </div>
            </CardContent>
          </Card>

          {newSecret && (
            <Card className="border-warning/50 bg-warning/10 dark:border-warning/30 dark:bg-warning/5">
              <CardContent className="p-4">
                <p className="mb-2 text-sm font-medium text-warning dark:text-warning">新 Client Secret（仅显示一次）</p>
                <div className="flex items-center gap-2">
                  <code className="rounded bg-card px-2 py-1 text-xs">{newSecret}</code>
                  <Button variant="outline" size="xs" onClick={() => copyToClipboard(newSecret)}>
                    <Copy className="size-3" /> 复制
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setRegenOpen(true)}>重新生成 Secret</Button>
          </div>

          <Separator />

          <Card className="border-destructive/50">
            <CardHeader>
              <CardTitle className="text-base text-destructive">危险操作</CardTitle>
            </CardHeader>
            <CardContent>
              <Button variant="destructive" onClick={() => setDeleteOpen(true)}>删除应用</Button>
            </CardContent>
          </Card>
        </div>
      </Page>

      <ConfirmDialog
        open={regenOpen}
        onOpenChange={setRegenOpen}
        title="重新生成 Secret"
        description="此操作将使旧 Secret 失效。确认继续？"
        onConfirm={handleRegenSecret}
        confirmLabel="确认"
      />

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="删除应用"
        description="确认删除此应用？此操作不可撤销。"
        onConfirm={handleDelete}
        destructive
        confirmLabel="删除"
      />
    </DataState>
  )
}
