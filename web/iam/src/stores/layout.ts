import { Store } from '@tanstack/store'

const SIDEBAR_COLLAPSED_KEY = 'iam.sidebar_collapsed'

function hasStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage.getItem === 'function'
}

export interface LayoutState {
  sidebarCollapsed: boolean
}

export const layoutStore = new Store<LayoutState>({
  sidebarCollapsed: hasStorage()
    ? window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true'
    : false,
})

export function toggleSidebar(): void {
  layoutStore.setState((prev) => {
    const next = !prev.sidebarCollapsed
    if (hasStorage()) window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(next))
    return { ...prev, sidebarCollapsed: next }
  })
}

export function setSidebarCollapsed(collapsed: boolean): void {
  if (layoutStore.state.sidebarCollapsed === collapsed) return
  layoutStore.setState((prev) => ({ ...prev, sidebarCollapsed: collapsed }))
  if (hasStorage()) window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(collapsed))
}

export const SIDEBAR_WIDTH = 224
export const SIDEBAR_COLLAPSED_WIDTH = 48
