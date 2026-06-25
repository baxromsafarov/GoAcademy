package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func doGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	rec := doGet(t, NewRouter(Deps{Logger: testLogger()}), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want %q", body["status"], "ok")
	}
}

func TestReadyz_NoChecks(t *testing.T) {
	rec := doGet(t, NewRouter(Deps{Logger: testLogger()}), "/readyz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body.Status != "ready" {
		t.Errorf("status = %q, want %q", body.Status, "ready")
	}
}

func TestReadyz_FailingCheck(t *testing.T) {
	failing := Check{Name: "database", Func: func(context.Context) error {
		return errors.New("connection refused")
	}}
	rec := doGet(t, NewRouter(Deps{Logger: testLogger(), ReadyChecks: []Check{failing}}), "/readyz")

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body.Status != "not_ready" {
		t.Errorf("status = %q, want %q", body.Status, "not_ready")
	}
	if body.Checks["database"] != "connection refused" {
		t.Errorf("checks[database] = %q, want %q", body.Checks["database"], "connection refused")
	}
}

func TestRequestIDHeaderSet(t *testing.T) {
	rec := doGet(t, NewRouter(Deps{Logger: testLogger()}), "/healthz")
	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Error("expected X-Request-Id response header to be set by middleware")
	}
}
