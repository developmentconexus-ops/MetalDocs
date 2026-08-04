package httpresponse

import (
	"encoding/json"
	"net/http"

	"metaldocs/internal/platform/problem"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code problem.Code, message string) {
	_ = problem.Write(w, problem.New(status, code, message))
}

// WriteMethodNotAllowed writes the canonical RFC 9457 405 response with the
// RFC 9110 Allow header (D-03: no bare-status error responses).
func WriteMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	WriteError(w, http.StatusMethodNotAllowed, problem.CodeRequestMethodNotAllowed, "Method not allowed")
}

func ReadJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
