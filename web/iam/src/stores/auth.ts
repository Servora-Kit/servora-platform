import { Store } from '@tanstack/store'

const ACCESS_TOKEN_KEY = 'iam_access_token'
const REFRESH_TOKEN_KEY = 'iam_refresh_token'
const USER_KEY = 'iam_user'

function hasStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage?.getItem === 'function'
}

function readStorage(key: string): string | null {
  if (!hasStorage()) return null
  return window.localStorage.getItem(key)
}

function writeStorage(key: string, value: string | null): void {
  if (!hasStorage()) return
  if (value === null) {
    window.localStorage.removeItem(key)
  } else {
    window.localStorage.setItem(key, value)
  }
}

export interface UserInfo {
  id: string
  name: string
  email: string
  role: string
}

export interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  user: UserInfo | null
  loginExpired: boolean
}

function readUser(): UserInfo | null {
  const raw = readStorage(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserInfo
  } catch {
    return null
  }
}

export const authStore = new Store<AuthState>({
  accessToken: readStorage(ACCESS_TOKEN_KEY),
  refreshToken: readStorage(REFRESH_TOKEN_KEY),
  user: readUser(),
  loginExpired: false,
})

export function setTokens(accessToken: string, refreshToken: string): void {
  const prev = authStore.state
  if (prev.accessToken === accessToken && prev.refreshToken === refreshToken) return
  writeStorage(ACCESS_TOKEN_KEY, accessToken)
  writeStorage(REFRESH_TOKEN_KEY, refreshToken)
  authStore.setState((s) => ({
    ...s,
    accessToken,
    refreshToken,
    loginExpired: false,
  }))
}

export function clearAuth(): void {
  writeStorage(ACCESS_TOKEN_KEY, null)
  writeStorage(REFRESH_TOKEN_KEY, null)
  writeStorage(USER_KEY, null)
  authStore.setState(() => ({
    accessToken: null,
    refreshToken: null,
    user: null,
    loginExpired: false,
  }))
}

export function setUser(user: UserInfo): void {
  writeStorage(USER_KEY, JSON.stringify(user))
  authStore.setState((prev) => ({ ...prev, user }))
}

export function setLoginExpired(expired: boolean): void {
  authStore.setState((prev) => ({ ...prev, loginExpired: expired }))
}

export function isAuthenticated(): boolean {
  return authStore.state.accessToken !== null
}

export function isSuperAdmin(): boolean {
  return authStore.state.user?.role === 'admin'
}
