package httpresponse

import (
	"encoding/json"
	"net/http"

	"metaldocs/internal/platform/problem"
)

// WriteJSON writes payload as a JSON response body with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteMethodNotAllowed writes the canonical RFC 9457 405 response with the
// RFC 9110 Allow header (D-03: no bare-status error responses).
//
// The former WriteError sibling is gone: it was a second emission entry point
// for the same act, which is what G-07 exists to remove. Error responses go
// through problem.Respond, and this helper does too — it only owns the Allow
// header, not the envelope.
func WriteMethodNotAllowed(w http.ResponseWriter, r *http.Request, allow string) {
	w.Header().Set("Allow", allow)
	problem.Respond(w, r, problem.New(http.StatusMethodNotAllowed, problem.CodeRequestMethodNotAllowed, "Method not allowed"))
}

// ReadJSON decodes the request body as JSON into v.
func ReadJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
