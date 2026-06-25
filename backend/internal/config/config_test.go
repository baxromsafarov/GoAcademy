package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	validDBURL    = "postgres://user:pass@localhost:5432/goacademy?sslmode=disable"
	testJWTSecret = "0123456789abcdef0123456789abcdef" // 32 chars
)

// setBaseEnv sets a complete, valid environment so individual cases can override
// just the variable under test without leaking values from the host environment.
func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", EnvDevelopment)
	t.Setenv("HTTP_HOST", "0.0.0.0")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("DATABASE_URL", validDBURL)
	t.Setenv("JWT_SECRET", testJWTSecret)
}

func TestLoad_ValidWithExplicitValues(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("LOG_LEVEL", "DEBUG") // case-insensitive

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.AppEnv != EnvDevelopment {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, EnvDevelopment)
	}
	if cfg.HTTPPort != 9000 {
		t.Errorf("HTTPPort = %d, want 9000", cfg.HTTPPort)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q (lowercased)", cfg.LogLevel, "debug")
	}
	if got, want := cfg.HTTPAddr(), "0.0.0.0:9000"; got != want {
		t.Errorf("HTTPAddr() = %q, want %q", got, want)
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true, want false for development")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	// Only the required variables are set; everything else must fall back to defaults.
	for _, k := range []string{
		"APP_ENV", "HTTP_HOST", "HTTP_PORT", "LOG_LEVEL", "LOG_FORMAT",
		"EMAIL_VERIFICATION_TTL", "PASSWORD_RESET_TTL", "RATE_LIMIT_AUTH_PER_MINUTE",
		"JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"COOKIE_DOMAIN", "COOKIE_SECURE", "COOKIE_SAMESITE",
		"STORAGE_DRIVER", "STORAGE_LOCAL_DIR", "STORAGE_PUBLIC_BASE_URL",
		"SANDBOX_ENABLED", "SANDBOX_IMAGE", "SANDBOX_WORK_DIR", "RATE_LIMIT_SANDBOX_PER_MINUTE",
		"CORS_ALLOWED_ORIGINS",
	} {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", validDBURL)
	t.Setenv("JWT_SECRET", testJWTSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	want := Config{
		AppEnv:                 EnvDevelopment,
		HTTPHost:               "0.0.0.0",
		HTTPPort:               8080,
		LogLevel:               "info",
		LogFormat:              "json",
		DatabaseURL:            validDBURL,
		EmailVerificationTTL:   24 * time.Hour,
		PasswordResetTTL:       time.Hour,
		AuthRateLimitPerMinute: 10,
		JWTSecret:              testJWTSecret,
		JWTAccessTTL:           15 * time.Minute,
		JWTRefreshTTL:          720 * time.Hour,
		CookieDomain:           "",
		CookieSecure:           false,
		CookieSameSite:         "lax",
		CORSAllowedOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		StorageDriver:          "local",
		StorageLocalDir:        "./storage",
		StoragePublicBaseURL:   "http://localhost:8080/static",

		SandboxEnabled:            false,
		SandboxImage:              "busybox",
		SandboxWorkDir:            "",
		SandboxRateLimitPerMinute: 6,
	}
	if !reflect.DeepEqual(*cfg, want) {
		t.Errorf("defaults mismatch:\n got  %+v\n want %+v", *cfg, want)
	}
}

func TestLoad_Invalid(t *testing.T) {
	cases := []struct {
		name        string
		override    map[string]string
		wantInError string
	}{
		{"missing database url", map[string]string{"DATABASE_URL": ""}, "DATABASE_URL is required"},
		{"non-postgres scheme", map[string]string{"DATABASE_URL": "mysql://localhost/db"}, "postgres scheme"},
		{"bad app env", map[string]string{"APP_ENV": "staging"}, "APP_ENV"},
		{"port out of range", map[string]string{"HTTP_PORT": "70000"}, "HTTP_PORT must be in 1..65535"},
		{"port not an int", map[string]string{"HTTP_PORT": "abc"}, "HTTP_PORT must be an integer"},
		{"bad log level", map[string]string{"LOG_LEVEL": "verbose"}, "LOG_LEVEL"},
		{"bad log format", map[string]string{"LOG_FORMAT": "xml"}, "LOG_FORMAT"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setBaseEnv(t)
			for k, v := range tc.override {
				t.Setenv(k, v)
			}
			_, err := Load()
			if err == nil {
				t.Fatalf("Load() = nil error, want error containing %q", tc.wantInError)
			}
			if !strings.Contains(err.Error(), tc.wantInError) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantInError)
			}
		})
	}
}

func TestLoad_AggregatesMultipleProblems(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("APP_ENV", "staging")
	t.Setenv("LOG_FORMAT", "xml")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() = nil error, want aggregated error")
	}
	for _, want := range []string{"APP_ENV", "LOG_FORMAT", "DATABASE_URL is required"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("aggregated error missing %q; got:\n%s", want, err.Error())
		}
	}
}
