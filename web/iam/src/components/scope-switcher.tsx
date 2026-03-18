import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'

interface ScopeSwitcherProps {
  label: string
  value: string | null
  items: Array<{ id: string; name: string }>
  onValueChange: (id: string | null) => void
  placeholder?: string
}

export function ScopeSwitcher({
  label,
  value,
  items,
  onValueChange,
  placeholder = '选择...',
}: ScopeSwitcherProps) {
  return (
    <div>
      <label className="mb-1 block text-[11px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
        {label}
      </label>
      <Select
        value={value ?? ''}
        onValueChange={(v) => onValueChange(v || null)}
      >
        <SelectTrigger className="h-8 w-full text-xs">
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          {items.map((item) => (
            <SelectItem key={item.id} value={item.id}>
              {item.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
