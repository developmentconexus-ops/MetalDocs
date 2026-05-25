package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
)

type familyUpsertRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) listFamilies(w http.ResponseWriter, r *http.Request) {
	includeInactive, err := parseBool(r.URL.Query().Get("includeInactive"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "includeInactive must be true or false")
		return
	}
	items, err := h.families.List(r.Context(), includeInactive)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list families")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) createFamily(w http.ResponseWriter, r *http.Request) {
	var req familyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid JSON payload")
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "code is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	f := &domain.DocumentFamily{
		Code:        req.Code,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if err := h.families.Create(r.Context(), f); err != nil {
		h.writeFamilyError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *Handler) getFamily(w http.ResponseWriter, r *http.Request) {
	f, err := h.families.Get(r.Context(), r.PathValue("code"))
	if err != nil {
		h.writeFamilyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) updateFamily(w http.ResponseWriter, r *http.Request) {
	var req familyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid JSON payload")
		return
	}
	f := &domain.DocumentFamily{
		Code:        r.PathValue("code"),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	updated, err := h.families.Update(r.Context(), f)
	if err != nil {
		h.writeFamilyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deactivateFamily(w http.ResponseWriter, r *http.Request) {
	if err := h.families.Deactivate(r.Context(), r.PathValue("code")); err != nil {
		h.writeFamilyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeFamilyError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrFamilyNotFound):
		writeError(w, http.StatusNotFound, "FAMILY_NOT_FOUND", "family not found")
	case errors.Is(err, domain.ErrFamilyAlreadyInactive):
		writeError(w, http.StatusConflict, "FAMILY_ALREADY_INACTIVE", "family is already inactive")
	case errors.Is(err, domain.ErrFamilyHasProfiles):
		writeError(w, http.StatusConflict, "FAMILY_HAS_PROFILES", "family has active profiles and cannot be deactivated")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeError(w, http.StatusConflict, "FAMILY_ALREADY_EXISTS", "family code already exists")
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "request violates data constraints")
	default:
		slog.Error("taxonomy family error", "err", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
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
