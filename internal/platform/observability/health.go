// Health is now mounted through the generated observabilityapi.ServerInterface
// router (HandlerWithOptions) rather than bare mux.HandleFunc — see
// RegisterRoutes. Handler implements observabilityapi.ServerInterface
// directly (not the strict variant): CheckLiveness/CheckReadiness's plain
// (w, r) shape already matches handleLive/handleReady's existing signature
// exactly, so strict's (ctx, RequestObject) -> (ResponseObject, error)
// indirection would buy nothing. Mirrors the audit/auth/security precedent.
//
// /healthz is deleted, not mounted here: the protocol has no exception
// mechanism (operator ruling C). Its four live consumers move to
// /api/v1/health/live in the same commit that deletes this mount.
package observability

import (
	"encoding/json"
	"net/http"

	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/httprouter"
	observabilityapi "metaldocs/internal/platform/observability/api"
)

type HealthHandler struct {
	provider RuntimeStatusProvider
}

func NewHealthHandler(provider RuntimeStatusProvider) *HealthHandler {
	return &HealthHandler{provider: provider}
}

func (h *HealthHandler) RegisterRoutes(mux httprouter.Muxer) {
	observabilityapi.HandlerWithOptions(h, observabilityapi.StdHTTPServerOptions{
		BaseURL:    apibase.BaseURL,
		BaseRouter: mux,
	})
}

// CheckLiveness adapts GET /health/live to the existing handler.
func (h *HealthHandler) CheckLiveness(w http.ResponseWriter, r *http.Request) {
	h.handleLive(w, r)
}

// CheckReadiness adapts GET /health/ready to the existing handler.
func (h *HealthHandler) CheckReadiness(w http.ResponseWriter, r *http.Request) {
	h.handleReady(w, r)
}

func (h *HealthHandler) handleLive(w http.ResponseWriter, r *http.Request) {
	if h.provider == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "live", "checks": []map[string]any{{"name": "process", "status": "up"}}})
		return
	}
	status, payload := h.provider.Live(r.Context())
	writeJSON(w, status, payload)
}

func (h *HealthHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	if h.provider == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "checks": []map[string]any{{"name": "process", "status": "up"}}})
		return
	}
	status, payload := h.provider.Ready(r.Context())
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
