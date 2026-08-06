package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampresence "metaldocs/internal/modules/iam/presence"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// UserAdminService is the narrow auth-module surface AdminHandler and
// PeopleHandler use for user CRUD and presence. Implemented by
// *authapp.Service; kept as an interface so handlers depend on a capability,
// not the concrete auth service.
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
	ListEvents(ctx context.Context, query auditdomain.ListEventsQuery) ([]auditdomain.Event, bool, error)
}

// KpiReader is the narrow port the overview uses to fold the KPI snapshot
// into its composed response. Implemented by *iamapp.ObservabilityService.
type KpiReader interface {
	GetKpi(ctx context.Context, tenantID string) (iamdomain.KpiSnapshot, error)
}

// PresenceReader is the narrow port the overview uses to fold real-time
// presence into its composed response. Implemented by
// *iampresence.PostgresRepository (PR-9). When wired, replaces the legacy
// 10-minute auth_sessions polling fallback derived from
// authService.ListOnlineUsers.
type PresenceReader interface {
	Snapshot(ctx context.Context, tenantID string, now time.Time) ([]iampresence.Item, error)
}

// AdminHandler serves the legacy role-edit endpoints (POST/PUT
// /iam/users/{user_id}/roles) plus the composed Admin Center overview.
// User CRUD lives on PeopleHandler; see RegisterRoutes.
type AdminHandler struct {
	service     *iamapp.AdminService
	authService UserAdminService
	auditEvents AuditEventLister
	kpi         KpiReader
	presence    PresenceReader
}

// UpsertUserRoleRequest is the POST /iam/users/{user_id}/roles request body.
type UpsertUserRoleRequest struct {
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	AssignedBy  string `json:"assigned_by,omitempty"`
}

// ReplaceUserRolesRequest is the PUT /iam/users/{user_id}/roles request body.
// Roles must decode to exactly one role (parseExactlyOneRole).
type ReplaceUserRolesRequest struct {
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	AssignedBy  string   `json:"assigned_by,omitempty"`
}

// NewAdminHandler constructs the handler. The variadic auditdomain.Writer
// parameter is accepted for call-site compatibility only and discarded:
// audit is now emitted in-tx by AdminService (H-3b).
func NewAdminHandler(service *iamapp.AdminService, authService UserAdminService, _ ...auditdomain.Writer) *AdminHandler {
	// auditWriter variadic is kept for call-site compatibility during the H-3b
	// migration; audit is now emitted in-tx by AdminService and is no longer
	// needed at the handler layer.
	return &AdminHandler{service: service, authService: authService}
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

// WithPresenceReader wires the real-time presence reader (PR-9). When
// set, the overview composes presence from iam_users.last_seen_at via
// this reader instead of the 10-minute auth_sessions window derived
// from authService.ListOnlineUsers.
func (h *AdminHandler) WithPresenceReader(r PresenceReader) *AdminHandler {
	h.presence = r
	return h
}

func (h *AdminHandler) handleUserRoleUpsertTyped(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
		return
	}
	h.handleUserRoleUpsert(w, r, userID)
}

func (h *AdminHandler) handleReplaceUserRolesTyped(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
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
		httpresponse.WriteMethodNotAllowed(w, "GET")
		return
	}
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "User management service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	now := time.Now().UTC()
	activeSince := now.Add(-10 * time.Minute)

	var (
		kpiSnapshot   iamdomain.KpiSnapshot
		onlineUsers   []authdomain.OnlineUser
		presenceItems []iampresence.Item
		recentEvents  []auditdomain.Event
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
		// PR-9: prefer iam_users.last_seen_at via PresenceReader when
		// wired; falls back to the legacy 10-minute auth_sessions
		// window in dev / in-memory mode.
		if h.presence != nil {
			items, err := h.presence.Snapshot(gctx, tenantID, now)
			if err != nil {
				return err
			}
			presenceItems = items
			return nil
		}
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
		events, _, err := h.auditEvents.ListEvents(gctx, auditdomain.ListEventsQuery{Limit: 25, TenantID: tenantID})
		if err != nil {
			return err
		}
		recentEvents = events
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("iam admin: overview composition failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to load admin overview"))
		return
	}

	presenceOut := make([]iamapi.OnlinePresenceItem, 0, len(onlineUsers)+len(presenceItems))
	if h.presence != nil {
		for _, item := range presenceItems {
			status := iamapi.OnlinePresenceItemStatus(string(item.Status))
			presenceOut = append(presenceOut, iamapi.OnlinePresenceItem{
				UserId:      item.UserID,
				Username:    item.Username,
				DisplayName: item.DisplayName,
				LastSeenAt:  item.LastSeenAt.UTC(),
				Status:      &status,
			})
		}
	} else {
		for _, item := range onlineUsers {
			presenceOut = append(presenceOut, iamapi.OnlinePresenceItem{
				UserId:      item.UserID,
				Username:    item.Username,
				DisplayName: item.DisplayName,
				LastSeenAt:  item.LastSeenAt.UTC(),
			})
		}
	}
	eventOut := make([]iamapi.AuditEventItem, 0, len(recentEvents))
	for _, item := range recentEvents {
		// AuditEventItem.Payload is map[string]interface{} per codegen contract.
		payload := make(map[string]interface{})
		if strings.TrimSpace(item.PayloadJSON) != "" {
			_ = json.Unmarshal([]byte(item.PayloadJSON), &payload)
		}
		eventOut = append(eventOut, iamapi.AuditEventItem{
			Id:           item.ID,
			OccurredAt:   item.OccurredAt.UTC(),
			ActorId:      item.ActorID,
			Action:       item.Action,
			ResourceType: item.ResourceType,
			ResourceId:   item.ResourceID,
			Payload:      payload,
			TraceId:      item.TraceID,
		})
	}
	writeJSON(w, http.StatusOK, iamapi.AdminOverviewResponse{
		Kpi:              kpiToOverviewTyped(kpiSnapshot),
		Presence:         presenceOut,
		RecentActivities: eventOut,
	})
}

// kpiToOverviewTyped folds the domain KpiSnapshot into the strict-server
// generated iamapi.IamKpiSnapshot — the same shape the standalone /iam/kpi
// endpoint returns. FailedLogins24h widens int64→int per the generated
// contract; Role carries through as the iamapi.UserRole enum.
func kpiToOverviewTyped(k iamdomain.KpiSnapshot) iamapi.IamKpiSnapshot {
	dist := make([]iamapi.IamKpiRoleCount, 0, len(k.RoleDistribution))
	for _, rc := range k.RoleDistribution {
		dist = append(dist, iamapi.IamKpiRoleCount{
			Role:  iamapi.UserRole(string(rc.Role)),
			Count: rc.Count,
		})
	}
	return iamapi.IamKpiSnapshot{
		LockedAccounts:       k.LockedAccounts,
		MfaCoveragePct:       k.MfaCoveragePct,
		FailedLogins24h:      int(k.FailedLogins24h),
		DormantUsers30d:      k.DormantUsers30d,
		RoleDistribution:     dist,
		AuditEventsPerMinute: k.AuditEventsPerMinute,
	}
}

// Legacy suffix-dispatcher / list / create / patch / reset / unlock methods
// were removed by PR-4. Those endpoints now live on PeopleHandler in
// people_handler.go, registered via Go 1.22 typed mux patterns under
// /api/v1/iam/users/* in the form "METHOD /path". The roles endpoints below
// stay here (PR-5 will own them next).

func (h *AdminHandler) handleUserRoleUpsert(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}

	var req UpsertUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}

	role, err := iamdomain.ParseRole(req.Role)
	if errors.Is(err, iamdomain.ErrInvalidRole) && strings.TrimSpace(req.Role) == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Role is required"))
		return
	}
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid role"))
		return
	}

	// actorID is the authenticated principal performing the action and becomes the
	// audit ActorID — it must never be client-controllable. assignedBy is the
	// domain "assigned_by" field, which may be overridden via the request body and
	// falls back to the actor when omitted (H-3b M2).
	actorID := authenticatedActor(r)
	assignedBy := strings.TrimSpace(req.AssignedBy)
	if assignedBy == "" {
		assignedBy = actorID
	}
	upsertTenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	if err := h.service.UpsertUserAndAssignRole(r.Context(), userID, req.DisplayName, upsertTenantID, role, assignedBy, actorID); err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to upsert user role"))
		return
	}

	dn := strings.TrimSpace(req.DisplayName)
	var dnPtr *string
	if dn != "" {
		dnPtr = &dn
	}
	writeJSON(w, http.StatusOK, iamapi.UpsertUserRoleResponse{
		UserId:      userID,
		Role:        string(role),
		DisplayName: dnPtr,
	})
	// Audit is now emitted in-tx by AdminService.UpsertUserAndAssignRole (H-3b).
}

func (h *AdminHandler) handleReplaceUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	var req ReplaceUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}

	role, err := parseExactlyOneRole(req.Roles)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
		return
	}

	actorID := authenticatedActor(r)
	assignedBy := strings.TrimSpace(req.AssignedBy)
	if assignedBy == "" {
		assignedBy = actorID
	}
	replaceTenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	if err := h.service.ReplaceUserRoles(r.Context(), userID, req.DisplayName, replaceTenantID, role, assignedBy, actorID); err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to replace user roles"))
		return
	}

	writeJSON(w, http.StatusOK, iamapi.ReplaceUserRolesResponse{
		UserId:      userID,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Roles:       []iamapi.UserRole{iamapi.UserRole(role)},
	})
	// Audit is now emitted in-tx by AdminService.ReplaceUserRoles (H-3b).
}

func (h *AdminHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if werr := problem.Write(w, p); werr != nil {
		slog.Warn("iam: write response failed", "err", werr)
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
