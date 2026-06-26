import { lazy, Suspense, useEffect, useState } from "react"
import { useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Play, RotateCcw } from "lucide-react"
import { useRunSandbox } from "@/lib/queries"
import { Button } from "@/components/ui/button"
import { SandboxOutput } from "@/components/SandboxOutput"

// CodeEditor pulls in highlight.js — keep it out of the initial bundle.
const CodeEditor = lazy(() =>
  import("@/components/CodeEditor").then((m) => ({ default: m.CodeEditor })),
)

const template = `package main

import "fmt"

func main() {
	fmt.Println("Hello, GoAcademy!")
}
`

// Persist the editor between visits so a student's work survives navigation and
// reloads; they can reset to the starter template whenever they like.
const STORAGE_KEY = "sandbox.code"

export function Sandbox() {
  const { t } = useTranslation()
  const location = useLocation()
  const initial = (location.state as { code?: string } | null)?.code
  const [source, setSource] = useState(() => {
    if (initial && initial.trim()) return initial // arrived via "open in sandbox"
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved && saved.trim() ? saved : template
  })
  const [stdin, setStdin] = useState("")
  const run = useRunSandbox()
  const result = run.data

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, source)
  }, [source])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">{t("nav.sandbox")}</h1>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => setSource(template)}
            disabled={source === template}
            title={t("sandbox.reset")}
          >
            <RotateCcw className="size-4" />
            <span className="hidden sm:inline">{t("sandbox.reset")}</span>
          </Button>
          <Button onClick={() => run.mutate({ source, stdin })} disabled={run.isPending || !source.trim()}>
            <Play className="size-4" />
            {run.isPending ? t("sandbox.running") : t("sandbox.run")}
          </Button>
        </div>
      </div>
      <p className="text-sm text-muted-foreground">{t("sandbox.stdlibNote")}</p>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">{t("sandbox.code")}</span>
          <Suspense fallback={<div className="h-[28rem] w-full animate-pulse rounded-lg bg-muted" />}>
            <CodeEditor
              value={source}
              onChange={setSource}
              language="go"
              ariaLabel={t("sandbox.code")}
              className="h-[28rem]"
            />
          </Suspense>
          <label className="text-sm font-medium" htmlFor="stdin">
            {t("sandbox.stdin")}
          </label>
          <textarea
            id="stdin"
            value={stdin}
            onChange={(e) => setStdin(e.target.value)}
            spellCheck={false}
            rows={3}
            placeholder={t("sandbox.stdinHint")}
            className="w-full resize-y rounded-md border bg-card p-3 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>

        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">{t("sandbox.output")}</span>
          <div className="min-h-[18rem] rounded-md border bg-card p-3 text-sm">
            {run.isPending && <p className="text-muted-foreground">{t("sandbox.running")}</p>}
            {run.isError && <p className="text-red-500">{t("sandbox.unavailable")}</p>}
            {result && <SandboxOutput result={result} />}
            {!run.isPending && !run.isError && !result && (
              <p className="text-muted-foreground">{t("sandbox.empty")}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
