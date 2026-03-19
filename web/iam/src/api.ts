import { env } from '#/env'
import { createIamClients } from '#/service/request/clients'
import type { ApiError, TokenStore } from '#/service/request/clients'
import { authStore, setTokens as storeSetTokens, setLoginExpired, clearAuth } from '#/stores/auth'
import { toast } from '#/lib/toast'

export const tokenStore: TokenStore = {
  getAccessToken() {
    return authStore.state.accessToken
  },
  getRefreshToken() {
    return authStore.state.refreshToken
  },
  setTokens(accessToken: string, refreshToken: string) {
    storeSetTokens(accessToken, refreshToken)
  },
  clear() {
    clearAuth()
  },
}

export const iamClients = createIamClients({
  baseUrl: env.VITE_API_BASE_URL ?? '',
  tokenStore,
  timeoutMs: 30_000,
  autoRefreshToken: true,
  onError(error: ApiError) {
    if (error.httpStatus === 401) {
      setLoginExpired(true)
      return
    }

    if (error.httpStatus === 403) {
      const body = error.responseBody as { reason?: string; message?: string } | null | undefined
      const reason = body?.reason
      if (reason === 'AUTHZ_DENIED') {
        // Genuine permission denial — show a toast.
        const msg = body?.message ?? 'Insufficient permissions'
        toast.error(msg)
        return
      }
    }

    // Global fallback: show API error
    toast.fromApiError(error)
  },
})
