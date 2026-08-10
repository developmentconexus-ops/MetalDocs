package http

import (
	"net/http"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) createNextVersion(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateCreate)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	v, err := h.svc.CreateNextVersion(r.Context(), application.CreateVersionCmd{
		TenantID:    tenantID,
		ActorUserID: actorID,
		TemplateID:  templateID,
	})
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}
