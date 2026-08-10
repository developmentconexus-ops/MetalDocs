package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/problem"
)

type listFamiliesResponse struct {
	Items []domain.DocumentFamily `json:"items"`
}

type familyUpsertRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) listFamilies(w http.ResponseWriter, r *http.Request) {
	includeInactive, err := parseBool(r.URL.Query().Get("include_inactive"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "include_inactive must be true or false")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	items, err := h.families.List(r.Context(), tenantID, includeInactive)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to list families")
		return
	}
	writeJSON(w, http.StatusOK, listFamiliesResponse{Items: items})
}

func (h *Handler) createFamily(w http.ResponseWriter, r *http.Request) {
	var req familyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "code is required")
		return
	}
	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "name is required")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	f := &domain.DocumentFamily{
		Code:        domain.FamilyCode(req.Code),
		TenantID:    tenantID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if err := h.families.Create(r.Context(), f); err != nil {
		h.writeFamilyError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *Handler) getFamily(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	f, err := h.families.Get(r.Context(), tenantID, domain.FamilyCode(r.PathValue("code")))
	if err != nil {
		h.writeFamilyError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) updateFamily(w http.ResponseWriter, r *http.Request) {
	var req familyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	f := &domain.DocumentFamily{
		Code:        domain.FamilyCode(r.PathValue("code")),
		TenantID:    tenantID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	updated, err := h.families.Update(r.Context(), f)
	if err != nil {
		h.writeFamilyError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deactivateFamily(w http.ResponseWriter, r *http.Request) {
	if err := h.families.Deactivate(r.Context(), domain.FamilyCode(r.PathValue("code"))); err != nil {
		h.writeFamilyError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeFamilyError(w http.ResponseWriter, r *http.Request, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrFamilyNotFound):
		writeError(w, r, http.StatusNotFound, codeTaxFamilyNotFound, "family not found")
	case errors.Is(err, domain.ErrFamilyAlreadyInactive):
		writeError(w, r, http.StatusConflict, codeTaxFamilyAlreadyInactive, "family is already inactive")
	case errors.Is(err, domain.ErrFamilyHasProfiles):
		writeError(w, r, http.StatusConflict, codeTaxFamilyHasProfiles, "family has active profiles and cannot be deactivated")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeError(w, r, http.StatusConflict, codeTaxFamilyAlreadyExists, "family code already exists")
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, "request violates data constraints")
	default:
		slog.Error("taxonomy family error", "err", err)
		writeError(w, r, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
	}
}

func parseBool(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return false, nil
	}
	switch s {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, errors.New("invalid bool")
	}
}
