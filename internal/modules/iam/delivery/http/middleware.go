package httpdelivery

import (
	"context"
	"errors"
	"net/http"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	httpresponse "metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type ctxKeyCapability struct{}
type ctxKeyAreaCode struct{}
type ctxKeyResourceID struct{}

var writeJSON = httpresponse.WriteJSON

// Visibility classifies a route's authorization expectation. The zero value
// (VisibilityPermissionGuarded) is the fail-closed default: a rule that
// forgets to set Visibility demands a capability rather than slipping through
// as public.
type Visibility int

const (
	VisibilityPermissionGuarded Visibility = iota
	VisibilitySessionRequired
	VisibilityPublic
)

type PermissionResolver func(method, path string) (iamdomain.Capability, Visibility)

type Middleware struct {
	caps         *iamapp.CapabilityService
	roleProvider iamdomain.RoleProvider
	enabled      bool
	legacyHeader bool
	resolver     PermissionResolver
}

func NewMiddleware(caps *iamapp.CapabilityService, roleProvider iamdomain.RoleProvider, enabled bool, legacyHeader ...bool) *Middleware {
	allowLegacy := false
	if len(legacyHeader) > 0 {
		allowLegacy = legacyHeader[0]
	}
	return &Middleware{
		caps:         caps,
		roleProvider: roleProvider,
		enabled:      enabled,
		legacyHeader: allowLegacy,
	}
}

func (m *Middleware) WithPermissionResolver(resolver PermissionResolver) *Middleware {
	m.resolver = resolver
	return m
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Roles")

		// C3: nil resolver is a misconfiguration — fail closed, never pass unauthenticated.
		if m.resolver == nil {
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Permission resolver not configured"))
			return
		}
		capability, visibility := m.resolver(r.Method, r.URL.Path)

		// VisibilityPublic: no auth checks at all.
		if visibility == VisibilityPublic {
			next.ServeHTTP(w, r)
			return
		}

		// C1: userID must come from the authenticated session context only.
		// Legacy X-User-Id header fallback removed — trusting a client-supplied
		// header bypasses cookie/HMAC/revocation/expiry/tenant checks entirely.
		userID := iamdomain.UserIDFromContext(r.Context())
		if userID == "" {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
			return
		}

		// C7: tenantID must come from the authenticated session context only.
		// Fallback to X-Tenant-ID header or DevTenantID removed — both are
		// client-controllable and allow cross-tenant access.
		tenantID, err := tenant.FromContext(r.Context())
		if err != nil {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
			return
		}

		// C4: VisibilitySessionRequired — session is verified above; skip capability
		// check but still enrich context with roles for downstream handlers.
		if visibility == VisibilitySessionRequired {
			ctx, ok := m.resolveRoles(w, r, userID, tenantID)
			if !ok {
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// VisibilityPermissionGuarded: full capability check required.

		// C8: nil caps is a misconfiguration — fail closed, never silently skip check.
		if m.caps == nil {
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Capability service not configured"))
			return
		}

		if err := m.caps.CanDo(r.Context(), userID, tenantID, string(capability)); err != nil {
			_ = problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
			return
		}

		ctx, ok := m.resolveRoles(w, r, userID, tenantID)
		if !ok {
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveRoles enriches the request context with the user's IAM roles when no
// CurrentUser is already present. Returns the updated context and true on
// success, or writes an error response and returns false on failure.
func (m *Middleware) resolveRoles(w http.ResponseWriter, r *http.Request, userID, tenantID string) (context.Context, bool) {
	rctx := r.Context()
	if _, hasUser := authdomain.CurrentUserFromContext(rctx); hasUser {
		return rctx, true //nolint:nilerr
	}
	var roles []iamdomain.Role
	if m.roleProvider != nil {
		resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID, tenantID)
		if errors.Is(err, iamdomain.ErrUserNotFound) || errors.Is(err, iamdomain.ErrUserInactive) {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "User is not authorized"))
			return nil, false
		}
		if errors.Is(err, iamdomain.ErrNoRolesAssigned) {
			_ = problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
			return nil, false
		}
		if err != nil {
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed"))
			return nil, false
		}
		roles = resolvedRoles
	}
	return iamdomain.WithAuthContext(rctx, userID, roles), true
}
