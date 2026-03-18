import { useState, useCallback, useMemo } from 'react'

interface CursorPaginationResult {
  currentCursor: string
  currentPage: number
  goNext: (nextCursor: string) => void
  goPrev: () => void
  reset: () => void
  hasPrev: boolean
}

export function useCursorPagination(): CursorPaginationResult {
  const [cursorHistory, setCursorHistory] = useState<string[]>([''])
  const [currentIndex, setCurrentIndex] = useState(0)

  const currentCursor = cursorHistory[currentIndex] ?? ''
  const hasPrev = currentIndex > 0

  const goNext = useCallback(
    (nextCursor: string) => {
      setCursorHistory((prev) => [...prev.slice(0, currentIndex + 1), nextCursor])
      setCurrentIndex((prev) => prev + 1)
    },
    [currentIndex],
  )

  const goPrev = useCallback(() => {
    if (currentIndex > 0) {
      setCurrentIndex((prev) => prev - 1)
    }
  }, [currentIndex])

  const reset = useCallback(() => {
    setCursorHistory([''])
    setCurrentIndex(0)
  }, [])

  return useMemo(
    () => ({
      currentCursor,
      currentPage: currentIndex + 1,
      goNext,
      goPrev,
      reset,
      hasPrev,
    }),
    [currentCursor, currentIndex, goNext, goPrev, reset, hasPrev],
  )
}
