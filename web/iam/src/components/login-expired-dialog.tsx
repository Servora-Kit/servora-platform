import { useState } from 'react'
import { useStore } from '@tanstack/react-store'
import { useQueryClient } from '@tanstack/react-query'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { iamClients } from '#/api'
import { authStore, setTokens, setLoginExpired, clearAuth } from '#/stores/auth'
import { useNavigate } from '@tanstack/react-router'

export function LoginExpiredDialog() {
  const loginExpired = useStore(authStore, (s) => s.loginExpired)
  const queryClient = useQueryClient()
  const navigate = useNavigate()
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
      if (res.accessToken && res.refreshToken) {
        setTokens(res.accessToken, res.refreshToken)
      }
      setLoginExpired(false)
      setEmail('')
      setPassword('')
      void queryClient.invalidateQueries()
    } catch {
      setError('登录失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  function handleLogout() {
    clearAuth()
    setLoginExpired(false)
    void navigate({ to: '/login' as string })
  }

  return (
    <Dialog open={loginExpired} onOpenChange={() => {}}>
      <DialogContent className="sm:max-w-md" onPointerDownOutside={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle>登录已过期</DialogTitle>
          <DialogDescription>请重新输入凭据以继续操作</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="expired-email">邮箱</Label>
            <Input
              id="expired-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="expired-password">密码</Label>
            <Input
              id="expired-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <div className="flex gap-2">
            <Button type="submit" className="flex-1" disabled={loading}>
              {loading ? '登录中...' : '重新登录'}
            </Button>
            <Button type="button" variant="outline" onClick={handleLogout}>
              退出
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
