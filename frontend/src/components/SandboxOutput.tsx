import { useTranslation } from "react-i18next"
import { Clock } from "lucide-react"
import type { SandboxRunResult } from "@/lib/types"

/** SandboxOutput renders a run result: status badges plus stdout/stderr. Shared
 * by the full sandbox page and the inline runner inside articles. */
export function SandboxOutput({ result }: { result: SandboxRunResult }) {
  const { t } = useTranslation()
  return (
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
        <p className="text-sm text-muted-foreground">{t("sandbox.noOutput")}</p>
      )}
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
        {truncated && (
          <span className="ml-1 text-amber-600 dark:text-amber-400">({t("sandbox.truncated")})</span>
        )}
      </div>
      <pre
        className={`overflow-x-auto rounded bg-muted p-2 font-mono text-xs ${red ? "text-red-600 dark:text-red-400" : ""}`}
      >
        {text}
      </pre>
    </div>
  )
}
