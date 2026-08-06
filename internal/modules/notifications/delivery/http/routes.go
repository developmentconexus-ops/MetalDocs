package notificationshttp

import (
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/httprouter"

	notificationsapi "metaldocs/internal/modules/notifications/api"
)

// RegisterRoutes wires the notifications endpoints onto mux via the generated
// strict handler. All endpoints share the /api/v1 base URL prefix.
func RegisterRoutes(h *Handler, mux httprouter.Muxer) {
	strict := notificationsapi.NewStrictHandler(h, nil)
	notificationsapi.HandlerWithOptions(strict, notificationsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    apibase.BaseURL,
	})
}
