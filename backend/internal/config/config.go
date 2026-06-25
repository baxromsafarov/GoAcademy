// Package config loads and validates application configuration from environment
// variables. Validation runs once on startup (fail-fast): if anything is wrong,
// Load returns an error listing every problem and the process must not start.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Application environments.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Config holds all runtime configuration for the API service.
//
// Fields are intentionally limited to what CHAPTER 1 needs (server + database +
// logging). Auth/mailer/storage settings are added in their respective chapters.
type Config struct {
	AppEnv      string // development | production
	HTTPHost    string
	HTTPPort    int
	LogLevel    string // debug | info | warn | error
	LogFormat   string // text | json
	DatabaseURL string // postgres connection string

	EmailVerificationTTL time.Duration // lifetime of email verification tokens
	PasswordResetTTL     time.Duration // lifetime of password reset tokens

	AuthRateLimitPerMinute int // per-IP request budget for auth endpoints

	JWTSecret     string        // HS256 signing secret for access tokens
	JWTAccessTTL  time.Duration // access token lifetime (short)
	JWTRefreshTTL time.Duration // refresh token / session lifetime

	CookieDomain   string // domain for the refresh cookie ("" = host-only)
	CookieSecure   bool   // Secure flag (true in production/HTTPS)
	CookieSameSite string // lax | strict | none

	CORSAllowedOrigins []string // browser origins allowed cross-origin (with credentials)

	StorageDriver        string // local (s3 later)
	StorageLocalDir      string // filesystem dir for local storage
	StoragePublicBaseURL string // public base URL for stored objects

	SandboxEnabled            bool   // enable the code-runner sandbox (needs Docker + Go toolchain)
	SandboxImage              string // minimal run image for the sandbox (e.g. busybox)
	SandboxWorkDir            string // host dir for transient build dirs ("" = os temp)
	SandboxRateLimitPerMinute int    // per-IP budget for /sandbox/run
}

var validLogLevels = map[string]struct{}{
	"debug": {}, "info": {}, "warn": {}, "error": {},
}

// Load reads configuration from the environment, applies defaults, and validates
// the result. The returned error aggregates all problems found so a misconfigured
// deployment surfaces every issue at once instead of one restart at a time.
func Load() (*Config, error) {
	port, err := getEnvInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}

	emailVerificationTTL, err := getEnvDuration("EMAIL_VERIFICATION_TTL", 24*time.Hour)
	if err != nil {
		return nil, err
	}
	passwordResetTTL, err := getEnvDuration("PASSWORD_RESET_TTL", time.Hour)
	if err != nil {
		return nil, err
	}
	authRateLimit, err := getEnvInt("RATE_LIMIT_AUTH_PER_MINUTE", 10)
	if err != nil {
		return nil, err
	}
	jwtAccessTTL, err := getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	jwtRefreshTTL, err := getEnvDuration("JWT_REFRESH_TTL", 720*time.Hour)
	if err != nil {
		return nil, err
	}
	cookieSecure, err := getEnvBool("COOKIE_SECURE", false)
	if err != nil {
		return nil, err
	}
	sandboxEnabled, err := getEnvBool("SANDBOX_ENABLED", false)
	if err != nil {
		return nil, err
	}
	sandboxRateLimit, err := getEnvInt("RATE_LIMIT_SANDBOX_PER_MINUTE", 6)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", EnvDevelopment),
		HTTPHost:               getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:               port,
		LogLevel:               strings.ToLower(getEnv("LOG_LEVEL", "info")),
		LogFormat:              strings.ToLower(getEnv("LOG_FORMAT", "json")),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		EmailVerificationTTL:   emailVerificationTTL,
		PasswordResetTTL:       passwordResetTTL,
		AuthRateLimitPerMinute: authRateLimit,
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTAccessTTL:           jwtAccessTTL,
		JWTRefreshTTL:          jwtRefreshTTL,
		CookieDomain:           getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:           cookieSecure,
		CookieSameSite:         strings.ToLower(getEnv("COOKIE_SAMESITE", "lax")),
		CORSAllowedOrigins:     splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		StorageDriver:          strings.ToLower(getEnv("STORAGE_DRIVER", "local")),
		StorageLocalDir:        getEnv("STORAGE_LOCAL_DIR", "./storage"),
		StoragePublicBaseURL:   getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080/static"),

		SandboxEnabled:            sandboxEnabled,
		SandboxImage:              getEnv("SANDBOX_IMAGE", "busybox"),
		SandboxWorkDir:            getEnv("SANDBOX_WORK_DIR", ""),
		SandboxRateLimitPerMinute: sandboxRateLimit,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// IsProduction reports whether the service runs in the production environment.
func (c *Config) IsProduction() bool { return c.AppEnv == EnvProduction }

// HTTPAddr returns the host:port the HTTP server should listen on.
func (c *Config) HTTPAddr() string {
	return net.JoinHostPort(c.HTTPHost, strconv.Itoa(c.HTTPPort))
}

func (c *Config) validate() error {
	var problems []string

	if c.AppEnv != EnvDevelopment && c.AppEnv != EnvProduction {
		problems = append(problems, fmt.Sprintf("APP_ENV must be %q or %q, got %q", EnvDevelopment, EnvProduction, c.AppEnv))
	}
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		problems = append(problems, fmt.Sprintf("HTTP_PORT must be in 1..65535, got %d", c.HTTPPort))
	}
	if _, ok := validLogLevels[c.LogLevel]; !ok {
		problems = append(problems, fmt.Sprintf("LOG_LEVEL must be one of debug|info|warn|error, got %q", c.LogLevel))
	}
	if c.LogFormat != "text" && c.LogFormat != "json" {
		problems = append(problems, fmt.Sprintf("LOG_FORMAT must be text|json, got %q", c.LogFormat))
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		problems = append(problems, "DATABASE_URL is required")
	} else if u, err := url.Parse(c.DatabaseURL); err != nil {
		problems = append(problems, fmt.Sprintf("DATABASE_URL is not a valid URL: %v", err))
	} else if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		problems = append(problems, fmt.Sprintf("DATABASE_URL must use the postgres scheme, got %q", u.Scheme))
	}
	if c.EmailVerificationTTL <= 0 {
		problems = append(problems, fmt.Sprintf("EMAIL_VERIFICATION_TTL must be positive, got %s", c.EmailVerificationTTL))
	}
	if c.PasswordResetTTL <= 0 {
		problems = append(problems, fmt.Sprintf("PASSWORD_RESET_TTL must be positive, got %s", c.PasswordResetTTL))
	}
	if c.AuthRateLimitPerMinute <= 0 {
		problems = append(problems, fmt.Sprintf("RATE_LIMIT_AUTH_PER_MINUTE must be positive, got %d", c.AuthRateLimitPerMinute))
	}
	if len(c.JWTSecret) < 32 {
		problems = append(problems, "JWT_SECRET is required and must be at least 32 characters")
	}
	if c.JWTAccessTTL <= 0 {
		problems = append(problems, fmt.Sprintf("JWT_ACCESS_TTL must be positive, got %s", c.JWTAccessTTL))
	}
	if c.JWTRefreshTTL <= 0 {
		problems = append(problems, fmt.Sprintf("JWT_REFRESH_TTL must be positive, got %s", c.JWTRefreshTTL))
	}
	if c.JWTAccessTTL > 0 && c.JWTRefreshTTL > 0 && c.JWTRefreshTTL <= c.JWTAccessTTL {
		problems = append(problems, "JWT_REFRESH_TTL must be greater than JWT_ACCESS_TTL")
	}
	switch c.CookieSameSite {
	case "lax", "strict", "none":
	default:
		problems = append(problems, fmt.Sprintf("COOKIE_SAMESITE must be lax|strict|none, got %q", c.CookieSameSite))
	}
	if c.StorageDriver != "local" {
		problems = append(problems, fmt.Sprintf("STORAGE_DRIVER must be local (only supported driver), got %q", c.StorageDriver))
	}
	if strings.TrimSpace(c.StorageLocalDir) == "" {
		problems = append(problems, "STORAGE_LOCAL_DIR is required")
	}
	if strings.TrimSpace(c.StoragePublicBaseURL) == "" {
		problems = append(problems, "STORAGE_PUBLIC_BASE_URL is required")
	}
	if c.SandboxEnabled {
		if strings.TrimSpace(c.SandboxImage) == "" {
			problems = append(problems, "SANDBOX_IMAGE is required when SANDBOX_ENABLED")
		}
		if c.SandboxRateLimitPerMinute <= 0 {
			problems = append(problems, fmt.Sprintf("RATE_LIMIT_SANDBOX_PER_MINUTE must be positive, got %d", c.SandboxRateLimitPerMinute))
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

// splitCSV splits a comma-separated value into trimmed, non-empty items.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// getEnv returns the value of key, or fallback when the variable is unset or empty.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getEnvInt parses key as an int, or returns fallback when the variable is unset
// or empty. A present-but-non-numeric value is a hard error.
func getEnvInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q", key, v)
	}
	return n, nil
}

// getEnvDuration parses key as a Go duration (e.g. "24h", "15m"), or returns
// fallback when the variable is unset or empty.
func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration (e.g. 24h), got %q", key, v)
	}
	return d, nil
}

// getEnvBool parses key as a bool ("true"/"false"/"1"/"0"), or returns fallback
// when the variable is unset or empty.
func getEnvBool(key string, fallback bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean, got %q", key, v)
	}
	return b, nil
}
