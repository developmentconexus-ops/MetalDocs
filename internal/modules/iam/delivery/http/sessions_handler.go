package httpdelivery

import (
	"context"
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

// defaultSessionsPageLimit mirrors authpg.ListActiveSessions's own default
// (limit<=0 -> 50) so behavior is unchanged when the caller omits ?limit.
// Kept local to the iam delivery layer rather than importing the auth
// module's repository (module boundary: iam depends on auth only through the
// SessionAdmin port below, never its internals).
const defaultSessionsPageLimit = 50

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

// SessionsHandler serves the /auth/sessions admin list + revoke surface.
// Revocation is tenant-guarded: a session belonging to another tenant
// returns 404, not 403, so cross-tenant probes cannot distinguish "exists
// elsewhere" from "does not exist".
type SessionsHandler struct {
	sessions          SessionAdmin
	sessionService    *iamapp.SessionService
	displayNameReader iamdomain.UserDisplayNameReader
	now               func() time.Time
}

// NewSessionsHandler constructs the handler with the Noop display-name
// reader (all names fall back to user_id until WithDisplayNameReader is
// called). The variadic auditdomain.Writer parameter is accepted for
// call-site compatibility only and discarded: audit is now emitted in-tx by
// SessionService (H-3b Site 1).
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

func (h *SessionsHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpresponse.WriteMethodNotAllowed(w, r, "GET")
		return
	}
	if h.sessions == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Sessions service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}

	q, requestedLimit, ok := h.parseSessionsListQuery(w, r, tenantID)
	if !ok {
		return
	}

	items, err := h.sessions.ListActiveSessions(r.Context(), q)
	if err != nil {
		slog.Error("iam sessions: list failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list sessions"))
		return
	}

	// Truthful has_more via the "over-fetch by one" trick: we asked for
	// requestedLimit+1 rows above. If the port returned that many, there is at
	// least one more row beyond this page — truncate the extra row before it
	// reaches the response (security-tech-debt.md #7; was hardcoded false).
	hasMore := len(items) > requestedLimit
	if hasMore {
		items = items[:requestedLimit]
	}

	// Enrich auth-owned session rows with iam-owned display names via the port
	// (M4/F4.4). Missing/empty names fall back to the user_id, reproducing the
	// old COALESCE(NULLIF(display_name,''), user_id) value consumer-side. The
	// read is best-effort: on error we render fallbacks rather than fail the list.
	displayNames := h.resolveDisplayNames(r.Context(), tenantID, items)
	out := toSessionItems(items, displayNames)

	// Full cursor pagination (opaque next_cursor) is still deferred (see
	// authpg.ListActiveSessions doc) — that would need a new contract param.
	// has_more is now truthful within the existing limit-only response shape
	// via the over-fetch-by-one truncation above (security-tech-debt.md #7).
	writeJSON(w, http.StatusOK, iamapi.ListSessionsResponse{
		Items: out,
		Page:  iamapi.CursorPage{HasMore: hasMore},
	})
}

// parseSessionsListQuery parses and validates the list-sessions query params
// (user_id, is_active, limit) into an authdomain.SessionAdminQuery.
// requestedLimit is the caller-facing page size (contract: 1-100, default 50
// — matches authpg.ListActiveSessions's own default so behavior is unchanged
// when the param is omitted); q.Limit asks the port for one extra row
// (requestedLimit+1) purely to detect has_more without a second round trip
// (security-tech-debt.md #7). On invalid input it writes the problem response
// itself and returns ok=false.
func (h *SessionsHandler) parseSessionsListQuery(w http.ResponseWriter, r *http.Request, tenantID string) (q authdomain.SessionAdminQuery, requestedLimit int, ok bool) {
	q = authdomain.SessionAdminQuery{
		TenantID: tenantID,
		UserID:   strings.TrimSpace(r.URL.Query().Get("user_id")),
	}
	// Default: active sessions only. ?is_active=false toggles to include revoked
	// + expired rows so admins can audit recent session activity.
	if v := strings.TrimSpace(r.URL.Query().Get("is_active")); v != "" {
		active, perr := strconv.ParseBool(v)
		if perr != nil {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "is_active must be a boolean"))
			return authdomain.SessionAdminQuery{}, 0, false
		}
		q.IncludeRevoked = !active
	}
	requestedLimit = defaultSessionsPageLimit
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		limit, perr := strconv.Atoi(v)
		if perr != nil || limit < 1 || limit > 100 {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "limit must be between 1 and 100"))
			return authdomain.SessionAdminQuery{}, 0, false
		}
		requestedLimit = limit
	}
	q.Limit = requestedLimit + 1
	return q, requestedLimit, true
}

// toSessionItems maps auth-owned session rows onto the generated
// iamapi.SessionItem wire shape, enriching with iam-owned display names
// (falling back to user_id) and normalizing sql.NullTime fields to UTC.
func toSessionItems(items []authdomain.SessionListItem, displayNames map[string]string) []iamapi.SessionItem {
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
	return out
}

func (h *SessionsHandler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httpresponse.WriteMethodNotAllowed(w, r, "DELETE")
		return
	}
	if h.sessions == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Sessions service is not configured"))
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/v1/auth/sessions/")
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || strings.Contains(sessionID, "/") {
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "Session not found"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}

	// Tenant guard: load the session and verify it belongs to the caller's
	// tenant. Cross-tenant access returns 404, not 403, to avoid leaking
	// the existence of session IDs across tenants (per spec).
	session, err := h.sessions.FindSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "Session not found"))
			return
		}
		slog.Error("iam sessions: find failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to revoke session"))
		return
	}
	if session.TenantID != tenantID {
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "Session not found"))
		return
	}

	reason := parseRevokeSessionReason(r)
	actor := ""
	if userID, ok := authn.UserIDFromContext(r.Context()); ok {
		actor = userID
	}

	if err := h.revokeSession(r.Context(), session, reason, actor); err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "Session not found"))
			return
		}
		slog.Error("iam sessions: revoke failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to revoke session"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseRevokeSessionReason best-effort decodes the optional {"reason":"..."}
// request body. An absent, empty, or malformed body yields "" rather than
// failing the revoke — rejecting the common empty-body UI path would be a
// regression, and the reason is recorded in the audit payload only.
func parseRevokeSessionReason(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
	if len(raw) == 0 {
		return ""
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal(raw, &body)
	return strings.TrimSpace(body.Reason)
}

// revokeSession dispatches to SessionService (atomic revoke + audit in one
// tx, H-3b) when wired, falling back to the bare RevokeSession port in test
// environments where no real DB / SessionService is available.
func (h *SessionsHandler) revokeSession(ctx context.Context, session authdomain.Session, reason, actor string) error {
	if h.sessionService != nil {
		return h.sessionService.RevokeSession(ctx, iamapp.RevokeSessionInfo{
			SessionID: session.SessionID,
			UserID:    session.UserID,
			TenantID:  session.TenantID,
			Reason:    reason,
		}, actor)
	}
	return h.sessions.RevokeSession(ctx, session.SessionID, h.now())
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
