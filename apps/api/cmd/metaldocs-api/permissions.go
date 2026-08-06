package main

import (
	"net/http"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// newPermissionResolver looks the request's matched mux pattern up in the
// generated httpSurface table (produced from the OpenAPI spec's
// x-authz-capability extensions by cmd/gen-http-surface). It is the sole
// tier-1 authority: the hand-written routeRules table and its
// resolveRoutePermission walker were deleted in Task 18 once this generated
// table became the single home for route visibility + capability.
func newPermissionResolver(mux *http.ServeMux) iamdelivery.PermissionResolver {
	return func(r *http.Request) (iamdomain.Capability, iamdelivery.Visibility) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			// The mux matched nothing: a 404 or a 405, not a route with a
			// policy. Demand a session — identical to today's fall-through —
			// and let the mux emit the status, which method_not_allowed
			// rewrites into problem+json.
			return "", iamdelivery.VisibilitySessionRequired
		}
		rule, ok := httpSurface[pattern]
		if !ok {
			// §5's boot assertion (assertSurface) makes this unreachable at
			// boot; reaching it here is a wiring bug, not a tier to guess at.
			return "", iamdelivery.VisibilityUnresolved
		}
		return rule.capability, rule.visibility
	}
}

func newPublicPathChecker(resolver iamdelivery.PermissionResolver) authdelivery.PublicPathChecker {
	return func(r *http.Request) bool {
		_, visibility := resolver(r)
		return visibility == iamdelivery.VisibilityPublic
	}
}

// newPasswordChangeAllowedChecker derives the password-change-allowed
// boolean from the same generated httpSurface table newPermissionResolver
// reads, so there is one authoritative source rather than a second
// hand-maintained list (mirrors newPublicPathChecker's derivation from the
// resolver). A request whose pattern carries no rule is treated as NOT
// allowed — fail closed, consistent with VisibilityUnresolved's 500 rather
// than a silent pass.
func newPasswordChangeAllowedChecker(mux *http.ServeMux) authdelivery.PasswordChangeAllowedChecker {
	return func(r *http.Request) bool {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			return false
		}
		rule, ok := httpSurface[pattern]
		if !ok {
			return false
		}
		return rule.allowedDuringPasswordChange
	}
}
