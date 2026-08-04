// Package httpdelivery implements the auth module's HTTP surface: the login/
// logout/me/change-password Handler and the session-resolving Middleware that
// gates every other route. Auth is pre-codegen (no api.gen.go); per ADR 0012,
// hand-rolled typed request/response structs are the sanctioned posture here.
package httpdelivery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

// Handler implements the auth module's HTTP endpoints: login, logout, me,
// and change-password.
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
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// authLoginResponse / changePasswordResponse are hand-rolled typed response bodies
// mirroring the OpenAPI AuthLoginResponse / ChangePasswordResponse schemas. Auth is
// pre-codegen (no api.gen.go); per ADR 0012 hand-rolled typed structs are the
// sanctioned posture for pre-codegen modules. They replace the prior untyped-map
// response literals (M7 F7.2 typed-body parity) with byte-identical wire output.
type authLoginResponse struct {
	User      authdomain.CurrentUser `json:"user"`
	ExpiresAt string                 `json:"expires_at"`
}

type changePasswordResponse struct {
	Changed bool                   `json:"changed"`
	User    authdomain.CurrentUser `json:"user"`
}

func (r loginRequest) String() string {
	return fmt.Sprintf("{identifier:%s}", strings.TrimSpace(r.Identifier))
}

func (r changePasswordRequest) String() string {
	return "changePasswordRequest{redacted}"
}

// NewHandler constructs a Handler backed by the given auth Service.
func NewHandler(service *authapp.Service) *Handler {
	return &Handler{service: service, now: time.Now}
}

// WithAudit wires an audit.Writer so login/logout/change-password events are
// recorded; without it, audit calls are silently skipped.
func (h *Handler) WithAudit(w auditdomain.Writer) *Handler {
	h.audit = w
	return h
}

// RegisterRoutes mounts the auth endpoints (login, logout, me, change-password) on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/v1/auth/me", h.handleMe)
	mux.HandleFunc("/api/v1/auth/change-password", h.handleChangePassword)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}

	session, err := h.service.Authenticate(r.Context(), req.Identifier, req.Password, r)
	if err != nil {
		// Log the sha256 of the identifier, never the raw plaintext: the raw value
		// is PII and an account-enumeration aid. The full identifier survives only
		// in the structured audit record below.
		identifierHash := hashIdentifier(req.Identifier)
		slog.Warn("auth login failed", "sha256", identifierHash, "err", err)
		http.SetCookie(w, h.service.ExpiredSessionCookie())
		h.recordAudit(r, "", "auth.login.failed", identifierHash, map[string]any{
			"identifier_sha256": identifierHash,
		})
		h.writeAuthError(w, err)
		return
	}
	http.SetCookie(w, h.service.SessionCookie(session.RawToken, session.ExpiresAt))
	httpresponse.WriteJSON(w, http.StatusOK, authLoginResponse{
		User:      session.CurrentUser,
		ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339),
	})
	h.recordAudit(r, session.CurrentUser.UserID, "auth.login", session.CurrentUser.UserID, map[string]any{
		"tenant_id": session.CurrentUser.TenantID,
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
	if cookie, err := r.Cookie(h.service.SessionCookieName()); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			slog.Error("auth logout failed", "err", err)
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
		httpresponse.WriteMethodNotAllowed(w, "GET")
		return
	}
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
	user, ok := authdomain.CurrentUserFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}
	if err := h.service.ChangePasswordForUser(r.Context(), user, req.CurrentPassword, req.NewPassword); err != nil {
		slog.Warn("auth change password failed", "user_id", strings.TrimSpace(user.UserID), "err", err)
		h.writeAuthError(w, err)
		return
	}
	currentUser, err := h.service.CurrentUser(r.Context(), user.UserID, user.TenantID)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Internal server error"))
		return
	}
	// F0.4: service already revoked all sessions (CWE-613). Mirror handleLogout
	// and expire the cookie client-side so the browser drops the dead value.
	http.SetCookie(w, h.service.ExpiredSessionCookie())
	httpresponse.WriteJSON(w, http.StatusOK, changePasswordResponse{
		Changed: true,
		User:    currentUser,
	})
	// Audit for auth.password.changed is now emitted inside the service tx (H-3b, F-07).
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthInvalidCredentials, "Invalid username/email or password"))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthInvalidCredentials, "Invalid username/email or password"))
	case errors.Is(err, authdomain.ErrIdentityLocked):
		h.writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthAccountLocked, "Account is temporarily locked"))
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
	case errors.Is(err, authdomain.ErrIdentityInactive):
		h.writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthAccountInactive, "User account is inactive"))
	case errors.Is(err, authdomain.ErrTenantNotPermitted):
		h.writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthTenantForbidden, "User has no role in the requested tenant"))
	case errors.Is(err, authdomain.ErrTenantClaimRequired):
		h.writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthTenantRequired, "Tenant selection required"))
	default:
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Internal server error"))
	}
}

func (h *Handler) recordAudit(r *http.Request, actorID, action, resourceID string, payload map[string]any) {
	if h.audit == nil {
		return
	}
	raw, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		slog.Warn("auth audit payload marshal failed", "action", action, "actor", actorID, "err", marshalErr)
		return
	}
	traceID := requesttrace.Resolve(r.Context())
	tenantID := ""
	if sess, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
		tenantID = sess.TenantID
	}
	if err := h.audit.Record(r.Context(), auditdomain.Event{ //cilint:allow-post-commit-audit auth login/logout best-effort: login.failed has no business tx to attach to (H-3b bounded defer)
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
		slog.Warn("auth audit write failed", "action", action, "actor", actorID, "err", err)
	}
}

func (h *Handler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("auth problem write failed", "err", err)
	}
}

func hashIdentifier(identifier string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(identifier)))
	return hex.EncodeToString(sum[:])
}
