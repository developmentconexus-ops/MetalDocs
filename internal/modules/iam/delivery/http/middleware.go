package httpdelivery

import (
	"errors"
	"net/http"
	"strings"

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

type PermissionResolver func(method, path string) (iamdomain.Capability, bool)

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

		if m.resolver == nil {
			next.ServeHTTP(w, r)
			return
		}
		capability, guarded := m.resolver(r.Method, r.URL.Path)
		if !guarded {
			next.ServeHTTP(w, r)
			return
		}

		userID := iamdomain.UserIDFromContext(r.Context())
		if userID == "" && m.legacyHeader {
			userID = strings.TrimSpace(r.Header.Get("X-User-Id"))
		}
		if userID == "" {
			_ = problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
			return
		}

		tenantID, err := tenant.FromContext(r.Context())
		if err != nil {
			tenantID = strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
			if tenantID == "" {
				tenantID = tenant.DevTenantID
			}
		}

		if m.caps != nil {
			if err := m.caps.CanDo(r.Context(), userID, tenantID, string(capability)); err != nil {
				_ = problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
				return
			}
		}

		ctx := r.Context()
		if _, ok := authdomain.CurrentUserFromContext(ctx); !ok {
			var roles []iamdomain.Role
			if m.roleProvider != nil {
				resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID, tenantID)
				if errors.Is(err, iamdomain.ErrUserNotFound) || errors.Is(err, iamdomain.ErrUserInactive) {
					_ = problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "User is not authorized"))
					return
				}
				if err != nil {
					_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed"))
					return
				}
				roles = resolvedRoles
			}
			ctx = iamdomain.WithAuthContext(ctx, userID, roles)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

