package http

import (
	"net/http"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

func (h *Handler) archiveTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateArchive)); err != nil {
		writeMappedErr(w, err)
		return
	}

	tpl, err := h.svc.ArchiveTemplate(r.Context(), application.ArchiveCmd{
		TenantID:    tenantID,
		ActorUserID: actorID,
		TemplateID:  templateID,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPITemplateDTO(tpl, h.resolveCreatedByDisplayName(r.Context(), tenantID, tpl.CreatedBy))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.ArchiveTemplateResponse
	resp.Data.Template = dto
	writeJSON(w, http.StatusOK, resp)
}
