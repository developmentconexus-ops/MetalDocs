package httpdelivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
	"metaldocs/internal/platform/useragent"
)

// SessionAdmin is the narrow port the SessionsHandler depends on. The
// production implementation is *authpg.Repository (Postgres). Memory mode
// returns 501 from the handler so dev/test paths don't pay for an in-memory
// approximation of the JOIN.
type SessionAdmin interface {
	ListActiveSessions(ctx context.Context, q authpg.SessionAdminQuery) ([]authpg.SessionListItem, error)
	FindSession(ctx context.Context, sessionID string) (authdomain.Session, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error
}

type SessionsHandler struct {
	sessions SessionAdmin
	audit    auditdomain.Writer
	now      func() time.Time
}

func NewSessionsHandler(sessions SessionAdmin, auditWriter auditdomain.Writer) *SessionsHandler {
	return &SessionsHandler{
		sessions: sessions,
		audit:    auditWriter,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (h *SessionsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/sessions", h.handleSessions)
	mux.HandleFunc("/api/v1/auth/sessions/", h.handleSessionByID)
}

func (h *SessionsHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.sessions == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Sessions service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	}

	q := authpg.SessionAdminQuery{
		TenantID: tenantID,
		UserID:   strings.TrimSpace(r.URL.Query().Get("userId")),
	}
	// Default: active sessions only. ?isActive=false toggles to include revoked
	// + expired rows so admins can audit recent session activity.
	if v := strings.TrimSpace(r.URL.Query().Get("isActive")); v != "" {
		active, perr := strconv.ParseBool(v)
		if perr != nil {
			h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "isActive must be a boolean"))
			return
		}
		q.IncludeRevoked = !active
	}
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		limit, perr := strconv.Atoi(v)
		if perr != nil || limit < 1 || limit > 200 {
			h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "limit must be between 1 and 200"))
			return
		}
		q.Limit = limit
	}

	items, err := h.sessions.ListActiveSessions(r.Context(), q)
	if err != nil {
		slog.Error("iam sessions: list failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list sessions"))
		return
	}

	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{
			"sessionId":   item.SessionID,
			"userId":      item.UserID,
			"displayName": item.DisplayName,
			"createdAt":   nullTimeRFC3339(item.CreatedAt),
			"lastSeenAt":  nullTimeRFC3339(item.LastSeenAt),
			"expiresAt":   nullTimeRFC3339(item.ExpiresAt),
		}
		if item.IPAddress != "" {
			entry["ipAddress"] = item.IPAddress
		}
		if item.UserAgent != "" {
			entry["userAgent"] = item.UserAgent
			entry["deviceLabel"] = useragent.Label(item.UserAgent)
		}
		out = append(out, entry)
	}

	// Cursor pagination is deferred (see authpg.ListActiveSessions doc).
	// has_more=false is honest for the limit-only MVP.
	writeJSON(w, http.StatusOK, map[string]any{
		"items": out,
		"page": map[string]any{
			"has_more":    false,
			"next_cursor": nil,
		},
	})
}

func (h *SessionsHandler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.sessions == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Sessions service is not configured"))
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/v1/auth/sessions/")
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || strings.Contains(sessionID, "/") {
		h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Session not found"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	}

	// Tenant guard: load the session and verify it belongs to the caller's
	// tenant. Cross-tenant access returns 404, not 403, to avoid leaking
	// the existence of session IDs across tenants (per spec).
	session, err := h.sessions.FindSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Session not found"))
			return
		}
		slog.Error("iam sessions: find failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke session"))
		return
	}
	if session.TenantID != tenantID {
		h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Session not found"))
		return
	}

	// Optional body: {"reason":"..."} — recorded in the audit payload only;
	// rejecting an empty/invalid body would block the common UI path.
	reason := ""
	if r.Body != nil {
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
		if len(raw) > 0 {
			var body struct {
				Reason string `json:"reason"`
			}
			_ = json.Unmarshal(raw, &body)
			reason = strings.TrimSpace(body.Reason)
		}
	}

	now := h.now()
	if err := h.sessions.RevokeSession(r.Context(), sessionID, now); err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Session not found"))
			return
		}
		slog.Error("iam sessions: revoke failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke session"))
		return
	}

	h.emitRevokeAudit(r, session, reason)
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) emitRevokeAudit(r *http.Request, session authdomain.Session, reason string) {
	if h.audit == nil {
		return
	}
	payload := map[string]any{
		"sessionId":    session.SessionID,
		"targetUserId": session.UserID,
	}
	if reason != "" {
		payload["reason"] = reason
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if err := h.audit.Record(r.Context(), auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   h.now(),
		ActorID:      authenticatedActor(r),
		Action:       "auth.session.revoked",
		ResourceType: "session",
		ResourceID:   session.SessionID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      r.Header.Get("X-Trace-Id"),
		TenantID:     session.TenantID,
	}); err != nil {
		slog.Warn("iam sessions: audit emit failed", "err", err)
	}
}

func (h *SessionsHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("iam sessions: write response failed", "err", err)
	}
}

func nullTimeRFC3339(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
