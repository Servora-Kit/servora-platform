import { Skeleton } from '#/components/ui/skeleton'
import { Button } from '#/components/ui/button'

interface DataStateProps {
  isLoading: boolean
  isError: boolean
  isEmpty: boolean
  onRetry?: () => void
  loadingText?: string
  emptyText?: string
  errorTitle?: string
  children: React.ReactNode
}

export function DataState({
  isLoading,
  isError,
  isEmpty,
  onRetry,
  loadingText = '加载中...',
  emptyText = '暂无数据。',
  errorTitle = '加载失败',
  children,
}: DataStateProps) {
  if (isLoading) {
    return (
      <div className="space-y-2 rounded border bg-card p-4">
        <p className="text-sm text-muted-foreground">{loadingText}</p>
        <Skeleton className="h-9 w-full" />
        <Skeleton className="h-9 w-full" />
        <Skeleton className="h-9 w-full" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="rounded border bg-card p-4">
        <p className="text-sm font-medium text-destructive">{errorTitle}</p>
        {onRetry && (
          <Button variant="outline" size="sm" className="mt-2" onClick={onRetry}>
            重试
          </Button>
        )}
      </div>
    )
  }

  if (isEmpty) {
    return (
      <div className="rounded border bg-card p-4 text-sm text-muted-foreground">
        {emptyText}
      </div>
    )
  }

  return <>{children}</>
}
