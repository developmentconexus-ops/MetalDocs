package httpdelivery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/requesttrace"
)

type Handler struct {
	service *authapp.Service
	audit   auditdomain.Writer
	now     func() time.Time
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (r loginRequest) String() string {
	return fmt.Sprintf("{identifier:%s}", strings.TrimSpace(r.Identifier))
}

func (r changePasswordRequest) String() string {
	return "changePasswordRequest{redacted}"
}

func NewHandler(service *authapp.Service) *Handler {
	return &Handler{service: service, now: time.Now}
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
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}

	session, err := h.service.Authenticate(r.Context(), req.Identifier, req.Password, r)
	if err != nil {
		log.Printf("auth login failed for %q: %v", strings.TrimSpace(req.Identifier), err)
		http.SetCookie(w, h.service.ExpiredSessionCookie())
		identifierHash := hashIdentifier(req.Identifier)
		h.recordAudit(r, "", "auth.login.failed", identifierHash, map[string]any{
			"identifier_sha256": identifierHash,
		})
		h.writeAuthError(w, err)
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
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			log.Printf("auth logout failed: %v", err)
			http.SetCookie(w, h.service.ExpiredSessionCookie())
			h.writeAuthError(w, err)
			return
		}
		if user, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
			h.recordAudit(r, user.UserID, "auth.logout", user.UserID, map[string]any{})
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
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if err := h.service.ChangePasswordForUser(r.Context(), user, req.CurrentPassword, req.NewPassword); err != nil {
		log.Printf("auth change password failed for %q: %v", strings.TrimSpace(user.UserID), err)
		h.writeAuthError(w, err)
		return
	}
	currentUser, err := h.service.CurrentUser(r.Context(), user.UserID, user.TenantID)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"changed": true,
		"user":    currentUser,
	})
	h.recordAudit(r, user.UserID, "auth.password.changed", user.UserID, map[string]any{})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password"))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password"))
	case errors.Is(err, authdomain.ErrIdentityLocked):
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_ACCOUNT_LOCKED", "Account is temporarily locked"))
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, authdomain.ErrIdentityInactive):
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_ACCOUNT_INACTIVE", "User account is inactive"))
	case errors.Is(err, authdomain.ErrTenantNotPermitted):
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_TENANT_FORBIDDEN", "User has no role in the requested tenant"))
	case errors.Is(err, authdomain.ErrTenantClaimRequired):
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_TENANT_REQUIRED", "Tenant selection required"))
	default:
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"))
	}
}

func (h *Handler) recordAudit(r *http.Request, actorID, action, resourceID string, payload map[string]any) {
	if h.audit == nil {
		return
	}
	raw, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		log.Printf("auth audit payload marshal failed action=%s actor=%s: %v", action, actorID, marshalErr)
		return
	}
	traceID := requesttrace.Resolve(r.Context())
	tenantID := ""
	if sess, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
		tenantID = sess.TenantID
	}
	if err := h.audit.Record(r.Context(), auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   h.now().UTC(),
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

func (h *Handler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		log.Printf("auth problem write failed: %v", err)
	}
}

func hashIdentifier(identifier string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(identifier)))
	return hex.EncodeToString(sum[:])
}
