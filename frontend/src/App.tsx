import { BrowserRouter, Routes, Route } from "react-router-dom"
import { Layout } from "@/components/Layout"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import { Dashboard } from "@/pages/Dashboard"
import { Placeholder } from "@/pages/Placeholder"
import { VideoList } from "@/pages/videos/VideoList"
import { VideoDetail } from "@/pages/videos/VideoDetail"
import { ArticleList } from "@/pages/articles/ArticleList"
import { ArticleDetail } from "@/pages/articles/ArticleDetail"
import { QuizList } from "@/pages/quizzes/QuizList"
import { QuizDetail } from "@/pages/quizzes/QuizDetail"
import { ProblemList } from "@/pages/problems/ProblemList"
import { ProblemDetail } from "@/pages/problems/ProblemDetail"
import { TrackList } from "@/pages/tracks/TrackList"
import { TrackDetail } from "@/pages/tracks/TrackDetail"
import { CheatsheetList } from "@/pages/cheatsheets/CheatsheetList"
import { CheatsheetDetail } from "@/pages/cheatsheets/CheatsheetDetail"
import { Glossary } from "@/pages/glossary/Glossary"
import { ProjectList } from "@/pages/projects/ProjectList"
import { ProjectDetail } from "@/pages/projects/ProjectDetail"
import { Sandbox } from "@/pages/sandbox/Sandbox"
import { Leaderboard } from "@/pages/leaderboard/Leaderboard"
import { MyNotes } from "@/pages/notes/MyNotes"
import { MyBookmarks } from "@/pages/bookmarks/MyBookmarks"
import { Profile } from "@/pages/profile/Profile"
import { AdminRoute } from "@/components/AdminRoute"
import { AdminHome } from "@/pages/admin/AdminHome"
import { AdminVideos } from "@/pages/admin/AdminVideos"
import { VideoForm } from "@/pages/admin/VideoForm"
import { AdminArticles } from "@/pages/admin/AdminArticles"
import { ArticleForm } from "@/pages/admin/ArticleForm"
import { AdminUsers } from "@/pages/admin/AdminUsers"
import { Login } from "@/pages/auth/Login"
import { Register } from "@/pages/auth/Register"
import { VerifyEmail } from "@/pages/auth/VerifyEmail"
import { ForgotPassword } from "@/pages/auth/ForgotPassword"
import { ResetPassword } from "@/pages/auth/ResetPassword"

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public auth routes */}
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/verify-email" element={<VerifyEmail />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />

        {/* Authenticated app */}
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="videos" element={<VideoList />} />
            <Route path="videos/:id" element={<VideoDetail />} />
            <Route path="articles" element={<ArticleList />} />
            <Route path="articles/:slug" element={<ArticleDetail />} />
            <Route path="quizzes" element={<QuizList />} />
            <Route path="quizzes/:id" element={<QuizDetail />} />
            <Route path="problems" element={<ProblemList />} />
            <Route path="problems/:slug" element={<ProblemDetail />} />
            <Route path="tracks" element={<TrackList />} />
            <Route path="tracks/:id" element={<TrackDetail />} />
            <Route path="sandbox" element={<Sandbox />} />
            <Route path="cheatsheets" element={<CheatsheetList />} />
            <Route path="cheatsheets/:id" element={<CheatsheetDetail />} />
            <Route path="projects" element={<ProjectList />} />
            <Route path="projects/:id" element={<ProjectDetail />} />
            <Route path="glossary" element={<Glossary />} />
            <Route path="leaderboard" element={<Leaderboard />} />
            <Route path="notes" element={<MyNotes />} />
            <Route path="bookmarks" element={<MyBookmarks />} />
            <Route path="profile" element={<Profile />} />

            {/* Admin-only section (gated client-side + enforced server-side) */}
            <Route element={<AdminRoute />}>
              <Route path="admin" element={<AdminHome />} />
              <Route path="admin/videos" element={<AdminVideos />} />
              <Route path="admin/videos/new" element={<VideoForm />} />
              <Route path="admin/videos/:id/edit" element={<VideoForm />} />
              <Route path="admin/articles" element={<AdminArticles />} />
              <Route path="admin/articles/new" element={<ArticleForm />} />
              <Route path="admin/articles/:slug/edit" element={<ArticleForm />} />
              <Route path="admin/users" element={<AdminUsers />} />
            </Route>
            <Route
              path="*"
              element={<Placeholder titleKey="common.notFoundTitle" descriptionKey="common.notFound" />}
            />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
