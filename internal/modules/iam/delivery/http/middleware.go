package httpdelivery

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	httpresponse "metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

var writeJSON = httpresponse.WriteJSON

// Visibility classifies a route's authorization expectation. The zero value
// (VisibilityPermissionGuarded) is the fail-closed default: a rule that
// forgets to set Visibility demands a capability rather than slipping through
// as public.
type Visibility int

// VisibilityPermissionGuarded, VisibilitySessionRequired, and VisibilityPublic
// are the three route authorization tiers a PermissionResolver may return.
// VisibilityPermissionGuarded (the zero value) is fail-closed: a resolver rule
// that omits Visibility demands a capability rather than defaulting to public.
const (
	VisibilityPermissionGuarded Visibility = iota
	VisibilitySessionRequired
	VisibilityPublic
	// VisibilityUnresolved means the mux matched a pattern that carries no
	// rule. §5's boot assertion makes this unreachable; reaching it is a
	// wiring bug, not a tier to guess at. Appended AFTER VisibilityPublic so
	// VisibilityPermissionGuarded keeps iota = 0 as the fail-closed zero value.
	VisibilityUnresolved
)

// PermissionResolver maps an HTTP method+path to the required tier-1
// capability and its Visibility tier. Implementations back the route table
// consulted by Middleware.Wrap.
type PermissionResolver func(*http.Request) (iamdomain.Capability, Visibility)

// Middleware is the tier-1 route→capability authorization gate (ADR 0022).
// It strips trusted identity headers before invoking the resolver and never
// trusts client-supplied X-User-ID/X-Tenant-ID for authorization decisions.
type Middleware struct {
	caps         *iamapp.CapabilityService
	roleProvider iamdomain.RoleProvider
	enabled      bool
	resolver     PermissionResolver
}

// NewMiddleware constructs the tier-1 authz middleware. When enabled is
// false, Wrap becomes a no-op passthrough (used for internal/unauthenticated
// entry points).
func NewMiddleware(caps *iamapp.CapabilityService, roleProvider iamdomain.RoleProvider, enabled bool) *Middleware {
	return &Middleware{
		caps:         caps,
		roleProvider: roleProvider,
		enabled:      enabled,
	}
}

// WithPermissionResolver wires the route→capability table. Must be called
// before Wrap serves traffic; a nil resolver at request time is treated as a
// misconfiguration and fails closed (500), never passes the request through.
func (m *Middleware) WithPermissionResolver(resolver PermissionResolver) *Middleware {
	m.resolver = resolver
	return m
}

// Wrap enforces tier-1 capability authorization on next. It strips
// X-User-ID/X-User-Roles from the inbound request (defense against
// header-spoofing), resolves userID/tenantID from authenticated session
// context only, and — for VisibilityPermissionGuarded routes — calls
// CapabilityService.CanDo before invoking next. Returns next unmodified when
// the middleware is disabled.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Roles")

		// C3: nil resolver is a misconfiguration — fail closed, never pass unauthenticated.
		if m.resolver == nil {
			m.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Permission resolver not configured"))
			return
		}
		capability, visibility, ok := m.resolveVisibilityCapability(w, r)
		if !ok {
			return
		}

		// VisibilityPublic: no auth checks at all.
		if visibility == VisibilityPublic {
			next.ServeHTTP(w, r)
			return
		}

		// C1/C7: userID and tenantID must come from the authenticated session
		// context only. Legacy X-User-Id/X-Tenant-ID header fallbacks removed —
		// trusting client-supplied headers bypasses cookie/HMAC/revocation/
		// expiry/tenant checks entirely.
		userID, tenantID, ok := m.sessionIdentity(w, r)
		if !ok {
			return
		}

		// C4: VisibilitySessionRequired — session is verified above; skip capability
		// check but still enrich context with roles for downstream handlers.
		if visibility == VisibilitySessionRequired {
			ctx, ok := m.resolveRoles(r.Context(), w, userID, tenantID)
			if !ok {
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// VisibilityPermissionGuarded: full capability check required.

		// C8: nil caps is a misconfiguration — fail closed, never silently skip check.
		if m.caps == nil {
			m.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Capability service not configured"))
			return
		}

		if err := m.caps.CanDo(r.Context(), userID, tenantID, string(capability)); err != nil {
			m.writeProblem(w, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Insufficient permissions"))
			return
		}

		ctx, ok := m.resolveRoles(r.Context(), w, userID, tenantID)
		if !ok {
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveVisibilityCapability resolves the route's capability/visibility
// tier via m.resolver and validates it against the known Visibility values.
// On an unresolved or unrecognized tier it writes the problem response
// itself (fail-closed) and returns ok=false.
func (m *Middleware) resolveVisibilityCapability(w http.ResponseWriter, r *http.Request) (capability iamdomain.Capability, visibility Visibility, ok bool) {
	capability, visibility = m.resolver(r)
	switch visibility {
	case VisibilityPublic, VisibilitySessionRequired, VisibilityPermissionGuarded:
		return capability, visibility, true
	case VisibilityUnresolved:
		slog.Warn("iam: route matched a pattern with no rule", "method", r.Method, "path", r.URL.Path) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		m.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Route resolved to no policy"))
		return capability, visibility, false
	default:
		// STAYS: genuinely unknown values.
		slog.Warn("iam: unknown route visibility", "method", r.Method, "path", r.URL.Path, "visibility", visibility) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		m.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Permission resolver returned unknown visibility"))
		return capability, visibility, false
	}
}

// sessionIdentity resolves userID/tenantID from the authenticated session
// context only (C1/C7 — never client-supplied headers). Writes the 401
// problem response itself and returns ok=false when either is missing.
func (m *Middleware) sessionIdentity(w http.ResponseWriter, r *http.Request) (userID, tenantID string, ok bool) {
	userID = iamdomain.UserIDFromContext(r.Context())
	if userID == "" {
		m.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return "", "", false
	}
	var err error
	tenantID, err = tenant.FromContext(r.Context())
	if err != nil {
		m.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return "", "", false
	}
	return userID, tenantID, true
}

// resolveRoles enriches the request context with the user's IAM roles when no
// CurrentUser is already present. Returns the updated context and true on
// success, or writes an error response and returns false on failure. ctx must
// be the caller's inherited request context (r.Context()) so role-lookup
// calls stay attached to the request's cancellation/deadline chain.
func (m *Middleware) resolveRoles(ctx context.Context, w http.ResponseWriter, userID, tenantID string) (context.Context, bool) {
	rctx := ctx
	if _, hasUser := authdomain.CurrentUserFromContext(rctx); hasUser {
		return rctx, true //nolint:nilerr
	}
	var roles []iamdomain.Role
	if m.roleProvider != nil {
		resolvedRoles, err := m.roleProvider.RolesByUserID(ctx, userID, tenantID)
		if errors.Is(err, iamdomain.ErrUserNotFound) || errors.Is(err, iamdomain.ErrUserInactive) {
			m.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "User is not authorized"))
			return nil, false
		}
		if errors.Is(err, iamdomain.ErrNoRolesAssigned) {
			m.writeProblem(w, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Insufficient permissions"))
			return nil, false
		}
		if err != nil {
			m.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Authorization lookup failed"))
			return nil, false
		}
		roles = resolvedRoles
	}
	return iamdomain.WithAuthContext(rctx, userID, roles), true
}

func (m *Middleware) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("iam: write response failed", "err", err)
	}
}
