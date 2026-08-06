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
}

// routeFamily binds a routeHandlers field name to the call that mounts it.
// Name and registration live in one struct literal, so a family cannot exist
// without a name and a name cannot exist without a registration — the pairing
// is structural, not a second enumeration kept in sync by hand.
type routeFamily struct {
	name     string
	register func(httprouter.Muxer)
}

// routeFamilies is the mount table: the ordered set of route families
// buildRouter installs. Mount is total (§4) — every routeHandlers field has
// exactly one entry here, unconditionally, including presence: on the
// SQLDB-less boot path h.presence is nil, but h.presence.RegisterRoutes binds
// fine on a nil *presence.Handler (it only creates a method value, it never
// dereferences the receiver), and the stream handler itself answers 501 when
// its receiver is nil, matching GetPresenceSnapshot's existing convention
// (iam/delivery/http/router.go:232).
//
// This is a plain production function with no test-only affordance.
// TestRouteCoverage (permissions_test.go) walks this same table to assert
// per-family route floors and to cross-check the names against
// routeHandlers' actual struct fields via reflection.
func routeFamilies(h routeHandlers) []routeFamily {
	families := []routeFamily{
		{"auth", h.auth.RegisterRoutes},
		{"health", h.health.RegisterRoutes},
		{"featureFlags", h.featureFlags.RegisterRoutes},
		{"audit", h.audit.RegisterRoutes},
		{"search", h.search.RegisterRoutes},
		{"security", h.security.RegisterRoutes},
		{"presence", h.presence.RegisterRoutes},
	}
	return append(families,
		routeFamily{"taxonomy", h.taxonomy.RegisterRoutes},
		routeFamily{"tokens", h.tokens.RegisterRoutes},
		routeFamily{"controlledDocuments", h.controlledDocuments.RegisterRoutes},
		routeFamily{"iamRouter", h.iamRouter.RegisterGenerated},
		routeFamily{"documents", func(mux httprouter.Muxer) {
			h.documents.RegisterRoutesWithRateLimit(mux, h.documentsRateLimit, h.documentsUserID)
		}},
		routeFamily{"templates", h.templates.Register},
		routeFamily{"approval", h.approval.RegisterRoutes},
		routeFamily{"distribution", func(mux httprouter.Muxer) {
			distributionhttp.RegisterRoutes(h.distribution, mux)
		}},
		routeFamily{"notifications", func(mux httprouter.Muxer) {
			notificationshttp.RegisterRoutes(h.notifications, mux)
		}},
	)
}

// buildRouter mounts every route family in routeHandlers onto mux. main()
// calls this; TestRouteCoverage walks routeFamilies directly (this function's
// entire body), so a family present in the table but absent from
// permissions.go's routeRules is a red test instead of a live 404/misrouted
// gap.
func buildRouter(mux httprouter.Muxer, h routeHandlers) {
	for _, f := range routeFamilies(h) {
		f.register(mux)
	}
}
