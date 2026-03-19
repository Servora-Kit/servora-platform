import { useMatches, Link } from '@tanstack/react-router'
import { useQueries } from '@tanstack/react-query'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '#/components/ui/breadcrumb'
import { Fragment } from 'react'
import { iamClients } from '#/api'

interface BreadcrumbSegment {
  label: string
  href: string
}

const ROUTE_LABELS: Record<string, string> = {
  dashboard: '概览',
  applications: '应用',
  users: '用户',
  settings: '设置',
  profile: '个人信息',
  security: '安全',
}

function getEntityQueryConfig(paramName: string, paramValue: string) {
  switch (paramName) {
    case 'userId':
      return {
        queryKey: ['user', paramValue] as const,
        queryFn: () => iamClients.user.GetUser({ id: paramValue }),
      }
    case 'appId':
      return {
        queryKey: ['application', paramValue] as const,
        queryFn: () => iamClients.application.GetApplication({ id: paramValue }),
      }
    default:
      return null
  }
}

function extractEntityName(paramName: string, data: unknown): string | undefined {
  if (!data) return undefined
  switch (paramName) {
    case 'userId': {
      const d = data as { user?: { username?: string } }
      return d.user?.username || undefined
    }
    case 'appId': {
      const d = data as { application?: { name?: string } }
      return d.application?.name || undefined
    }
    default:
      return undefined
  }
}

function looksLikeDynamicId(segment: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment)
}

export function AppBreadcrumb() {
  const matches = useMatches()
  const leafMatch = matches.length > 0 ? matches[matches.length - 1] : null
  const pathname = leafMatch?.pathname ?? ''
  const params = (leafMatch?.params ?? {}) as Record<string, string>

  const valueToParamName = new Map(
    Object.entries(params)
      .filter(([, v]) => v && looksLikeDynamicId(v))
      .map(([k, v]) => [v, k]),
  )

  const pathParts = pathname.split('/').filter(Boolean)

  const dynamicSegmentsMap = new Map<string, { paramName: string; paramValue: string }>()
  pathParts.forEach((segment) => {
    const paramName = valueToParamName.get(segment)
    if (paramName && !dynamicSegmentsMap.has(paramName)) {
      dynamicSegmentsMap.set(paramName, { paramName, paramValue: segment })
    }
  })
  const dynamicSegments = [...dynamicSegmentsMap.values()]

  const entityQueries = useQueries({
    queries: dynamicSegments.map(({ paramName, paramValue }) => {
      const config = getEntityQueryConfig(paramName, paramValue)
      if (!config || !paramValue) {
        return {
          queryKey: ['_noop', paramName, paramValue] as const,
          queryFn: (): null => null,
          enabled: false,
        }
      }
      return { ...config, enabled: true, staleTime: 60_000 }
    }),
  })

  const entityNameByValue = new Map<string, string>()
  dynamicSegments.forEach(({ paramName, paramValue }, idx) => {
    const name = extractEntityName(paramName, entityQueries[idx]?.data)
    if (name) {
      entityNameByValue.set(paramValue, name)
    }
  })

  const segments: BreadcrumbSegment[] = []
  let accumulatedPath = ''

  for (const part of pathParts) {
    accumulatedPath += '/' + part
    const isDynamic = valueToParamName.has(part)
    const label = isDynamic
      ? (entityNameByValue.get(part) ?? part.slice(0, 8))
      : (ROUTE_LABELS[part] ?? part)
    segments.push({ label, href: accumulatedPath })
  }

  if (segments.length === 0) return null

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {segments.map((seg, idx) => (
          <Fragment key={seg.href}>
            {idx > 0 && <BreadcrumbSeparator />}
            <BreadcrumbItem>
              {idx < segments.length - 1 ? (
                <BreadcrumbLink asChild>
                  <Link to={seg.href as '/'}>{seg.label}</Link>
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
