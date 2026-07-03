package http

import (
	"net/http"
	"net/url"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
)

var _ taxonomyapi.ServerInterface = (*Handler)(nil)

func (h *Handler) ListTaxonomyProfiles(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyProfilesParams) {
	applyIncludeArchivedParam(r, params.IncludeArchived)
	h.listProfiles(w, r)
}

func (h *Handler) CreateTaxonomyProfile(w http.ResponseWriter, r *http.Request) {
	h.createProfile(w, r)
}

func (h *Handler) GetTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.getProfile(w, r)
}

func (h *Handler) UpdateTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.updateProfile(w, r)
}

func (h *Handler) ArchiveTaxonomyProfile(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveProfile(w, r)
}

func (h *Handler) SetTaxonomyProfileDefaultTemplate(w http.ResponseWriter, r *http.Request, code string) {
	h.setDefaultTemplate(w, r)
}

func (h *Handler) ListTaxonomyAreas(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyAreasParams) {
	applyIncludeArchivedParam(r, params.IncludeArchived)
	h.listAreas(w, r)
}

func (h *Handler) CreateTaxonomyArea(w http.ResponseWriter, r *http.Request) {
	h.createArea(w, r)
}

func (h *Handler) GetTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.getArea(w, r)
}

func (h *Handler) UpdateTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.updateArea(w, r)
}

func (h *Handler) ArchiveTaxonomyArea(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveArea(w, r)
}

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

func (h *Handler) CreateTaxonomyFamily(w http.ResponseWriter, r *http.Request) {
	h.createFamily(w, r)
}

func (h *Handler) GetTaxonomyFamily(w http.ResponseWriter, r *http.Request, code string) {
	h.getFamily(w, r)
}

func (h *Handler) UpdateTaxonomyFamily(w http.ResponseWriter, r *http.Request, code string) {
	h.updateFamily(w, r)
}

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
