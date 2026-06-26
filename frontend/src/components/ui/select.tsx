import { useEffect, useRef, useState } from "react"
import { Check, ChevronDown } from "lucide-react"
import { cn } from "@/lib/utils"

export interface SelectOption {
  value: string
  label: string
}

/**
 * Select is a custom dropdown (not a native <select>) so its popup is always
 * painted with the theme tokens. Native option popups render white-on-white in
 * dark mode on some browsers/OSes, which is the bug this replaces everywhere.
 */
export function Select({
  value,
  onChange,
  options,
  ariaLabel,
  className,
  placeholder,
}: {
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  ariaLabel?: string
  className?: string
  placeholder?: string
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    function onDoc(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false)
    }
    document.addEventListener("mousedown", onDoc)
    document.addEventListener("keydown", onKey)
    return () => {
      document.removeEventListener("mousedown", onDoc)
      document.removeEventListener("keydown", onKey)
    }
  }, [open])

  const current = options.find((o) => o.value === value)

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex h-9 w-full items-center justify-between gap-2 rounded-md border bg-transparent px-2.5 text-sm transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label={ariaLabel}
      >
        <span className={cn("truncate", !current && "text-muted-foreground")}>
          {current?.label ?? placeholder ?? ""}
        </span>
        <ChevronDown className={cn("size-3.5 shrink-0 opacity-60 transition-transform", open && "rotate-180")} />
      </button>
      {open && (
        <ul
          role="listbox"
          className="absolute left-0 z-50 mt-1 max-h-64 min-w-full overflow-auto rounded-md border bg-card py-1 text-foreground shadow-lg"
        >
          {options.map((o) => (
            <li key={o.value}>
              <button
                type="button"
                role="option"
                aria-selected={o.value === value}
                onClick={() => {
                  onChange(o.value)
                  setOpen(false)
                }}
                className={cn(
                  "flex w-full items-center justify-between gap-3 px-3 py-1.5 text-left text-sm whitespace-nowrap hover:bg-accent hover:text-accent-foreground",
                  o.value === value && "font-medium",
                )}
              >
                <span>{o.label}</span>
                {o.value === value && <Check className="size-4 shrink-0 text-primary" />}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
