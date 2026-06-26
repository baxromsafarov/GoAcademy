import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api, ApiError } from "@/lib/api"

/** Pagination shared by every content list filter. */
export interface PageFilter {
  limit?: number
  offset?: number
}

/** appendPage adds limit/offset to a query string when present (offset may be 0). */
function appendPage(params: URLSearchParams, f: PageFilter) {
  if (f.limit != null) params.set("limit", String(f.limit))
  if (f.offset != null) params.set("offset", String(f.offset))
}
import type {
  ActivityResponse,
  AdminUserListResponse,
  Article,
  ArticleListResponse,
  ArticleReadStatus,
  BadgesResponse,
  BookmarksResponse,
  CheatsheetDetail,
  CheatsheetListResponse,
  GlossaryListResponse,
  LeaderboardResponse,
  Note,
  NotesResponse,
  ProblemDetail,
  ProblemListResponse,
  ProblemSubmissionResult,
  ProgressSummary,
  ProjectDetail,
  ProjectListResponse,
  ProjectProgress,
  QuizAttemptResult,
  QuizDetail,
  QuizListResponse,
  SandboxRunResult,
  Stats,
  TrackDetail,
  TrackListResponse,
  TrackProgress,
  User,
  Video,
  VideoListResponse,
  VideoProgress,
} from "@/lib/types"

export function useStats() {
  return useQuery({ queryKey: ["me", "stats"], queryFn: () => api.get<Stats>("/me/stats") })
}

export function useProgressSummary() {
  return useQuery({
    queryKey: ["me", "progress"],
    queryFn: () => api.get<ProgressSummary>("/me/progress"),
  })
}

export function useBadges() {
  return useQuery({ queryKey: ["me", "badges"], queryFn: () => api.get<BadgesResponse>("/me/badges") })
}

export function useActivity() {
  return useQuery({
    queryKey: ["me", "activity"],
    queryFn: () => api.get<ActivityResponse>("/me/activity"),
  })
}

export interface VideoFilters extends PageFilter {
  difficulty?: string
  tag?: string
  language?: string
  q?: string
}

export function useVideos(filters: VideoFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.tag) params.set("tag", filters.tag)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["videos", filters],
    queryFn: () => api.get<VideoListResponse>("/videos" + (qs ? "?" + qs : "")),
  })
}

export function useVideo(id: string) {
  return useQuery({ queryKey: ["video", id], queryFn: () => api.get<Video>(`/videos/${id}`) })
}

export function useVideoProgress(id: string) {
  return useQuery({
    queryKey: ["video", id, "progress"],
    queryFn: () => api.get<VideoProgress>(`/videos/${id}/progress`),
  })
}

export function usePostVideoProgress(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { percent: number; position: number; completed?: boolean }) =>
      api.post<VideoProgress>(`/videos/${id}/progress`, body),
    onSuccess: (data) => {
      qc.setQueryData(["video", id, "progress"], data)
    },
  })
}

export interface ArticleFilters extends PageFilter {
  difficulty?: string
  tag?: string
  language?: string
  q?: string
}

export function useArticles(filters: ArticleFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.tag) params.set("tag", filters.tag)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["articles", filters],
    queryFn: () => api.get<ArticleListResponse>("/articles" + (qs ? "?" + qs : "")),
  })
}

export function useArticle(slug: string) {
  return useQuery({
    queryKey: ["article", slug],
    queryFn: () => api.get<Article>(`/articles/${slug}`),
  })
}

export function useArticleReadStatus(slug: string) {
  return useQuery({
    queryKey: ["article", slug, "read"],
    queryFn: () => api.get<ArticleReadStatus>(`/articles/${slug}/read`),
  })
}

export function useMarkArticleRead(slug: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ article_id: string; completed_at: string }>(`/articles/${slug}/complete`, {}),
    onSuccess: (data) => {
      qc.setQueryData<ArticleReadStatus>(["article", slug, "read"], {
        read: true,
        completed_at: data.completed_at,
      })
      qc.invalidateQueries({ queryKey: ["me", "progress"] })
    },
  })
}

export interface QuizFilters extends PageFilter {
  difficulty?: string
  tag?: string
  language?: string
  q?: string
}

export function useQuizzes(filters: QuizFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.tag) params.set("tag", filters.tag)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["quizzes", filters],
    queryFn: () => api.get<QuizListResponse>("/quizzes" + (qs ? "?" + qs : "")),
  })
}

export function useQuiz(id: string) {
  return useQuery({ queryKey: ["quiz", id], queryFn: () => api.get<QuizDetail>(`/quizzes/${id}`) })
}

export function useSubmitQuiz(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (answers: Record<string, string[]>) =>
      api.post<QuizAttemptResult>(`/quizzes/${id}/attempts`, { answers }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["me", "progress"] })
      qc.invalidateQueries({ queryKey: ["me", "stats"] })
    },
  })
}

export interface ProblemFilters extends PageFilter {
  difficulty?: string
  tag?: string
  language?: string
  q?: string
}

export function useProblems(filters: ProblemFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.tag) params.set("tag", filters.tag)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["problems", filters],
    queryFn: () => api.get<ProblemListResponse>("/problems" + (qs ? "?" + qs : "")),
  })
}

export function useProblem(slug: string) {
  return useQuery({
    queryKey: ["problem", slug],
    queryFn: () => api.get<ProblemDetail>(`/problems/${slug}`),
  })
}

/**
 * useProblemSolution probes the reference solution. The endpoint returns 403
 * until the user has solved the problem, so a 403 is mapped to null ("not yet
 * solved") rather than an error — that lets a solved state survive reloads.
 */
export function useProblemSolution(slug: string) {
  return useQuery({
    queryKey: ["problem", slug, "solution"],
    retry: false,
    queryFn: async () => {
      try {
        return await api.get<{ reference_solution_markdown: string }>(`/problems/${slug}/solution`)
      } catch (err) {
        if (err instanceof ApiError && err.status === 403) return null
        throw err
      }
    },
  })
}

export function useSubmitProblem(slug: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { code: string; language: string; solved: boolean }) =>
      api.post<ProblemSubmissionResult>(`/problems/${slug}/submissions`, body),
    onSuccess: (data) => {
      if (data.status === "solved") {
        if (data.reference_solution_markdown) {
          qc.setQueryData(["problem", slug, "solution"], {
            reference_solution_markdown: data.reference_solution_markdown,
          })
        } else {
          qc.invalidateQueries({ queryKey: ["problem", slug, "solution"] })
        }
        qc.invalidateQueries({ queryKey: ["me", "progress"] })
        qc.invalidateQueries({ queryKey: ["me", "stats"] })
      }
    },
  })
}

export interface TrackFilters extends PageFilter {
  difficulty?: string
  language?: string
  q?: string
}

export function useTracks(filters: TrackFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["tracks", filters],
    queryFn: () => api.get<TrackListResponse>("/tracks" + (qs ? "?" + qs : "")),
  })
}

export function useTrack(id: string) {
  return useQuery({ queryKey: ["track", id], queryFn: () => api.get<TrackDetail>(`/tracks/${id}`) })
}

export function useTrackProgress(id: string) {
  return useQuery({
    queryKey: ["track", id, "progress"],
    queryFn: () => api.get<TrackProgress>(`/tracks/${id}/progress`),
  })
}

/** The tracks the current user has enrolled in (followed). */
export function useMyTracks() {
  return useQuery({
    queryKey: ["me", "tracks"],
    queryFn: () => api.get<TrackListResponse>("/me/tracks"),
  })
}

export function useEnrollTrack() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, enrolled }: { id: string; enrolled: boolean }) =>
      enrolled ? api.del(`/tracks/${id}/enroll`) : api.post(`/tracks/${id}/enroll`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "tracks"] }),
  })
}

export interface CheatsheetFilters extends PageFilter {
  category?: string
  q?: string
  language?: string
}

export function useCheatsheets(filters: CheatsheetFilters = {}) {
  const params = new URLSearchParams()
  if (filters.category) params.set("category", filters.category)
  if (filters.q) params.set("q", filters.q)
  if (filters.language) params.set("language", filters.language)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["cheatsheets", filters],
    queryFn: () => api.get<CheatsheetListResponse>("/cheatsheets" + (qs ? "?" + qs : "")),
  })
}

export function useCheatsheet(id: string) {
  return useQuery({
    queryKey: ["cheatsheet", id],
    queryFn: () => api.get<CheatsheetDetail>(`/cheatsheets/${id}`),
  })
}

export function useGlossary(filters: { q?: string; language?: string } & PageFilter = {}) {
  const params = new URLSearchParams()
  if (filters.q) params.set("q", filters.q)
  if (filters.language) params.set("language", filters.language)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["glossary", filters],
    queryFn: () => api.get<GlossaryListResponse>("/glossary" + (qs ? "?" + qs : "")),
  })
}

export interface ProjectFilters extends PageFilter {
  difficulty?: string
  tag?: string
  language?: string
  q?: string
}

export function useProjects(filters: ProjectFilters = {}) {
  const params = new URLSearchParams()
  if (filters.difficulty) params.set("difficulty", filters.difficulty)
  if (filters.tag) params.set("tag", filters.tag)
  if (filters.language) params.set("language", filters.language)
  if (filters.q) params.set("q", filters.q)
  appendPage(params, filters)
  const qs = params.toString()
  return useQuery({
    queryKey: ["projects", filters],
    queryFn: () => api.get<ProjectListResponse>("/projects" + (qs ? "?" + qs : "")),
  })
}

export function useProject(id: string) {
  return useQuery({ queryKey: ["project", id], queryFn: () => api.get<ProjectDetail>(`/projects/${id}`) })
}

export function useProjectProgress(id: string) {
  return useQuery({
    queryKey: ["project", id, "progress"],
    queryFn: () => api.get<ProjectProgress>(`/projects/${id}/progress`),
  })
}

export function useToggleProjectStep(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (stepId: string) =>
      api.post<ProjectProgress>(`/projects/${id}/steps/${stepId}/toggle`, {}),
    onSuccess: (data) => {
      qc.setQueryData(["project", id, "progress"], data)
      qc.invalidateQueries({ queryKey: ["me", "progress"] })
      qc.invalidateQueries({ queryKey: ["me", "stats"] })
    },
  })
}

export function useLeaderboard(period: string, limit?: number) {
  const qs = limit ? `&limit=${limit}` : ""
  return useQuery({
    queryKey: ["leaderboard", period, limit],
    queryFn: () => api.get<LeaderboardResponse>(`/leaderboard?period=${period}${qs}`),
  })
}

export function useNotes() {
  return useQuery({ queryKey: ["me", "notes"], queryFn: () => api.get<NotesResponse>("/me/notes") })
}

export function useCreateNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { content_type: string; content_id: string; body: string }) =>
      api.post<Note>("/notes", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "notes"] }),
  })
}

export function useUpdateNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) => api.patch<Note>(`/notes/${id}`, { body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "notes"] }),
  })
}

export function useDeleteNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.del(`/notes/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "notes"] }),
  })
}

export function useBookmarks() {
  return useQuery({
    queryKey: ["me", "bookmarks"],
    queryFn: () => api.get<BookmarksResponse>("/me/bookmarks"),
  })
}

export function useCreateBookmark() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { content_type: string; content_id: string }) =>
      api.post("/bookmarks", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "bookmarks"] }),
  })
}

export function useDeleteBookmark() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.del(`/bookmarks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "bookmarks"] }),
  })
}

export interface ProfileUpdate {
  display_name?: string
  bio?: string
  location?: string
  locale?: string
  is_public?: boolean
}

export function useUpdateProfile() {
  return useMutation({
    mutationFn: (body: ProfileUpdate) => api.patch<User>("/me", body),
  })
}

export function useUploadAvatar() {
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData()
      form.append("avatar", file)
      return api.upload<User>("/me/avatar", form)
    },
  })
}

// ---- Admin ----

export interface AdminVideoInput {
  title: string
  description: string
  youtube_id: string
  duration_seconds: number
  difficulty: string
  language: string
  tags: string[]
}

export function useSaveVideo() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: AdminVideoInput }) =>
      id ? api.patch<Video>(`/admin/videos/${id}`, input) : api.post<Video>("/admin/videos", input),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ["videos"] })
      if (vars.id) qc.invalidateQueries({ queryKey: ["video", vars.id] })
    },
  })
}

export function useDeleteVideo() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.del(`/admin/videos/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["videos"] }),
  })
}

export interface AdminArticleInput {
  title: string
  slug: string
  body_markdown: string
  difficulty: string
  language: string
  tags: string[]
}

export function useSaveArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: AdminArticleInput }) =>
      id ? api.patch<Article>(`/admin/articles/${id}`, input) : api.post<Article>("/admin/articles", input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  })
}

export function useDeleteArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.del(`/admin/articles/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  })
}

// ---- Admin: cheatsheets / glossary / projects / tracks / quizzes / problems ----

/** makeAdminCrud builds matching save (create via POST, update via PATCH) and
 * delete hooks for an admin content type, invalidating its public list. */
function adminSave<I>(path: string, listKey: string) {
  return function useSave() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: ({ id, input }: { id?: string; input: I }) =>
        id ? api.patch(`/admin/${path}/${id}`, input) : api.post(`/admin/${path}`, input),
      onSuccess: () => qc.invalidateQueries({ queryKey: [listKey] }),
    })
  }
}
function adminDelete(path: string, listKey: string) {
  return function useDel() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: (id: string) => api.del(`/admin/${path}/${id}`),
      onSuccess: () => qc.invalidateQueries({ queryKey: [listKey] }),
    })
  }
}

export interface AdminCheatsheetInput {
  title: string
  category: string
  body_markdown: string
  language: string
}
export const useSaveCheatsheet = adminSave<AdminCheatsheetInput>("cheatsheets", "cheatsheets")
export const useDeleteCheatsheet = adminDelete("cheatsheets", "cheatsheets")

export interface AdminGlossaryInput {
  term: string
  definition_markdown: string
  language: string
}
export const useSaveGlossary = adminSave<AdminGlossaryInput>("glossary", "glossary")
export const useDeleteGlossary = adminDelete("glossary", "glossary")

export interface AdminProjectInput {
  title: string
  description_markdown: string
  difficulty: string
  language: string
  tags: string[]
  steps: { text: string }[]
}
export const useSaveProject = adminSave<AdminProjectInput>("projects", "projects")
export const useDeleteProject = adminDelete("projects", "projects")

export interface AdminTrackItemInput {
  content_type: string
  content_id: string
}
export interface AdminTrackInput {
  title: string
  description: string
  level: string
  position: number
  language: string
  items: AdminTrackItemInput[]
}
export const useSaveTrack = adminSave<AdminTrackInput>("tracks", "tracks")
export const useDeleteTrack = adminDelete("tracks", "tracks")

export interface AdminQuizOptionInput {
  text: string
  is_correct: boolean
}
export interface AdminQuizQuestionInput {
  prompt: string
  type: string
  options: AdminQuizOptionInput[]
}
export interface AdminQuizInput {
  title: string
  description: string
  pass_threshold: number
  difficulty: string
  language: string
  tags: string[]
  questions: AdminQuizQuestionInput[]
}
export const useSaveQuiz = adminSave<AdminQuizInput>("quizzes", "quizzes")
export const useDeleteQuiz = adminDelete("quizzes", "quizzes")

export interface AdminTestCaseInput {
  input: string
  expected_output: string
  is_sample: boolean
}
export interface AdminProblemInput {
  title: string
  slug: string
  statement_markdown: string
  reference_solution_markdown: string
  difficulty: string
  language: string
  tags: string[]
  sample_io: { input: string; output: string }[]
  test_cases: AdminTestCaseInput[]
}
export const useSaveProblem = adminSave<AdminProblemInput>("problems", "problems")
export const useDeleteProblem = adminDelete("problems", "problems")

export function useAdminUsers(params: { q?: string; limit?: number; offset?: number }) {
  const sp = new URLSearchParams()
  if (params.q) sp.set("q", params.q)
  if (params.limit) sp.set("limit", String(params.limit))
  if (params.offset) sp.set("offset", String(params.offset))
  const qs = sp.toString()
  return useQuery({
    queryKey: ["admin", "users", params],
    queryFn: () => api.get<AdminUserListResponse>("/admin/users" + (qs ? "?" + qs : "")),
  })
}

export function useUpdateAdminUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: { id: string; role?: string; is_blocked?: boolean }) =>
      api.patch(`/admin/users/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "users"] }),
  })
}

export function useRunSandbox() {
  return useMutation({
    mutationFn: (body: { source: string; stdin: string }) =>
      api.post<SandboxRunResult>("/sandbox/run", body),
  })
}
