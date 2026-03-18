import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from '#/components/ui/sheet'
import { Button } from '#/components/ui/button'
import { Separator } from '#/components/ui/separator'

interface FormDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  loading?: boolean
  onSubmit: () => void | Promise<void>
  submitLabel?: string
  children: React.ReactNode
}

export function FormDrawer({
  open,
  onOpenChange,
  title,
  loading,
  onSubmit,
  submitLabel = '确认',
  children,
}: FormDrawerProps) {
  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    await onSubmit()
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex flex-col sm:max-w-md" aria-describedby={undefined}>
        <SheetHeader>
          <SheetTitle>{title}</SheetTitle>
        </SheetHeader>
        <Separator />
        <form onSubmit={handleSubmit} className="flex flex-1 flex-col gap-4 overflow-y-auto px-1 py-4">
          {children}
        </form>
        <Separator />
        <SheetFooter className="flex-row justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type="submit" disabled={loading} onClick={handleSubmit}>
            {loading ? '提交中...' : submitLabel}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
