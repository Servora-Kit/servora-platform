import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useLocation } from '@tanstack/react-router'
import { Bell, Maximize2, Minimize2, RotateCw, ShieldAlert } from 'lucide-react'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { Button } from '#/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu'
import { Separator } from '#/components/ui/separator'
import { AppBreadcrumb } from '#/components/app-breadcrumb'
import { OrgContextPicker } from '#/components/org-context-picker'
import ThemeToggle from '#/components/ThemeToggle'
import { useStore } from '@tanstack/react-store'
import { authStore } from '#/stores/auth'
import type { UserInfo } from '#/stores/auth'

interface HeaderProps {
  user: UserInfo | null
  onLogout: () => void
  /** Shown in PlatformShell to let superadmin switch back to tenant view */
  platformMode?: boolean
}

export function Header({ user, onLogout, platformMode = false }: HeaderProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const superAdmin = useStore(authStore, (s) => s.user?.role === 'admin')
  // true when the user is in the platform admin area
  const inPlatform = platformMode || location.pathname.startsWith('/tenants')
  const [isFullscreen, setIsFullscreen] = useState(false)

  useEffect(() => {
    const handler = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', handler)
    return () => document.removeEventListener('fullscreenchange', handler)
  }, [])

  const toggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) {
      void document.exitFullscreen()
    } else {
      void document.documentElement.requestFullscreen()
    }
  }, [])

  const handleRefresh = useCallback(() => {
    void navigate({ to: '.' })
  }, [navigate])

  const initials = user
    ? user.name
        .split(/\s+/)
        .map((w) => w[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : ''

  return (
    <header className="flex h-[50px] shrink-0 items-center justify-between border-b border-border bg-header px-4">
      <div className="flex items-center gap-2">
        <AppBreadcrumb />
        <OrgContextPicker />
      </div>

      <div className="flex items-center gap-1">
        <ThemeToggle />

        <Separator orientation="vertical" className="mx-1 h-5" />

        <Button variant="ghost" size="icon-sm" onClick={handleRefresh} aria-label="刷新">
          <RotateCw className="size-4" />
        </Button>

        <Button variant="ghost" size="icon-sm" onClick={toggleFullscreen} aria-label="全屏">
          {isFullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
        </Button>

        <div className="relative">
          <Button variant="ghost" size="icon-sm" aria-label="通知">
            <Bell className="size-4" />
          </Button>
          <span className="absolute right-0.5 top-0.5 size-2 rounded-full bg-primary" />
        </div>

        <Separator orientation="vertical" className="mx-1 h-5" />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="flex items-center gap-2 rounded-full p-1.5 transition-colors hover:bg-accent"
            >
              <Avatar className="size-8">
                <AvatarFallback className="bg-primary/15 text-primary text-xs font-semibold">
                  {initials}
                </AvatarFallback>
              </Avatar>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[240px] p-0">
            <div className="flex items-center gap-3 p-3">
              <Avatar className="size-12">
                <AvatarFallback className="bg-primary/15 text-primary text-lg font-semibold">
                  {initials}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-foreground">
                  {user?.name ?? '...'}
                </p>
                <p className="truncate text-xs text-muted-foreground">
                  {user?.email ?? ''}
                </p>
              </div>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild className="mx-1 rounded-sm py-1 leading-8">
              <Link to="/settings/profile">个人设置</Link>
            </DropdownMenuItem>
            {superAdmin && (
              <DropdownMenuItem asChild className="mx-1 rounded-sm py-1 leading-8">
                {inPlatform ? (
                  <Link to="/dashboard">
                    <span className="flex items-center gap-2">
                      <ShieldAlert className="size-3.5 text-primary" />
                      切换到租户管理
                    </span>
                  </Link>
                ) : (
                  <Link to="/tenants">
                    <span className="flex items-center gap-2">
                      <ShieldAlert className="size-3.5 text-primary" />
                      切换到平台管理
                    </span>
                  </Link>
                )}
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onLogout} className="mx-1 rounded-sm py-1 leading-8 text-destructive focus:text-destructive">
              退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
