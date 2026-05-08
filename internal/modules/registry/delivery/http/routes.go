package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	registryapi "metaldocs/internal/modules/registry/api"
	"metaldocs/internal/modules/registry/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/tenant"
)

type activeDocumentResponse struct {
	DocumentID          *string `json:"documentId,omitempty"`
	ApprovalState       *string `json:"approvalState,omitempty"`
	ContentHash         *string `json:"contentHash,omitempty"`
	RevisionVersion     *int    `json:"revisionVersion,omitempty"`
	PublishedDocumentID *string `json:"publishedDocumentId,omitempty"`
	ApprovalInstanceID  *string `json:"approvalInstanceId,omitempty"`
}

func (h *Handler) listDocs(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFilter(r)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	items, err := h.svc.List(r.Context(), tenantIDFromRequest(r), filter)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params registryapi.AtomicCreateControlledDocumentParams) {
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
		TenantID:                  tenantIDFromRequest(r),
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

func (h *Handler) previewCode(w http.ResponseWriter, r *http.Request) {
	profileCode := strings.TrimSpace(r.URL.Query().Get("profileCode"))
	areaCode := strings.TrimSpace(r.URL.Query().Get("areaCode"))
	if profileCode == "" || areaCode == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "profileCode and areaCode query parameters are required")
		return
	}
	tenantID := tenantIDFromRequest(r)
	next, err := h.svc.PeekSeq(r.Context(), tenantID, profileCode, areaCode)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"profileCode": strings.ToUpper(profileCode),
		"areaCode":    strings.ToUpper(areaCode),
		"nextSeq":     next,
		"code":        registrydomain.AutoCode(profileCode, areaCode, next),
	})
}

func (h *Handler) createRevision(w http.ResponseWriter, r *http.Request) {
	cdID := r.PathValue("id")
	var body struct {
		Name              string         `json:"name"`
		FormData          map[string]any `json:"formData"`
		TemplateVersionID *string        `json:"templateVersionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid JSON payload")
		return
	}
	ref, err := h.svc.CreateRevision(r.Context(), application.CreateRevisionCmd{
		TenantID:          tenantIDFromRequest(r),
		CDID:              cdID,
		Name:              body.Name,
		FormData:          body.FormData,
		TemplateVersionID: body.TemplateVersionID,
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{"document": ref})
}

func (h *Handler) getDoc(w http.ResponseWriter, r *http.Request) {
	doc, err := h.svc.Get(r.Context(), tenantIDFromRequest(r), r.PathValue("id"))
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) getActiveDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromRequest(r)
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
	err := h.db.QueryRowContext(r.Context(), `
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

	var resp activeDocumentResponse
	if docID.Valid {
		resp.DocumentID = &docID.String
	}
	if contentHash.Valid {
		resp.ContentHash = &contentHash.String
	}
	if revisionVer.Valid {
		v := int(revisionVer.Int64)
		resp.RevisionVersion = &v
	}
	if approvalState.Valid {
		resp.ApprovalState = &approvalState.String
	}
	if publishedDocID.Valid {
		resp.PublishedDocumentID = &publishedDocID.String
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
			resp.ApprovalInstanceID = &approvalInstanceID.String
		}
	}

	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) obsoleteDoc(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Obsolete(r.Context(), tenantIDFromRequest(r), r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) supersedeDoc(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Supersede(r.Context(), tenantIDFromRequest(r), r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListControlledDocuments(w http.ResponseWriter, r *http.Request, params registryapi.ListControlledDocumentsParams) {
	h.listDocs(w, r)
}

func (h *Handler) PreviewControlledDocumentCode(w http.ResponseWriter, r *http.Request, params registryapi.PreviewControlledDocumentCodeParams) {
	h.previewCode(w, r)
}

func (h *Handler) GetControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.getDoc(w, r)
}

func (h *Handler) GetActiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.getActiveDocument(w, r)
}

func (h *Handler) ObsoleteControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.obsoleteDoc(w, r)
}

func (h *Handler) CreateControlledDocumentRevision(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params registryapi.CreateControlledDocumentRevisionParams) {
	r.SetPathValue("id", id.String())
	h.createRevision(w, r)
}

func (h *Handler) SupersedeControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.supersedeDoc(w, r)
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
		httpresponse.WriteError(w, http.StatusConflict, "TEMPLATE_PROFILE_MISMATCH", "template profile mismatch")
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

func tenantIDFromRequest(r *http.Request) string {
	tid := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tid == "" {
		return tenant.DevTenantID
	}
	return tid
}

func parseFilter(r *http.Request) (application.CDFilter, error) {
	query := r.URL.Query()
	filter := application.CDFilter{}

	if value := strings.TrimSpace(query.Get("profileCode")); value != "" {
		filter.ProfileCode = &value
	}
	if value := strings.TrimSpace(query.Get("processAreaCode")); value != "" {
		filter.ProcessAreaCode = &value
	}
	if value := strings.TrimSpace(query.Get("departmentCode")); value != "" {
		filter.DepartmentCode = &value
	}
	if value := strings.TrimSpace(query.Get("ownerUserId")); value != "" {
		filter.OwnerUserID = &value
	}
	if value := strings.TrimSpace(query.Get("q")); value != "" {
		filter.Query = &value
	}
	if value := strings.TrimSpace(query.Get("status")); value != "" {
		status := registrydomain.CDStatus(value)
		switch status {
		case registrydomain.CDStatusActive, registrydomain.CDStatusObsolete, registrydomain.CDStatusSuperseded:
			filter.Status = &status
		default:
			return application.CDFilter{}, errors.New("invalid status value")
		}
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 0 {
			return application.CDFilter{}, errors.New("invalid limit value")
		}
		filter.Limit = limit
	}
	if value := strings.TrimSpace(query.Get("offset")); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return application.CDFilter{}, errors.New("invalid offset value")
		}
		filter.Offset = offset
	}

	return filter, nil
}
