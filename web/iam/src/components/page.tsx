import { cn } from '#/lib/utils'

interface PageProps {
  title?: React.ReactNode
  description?: React.ReactNode
  extra?: React.ReactNode
  footer?: React.ReactNode
  contentClass?: string
  children: React.ReactNode
}

export function Page({ title, description, extra, footer, contentClass, children }: PageProps) {
  const hasHeader = title || description || extra
  return (
    <div className="relative flex min-h-full flex-col">
      {hasHeader && (
        <div className="flex items-start justify-between gap-4 rounded-t-md bg-card border border-border px-6 py-4">
          <div className="flex-auto">
            {title && <h1 className="text-xl font-semibold text-foreground">{title}</h1>}
            {description && (
              <p className="mt-1 text-sm text-muted-foreground">{description}</p>
            )}
          </div>
          {extra && <div className="flex shrink-0 items-center gap-2">{extra}</div>}
        </div>
      )}
      <div className={cn('flex-1 p-4', hasHeader && 'rounded-b-md bg-card border border-t-0 border-border', !hasHeader && 'rounded-md bg-card border border-border', contentClass)}>
        {children}
      </div>
      {footer && (
        <div className="mt-4 flex items-center rounded-md bg-card border border-border px-6 py-4">
          {footer}
        </div>
      )}
    </div>
  )
}
