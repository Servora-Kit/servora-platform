import { useMatches, Link } from '@tanstack/react-router'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '#/components/ui/breadcrumb'
import { Fragment } from 'react'

interface BreadcrumbSegment {
  label: string
  href?: string
}

const ROUTE_LABELS: Record<string, string> = {
  dashboard: '概览',
  organizations: '组织',
  applications: '应用',
  users: '用户',
  tenants: '租户',
  settings: '设置',
  members: '成员',
  profile: '个人信息',
  security: '安全',
}

export function AppBreadcrumb() {
  const matches = useMatches()

  const segments: BreadcrumbSegment[] = []
  for (const match of matches) {
    const routeId = match.routeId
    if (routeId === '__root__') continue

    const parts = routeId.split('/').filter(Boolean)
    const lastPart = parts[parts.length - 1]
    if (!lastPart || lastPart.startsWith('_') || lastPart.startsWith('$')) continue

    const label = ROUTE_LABELS[lastPart] ?? lastPart
    segments.push({ label, href: match.pathname })
  }

  if (segments.length === 0) return null

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {segments.map((seg, idx) => (
          <Fragment key={seg.href ?? idx}>
            {idx > 0 && <BreadcrumbSeparator />}
            <BreadcrumbItem>
              {idx < segments.length - 1 && seg.href ? (
                <BreadcrumbLink asChild>
                  <Link to={seg.href}>{seg.label}</Link>
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage>{seg.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
