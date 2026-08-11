package http

import (
	"net/http"
	"strconv"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) updateSchemas(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	actorID, err := userIDFromReq(r)
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidParam, "version must be an integer"))
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	var req struct {
		MetadataSchema      domain.MetadataSchema `json:"metadata_schema"`
		PlaceholderSchema   []domain.Placeholder  `json:"placeholder_schema"`
		ExpectedLockVersion *int                  `json:"expected_lock_version"`
	}
	if err := readJSON(r, &req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error()))
		return
	}
	if req.ExpectedLockVersion == nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, "expected_lock_version is required"))
		return
	}
	if *req.ExpectedLockVersion < 0 {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, "expected_lock_version must be >= 0"))
		return
	}

	v, err := h.svc.UpdateSchemas(r.Context(), application.UpdateSchemasCmd{
		TenantID:            tenantID,
		ActorUserID:         actorID,
		TemplateID:          templateID,
		VersionNumber:       versionNum,
		MetadataSchema:      req.MetadataSchema,
		PlaceholderSchema:   req.PlaceholderSchema,
		ExpectedLockVersion: *req.ExpectedLockVersion,
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
	var resp templatesapi.UpdateTemplateSchema200JSONResponse
	resp.Data.Version = dto
	writeJSON(w, http.StatusOK, resp)
}
