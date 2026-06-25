import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import { en } from "./locales/en"
import { ru } from "./locales/ru"
import { uz } from "./locales/uz"
import { ja } from "./locales/ja"

export const supportedLngs = ["ru", "en", "uz", "ja"] as const
export type Lang = (typeof supportedLngs)[number]

function isLang(v: string | null): v is Lang {
  return v !== null && (supportedLngs as readonly string[]).includes(v)
}

function initialLang(): Lang {
  const stored = localStorage.getItem("lang")
  return isLang(stored) ? stored : "en"
}

void i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    ru: { translation: ru },
    uz: { translation: uz },
    ja: { translation: ja },
  },
  lng: initialLang(),
  fallbackLng: "en",
  interpolation: { escapeValue: false },
})

export default i18n

/** setLang switches the UI language and persists the choice. */
export function setLang(lang: Lang) {
  void i18n.changeLanguage(lang)
  localStorage.setItem("lang", lang)
}

/** applyProfileLocale applies a user's profile locale if it is supported (login). */
export function applyProfileLocale(locale: string) {
  if (isLang(locale)) setLang(locale)
}
