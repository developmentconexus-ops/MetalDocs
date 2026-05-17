package http

import (
	"net/http"
	"strconv"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

var _ templatesapi.ServerInterface = (*Handler)(nil)

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	h.listTemplates(w, r)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request, _ templatesapi.CreateTemplateParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", "template.create"); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req templatesapi.CreateTemplateJSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingCreateTemplateField(req); field != "" {
		writeErr(w, http.StatusBadRequest, "invalid_body", "field "+field+" is required")
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	docTypeCode := ""
	if req.DocTypeCode != nil {
		docTypeCode = strings.TrimSpace(*req.DocTypeCode)
	}
	res, err := h.svc.CreateTemplate(r.Context(), application.CreateTemplateCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		Key:          strings.TrimSpace(req.Key),
		Name:         strings.TrimSpace(req.Name),
		Description:  description,
		DocTypeCode:  docTypeCode,
		ApproverRole: "approver",
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         res.Template.ID,
		"version_id": res.Version.ID,
		"data": map[string]any{
			"template": toTemplateResponse(res.Template),
			"version":  toVersionResponse(res.Version),
		},
	})
}

func missingCreateTemplateField(req templatesapi.CreateTemplateJSONRequestBody) string {
	switch {
	case strings.TrimSpace(req.Key) == "":
		return "key"
	case strings.TrimSpace(req.Name) == "":
		return "name"
	default:
		return ""
	}
}

func (h *Handler) GetTemplateVersion(w http.ResponseWriter, r *http.Request, id string, n int) {
	r.SetPathValue("id", id)
	r.SetPathValue("n", intString(n))
	h.getVersion(w, r)
}

func (h *Handler) PresignTemplateDocxUploadUrl(w http.ResponseWriter, r *http.Request, id string, n int) {
	h.presignTemplateUpload(w, r, id, n, "")
}

func (h *Handler) PresignTemplateSchemaUploadUrl(w http.ResponseWriter, r *http.Request, id string, n int) {
	h.presignTemplateUpload(w, r, id, n, "templates/"+id+"/versions/"+intString(n)+".schema.json")
}

func (h *Handler) presignTemplateUpload(w http.ResponseWriter, r *http.Request, id string, n int, storageKey string) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", "template.edit"); err != nil {
		writeMappedErr(w, err)
		return
	}

	res, err := h.svc.PresignTemplateUpload(r.Context(), application.PresignTemplateUploadCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    id,
		VersionNumber: n,
		StorageKey:    storageKey,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"url":         res.UploadURL,
		"storage_key": res.StorageKey,
	})
}

func (h *Handler) RedirectSignedUrl(w http.ResponseWriter, r *http.Request, params templatesapi.RedirectSignedUrlParams) {
	key := strings.TrimSpace(params.Key)
	if key == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "key query parameter is required")
		return
	}
	url, err := h.svc.PresignStoredObject(r.Context(), key)
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) SaveTemplateDraft(w http.ResponseWriter, r *http.Request, id string, n int) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", "template.edit"); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req templatesapi.SaveTemplateDraftJSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingSaveTemplateDraftField(req); field != "" {
		writeErr(w, http.StatusBadRequest, "invalid_body", "field "+field+" is required")
		return
	}

	err = h.svc.SaveTemplateDraft(r.Context(), application.SaveTemplateDraftCmd{
		TenantID:            tenantID,
		ActorUserID:         actorID,
		TemplateID:          id,
		VersionNumber:       n,
		ExpectedLockVersion: req.ExpectedLockVersion,
		DocxStorageKey:      strings.TrimSpace(req.DocxStorageKey),
		SchemaStorageKey:    strings.TrimSpace(req.SchemaStorageKey),
		DocxContentHash:     strings.TrimSpace(req.DocxContentHash),
		SchemaContentHash:   strings.TrimSpace(req.SchemaContentHash),
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func missingSaveTemplateDraftField(req templatesapi.SaveTemplateDraftJSONRequestBody) string {
	switch {
	case strings.TrimSpace(req.DocxStorageKey) == "":
		return "docx_storage_key"
	case strings.TrimSpace(req.SchemaStorageKey) == "":
		return "schema_storage_key"
	case strings.TrimSpace(req.DocxContentHash) == "":
		return "docx_content_hash"
	case strings.TrimSpace(req.SchemaContentHash) == "":
		return "schema_content_hash"
	default:
		return ""
	}
}

func (h *Handler) PublishTemplateVersion(w http.ResponseWriter, r *http.Request, id string, n int, _ templatesapi.PublishTemplateVersionParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", "template.approve"); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req templatesapi.PublishTemplateVersionJSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingPublishTemplateVersionField(req); field != "" {
		writeErr(w, http.StatusBadRequest, "invalid_body", "field "+field+" is required")
		return
	}

	res, err := h.svc.PublishTemplateVersion(r.Context(), application.PublishTemplateVersionCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    id,
		VersionNumber: n,
		DocxKey:       strings.TrimSpace(req.DocxKey),
		SchemaKey:     strings.TrimSpace(req.SchemaKey),
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"published_version_id":     res.PublishedVersion.ID,
		"next_draft_id":            res.NextDraft.ID,
		"next_draft_version_num":   res.NextDraft.VersionNumber,
		"published_version_number": res.PublishedVersion.VersionNumber,
	})
}

func missingPublishTemplateVersionField(req templatesapi.PublishTemplateVersionJSONRequestBody) string {
	switch {
	case strings.TrimSpace(req.DocxKey) == "":
		return "docx_key"
	case strings.TrimSpace(req.SchemaKey) == "":
		return "schema_key"
	default:
		return ""
	}
}

func intString(v int) string {
	return strconv.Itoa(v)
}

func (h *Handler) CreateTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.createNextVersion(w, r)
}

func (h *Handler) UpdateTemplateSchema(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.updateSchemas(w, r)
}

func (h *Handler) PresignTemplateAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.presignAutosave(w, r)
}

func (h *Handler) CommitTemplateAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.commitAutosave(w, r)
}

func (h *Handler) SubmitTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.SubmitTemplateVersionParams) {
	h.submitForReview(w, r)
}

func (h *Handler) ReviewTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.ReviewTemplateVersionParams) {
	h.review(w, r)
}

func (h *Handler) ApproveTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.ApproveTemplateVersionParams) {
	h.approve(w, r)
}

func (h *Handler) ArchiveTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.archiveTemplate(w, r)
}

func (h *Handler) UpsertTemplateApprovalConfig(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.upsertApprovalConfig(w, r)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.getTemplate(w, r)
}

func (h *Handler) GetTemplateDocxUrl(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.getDocxURL(w, r)
}

func (h *Handler) ListTemplateAudit(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.listAudit(w, r)
}

func (h *Handler) ListTemplatePlaceholderCatalog(w http.ResponseWriter, r *http.Request) {
	h.listPlaceholderCatalog(w, r)
}
