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
	"log/slog"
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

// PeopleHandler owns the People-tab REST surface introduced in PR-4.
// Audit rows for Patch, Invite, Reset, and Unlock are now emitted by the
// application layer (H-3b Stage 2). The audit field is retained for the
// bulk-action path, which is not yet moved to the application layer.
type PeopleHandler struct {
	service *iamapp.PeopleService
	authSvc UserAdminService
	audit   auditdomain.Writer
}

// NewPeopleHandler constructs the handler. audit is used only by the
// bulk-action path (recordAudit); Patch/Invite/Reset/Unlock emit audit rows
// at the application layer instead.
func NewPeopleHandler(service *iamapp.PeopleService, authSvc UserAdminService, audit auditdomain.Writer) *PeopleHandler {
	return &PeopleHandler{service: service, authSvc: authSvc, audit: audit}
}

// ─── GET /iam/users ─────────────────────────────────────────────────────────

func (h *PeopleHandler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	filters := iamapp.ListFilters{
		IsActive: parseOptionalBool(r.URL.Query().Get("is_active")),
		Q:        nonEmptyPtr(r.URL.Query().Get("q")),
		Cursor:   nonEmptyPtr(r.URL.Query().Get("cursor")),
		AreaCode: nonEmptyPtr(r.URL.Query().Get("area_code")),
		Limit:    parseLimit(r.URL.Query().Get("limit")),
	}
	if rawRole := strings.TrimSpace(r.URL.Query().Get("role")); rawRole != "" {
		role := iamdomain.Role(rawRole)
		if !iamdomain.IsValidRole(role) {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid role filter"))
			return
		}
		filters.Role = &role
	}

	result, err := h.service.ListFiltered(r.Context(), tenantID, filters)
	if err != nil {
		if errors.Is(err, iamapp.ErrCursorExpired) {
			problem.Respond(w, r, problem.New(http.StatusGone, problem.CodeRequestCursorExpired, "The pagination cursor refers to an item that is no longer available. Restart from the first page."))
			return
		}
		slog.Error("iam people: list users failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list users"))
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
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	var body iamapi.UserInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
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
		h.writePeopleError(w, r, err)
		return
	}
	// Audit is now emitted by PeopleService.Invite at the application layer (H-3b).

	writeJSON(w, http.StatusCreated, iamapi.UserInviteResponse{
		UserId:       result.UserID,
		TempPassword: result.TempPassword,
	})
}

// ─── PATCH /iam/users/{user_id} ─────────────────────────────────────────────

func (h *PeopleHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
		return
	}
	if !h.guardUserInTenant(w, r, userID) {
		return
	}
	var body iamapi.UpdateManagedUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
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
	if _, err := h.service.PatchAtomic(r.Context(), tenantID, actor, userID, input); err != nil {
		h.writePeopleError(w, r, err)
		return
	}
	// Audit is now emitted in-tx by PeopleService.PatchAtomic (H-3b).

	// Spec (operationId patchUser) declares the 200 body as ManagedUserCore —
	// read the user back so the wire shape matches the generated FE types (A5).
	updated, err := h.service.Get(r.Context(), tenantID, userID)
	if err != nil {
		h.writePeopleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toManagedUserCore(updated))
}

// ─── POST /iam/users/{user_id}/reset-password ──────────────────────────────

func (h *PeopleHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Auth service is not configured"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
		return
	}
	var body iamapi.ResetManagedUserPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}
	if !h.guardUserInTenant(w, r, userID) {
		return
	}
	if err := h.authSvc.AdminResetPassword(r.Context(), userID, body.NewPassword); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	// Audit is now emitted in-tx by auth.Service.AdminResetPassword (H-3b Stage 1).
	writeJSON(w, http.StatusOK, iamapi.ResetManagedUserPasswordResponse{
		UserId:             userID,
		Reset:              true,
		MustChangePassword: true,
	})
}

// ─── POST /iam/users/{user_id}/unlock ──────────────────────────────────────

func (h *PeopleHandler) handleUnlock(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Auth service is not configured"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
		return
	}
	if !h.guardUserInTenant(w, r, userID) {
		return
	}
	if err := h.authSvc.UnlockUser(r.Context(), userID); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	// Audit is now emitted in-tx by auth.Service.UnlockUser (H-3b Stage 1).
	writeJSON(w, http.StatusOK, iamapi.UnlockManagedUserResponse{UserId: userID, Unlocked: true})
}

// ─── POST /iam/users/bulk ──────────────────────────────────────────────────

func (h *PeopleHandler) handleBulk(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	var body iamapi.UserBulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}
	if len(body.UserIds) == 0 {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_ids required"))
		return
	}
	action := string(body.Action)
	actor := authenticatedActor(r)
	outcome, err := h.service.BulkAction(r.Context(), tenantID, actor, action, body.UserIds)
	if err != nil {
		if iamapp.IsForceLogoutDeferred(err) {
			// Spec: 501 with code NOT_IMPLEMENTED + detail "PR-7 dependency".
			detail := "PR-7 dependency"
			problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalNotImplemented, "Force logout is not yet implemented").WithDetail(detail))
			return
		}
		if errors.Is(err, iamapp.ErrPeopleValidation) {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
			return
		}
		slog.Error("iam people: bulk action failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Bulk action failed"))
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

// ─── GET /iam/users/{user_id}/memberships ──────────────────────────────────

func (h *PeopleHandler) handleListMemberships(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if userID == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id required"))
		return
	}
	if !h.guardUserInTenant(w, r, userID) {
		return
	}
	memberships, err := h.service.ListMemberships(r.Context(), tenantID, userID)
	if err != nil {
		h.writePeopleError(w, r, err)
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

// guardUserInTenant verifies the target userID is a member of the caller's
// tenant before delegating to tenant-agnostic auth mutations. Returns true on
// match. On miss writes 404 NOT_FOUND (NOT 403) so cross-tenant probes cannot
// distinguish "exists elsewhere" from "does not exist".
func (h *PeopleHandler) guardUserInTenant(w http.ResponseWriter, r *http.Request, userID string) bool {
	if h.service == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "People service is not configured"))
		return false
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return false
	}
	if err := h.service.VerifyUserInTenant(r.Context(), tenantID, userID); err != nil {
		if errors.Is(err, iamapp.ErrUserNotInTenant) {
			problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "User not found"))
			return false
		}
		slog.Error("iam people: verify user-in-tenant failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to verify user"))
		return false
	}
	return true
}

func (h *PeopleHandler) writePeopleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, iamapp.ErrPeopleValidation):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
	case errors.Is(err, iamapp.ErrAreaUnknown):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
	case errors.Is(err, iamapp.ErrUnknownRole):
		// Invite carries area memberships, so it shares the membership route's
		// vocabulary failure (F-QA4-2). Same sentinel → same 400 UNKNOWN_ROLE
		// as POST /iam/area-memberships, never a 23514-driven 500.
		problem.Respond(w, r, problem.NewFor(problem.CodeValidationRoleUnknown, "Unknown role"))
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		problem.Respond(w, r, problem.New(http.StatusConflict, problem.CodeConflictGeneric, "User already exists"))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "User not found"))
	case errors.Is(err, iamapp.ErrUserNotInTenant):
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "User not found"))
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
	case errors.Is(err, iamdomain.ErrInvalidRole):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid role"))
	default:
		slog.Error("iam people: handler error", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to process people request"))
	}
}

func (h *PeopleHandler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "User not found"))
	default:
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to process auth request"))
	}
}

// recordAudit emits a best-effort (non-tx) audit event from the handler.
// Retained for the bulk-action path; Patch/Invite/Reset/Unlock now emit
// audit at the application layer (H-3b Stage 2).
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
	if err := h.audit.Record(r.Context(), auditdomain.Event{ //cilint:allow-post-commit-audit bulk-action envelope; per-item mutations already audited in-app (H-3b bounded defer)
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
		slog.Warn("audit: failed to record", "action", action, "user_id", userID, "err", err)
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
