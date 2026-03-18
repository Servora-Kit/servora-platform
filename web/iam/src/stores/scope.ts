import { Store } from '@tanstack/store'

const TENANT_ID_KEY = 'iam.current_tenant_id'
const ORG_ID_KEY = 'iam.current_organization_id'

function hasStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage.getItem === 'function'
}

function readStorage(key: string): string | null {
  if (!hasStorage()) return null
  const value = window.localStorage.getItem(key)
  return value && value.trim().length > 0 ? value : null
}

function writeStorage(key: string, value: string | null): void {
  if (!hasStorage()) return
  if (!value) {
    window.localStorage.removeItem(key)
  } else {
    window.localStorage.setItem(key, value)
  }
}

export interface ScopeState {
  currentTenantId: string | null
  currentOrganizationId: string | null
}

export const scopeStore = new Store<ScopeState>({
  currentTenantId: readStorage(TENANT_ID_KEY),
  currentOrganizationId: readStorage(ORG_ID_KEY),
})

export function setCurrentTenantId(id: string | null): void {
  const trimmed = id?.trim() || null
  if (scopeStore.state.currentTenantId === trimmed) return
  scopeStore.setState((prev) => ({ ...prev, currentTenantId: trimmed }))
  writeStorage(TENANT_ID_KEY, trimmed)
}

export function setCurrentOrganizationId(id: string | null): void {
  const trimmed = id?.trim() || null
  if (scopeStore.state.currentOrganizationId === trimmed) return
  scopeStore.setState((prev) => ({ ...prev, currentOrganizationId: trimmed }))
  writeStorage(ORG_ID_KEY, trimmed)
}

export function clearScope(): void {
  scopeStore.setState(() => ({
    currentTenantId: null,
    currentOrganizationId: null,
  }))
  writeStorage(TENANT_ID_KEY, null)
  writeStorage(ORG_ID_KEY, null)
}

/**
 * Extract organization ID from a URL path matching `/org/{orgId}/...`.
 * Returns null if the path does not match.
 */
export function orgIdFromPath(pathname: string): string | null {
  const match = pathname.match(/^\/org\/([^/]+)(?:\/|$)/)
  if (!match) return null
  const segment = decodeURIComponent(match[1]).trim()
  return segment || null
}
