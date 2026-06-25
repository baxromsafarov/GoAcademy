import { useTranslation } from "react-i18next"

/** Placeholder is a stub section page until its real UI ships (CHAPTER 15).
 * titleKey/descriptionKey are i18n keys. */
export function Placeholder({ titleKey, descriptionKey }: { titleKey: string; descriptionKey?: string }) {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight">{t(titleKey)}</h1>
      <p className="mt-2 text-muted-foreground">{t(descriptionKey ?? "common.comingSoon")}</p>
    </div>
  )
}
