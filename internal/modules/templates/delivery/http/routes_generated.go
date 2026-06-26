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

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	h.listTemplates(w, r)
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

	tplDTO, err := toAPITemplateDTO(res.Template)
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
	url, key, ok := h.presignTemplateUpload(w, r, id, n, "")
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, templatesapi.PresignTemplateDocxUploadUrl200JSONResponse{
		Url:        &url,
		StorageKey: &key,
	})
}

func (h *Handler) PresignTemplateSchemaUploadUrl(w http.ResponseWriter, r *http.Request, id string, n int) {
	url, key, ok := h.presignTemplateUpload(w, r, id, n, "templates/"+id+"/versions/"+intString(n)+".schema.json")
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, templatesapi.PresignTemplateSchemaUploadUrl200JSONResponse{
		Url:        &url,
		StorageKey: &key,
	})
}

// presignTemplateUpload runs the shared authz + service flow and returns the
// presigned URL and storage key. Each caller writes the op-specific generated
// 200 type so each route is pinned to its strict-server response contract
// (M5/F5.3 H-D remediation; M1/F1.3 declared-fields-only). On error this writes
// the problem+json response and returns ok=false.
func (h *Handler) presignTemplateUpload(w http.ResponseWriter, r *http.Request, id string, n int, storageKey string) (string, string, bool) {
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
		StorageKey:    storageKey,
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
	if field := missingPublishTemplateVersionField(req); field != "" {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, "field "+field+" is required")
		return
	}

	res, err := h.svc.PublishTemplateVersion(r.Context(), application.PublishTemplateVersionCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		ActorRoles:    actorRolesFromReq(r),
		TemplateID:    id,
		VersionNumber: n,
		DocxKey:       strings.TrimSpace(req.DocxKey),
		SchemaKey:     strings.TrimSpace(req.SchemaKey),
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	// Strict-server typed response — exactly the 3 fields declared at
	// openapi.yaml:1331 (M5/F5.3 H-D remediation; closes the M1/F1.3 declared-
	// fields-only leak that emitted an undeclared `published_version_number`).
	writeJSON(w, http.StatusOK, templatesapi.PublishTemplateVersion200JSONResponse{
		PublishedVersionId:  res.PublishedVersion.ID,
		NextDraftId:         res.NextDraft.ID,
		NextDraftVersionNum: res.NextDraft.VersionNumber,
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
