package notificationshttp

import (
	"net/http"

	notificationsapi "metaldocs/internal/modules/notifications/api"
)

// RegisterRoutes wires the notifications endpoints onto mux via the generated
// strict handler. All endpoints share the /api/v1 base URL prefix.
func RegisterRoutes(h *Handler, mux *http.ServeMux) {
	strict := notificationsapi.NewStrictHandler(h, nil)
	notificationsapi.HandlerWithOptions(strict, notificationsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
	})
}
