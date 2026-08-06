package main

import (
	approvalhttp "metaldocs/internal/modules/approval/http"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	"metaldocs/internal/modules/controlleddocuments"
	distributionhttp "metaldocs/internal/modules/distribution/delivery/http"
	documents "metaldocs/internal/modules/documents"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	securitydelivery "metaldocs/internal/modules/security/delivery/http"
	"metaldocs/internal/modules/taxonomy"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/tokens"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/httprouter"
	"metaldocs/internal/platform/observability"
)

// routeHandlers collects every handler main() constructs that mounts routes
// onto the API mux. buildRouter is the single call site that mounts them —
// main() and TestRouteCoverage (permissions_test.go) call the exact same
// function, so a handler that's constructed but never added here is a
// build-time-visible gap in permissions.go's routeRules, not a live
// discovery.
//
// Every field is an httprouter.SurfacePublisher (Task 15a): 16 fields, 16
// OpenAPI tags, one-to-one. documents' rate limiter and user-ID extractor are
// no longer routeFamilies closure params — they are wired onto *documents.
// Module as constructor fields via WithRateLimit before the module is placed
// here (main.go), because Mount(Muxer) takes exactly one argument.
// health/observability were one field (and one generated package) prior to
// this task; iamRouter now also mounts the hand-written presence stream
// (folded in, see iamdelivery.Router.Mount) — presence is no longer a field
// of its own.
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
	observability       *observability.MetricsHandler
	featureFlags        *featureflags.Handler
	audit               *auditdelivery.Handler
	search              *searchdelivery.Handler
	security            *securitydelivery.Handler
	taxonomy            *taxonomy.Module
	tokens              *tokens.Module
	controlledDocuments *controlleddocuments.Module
	iamRouter           *iamdelivery.Router
	documents           *documents.Module
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
// exactly one entry here, unconditionally. Every registration below is a
// SurfacePublisher's Mount method value (Task 15a): binding it never
// dereferences a nil receiver, and each publisher's own Mount documents its
// nil-safety (e.g. iamdelivery.Router.Mount calls presence.MountStream even
// when presence is nil; HealthHandler/MetricsHandler tolerate a nil provider/
// metrics field the same way they always have).
//
// This is a plain production function with no test-only affordance.
// TestRouteCoverage (permissions_test.go) walks this same table to assert
// per-family route floors and to cross-check the names against
// routeHandlers' actual struct fields via reflection.
func routeFamilies(h routeHandlers) []routeFamily {
	return []routeFamily{
		{"auth", h.auth.Mount},
		{"health", h.health.Mount},
		{"observability", h.observability.Mount},
		{"featureFlags", h.featureFlags.Mount},
		{"audit", h.audit.Mount},
		{"search", h.search.Mount},
		{"security", h.security.Mount},
		{"taxonomy", h.taxonomy.Mount},
		{"tokens", h.tokens.Mount},
		{"controlledDocuments", h.controlledDocuments.Mount},
		{"iamRouter", h.iamRouter.Mount},
		{"documents", h.documents.Mount},
		{"templates", h.templates.Mount},
		{"approval", h.approval.Mount},
		{"distribution", h.distribution.Mount},
		{"notifications", h.notifications.Mount},
	}
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
