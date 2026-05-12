package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	registryapi "metaldocs/internal/modules/registry/api"
	"metaldocs/internal/modules/registry/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

func (h *Handler) ListControlledDocuments(w http.ResponseWriter, r *http.Request, params registryapi.ListControlledDocumentsParams) {
	filter, err := filterFromListParams(params)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	items, err := h.svc.List(r.Context(), tenantID, filter)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	respItems, err := controlledDocumentResponses(items)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, registryapi.ListControlledDocuments200JSONResponse{Items: respItems})
}

func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params registryapi.AtomicCreateControlledDocumentParams) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	var req registryapi.CreateAtomicRequest
	if err := decodeStrictJSON(r, &req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if field := missingAtomicCreateField(req); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "field "+field+" is required")
		return
	}

	templateVersionID := uuidStringPtr(req.TemplateVersionId)
	overrideTemplateVersionID := uuidStringPtr(req.OverrideTemplateVersionId)
	formData := map[string]any(nil)
	if req.FormData != nil {
		formData = *req.FormData
	}

	res, err := h.svc.Create(r.Context(), application.CreateControlledDocumentCmd{
		TenantID:                  tenantID,
		ProfileCode:               strings.TrimSpace(req.ProfileCode),
		ProcessAreaCode:           strings.TrimSpace(req.ProcessAreaCode),
		DepartmentCode:            req.DepartmentCode,
		Title:                     strings.TrimSpace(req.Title),
		OwnerUserID:               strings.TrimSpace(req.OwnerUserId),
		ActorUserID:               authn.UserIDFromContext(r.Context()),
		ManualCode:                req.ManualCode,
		ManualCodeReason:          req.ManualCodeReason,
		OverrideTemplateVersionID: overrideTemplateVersionID,
		OverrideTemplateReason:    req.OverrideTemplateReason,
		TemplateVersionID:         templateVersionID,
		DocumentName:              strings.TrimSpace(req.DocumentName),
		FormData:                  formData,
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
		"controlledDocument": res.ControlledDocument,
		"document":           res.DocumentRef,
	})
}

func decodeStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}
		return err
	}
	return nil
}

func missingAtomicCreateField(req registryapi.CreateAtomicRequest) string {
	switch {
	case strings.TrimSpace(req.DocumentName) == "":
		return "documentName"
	case strings.TrimSpace(req.ProfileCode) == "":
		return "profileCode"
	case strings.TrimSpace(req.ProcessAreaCode) == "":
		return "processAreaCode"
	case strings.TrimSpace(req.Title) == "":
		return "title"
	case strings.TrimSpace(req.OwnerUserId) == "":
		return "ownerUserId"
	default:
		return ""
	}
}

func uuidStringPtr(id *openapi_types.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}

func (h *Handler) PreviewControlledDocumentCode(w http.ResponseWriter, r *http.Request, params registryapi.PreviewControlledDocumentCodeParams) {
	profileCode := strings.TrimSpace(params.ProfileCode)
	areaCode := strings.TrimSpace(params.AreaCode)
	if profileCode == "" || areaCode == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "profileCode and areaCode query parameters are required")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	next, err := h.svc.PeekSeq(r.Context(), tenantID, profileCode, areaCode)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, registryapi.PreviewCodeResponse{
		ProfileCode: strings.ToUpper(profileCode),
		AreaCode:    strings.ToUpper(areaCode),
		NextSeq:     next,
		Code:        registrydomain.AutoCode(profileCode, areaCode, next),
	})
}

func (h *Handler) CreateControlledDocumentRevision(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params registryapi.CreateControlledDocumentRevisionParams) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cdID := r.PathValue("id")
	var body registryapi.CreateRevisionRequest
	if err := decodeStrictJSON(r, &body); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if field := missingCreateRevisionField(body); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "field "+field+" is required")
		return
	}
	formData := map[string]any(nil)
	if body.FormData != nil {
		formData = *body.FormData
	}
	ref, err := h.svc.CreateRevision(r.Context(), application.CreateRevisionCmd{
		TenantID:          tenantID,
		CDID:              cdID,
		Name:              strings.TrimSpace(body.Name),
		FormData:          formData,
		TemplateVersionID: uuidStringPtr(body.TemplateVersionId),
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	respRef, err := documentRefResponse(ref)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, registryapi.RevisionResponse{Document: respRef})
}

func missingCreateRevisionField(req registryapi.CreateRevisionRequest) string {
	if strings.TrimSpace(req.Name) == "" {
		return "name"
	}
	return ""
}

func (h *Handler) GetControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	doc, err := h.svc.Get(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	resp, err := controlledDocumentResponse(*doc)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetActiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cdID := r.PathValue("id")

	// FULL OUTER JOIN so we get a row whenever either an active doc or a published
	// doc exists for this controlled document.  If neither exists the query returns
	// no rows and we 404.
	var (
		docID          sql.NullString
		contentHash    sql.NullString
		revisionVer    sql.NullInt64
		approvalState  sql.NullString
		publishedDocID sql.NullString
	)
	err = h.db.QueryRowContext(r.Context(), `
SELECT active.id,
       COALESCE(active.content_hash_at_submit,
                (SELECT r.content_hash FROM document_revisions r
                  WHERE r.document_id = active.id
                  ORDER BY r.created_at DESC LIMIT 1)),
       active.revision_version,
       COALESCE(
         (SELECT CASE ai.status
            WHEN 'in_progress' THEN 'under_review'
            WHEN 'approved'    THEN 'approved'
            WHEN 'scheduled'   THEN 'scheduled'
            WHEN 'rejected'    THEN 'rejected'
            WHEN 'cancelled'   THEN 'cancelled'
          END
          FROM approval_instances ai
          WHERE ai.document_v2_id = active.id
          ORDER BY ai.submitted_at DESC
          LIMIT 1),
         'draft'
       ),
       pub.id::text
  FROM (SELECT id, content_hash_at_submit, revision_version
          FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status IN ('draft','under_review','approved','rejected','scheduled')
         LIMIT 1) active
  FULL OUTER JOIN
       (SELECT id FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status = 'published'
         ORDER BY revision_number DESC
         LIMIT 1) pub ON TRUE`,
		tenantID, cdID,
	).Scan(&docID, &contentHash, &revisionVer, &approvalState, &publishedDocID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpresponse.WriteError(w, http.StatusNotFound, "NO_ACTIVE_INSTANCE", "no active document instance for this controlled document")
			return
		}
		httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// If both sides are NULL the controlled document simply does not exist.
	if !docID.Valid && !publishedDocID.Valid {
		httpresponse.WriteError(w, http.StatusNotFound, "NO_ACTIVE_INSTANCE", "no active document instance for this controlled document")
		return
	}

	var resp registryapi.ActiveDocumentResponse
	if docID.Valid {
		id, err := uuid.Parse(docID.String)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
		resp.DocumentId = &id
	}
	if contentHash.Valid {
		resp.ContentHash = &contentHash.String
	}
	if revisionVer.Valid {
		v := int(revisionVer.Int64)
		resp.RevisionVersion = &v
	}
	if approvalState.Valid {
		state := registryapi.ActiveDocumentResponseApprovalState(approvalState.String)
		resp.ApprovalState = &state
	}
	if publishedDocID.Valid {
		id, err := uuid.Parse(publishedDocID.String)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
		resp.PublishedDocumentId = &id
	}

	// Fetch in-progress approval instance only when an active doc exists.
	if docID.Valid {
		var approvalInstanceID sql.NullString
		_ = h.db.QueryRowContext(r.Context(), `
SELECT id::text
  FROM approval_instances
 WHERE document_v2_id = $1::uuid
   AND tenant_id = $2::uuid
   AND status = 'in_progress'
 ORDER BY submitted_at DESC
 LIMIT 1`,
			docID.String, tenantID,
		).Scan(&approvalInstanceID)
		if approvalInstanceID.Valid {
			id, err := uuid.Parse(approvalInstanceID.String)
			if err != nil {
				httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
				return
			}
			resp.ApprovalInstanceId = &id
		}
	}

	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ObsoleteControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	if err := h.svc.Obsolete(r.Context(), tenantID, r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SupersedeControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	if err := h.svc.Supersede(r.Context(), tenantID, r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func controlledDocumentResponses(docs []registrydomain.ControlledDocument) ([]registryapi.ControlledDocument, error) {
	items := make([]registryapi.ControlledDocument, 0, len(docs))
	for _, doc := range docs {
		item, err := controlledDocumentResponse(doc)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func controlledDocumentResponse(doc registrydomain.ControlledDocument) (registryapi.ControlledDocument, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return registryapi.ControlledDocument{}, err
	}
	tenantID, err := uuid.Parse(doc.TenantID)
	if err != nil {
		return registryapi.ControlledDocument{}, err
	}
	overrideTemplateVersionID, err := optionalUUID(doc.OverrideTemplateVersionID)
	if err != nil {
		return registryapi.ControlledDocument{}, err
	}
	return registryapi.ControlledDocument{
		Id:                        id,
		TenantId:                  tenantID,
		ProfileCode:               doc.ProfileCode,
		ProcessAreaCode:           doc.ProcessAreaCode,
		DepartmentCode:            doc.DepartmentCode,
		Code:                      doc.Code,
		SequenceNum:               doc.SequenceNum,
		Title:                     doc.Title,
		OwnerUserId:               doc.OwnerUserID,
		OverrideTemplateVersionId: overrideTemplateVersionID,
		Status:                    registryapi.ControlledDocumentStatus(doc.Status),
		CreatedAt:                 doc.CreatedAt,
		UpdatedAt:                 doc.UpdatedAt,
	}, nil
}

func documentRefResponse(ref *registrydomain.DocumentRef) (registryapi.DocumentRef, error) {
	id, err := uuid.Parse(ref.ID)
	if err != nil {
		return registryapi.DocumentRef{}, err
	}
	return registryapi.DocumentRef{
		Id:          id,
		ContentHash: ref.ContentHash,
	}, nil
}

func optionalUUID(value *string) (*openapi_types.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registrydomain.ErrCDNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, "CONTROLLED_DOCUMENT_NOT_FOUND", "controlled document not found")
	case errors.Is(err, registrydomain.ErrCDNotActive):
		httpresponse.WriteError(w, http.StatusConflict, "CONTROLLED_DOCUMENT_NOT_ACTIVE", "controlled document is not active")
	case errors.Is(err, registrydomain.ErrCDCodeTaken):
		httpresponse.WriteError(w, http.StatusConflict, "CONTROLLED_DOCUMENT_CODE_TAKEN", "controlled document code already taken")
	case errors.Is(err, registrydomain.ErrCDArchivedCodeReuse):
		httpresponse.WriteError(w, http.StatusConflict, "CONTROLLED_DOCUMENT_CODE_ARCHIVED", "cannot reuse code from archived controlled document")
	case errors.Is(err, registrydomain.ErrManualCodeReasonRequired):
		httpresponse.WriteError(w, http.StatusBadRequest, "MANUAL_CODE_REASON_REQUIRED", "manual code reason is required")
	case errors.Is(err, registrydomain.ErrOverrideReasonRequired):
		httpresponse.WriteError(w, http.StatusBadRequest, "OVERRIDE_REASON_REQUIRED", "override reason is required")
	case errors.Is(err, registrydomain.ErrOverrideTemplateDeleted):
		httpresponse.WriteError(w, http.StatusConflict, "OVERRIDE_TEMPLATE_DELETED", "override template deleted")
	case errors.Is(err, registrydomain.ErrOverrideNotPublished):
		httpresponse.WriteError(w, http.StatusConflict, "OVERRIDE_TEMPLATE_NOT_PUBLISHED", "override template is not published")
	case errors.Is(err, registrydomain.ErrTemplateProfileMismatch):
		_ = problem.Write(w, problem.New(http.StatusUnprocessableEntity, "template_invalid", "template version does not match the document profile"))
	case errors.Is(err, registrydomain.ErrProfileHasNoDefaultTemplate):
		httpresponse.WriteError(w, http.StatusConflict, "PROFILE_NO_DEFAULT_TEMPLATE", "profile has no default template")
	case errors.Is(err, registrydomain.ErrDefaultObsolete):
		httpresponse.WriteError(w, http.StatusConflict, "DEFAULT_TEMPLATE_OBSOLETE", "default template is obsolete")
	case errors.Is(err, taxonomydomain.ErrProfileNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "profile not found")
	case errors.Is(err, taxonomydomain.ErrAreaNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, "AREA_NOT_FOUND", "process area not found")
	case errors.Is(err, taxonomydomain.ErrProfileArchived):
		httpresponse.WriteError(w, http.StatusConflict, "PROFILE_ARCHIVED", "profile is archived")
	case errors.Is(err, taxonomydomain.ErrAreaArchived):
		httpresponse.WriteError(w, http.StatusConflict, "AREA_ARCHIVED", "process area is archived")
	default:
		httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func filterFromListParams(params registryapi.ListControlledDocumentsParams) (application.CDFilter, error) {
	filter := application.CDFilter{}

	if params.ProfileCode != nil {
		value := strings.TrimSpace(*params.ProfileCode)
		if value != "" {
			filter.ProfileCode = &value
		}
	}
	if params.ProcessAreaCode != nil {
		value := strings.TrimSpace(*params.ProcessAreaCode)
		if value != "" {
			filter.ProcessAreaCode = &value
		}
	}
	if params.DepartmentCode != nil {
		value := strings.TrimSpace(*params.DepartmentCode)
		if value != "" {
			filter.DepartmentCode = &value
		}
	}
	if params.OwnerUserId != nil {
		value := strings.TrimSpace(*params.OwnerUserId)
		if value != "" {
			filter.OwnerUserID = &value
		}
	}
	if params.Q != nil {
		value := strings.TrimSpace(*params.Q)
		if value != "" {
			filter.Query = &value
		}
	}
	if params.Status != nil {
		if !params.Status.Valid() {
			return application.CDFilter{}, errors.New("invalid status value")
		}
		status := registrydomain.CDStatus(*params.Status)
		filter.Status = &status
	}
	if params.Limit != nil {
		if *params.Limit < 0 {
			return application.CDFilter{}, errors.New("invalid limit value")
		}
		filter.Limit = *params.Limit
	}
	if params.Offset != nil {
		if *params.Offset < 0 {
			return application.CDFilter{}, errors.New("invalid offset value")
		}
		filter.Offset = *params.Offset
	}

	return filter, nil
}
