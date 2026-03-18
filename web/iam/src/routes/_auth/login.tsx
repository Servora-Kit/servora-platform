import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { iamClients } from '#/api'
import { setTokens, setUser, clearAuth } from '#/stores/auth'
import { setCurrentTenantId } from '#/stores/scope'

export const Route = createFileRoute('/_auth/login')({
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: (search.redirect as string) || '',
  }),
  component: LoginPage,
})

function LoginPage() {
  const navigate = useNavigate()
  const { redirect: redirectTo } = useSearch({ from: '/_auth/login' })
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await iamClients.authn.LoginByEmailPassword({ email, password })
      setTokens(res.accessToken ?? '', res.refreshToken ?? '')

      try {
        const userInfo = await iamClients.user.CurrentUserInfo({})
        setUser({
          id: userInfo.id ?? '',
          name: userInfo.name ?? '',
          email: userInfo.email ?? '',
          role: userInfo.role ?? '',
        })
      } catch {
        // non-critical
      }

      try {
        const tenantsRes = await iamClients.tenant.ListTenants({
          pagination: { page: { page: 1, pageSize: 100 } },
        })
        const tenantIds = (tenantsRes.tenants ?? [])
          .map((t) => t.id)
          .filter(Boolean) as string[]

        if (tenantIds.length === 0) {
          setError('当前账号未加入任何租户，请联系管理员')
          clearAuth()
          return
        }
        setCurrentTenantId(tenantIds[0])
      } catch {
        // non-critical: app-shell will retry on mount
      }

      const target = redirectTo || '/dashboard'
      void navigate({ to: target })
    } catch (err: unknown) {
      const apiErr = err as { responseBody?: { message?: string } }
      setError(apiErr.responseBody?.message ?? '登录失败，请检查邮箱和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">欢迎回来</h1>
        <p className="mt-2 text-muted-foreground">登录到 Servora IAM 管理平台</p>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="space-y-2">
          <Label htmlFor="email">邮箱</Label>
          <Input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
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
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            required
            className="h-10"
          />
        </div>

        <Button type="submit" className="mt-2 h-10 w-full" disabled={loading}>
          {loading ? '登录中...' : '登录'}
        </Button>
      </form>
    </div>
  )
}
