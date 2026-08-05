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

// routeFamily binds a routeHandlers field name to the call that mounts it.
// Name and registration live in one struct literal, so a family cannot exist
// without a name and a name cannot exist without a registration — the pairing
// is structural, not a second enumeration kept in sync by hand.
type routeFamily struct {
	name     string
	register func(httprouter.Muxer)
}

// conditionalRouteFamilies names every routeHandlers field that legitimately
// has no entry in the mount table on some boot path. It is exactly one:
// presence, which buildPresence returns nil for when deps.SQLDB is nil
// (main.go:1173-1175).
//
// Every other field is constructed on every path — including security, which
// takes a nil service but is itself always non-nil (main.go:404 and :406 both
// assign), and metrics, which is an unconditional http.HandlerFunc
// (observability/http.go:211). Guarding those would be a branch no
// environment executes; a zero-value field there is a wiring bug and must
// fail loudly, not be skipped silently.
//
// TestRouteCoverage consumes this set to decide which fields may be absent
// from the table — so "conditional" is declared once, here, rather than
// inferred from a zero value at two sites that could disagree.
var conditionalRouteFamilies = map[string]bool{"presence": true}

// routeFamilies is the mount table: the ordered set of route families
// buildRouter installs. See conditionalRouteFamilies for the one entry that
// is boot-path dependent.
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
	}
	if h.presence != nil {
		families = append(families, routeFamily{"presence", h.presence.RegisterRoutes})
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
		routeFamily{"metrics", func(mux httprouter.Muxer) {
			mux.Handle("/api/v1/metrics", h.metrics)
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
