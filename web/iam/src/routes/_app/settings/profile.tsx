import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { iamClients } from '#/api'
import { setUser } from '#/stores/auth'
import { toast } from '#/lib/toast'
import { DataState } from '#/components/data-state'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'

export const Route = createFileRoute('/_app/settings/profile')({
  component: ProfilePage,
})

function ProfilePage() {
  const { data: userInfo, isLoading, isError, refetch } = useQuery({
    queryKey: ['current-user'],
    queryFn: () => iamClients.user.CurrentUserInfo({}),
  })

  const [username, setUsername] = useState('')
  const [saving, setSaving] = useState(false)

  const currentUser = userInfo?.user

  useEffect(() => {
    if (currentUser?.username) setUsername(currentUser.username)
  }, [currentUser?.username])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    try {
      await iamClients.user.UpdateUser({
        id: currentUser?.id ?? '',
        data: {
          id: currentUser?.id ?? '',
          username,
          email: currentUser?.email ?? '',
          role: currentUser?.role ?? '',
          status: currentUser?.status ?? '',
          emailVerified: currentUser?.emailVerified ?? false,
          phone: currentUser?.phone ?? '',
          phoneVerified: currentUser?.phoneVerified ?? false,
          profile: currentUser?.profile,
          emailVerifiedAt: currentUser?.emailVerifiedAt,
          createdAt: currentUser?.createdAt,
          updatedAt: currentUser?.updatedAt,
        },
      })
      setUser({
        id: currentUser?.id ?? '',
        name: username,
        email: currentUser?.email ?? '',
        role: currentUser?.role ?? '',
      })
      void refetch()
      toast.success('保存成功')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">个人信息</h1>
      <DataState isLoading={isLoading} isError={isError} isEmpty={!userInfo} onRetry={() => void refetch()}>
        <Card>
          <CardHeader><CardTitle className="text-base">编辑资料</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label>邮箱</Label>
                <Input value={currentUser?.email ?? ''} disabled />
              </div>
              <div className="space-y-2">
                <Label>用户名</Label>
                <Input value={username} onChange={(e) => setUsername(e.target.value)} required />
              </div>
              <Button type="submit" disabled={saving}>{saving ? '保存中...' : '保存'}</Button>
            </form>
          </CardContent>
        </Card>
      </DataState>
    </div>
  )
}
