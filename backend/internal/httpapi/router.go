// Package httpapi wires the HTTP router, middleware and handlers for the
// GoAcademy API. It is named httpapi (not http) to avoid clashing with the
// standard library net/http package.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/auth"
	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/gamification"
	"github.com/goacademy/backend/internal/judge"
	"github.com/goacademy/backend/internal/progress"
	"github.com/goacademy/backend/internal/quiz"
	"github.com/goacademy/backend/internal/runner"
	"github.com/goacademy/backend/internal/social"
	"github.com/goacademy/backend/internal/user"
)

// Deps are the dependencies the router needs to serve all routes.
type Deps struct {
	Logger                    *slog.Logger
	Auth                      *auth.Service               // when nil, /api/v1/auth routes are not mounted (e.g. in health-only tests)
	User                      *user.Service               // when nil, /api/v1/me routes are not mounted
	Content                   *content.Service            // when nil, content read routes are not mounted
	Progress                  *progress.Service           // when nil, progress routes are not mounted
	Quiz                      *quiz.Service               // when nil, quiz attempt route is not mounted
	Gamification              *gamification.Service       // when nil, /api/v1/me/stats and /me/badges are not mounted
	Daily                     *gamification.DailyService  // when nil, /api/v1/daily-challenge is not mounted
	Leaderboard               *social.Service             // when nil, /api/v1/leaderboard is not mounted
	Notes                     *social.NotesService        // when nil, notes routes are not mounted
	Bookmarks                 *social.BookmarksService    // when nil, bookmarks routes are not mounted
	Certificates              *social.CertificatesService // when nil, certificate routes are not mounted
	Admin                     *admin.Service              // when nil, /admin routes are not mounted
	Runner                    *runner.Runner              // when nil, /sandbox/run is not mounted
	Judge                     *judge.Service              // when nil, problem submissions use manual marking
	Tokens                    *auth.TokenManager          // used by RequireAuth for protected routes
	Cookie                    CookieConfig
	AuthRateLimitPerMinute    int      // per-IP rate limit on auth endpoints (0 = disabled)
	SandboxRateLimitPerMinute int      // per-IP rate limit on /sandbox/run (0 = disabled)
	CORSAllowedOrigins        []string // browser origins allowed cross-origin (empty = CORS off)
	StaticDir                 string   // local dir served at /static/ (empty = not mounted)
	ReadyChecks               []Check
}

// NewRouter builds the application router: base middleware, health endpoints at
// the root, and the versioned API under /api/v1.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(securityHeaders)
	if len(deps.CORSAllowedOrigins) > 0 {
		r.Use(cors(deps.CORSAllowedOrigins))
	}
	r.Use(middleware.RequestID)
	r.Use(requestIDHeader)
	// NB: chi's RealIP is intentionally NOT used — it trusts client-supplied
	// X-Forwarded-For / X-Real-IP headers, which lets a client spoof its source
	// IP and so bypass the per-IP rate limiters. The rate limiter therefore keys
	// on the real TCP peer (r.RemoteAddr). Behind a trusted reverse proxy, a
	// proxy-aware real-IP extractor should be configured in the deployment.
	r.Use(recoverer(deps.Logger))
	r.Use(requestLogger(deps.Logger))

	// Liveness/readiness live at the root (not versioned).
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(deps.ReadyChecks))

	// Serve uploaded files (avatars, ...) from local storage.
	if deps.StaticDir != "" {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(deps.StaticDir))))
	}

	r.Route("/api/v1", func(r chi.Router) {
		if deps.Auth != nil {
			ah := newAuthHandler(deps.Auth, deps.Cookie, deps.Logger)
			r.Route("/auth", func(r chi.Router) {
				if deps.AuthRateLimitPerMinute > 0 {
					r.Use(RateLimit(deps.AuthRateLimitPerMinute))
				}
				r.Post("/register", ah.register)
				r.Post("/verify-email", ah.verifyEmail)
				r.Post("/login", ah.login)
				r.Post("/refresh", ah.refresh)
				r.Post("/logout", ah.logout)
				r.Post("/forgot-password", ah.forgotPassword)
				r.Post("/reset-password", ah.resetPassword)
			})
		}

		// Content read routes (public) + authenticated progress write.
		if deps.Content != nil {
			vh := newVideoHandler(deps.Content, deps.Progress, deps.Logger)
			r.Get("/videos", vh.list)
			r.Get("/videos/{id}", vh.get)
			if deps.Progress != nil && deps.Tokens != nil {
				r.Group(func(r chi.Router) {
					r.Use(RequireAuth(deps.Tokens, deps.Logger))
					r.Post("/videos/{id}/progress", vh.postProgress)
					r.Get("/videos/{id}/progress", vh.getProgress)
				})
			}

			ah := newArticleHandler(deps.Content, deps.Progress, deps.Logger)
			r.Get("/articles", ah.list)
			r.Get("/articles/{slug}", ah.get)
			if deps.Progress != nil && deps.Tokens != nil {
				r.Group(func(r chi.Router) {
					r.Use(RequireAuth(deps.Tokens, deps.Logger))
					r.Post("/articles/{slug}/complete", ah.complete)
					r.Get("/articles/{slug}/read", ah.readStatus)
				})
			}

			qh := newQuizHandler(deps.Content, deps.Quiz, deps.Logger)
			r.Get("/quizzes", qh.list)
			r.Get("/quizzes/{id}", qh.get)
			if deps.Quiz != nil && deps.Tokens != nil {
				r.With(RequireAuth(deps.Tokens, deps.Logger)).Post("/quizzes/{id}/attempts", qh.submit)
			}

			ph := newProblemHandler(deps.Content, deps.Progress, deps.Judge, deps.Logger)
			r.Get("/problems", ph.list)
			r.Get("/problems/{slug}", ph.get)
			if deps.Progress != nil && deps.Tokens != nil {
				r.Group(func(r chi.Router) {
					r.Use(RequireAuth(deps.Tokens, deps.Logger))
					r.Post("/problems/{slug}/submissions", ph.submit)
					r.Get("/problems/{slug}/solution", ph.getSolution)
				})
			}

			th := newTrackHandler(deps.Content, deps.Progress, deps.Logger)
			r.Get("/tracks", th.list)
			r.Get("/tracks/{id}", th.get)
			if deps.Progress != nil && deps.Tokens != nil {
				r.With(RequireAuth(deps.Tokens, deps.Logger)).Get("/tracks/{id}/progress", th.getProgress)
			}
			if deps.Tokens != nil {
				auth := RequireAuth(deps.Tokens, deps.Logger)
				r.With(auth).Post("/tracks/{id}/enroll", th.enroll)
				r.With(auth).Delete("/tracks/{id}/enroll", th.unenroll)
				r.With(auth).Get("/me/tracks", th.myTracks)
				r.With(auth).Get("/me/recent", th.recentCompletions)
			}

			rh := newReferenceHandler(deps.Content, deps.Logger)
			r.Get("/cheatsheets", rh.listCheatsheets)
			r.Get("/cheatsheets/{id}", rh.getCheatsheet)
			r.Get("/glossary", rh.listGlossary)

			projh := newProjectHandler(deps.Content, deps.Progress, deps.Logger)
			r.Get("/projects", projh.list)
			r.Get("/projects/{id}", projh.get)
			if deps.Progress != nil && deps.Tokens != nil {
				r.Group(func(r chi.Router) {
					r.Use(RequireAuth(deps.Tokens, deps.Logger))
					r.Get("/projects/{id}/progress", projh.getProgress)
					r.Post("/projects/{id}/steps/{stepId}/toggle", projh.toggleStep)
				})
			}
		}

		// Authenticated routes.
		if deps.User != nil && deps.Tokens != nil {
			mh := newMeHandler(deps.User, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Get("/me", mh.get)
				r.Patch("/me", mh.patch)
				r.Post("/me/avatar", mh.uploadAvatar)
			})
		}

		// Authenticated progress summary + activity heatmap.
		if deps.Progress != nil && deps.Tokens != nil {
			mph := newMeProgressHandler(deps.Progress, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Get("/me/progress", mph.summary)
				r.Get("/me/activity", mph.activity)
			})
		}

		// Authenticated gamification stats (XP, level, streaks) and badges.
		if deps.Gamification != nil && deps.Tokens != nil {
			sh := newMeStatsHandler(deps.Gamification, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Get("/me/stats", sh.get)
				r.Get("/me/badges", sh.badges)
			})
		}

		// Authenticated daily challenge.
		if deps.Daily != nil && deps.Tokens != nil {
			dh := newDailyHandler(deps.Daily, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Get("/daily-challenge", dh.get)
				r.Post("/daily-challenge/complete", dh.complete)
			})
		}

		// Public leaderboard (only public users are listed).
		if deps.Leaderboard != nil {
			lh := newLeaderboardHandler(deps.Leaderboard, deps.Logger)
			r.Get("/leaderboard", lh.get)
		}

		// Code sandbox (authenticated + rate-limited): runs untrusted Go.
		if deps.Runner != nil && deps.Tokens != nil {
			sh := newSandboxHandler(deps.Runner, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				if deps.SandboxRateLimitPerMinute > 0 {
					r.Use(RateLimit(deps.SandboxRateLimitPerMinute))
				}
				r.Post("/sandbox/run", sh.run)
			})
		}

		// Authenticated personal notes (owner-only).
		if deps.Notes != nil && deps.Tokens != nil {
			nh := newNotesHandler(deps.Notes, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Post("/notes", nh.create)
				r.Patch("/notes/{id}", nh.update)
				r.Delete("/notes/{id}", nh.delete)
				r.Get("/me/notes", nh.list)
			})
		}

		// Authenticated bookmarks (owner-only).
		if deps.Bookmarks != nil && deps.Tokens != nil {
			bh := newBookmarksHandler(deps.Bookmarks, deps.Logger)
			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Post("/bookmarks", bh.create)
				r.Delete("/bookmarks/{id}", bh.delete)
				r.Get("/me/bookmarks", bh.list)
			})
		}

		// Certificates: issue/list require auth; verification by code is public.
		if deps.Certificates != nil {
			ch := newCertificatesHandler(deps.Certificates, deps.Logger)
			r.Get("/certificates/{code}", ch.verify)
			if deps.Tokens != nil {
				r.Group(func(r chi.Router) {
					r.Use(RequireAuth(deps.Tokens, deps.Logger))
					r.Post("/tracks/{id}/certificate", ch.issue)
					r.Get("/me/certificates", ch.list)
				})
			}
		}

		// Admin content CRUD (admin role only).
		if deps.Admin != nil && deps.Tokens != nil {
			adm := newAdminHandler(deps.Admin, deps.Logger)
			r.Route("/admin", func(r chi.Router) {
				r.Use(RequireAuth(deps.Tokens, deps.Logger))
				r.Use(RequireRole("admin"))

				r.Post("/videos", adm.createVideo)
				r.Patch("/videos/{id}", adm.updateVideo)
				r.Delete("/videos/{id}", adm.deleteVideo)

				r.Post("/articles", adm.createArticle)
				r.Patch("/articles/{id}", adm.updateArticle)
				r.Delete("/articles/{id}", adm.deleteArticle)

				r.Post("/quizzes", adm.createQuiz)
				r.Patch("/quizzes/{id}", adm.updateQuiz)
				r.Delete("/quizzes/{id}", adm.deleteQuiz)

				r.Post("/problems", adm.createProblem)
				r.Patch("/problems/{id}", adm.updateProblem)
				r.Delete("/problems/{id}", adm.deleteProblem)

				r.Post("/tracks", adm.createTrack)
				r.Patch("/tracks/{id}", adm.updateTrack)
				r.Delete("/tracks/{id}", adm.deleteTrack)

				r.Post("/cheatsheets", adm.createCheatsheet)
				r.Patch("/cheatsheets/{id}", adm.updateCheatsheet)
				r.Delete("/cheatsheets/{id}", adm.deleteCheatsheet)

				r.Post("/projects", adm.createProject)
				r.Patch("/projects/{id}", adm.updateProject)
				r.Delete("/projects/{id}", adm.deleteProject)

				r.Post("/glossary", adm.createGlossary)
				r.Patch("/glossary/{id}", adm.updateGlossary)
				r.Delete("/glossary/{id}", adm.deleteGlossary)

				r.Post("/badges", adm.createBadge)
				r.Patch("/badges/{id}", adm.updateBadge)
				r.Delete("/badges/{id}", adm.deleteBadge)

				r.Post("/daily-challenges", adm.createDailyChallenge)
				r.Patch("/daily-challenges/{id}", adm.updateDailyChallenge)
				r.Delete("/daily-challenges/{id}", adm.deleteDailyChallenge)

				r.Get("/users", adm.listUsers)
				r.Patch("/users/{id}", adm.updateUser)
			})
		}
	})

	return r
}
