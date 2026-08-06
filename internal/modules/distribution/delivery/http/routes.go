package distributionhttp

import (
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/httprouter"

	distributionapi "metaldocs/internal/modules/distribution/api"
)

// Name identifies this publisher in boot assertion messages.
func (h *Handler) Name() string { return "distribution" }

// Tag is the OpenAPI tag this publisher owns.
func (h *Handler) Tag() string { return "distribution" }

// Mount wires the distribution endpoints onto mux via the generated
// strict handler. All three endpoints share the /api/v1 base URL prefix.
func (h *Handler) Mount(mux httprouter.Muxer) {
	strict := distributionapi.NewStrictHandler(h, nil)
	distributionapi.HandlerWithOptions(strict, distributionapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    apibase.BaseURL,
	})
}
