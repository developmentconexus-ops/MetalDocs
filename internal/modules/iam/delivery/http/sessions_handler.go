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

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
	"metaldocs/internal/platform/useragent"
)

// SessionAdmin is the narrow port the SessionsHandler depends on. The
// production implementation is *authpg.Repository (Postgres). Memory mode
// returns 501 from the handler so dev/test paths don't pay for an in-memory
// approximation of the JOIN.
type SessionAdmin interface {
	ListActiveSessions(ctx context.Context, q authdomain.SessionAdminQuery) ([]authdomain.SessionListItem, error)
	FindSession(ctx context.Context, sessionID string) (authdomain.Session, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error
}

type SessionsHandler struct {
	sessions          SessionAdmin
	sessionService    *iamapp.SessionService
	displayNameReader iamdomain.UserDisplayNameReader
	now               func() time.Time
}

func NewSessionsHandler(sessions SessionAdmin, _ ...auditdomain.Writer) *SessionsHandler {
	// The second variadic arg (legacy auditdomain.Writer) is intentionally
	// discarded: audit is now emitted in-tx by SessionService (H-3b Site 1).
	// Call sites that pass deps.AuditWriter continue to compile without change.
	return &SessionsHandler{
		sessions:          sessions,
		displayNameReader: iamdomain.NoopUserDisplayNameReader{},
		now:               func() time.Time { return time.Now().UTC() },
	}
}

// WithDisplayNameReader wires the iam-owned UserDisplayNameReader port used to
// enrich the session list with display names (M4/F4.4). Auth returns auth-owned
// session rows only; the iam consumer resolves names from metaldocs.iam_users
// via this port (read off-tx on the pool), so auth never reaches across the
// module boundary. Defaults to the Noop reader (all names fall back to user_id).
func (h *SessionsHandler) WithDisplayNameReader(r iamdomain.UserDisplayNameReader) *SessionsHandler {
	if r != nil {
		h.displayNameReader = r
	}
	return h
}

// WithSessionService wires the application-layer service that performs the
// atomic revoke + audit write (H-3b Site 1). When nil the revoke path is
// unavailable; this keeps unit tests that construct the handler without a
// full DB stack from breaking.
func (h *SessionsHandler) WithSessionService(svc *iamapp.SessionService) *SessionsHandler {
	h.sessionService = svc
	return h
}

func (h *SessionsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/sessions", h.handleSessions)
	mux.HandleFunc("/api/v1/auth/sessions/", h.handleSessionByID)
}

func (h *SessionsHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpresponse.WriteMethodNotAllowed(w, "GET")
		return
	}
	if h.sessions == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalError, "Sessions service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
		return
	}

	q := authdomain.SessionAdminQuery{
		TenantID: tenantID,
		UserID:   strings.TrimSpace(r.URL.Query().Get("user_id")),
	}
	// Default: active sessions only. ?is_active=false toggles to include revoked
	// + expired rows so admins can audit recent session activity.
	if v := strings.TrimSpace(r.URL.Query().Get("is_active")); v != "" {
		active, perr := strconv.ParseBool(v)
		if perr != nil {
			h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "is_active must be a boolean"))
			return
		}
		q.IncludeRevoked = !active
	}
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		limit, perr := strconv.Atoi(v)
		if perr != nil || limit < 1 || limit > 100 {
			h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "limit must be between 1 and 100"))
			return
		}
		q.Limit = limit
	}

	items, err := h.sessions.ListActiveSessions(r.Context(), q)
	if err != nil {
		slog.Error("iam sessions: list failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to list sessions"))
		return
	}

	// Enrich auth-owned session rows with iam-owned display names via the port
	// (M4/F4.4). Missing/empty names fall back to the user_id, reproducing the
	// old COALESCE(NULLIF(display_name,''), user_id) value consumer-side. The
	// read is best-effort: on error we render fallbacks rather than fail the list.
	displayNames := h.resolveDisplayNames(r.Context(), tenantID, items)

	out := make([]iamapi.SessionItem, 0, len(items))
	for _, item := range items {
		name := item.UserID
		if n, ok := displayNames[item.UserID]; ok && n != "" {
			name = n
		}
		// sql.NullTime fields: zero time.Time for invalid (never-set) values.
		// Real active sessions always have valid CreatedAt/ExpiresAt; LastSeenAt
		// may be zero on very fresh sessions that haven't been refreshed yet.
		createdAt := item.CreatedAt.Time
		if item.CreatedAt.Valid {
			createdAt = item.CreatedAt.Time.UTC()
		}
		lastSeenAt := item.LastSeenAt.Time
		if item.LastSeenAt.Valid {
			lastSeenAt = item.LastSeenAt.Time.UTC()
		}
		expiresAt := item.ExpiresAt.Time
		if item.ExpiresAt.Valid {
			expiresAt = item.ExpiresAt.Time.UTC()
		}
		si := iamapi.SessionItem{
			SessionId:   item.SessionID,
			UserId:      item.UserID,
			DisplayName: name,
			CreatedAt:   createdAt,
			LastSeenAt:  lastSeenAt,
			ExpiresAt:   expiresAt,
		}
		if item.IPAddress != "" {
			ip := item.IPAddress
			si.IpAddress = &ip
		}
		if item.UserAgent != "" {
			ua := item.UserAgent
			si.UserAgent = &ua
			dl := useragent.Label(item.UserAgent)
			si.DeviceLabel = &dl
		}
		out = append(out, si)
	}

	// Cursor pagination is deferred (see authpg.ListActiveSessions doc).
	// has_more=false is honest for the limit-only MVP.
	writeJSON(w, http.StatusOK, iamapi.ListSessionsResponse{
		Items: out,
		Page:  iamapi.CursorPage{HasMore: false},
	})
}

func (h *SessionsHandler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httpresponse.WriteMethodNotAllowed(w, "DELETE")
		return
	}
	if h.sessions == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalError, "Sessions service is not configured"))
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/v1/auth/sessions/")
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || strings.Contains(sessionID, "/") {
		h.writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Session not found"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
		return
	}

	// Tenant guard: load the session and verify it belongs to the caller's
	// tenant. Cross-tenant access returns 404, not 403, to avoid leaking
	// the existence of session IDs across tenants (per spec).
	session, err := h.sessions.FindSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			h.writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Session not found"))
			return
		}
		slog.Error("iam sessions: find failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to revoke session"))
		return
	}
	if session.TenantID != tenantID {
		h.writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Session not found"))
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

	// Use SessionService when wired (atomic revoke + audit in one tx, H-3b).
	// Falls back to the bare RevokeSession path in test environments where no
	// real DB / SessionService is available.
	if h.sessionService != nil {
		actor := ""
		if userID, ok := authn.UserIDFromContext(r.Context()); ok {
			actor = userID
		}
		err := h.sessionService.RevokeSession(r.Context(), iamapp.RevokeSessionInfo{
			SessionID: session.SessionID,
			UserID:    session.UserID,
			TenantID:  session.TenantID,
			Reason:    reason,
		}, actor)
		if err != nil {
			if errors.Is(err, authdomain.ErrSessionNotFound) {
				h.writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Session not found"))
				return
			}
			slog.Error("iam sessions: revoke failed", "err", err)
			h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to revoke session"))
			return
		}
	} else {
		now := h.now()
		if err := h.sessions.RevokeSession(r.Context(), sessionID, now); err != nil {
			if errors.Is(err, authdomain.ErrSessionNotFound) {
				h.writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Session not found"))
				return
			}
			slog.Error("iam sessions: revoke failed", "err", err)
			h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to revoke session"))
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}


// resolveDisplayNames batches the unique user_ids of the session rows and looks
// up their display names through the iam-owned port. Returns user_id -> name for
// users with a non-empty display_name; callers fall back to user_id for the rest.
func (h *SessionsHandler) resolveDisplayNames(ctx context.Context, tenantID string, items []authdomain.SessionListItem) map[string]string {
	if len(items) == 0 {
		return map[string]string{}
	}
	seen := make(map[string]struct{}, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.UserID]; ok {
			continue
		}
		seen[item.UserID] = struct{}{}
		ids = append(ids, item.UserID)
	}
	names, err := h.displayNameReader.DisplayNames(ctx, tenantID, ids)
	if err != nil {
		slog.Warn("iam sessions: display-name resolve failed; falling back to user_id", "err", err)
		return map[string]string{}
	}
	return names
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
