import { useId, useRef, type KeyboardEvent, type UIEvent } from "react"
import hljs from "highlight.js/lib/core"
import go from "highlight.js/lib/languages/go"
import javascript from "highlight.js/lib/languages/javascript"
import python from "highlight.js/lib/languages/python"
import rust from "highlight.js/lib/languages/rust"
import java from "highlight.js/lib/languages/java"
// github-dark styles the .hljs token classes; the editor chrome is always dark
// (like a real IDE / Practicum's editor) so these colours read well regardless
// of the app's light/dark theme.
import "highlight.js/styles/github-dark.css"
import { cn } from "@/lib/utils"

const grammars: Record<string, unknown> = { go, javascript, python, rust, java }
for (const [name, grammar] of Object.entries(grammars)) {
  // Registering is a no-op if the language is already known, so this is safe
  // even when several editors mount.
  if (!hljs.getLanguage(name)) hljs.registerLanguage(name, grammar as never)
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;")
}

function highlight(code: string, language: string): string {
  // A trailing newline keeps the last line's height when the source ends in \n.
  const src = code.endsWith("\n") ? code : code + "\n"
  if (hljs.getLanguage(language)) {
    return hljs.highlight(src, { language, ignoreIllegals: true }).value
  }
  return escapeHtml(src)
}

/**
 * CodeEditor is a lightweight syntax-highlighting editor: a transparent
 * <textarea> layered over a highlighted <pre> that scrolls in lockstep. It
 * pulls in highlight.js, so import it lazily (React.lazy) to keep that weight
 * out of the initial bundle. The chrome is intentionally always dark.
 */
export function CodeEditor({
  value,
  onChange,
  language = "go",
  placeholder,
  className,
  ariaLabel,
}: {
  value: string
  onChange: (value: string) => void
  language?: string
  placeholder?: string
  className?: string
  ariaLabel?: string
}) {
  const preRef = useRef<HTMLPreElement>(null)
  const id = useId()

  function syncScroll(e: UIEvent<HTMLTextAreaElement>) {
    const pre = preRef.current
    if (!pre) return
    pre.scrollTop = e.currentTarget.scrollTop
    pre.scrollLeft = e.currentTarget.scrollLeft
  }

  function onKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Tab") {
      e.preventDefault()
      const ta = e.currentTarget
      const { selectionStart: start, selectionEnd: end } = ta
      const next = value.slice(0, start) + "\t" + value.slice(end)
      onChange(next)
      // Restore caret just after the inserted tab on the next tick.
      requestAnimationFrame(() => {
        ta.selectionStart = ta.selectionEnd = start + 1
      })
    }
  }

  const shared = "m-0 p-3 font-mono text-sm leading-6 whitespace-pre [tab-size:4]"

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-lg border border-white/10 bg-[#0d1117] text-gray-100",
        className,
      )}
    >
      <pre ref={preRef} aria-hidden className={cn(shared, "absolute inset-0 overflow-auto")}>
        <code
          className={`hljs language-${language} !bg-transparent !p-0`}
          dangerouslySetInnerHTML={{ __html: highlight(value, language) }}
        />
      </pre>
      <textarea
        id={id}
        aria-label={ariaLabel}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onScroll={syncScroll}
        onKeyDown={onKeyDown}
        placeholder={placeholder}
        spellCheck={false}
        autoComplete="off"
        autoCorrect="off"
        autoCapitalize="off"
        wrap="off"
        className={cn(
          shared,
          "absolute inset-0 size-full resize-none overflow-auto bg-transparent text-transparent caret-white outline-none placeholder:text-gray-500",
        )}
      />
    </div>
  )
}
