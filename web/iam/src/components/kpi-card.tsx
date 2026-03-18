import { Link } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Skeleton } from '#/components/ui/skeleton'
import type { LucideIcon } from 'lucide-react'

interface KpiCardProps {
  title: string
  value: number | undefined
  icon: LucideIcon
  href: string
  isLoading: boolean
}

export function KpiCard({ title, value, icon: Icon, href, isLoading }: KpiCardProps) {
  return (
    <Link to={href} className="block">
      <Card className="cursor-pointer transition-colors hover:bg-accent/50">
        <CardContent className="flex items-center gap-4 p-4">
          <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
            <Icon className="size-5 text-muted-foreground" />
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            {isLoading ? (
              <Skeleton className="mt-1 h-7 w-16" />
            ) : (
              <p className="text-2xl font-semibold">{value ?? 0}</p>
            )}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
