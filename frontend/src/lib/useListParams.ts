import { useSearchParams } from "react-router-dom"
import { useTranslation } from "react-i18next"

/** Items per page for content lists. A multiple of the 1/2/3-column grids. */
export const PAGE_SIZE = 12

/**
 * useListParams keeps a content list's filters and page in the URL query string.
 * Because the state lives in the URL, it survives navigating into a card and
 * back (the browser restores the URL) and is shared by pagination links — fixing
 * both the "filter resets on back" bug and filters being lost between pages.
 *
 * Language is special: with no `lang` param the content language follows the UI
 * language (as before); `lang=all` means every language; `lang=ru` pins one.
 */
export function useListParams() {
  const [sp, setSp] = useSearchParams()
  const { i18n } = useTranslation()
  const ui = i18n.resolvedLanguage ?? "en"

  const get = (key: string) => sp.get(key) ?? ""

  const langParam = sp.get("lang")
  const language = langParam === null ? ui : langParam === "all" ? "" : langParam

  const page = Math.max(1, Number(sp.get("page")) || 1)
  const offset = (page - 1) * PAGE_SIZE

  function patch(updates: Record<string, string | null>, resetPage = true) {
    const next = new URLSearchParams(sp)
    for (const [key, value] of Object.entries(updates)) {
      if (!value) next.delete(key)
      else next.set(key, value)
    }
    if (resetPage) next.delete("page")
    setSp(next, { replace: true })
  }

  return {
    get,
    language,
    page,
    offset,
    pageSize: PAGE_SIZE,
    /** Set (or clear, when empty) a filter param; resets to the first page. */
    setParam: (key: string, value: string) => patch({ [key]: value }),
    /** "" selects all languages; otherwise pins a specific language. */
    setLanguage: (value: string) => patch({ lang: value === "" ? "all" : value }),
    setPage: (p: number) => patch({ page: p <= 1 ? null : String(p) }, false),
  }
}
