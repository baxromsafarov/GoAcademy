import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Check, ChevronDown } from "lucide-react"
import { setLang, supportedLngs, type Lang } from "@/i18n"
import { cn } from "@/lib/utils"

const labels: Record<Lang, string> = { ru: "RU", en: "EN", uz: "UZ", ja: "JA" }
const names: Record<Lang, string> = { ru: "Русский", en: "English", uz: "Oʻzbekcha", ja: "日本語" }

/**
 * LanguageSwitcher is a custom dropdown (not a native <select>) so the menu is
 * always styled with the theme tokens — readable in dark mode instead of the
 * white-on-white native popup some browsers render.
 */
export function LanguageSwitcher() {
  const { i18n } = useTranslation()
  const current = (i18n.resolvedLanguage ?? i18n.language) as Lang
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

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex h-9 items-center gap-1 rounded-md border bg-transparent px-2 text-sm hover:bg-accent"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label="Language"
      >
        {labels[current] ?? "EN"}
        <ChevronDown className="size-3.5 opacity-60" />
      </button>
      {open && (
        <ul
          role="listbox"
          className="animate-pop absolute right-0 z-50 mt-1 w-36 overflow-hidden rounded-md border bg-card py-1 text-foreground shadow-lg"
        >
          {supportedLngs.map((l) => (
            <li key={l}>
              <button
                type="button"
                role="option"
                aria-selected={l === current}
                onClick={() => {
                  setLang(l)
                  setOpen(false)
                }}
                className={cn(
                  "flex w-full items-center justify-between px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground",
                  l === current && "font-medium",
                )}
              >
                <span>{names[l]}</span>
                {l === current && <Check className="size-4 text-primary" />}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
