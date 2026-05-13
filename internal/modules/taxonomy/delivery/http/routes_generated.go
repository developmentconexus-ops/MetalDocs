package http

import (
	"net/http"
	"net/url"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
)

var _ taxonomyapi.ServerInterface = (*Handler)(nil)

func (h *Handler) ListTaxonomyProfilesV2(w http.ResponseWriter, r *http.Request) {
	h.listProfiles(w, r)
}

func (h *Handler) CreateTaxonomyProfileV2(w http.ResponseWriter, r *http.Request) {
	h.createProfile(w, r)
}

func (h *Handler) GetTaxonomyProfileV2(w http.ResponseWriter, r *http.Request, code string) {
	h.getProfile(w, r)
}

func (h *Handler) UpdateTaxonomyProfileV2(w http.ResponseWriter, r *http.Request, code string) {
	h.updateProfile(w, r)
}

func (h *Handler) ArchiveTaxonomyProfileV2(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveProfile(w, r)
}

func (h *Handler) SetTaxonomyProfileDefaultTemplateV2(w http.ResponseWriter, r *http.Request, code string) {
	h.setDefaultTemplate(w, r)
}

func (h *Handler) ListTaxonomyAreasV2(w http.ResponseWriter, r *http.Request) {
	h.listAreas(w, r)
}

func (h *Handler) CreateTaxonomyAreaV2(w http.ResponseWriter, r *http.Request) {
	h.createArea(w, r)
}

func (h *Handler) GetTaxonomyAreaV2(w http.ResponseWriter, r *http.Request, code string) {
	h.getArea(w, r)
}

func (h *Handler) UpdateTaxonomyAreaV2(w http.ResponseWriter, r *http.Request, code string) {
	h.updateArea(w, r)
}

func (h *Handler) ArchiveTaxonomyAreaV2(w http.ResponseWriter, r *http.Request, code string) {
	h.archiveArea(w, r)
}

func (h *Handler) ListTaxonomyFamiliesV2(w http.ResponseWriter, r *http.Request, params taxonomyapi.ListTaxonomyFamiliesV2Params) {
	if params.IncludeInactive != nil {
		q, _ := url.ParseQuery(r.URL.RawQuery)
		if *params.IncludeInactive {
			q.Set("includeInactive", "true")
		} else {
			q.Set("includeInactive", "false")
		}
		r.URL.RawQuery = q.Encode()
	}
	h.listFamilies(w, r)
}

func (h *Handler) CreateTaxonomyFamilyV2(w http.ResponseWriter, r *http.Request) {
	h.createFamily(w, r)
}

func (h *Handler) GetTaxonomyFamilyV2(w http.ResponseWriter, r *http.Request, code string) {
	h.getFamily(w, r)
}

func (h *Handler) UpdateTaxonomyFamilyV2(w http.ResponseWriter, r *http.Request, code string) {
	h.updateFamily(w, r)
}

func (h *Handler) DeactivateTaxonomyFamilyV2(w http.ResponseWriter, r *http.Request, code string) {
	h.deactivateFamily(w, r)
}
