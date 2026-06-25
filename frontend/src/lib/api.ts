// Relative by default: the dev server (Vite proxy) and production (nginx) both
// reverse-proxy /api to the backend, keeping requests same-origin (no CORS).
// Override with VITE_API_BASE_URL only when the API is on a different origin.
const BASE = import.meta.env.VITE_API_BASE_URL ?? "/api/v1"

// Access token is held in memory only (the refresh token lives in an httpOnly cookie).
let accessToken: string | null = null
let onAuthFailure: (() => void) | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

/** Registered by the AuthProvider; called when a refresh ultimately fails. */
export function setOnAuthFailure(cb: (() => void) | null) {
  onAuthFailure = cb
}

export interface ApiErrorBody {
  code: string
  message: string
  details?: unknown
}

/** ApiError mirrors the backend's {error:{code,message,details}} envelope. */
export class ApiError extends Error {
  status: number
  code: string
  details?: unknown
  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.name = "ApiError"
    this.status = status
    this.code = body.code
    this.details = body.details
  }
}

/**
 * errorMessage extracts a human-readable reason from any thrown error so forms
 * can show the actual cause instead of a generic message:
 * - ApiError → the backend message, plus any field-level validation details
 * - a network/fetch failure (server unreachable, cross-origin block) → its
 *   message, prefixed so the user knows it's a connectivity problem
 * - anything else → its string form
 */
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.details && typeof err.details === "object") {
      const parts = Object.entries(err.details as Record<string, unknown>).map(
        ([k, v]) => `${k}: ${String(v)}`,
      )
      if (parts.length > 0) return `${err.message} (${parts.join("; ")})`
    }
    return err.message
  }
  if (err instanceof Error) {
    // e.g. "Failed to fetch" — the API is unreachable.
    return `${err.message} — не удалось связаться с сервером`
  }
  return String(err)
}

interface RequestOptions {
  method?: string
  body?: unknown
  auth?: boolean
  _retry?: boolean
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, auth = true, _retry = false } = options
  const isForm = typeof FormData !== "undefined" && body instanceof FormData
  const headers: Record<string, string> = {}
  // For multipart (FormData) the browser sets Content-Type + boundary itself.
  if (body !== undefined && !isForm) headers["Content-Type"] = "application/json"
  if (auth && accessToken) headers["Authorization"] = `Bearer ${accessToken}`

  const res = await fetch(BASE + path, {
    method,
    headers,
    credentials: "include",
    body: body === undefined ? undefined : isForm ? (body as FormData) : JSON.stringify(body),
  })

  // On a 401 for an authenticated request, try a single refresh + replay.
  if (res.status === 401 && auth && !_retry) {
    if (await refresh()) {
      return request<T>(path, { ...options, _retry: true })
    }
  }

  if (res.status === 204) return undefined as T

  const data = (await res.json().catch(() => null)) as { error?: ApiErrorBody } | null
  if (!res.ok) {
    const err = data?.error ?? { code: "internal", message: `request failed (${res.status})` }
    throw new ApiError(res.status, err)
  }
  return data as T
}

/** refresh exchanges the refresh cookie for a new access token. */
export async function refresh(): Promise<boolean> {
  try {
    const res = await fetch(BASE + "/auth/refresh", { method: "POST", credentials: "include" })
    if (!res.ok) {
      accessToken = null
      onAuthFailure?.()
      return false
    }
    const data = (await res.json()) as { access_token: string }
    accessToken = data.access_token
    return true
  } catch {
    accessToken = null
    onAuthFailure?.()
    return false
  }
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown, auth = true) =>
    request<T>(path, { method: "POST", body, auth }),
  patch: <T>(path: string, body?: unknown) => request<T>(path, { method: "PATCH", body }),
  del: (path: string) => request<void>(path, { method: "DELETE" }),
  upload: <T>(path: string, form: FormData) => request<T>(path, { method: "POST", body: form }),
}
