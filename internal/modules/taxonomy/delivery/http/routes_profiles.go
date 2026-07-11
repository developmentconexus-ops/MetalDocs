package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
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
	codeTaxProfileNotFound           problem.Code = "PROFILE_NOT_FOUND"
	codeTaxProfileArchived           problem.Code = "PROFILE_ARCHIVED"
	codeTaxTemplateNotPublished      problem.Code = "TEMPLATE_NOT_PUBLISHED"
	codeTaxTemplateProfileMismatch   problem.Code = "TEMPLATE_PROFILE_MISMATCH"
	codeTaxProfileCodeImmutable      problem.Code = "PROFILE_CODE_IMMUTABLE"
	codeTaxProfileAlreadyExists      problem.Code = "PROFILE_ALREADY_EXISTS"
	codeTaxFamilyNotFound            problem.Code = "FAMILY_NOT_FOUND"
	codeTaxProfileClassRouteConflict problem.Code = "PROFILE_CLASS_ROUTE_CONFLICT"
)

type profileUpsertRequest struct {
	Code                     string  `json:"code"`
	FamilyCode               string  `json:"family_code"`
	Name                     string  `json:"name"`
	Description              string  `json:"description"`
	Alias                    string  `json:"alias"`
	ReviewIntervalDays       int     `json:"review_interval_days"`
	DefaultTemplateVersionID *string `json:"default_template_version_id"`
	OwnerUserID              *string `json:"owner_user_id"`
	EditableByRole           string  `json:"editable_by_role"`
	GovernanceClass          string  `json:"governance_class"`
}

type setDefaultTemplateRequest struct {
	TemplateVersionID string `json:"template_version_id"`
}

func (h *Handler) listProfiles(w http.ResponseWriter, r *http.Request) {
	includeArchived, err := parseIncludeArchived(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "include_archived must be true or false")
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
	dtos := make([]taxonomyapi.DocumentProfileItem, len(items))
	for i := range items {
		dtos[i] = toDocumentProfileItem(&items[i])
	}
	httpresponse.WriteJSON(w, http.StatusOK, taxonomyapi.ListDocumentProfilesResponse{Items: dtos})
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
		Code:                     domain.ProfileCode(code),
		TenantID:                 tenantID,
		FamilyCode:               domain.FamilyCode(strings.TrimSpace(req.FamilyCode)),
		Name:                     strings.TrimSpace(req.Name),
		Description:              strings.TrimSpace(req.Description),
		Alias:                    alias,
		ReviewIntervalDays:       req.ReviewIntervalDays,
		DefaultTemplateVersionID: req.DefaultTemplateVersionID,
		OwnerUserID:              req.OwnerUserID,
		EditableByRole:           strings.TrimSpace(req.EditableByRole),
		GovernanceClass:          governanceClassOrDefault(req.GovernanceClass),
	}
	if profile.Code == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "code is required")
		return
	}

	if err := h.profiles.Create(r.Context(), profile); err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, toDocumentProfileItem(profile))
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	profile, err := h.profiles.Get(r.Context(), tenantID, domain.ProfileCode(r.PathValue("code")))
	if err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile))
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
	current, err := h.profiles.Get(r.Context(), tenantID, domain.ProfileCode(r.PathValue("code")))
	if err != nil {
		h.writeProfileError(w, err)
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
		Code:                     domain.ProfileCode(updateCode),
		TenantID:                 tenantID,
		FamilyCode:               domain.FamilyCode(strings.TrimSpace(req.FamilyCode)),
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

	// Governance class changes NEVER travel through the generic Update path
	// (the repository deliberately leaves governance_class untouched there, so
	// an ordinary edit cannot silently reclassify). When the client supplies a
	// governance_class that differs from the current one, route it through the
	// dedicated reclassification use-case, which the DB reclassify-guard trigger
	// can reject with a 409 if an active approval route would then conflict.
	newClass := governanceClassOrDefault(req.GovernanceClass)
	if req.GovernanceClass != "" && newClass != current.GovernanceClass {
		actorUserID, ok := authn.UserIDFromContext(r.Context())
		if !ok {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
			return
		}
		if err := h.profiles.Reclassify(r.Context(), tenantID, domain.ProfileCode(updateCode), newClass, actorUserID); err != nil {
			h.writeProfileError(w, err)
			return
		}
		profile.GovernanceClass = newClass
	} else {
		// Unchanged (or omitted): echo the persisted class in the response.
		profile.GovernanceClass = current.GovernanceClass
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile))
}

// governanceClassOrDefault trims the request value and defaults an empty class
// to controlado (the server-side default declared in the OpenAPI contract and
// the DB column default). Validation of a non-empty, non-enum value is left to
// the domain (ErrInvalidGovernanceClass) so the mapping stays single-sourced.
func governanceClassOrDefault(raw string) domain.GovernanceClass {
	c := domain.GovernanceClass(strings.TrimSpace(raw))
	if c == "" {
		return domain.GovernanceControlado
	}
	return c
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
		domain.ProfileCode(r.PathValue("code")),
		req.TemplateVersionID,
		actorUserID,
	); err != nil {
		h.writeProfileError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, taxonomyapi.SetTaxonomyProfileDefaultTemplate200Response{})
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
		domain.ProfileCode(r.PathValue("code")),
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
	case errors.Is(err, domain.ErrInvalidGovernanceClass):
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "governance_class must be one of: controlado, simples, livre")
	case errors.Is(err, domain.ErrClassChangeRouteConflict):
		httpresponse.WriteError(w, http.StatusConflict, codeTaxProfileClassRouteConflict, "reclassification conflicts with an active approval route")
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "request violates data constraints")
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
	return parseBool(r.URL.Query().Get("include_archived"))
}
