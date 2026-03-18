import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_auth/callback')({
  component: CallbackPage,
})

function CallbackPage() {
  const navigate = useNavigate()

  useEffect(() => {
    // TODO: Process OAuth callback with authorization code from URL params
    // For now, redirect to login
    void navigate({ to: '/login', search: { redirect: '' } })
  }, [navigate])

  return (
    <div className="flex items-center justify-center p-8">
      <p className="text-sm text-muted-foreground">处理认证回调中...</p>
    </div>
  )
}
