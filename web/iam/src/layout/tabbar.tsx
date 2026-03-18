import { useRef, useEffect, useCallback } from 'react'
import { useStore } from '@tanstack/react-store'
import { useNavigate } from '@tanstack/react-router'
import { X, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '#/lib/utils'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '#/components/ui/context-menu'
import {
  tabbarStore,
  setActiveTab,
  removeTab,
  removeOtherTabs,
  removeRightTabs,
} from '#/stores/tabbar'

export function Tabbar() {
  const { tabs, activeTab } = useStore(tabbarStore, (s) => s)
  const navigate = useNavigate()
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const activeEl: HTMLElement | null = el.querySelector(`[data-tab-path="${CSS.escape(activeTab)}"]`)
    activeEl?.scrollIntoView({ inline: 'nearest', behavior: 'smooth' })
  }, [activeTab])

  const handleClick = useCallback(
    (path: string) => {
      setActiveTab(path)
      void navigate({ to: path })
    },
    [navigate],
  )

  const handleClose = useCallback(
    (e: React.MouseEvent, path: string) => {
      e.stopPropagation()
      const nextPath = removeTab(path)
      if (nextPath && nextPath !== activeTab) {
        void navigate({ to: nextPath })
      }
    },
    [navigate, activeTab],
  )

  const handleContextClose = useCallback(
    (path: string) => {
      const nextPath = removeTab(path)
      if (nextPath && nextPath !== activeTab) void navigate({ to: nextPath })
    },
    [navigate, activeTab],
  )

  const scroll = useCallback((dir: 'left' | 'right') => {
    scrollRef.current?.scrollBy({ left: dir === 'left' ? -200 : 200, behavior: 'smooth' })
  }, [])

  return (
    <div className="flex h-[38px] shrink-0 items-center border-b border-border bg-background">
      <button
        type="button"
        onClick={() => scroll('left')}
        className="flex h-full shrink-0 items-center border-r border-border px-1.5 text-muted-foreground hover:text-foreground"
      >
        <ChevronLeft className="size-4" />
      </button>

      <div ref={scrollRef} className="flex flex-1 items-end gap-0 overflow-x-auto scrollbar-none">
        {tabs.map((tab) => {
          const active = tab.path === activeTab
          return (
            <ContextMenu key={tab.path}>
              <ContextMenuTrigger asChild>
                <button
                  type="button"
                  data-tab-path={tab.path}
                  onClick={() => handleClick(tab.path)}
                  className={cn(
                    'group relative flex h-[34px] shrink-0 items-center gap-1.5 rounded-t-lg px-3 text-[13px] transition-colors',
                    active
                      ? 'bg-primary/15 text-primary dark:bg-accent dark:text-accent-foreground'
                      : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground',
                  )}
                >
                  <span className="max-w-[120px] truncate">{tab.title}</span>
                  {tab.closable && (
                    <span
                      role="button"
                      tabIndex={-1}
                      onClick={(e) => handleClose(e, tab.path)}
                      onKeyDown={() => {}}
                      className={cn(
                        'flex size-4 items-center justify-center rounded-full transition-colors',
                        active
                          ? 'hover:bg-primary/20 dark:hover:bg-accent-hover'
                          : 'opacity-0 group-hover:opacity-100 hover:bg-muted',
                      )}
                    >
                      <X className="size-3" />
                    </span>
                  )}
                </button>
              </ContextMenuTrigger>
              <ContextMenuContent>
                {tab.closable && (
                  <ContextMenuItem onClick={() => handleContextClose(tab.path)}>
                    关闭当前
                  </ContextMenuItem>
                )}
                <ContextMenuItem onClick={() => removeOtherTabs(tab.path)}>
                  关闭其他
                </ContextMenuItem>
                <ContextMenuItem onClick={() => removeRightTabs(tab.path)}>
                  关闭右侧
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          )
        })}
      </div>

      <button
        type="button"
        onClick={() => scroll('right')}
        className="flex h-full shrink-0 items-center border-l border-border px-1.5 text-muted-foreground hover:text-foreground"
      >
        <ChevronRight className="size-4" />
      </button>
    </div>
  )
}
