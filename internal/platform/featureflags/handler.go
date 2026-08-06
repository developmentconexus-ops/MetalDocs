// Package featureflags exposes server-controlled feature flag values to the
// frontend via GET /api/v1/feature-flags.
//
// The spec tag for this operation is "configuration" (api/openapi/v1/
// openapi.yaml), not "featureflags" -- the Go package and the spec tag are
// the same slice under two different names. cfg.yaml's include-tags is set
// to the tag, per §5 check 4 (tags govern, not package names).
//
// Mounted through the generated featureflagsapi.ServerInterface router
// (HandlerWithOptions) rather than bare mux.HandleFunc -- see RegisterRoutes.
// Handler implements the plain ServerInterface directly (not strict):
// GetFeatureFlags's (w, r) shape already matches the existing handle
// method's signature exactly, so strict's (ctx, RequestObject) ->
// (ResponseObject, error) indirection would buy nothing. Mirrors the
// audit/auth/security/health precedent.
package featureflags

import (
	"encoding/json"
	"net/http"

	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/config"
	featureflagsapi "metaldocs/internal/platform/featureflags/api"
	"metaldocs/internal/platform/httprouter"
)

// Handler serves GET /api/v1/feature-flags.
type Handler struct {
	cfg config.FeatureFlagsConfig
}

// NewHandler creates a Handler backed by the given config.
func NewHandler(cfg config.FeatureFlagsConfig) *Handler {
	return &Handler{cfg: cfg}
}

// RegisterRoutes registers the feature-flags route on mux.
func (h *Handler) RegisterRoutes(mux httprouter.Muxer) {
	featureflagsapi.HandlerWithOptions(h, featureflagsapi.StdHTTPServerOptions{
		BaseURL:    apibase.BaseURL,
		BaseRouter: mux,
	})
}

type featureFlagsResponse struct {
	MDDMNativeExportRolloutPct int `json:"MDDM_NATIVE_EXPORT_ROLLOUT_PCT"`
}

// GetFeatureFlags adapts GET /feature-flags to the existing handler.
func (h *Handler) GetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r)
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(featureFlagsResponse{
		MDDMNativeExportRolloutPct: h.cfg.MDDMNativeExportRolloutPercent,
	})
}
