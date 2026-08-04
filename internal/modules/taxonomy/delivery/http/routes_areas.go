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
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "include_archived must be true or false")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	items, err := h.areas.List(r.Context(), tenantID, includeArchived)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to list areas")
		return
	}
	writeJSON(w, http.StatusOK, listAreasResponse{Items: items})
}

func (h *Handler) createArea(w http.ResponseWriter, r *http.Request) {
	var req areaUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "code is required")
		return
	}

	if err := h.areas.Create(r.Context(), area); err != nil {
		h.writeAreaError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, area)
}

func (h *Handler) getArea(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	area, err := h.areas.Get(r.Context(), tenantID, domain.AreaCode(r.PathValue("code")))
	if err != nil {
		h.writeAreaError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *Handler) updateArea(w http.ResponseWriter, r *http.Request) {
	var req areaUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	if _, err := h.areas.Get(r.Context(), tenantID, domain.AreaCode(r.PathValue("code"))); err != nil {
		h.writeAreaError(w, err)
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
		h.writeAreaError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *Handler) archiveArea(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	if err := h.areas.Archive(
		r.Context(),
		tenantID,
		domain.AreaCode(r.PathValue("code")),
		actorUserID,
	); err != nil {
		h.writeAreaError(w, err)
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

func (h *Handler) writeAreaError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrAreaNotFound):
		writeError(w, http.StatusNotFound, codeTaxAreaNotFound, "process area not found")
	case errors.Is(err, domain.ErrAreaArchived):
		writeError(w, http.StatusConflict, codeTaxAreaArchived, "process area is archived")
	case errors.Is(err, domain.ErrAreaParentCycle):
		writeError(w, http.StatusBadRequest, codeTaxAreaParentCycle, "area parent assignment creates cycle")
	case errors.Is(err, domain.ErrAreaParentCodeRequired):
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "parentCode is required")
	case errors.Is(err, domain.ErrAreaCodeImmutable):
		writeError(w, http.StatusBadRequest, codeTaxAreaCodeImmutable, "area code is immutable")
	case errors.Is(err, domain.ErrInvalidDefaultApproverRole):
		writeError(w, http.StatusUnprocessableEntity, problem.CodeRequestInvalid, defaultApproverRoleMessage())
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		writeError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "request violates data constraints")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeError(w, http.StatusConflict, codeTaxAreaAlreadyExists, "area code already exists")
	default:
		writeError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
	}
}
