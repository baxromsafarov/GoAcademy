import type { TFunction } from "i18next"
import type { SelectOption } from "@/components/ui/select"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]

/** "Difficulty: All" plus each difficulty, for the content list filters. */
export function difficultyOptions(t: TFunction): SelectOption[] {
  return [
    { value: "", label: `${t("videos.filterDifficulty")}: ${t("common.all")}` },
    ...difficulties.map((d) => ({ value: d, label: t(`difficulty.${d}`) })),
  ]
}

/** "Language: All" plus each UI language, for the content list filters. */
export function languageOptions(t: TFunction): SelectOption[] {
  return [
    { value: "", label: `${t("videos.filterLanguage")}: ${t("common.all")}` },
    ...langs.map((l) => ({ value: l, label: l.toUpperCase() })),
  ]
}
