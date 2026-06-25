package respond

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goacademy/backend/internal/platform/apierr"
)

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`
} {
	t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details any    `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return body
}

func TestJSON_WritesStatusAndBody(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, map[string]string{"hello": "world"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["hello"] != "world" {
		t.Errorf("body = %v", got)
	}
}

func TestError_MapsAPIError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", apierr.NotFound("user not found"), http.StatusNotFound, apierr.CodeNotFound},
		{"unauthorized", apierr.Unauthorized("bad token"), http.StatusUnauthorized, apierr.CodeUnauthorized},
		{"conflict", apierr.Conflict("email taken"), http.StatusConflict, apierr.CodeConflict},
		{"rate limited", apierr.RateLimited("slow down"), http.StatusTooManyRequests, apierr.CodeRateLimited},
		{"validation", apierr.Validation("invalid"), http.StatusBadRequest, apierr.CodeValidation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			Error(rec, req, nil, tc.err)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			body := decodeError(t, rec)
			if body.Error.Code != tc.wantCode {
				t.Errorf("code = %q, want %q", body.Error.Code, tc.wantCode)
			}
			if body.Error.Message == "" {
				t.Error("message is empty")
			}
		})
	}
}

func TestError_UnknownErrorBecomesInternalAndHidesCause(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	secret := errors.New("connection string with secret=hunter2")
	Error(rec, req, nil, secret)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := decodeError(t, rec)
	if body.Error.Code != apierr.CodeInternal {
		t.Errorf("code = %q, want %q", body.Error.Code, apierr.CodeInternal)
	}
	if got := rec.Body.String(); contains(got, "hunter2") {
		t.Errorf("internal cause leaked to client: %s", got)
	}
}

func TestError_IncludesValidationDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := apierr.Validation("validation failed").
		WithDetails(map[string]string{"email": "must be a valid email"})
	Error(rec, req, nil, err)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	body := decodeError(t, rec)
	details, ok := body.Error.Details.(map[string]any)
	if !ok {
		t.Fatalf("details not an object: %v", body.Error.Details)
	}
	if details["email"] != "must be a valid email" {
		t.Errorf("details = %v", details)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
