package http

import (
	"net/http"
	"strconv"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func (h *Handler) updateSchemas(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_version_number", "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", "template.edit"); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req struct {
		MetadataSchema      domain.MetadataSchema `json:"metadata_schema"`
		PlaceholderSchema   []domain.Placeholder  `json:"placeholder_schema"`
		ExpectedLockVersion *int                  `json:"expected_lock_version"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if req.ExpectedLockVersion == nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", "expected_lock_version is required")
		return
	}
	if *req.ExpectedLockVersion < 0 {
		writeErr(w, http.StatusBadRequest, "invalid_body", "expected_lock_version must be >= 0")
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
		writeMappedErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"version": toVersionResponse(v),
		},
	})
}
