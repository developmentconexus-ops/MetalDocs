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

type AdminHandler struct {
	service     *iamapp.AdminService
	authService UserAdminService
	audit       auditdomain.Writer
	auditReader auditdomain.Reader
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

func (h *AdminHandler) WithAuditReader(reader auditdomain.Reader) *AdminHandler {
	h.auditReader = reader
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

// handleUserRoleUpsertTyped is the Go 1.22 typed-pattern entrypoint to the
// legacy role upsert. The path parameter is extracted with r.PathValue so we
// no longer split the URL by hand.
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
	users, err := h.authService.ListUsers(r.Context(), tenantID)
	if err != nil {
		log.Printf("iam admin: list users failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users"))
		return
	}
	onlineUsers, err := h.authService.ListOnlineUsers(r.Context(), tenantID, activeSince)
	if err != nil {
		log.Printf("iam admin: list online users failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list online users"))
		return
	}
	recentEvents := []auditdomain.Event{}
	if h.auditReader != nil {
		// TODO: keep this tenant filter in place until the backing governance_events resource index includes tenant_id.
		events, err := h.auditReader.ListEvents(r.Context(), auditdomain.ListEventsQuery{Limit: 25, TenantID: tenantID})
		if err != nil {
			log.Printf("iam admin: list audit events failed: %v", err)
			h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit events"))
			return
		}
		recentEvents = events
	}
	userOut := make([]map[string]any, 0, len(users))
	for _, item := range users {
		roles := make([]string, 0, len(item.Roles))
		for _, role := range item.Roles {
			roles = append(roles, string(role))
		}
		userOut = append(userOut, map[string]any{
			"userId":              item.UserID,
			"username":            item.Username,
			"email":               item.Email,
			"displayName":         item.DisplayName,
			"isActive":            item.IsActive,
			"mustChangePassword":  item.MustChangePassword,
			"failedLoginAttempts": item.FailedLoginAttempts,
			"roles":               roles,
			"lastLoginAt":         formatOptionalTime(item.LastLoginAt),
			"lockedUntil":         formatOptionalTime(item.LockedUntil),
			"createdAt":           item.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":           item.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	onlineOut := make([]map[string]any, 0, len(onlineUsers))
	for _, item := range onlineUsers {
		onlineOut = append(onlineOut, map[string]any{
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
		"users":            userOut,
		"onlineUsers":      onlineOut,
		"recentActivities": eventOut,
	})
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
	// Policy: tolerated fallback. Used as a label-only field on a path where
	// no authenticated actor existing legitimately maps to "system" (e.g.
	// bootstrap admin routes). Audit-bearing IAM mutations fail-closed in
	// routes_memberships.go.
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
