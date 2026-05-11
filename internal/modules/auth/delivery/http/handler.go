package httpdelivery

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/httpresponse"
)

type Handler struct {
	service *authapp.Service
	audit   auditdomain.Writer
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func NewHandler(service *authapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) WithAudit(w auditdomain.Writer) *Handler {
	h.audit = w
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/v1/auth/me", h.handleMe)
	mux.HandleFunc("/api/v1/auth/change-password", h.handleChangePassword)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	traceID := requestTraceID(r)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", traceID)
		return
	}

	session, err := h.service.Authenticate(r.Context(), req.Identifier, req.Password, r)
	if err != nil {
		log.Printf("auth login failed for %q: %v", strings.TrimSpace(req.Identifier), err)
		http.SetCookie(w, h.service.ExpiredSessionCookie())
		h.writeAuthError(w, err, traceID)
		return
	}
	http.SetCookie(w, h.service.SessionCookie(session.RawToken, session.ExpiresAt))
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"user":      session.CurrentUser,
		"expiresAt": session.ExpiresAt.UTC().Format(time.RFC3339),
	})
	h.recordAudit(r, session.CurrentUser.UserID, "auth.login", session.CurrentUser.UserID, map[string]any{
		"tenant_id": session.CurrentUser.TenantID,
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie(h.service.SessionCookieName()); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err == nil {
			if user, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
				h.recordAudit(r, user.UserID, "auth.logout", user.UserID, map[string]any{})
			}
		}
	}
	http.SetCookie(w, h.service.ExpiredSessionCookie())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	traceID := requestTraceID(r)
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	traceID := requestTraceID(r)
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", traceID)
		return
	}
	if err := h.service.ChangePasswordForUser(r.Context(), user, req.CurrentPassword, req.NewPassword); err != nil {
		log.Printf("auth change password failed for %q: %v", strings.TrimSpace(user.UserID), err)
		h.writeAuthError(w, err, traceID)
		return
	}
	currentUser, err := h.service.CurrentUser(r.Context(), user.UserID, user.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", traceID)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"changed": true,
		"user":    currentUser,
	})
	h.recordAudit(r, user.UserID, "auth.password.changed", user.UserID, map[string]any{})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error, traceID string) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		writeAPIError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password", traceID)
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		writeAPIError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password", traceID)
	case errors.Is(err, authdomain.ErrIdentityLocked):
		writeAPIError(w, http.StatusForbidden, "AUTH_ACCOUNT_LOCKED", "Account is temporarily locked", traceID)
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), traceID)
	case errors.Is(err, authdomain.ErrIdentityInactive):
		writeAPIError(w, http.StatusForbidden, "AUTH_ACCOUNT_INACTIVE", "User account is inactive", traceID)
	case errors.Is(err, authdomain.ErrTenantNotPermitted):
		writeAPIError(w, http.StatusForbidden, "AUTH_TENANT_FORBIDDEN", "User has no role in the requested tenant", traceID)
	case errors.Is(err, authdomain.ErrTenantClaimRequired):
		writeAPIError(w, http.StatusForbidden, "AUTH_TENANT_REQUIRED", "Tenant selection required — user belongs to multiple tenants", traceID)
	default:
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", traceID)
	}
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

func (h *Handler) recordAudit(r *http.Request, actorID, action, resourceID string, payload map[string]any) {
	if h.audit == nil {
		return
	}
	raw, _ := json.Marshal(payload)
	traceID := "trace-local"
	if v := strings.TrimSpace(r.Header.Get("X-Trace-Id")); v != "" {
		traceID = v
	}
	tenantID := ""
	if sess, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
		tenantID = sess.TenantID
	}
	if err := h.audit.Record(r.Context(), auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "user",
		ResourceID:   resourceID,
		PayloadJSON:  string(raw),
		TraceID:      traceID,
		TenantID:     tenantID,
	}); err != nil {
		log.Printf("auth audit write failed action=%s actor=%s: %v", action, actorID, err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	httpresponse.WriteJSON(w, status, apiErrorEnvelope{Error: apiError{Code: code, Message: message, Details: map[string]any{}, TraceID: traceID}})
}
