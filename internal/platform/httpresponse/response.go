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

func ReadJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
