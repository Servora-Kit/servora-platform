import { useEffect, useState, useCallback } from 'react'
import { Sun, Moon, SunMoon } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '#/components/ui/tooltip'

type ThemeMode = 'light' | 'dark' | 'auto'

const CYCLE: ThemeMode[] = ['light', 'dark', 'auto']

const LABELS: Record<ThemeMode, string> = {
  light: '浅色模式',
  dark: '深色模式',
  auto: '跟随系统',
}

function getInitialMode(): ThemeMode {
  if (typeof window === 'undefined') return 'auto'
  const stored = window.localStorage.getItem('theme')
  if (stored === 'light' || stored === 'dark' || stored === 'auto') return stored
  return 'auto'
}

function applyThemeMode(mode: ThemeMode) {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  const resolved = mode === 'auto' ? (prefersDark ? 'dark' : 'light') : mode
  document.documentElement.classList.remove('light', 'dark')
  document.documentElement.classList.add(resolved)
  document.documentElement.style.colorScheme = resolved
  if (mode === 'auto') {
    document.documentElement.removeAttribute('data-theme')
  } else {
    document.documentElement.setAttribute('data-theme', mode)
  }
}

export default function ThemeToggle() {
  const [mode, setMode] = useState<ThemeMode>('auto')

  useEffect(() => {
    const init = getInitialMode()
    setMode(init)
    applyThemeMode(init)
  }, [])

  useEffect(() => {
    if (mode !== 'auto') return
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => applyThemeMode('auto')
    media.addEventListener('change', onChange)
    return () => media.removeEventListener('change', onChange)
  }, [mode])

  const handleClick = useCallback(() => {
    const next = CYCLE[(CYCLE.indexOf(mode) + 1) % CYCLE.length]
    setMode(next)
    applyThemeMode(next)
    window.localStorage.setItem('theme', next)
  }, [mode])

  const Icon = mode === 'light' ? Sun : mode === 'dark' ? Moon : SunMoon

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button variant="ghost" size="icon-xs" onClick={handleClick} aria-label={LABELS[mode]}>
          <Icon className="size-4" />
        </Button>
      </TooltipTrigger>
      <TooltipContent>{LABELS[mode]}</TooltipContent>
    </Tooltip>
  )
}
