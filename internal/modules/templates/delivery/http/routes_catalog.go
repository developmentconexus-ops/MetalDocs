package http

import (
	"net/http"

	iamdomain "metaldocs/internal/modules/iam/domain"
	renderdomain "metaldocs/internal/modules/render/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) listPlaceholderCatalog(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, r, err)
		return
	}
	var entries []templatesapi.PlaceholderCatalogEntry
	for _, t := range renderdomain.ComputedCatalog() {
		if !t.AuthorVisible {
			continue
		}
		entries = append(entries, templatesapi.PlaceholderCatalogEntry{
			Key: t.Key, Label: t.Label, Description: t.Description,
		})
	}
	writeJSON(w, http.StatusOK, templatesapi.PlaceholderCatalogResponse{Items: entries})
}
