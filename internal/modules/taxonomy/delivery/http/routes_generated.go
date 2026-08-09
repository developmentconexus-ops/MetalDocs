package http

import (
	"net/http"
	"net/url"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
)

var _ taxonomyapi.ServerInterface = (*Handler)(nil)

// ListTaxonomyProfiles handles GET /taxonomy/profiles.
func (h *Handler) ListTaxonomyProfiles(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyProfilesParams) {
	applyIncludeArchivedParam(r, params.IncludeArchived)
	h.listProfiles(w, r)
}

// CreateTaxonomyProfile handles POST /taxonomy/profiles.
func (h *Handler) CreateTaxonomyProfile(w http.ResponseWriter, r *http.Request, _ taxonomyapi.CreateTaxonomyProfileParams) {
	h.createProfile(w, r)
}

// GetTaxonomyProfile handles GET /taxonomy/profiles/{code}.
func (h *Handler) GetTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.getProfile(w, r)
}

// UpdateTaxonomyProfile handles PATCH /taxonomy/profiles/{code}.
func (h *Handler) UpdateTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.updateProfile(w, r)
}

// ArchiveTaxonomyProfile handles DELETE /taxonomy/profiles/{code}.
func (h *Handler) ArchiveTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveProfile(w, r)
}

// SetTaxonomyProfileDefaultTemplate handles PUT /taxonomy/profiles/{code}/default-template.
func (h *Handler) SetTaxonomyProfileDefaultTemplate(w http.ResponseWriter, r *http.Request, code string) {
	h.setDefaultTemplate(w, r)
}

// ListTaxonomyAreas handles GET /taxonomy/areas.
func (h *Handler) ListTaxonomyAreas(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyAreasParams) {
	applyIncludeArchivedParam(r, params.IncludeArchived)
	h.listAreas(w, r)
}

// CreateTaxonomyArea handles POST /taxonomy/areas.
func (h *Handler) CreateTaxonomyArea(w http.ResponseWriter, r *http.Request, _ taxonomyapi.CreateTaxonomyAreaParams) {
	h.createArea(w, r)
}

// GetTaxonomyArea handles GET /taxonomy/areas/{code}.
func (h *Handler) GetTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.getArea(w, r)
}

// UpdateTaxonomyArea handles PUT /taxonomy/areas/{code}.
func (h *Handler) UpdateTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.updateArea(w, r)
}

// ArchiveTaxonomyArea handles DELETE /taxonomy/areas/{code}.
func (h *Handler) ArchiveTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveArea(w, r)
}

// ListTaxonomyFamilies handles GET /taxonomy/families.
func (h *Handler) ListTaxonomyFamilies(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyFamiliesParams) {
	if params.IncludeInactive != nil {
		q, _ := url.ParseQuery(r.URL.RawQuery)
		if *params.IncludeInactive {
			q.Set("include_inactive", "true")
		} else {
			q.Set("include_inactive", "false")
		}
		r.URL.RawQuery = q.Encode()
	}
	h.listFamilies(w, r)
}

// CreateTaxonomyFamily handles POST /taxonomy/families.
func (h *Handler) CreateTaxonomyFamily(w http.ResponseWriter, r *http.Request, _ taxonomyapi.CreateTaxonomyFamilyParams) {
	h.createFamily(w, r)
}

// GetTaxonomyFamily handles GET /taxonomy/families/{code}.
func (h *Handler) GetTaxonomyFamily(w http.ResponseWriter, r *http.Request, code string) {
	h.getFamily(w, r)
}

// UpdateTaxonomyFamily handles PATCH /taxonomy/families/{code}.
func (h *Handler) UpdateTaxonomyFamily(w http.ResponseWriter, r *http.Request, code string) {
	h.updateFamily(w, r)
}

// DeactivateTaxonomyFamily handles DELETE /taxonomy/families/{code}.
func (h *Handler) DeactivateTaxonomyFamily(w http.ResponseWriter, r *http.Request, code string) {
	h.deactivateFamily(w, r)
}

// applyIncludeArchivedParam rewrites r.URL.RawQuery with the typed
// include_archived value from the generated params struct so downstream
// handlers (listProfiles, listAreas — parseIncludeArchived in
// routes_profiles.go) keep reading it via r.URL.Query(), matching the
// ListTaxonomyFamilies/include_inactive adapter pattern above.
func applyIncludeArchivedParam(r *http.Request, includeArchived *bool) {
	if includeArchived == nil {
		return
	}
	q, _ := url.ParseQuery(r.URL.RawQuery)
	if *includeArchived {
		q.Set("include_archived", "true")
	} else {
		q.Set("include_archived", "false")
	}
	r.URL.RawQuery = q.Encode()
}
