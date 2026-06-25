package apierr

import (
	"errors"
	"net/http"
	"testing"
)

func TestConstructorsSetStatusAndCode(t *testing.T) {
	cases := []struct {
		err        *APIError
		wantStatus int
		wantCode   string
	}{
		{Validation("x"), http.StatusBadRequest, CodeValidation},
		{Unauthorized("x"), http.StatusUnauthorized, CodeUnauthorized},
		{Forbidden("x"), http.StatusForbidden, CodeForbidden},
		{NotFound("x"), http.StatusNotFound, CodeNotFound},
		{Conflict("x"), http.StatusConflict, CodeConflict},
		{RateLimited("x"), http.StatusTooManyRequests, CodeRateLimited},
		{Internal(), http.StatusInternalServerError, CodeInternal},
	}
	for _, tc := range cases {
		if tc.err.Status != tc.wantStatus {
			t.Errorf("%s: status = %d, want %d", tc.wantCode, tc.err.Status, tc.wantStatus)
		}
		if tc.err.Code != tc.wantCode {
			t.Errorf("code = %q, want %q", tc.err.Code, tc.wantCode)
		}
	}
}

func TestWithCause_UnwrapsAndIncludesInError(t *testing.T) {
	cause := errors.New("root cause")
	err := NotFound("missing").WithCause(cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is should find the wrapped cause")
	}
	if got := err.Error(); got != "missing: root cause" {
		t.Errorf("Error() = %q", got)
	}
}

func TestAsAPIError(t *testing.T) {
	var target *APIError
	if !errors.As(NotFound("x").WithCause(errors.New("y")), &target) {
		t.Fatal("errors.As should match *APIError")
	}
	if target.Status != http.StatusNotFound {
		t.Errorf("status = %d", target.Status)
	}
}
