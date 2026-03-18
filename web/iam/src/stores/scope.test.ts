// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest'
import type { Store } from '@tanstack/store'
import type { ScopeState } from './scope'

// Dynamic import to ensure jsdom localStorage is available before module init
let scopeStore: Store<ScopeState>
let setCurrentOrganizationId: (id: string | null) => void
let clearScope: () => void
let orgIdFromPath: (pathname: string) => string | null

const ORG_ID_KEY = 'iam.current_organization_id'

beforeEach(async () => {
  // Ensure jsdom localStorage is set up
  if (typeof window.localStorage.getItem !== 'function') {
    const store: Record<string, string> = {}
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: (key: string) => store[key] ?? null,
        setItem: (key: string, value: string) => { store[key] = value },
        removeItem: (key: string) => { delete store[key] },
        clear: () => { for (const k of Object.keys(store)) delete store[k] },
        get length() { return Object.keys(store).length },
        key: (index: number) => Object.keys(store)[index] ?? null,
      },
      writable: true,
      configurable: true,
    })
  }

  localStorage.clear()

  // Fresh import each time
  const mod = await import('./scope')
  scopeStore = mod.scopeStore
  setCurrentOrganizationId = mod.setCurrentOrganizationId
  clearScope = mod.clearScope
  orgIdFromPath = mod.orgIdFromPath

  clearScope()
})

describe('orgIdFromPath', () => {
  it('extracts org ID from /org/{id}/...', () => {
    expect(orgIdFromPath('/org/abc-123/dashboard')).toBe('abc-123')
  })

  it('extracts org ID from /org/{id} (no trailing slash)', () => {
    expect(orgIdFromPath('/org/my-org')).toBe('my-org')
  })

  it('extracts org ID from /org/{id}/', () => {
    expect(orgIdFromPath('/org/my-org/')).toBe('my-org')
  })

  it('decodes URL-encoded segments', () => {
    expect(orgIdFromPath('/org/my%20org/dashboard')).toBe('my org')
  })

  it('returns null for non-matching paths', () => {
    expect(orgIdFromPath('/applications')).toBeNull()
    expect(orgIdFromPath('/settings/org/123')).toBeNull()
    expect(orgIdFromPath('/')).toBeNull()
  })

  it('returns null for empty segment', () => {
    expect(orgIdFromPath('/org/')).toBeNull()
  })

  it('returns null for whitespace-only segment', () => {
    expect(orgIdFromPath('/org/%20/')).toBeNull()
  })
})

describe('setCurrentOrganizationId', () => {
  it('sets org ID in store and localStorage', () => {
    setCurrentOrganizationId('org-1')
    expect(scopeStore.state.currentOrganizationId).toBe('org-1')
    expect(localStorage.getItem(ORG_ID_KEY)).toBe('org-1')
  })

  it('trims whitespace', () => {
    setCurrentOrganizationId('  org-2  ')
    expect(scopeStore.state.currentOrganizationId).toBe('org-2')
  })

  it('clears on null', () => {
    setCurrentOrganizationId('org-1')
    setCurrentOrganizationId(null)
    expect(scopeStore.state.currentOrganizationId).toBeNull()
    expect(localStorage.getItem(ORG_ID_KEY)).toBeNull()
  })

  it('clears on empty string', () => {
    setCurrentOrganizationId('org-1')
    setCurrentOrganizationId('')
    expect(scopeStore.state.currentOrganizationId).toBeNull()
  })

  it('skips update when value unchanged', () => {
    setCurrentOrganizationId('org-1')
    const snapshot = scopeStore.state
    setCurrentOrganizationId('org-1')
    expect(scopeStore.state).toBe(snapshot)
  })
})

describe('clearScope', () => {
  it('clears org ID', () => {
    setCurrentOrganizationId('org-1')
    clearScope()
    expect(scopeStore.state.currentOrganizationId).toBeNull()
    expect(localStorage.getItem(ORG_ID_KEY)).toBeNull()
  })
})
