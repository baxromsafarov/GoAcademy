import { lazy, Suspense, useState } from "react"
import { useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Play, Clock } from "lucide-react"
import { useRunSandbox } from "@/lib/queries"
import { Button } from "@/components/ui/button"

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

export function Sandbox() {
  const { t } = useTranslation()
  const location = useLocation()
  const initial = (location.state as { code?: string } | null)?.code
  const [source, setSource] = useState(initial && initial.trim() ? initial : template)
  const [stdin, setStdin] = useState("")
  const run = useRunSandbox()
  const result = run.data

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">{t("nav.sandbox")}</h1>
        <Button onClick={() => run.mutate({ source, stdin })} disabled={run.isPending || !source.trim()}>
          <Play className="size-4" />
          {run.isPending ? t("sandbox.running") : t("sandbox.run")}
        </Button>
      </div>
      <p className="text-sm text-muted-foreground">{t("sandbox.stdlibNote")}</p>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">{t("sandbox.code")}</span>
          <Suspense
            fallback={<div className="h-[28rem] w-full animate-pulse rounded-lg bg-muted" />}
          >
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
            {result && (
              <div className="flex flex-col gap-3">
                <div className="flex flex-wrap items-center gap-2 text-xs">
                  {result.compile_error && <Badge tone="red">{t("sandbox.compileError")}</Badge>}
                  {result.timed_out && <Badge tone="amber">{t("sandbox.timedOut")}</Badge>}
                  {result.oom_killed && <Badge tone="amber">{t("sandbox.oom")}</Badge>}
                  {!result.compile_error && !result.timed_out && (
                    <Badge tone={result.exit_code === 0 ? "green" : "red"}>
                      {t("sandbox.exit", { code: result.exit_code })}
                    </Badge>
                  )}
                  <span className="flex items-center gap-1 text-muted-foreground">
                    <Clock className="size-3" /> {result.duration_ms} ms
                  </span>
                </div>

                {result.stdout && (
                  <Stream label="stdout" text={result.stdout} truncated={result.stdout_truncated} />
                )}
                {result.stderr && (
                  <Stream label="stderr" text={result.stderr} truncated={result.stderr_truncated} red />
                )}
                {!result.stdout && !result.stderr && (
                  <p className="text-muted-foreground">{t("sandbox.noOutput")}</p>
                )}
              </div>
            )}
            {!run.isPending && !run.isError && !result && (
              <p className="text-muted-foreground">{t("sandbox.empty")}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function Badge({ tone, children }: { tone: "green" | "red" | "amber"; children: React.ReactNode }) {
  const tones = {
    green: "border-green-500/40 bg-green-500/10 text-green-600 dark:text-green-400",
    red: "border-red-500/40 bg-red-500/10 text-red-600 dark:text-red-400",
    amber: "border-amber-500/40 bg-amber-500/10 text-amber-600 dark:text-amber-400",
  }
  return <span className={`rounded border px-1.5 py-0.5 ${tones[tone]}`}>{children}</span>
}

function Stream({
  label,
  text,
  truncated,
  red,
}: {
  label: string
  text: string
  truncated: boolean
  red?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div>
      <div className="mb-1 text-xs text-muted-foreground">
        {label}
        {truncated && <span className="ml-1 text-amber-600 dark:text-amber-400">({t("sandbox.truncated")})</span>}
      </div>
      <pre className={`overflow-x-auto rounded bg-muted p-2 font-mono text-xs ${red ? "text-red-600 dark:text-red-400" : ""}`}>
        {text}
      </pre>
    </div>
  )
}
