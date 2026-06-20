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

// PublicPathChecker returns true if the given method+path requires no session
// cookie (i.e. it is fully unauthenticated). Injecting this function into the
// middleware lets the composition root own the single authoritative list of
// public routes, preventing the auth layer and the IAM permission layer from
// maintaining two independent lists that can drift apart.
type PublicPathChecker func(method, path string) bool

type Middleware struct {
	service       *authapp.Service
	cfg           authapp.Config
	enabled       bool
	publicChecker PublicPathChecker // optional; falls back to defaultPublicPaths
}

func NewMiddleware(service *authapp.Service, cfg authapp.Config, enabled bool) *Middleware {
	return &Middleware{service: service, cfg: cfg, enabled: enabled}
}

// WithPublicPathChecker replaces the built-in public-path list with the
// provided checker. Use this in the composition root so there is one
// authoritative source of truth for which routes bypass authentication.
func (m *Middleware) WithPublicPathChecker(fn PublicPathChecker) *Middleware {
	m.publicChecker = fn
	return m
}

func (m *Middleware) isPublic(method, path string) bool {
	if m.publicChecker != nil {
		return m.publicChecker(method, path)
	}
	return defaultPublicPaths(method, path)
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublic(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(m.cfg.SessionCookieName)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
			return
		}

		currentUser, err := m.service.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, authdomain.ErrSessionNotFound) || errors.Is(err, authdomain.ErrSessionExpired) || errors.Is(err, authdomain.ErrSessionRevoked) || errors.Is(err, authdomain.ErrIdentityInactive) {
				_ = problem.Write(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
				return
			}
			slog.Error("auth resolve session failed", "err", err)
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Authentication failed"))
			return
		}
		if currentUser.MustChangePassword && !isPasswordChangeAllowedPath(r.URL.Path, r.Method) {
			_ = problem.Write(w, problem.New(http.StatusForbidden, problem.CodeAuthPasswordChangeRequired, "Password change is required before accessing the application"))
			return
		}

		// Report the principal outward to the observability middleware,
		// which runs outside authn (REQ-MW-4) and cannot see this
		// context enrichment from its layer.
		observability.SetPrincipal(r.Context(), currentUser.UserID)

		ctx := authdomain.WithCurrentUser(r.Context(), currentUser)
		ctx = iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)
		ctx = platformtenant.WithTenantID(ctx, currentUser.TenantID)
		r2 := r.WithContext(ctx)
		r2.Header = r2.Header.Clone()
		r2.Header.Del("X-Tenant-ID")
		next.ServeHTTP(w, r2)
	})
}

// defaultPublicPaths is the fallback used when no PublicPathChecker is
// injected. Keep this in sync with the composition root's authoritative list
// whenever WithPublicPathChecker is not used (e.g. tests).
func defaultPublicPaths(method, path string) bool {
	switch {
	case path == "/api/v1/health/live", path == "/api/v1/health/ready":
		return true
	case method == http.MethodPost && path == PathLogin:
		return true
	case method == http.MethodPost && path == "/api/v1/auth/logout":
		return true
	default:
		return false
	}
}

func isPasswordChangeAllowedPath(path, method string) bool {
	return (method == http.MethodGet && path == "/api/v1/auth/me") ||
		(method == http.MethodPost && path == "/api/v1/auth/change-password") ||
		(method == http.MethodPost && path == "/api/v1/auth/logout")
}
