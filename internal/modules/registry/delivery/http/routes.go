package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"metaldocs/internal/modules/registry/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/tenant"
)

type createDocRequest struct {
	ProfileCode               string  `json:"profileCode"`
	ProcessAreaCode           string  `json:"processAreaCode"`
	DepartmentCode            *string `json:"departmentCode"`
	Title                     string  `json:"title"`
	OwnerUserID               string  `json:"ownerUserId"`
	ManualCode                *string `json:"manualCode"`
	ManualCodeReason          *string `json:"manualCodeReason"`
	OverrideTemplateVersionID *string `json:"overrideTemplateVersionId"`
	OverrideTemplateReason    *string `json:"overrideTemplateReason"`
}

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

func (h *Handler) createDoc(w http.ResponseWriter, r *http.Request) {
	var req createDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid JSON payload")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title is required")
		return
	}

	res, err := h.svc.Create(r.Context(), application.CreateControlledDocumentCmd{
		TenantID:                  tenantIDFromRequest(r),
		ProfileCode:               strings.TrimSpace(req.ProfileCode),
		ProcessAreaCode:           strings.TrimSpace(req.ProcessAreaCode),
		DepartmentCode:            req.DepartmentCode,
		Title:                     strings.TrimSpace(req.Title),
		OwnerUserID:               strings.TrimSpace(req.OwnerUserID),
		ActorUserID:               authn.UserIDFromContext(r.Context()),
		ManualCode:                req.ManualCode,
		ManualCodeReason:          req.ManualCodeReason,
		OverrideTemplateVersionID: req.OverrideTemplateVersionID,
		OverrideTemplateReason:    req.OverrideTemplateReason,
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, res.ControlledDocument)
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
	tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantID == "" {
		return tenant.DevTenantID
	}
	return tenantID
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
