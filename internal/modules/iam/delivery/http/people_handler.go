// people_handler.go — PR-4 People-tab HTTP wiring.
//
// Why this file does not call HandlerFromMux from the codegen:
// PR-3 authored the OpenAPI with mixed-prefix paths (some entries bare
// `/iam/users`, others prefixed `/api/v1/...`). HandlerFromMux therefore
// cannot register the People routes cleanly. Until PR-5 reconciles the spec
// prefix we register each People route by hand with Go 1.22 typed mux patterns
// and decode/encode the codegen request/response models via encoding/json.
//
// All routes are tenant-scoped via tenant.FromContext. Audit emission for the
// mutating paths reuses the existing AdminHandler.recordAudit helper.
package httpdelivery

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// PeopleHandler owns the People-tab REST surface introduced in PR-4. It sits
// next to AdminHandler in the same package so the existing recordAudit /
// writeProblem helpers stay reusable without exporting them.
type PeopleHandler struct {
	service *iamapp.PeopleService
	authSvc UserAdminService
	audit   auditdomain.Writer
}

func NewPeopleHandler(service *iamapp.PeopleService, authSvc UserAdminService, audit auditdomain.Writer) *PeopleHandler {
	return &PeopleHandler{service: service, authSvc: authSvc, audit: audit}
}

// RegisterRoutes attaches People-tab routes with Go 1.22 typed mux patterns.
// The path-param wildcard {userId} is extracted via r.PathValue.
func (h *PeopleHandler) RegisterRoutes(mux *http.ServeMux) {
	if h == nil {
		return
	}
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/users", h.handleListUsers)
	mux.HandleFunc(http.MethodPost+" /api/v1/iam/users/invite", h.handleInvite)
	mux.HandleFunc(http.MethodPost+" /api/v1/iam/users/bulk", h.handleBulk)
	mux.HandleFunc(http.MethodPatch+" /api/v1/iam/users/{userId}", h.handlePatch)
	mux.HandleFunc(http.MethodPost+" /api/v1/iam/users/{userId}/reset-password", h.handleResetPassword)
	mux.HandleFunc(http.MethodPost+" /api/v1/iam/users/{userId}/unlock", h.handleUnlock)
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/users/{userId}/memberships", h.handleListMemberships)
}

// ─── GET /iam/users ─────────────────────────────────────────────────────────

func (h *PeopleHandler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	filters := iamapp.ListFilters{
		IsActive: parseOptionalBool(r.URL.Query().Get("isActive")),
		Q:        nonEmptyPtr(r.URL.Query().Get("q")),
		Cursor:   nonEmptyPtr(r.URL.Query().Get("cursor")),
		AreaCode: nonEmptyPtr(r.URL.Query().Get("areaCode")),
		Limit:    parseLimit(r.URL.Query().Get("limit")),
	}
	if rawRole := strings.TrimSpace(r.URL.Query().Get("role")); rawRole != "" {
		role := iamdomain.Role(rawRole)
		if !iamdomain.IsValidRole(role) {
			h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid role filter"))
			return
		}
		filters.Role = &role
	}

	result, err := h.service.ListFiltered(r.Context(), tenantID, filters)
	if err != nil {
		log.Printf("iam people: list users failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users"))
		return
	}

	items := make([]iamapi.ManagedUserCore, 0, len(result.Items))
	for _, u := range result.Items {
		items = append(items, toManagedUserCore(u))
	}
	var nextCursor *string
	if result.HasMore && result.NextCursor != "" {
		next := result.NextCursor
		nextCursor = &next
	}
	resp := iamapi.ListUsersResponse{
		Items: items,
		Page: iamapi.CursorPage{
			HasMore:    result.HasMore,
			NextCursor: nextCursor,
		},
		Total: result.Total,
	}
	writeJSON(w, http.StatusOK, resp)
}

// ─── POST /iam/users/invite ────────────────────────────────────────────────

func (h *PeopleHandler) handleInvite(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	var body iamapi.UserInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	input := iamapp.InviteInput{
		Username:    strings.TrimSpace(body.Username),
		Email:       strings.TrimSpace(string(body.Email)),
		DisplayName: strings.TrimSpace(body.DisplayName),
		TenantRole:  iamdomain.Role(string(body.TenantRole)),
	}
	if body.AreaMemberships != nil {
		for _, m := range *body.AreaMemberships {
			input.AreaMemberships = append(input.AreaMemberships, iamapp.InviteAreaInput{
				AreaCode: strings.TrimSpace(m.AreaCode),
				Role:     iamdomain.Role(string(m.Role)),
			})
		}
	}

	actor := authenticatedActor(r)
	result, err := h.service.Invite(r.Context(), tenantID, actor, input)
	if err != nil {
		h.writePeopleError(w, err)
		return
	}

	// CRITICAL: tempPassword is returned to the operator one time only. The
	// audit payload deliberately excludes it; only username/email/tenantRole
	// and area count are journalled. Do NOT log tempPassword anywhere.
	areaCount := 0
	if input.AreaMemberships != nil {
		areaCount = len(input.AreaMemberships)
	}
	h.recordAudit(r, result.UserID, "iam.user.invited", map[string]any{
		"username":   input.Username,
		"email":      input.Email,
		"tenantRole": string(input.TenantRole),
		"areaCount":  areaCount,
	})

	writeJSON(w, http.StatusCreated, iamapi.UserInviteResponse{
		UserId:       result.UserID,
		TempPassword: result.TempPassword,
	})
}

// ─── PATCH /iam/users/{userId} ─────────────────────────────────────────────

func (h *PeopleHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	var body iamapi.UpdateManagedUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}

	input := iamapp.PatchInput{
		DisplayName:        body.DisplayName,
		Email:              body.Email,
		IsActive:           body.IsActive,
		MustChangePassword: body.MustChangePassword,
	}
	if body.TenantRole != nil {
		role := iamdomain.Role(string(*body.TenantRole))
		input.TenantRole = &role
	}

	actor := authenticatedActor(r)
	changes, err := h.service.PatchAtomic(r.Context(), tenantID, actor, userID, input)
	if err != nil {
		h.writePeopleError(w, err)
		return
	}
	// Audit deliberately omits password material — UpdateManagedUserRequest's
	// NewPassword (if present) is not forwarded through PatchAtomic, so the
	// audit payload only sees the metadata + tenantRole changes.
	h.recordAudit(r, userID, "iam.user.updated", changes)

	writeJSON(w, http.StatusOK, map[string]any{
		"userId":  userID,
		"updated": true,
		"changes": changes,
	})
}

// ─── POST /iam/users/{userId}/reset-password ──────────────────────────────

func (h *PeopleHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Auth service is not configured"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	var body iamapi.ResetManagedUserPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if err := h.authSvc.AdminResetPassword(r.Context(), userID, body.NewPassword); err != nil {
		h.writeAuthError(w, err)
		return
	}
	h.recordAudit(r, userID, "auth.user.password_reset", map[string]any{"mustChangePassword": true})
	writeJSON(w, http.StatusOK, iamapi.ResetManagedUserPasswordResponse{
		UserId:             userID,
		Reset:              true,
		MustChangePassword: true,
	})
}

// ─── POST /iam/users/{userId}/unlock ──────────────────────────────────────

func (h *PeopleHandler) handleUnlock(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Auth service is not configured"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	if err := h.authSvc.UnlockUser(r.Context(), userID); err != nil {
		h.writeAuthError(w, err)
		return
	}
	h.recordAudit(r, userID, "auth.user.unlocked", map[string]any{})
	writeJSON(w, http.StatusOK, iamapi.UnlockManagedUserResponse{UserId: userID, Unlocked: true})
}

// ─── POST /iam/users/bulk ──────────────────────────────────────────────────

func (h *PeopleHandler) handleBulk(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	var body iamapi.UserBulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if len(body.UserIds) == 0 {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userIds required"))
		return
	}
	action := string(body.Action)
	actor := authenticatedActor(r)
	outcome, err := h.service.BulkAction(r.Context(), tenantID, actor, action, body.UserIds)
	if err != nil {
		if iamapp.IsForceLogoutDeferred(err) {
			// Spec: 501 with code NOT_IMPLEMENTED + detail "PR-7 dependency".
			detail := "PR-7 dependency"
			h.writeProblem(w, problem.New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Force logout is not yet implemented").WithDetail(detail))
			return
		}
		if errors.Is(err, iamapp.ErrPeopleValidation) {
			h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
			return
		}
		log.Printf("iam people: bulk action failed: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Bulk action failed"))
		return
	}

	// Per-user audit so the trail records each successful mutation. Failures
	// are surfaced in the response body — we do not emit a denied-event per
	// failure; the operator sees them inline.
	for _, userID := range outcome.Succeeded {
		h.recordAudit(r, userID, "iam.user.bulk."+action, map[string]any{
			"action": action,
			"reason": ptrDeref(body.Reason),
		})
	}

	resp := iamapi.UserBulkActionResult{
		Succeeded: outcome.Succeeded,
		Failed:    make([]iamapi.UserBulkActionFailure, 0, len(outcome.Failed)),
	}
	for _, f := range outcome.Failed {
		resp.Failed = append(resp.Failed, iamapi.UserBulkActionFailure{
			UserId:  f.UserID,
			Code:    f.Code,
			Message: f.Message,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// ─── GET /iam/users/{userId}/memberships ──────────────────────────────────

func (h *PeopleHandler) handleListMemberships(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId required"))
		return
	}
	memberships, err := h.service.ListMemberships(r.Context(), tenantID, userID)
	if err != nil {
		h.writePeopleError(w, err)
		return
	}
	resp := iamapi.ListMembershipsResponse{Items: make([]iamapi.AreaMembership, 0, len(memberships))}
	for _, m := range memberships {
		item := iamapi.AreaMembership{
			UserId:   m.UserID,
			AreaCode: m.AreaCode,
			Role:     iamapi.UserRole(string(m.Role)),
		}
		if m.GrantedBy != nil {
			gb := *m.GrantedBy
			item.GrantedBy = &gb
		}
		grantedAt := m.EffectiveFrom
		item.GrantedAt = &grantedAt
		resp.Items = append(resp.Items, item)
	}
	writeJSON(w, http.StatusOK, resp)
}

// ─── helpers ───────────────────────────────────────────────────────────────

func (h *PeopleHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if werr := problem.Write(w, p); werr != nil {
		log.Printf("iam people: write response failed: %v", werr)
	}
}

func (h *PeopleHandler) writePeopleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, iamapp.ErrPeopleValidation):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, iamapp.ErrAreaUnknown):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		h.writeProblem(w, problem.New(http.StatusConflict, "CONFLICT_ERROR", "User already exists"))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "User not found"))
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, iamdomain.ErrInvalidRole):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid role"))
	default:
		log.Printf("iam people: handler error: %v", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process people request"))
	}
}

func (h *PeopleHandler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "User not found"))
	default:
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process auth request"))
	}
}

func (h *PeopleHandler) recordAudit(r *http.Request, userID, action string, payload map[string]any) {
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

func toManagedUserCore(u iamapp.ListedUser) iamapi.ManagedUserCore {
	var email *string
	if strings.TrimSpace(u.Email) != "" {
		e := u.Email
		email = &e
	}
	areas := make([]iamapi.AreaMembership, 0, len(u.AreaMemberships))
	for _, a := range u.AreaMemberships {
		item := iamapi.AreaMembership{
			UserId:   a.UserID,
			AreaCode: a.AreaCode,
			Role:     iamapi.UserRole(string(a.Role)),
		}
		grantedAt := a.EffectiveFrom
		item.GrantedAt = &grantedAt
		if a.GrantedBy != nil {
			gb := *a.GrantedBy
			item.GrantedBy = &gb
		}
		areas = append(areas, item)
	}
	return iamapi.ManagedUserCore{
		UserId:              u.UserID,
		Username:            u.Username,
		DisplayName:         u.DisplayName,
		Email:               email,
		IsActive:            u.IsActive,
		MustChangePassword:  u.MustChangePassword,
		FailedLoginAttempts: u.FailedLoginAttempts,
		LastLoginAt:         u.LastLoginAt,
		LockedUntil:         u.LockedUntil,
		TenantRole:          iamapi.UserRole(string(u.TenantRole)),
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
		AreaMemberships:     areas,
	}
}

func parseOptionalBool(raw string) *bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}

func nonEmptyPtr(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func parseLimit(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func ptrDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
