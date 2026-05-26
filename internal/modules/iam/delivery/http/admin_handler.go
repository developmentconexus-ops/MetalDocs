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
	ListOnlineUsers(ctx context.Context, activeSince time.Time) ([]authdomain.OnlineUser, error)
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

type CreateUserRequest struct {
	UserID      string   `json:"userId,omitempty"`
	Username    string   `json:"username"`
	Email       string   `json:"email,omitempty"`
	DisplayName string   `json:"displayName"`
	Password    string   `json:"password"`
	Roles       []string `json:"roles"`
}

type UpdateUserRequest struct {
	DisplayName        *string `json:"displayName,omitempty"`
	Email              *string `json:"email,omitempty"`
	IsActive           *bool   `json:"isActive,omitempty"`
	NewPassword        string  `json:"newPassword,omitempty"`
	MustChangePassword *bool   `json:"mustChangePassword,omitempty"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword"`
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

func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/iam/users", h.handleUsers)
	mux.HandleFunc("/api/v1/iam/users/", h.handleUserRoute)
	mux.HandleFunc("/api/v1/iam/admin/overview", h.handleAdminOverview)
}

func (h *AdminHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListUsers(w, r)
	case http.MethodPost:
		h.handleCreateUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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
	onlineUsers, err := h.authService.ListOnlineUsers(r.Context(), activeSince)
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
func (h *AdminHandler) handleUserRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/iam/users/")
	parts := strings.Split(path, "/")
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && parts[1] == "roles" {
		switch r.Method {
		case http.MethodPost:
			h.handleUserRoleUpsert(w, r, strings.TrimSpace(parts[0]))
		case http.MethodPut:
			h.handleReplaceUserRoles(w, r, strings.TrimSpace(parts[0]))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && parts[1] == "reset-password" && r.Method == http.MethodPost {
		h.handleResetPassword(w, r, strings.TrimSpace(parts[0]))
		return
	}
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && parts[1] == "unlock" && r.Method == http.MethodPost {
		h.handleUnlockUser(w, r, strings.TrimSpace(parts[0]))
		return
	}
	if len(parts) == 1 && strings.TrimSpace(parts[0]) != "" && r.Method == http.MethodPatch {
		h.handlePatchUser(w, r, strings.TrimSpace(parts[0]))
		return
	}
	h.writeProblem(w, problem.New(http.StatusNotFound, "INTERNAL_ERROR", "Route not found"))
}

func (h *AdminHandler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	items, err := h.authService.ListUsers(r.Context(), tenantID)
	if err != nil {
		log.Printf("iam admin: list users failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users"))
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		roles := make([]string, 0, len(item.Roles))
		for _, role := range item.Roles {
			roles = append(roles, string(role))
		}
		out = append(out, map[string]any{
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
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *AdminHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	roles, ok := parseRoles(req.Roles)
	if !ok {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid roles"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	assignedBy := authenticatedActor(r)
	if err := h.authService.CreateUser(r.Context(), req.UserID, req.Username, req.Email, req.DisplayName, req.Password, tenantID, roles, assignedBy); err != nil {
		h.writeAuthError(w, err)
		return
	}
	createdUserID := strings.TrimSpace(defaultString(req.UserID, req.Username))
	writeJSON(w, http.StatusCreated, map[string]any{"userId": createdUserID})
	h.recordAudit(r, createdUserID, "auth.user.created", map[string]any{
		"username": req.Username,
		"roles":    req.Roles,
		"email":    req.Email,
	})
}

func (h *AdminHandler) handlePatchUser(w http.ResponseWriter, r *http.Request, userID string) {
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if err := h.authService.UpdateUser(r.Context(), authdomain.UpdateUserParams{
		UserID:             userID,
		DisplayName:        req.DisplayName,
		Email:              req.Email,
		IsActive:           req.IsActive,
		MustChangePassword: req.MustChangePassword,
	}, req.NewPassword); err != nil {
		h.writeAuthError(w, err)
		return
	}
	h.recordAudit(r, userID, "iam.user.updated", map[string]any{
		"displayName":        req.DisplayName,
		"email":              req.Email,
		"isActive":           req.IsActive,
		"mustChangePassword": req.MustChangePassword,
	})
	writeJSON(w, http.StatusOK, map[string]any{"userId": userID, "updated": true})
}

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

	roles, ok := parseRoles(req.Roles)
	if !ok {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid roles"))
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

	if err := h.service.ReplaceUserRoles(r.Context(), userID, req.DisplayName, replaceTenantID, roles, assignedBy); err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to replace user roles"))
		return
	}

	roleStrings := make([]string, 0, len(roles))
	for _, role := range roles {
		roleStrings = append(roleStrings, string(role))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"userId":      userID,
		"displayName": strings.TrimSpace(req.DisplayName),
		"roles":       roleStrings,
	})
	h.recordAudit(r, userID, "iam.user.roles.replaced", map[string]any{
		"roles": roleStrings,
	})
}

func (h *AdminHandler) handleResetPassword(w http.ResponseWriter, r *http.Request, userID string) {
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if err := h.authService.AdminResetPassword(r.Context(), userID, req.NewPassword); err != nil {
		h.writeAuthError(w, err)
		return
	}
	h.recordAudit(r, userID, "auth.user.password_reset", map[string]any{
		"mustChangePassword": true,
	})
	writeJSON(w, http.StatusOK, map[string]any{"userId": userID, "reset": true, "mustChangePassword": true})
}

func (h *AdminHandler) handleUnlockUser(w http.ResponseWriter, r *http.Request, userID string) {
	if h.authService == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "User management service is not configured"))
		return
	}
	if err := h.authService.UnlockUser(r.Context(), userID); err != nil {
		h.writeAuthError(w, err)
		return
	}
	h.recordAudit(r, userID, "auth.user.unlocked", map[string]any{})
	writeJSON(w, http.StatusOK, map[string]any{"userId": userID, "unlocked": true})
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
