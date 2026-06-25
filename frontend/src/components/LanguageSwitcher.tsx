import { useTranslation } from "react-i18next"
import { setLang, supportedLngs, type Lang } from "@/i18n"

const labels: Record<Lang, string> = { ru: "RU", en: "EN", uz: "UZ", ja: "JA" }

export function LanguageSwitcher() {
  const { i18n } = useTranslation()
  return (
    <select
      value={i18n.resolvedLanguage ?? i18n.language}
      onChange={(e) => setLang(e.target.value as Lang)}
      className="h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
      aria-label="Language"
    >
      {supportedLngs.map((l) => (
        <option key={l} value={l}>
          {labels[l]}
        </option>
      ))}
    </select>
  )
}
