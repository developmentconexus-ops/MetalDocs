package main

import (
	"net/http"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// newPermissionResolver looks the request's matched mux pattern up in the
// table surface points to — the composition root's httpSurface, possibly
// widened to include httpSurfaceE2E (produced from the OpenAPI spec's
// x-authz-capability extensions by cmd/gen-http-surface). It is the sole
// tier-1 authority: the hand-written routeRules table and its
// resolveRoutePermission walker were deleted in Task 18 once this generated
// table became the single home for route visibility + capability.
//
// surface is a *map, not a map, because the composition root's E2E decision
// (main.go:~858) reassigns the variable it points to AFTER this closure is
// built but BEFORE the first request is served — the same reason mux is
// passed as a pointer-shaped type (*http.ServeMux) whose routes are
// populated later. A map passed by value would freeze the table as it stood
// at construction, missing that widening entirely; every lookup here reads
// *surface so it always sees whatever the composition root last assigned.
func newPermissionResolver(mux *http.ServeMux, surface *map[string]surfaceRule) iamdelivery.PermissionResolver {
	return func(r *http.Request) (iamdomain.Capability, iamdelivery.Visibility) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			// The mux matched nothing: a 404 or a 405, not a route with a
			// policy. Demand a session — identical to today's fall-through —
			// and let the mux emit the status, which method_not_allowed
			// rewrites into problem+json.
			return "", iamdelivery.VisibilitySessionRequired
		}
		rule, ok := (*surface)[pattern]
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
// boolean from the same table newPermissionResolver reads — the composition
// root's httpSurface, possibly widened by the E2E decision — so there is one
// authoritative source rather than a second hand-maintained list (mirrors
// newPublicPathChecker's derivation from the resolver). A request whose
// pattern carries no rule is treated as NOT allowed — fail closed, consistent
// with VisibilityUnresolved's 500 rather than a silent pass. surface is a
// *map for the same reason as newPermissionResolver's: the table it points
// to may still be reassigned after this closure is built.
func newPasswordChangeAllowedChecker(mux *http.ServeMux, surface *map[string]surfaceRule) authdelivery.PasswordChangeAllowedChecker {
	return func(r *http.Request) bool {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			return false
		}
		rule, ok := (*surface)[pattern]
		if !ok {
			return false
		}
		return rule.allowedDuringPasswordChange
	}
}
