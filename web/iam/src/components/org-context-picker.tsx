import { useEffect } from 'react'
import { useStore } from '@tanstack/react-store'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, Building2, Check } from 'lucide-react'
import { Button } from '#/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '#/components/ui/popover'
import { Separator } from '#/components/ui/separator'
import { cn } from '#/lib/utils'
import {
  scopeStore,
  setCurrentOrganizationId,
} from '#/stores/scope'
import { iamClients } from '#/api'

/**
 * Compact org/project context indicator shown in the topbar.
 * Replaces the bottom-sidebar scope switcher so the current scope
 * is always visible and the switch is surfaced in the header.
 */
export function OrgContextPicker() {
  const currentTenantId = useStore(scopeStore, (s) => s.currentTenantId)
  const currentOrgId = useStore(scopeStore, (s) => s.currentOrganizationId)

  const { data: orgsData } = useQuery({
    queryKey: ['organizations', 'org-picker', currentTenantId],
    queryFn: () =>
      iamClients.organization.ListOrganizations({
        pagination: { page: { page: 1, pageSize: 100 } },
      }),
    enabled: !!currentTenantId,
    staleTime: 60_000,
    placeholderData: (prev) => prev,
  })

  const orgs = orgsData?.organizations ?? []
  const currentOrg = orgs.find((o) => o.id === currentOrgId) ?? orgs[0]
  const displayName = currentOrg?.displayName || currentOrg?.name || '未选择组织'

  // Auto-select the first org when none is selected
  useEffect(() => {
    if (orgs.length > 0 && !currentOrgId) {
      setCurrentOrganizationId(orgs[0].id ?? null)
    }
  }, [orgs, currentOrgId])

  function handleSelect(orgId: string) {
    if (orgId === currentOrgId) return
    setCurrentOrganizationId(orgId)
  }

  if (orgs.length === 0) return null

  // Single org: non-interactive label only
  if (orgs.length === 1) {
    return (
      <span className="flex items-center gap-1.5 px-2.5 text-xs text-muted-foreground">
        <Building2 className="size-3.5 shrink-0" />
        <span className="max-w-[120px] truncate">{displayName}</span>
      </span>
    )
  }

  // Multiple orgs: interactive switcher
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 rounded-full px-2.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <Building2 className="size-3.5 shrink-0" />
          <span className="max-w-[120px] truncate">{displayName}</span>
          <ChevronDown className="size-3 shrink-0" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-56 p-1" align="start">
        <p className="px-2 pb-1 pt-1 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
          切换组织
        </p>
        <Separator className="my-1" />
        {orgs.map((org) => {
          const active = org.id === currentOrgId
          return (
            <button
              key={org.id}
              type="button"
              onClick={() => handleSelect(org.id ?? '')}
              className={cn(
                'flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm transition-colors hover:bg-accent',
                active && 'text-primary',
              )}
            >
              <Building2 className="size-4 shrink-0" />
              <span className="flex-1 truncate text-left">
                {org.displayName || org.name}
              </span>
              {active && <Check className="size-4 shrink-0" />}
            </button>
          )
        })}
      </PopoverContent>
    </Popover>
  )
}
