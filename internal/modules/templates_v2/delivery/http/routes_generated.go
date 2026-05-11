package http

import (
	"net/http"
	"strconv"
	"strings"

	templatesapi "metaldocs/internal/modules/templates_v2/api"
	"metaldocs/internal/modules/templates_v2/application"
	"metaldocs/internal/modules/templates_v2/domain"
)

var _ templatesapi.ServerInterface = (*Handler)(nil)

func (h *Handler) ListTemplatesV2(w http.ResponseWriter, r *http.Request) {
	h.listTemplates(w, r)
}

func (h *Handler) CreateTemplateV2(w http.ResponseWriter, r *http.Request) {
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

	var req templatesapi.CreateTemplateV2JSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingCreateTemplateV2Field(req); field != "" {
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
		Visibility:   domain.VisibilityPublic,
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

func missingCreateTemplateV2Field(req templatesapi.CreateTemplateV2JSONRequestBody) string {
	switch {
	case strings.TrimSpace(req.Key) == "":
		return "key"
	case strings.TrimSpace(req.Name) == "":
		return "name"
	default:
		return ""
	}
}

func (h *Handler) GetTemplateVersionV2(w http.ResponseWriter, r *http.Request, id string, n int) {
	r.SetPathValue("id", id)
	r.SetPathValue("n", intString(n))
	h.getVersion(w, r)
}

func (h *Handler) PresignTemplateDocxUploadUrlV2(w http.ResponseWriter, r *http.Request, id string, n int) {
	h.presignTemplateUpload(w, r, id, n, "")
}

func (h *Handler) PresignTemplateSchemaUploadUrlV2(w http.ResponseWriter, r *http.Request, id string, n int) {
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

func (h *Handler) RedirectSignedUrlV2(w http.ResponseWriter, r *http.Request, params templatesapi.RedirectSignedUrlV2Params) {
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

func (h *Handler) SaveTemplateDraftV2(w http.ResponseWriter, r *http.Request, id string, n int) {
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

	var req templatesapi.SaveTemplateDraftV2JSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingSaveTemplateDraftV2Field(req); field != "" {
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

func missingSaveTemplateDraftV2Field(req templatesapi.SaveTemplateDraftV2JSONRequestBody) string {
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

func (h *Handler) PublishTemplateVersionV2(w http.ResponseWriter, r *http.Request, id string, n int) {
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

	var req templatesapi.PublishTemplateVersionV2JSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if field := missingPublishTemplateVersionV2Field(req); field != "" {
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

func missingPublishTemplateVersionV2Field(req templatesapi.PublishTemplateVersionV2JSONRequestBody) string {
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
