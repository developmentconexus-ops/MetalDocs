package httpdelivery

import (
	"errors"
	"net/http"
	"strings"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	httpresponse "metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/tenant"
)

type ctxKeyCapability struct{}
type ctxKeyAreaCode struct{}
type ctxKeyResourceID struct{}

var writeJSON = httpresponse.WriteJSON

type PermissionResolver func(method, path string) (string, bool)

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

		traceID := requestTraceID(r)
		userID := iamdomain.UserIDFromContext(r.Context())
		if userID == "" && m.legacyHeader {
			userID = strings.TrimSpace(r.Header.Get("X-User-Id"))
		}
		if userID == "" {
			writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
			return
		}

		tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		if tenantID == "" {
			tenantID = tenant.DevTenantID
		}

		if m.caps != nil {
			if err := m.caps.CanDo(r.Context(), userID, tenantID, capability); err != nil {
				writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", traceID)
				return
			}
		}

		ctx := r.Context()
		if _, ok := authdomain.CurrentUserFromContext(ctx); !ok {
			var roles []iamdomain.Role
			if m.roleProvider != nil {
				resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID, tenantID)
				if errors.Is(err, iamdomain.ErrUserNotFound) || errors.Is(err, iamdomain.ErrUserInactive) {
					writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "User is not authorized", traceID)
					return
				}
				if err != nil {
					writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", traceID)
					return
				}
				roles = resolvedRoles
			}
			ctx = iamdomain.WithAuthContext(ctx, userID, roles)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func requestTraceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id")); traceID != "" {
		return traceID
	}
	return "trace-local"
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	httpresponse.WriteJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}
