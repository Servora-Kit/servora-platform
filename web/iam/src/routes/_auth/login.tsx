import {
  createFileRoute,
  Link,
  useNavigate,
  useSearch,
} from '@tanstack/react-router'
import { useReducer } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { iamClients } from '#/api'
import { clearAuth, setTokens, setUser } from '#/stores/auth'
import { useResendVerification } from '#/hooks/use-resend-verification'
import type { ApiError } from '@servora/web-pkg/request'
import { isKratosReason, kratosMessage } from '@servora/web-pkg/errors'

export const Route = createFileRoute('/_auth/login')({
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: (search.redirect as string) || '',
    authRequestID: (search.authRequestID as string) || '',
  }),
  component: LoginPage,
})

// ---------- State / Reducer ----------

type Status = 'idle' | 'submitting'

interface LoginState {
  email: string
  password: string
  status: Status
  error: string
  emailNotVerified: boolean
}

type LoginAction =
  | { type: 'SET_FIELD'; field: 'email' | 'password'; value: string }
  | { type: 'SUBMIT' }
  | { type: 'SUBMIT_ERROR'; error: string }
  | { type: 'EMAIL_NOT_VERIFIED' }
  | { type: 'RESET_NOTIFICATION' }

const initialState: LoginState = {
  email: '',
  password: '',
  status: 'idle',
  error: '',
  emailNotVerified: false,
}

function reducer(state: LoginState, action: LoginAction): LoginState {
  switch (action.type) {
    case 'SET_FIELD':
      return { ...state, [action.field]: action.value }
    case 'SUBMIT':
      return {
        ...state,
        status: 'submitting',
        error: '',
        emailNotVerified: false,
      }
    case 'SUBMIT_ERROR':
      return { ...state, status: 'idle', error: action.error }
    case 'EMAIL_NOT_VERIFIED':
      return { ...state, status: 'idle', error: '', emailNotVerified: true }
    case 'RESET_NOTIFICATION':
      return { ...state, error: '', emailNotVerified: false }
    default:
      return state
  }
}

// ---------- Component ----------

function LoginPage() {
  const navigate = useNavigate()
  const { redirect: redirectTo, authRequestID } = useSearch({
    from: '/_auth/login',
  })
  const [state, dispatch] = useReducer(reducer, initialState)
  const { resend, resending, message: resendMsg } = useResendVerification(
    state.email,
  )

  const isOidcFlow = Boolean(authRequestID)
  const loading = state.status === 'submitting'

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    dispatch({ type: 'SUBMIT' })

    try {
      const res = await iamClients.authn.LoginByEmailPassword({
        email: state.email,
        password: state.password,
      })

      if (isOidcFlow) {
        // OIDC 流程：调后端 complete，由后端 302 到 callbackURL
        const completeRes = await fetch('/login/complete', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            authRequestID,
            accessToken: res.accessToken,
          }),
        })
        if (!completeRes.ok) {
          const body = (await completeRes.json()) as { message?: string }
          dispatch({
            type: 'SUBMIT_ERROR',
            error: body.message ?? '授权流程失败，请重试',
          })
          return
        }
        const { callbackURL } = (await completeRes.json()) as {
          callbackURL: string
        }
        window.location.href = callbackURL
        return
      }

      // 管理端流程：存 JWT + 校验 admin 角色
      setTokens(res.accessToken ?? '', res.refreshToken ?? '')

      let role = ''
      try {
        const userInfo = await iamClients.user.CurrentUserInfo({})
        const currentUser = userInfo.user
        role = currentUser?.role ?? ''
        setUser({
          id: currentUser?.id ?? '',
          name: currentUser?.username ?? currentUser?.email ?? '',
          email: currentUser?.email ?? '',
          role,
        })
      } catch {
        // non-critical
      }

      if (role !== 'admin') {
        clearAuth()
        dispatch({
          type: 'SUBMIT_ERROR',
          error: '此账号无权访问 IAM 管理平台，请使用管理员账号登录',
        })
        return
      }

      void navigate({ to: redirectTo || '/dashboard' })
    } catch (err: unknown) {
      const apiErr = err as ApiError
      if (isKratosReason(apiErr, 'EMAIL_NOT_VERIFIED')) {
        dispatch({ type: 'EMAIL_NOT_VERIFIED' })
      } else {
        dispatch({
          type: 'SUBMIT_ERROR',
          error: kratosMessage(apiErr, '登录失败，请检查邮箱和密码'),
        })
      }
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">欢迎回来</h1>
        <p className="mt-2 text-muted-foreground">
          {isOidcFlow
            ? '请登录以继续授权'
            : '登录到 Servora IAM 管理平台'}
        </p>
      </div>

      {state.error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {state.error}
        </div>
      )}

      {state.emailNotVerified && (
        <div className="rounded-lg border border-amber-500/50 bg-amber-500/10 px-4 py-3 text-sm text-amber-700 dark:text-amber-400">
          <p className="font-medium">邮箱尚未验证</p>
          <p className="mt-1">请先验证邮箱后再登录。</p>
          {resendMsg && (
            <p className="mt-2 text-muted-foreground">{resendMsg}</p>
          )}
          <button
            type="button"
            onClick={resend}
            disabled={resending}
            className="mt-2 text-primary hover:underline disabled:opacity-50"
          >
            {resending ? '发送中...' : '重新发送验证邮件'}
          </button>
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="space-y-2">
          <Label htmlFor="email">邮箱</Label>
          <Input
            id="email"
            type="email"
            value={state.email}
            onChange={(e) =>
              dispatch({ type: 'SET_FIELD', field: 'email', value: e.target.value })
            }
            placeholder="you@example.com"
            required
            autoFocus
            className="h-10"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">密码</Label>
          <Input
            id="password"
            type="password"
            value={state.password}
            onChange={(e) =>
              dispatch({
                type: 'SET_FIELD',
                field: 'password',
                value: e.target.value,
              })
            }
            placeholder="••••••••"
            required
            className="h-10"
          />
        </div>

        <Button type="submit" className="mt-2 h-10 w-full" disabled={loading}>
          {loading ? '登录中...' : '登录'}
        </Button>
      </form>

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>
          没有账号？{' '}
          <Link
            to="/register"
            className="font-medium text-primary hover:underline"
          >
            立即注册
          </Link>
        </span>
        <Link
          to="/reset-password"
          search={{ token: '' }}
          className="font-medium text-primary hover:underline"
        >
          忘记密码？
        </Link>
      </div>
    </div>
  )
}
