package notificationshttp

import (
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/httprouter"

	notificationsapi "metaldocs/internal/modules/notifications/api"
)

// Name identifies this publisher in boot assertion messages.
func (h *Handler) Name() string { return "notifications" }

// Tag is the OpenAPI tag this publisher owns.
func (h *Handler) Tag() string { return "notifications" }

// Mount wires the notifications endpoints onto mux via the generated
// strict handler. All endpoints share the /api/v1 base URL prefix.
func (h *Handler) Mount(mux httprouter.Muxer) {
	strict := notificationsapi.NewStrictHandler(h, nil)
	notificationsapi.HandlerWithOptions(strict, notificationsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    apibase.BaseURL,
	})
}
