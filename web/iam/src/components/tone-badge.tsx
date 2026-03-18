import { cn } from '#/lib/utils'

type Tone = 'green' | 'yellow' | 'red' | 'zinc'

const toneStyles: Record<Tone, string> = {
  green: 'border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-400',
  yellow: 'border-yellow-200 bg-yellow-50 text-yellow-800 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-400',
  red: 'border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400',
  zinc: 'border-border bg-muted text-muted-foreground',
}

interface ToneBadgeProps {
  tone: Tone
  children: React.ReactNode
  className?: string
}

export function ToneBadge({ tone, children, className }: ToneBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium',
        toneStyles[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}

export function statusTone(status: string): Tone {
  switch (status) {
    case 'active':
    case 'accepted':
      return 'green'
    case 'pending':
    case 'invited':
      return 'yellow'
    case 'rejected':
    case 'deleted':
      return 'red'
    default:
      return 'zinc'
  }
}
