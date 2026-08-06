package approvalhttp

import (
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/httprouter"
	"net/http"

	approvalapi "metaldocs/internal/modules/approval/api"
)

// idempotentRoutes lists the route templates (method + path, matching the
// net/http 1.22 mux pattern syntax) that are gated by the router-level
// platform idempotency middleware (h.idempotentHandler, internal/platform/idempotency).
// Kept as a set so the Middlewares closure below can dispatch per-route
// without hand-registering each mutation separately — CON-03 approval slice:
// mounts the generated ServerInterface via HandlerWithOptions instead of a
// hand-written route list, so a spec'd route can no longer silently go
// unserved (missing ServerInterface methods fail to compile). Mirrors the
// templates pattern (internal/modules/templates/delivery/http/handler.go).
//
// Five other Idempotency-Key-bearing routes are deliberately excluded here:
// RecordApprovalStageSignoff, RecordDocumentSignoff (signoff_handler.go /
// doc_approval_handler.go, via h.idempStore / signoffIdempStore) and
// CreateApprovalRoute, UpdateApprovalRoute, DeactivateApprovalRoute
// (route_admin_handler.go, via h.routeAdmin's RouteAdminIdempStore) validate
// presence and replay themselves in-handler against a domain-specific replay
// envelope (signoff outcome / route id+version) — wrapping them in the
// generic h.idempotentHandler middleware too would double-apply idempotency
// against two different stores for the same request.
var idempotentRoutes = map[string]bool{
	"POST /api/v1/approval/instances/{instance_id}/cancel": true,
	"POST /api/v1/documents/{id}/cancel":                   true,
	"POST /api/v1/documents/{id}/obsolete":                 true,
	"POST /api/v1/documents/{id}/submit":                   true,
	"POST /api/v1/documents/{id}/review":                   true,
}

// RegisterRoutes wires all approval routes onto mux via the generated
// ServerInterface router, so a route added to the OpenAPI spec that lacks a
// Handler method fails to compile instead of silently 404ing.
func (h *Handler) RegisterRoutes(mux httprouter.Muxer) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// r.Pattern already carries the method prefix ("POST /api/v1/...")
			// because the generated router registers method-qualified patterns —
			// do NOT prepend r.Method again or the lookup silently misses.
			routeTemplate := r.Pattern
			if idempotentRoutes[routeTemplate] {
				h.idempotentHandler(routeTemplate, next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	approvalapi.HandlerWithOptions(h, approvalapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: apibase.BaseURL,
		Middlewares: []approvalapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			WriteError(w, err)
		},
	})
}
