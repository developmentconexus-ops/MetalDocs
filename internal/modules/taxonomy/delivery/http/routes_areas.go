package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/problem"
)

type listAreasResponse struct {
	Items []domain.ProcessArea `json:"items"`
}

type areaUpsertRequest struct {
	Code                string  `json:"code"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	ParentCode          *string `json:"parent_code"`
	OwnerUserID         *string `json:"owner_user_id"`
	DefaultApproverRole *string `json:"default_approver_role"`
}

func (h *Handler) listAreas(w http.ResponseWriter, r *http.Request) {
	includeArchived, err := parseIncludeArchived(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "include_archived must be true or false")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	items, err := h.areas.List(r.Context(), tenantID, includeArchived)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to list areas")
		return
	}
	writeJSON(w, http.StatusOK, listAreasResponse{Items: items})
}

func (h *Handler) createArea(w http.ResponseWriter, r *http.Request) {
	var req areaUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	area := &domain.ProcessArea{
		Code:                domain.AreaCode(strings.TrimSpace(req.Code)),
		TenantID:            tenantID,
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		ParentCode:          areaCodePtr(req.ParentCode),
		OwnerUserID:         req.OwnerUserID,
		DefaultApproverRole: req.DefaultApproverRole,
	}
	if area.Code == "" {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "code is required")
		return
	}

	if err := h.areas.Create(r.Context(), area); err != nil {
		h.writeAreaError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, area)
}

func (h *Handler) getArea(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	area, err := h.areas.Get(r.Context(), tenantID, domain.AreaCode(r.PathValue("code")))
	if err != nil {
		h.writeAreaError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *Handler) updateArea(w http.ResponseWriter, r *http.Request) {
	var req areaUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	if _, err := h.areas.Get(r.Context(), tenantID, domain.AreaCode(r.PathValue("code"))); err != nil {
		h.writeAreaError(w, r, err)
		return
	}

	area := &domain.ProcessArea{
		Code:                domain.AreaCode(r.PathValue("code")),
		TenantID:            tenantID,
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		ParentCode:          areaCodePtr(req.ParentCode),
		OwnerUserID:         req.OwnerUserID,
		DefaultApproverRole: req.DefaultApproverRole,
	}
	if err := h.areas.Update(r.Context(), area); err != nil {
		h.writeAreaError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *Handler) archiveArea(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	if err := h.areas.Archive(
		r.Context(),
		tenantID,
		domain.AreaCode(r.PathValue("code")),
		actorUserID,
	); err != nil {
		h.writeAreaError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func areaCodePtr(v *string) *domain.AreaCode {
	if v == nil {
		return nil
	}
	code := domain.AreaCode(*v)
	return &code
}

// defaultApproverRoleMessage renders the friendly 422 detail from the canonical
// area-role registry so the message can never drift from the AreaRole enum the
// contract publishes.
func defaultApproverRoleMessage() string {
	roles := iamtypes.AreaRoles()
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, string(r))
	}
	return "default_approver_role must be one of: " + strings.Join(codes, ", ")
}

func (h *Handler) writeAreaError(w http.ResponseWriter, r *http.Request, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrAreaNotFound):
		writeError(w, r, http.StatusNotFound, codeTaxAreaNotFound, "process area not found")
	case errors.Is(err, domain.ErrAreaArchived):
		writeError(w, r, http.StatusConflict, codeTaxAreaArchived, "process area is archived")
	case errors.Is(err, domain.ErrAreaParentCycle):
		// R-25: 400 -> 422, bound to validation.area_parent_cycle.
		writeError(w, r, http.StatusUnprocessableEntity, codeTaxAreaParentCycle, "area parent assignment creates cycle")
	case errors.Is(err, domain.ErrAreaParentCodeRequired):
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "parentCode is required")
	case errors.Is(err, domain.ErrAreaCodeImmutable):
		// R-25: 400 -> 422, bound to validation.area_code_immutable.
		writeError(w, r, http.StatusUnprocessableEntity, codeTaxAreaCodeImmutable, "area code is immutable")
	case errors.Is(err, domain.ErrInvalidDefaultApproverRole):
		// annex R-6: this site answered 422 while carrying request.invalid, whose
		// registered default is 400 — a code/status contradiction. The generic 422
		// is validation.failed (annex row #121); the status now comes from the
		// registration via NewFor instead of being restated here.
		problem.Respond(w, r, problem.NewFor(problem.CodeValidationFailed, defaultApproverRoleMessage()))
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "request violates data constraints")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeError(w, r, http.StatusConflict, codeTaxAreaAlreadyExists, "area code already exists")
	default:
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
	}
}
