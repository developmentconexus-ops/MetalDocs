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
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

var writeJSON = httpresponse.WriteJSON

// writeError is taxonomy's problem CONSTRUCTOR shorthand: it builds the Problem
// and hands it to the canonical writer. It owns no serialization of its own
// (problem.Respond is the only emitter in the process).
func writeError(w http.ResponseWriter, r *http.Request, status int, code problem.Code, message string) {
	problem.Respond(w, r, problem.New(status, code, message))
}

// Taxonomy document-profile codes.
//
// ADR 0089 execution step 9 (annex §2.6, rows #147-#154): the SCREAMING_SNAKE
// names and their RegisterLegacy escape hatch are gone. Every code now carries a
// semantic family prefix from the closed set, and its status is BOUND to the
// code rather than decided per call site.
//
// PROFILE_NOT_FOUND / PROFILE_ARCHIVED are also emitted by controlleddocuments,
// so their single registration lives in the shared block of the platform catalog
// (collision C-15) — the registry forbids declaring one wire string twice, and a
// module may not import another module's delivery package.
var (
	codeTaxProfileNotFound      = problem.CodeNotFoundDocumentProfile
	codeTaxProfileArchived      = problem.CodeStateDocumentProfileArchived
	codeTaxTemplateNotPublished = problem.Register("taxonomy", "state.template_not_published", 409)

	// annex #150 / R-22 / C-14: "this template version does not belong to that
	// document profile" was 409 here and 422 (`template_invalid`) in
	// controlleddocuments. The caller supplied an invalid template/profile pair;
	// nothing is racing, so 422 wins and the two codes collapse onto this one —
	// declared in the platform catalog's shared block because both modules emit it
	// and neither may import the other's delivery package.
	codeTaxTemplateProfileMismatch = problem.CodeValidationTemplateProfileMismatch

	// annex #151 / R-20: the body parses and a supplied value fails a business
	// rule (the code may not change after creation) — 422, not the pre-0089 400.
	codeTaxProfileCodeImmutable = problem.Register("taxonomy", "validation.profile_code_immutable", 422)

	codeTaxProfileAlreadyExists = problem.Register("taxonomy", "conflict.document_profile_exists", 409)

	// annex #153b / R-26 (ratified into ADR 0089 decision 3): FAMILY_NOT_FOUND
	// named two different conditions. Here the family_code is a value REFERENCED
	// INSIDE a profile body that does not resolve (the 23503 FK arm), which is a
	// supplied-value rejection — 422, not the pre-0089 409. The other condition,
	// where the family IS the request target, keeps 404 as
	// codeTaxFamilyNotFound below.
	codeTaxFamilyUnknown = problem.Register("taxonomy", "validation.family_unknown", 422)

	codeTaxProfileClassRouteConflict = problem.Register("taxonomy", "state.profile_class_route_conflict", 409)
)

// Process-area and document-family codes (annex §2.6b, rows #166-#173 — the
// codes the original §2.6 inventory missed because it read only this file).
//
// AREA_NOT_FOUND / AREA_ARCHIVED are also emitted by controlleddocuments
// (collision C-19), so they share the platform catalog's single registration.
var (
	codeTaxAreaNotFound = problem.CodeNotFoundProcessArea
	codeTaxAreaArchived = problem.CodeStateProcessAreaArchived

	// annex #168/#169 / R-25: both requests parse and a supplied value fails a
	// business rule — a parent_code that would close a cycle, and a code that is
	// immutable after creation. Neither is a syntax/shape defect, so both move
	// 400 -> 422. #169 is the exact sibling of #151 above, which R-20 already
	// moved; leaving it at 400 would recreate the same-condition-different-status
	// defect inside one module.
	codeTaxAreaParentCycle   = problem.Register("taxonomy", "validation.area_parent_cycle", 422)
	codeTaxAreaCodeImmutable = problem.Register("taxonomy", "validation.area_code_immutable", 422)

	codeTaxAreaAlreadyExists = problem.Register("taxonomy", "conflict.process_area_exists", 409)

	// annex #153a / R-26: the family IS the request target of
	// /taxonomy/families/{code} (§1.4 rule 5), so this half of the old
	// FAMILY_NOT_FOUND keeps 404. See codeTaxFamilyUnknown above for the other.
	codeTaxFamilyNotFound = problem.Register("taxonomy", "notfound.document_family", 404)

	codeTaxFamilyAlreadyInactive = problem.Register("taxonomy", "state.document_family_already_inactive", 409)
	codeTaxFamilyHasProfiles     = problem.Register("taxonomy", "state.document_family_has_profiles", 409)
	codeTaxFamilyAlreadyExists   = problem.Register("taxonomy", "conflict.document_family_exists", 409)
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
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "include_archived must be true or false"))
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	items, err := h.profiles.List(r.Context(), tenantID, includeArchived)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to list profiles"))
		return
	}
	// Bulk approval-route readiness for the whole page (never N+1), one read
	// per subject kind, and it fails the request rather than defaulting to
	// false — a wrong badge would misreport which profiles can accept new
	// documents (or new templates, ADR 0086).
	ready, err := h.profiles.RouteReadySubjects(r.Context(), tenantID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to resolve approval route readiness"))
		return
	}
	dtos := make([]taxonomyapi.DocumentProfileItem, len(items))
	for i := range items {
		_, hasRoute := ready.Documents[string(items[i].Code)]
		_, hasTemplateRoute := ready.Templates[string(items[i].Code)]
		dtos[i] = toDocumentProfileItem(&items[i], hasRoute, hasTemplateRoute)
	}
	httpresponse.WriteJSON(w, http.StatusOK, taxonomyapi.ListDocumentProfilesResponse{Items: dtos})
}

func (h *Handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var req profileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
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
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "code is required"))
		return
	}

	if err := h.profiles.Create(r.Context(), profile); err != nil {
		h.writeProfileError(w, r, err)
		return
	}
	// A just-created profile cannot already have an approval route of either
	// kind (routes are keyed by profile code and created separately), so this
	// is a fact, not a default.
	httpresponse.WriteJSON(w, http.StatusCreated, toDocumentProfileItem(profile, false, false))
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	profile, err := h.profiles.Get(r.Context(), tenantID, domain.ProfileCode(r.PathValue("code")))
	if err != nil {
		h.writeProfileError(w, r, err)
		return
	}
	hasRoute, hasTemplateRoute, err := h.profileHasActiveRoute(r, tenantID, profile.Code)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to resolve approval route readiness"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile, hasRoute, hasTemplateRoute))
}

// profileHasActiveRoute resolves the single-profile has_active_route /
// has_active_template_route flags from the same bulk readiness read the list
// route uses, so both surfaces answer from one query shape.
func (h *Handler) profileHasActiveRoute(r *http.Request, tenantID string, code domain.ProfileCode) (bool, bool, error) {
	ready, err := h.profiles.RouteReadySubjects(r.Context(), tenantID)
	if err != nil {
		return false, false, err
	}
	_, hasDocumentRoute := ready.Documents[string(code)]
	_, hasTemplateRoute := ready.Templates[string(code)]
	return hasDocumentRoute, hasTemplateRoute, nil
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req profileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	current, err := h.profiles.Get(r.Context(), tenantID, domain.ProfileCode(r.PathValue("code")))
	if err != nil {
		h.writeProfileError(w, r, err)
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
		h.writeProfileError(w, r, err)
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
			problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
			return
		}
		if err := h.profiles.Reclassify(r.Context(), tenantID, domain.ProfileCode(updateCode), newClass, actorUserID); err != nil {
			h.writeProfileError(w, r, err)
			return
		}
		profile.GovernanceClass = newClass
	} else {
		// Unchanged (or omitted): echo the persisted class in the response.
		profile.GovernanceClass = current.GovernanceClass
	}
	hasRoute, hasTemplateRoute, err := h.profileHasActiveRoute(r, tenantID, profile.Code)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "failed to resolve approval route readiness"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile, hasRoute, hasTemplateRoute))
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
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload"))
		return
	}
	if strings.TrimSpace(req.TemplateVersionID) == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "templateVersionId is required"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	if err := h.profiles.SetDefaultTemplate(
		r.Context(),
		tenantID,
		domain.ProfileCode(r.PathValue("code")),
		req.TemplateVersionID,
		actorUserID,
	); err != nil {
		h.writeProfileError(w, r, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, taxonomyapi.SetTaxonomyProfileDefaultTemplate200Response{})
}

func (h *Handler) archiveProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	if err := h.profiles.Archive(
		r.Context(),
		tenantID,
		domain.ProfileCode(r.PathValue("code")),
		actorUserID,
	); err != nil {
		h.writeProfileError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// editableByRoleMessage renders the friendly 422 detail from the canonical role
// registry so the message can never drift from the enum the contract publishes.
func editableByRoleMessage() string {
	roles := iamtypes.Roles()
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, string(r))
	}
	return "editable_by_role must be one of: " + strings.Join(codes, ", ")
}

func (h *Handler) writeProfileError(w http.ResponseWriter, r *http.Request, err error) {
	var pgErr *pgconn.PgError
	switch {
	// A3.3 (review round 1, finding 1): ProfileService's mutating methods
	// resolve the actor before any mutation work (T1) and return
	// authn.ErrMissingActor when the request context carries none. That is
	// the platform's 401 + auth.unauthenticated, not a taxonomy dialect.
	case errors.Is(err, authn.ErrMissingActor):
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
	case errors.Is(err, domain.ErrProfileNotFound):
		problem.Respond(w, r, problem.New(http.StatusNotFound, codeTaxProfileNotFound, "profile not found"))
	case errors.Is(err, domain.ErrProfileArchived):
		problem.Respond(w, r, problem.New(http.StatusConflict, codeTaxProfileArchived, "profile is archived"))
	case errors.Is(err, domain.ErrTemplateNotPublished):
		problem.Respond(w, r, problem.New(http.StatusConflict, codeTaxTemplateNotPublished, "template version is not published"))
	case errors.Is(err, domain.ErrTemplateProfileMismatch):
		// R-22: 409 -> 422, bound to validation.template_profile_mismatch.
		problem.Respond(w, r, problem.New(http.StatusUnprocessableEntity, codeTaxTemplateProfileMismatch, "template version belongs to different profile"))
	case errors.Is(err, domain.ErrProfileCodeImmutable):
		// R-20: 400 -> 422, bound to validation.profile_code_immutable.
		problem.Respond(w, r, problem.New(http.StatusUnprocessableEntity, codeTaxProfileCodeImmutable, "profile code is immutable"))
	case errors.Is(err, domain.ErrInvalidGovernanceClass):
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "governance_class must be one of: controlado, simples, livre"))
	case errors.Is(err, domain.ErrInvalidEditableByRole):
		// annex R-6: this site answered 422 while carrying request.invalid, whose
		// registered default is 400 — a code/status contradiction. The generic 422
		// is validation.failed (annex row #121); the status now comes from the
		// registration via NewFor instead of being restated here.
		problem.Respond(w, r, problem.NewFor(problem.CodeValidationFailed, editableByRoleMessage()))
	case errors.Is(err, domain.ErrClassChangeRouteConflict):
		problem.Respond(w, r, problem.New(http.StatusConflict, codeTaxProfileClassRouteConflict, "reclassification conflicts with an active approval route"))
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "request violates data constraints"))
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		problem.Respond(w, r, problem.New(http.StatusConflict, codeTaxProfileAlreadyExists, "profile code already exists"))
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		// R-26 / annex #153b: the referenced family_code is a supplied value that
		// does not resolve — 409 -> 422 (behaviour-visible, ratified).
		problem.Respond(w, r, problem.New(http.StatusUnprocessableEntity, codeTaxFamilyUnknown, "referenced family does not exist"))
	default:
		slog.Error("taxonomy profile error", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func parseIncludeArchived(r *http.Request) (bool, error) {
	return parseBool(r.URL.Query().Get("include_archived"))
}
