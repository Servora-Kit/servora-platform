import { Store } from '@tanstack/store'

export interface TabItem {
  path: string
  title: string
  closable: boolean
}

const DEFAULT_TAB: TabItem = { path: '/dashboard', title: '概览', closable: false }

export interface TabbarState {
  tabs: TabItem[]
  activeTab: string
}

export const tabbarStore = new Store<TabbarState>({
  tabs: [DEFAULT_TAB],
  activeTab: DEFAULT_TAB.path,
})

export function addTab(tab: Omit<TabItem, 'closable'> & { closable?: boolean }): void {
  tabbarStore.setState((prev) => {
    const exists = prev.tabs.some((t) => t.path === tab.path)
    const nextTabs = exists
      ? prev.tabs
      : [...prev.tabs, { ...tab, closable: tab.closable ?? true }]
    return { tabs: nextTabs, activeTab: tab.path }
  })
}

export function setActiveTab(path: string): void {
  if (tabbarStore.state.activeTab === path) return
  tabbarStore.setState((prev) => ({ ...prev, activeTab: path }))
}

export function removeTab(path: string): string | null {
  const { tabs, activeTab } = tabbarStore.state
  const tab = tabs.find((t) => t.path === path)
  if (!tab || !tab.closable) return null

  const idx = tabs.indexOf(tab)
  const nextTabs = tabs.filter((t) => t.path !== path)

  let nextActive = activeTab
  if (activeTab === path) {
    nextActive = nextTabs[Math.min(idx, nextTabs.length - 1)]?.path ?? '/dashboard'
  }
  tabbarStore.setState(() => ({ tabs: nextTabs, activeTab: nextActive }))
  return nextActive
}

export function removeOtherTabs(path: string): void {
  tabbarStore.setState((prev) => ({
    tabs: prev.tabs.filter((t) => !t.closable || t.path === path),
    activeTab: path,
  }))
}

export function resetTabs(defaultTab: Omit<TabItem, 'closable'>): void {
  tabbarStore.setState(() => ({
    tabs: [{ ...defaultTab, closable: false }],
    activeTab: defaultTab.path,
  }))
}

export function removeRightTabs(path: string): void {
  tabbarStore.setState((prev) => {
    const idx = prev.tabs.findIndex((t) => t.path === path)
    if (idx < 0) return prev
    const kept = prev.tabs.filter((t, i) => i <= idx || !t.closable)
    const nextActive = kept.some((t) => t.path === prev.activeTab) ? prev.activeTab : path
    return { tabs: kept, activeTab: nextActive }
  })
}
