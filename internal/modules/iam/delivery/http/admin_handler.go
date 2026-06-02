package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type UserAdminService interface {
	ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)
	ListOnlineUsers(ctx context.Context, tenantID string, activeSince time.Time) ([]authdomain.OnlineUser, error)
	CreateUser(ctx context.Context, userID, username, email, displayName, password, tenantID string, roles []iamdomain.Role, createdBy string) error
	UpdateUser(ctx context.Context, params authdomain.UpdateUserParams, newPassword string) error
	AdminResetPassword(ctx context.Context, userID, newPassword string) error
	UnlockUser(ctx context.Context, userID string) error
}

// AuditEventLister is the narrow port the Admin Center overview uses to
// surface recent activity. Mirrors audit.Service.ListEvents — kept scoped
// to the small surface the handler actually needs so tests can pass fakes
// without depending on the full audit reader/counter/export plumbing.
type AuditEventLister interface {
	ListEvents(ctx context.Context, query auditdomain.ListEventsQuery) ([]auditdomain.Event, error)
}

// KpiReader is the narrow port the overview uses to fold the KPI snapshot
// into its composed response. Implemented by *iamapp.ObservabilityService.
type KpiReader interface {
	GetKpi(ctx context.Context, tenantID string) (iamdomain.KpiSnapshot, error)
}

type AdminHandler struct {
	service     *iamapp.AdminService
	authService UserAdminService
	audit       auditdomain.Writer
	auditEvents AuditEventLister
	kpi         KpiReader
}

type UpsertUserRoleRequest struct {
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	AssignedBy  string `json:"assignedBy,omitempty"`
}

type ReplaceUserRolesRequest struct {
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
	AssignedBy  string   `json:"assignedBy,omitempty"`
}

func NewAdminHandler(service *iamapp.AdminService, authService UserAdminService, auditWriter ...auditdomain.Writer) *AdminHandler {
	var writer auditdomain.Writer
	if len(auditWriter) > 0 {
		writer = auditWriter[0]
	}
	return &AdminHandler{service: service, authService: authService, audit: writer}
}

// WithAuditEventLister wires the audit-events reader used by the composed
// admin overview to fetch the recent-activities stream. Replaces the
// legacy direct auditReader field (dropped in PR-8): the handler now
// depends on an application-layer interface, not a raw repository port.
func (h *AdminHandler) WithAuditEventLister(events AuditEventLister) *AdminHandler {
	h.auditEvents = events
	return h
}

// WithObservabilityService wires the KPI source used by the composed
// admin overview. Added in PR-8 — the overview response shape was
// retyped in PR-3 to carry { kpi, presence, recentActivities }.
func (h *AdminHandler) WithObservabilityService(svc KpiReader) *AdminHandler {
	h.kpi = svc
	return h
}

// RegisterRoutes wires the legacy AdminHandler surface: roles + overview only.
//
// PR-4 moved the People-tab user CRUD (list/invite/patch/bulk/reset/unlock/
// memberships) to PeopleHandler, which registers its own typed Go 1.22 mux
// patterns under /api/v1/iam/users/*. The role-edit endpoints stay here
// because PR-5 (Roles & Caps matrix) owns them and will restructure them
// next.
func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPost+" /api/v1/iam/users/{userId}/roles", h.handleUserRoleUpsertTyped)
	mux.HandleFunc(http.MethodPut+" /api/v1/iam/users/{userId}/roles", h.handleReplaceUserRolesTyped)
	mux.HandleFunc("/api/v1/iam/admin/overview", h.handleAdminOverview)
}

func (h *AdminHandler) handleUserRoleUpsertTyped(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	h.handleUserRoleUpsert(w, r, userID)
}

func (h *AdminHandler) handleReplaceUserRolesTyped(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	h.handleReplaceUserRoles(w, r, userID)
}

// handleAdminOverview composes the three independent snapshots that back
// the Admin Center Overview tab: KPI, presence, and recent activities.
// PR-8 refactor:
//   - drops the legacy users[] field (People tab owns it now);
//   - retypes the response to {kpi, presence, recentActivities};
//   - runs the three reads concurrently via errgroup so the wall-clock
//     budget is max(kpi, presence, audit) rather than the legacy sum.
func (h *AdminHandler) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	now := time.Now().UTC()
	activeSince := now.Add(-10 * time.Minute)

	var (
		kpiSnapshot  iamdomain.KpiSnapshot
		onlineUsers  []authdomain.OnlineUser
		recentEvents []auditdomain.Event
	)

	g, gctx := errgroup.WithContext(r.Context())

	g.Go(func() error {
		if h.kpi == nil {
			return nil
		}
		snap, err := h.kpi.GetKpi(gctx, tenantID)
		if err != nil {
			return err
		}
		kpiSnapshot = snap
		return nil
	})

	g.Go(func() error {
		items, err := h.authService.ListOnlineUsers(gctx, tenantID, activeSince)
		if err != nil {
			return err
		}
		onlineUsers = items
		return nil
	})

	g.Go(func() error {
		if h.auditEvents == nil {
			return nil
		}
		events, err := h.auditEvents.ListEvents(gctx, auditdomain.ListEventsQuery{Limit: 25, TenantID: tenantID})
		if err != nil {
			return err
		}
		recentEvents = events
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Printf("iam admin: overview composition failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load admin overview"))
		return
	}

	presenceOut := make([]map[string]any, 0, len(onlineUsers))
	for _, item := range onlineUsers {
		presenceOut = append(presenceOut, map[string]any{
			"userId":      item.UserID,
			"username":    item.Username,
			"displayName": item.DisplayName,
			"lastSeenAt":  item.LastSeenAt.UTC().Format(time.RFC3339),
		})
	}
	eventOut := make([]map[string]any, 0, len(recentEvents))
	for _, item := range recentEvents {
		payload := map[string]any{}
		if strings.TrimSpace(item.PayloadJSON) != "" {
			_ = json.Unmarshal([]byte(item.PayloadJSON), &payload)
		}
		eventOut = append(eventOut, map[string]any{
			"id":           item.ID,
			"occurredAt":   item.OccurredAt.UTC().Format(time.RFC3339),
			"actorId":      item.ActorID,
			"action":       item.Action,
			"resourceType": item.ResourceType,
			"resourceId":   item.ResourceID,
			"payload":      payload,
			"traceId":      item.TraceID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kpi":              kpiToOverviewJSON(kpiSnapshot),
		"presence":         presenceOut,
		"recentActivities": eventOut,
	})
}

// kpiToOverviewJSON mirrors the shape produced by ObservabilityHandler so
// the composed overview and the standalone /iam/kpi endpoint return
// identical kpi payloads.
func kpiToOverviewJSON(k iamdomain.KpiSnapshot) map[string]any {
	dist := make([]map[string]any, 0, len(k.RoleDistribution))
	for _, rc := range k.RoleDistribution {
		dist = append(dist, map[string]any{
			"role":  string(rc.Role),
			"count": rc.Count,
		})
	}
	return map[string]any{
		"lockedAccounts":       k.LockedAccounts,
		"mfaCoveragePct":       k.MfaCoveragePct,
		"failedLogins24h":      k.FailedLogins24h,
		"dormantUsers30d":      k.DormantUsers30d,
		"roleDistribution":     dist,
		"auditEventsPerMinute": k.AuditEventsPerMinute,
	}
}

// Legacy suffix-dispatcher / list / create / patch / reset / unlock methods
// were removed by PR-4. Those endpoints now live on PeopleHandler in
// people_handler.go, registered via Go 1.22 typed mux patterns under
// /api/v1/iam/users/* in the form "METHOD /path". The roles endpoints below
// stay here (PR-5 will own them next).

func (h *AdminHandler) handleUserRoleUpsert(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req UpsertUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}

	role, err := iamdomain.ParseRole(req.Role)
	if errors.Is(err, iamdomain.ErrInvalidRole) && strings.TrimSpace(req.Role) == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Role is required"))
		return
	}
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid role"))
		return
	}

	assignedBy := strings.TrimSpace(req.AssignedBy)
	if assignedBy == "" {
		assignedBy = authenticatedActor(r)
	}
	upsertTenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	if err := h.service.UpsertUserAndAssignRole(r.Context(), userID, req.DisplayName, upsertTenantID, role, assignedBy); err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upsert user role"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"userId":      userID,
		"role":        string(role),
		"displayName": strings.TrimSpace(req.DisplayName),
	})
	h.recordAudit(r, userID, "iam.user.role.upserted", map[string]any{
		"role":       string(role),
		"assignedBy": assignedBy,
	})
}

func (h *AdminHandler) handleReplaceUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	var req ReplaceUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}

	role, err := parseExactlyOneRole(req.Roles)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
		return
	}

	assignedBy := strings.TrimSpace(req.AssignedBy)
	if assignedBy == "" {
		assignedBy = authenticatedActor(r)
	}
	replaceTenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	if err := h.service.ReplaceUserRoles(r.Context(), userID, req.DisplayName, replaceTenantID, role, assignedBy); err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to replace user roles"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"userId":      userID,
		"displayName": strings.TrimSpace(req.DisplayName),
		"roles":       []string{string(role)},
	})
	h.recordAudit(r, userID, "iam.user.roles.replaced", map[string]any{
		"roles": []string{string(role)},
	})
}

func (h *AdminHandler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		h.writeProblem(w, problem.New(http.StatusConflict, "CONFLICT_ERROR", "User already exists"))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "User not found"))
	default:
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process user request"))
	}
}

func (h *AdminHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if werr := problem.Write(w, p); werr != nil {
		slog.Warn("iam: write response failed", "err", werr)
	}
}

func (h *AdminHandler) recordAudit(r *http.Request, userID, action string, payload map[string]any) {
	if h.audit == nil {
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return
	}
	now := time.Now().UTC()
	if err := h.audit.Record(r.Context(), auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   now,
		ActorID:      authenticatedActor(r),
		Action:       action,
		ResourceType: "user",
		ResourceID:   userID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      r.Header.Get("X-Trace-Id"),
		TenantID:     tenantID,
	}); err != nil {
		log.Printf("audit: failed to record %s for user %s: %v", action, userID, err)
	}
}

func parseRoles(items []string) ([]iamdomain.Role, bool) {
	out := make([]iamdomain.Role, 0, len(items))
	seen := map[iamdomain.Role]bool{}
	for _, item := range items {
		role, err := iamdomain.ParseRole(item)
		if err != nil {
			return nil, false
		}
		if !seen[role] {
			seen[role] = true
			out = append(out, role)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func parseExactlyOneRole(items []string) (iamdomain.Role, error) {
	roles, ok := parseRoles(items)
	if !ok {
		return "", errors.New("Invalid roles")
	}
	if len(roles) != 1 {
		return "", errors.New("Exactly one role is required")
	}
	return roles[0], nil
}

func authenticatedActor(r *http.Request) string {
	if userID, ok := authn.UserIDFromContext(r.Context()); ok {
		return userID
	}
	return "system"
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
