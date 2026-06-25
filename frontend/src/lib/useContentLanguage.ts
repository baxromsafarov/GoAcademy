import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

/**
 * useContentLanguage returns the language to filter content by. It defaults to —
 * and follows — the current UI language, so switching the interface language also
 * switches which content (videos, articles, quizzes, …) is shown. The user can
 * still override it via the returned setter, including "" to mean "all languages";
 * changing the UI language clears that override so content follows the UI again.
 */
export function useContentLanguage(): [string, (language: string) => void] {
  const { i18n } = useTranslation()
  const ui = i18n.resolvedLanguage ?? "en"
  const [override, setOverride] = useState<string | null>(null)

  useEffect(() => {
    setOverride(null)
  }, [ui])

  return [override ?? ui, setOverride]
}
