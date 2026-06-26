import type { FormEvent, ReactNode } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft } from "lucide-react"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"

/** Field is a labelled form row. */
export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label>{label}</Label>
      {children}
    </div>
  )
}

/** AdminFormShell is the common chrome for an admin create/edit form: a back
 * link, a title, the fields, and a save button with error state. */
export function AdminFormShell({
  backTo,
  title,
  onSubmit,
  saving,
  isError,
  children,
}: {
  backTo: string
  title: string
  onSubmit: (e: FormEvent) => void
  saving: boolean
  isError: boolean
  children: ReactNode
}) {
  const { t } = useTranslation()
  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-4">
      <Link
        to={backTo}
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>
      <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        {children}
        <div className="flex items-center gap-3">
          <Button type="submit" disabled={saving}>
            {saving ? t("admin.saving") : t("admin.save")}
          </Button>
          {isError && <span className="text-sm text-red-500">{t("admin.saveError")}</span>}
        </div>
      </form>
    </div>
  )
}

/**
 * VisibilityField is the show/hide toggle. Visibility is stored as a "hidden"
 * tag on the content; checking the box adds it (and the public lists then skip
 * the item), unchecking removes it.
 */
export function VisibilityField({
  tags,
  onChange,
}: {
  tags: string[]
  onChange: (tags: string[]) => void
}) {
  const { t } = useTranslation()
  const hidden = tags.includes("hidden")
  return (
    <label className="flex w-fit cursor-pointer items-center gap-2 text-sm">
      <input
        type="checkbox"
        checked={hidden}
        onChange={(e) =>
          onChange(
            e.target.checked
              ? [...tags.filter((x) => x !== "hidden"), "hidden"]
              : tags.filter((x) => x !== "hidden"),
          )
        }
        className="size-4 accent-primary"
      />
      {t("admin.hiddenToggle")}
    </label>
  )
}

/** Plain textarea styled for the admin forms. */
export function TextArea({
  value,
  onChange,
  rows = 4,
  mono,
}: {
  value: string
  onChange: (v: string) => void
  rows?: number
  mono?: boolean
}) {
  return (
    <textarea
      value={value}
      onChange={(e) => onChange(e.target.value)}
      rows={rows}
      className={`w-full resize-y rounded-md border bg-transparent p-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring ${mono ? "font-mono" : ""}`}
    />
  )
}
