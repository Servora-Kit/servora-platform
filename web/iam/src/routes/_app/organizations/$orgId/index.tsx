import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { iamClients } from '#/api'
import { DataState } from '#/components/data-state'
import { Page } from '#/components/page'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'

export const Route = createFileRoute('/_app/organizations/$orgId/')({
  component: OrgDetailPage,
})

function OrgDetailPage() {
  const { orgId } = Route.useParams()

  const { data: org, isLoading, isError, refetch } = useQuery({
    queryKey: ['organization', orgId],
    queryFn: () => iamClients.organization.GetOrganization({ id: orgId }),
  })

  return (
    <DataState isLoading={isLoading} isError={isError} isEmpty={!org} onRetry={() => void refetch()}>
      <Page
        title={org?.organization?.displayName || org?.organization?.name}
        extra={
          <>
            <Button variant="outline" size="sm" asChild>
              <Link to="/organizations/$orgId/members" params={{ orgId }}>
                成员管理
              </Link>
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link to="/organizations/$orgId/settings" params={{ orgId }}>
                设置
              </Link>
            </Button>
          </>
        }
      >
        <Card>
          <CardHeader>
            <CardTitle className="text-base">基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex gap-2">
              <span className="w-20 text-muted-foreground">名称</span>
              <span className="text-foreground">{org?.organization?.displayName || org?.organization?.name}</span>
            </div>
            <div className="flex gap-2">
              <span className="w-20 text-muted-foreground">创建时间</span>
              <span className="text-foreground">
                {org?.organization?.createdAt ? new Date(org.organization.createdAt).toLocaleString('zh-CN') : '-'}
              </span>
            </div>
          </CardContent>
        </Card>
      </Page>
    </DataState>
  )
}
