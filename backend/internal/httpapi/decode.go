package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/goacademy/backend/internal/platform/apierr"
)

// maxRequestBody caps decoded request bodies to guard against oversized payloads.
const maxRequestBody = 1 << 20 // 1 MiB

// decodeJSON decodes the request body into dst, rejecting unknown fields and
// bodies larger than maxRequestBody. A decode failure is returned as a 400.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apierr.Validation("invalid request body").WithCause(err)
	}
	return nil
}
