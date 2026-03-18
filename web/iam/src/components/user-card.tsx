import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { Button } from '#/components/ui/button'
import { LogOut } from 'lucide-react'
import type { UserInfo } from '#/stores/auth'

interface UserCardProps {
  user: UserInfo | null
  onLogout: () => void
}

export function UserCard({ user, onLogout }: UserCardProps) {
  if (!user) return null

  const initials = user.name
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <div className="mx-3 mb-3 rounded-xl bg-muted p-3 ring-1 ring-border">
      <div className="flex items-center gap-3">
        <Avatar className="size-8">
          <AvatarFallback className="text-xs">{initials}</AvatarFallback>
        </Avatar>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-foreground">{user.name}</p>
          <p className="truncate text-xs text-muted-foreground">{user.email}</p>
        </div>
      </div>
      <Button
        variant="ghost"
        size="sm"
        className="mt-2 w-full justify-start text-muted-foreground hover:text-foreground"
        onClick={onLogout}
      >
        <LogOut className="size-4" />
        退出登录
      </Button>
    </div>
  )
}
