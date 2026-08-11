package http

import (
	"net/http"
	"strconv"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) presignAutosave(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.svc.PresignAutosave(r.Context(), application.PresignAutosaveCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    templateID,
		VersionNumber: versionNum,
	})
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, templatesapi.TemplatePresignAutosaveResponse{
		UploadUrl:  res.UploadURL,
		StorageKey: res.StorageKey,
		ExpiresAt:  res.ExpiresAt.UTC(),
	})
}

func (h *Handler) commitAutosave(w http.ResponseWriter, r *http.Request) {
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
		ExpectedContentHash string `json:"expected_content_hash"`
	}
	if err := readStrictJSON(r, &req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error()))
		return
	}

	v, err := h.svc.CommitAutosave(r.Context(), application.CommitAutosaveCmd{
		TenantID:            tenantID,
		ActorUserID:         actorID,
		TemplateID:          templateID,
		VersionNumber:       versionNum,
		ExpectedContentHash: req.ExpectedContentHash,
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
	writeJSON(w, http.StatusOK, dto)
}
