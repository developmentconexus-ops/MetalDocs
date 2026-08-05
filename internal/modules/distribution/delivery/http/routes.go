package distributionhttp

import (
	"metaldocs/internal/platform/httprouter"

	distributionapi "metaldocs/internal/modules/distribution/api"
)

// RegisterRoutes wires the distribution endpoints onto mux via the generated
// strict handler. All three endpoints share the /api/v1 base URL prefix.
func RegisterRoutes(h *Handler, mux httprouter.Muxer) {
	strict := distributionapi.NewStrictHandler(h, nil)
	distributionapi.HandlerWithOptions(strict, distributionapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
	})
}
