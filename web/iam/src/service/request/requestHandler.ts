import { ofetch } from 'ofetch'
import type { FetchOptions } from 'ofetch'

export type RequestType = {
  path: string
  method: string
  body: string | null
}

export type RequestMeta = {
  service: string
  method: string
}

export type RequestHandler = (
  request: RequestType,
  meta: RequestMeta,
) => Promise<unknown>

export interface TokenStore {
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
  setTokens: (accessToken: string, refreshToken: string) => void
  clear: () => void
}

export interface RequestHandlerOptions {
  baseUrl?: string
  tokenStore?: TokenStore
  contextHeaders?: (meta: RequestMeta) => Record<string, string>
  timeoutMs?: number
  onError?: (error: ApiError, meta: RequestMeta) => void
  autoRefreshToken?: boolean
}

export type ApiErrorKind = 'http' | 'network' | 'timeout'

export class ApiError extends Error {
  readonly kind: ApiErrorKind
  readonly httpStatus?: number
  readonly responseBody?: unknown
  readonly service: string
  readonly method: string

  constructor(opts: {
    kind: ApiErrorKind
    message: string
    httpStatus?: number
    responseBody?: unknown
    service: string
    method: string
    cause?: unknown
  }) {
    super(opts.message, { cause: opts.cause })
    this.name = 'ApiError'
    this.kind = opts.kind
    this.httpStatus = opts.httpStatus
    this.responseBody = opts.responseBody
    this.service = opts.service
    this.method = opts.method
  }
}

function ensureLeadingSlash(path: string): string {
  return path.startsWith('/') ? path : `/${path}`
}

const REFRESH_PATH = '/v1/auth/refresh-token'

export function createRequestHandler(
  options: RequestHandlerOptions = {},
): RequestHandler {
  const {
    baseUrl = '',
    tokenStore,
    contextHeaders,
    timeoutMs = 30_000,
    onError,
    autoRefreshToken = false,
  } = options

  let refreshPromise: Promise<boolean> | null = null

  async function tryRefreshToken(): Promise<boolean> {
    if (!tokenStore) return false
    const refreshToken = tokenStore.getRefreshToken()
    if (!refreshToken) return false

    if (refreshPromise) return refreshPromise

    refreshPromise = (async () => {
      try {
        const data = await ofetch<{
          accessToken: string
          refreshToken: string
        }>(REFRESH_PATH, {
          baseURL: baseUrl,
          method: 'POST',
          body: { refreshToken },
          timeout: timeoutMs,
        })
        tokenStore.setTokens(data.accessToken, data.refreshToken)
        return true
      } catch {
        tokenStore.clear()
        return false
      } finally {
        refreshPromise = null
      }
    })()

    return refreshPromise
  }

  async function doRequest(
    request: RequestType,
    meta: RequestMeta,
  ): Promise<unknown> {
    const headers: Record<string, string> = {
      Accept: 'application/json',
    }

    if (request.body) {
      headers['Content-Type'] = 'application/json'
    }

    if (tokenStore) {
      const token = tokenStore.getAccessToken()
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }
    }

    if (contextHeaders) {
      Object.assign(headers, contextHeaders(meta))
    }

    const fetchOptions: FetchOptions = {
      baseURL: baseUrl,
      method: request.method as FetchOptions['method'],
      headers,
      timeout: timeoutMs,
    }

    if (request.body) {
      fetchOptions.body = request.body
    }

    try {
      return await ofetch(ensureLeadingSlash(request.path), fetchOptions)
    } catch (err: unknown) {
      const fetchError = err as {
        response?: { status: number; _data?: unknown }
        message?: string
      }

      if (fetchError.response) {
        const status = fetchError.response.status
        const body = fetchError.response._data

        const apiErr = new ApiError({
          kind: 'http',
          message: `HTTP ${status} on ${meta.service}.${meta.method}`,
          httpStatus: status,
          responseBody: body,
          service: meta.service,
          method: meta.method,
          cause: err,
        })

        if (
          autoRefreshToken &&
          status === 401 &&
          ensureLeadingSlash(request.path) !== REFRESH_PATH
        ) {
          const refreshed = await tryRefreshToken()
          if (refreshed) {
            return doRequest(request, meta)
          }
        }

        onError?.(apiErr, meta)
        throw apiErr
      }

      const message = (fetchError.message ?? '').toLowerCase()
      const kind: ApiErrorKind = message.includes('timeout')
        ? 'timeout'
        : 'network'

      const apiErr = new ApiError({
        kind,
        message: `${kind} error on ${meta.service}.${meta.method}: ${fetchError.message}`,
        service: meta.service,
        method: meta.method,
        cause: err,
      })
      onError?.(apiErr, meta)
      throw apiErr
    }
  }

  return doRequest
}
