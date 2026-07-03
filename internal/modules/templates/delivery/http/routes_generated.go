package http

import (
	"net/http"
	"strconv"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

var _ templatesapi.ServerInterface = (*Handler)(nil)

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request, params templatesapi.ListTemplatesParams) {
	h.listTemplates(w, r, params)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request, _ templatesapi.CreateTemplateParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateCreate)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req templatesapi.CreateTemplateJSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}
	if field := missingCreateTemplateField(req); field != "" {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, "field "+field+" is required")
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
	// approver_role is now caller-configurable (F-T4); default to "approver"
	// server-side when omitted or blank so the creation contract is self-contained
	// without changing the historical default binding.
	approverRole := "approver"
	if req.ApproverRole != nil && strings.TrimSpace(*req.ApproverRole) != "" {
		approverRole = strings.TrimSpace(*req.ApproverRole)
	}
	var reviewerRole *string
	if req.ReviewerRole != nil {
		if trimmed := strings.TrimSpace(*req.ReviewerRole); trimmed != "" {
			reviewerRole = &trimmed
		}
	}
	res, err := h.svc.CreateTemplate(r.Context(), application.CreateTemplateCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		Key:          strings.TrimSpace(req.Key),
		Name:         strings.TrimSpace(req.Name),
		Description:  description,
		DocTypeCode:  docTypeCode,
		ApproverRole: approverRole,
		ReviewerRole: reviewerRole,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	tplDTO, err := toAPITemplateDTO(res.Template, h.resolveCreatedByDisplayName(r.Context(), tenantID, res.Template.CreatedBy))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	vDTO, err := toAPIVersionDTO(res.Version)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.CreateTemplateResponse
	resp.Data.Template = tplDTO
	resp.Data.Version = vDTO
	writeJSON(w, http.StatusCreated, resp)
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
	url, key, ok := h.presignTemplateUpload(w, r, id, n)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, templatesapi.PresignTemplateDocxUploadUrl200JSONResponse{
		Url:        &url,
		StorageKey: &key,
	})
}

func (h *Handler) PresignTemplateSchemaUploadUrl(w http.ResponseWriter, r *http.Request, id string, n int) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, err)
		return
	}
	res, err := h.svc.PresignTemplateSchemaUpload(r.Context(), application.PresignTemplateUploadCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    id,
		VersionNumber: n,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, templatesapi.PresignTemplateSchemaUploadUrl200JSONResponse{
		Url:        &res.UploadURL,
		StorageKey: &res.StorageKey,
	})
}

// presignTemplateUpload runs the shared authz + service flow for docx uploads
// and returns the presigned URL and storage key. On error writes the problem+json
// response and returns ok=false.
func (h *Handler) presignTemplateUpload(w http.ResponseWriter, r *http.Request, id string, n int) (string, string, bool) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return "", "", false
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, err)
		return "", "", false
	}

	res, err := h.svc.PresignTemplateUpload(r.Context(), application.PresignTemplateUploadCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    id,
		VersionNumber: n,
	})
	if err != nil {
		writeMappedErr(w, err)
		return "", "", false
	}
	return res.UploadURL, res.StorageKey, true
}


func (h *Handler) PublishTemplateVersion(w http.ResponseWriter, r *http.Request, id string, n int, _ templatesapi.PublishTemplateVersionParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplatePublish)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req templatesapi.PublishTemplateVersionJSONRequestBody
	if err := readStrictJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}
	schemaKey := strings.TrimSpace(req.SchemaKey)
	if schemaKey == "" {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, "field schema_key is required")
		return
	}

	res, err := h.svc.PublishTemplateVersion(r.Context(), application.PublishTemplateVersionCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		ActorRoles:    actorRolesFromReq(r),
		TemplateID:    id,
		VersionNumber: n,
		SchemaKey:     schemaKey,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	// Strict-server typed response — exactly the 1 field declared at
	// openapi.yaml (M1·T2: next_draft_* removed; published_version_id only).
	writeJSON(w, http.StatusOK, templatesapi.PublishTemplateVersion200JSONResponse{
		PublishedVersionId: res.PublishedVersion.ID,
	})
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
