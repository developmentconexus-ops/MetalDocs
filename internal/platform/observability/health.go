// Health AND metrics are mounted through the generated
// observabilityapi.ServerInterface router (HandlerWithOptions) rather than
// bare mux.HandleFunc/mux.Handle — see RegisterRoutes. Handler implements
// observabilityapi.ServerInterface directly (not the strict variant):
// CheckLiveness/CheckReadiness/GetMetrics's plain (w, r) shape already
// matches handleLive/handleReady/MetricsHandler's existing signatures
// exactly, so strict's (ctx, RequestObject) -> (ResponseObject, error)
// indirection would buy nothing. Mirrors the audit/auth/security precedent.
//
// health and observability share one generated package (cfg.yaml carries
// both tags) because one package — this one — owns both operations; GetMetrics
// is a thin delegation to the existing HTTPObservability.MetricsHandler()
// http.Handler, not a reimplementation.
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
	metrics  http.Handler
}

// NewHealthHandler builds the combined health+metrics route handler. metrics
// may be nil in tests that never exercise GetMetrics — RegisterRoutes only
// binds method values to the mux, it never invokes them at mount time.
func NewHealthHandler(provider RuntimeStatusProvider, metrics http.Handler) *HealthHandler {
	return &HealthHandler{provider: provider, metrics: metrics}
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

// GetMetrics adapts GET /metrics to the existing HTTPObservability handler.
// Method-qualifying the mount ("GET /api/v1/metrics") does NOT make this a
// no-op relative to MetricsHandler's own 405 (http.go:211-216): a real
// http.ServeMux intercepts non-GET/HEAD requests to a method-qualified
// pattern before this handler ever runs, so MetricsHandler's in-handler
// check never fires for a mounted request. The status code is unchanged
// (405), but the body/Content-Type/Allow header change from the handler's
// application/problem+json RFC 9457 body to the stdlib mux's own plain-text
// "Method Not Allowed" response with Allow: GET, HEAD. That is the intended,
// uniform outcome — every migrated route family (auth, security, health,
// search, configuration) now rejects wrong methods the same way through the
// mux, and metrics joining them is correct rather than a regression to
// patch back. See TestMetricsRejectsNonGET in health_routing_test.go.
func (h *HealthHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.ServeHTTP(w, r)
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
