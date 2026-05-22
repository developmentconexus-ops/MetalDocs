package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

var (
	writeJSON  = httpresponse.WriteJSON
	writeError = httpresponse.WriteError
)

const (
	codeTaxProfileNotFound         problem.Code = "PROFILE_NOT_FOUND"
	codeTaxProfileArchived         problem.Code = "PROFILE_ARCHIVED"
	codeTaxTemplateNotPublished    problem.Code = "TEMPLATE_NOT_PUBLISHED"
	codeTaxTemplateProfileMismatch problem.Code = "TEMPLATE_PROFILE_MISMATCH"
	codeTaxProfileCodeImmutable    problem.Code = "PROFILE_CODE_IMMUTABLE"
	codeTaxProfileAlreadyExists    problem.Code = "PROFILE_ALREADY_EXISTS"
	codeTaxFamilyNotFound          problem.Code = "FAMILY_NOT_FOUND"
)

type profileUpsertRequest struct {
	Code                     string  `json:"code"`
	FamilyCode               string  `json:"familyCode"`
	Name                     string  `json:"name"`
	Description              string  `json:"description"`
	Alias                    string  `json:"alias"`
	ReviewIntervalDays       int     `json:"reviewIntervalDays"`
	DefaultTemplateVersionID *string `json:"defaultTemplateVersionId"`
	OwnerUserID              *string `json:"ownerUserId"`
	EditableByRole           string  `json:"editableByRole"`
}

type setDefaultTemplateRequest struct {
	TemplateVersionID string `json:"templateVersionId"`
}

func (h *Handler) listProfiles(w http.ResponseWriter, r *http.Request) {
	includeArchived, err := parseIncludeArchived(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "includeArchived must be true or false")
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	items, err := h.profiles.List(r.Context(), tenantID, includeArchived)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "failed to list profiles")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var req profileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	code := strings.TrimSpace(req.Code)
	alias := strings.TrimSpace(req.Alias)
	if alias == "" {
		alias = code
		if len(alias) > 24 {
			alias = alias[:24]
		}
	}
	profile := &domain.DocumentProfile{
		Code:                     code,
		TenantID:                 tenantID,
		FamilyCode:               strings.TrimSpace(req.FamilyCode),
		Name:                     strings.TrimSpace(req.Name),
		Description:              strings.TrimSpace(req.Description),
		Alias:                    alias,
		ReviewIntervalDays:       req.ReviewIntervalDays,
		DefaultTemplateVersionID: req.DefaultTemplateVersionID,
		OwnerUserID:              req.OwnerUserID,
		EditableByRole:           strings.TrimSpace(req.EditableByRole),
	}
	if profile.Code == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "code is required")
		return
	}

	if err := h.profiles.Create(r.Context(), profile); err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, profile)
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	profile, err := h.profiles.Get(r.Context(), tenantID, r.PathValue("code"))
	if err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req profileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "invalid JSON payload")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	updateCode := r.PathValue("code")
	updateAlias := strings.TrimSpace(req.Alias)
	if updateAlias == "" {
		updateAlias = updateCode
		if len(updateAlias) > 24 {
			updateAlias = updateAlias[:24]
		}
	}
	profile := &domain.DocumentProfile{
		Code:                     updateCode,
		TenantID:                 tenantID,
		FamilyCode:               strings.TrimSpace(req.FamilyCode),
		Name:                     strings.TrimSpace(req.Name),
		Description:              strings.TrimSpace(req.Description),
		Alias:                    updateAlias,
		ReviewIntervalDays:       req.ReviewIntervalDays,
		DefaultTemplateVersionID: req.DefaultTemplateVersionID,
		OwnerUserID:              req.OwnerUserID,
		EditableByRole:           strings.TrimSpace(req.EditableByRole),
	}
	if err := h.profiles.Update(r.Context(), profile); err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) setDefaultTemplate(w http.ResponseWriter, r *http.Request) {
	var req setDefaultTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "invalid JSON payload")
		return
	}
	if strings.TrimSpace(req.TemplateVersionID) == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "templateVersionId is required")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	if err := h.profiles.SetDefaultTemplate(
		r.Context(),
		tenantID,
		r.PathValue("code"),
		req.TemplateVersionID,
		actorUserID,
	); err != nil {
		h.writeProfileError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func (h *Handler) archiveProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	if err := h.profiles.Archive(
		r.Context(),
		tenantID,
		r.PathValue("code"),
		actorUserID,
	); err != nil {
		h.writeProfileError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeProfileError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrProfileNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, codeTaxProfileNotFound, "profile not found")
	case errors.Is(err, domain.ErrProfileArchived):
		httpresponse.WriteError(w, http.StatusConflict, codeTaxProfileArchived, "profile is archived")
	case errors.Is(err, domain.ErrTemplateNotPublished):
		httpresponse.WriteError(w, http.StatusConflict, codeTaxTemplateNotPublished, "template version is not published")
	case errors.Is(err, domain.ErrTemplateProfileMismatch):
		httpresponse.WriteError(w, http.StatusConflict, codeTaxTemplateProfileMismatch, "template version belongs to different profile")
	case errors.Is(err, domain.ErrProfileCodeImmutable):
		httpresponse.WriteError(w, http.StatusBadRequest, codeTaxProfileCodeImmutable, "profile code is immutable")
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, pgErr.Message)
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		httpresponse.WriteError(w, http.StatusConflict, codeTaxProfileAlreadyExists, "profile code already exists")
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		httpresponse.WriteError(w, http.StatusConflict, codeTaxFamilyNotFound, "referenced family does not exist")
	default:
		slog.Error("taxonomy profile error", "err", err)
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func parseIncludeArchived(r *http.Request) (bool, error) {
	return parseBool(r.URL.Query().Get("includeArchived"))
}
