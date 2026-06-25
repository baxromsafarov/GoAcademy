export interface User {
  id: string
  email: string
  display_name: string
  role: "student" | "admin"
  locale: string
  bio: string
  location: string
  avatar_url: string
  email_verified: boolean
  is_public: boolean
  created_at: string
}

export interface Stats {
  total_xp: number
  level: number
  current_streak: number
  longest_streak: number
  last_active_date: string
}

export interface ProgressSummary {
  videos_completed: number
  articles_read: number
  quizzes_passed: number
  problems_solved: number
  projects_completed: number
}

export interface Badge {
  code: string
  title: string
  description: string
  icon: string
  awarded_at: string
}

export interface BadgesResponse {
  badges: Badge[]
}

export interface ActivityDay {
  day: string
  count: number
  xp: number
}

export interface ActivityResponse {
  from: string
  to: string
  days: ActivityDay[]
}

export interface Video {
  id: string
  title: string
  description: string
  youtube_id: string
  duration_seconds: number
  difficulty: string
  tags: string[]
  language: string
  created_at: string
  updated_at: string
}

export interface VideoListResponse {
  items: Video[]
  total: number
  limit: number
  offset: number
}

export interface VideoProgress {
  video_id: string
  watched_percent: number
  last_position_seconds: number
  completed: boolean
  updated_at: string
}

export interface Article {
  id: string
  title: string
  slug: string
  body_markdown: string
  difficulty: string
  tags: string[]
  language: string
  created_at: string
  updated_at: string
}

export interface ArticleListItem {
  id: string
  title: string
  slug: string
  difficulty: string
  tags: string[]
  language: string
  created_at: string
}

export interface ArticleListResponse {
  items: ArticleListItem[]
  total: number
  limit: number
  offset: number
}

export interface ArticleReadStatus {
  read: boolean
  completed_at?: string
}

export interface QuizListItem {
  id: string
  title: string
  description: string
  pass_threshold: number
  difficulty: string
  tags: string[]
  language: string
  created_at: string
}

export interface QuizListResponse {
  items: QuizListItem[]
  total: number
  limit: number
  offset: number
}

export interface QuizOption {
  id: string
  text: string
}

export interface QuizQuestion {
  id: string
  prompt: string
  type: "single" | "multiple"
  options: QuizOption[]
}

export interface QuizDetail {
  id: string
  title: string
  description: string
  pass_threshold: number
  difficulty: string
  tags: string[]
  language: string
  questions: QuizQuestion[]
}

export interface QuizQuestionReview {
  question_id: string
  correct: boolean
  correct_option_ids: string[]
}

export interface QuizAttemptResult {
  attempt_id: string
  score: number
  passed: boolean
  created_at: string
  review: QuizQuestionReview[]
}

export interface ProblemListItem {
  id: string
  title: string
  slug: string
  difficulty: string
  tags: string[]
  language: string
  created_at: string
}

export interface ProblemListResponse {
  items: ProblemListItem[]
  total: number
  limit: number
  offset: number
}

export interface ProblemSample {
  input?: string
  output?: string
  [key: string]: unknown
}

export interface ProblemDetail {
  id: string
  title: string
  slug: string
  statement_markdown: string
  difficulty: string
  tags: string[]
  language: string
  sample_io: ProblemSample[]
  created_at: string
  updated_at: string
}

export interface JudgeCaseResult {
  index: number
  is_sample: boolean
  verdict: string
  duration_ms: number
}

export interface JudgeVerdict {
  verdict: "OK" | "WA" | "TLE" | "RE" | "CE"
  passed: number
  total: number
  compile_error?: string
  cases: JudgeCaseResult[]
}

export interface ProblemSubmissionResult {
  id: string
  status: "attempted" | "solved"
  language: string
  created_at: string
  verdict?: JudgeVerdict
  reference_solution_markdown?: string
}

export type TrackContentType = "video" | "article" | "quiz" | "problem" | "project"

export interface TrackListItem {
  id: string
  title: string
  description: string
  level: string
  language: string
  position: number
  created_at: string
}

export interface TrackListResponse {
  items: TrackListItem[]
  total: number
  limit: number
  offset: number
}

export interface TrackItem {
  content_type: TrackContentType
  content_id: string
  position: number
}

export interface TrackDetail {
  id: string
  title: string
  description: string
  level: string
  language: string
  position: number
  items: TrackItem[]
}

export interface TrackItemProgress {
  content_type: TrackContentType
  content_id: string
  position: number
  completed: boolean
}

export interface TrackProgress {
  track_id: string
  total: number
  completed: number
  percent: number
  track_complete: boolean
  items: TrackItemProgress[]
}

export interface CheatsheetListItem {
  id: string
  title: string
  category: string
  language: string
  created_at: string
}

export interface CheatsheetListResponse {
  items: CheatsheetListItem[]
  total: number
  limit: number
  offset: number
}

export interface CheatsheetDetail {
  id: string
  title: string
  category: string
  body_markdown: string
  language: string
  created_at: string
  updated_at: string
}

export interface GlossaryItem {
  id: string
  term: string
  definition_markdown: string
  language: string
}

export interface GlossaryListResponse {
  items: GlossaryItem[]
  total: number
  limit: number
  offset: number
}

export interface ProjectListItem {
  id: string
  title: string
  difficulty: string
  tags: string[]
  language: string
  created_at: string
}

export interface ProjectListResponse {
  items: ProjectListItem[]
  total: number
  limit: number
  offset: number
}

export interface ProjectStep {
  id: string
  text: string
  position: number
}

export interface ProjectDetail {
  id: string
  title: string
  description_markdown: string
  difficulty: string
  tags: string[]
  language: string
  steps: ProjectStep[]
}

export interface ProjectProgress {
  project_id: string
  completed_step_ids: string[]
  total: number
  completed: number
  project_complete: boolean
}

export interface LeaderboardEntry {
  rank: number
  user_id: string
  display_name: string
  avatar_url: string
  xp: number
}

export interface LeaderboardResponse {
  period: string
  entries: LeaderboardEntry[]
}

export interface Note {
  id: string
  content_type: string
  content_id: string
  body: string
  created_at: string
  updated_at: string
}

export interface NotesResponse {
  notes: Note[]
}

export interface Bookmark {
  id: string
  content_type: string
  content_id: string
  created_at: string
}

export interface BookmarksResponse {
  bookmarks: Bookmark[]
}

export interface AdminUser {
  id: string
  email: string
  display_name: string
  role: "student" | "admin"
  is_blocked: boolean
  email_verified: boolean
  locale: string
  created_at: string
}

export interface AdminUserListResponse {
  items: AdminUser[]
  total: number
  limit: number
  offset: number
}

export interface SandboxRunResult {
  stdout: string
  stderr: string
  exit_code: number
  compile_error: boolean
  timed_out: boolean
  oom_killed: boolean
  stdout_truncated: boolean
  stderr_truncated: boolean
  duration_ms: number
}
