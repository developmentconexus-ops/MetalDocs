package httpdelivery

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/problem"
	platformtenant "metaldocs/internal/platform/tenant"
)

// PathLogin is the login endpoint path. Exported so the composition root's
// pre-auth rate limiter (REQ-MW-5) and this middleware's public-path fallback
// reference one constant — renaming the route cannot silently detach the
// rate-limit gate from the real endpoint.
const PathLogin = "/api/v1/auth/login"

// PublicPathChecker reports whether a request requires no session cookie
// (i.e. it is fully unauthenticated). Injecting this function into the
// middleware lets the composition root own the single authoritative
// route→visibility table (the generated httpSurface map), preventing the
// auth layer and the IAM permission layer from maintaining two independent
// lists that can drift apart.
type PublicPathChecker func(*http.Request) bool

// PasswordChangeAllowedChecker reports whether a request is one of the
// operations a principal with MustChangePassword set may still reach (e.g.
// change-password itself, logout, whoami). Like PublicPathChecker, this is
// derived from the single authoritative httpSurface table rather than a
// second hand-maintained list.
type PasswordChangeAllowedChecker func(*http.Request) bool

// Middleware resolves the session cookie into a CurrentUser and injects it
// into the request context, rejecting unauthenticated requests to non-public
// routes. When enabled is false, Wrap is a no-op passthrough.
type Middleware struct {
	service               *authapp.Service
	cfg                   authapp.Config
	enabled               bool
	publicChecker         PublicPathChecker            // mandatory once enabled; nil fails closed (500)
	passwordChangeChecker PasswordChangeAllowedChecker // mandatory once enabled; nil fails closed (500)
}

// NewMiddleware constructs a Middleware. enabled=false makes Wrap a no-op passthrough.
func NewMiddleware(service *authapp.Service, cfg authapp.Config, enabled bool) *Middleware {
	return &Middleware{service: service, cfg: cfg, enabled: enabled}
}

// WithPublicPathChecker wires the route→visibility table's public-path
// lookup. Must be called before Wrap serves traffic; a nil checker at
// request time is a misconfiguration and fails closed (500), never passes
// the request through as if it were public.
func (m *Middleware) WithPublicPathChecker(fn PublicPathChecker) *Middleware {
	m.publicChecker = fn
	return m
}

// WithPasswordChangeAllowedChecker wires the route→visibility table's
// password-change-allowed lookup. Must be called before Wrap serves traffic;
// a nil checker at request time is a misconfiguration and fails closed
// (500), never silently permits or blocks a must-change-password principal.
func (m *Middleware) WithPasswordChangeAllowedChecker(fn PasswordChangeAllowedChecker) *Middleware {
	m.passwordChangeChecker = fn
	return m
}

// Wrap enforces authentication on next: public paths (per isPublic) pass
// through unauthenticated; everything else requires a valid, non-expired,
// non-revoked session, resolves to an active identity, and — unless the
// principal has MustChangePassword and the path is one of the allowed
// exceptions — is rejected with 403 until the password is changed. On
// success it injects CurrentUser, the IAM auth context, and the tenant ID
// into the request context and strips any client-supplied X-Tenant-ID header.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Nil checker is a misconfiguration — fail closed, never pass through
		// as if the route were public.
		if m.publicChecker == nil {
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Public path checker not configured"))
			return
		}
		if m.publicChecker(r) {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(m.cfg.SessionCookieName)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
			return
		}

		currentUser, err := m.service.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, authdomain.ErrSessionNotFound) || errors.Is(err, authdomain.ErrSessionExpired) || errors.Is(err, authdomain.ErrSessionRevoked) || errors.Is(err, authdomain.ErrIdentityInactive) {
				_ = problem.Write(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
				return
			}
			slog.Error("auth resolve session failed", "err", err)
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Authentication failed"))
			return
		}
		if currentUser.MustChangePassword {
			// Nil checker is a misconfiguration — fail closed, never silently
			// admit a must-change-password principal.
			if m.passwordChangeChecker == nil {
				_ = problem.Write(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Password change allowed checker not configured"))
				return
			}
			if !m.passwordChangeChecker(r) {
				_ = problem.Write(w, problem.New(http.StatusForbidden, problem.CodeAuthPasswordChangeRequired, "Password change is required before accessing the application"))
				return
			}
		}

		// Report the principal outward to the observability middleware,
		// which runs outside authn (REQ-MW-4) and cannot see this
		// context enrichment from its layer.
		observability.SetPrincipal(r.Context(), currentUser.UserID)

		ctx := authdomain.WithCurrentUser(r.Context(), currentUser)
		ctx = iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)
		ctx = platformtenant.WithTenantID(ctx, currentUser.TenantID)
		ctx = platformtenant.WithActorID(ctx, currentUser.UserID)
		r2 := r.WithContext(ctx)
		r2.Header = r2.Header.Clone()
		r2.Header.Del("X-Tenant-ID")
		next.ServeHTTP(w, r2)
	})
}
