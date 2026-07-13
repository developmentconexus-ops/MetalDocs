package http

import (
	"net/http"
	"strconv"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

var _ templatesapi.ServerInterface = (*Handler)(nil)

// ListTemplates handles GET /templates, delegating to the underlying
// h.listTemplates use case which authorizes, applies pagination/doc-type
// filters, and returns the tenant's templates.
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request, params templatesapi.ListTemplatesParams) {
	h.listTemplates(w, r, params)
}

// CreateTemplate handles POST /templates: authorizes CapTemplateCreate,
// validates the required key/name fields, and creates a new template with
// its initial draft version. ADR 0082 phase (a) part 1: this no longer
// accepts or seeds approver/reviewer role bindings.
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
	res, err := h.svc.CreateTemplate(r.Context(), application.CreateTemplateCmd{
		TenantID:    tenantID,
		ActorUserID: actorID,
		Key:         strings.TrimSpace(req.Key),
		Name:        strings.TrimSpace(req.Name),
		Description: description,
		DocTypeCode: docTypeCode,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	// ADR 0065: assemble the read model at the boundary. A just-created
	// template is never published (Published: nil); its latest_version ref is
	// the version the create returned.
	read := &domain.TemplateRead{
		Template: *res.Template,
		Latest: domain.VersionRef{
			ID:             res.Version.ID,
			Number:         res.Version.VersionNumber,
			RevisionNumber: res.Version.RevisionNumber,
			Status:         res.Version.Status,
		},
	}
	tplDTO, err := toAPITemplateDTO(read, h.resolveCreatedByDisplayName(r.Context(), tenantID, res.Template.CreatedBy))
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

// GetTemplateVersion handles GET /templates/{id}/versions/{n}, delegating to
// h.getVersion after re-populating the path values (this variant of the
// generated interface receives id/n as typed parameters rather than via
// r.PathValue).
func (h *Handler) GetTemplateVersion(w http.ResponseWriter, r *http.Request, id string, n int) {
	r.SetPathValue("id", id)
	r.SetPathValue("n", intString(n))
	h.getVersion(w, r)
}

// PresignTemplateDocxUploadUrl handles POST for presigning a DOCX upload URL
// for the given template version, delegating to h.presignTemplateUpload
// (authorizes CapTemplateEdit and asks the application service for a
// presigned upload URL and storage key).
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

// PresignTemplateSchemaUploadUrl handles POST for presigning a schema-file
// upload URL for the given template version: authorizes CapTemplateEdit and
// returns a presigned URL and storage key from the application service.
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

// PublishTemplateVersion handles POST /templates/{id}/versions/{n}/publish:
// authorizes CapTemplatePublish and publishes the given approved version via
// the application service (bodyless — the schema object key is derived
// server-side), returning the published version's id.
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

	res, err := h.svc.PublishTemplateVersion(r.Context(), application.PublishTemplateVersionCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    id,
		VersionNumber: n,
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

// CreateTemplateVersion handles POST /templates/{id}/versions, delegating to
// the underlying create-next-version use case (h.createNextVersion).
func (h *Handler) CreateTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.createNextVersion(w, r)
}

// UpdateTemplateSchema handles PUT/PATCH for a template version's schema,
// delegating to the underlying update-schemas use case (h.updateSchemas).
func (h *Handler) UpdateTemplateSchema(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.updateSchemas(w, r)
}

// PresignTemplateAutosave handles the autosave presign route, delegating to
// the underlying presign-autosave use case (h.presignAutosave).
func (h *Handler) PresignTemplateAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.presignAutosave(w, r)
}

// CommitTemplateAutosave handles the autosave commit route, delegating to
// the underlying commit-autosave use case (h.commitAutosave).
func (h *Handler) CommitTemplateAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.commitAutosave(w, r)
}

// SubmitTemplateVersion handles POST /templates/{id}/versions/{n}/submit,
// delegating to the underlying submit-for-review use case (h.submitForReview).
func (h *Handler) SubmitTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.SubmitTemplateVersionParams) {
	h.submitForReview(w, r)
}

// ReviewTemplateVersion handles POST /templates/{id}/versions/{n}/review,
// delegating to the underlying review use case (h.review).
func (h *Handler) ReviewTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.ReviewTemplateVersionParams) {
	h.review(w, r)
}

// ApproveTemplateVersion handles POST /templates/{id}/versions/{n}/approve,
// delegating to the underlying approve use case (h.approve).
func (h *Handler) ApproveTemplateVersion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int, _ templatesapi.ApproveTemplateVersionParams) {
	h.approve(w, r)
}

// ArchiveTemplate handles the archive-template route, delegating to the
// underlying archive-template use case (h.archiveTemplate).
func (h *Handler) ArchiveTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.archiveTemplate(w, r)
}

// UpsertTemplateApprovalConfig handles the approval-config upsert route,
// delegating to the underlying upsert-approval-config use case
// (h.upsertApprovalConfig).
func (h *Handler) UpsertTemplateApprovalConfig(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.upsertApprovalConfig(w, r)
}

// GetTemplate handles GET /templates/{id}, delegating to the underlying
// get-template use case (h.getTemplate).
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.getTemplate(w, r)
}

// GetTemplateDocxUrl handles the DOCX download-URL route for a template
// version, delegating to the underlying get-docx-URL use case (h.getDocxURL).
func (h *Handler) GetTemplateDocxUrl(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, n int) {
	h.getDocxURL(w, r)
}

// ListTemplateAudit handles the template audit-log listing route,
// delegating to the underlying list-audit use case (h.listAudit).
func (h *Handler) ListTemplateAudit(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.listAudit(w, r)
}

// ListTemplatePlaceholderCatalog handles the placeholder-catalog listing
// route, delegating to the underlying list-placeholder-catalog use case
// (h.listPlaceholderCatalog).
func (h *Handler) ListTemplatePlaceholderCatalog(w http.ResponseWriter, r *http.Request) {
	h.listPlaceholderCatalog(w, r)
}
