// Health and observability (metrics) are mounted through two separate
// generated ServerInterface routers -- healthapi and observabilityapi -- each
// carrying exactly one OpenAPI tag (Task 15a, Ruling 1). They used to share
// one generated package (cfg.yaml carried both tags), but that conflation
// meant one Go package answered for two spec tags, breaking the
// SurfacePublisher "one tag, one publisher" contract (internal/platform/
// httprouter/publisher.go). HealthHandler and MetricsHandler below are two
// separate publishers in this same Go package, each owning exactly one tag:
// HealthHandler mounts healthapi (CheckLiveness/CheckReadiness), MetricsHandler
// mounts observabilityapi (GetMetrics, a thin delegation to the existing
// HTTPObservability.MetricsHandler() http.Handler).
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
	healthapi "metaldocs/internal/platform/observability/healthapi"
)

// HealthHandler serves the health/live and health/ready checks. It
// implements healthapi.ServerInterface directly (not the strict variant):
// CheckLiveness/CheckReadiness's plain (w, r) shape already matches
// handleLive/handleReady's existing signatures exactly, so strict's (ctx,
// RequestObject) -> (ResponseObject, error) indirection would buy nothing.
// Mirrors the audit/auth/security precedent.
type HealthHandler struct {
	provider RuntimeStatusProvider
}

// NewHealthHandler builds the health-check route handler.
func NewHealthHandler(provider RuntimeStatusProvider) *HealthHandler {
	return &HealthHandler{provider: provider}
}

// Name identifies this publisher in boot assertion messages.
func (h *HealthHandler) Name() string { return "health" }

// Tag is the OpenAPI tag this publisher owns.
func (h *HealthHandler) Tag() string { return "health" }

// Mount registers the health routes on mux.
func (h *HealthHandler) Mount(mux httprouter.Muxer) {
	healthapi.HandlerWithOptions(h, healthapi.StdHTTPServerOptions{
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

// MetricsHandler serves GET /metrics. It implements observabilityapi.
// ServerInterface directly (not the strict variant), matching HealthHandler's
// rationale above.
type MetricsHandler struct {
	metrics http.Handler
}

// NewMetricsHandler builds the metrics route handler. metrics may be nil in
// tests that never exercise GetMetrics -- Mount only binds method values to
// the mux, it never invokes them at mount time.
func NewMetricsHandler(metrics http.Handler) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

// Name identifies this publisher in boot assertion messages.
func (h *MetricsHandler) Name() string { return "observability" }

// Tag is the OpenAPI tag this publisher owns.
func (h *MetricsHandler) Tag() string { return "observability" }

// Mount registers the metrics route on mux.
func (h *MetricsHandler) Mount(mux httprouter.Muxer) {
	observabilityapi.HandlerWithOptions(h, observabilityapi.StdHTTPServerOptions{
		BaseURL:    apibase.BaseURL,
		BaseRouter: mux,
	})
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
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
