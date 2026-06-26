import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Loader2, Play, SquarePen, X } from "lucide-react"
import ReactMarkdown, { type Components } from "react-markdown"
import { useRunSandbox } from "@/lib/queries"
import { SandboxOutput } from "@/components/SandboxOutput"
import remarkGfm from "remark-gfm"
import rehypeHighlight from "rehype-highlight"
// Register only the grammars a Go course needs instead of highlight.js's full
// "common" set (~37 languages); keeps this lazy-loaded chunk small.
import go from "highlight.js/lib/languages/go"
import bash from "highlight.js/lib/languages/bash"
import json from "highlight.js/lib/languages/json"
import yaml from "highlight.js/lib/languages/yaml"
import sql from "highlight.js/lib/languages/sql"
import dockerfile from "highlight.js/lib/languages/dockerfile"
// highlight.js theme for fenced code blocks (styles the .hljs token classes).
import "highlight.js/styles/github-dark.css"

const hljsLanguages = { go, bash, shell: bash, json, yaml, sql, dockerfile }

/** nodeText walks an arbitrary React node tree and concatenates all of its text.
 * rehype-highlight wraps keywords, numbers, strings etc. in <span> tokens, so a
 * shallow "keep only string children" pass would silently drop them — which is
 * exactly what mangled the code sent to the sandbox. */
function nodeText(node: unknown): string {
  if (node == null || typeof node === "boolean") return ""
  if (typeof node === "string") return node
  if (typeof node === "number") return String(node)
  if (Array.isArray(node)) return node.map(nodeText).join("")
  if (typeof node === "object" && "props" in node) {
    return nodeText((node as { props?: { children?: unknown } }).props?.children)
  }
  return ""
}

/** extractCode pulls the raw source text out of a fenced code block's rendered
 * <code> element so the "Open in sandbox" button can prefill the editor. It
 * recurses through the highlight.js token spans to keep the source intact. */
function extractCode(children: unknown): string {
  return nodeText(children)
}

/** codeLang reads the highlight.js language class off the <code> element. */
function codeLang(children: unknown): string {
  const cn = (children as { props?: { className?: string } })?.props?.className ?? ""
  const m = cn.match(/language-(\w+)/)
  return m ? m[1] : ""
}

/**
 * RunnableCodeBlock wraps a fenced code block with a toolbar: run it inline
 * (right here in the article, no navigation) and open it in the full sandbox.
 * Output appears directly under the block. Run/Open show only for Go code.
 */
function RunnableCodeBlock({ code, runnable, children }: { code: string; runnable: boolean; children: React.ReactNode }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const run = useRunSandbox()
  const [showOutput, setShowOutput] = useState(false)

  return (
    <div className="my-4 overflow-hidden rounded-md border">
      <div className="group relative">
        {runnable && (
          <div className="absolute top-2 right-2 z-10 flex items-center gap-1 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100">
            <button
              type="button"
              onClick={() => {
                setShowOutput(true)
                run.mutate({ source: code, stdin: "" })
              }}
              disabled={run.isPending}
              title={t("sandbox.run")}
              className="flex items-center gap-1 rounded border border-white/20 bg-black/40 px-2 py-1 text-xs text-white backdrop-blur hover:bg-black/60 disabled:opacity-60"
            >
              {run.isPending ? <Loader2 className="size-3 animate-spin" /> : <Play className="size-3" />}
              {t("sandbox.run")}
            </button>
            <button
              type="button"
              onClick={() => navigate("/sandbox", { state: { code } })}
              title={t("articles.openInSandbox")}
              className="flex items-center gap-1 rounded border border-white/20 bg-black/40 px-2 py-1 text-xs text-white backdrop-blur hover:bg-black/60"
            >
              <SquarePen className="size-3" />
              {t("articles.openInSandbox")}
            </button>
          </div>
        )}
        <pre className="overflow-x-auto text-sm">{children}</pre>
      </div>

      {runnable && showOutput && (
        <div className="border-t bg-card/50 p-3">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t("sandbox.output")}</span>
            <button
              type="button"
              onClick={() => setShowOutput(false)}
              className="rounded p-0.5 text-muted-foreground hover:bg-accent"
              aria-label={t("common.back")}
            >
              <X className="size-3.5" />
            </button>
          </div>
          {run.isPending && <p className="text-sm text-muted-foreground">{t("sandbox.running")}</p>}
          {run.isError && <p className="text-sm text-red-500">{t("sandbox.unavailable")}</p>}
          {run.data && <SandboxOutput result={run.data} />}
        </div>
      )}
    </div>
  )
}

/**
 * Markdown renders trusted-but-untrusted-content markdown safely.
 *
 * Security: react-markdown does NOT parse embedded raw HTML (no rehype-raw is
 * used), so any <script>/<img onerror> in the source is rendered as inert text
 * rather than live DOM. Its default URL transform also strips javascript: and
 * other dangerous URI schemes from links/images. That is the sanitization layer
 * required by the article reader — rehype-highlight only adds className tokens
 * to code we already control.
 */
export function Markdown({ children }: { children: string }) {
  const components: Components = {
    h1: (props) => <h1 className="mt-6 mb-3 text-2xl font-semibold tracking-tight" {...props} />,
    h2: (props) => <h2 className="mt-6 mb-3 text-xl font-semibold tracking-tight" {...props} />,
    h3: (props) => <h3 className="mt-4 mb-2 text-lg font-semibold tracking-tight" {...props} />,
    h4: (props) => <h4 className="mt-4 mb-2 font-semibold" {...props} />,
    p: (props) => <p className="my-3 leading-7" {...props} />,
    a: ({ href, ...props }) => (
      <a
        href={href}
        target="_blank"
        rel="noreferrer noopener"
        className="font-medium text-primary underline underline-offset-2"
        {...props}
      />
    ),
    ul: (props) => <ul className="my-3 list-disc space-y-1 pl-6" {...props} />,
    ol: (props) => <ol className="my-3 list-decimal space-y-1 pl-6" {...props} />,
    blockquote: (props) => (
      <blockquote
        className="my-5 rounded-lg border border-primary/20 bg-primary/5 px-4 py-3 text-[0.95em] [&>p]:my-1"
        {...props}
      />
    ),
    hr: (props) => <hr className="my-6 border-t" {...props} />,
    img: ({ alt, ...props }) => (
      <img alt={alt ?? ""} className="my-4 max-w-full rounded-md border" {...props} />
    ),
    table: (props) => (
      <div className="my-4 overflow-x-auto">
        <table className="w-full border-collapse text-sm" {...props} />
      </div>
    ),
    th: (props) => <th className="border px-3 py-1.5 text-left font-semibold" {...props} />,
    td: (props) => <td className="border px-3 py-1.5" {...props} />,
    code: ({ className, children, ...props }) => {
      const text = String(children)
      const isBlock = text.includes("\n") || /language-|hljs/.test(className ?? "")
      if (isBlock) {
        return (
          <code className={className} {...props}>
            {children}
          </code>
        )
      }
      return (
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.85em]" {...props}>
          {children}
        </code>
      )
    },
    pre: ({ children }) => {
      const code = extractCode(children)
      const lang = codeLang(children)
      // The sandbox only runs Go; show the run/open toolbar for Go (or untagged)
      // blocks that look like a program, not for bash/json/yaml snippets.
      const runnable = (lang === "" || lang === "go") && /\b(package|func|import)\b/.test(code)
      return (
        <RunnableCodeBlock code={code} runnable={runnable}>
          {children}
        </RunnableCodeBlock>
      )
    },
  }

  return (
    <div className="max-w-none">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeHighlight, { languages: hljsLanguages }]]}
        components={components}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
