import { Store } from '@tanstack/store'

const ORG_ID_KEY = 'iam.current_organization_id'
const PROJECT_ID_KEY = 'iam.current_project_id'

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
  currentOrganizationId: string | null
  currentProjectId: string | null
}

export const scopeStore = new Store<ScopeState>({
  currentOrganizationId: readStorage(ORG_ID_KEY),
  currentProjectId: readStorage(PROJECT_ID_KEY),
})

export function setCurrentOrganizationId(id: string | null): void {
  const trimmed = id?.trim() || null
  if (scopeStore.state.currentOrganizationId === trimmed) return
  scopeStore.setState((prev) => ({ ...prev, currentOrganizationId: trimmed }))
  writeStorage(ORG_ID_KEY, trimmed)
}

export function setCurrentProjectId(id: string | null): void {
  const trimmed = id?.trim() || null
  if (scopeStore.state.currentProjectId === trimmed) return
  scopeStore.setState((prev) => ({ ...prev, currentProjectId: trimmed }))
  writeStorage(PROJECT_ID_KEY, trimmed)
}

export function clearScope(): void {
  scopeStore.setState(() => ({
    currentOrganizationId: null,
    currentProjectId: null,
  }))
  writeStorage(ORG_ID_KEY, null)
  writeStorage(PROJECT_ID_KEY, null)
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
