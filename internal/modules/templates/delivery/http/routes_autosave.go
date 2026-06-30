package http

import (
	"net/http"
	"strconv"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

func (h *Handler) presignAutosave(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, err)
		return
	}

	res, err := h.svc.PresignAutosave(r.Context(), application.PresignAutosaveCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    templateID,
		VersionNumber: versionNum,
	})
	if err != nil {
		writeMappedErr(w, err)
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
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req struct {
		ExpectedContentHash string `json:"expected_content_hash"`
	}
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
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
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, dto)
}
