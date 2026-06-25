// Command api is the GoAcademy HTTP API service entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/auth"
	"github.com/goacademy/backend/internal/config"
	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/gamification"
	"github.com/goacademy/backend/internal/httpapi"
	"github.com/goacademy/backend/internal/judge"
	"github.com/goacademy/backend/internal/mailer"
	"github.com/goacademy/backend/internal/platform/logging"
	"github.com/goacademy/backend/internal/platform/postgres"
	"github.com/goacademy/backend/internal/platform/storage"
	"github.com/goacademy/backend/internal/progress"
	"github.com/goacademy/backend/internal/quiz"
	"github.com/goacademy/backend/internal/runner"
	"github.com/goacademy/backend/internal/social"
	"github.com/goacademy/backend/internal/user"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel, cfg.LogFormat)

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelStartup()

	pool, err := postgres.Connect(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to postgres")

	dbReady := httpapi.Check{Name: "postgres", Func: pool.Ping}

	stubMailer := mailer.NewLogMailer(logger)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	authService := auth.NewService(pool, stubMailer, tokenManager, logger, auth.TTLConfig{
		EmailVerification: cfg.EmailVerificationTTL,
		Refresh:           cfg.JWTRefreshTTL,
		PasswordReset:     cfg.PasswordResetTTL,
	})
	avatarStorage, err := storage.NewLocalStorage(cfg.StorageLocalDir, cfg.StoragePublicBaseURL)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	userService := user.NewService(pool, avatarStorage)
	contentService := content.NewService(pool)
	activityRecorder := gamification.NewRecorder(pool)
	gamificationService := gamification.NewService(pool)
	dailyService := gamification.NewDailyService(pool, activityRecorder)
	leaderboardService := social.NewService(pool)
	notesService := social.NewNotesService(pool)
	bookmarksService := social.NewBookmarksService(pool)
	progressService := progress.NewService(pool, activityRecorder)
	certificatesService := social.NewCertificatesService(pool, progressService)
	adminService := admin.NewService(pool)
	quizService := quiz.NewService(pool, contentService, activityRecorder)

	// The code sandbox + online judge are opt-in (they need Docker + a Go
	// toolchain on the host) and share the same runner.
	var sandboxRunner *runner.Runner
	var judgeService *judge.Service
	if cfg.SandboxEnabled {
		sandboxRunner = runner.New(cfg.SandboxImage, cfg.SandboxWorkDir)
		judgeService = judge.NewService(sandboxRunner, pool, activityRecorder, runner.Limits{})
		logger.Info("code sandbox + online judge enabled", "image", cfg.SandboxImage)
	}

	srv := &http.Server{
		Addr: cfg.HTTPAddr(),
		Handler: httpapi.NewRouter(httpapi.Deps{
			Logger:       logger,
			Auth:         authService,
			User:         userService,
			Content:      contentService,
			Progress:     progressService,
			Quiz:         quizService,
			Gamification: gamificationService,
			Daily:        dailyService,
			Leaderboard:  leaderboardService,
			Notes:        notesService,
			Bookmarks:    bookmarksService,
			Certificates: certificatesService,
			Admin:        adminService,
			Runner:       sandboxRunner,
			Judge:        judgeService,
			Tokens:       tokenManager,
			Cookie: httpapi.CookieConfig{
				Domain:   cfg.CookieDomain,
				Secure:   cfg.CookieSecure,
				SameSite: sameSite(cfg.CookieSameSite),
			},
			AuthRateLimitPerMinute:    cfg.AuthRateLimitPerMinute,
			SandboxRateLimitPerMinute: cfg.SandboxRateLimitPerMinute,
			CORSAllowedOrigins:        cfg.CORSAllowedOrigins,
			StaticDir:                 cfg.StorageLocalDir,
			ReadyChecks:               []httpapi.Check{dbReady},
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the listener; a non-graceful failure is reported on this channel.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", cfg.HTTPAddr(), "app_env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Block until either the server dies or we receive a termination signal.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server failed: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server stopped cleanly")
	return nil
}

// sameSite maps the configured cookie SameSite policy to its net/http value.
func sameSite(mode string) http.SameSite {
	switch mode {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
