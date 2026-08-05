package main

import (
	"net/http"

	approvalhttp "metaldocs/internal/modules/approval/http"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	"metaldocs/internal/modules/controlleddocuments"
	distributionhttp "metaldocs/internal/modules/distribution/delivery/http"
	documents "metaldocs/internal/modules/documents"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iampresence "metaldocs/internal/modules/iam/presence"
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	securitydelivery "metaldocs/internal/modules/security/delivery/http"
	"metaldocs/internal/modules/taxonomy"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/tokens"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/httprouter"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/ratelimit"
)

// routeHandlers collects every handler main() constructs that mounts routes
// onto the API mux. buildRouter is the single call site that mounts them —
// main() and TestRouteCoverage (permissions_test.go) call the exact same
// function, so a handler that's constructed but never added here is a
// build-time-visible gap in permissions.go's routeRules, not a live
// discovery.
//
// Deliberately NOT part of this struct, and so NOT covered by
// TestRouteCoverage: the /internal/test/* e2e scaffolding
// (mountE2EHandlersIfEnabled in main.go). Those routes are gated behind
// METALDOCS_E2E and are intentionally absent from permissions.go/routeRules
// (see e2eHandlersEnabled's doc comment) — they are mounted separately in
// main(), after buildRouter runs.
type routeHandlers struct {
	auth                *authdelivery.Handler
	health              *observability.HealthHandler
	featureFlags        *featureflags.Handler
	audit               *auditdelivery.Handler
	search              *searchdelivery.Handler
	security            *securitydelivery.Handler
	presence            *iampresence.Handler
	taxonomy            *taxonomy.Module
	tokens              *tokens.Module
	controlledDocuments *controlleddocuments.Module
	iamRouter           *iamdelivery.Router
	documents           *documents.Module
	documentsRateLimit  *ratelimit.Middleware
	documentsUserID     func(*http.Request) string
	templates           *templateshttp.Handler
	approval            *approvalhttp.Handler
	distribution        *distributionhttp.Handler
	notifications       *notificationshttp.Handler
	metrics             http.Handler
}

// buildRouter mounts every route family in routeHandlers onto mux. main()
// and TestRouteCoverage (permissions_test.go) call this exact function, so a
// forgotten mount surfaces as a red test at build time instead of a live
// 404/misrouted gap. presence and security are nil-guarded because both are
// nil on boot paths without a SQLDB (see their construction sites in
// main.go); every other handler is unconditionally constructed.
//
// onMount is an optional test hook: TestRouteCoverage passes a callback
// invoked with each family's name right after its RegisterRoutes call, so it
// can assert a per-family route floor (a family that silently registers zero
// patterns — e.g. a miswired test fixture — would otherwise hide behind the
// aggregate "did anything register" check). main() never passes one.
func buildRouter(mux httprouter.Muxer, h routeHandlers, onMount ...func(family string)) {
	mount := func(family string) {
		for _, fn := range onMount {
			fn(family)
		}
	}
	h.auth.RegisterRoutes(mux)
	mount("auth")
	h.health.RegisterRoutes(mux)
	mount("health")
	h.featureFlags.RegisterRoutes(mux)
	mount("featureFlags")
	h.audit.RegisterRoutes(mux)
	mount("audit")
	h.search.RegisterRoutes(mux)
	mount("search")
	if h.security != nil {
		h.security.RegisterRoutes(mux)
		mount("security")
	}
	if h.presence != nil {
		h.presence.RegisterRoutes(mux)
		mount("presence")
	}
	h.taxonomy.RegisterRoutes(mux)
	mount("taxonomy")
	h.tokens.RegisterRoutes(mux)
	mount("tokens")
	h.controlledDocuments.RegisterRoutes(mux)
	mount("controlledDocuments")
	h.iamRouter.RegisterGenerated(mux)
	mount("iamRouter")
	h.documents.RegisterRoutesWithRateLimit(mux, h.documentsRateLimit, h.documentsUserID)
	mount("documents")
	h.templates.Register(mux)
	mount("templates")
	h.approval.RegisterRoutes(mux)
	mount("approval")
	distributionhttp.RegisterRoutes(h.distribution, mux)
	mount("distribution")
	notificationshttp.RegisterRoutes(h.notifications, mux)
	mount("notifications")
	if h.metrics != nil {
		mux.Handle("/api/v1/metrics", h.metrics)
		mount("metrics")
	}
}
