package http

import (
	"net/http"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func (h *Handler) createNextVersion(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateCreate)); err != nil {
		writeMappedErr(w, err)
		return
	}

	v, err := h.svc.CreateNextVersion(r.Context(), application.CreateVersionCmd{
		TenantID:    tenantID,
		ActorUserID: actorID,
		TemplateID:  templateID,
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
	writeJSON(w, http.StatusCreated, dto)
}

func toTemplateResponse(t *domain.Template) map[string]any {
	if t == nil {
		return nil
	}

	return map[string]any{
		"id":                       t.ID,
		"tenant_id":                t.TenantID,
		"doc_type_code":            t.DocTypeCode,
		"key":                      t.Key,
		"name":                     t.Name,
		"description":              t.Description,
		"latest_version":           t.LatestVersion,
		"latest_revision_number":   t.LatestRevisionNumber,
		"published_version_id":     t.PublishedVersionID,
		"published_version_number": t.PublishedVersionNumber,
		"current_revision_number":  t.CurrentRevisionNumber,
		"created_by":               t.CreatedBy,
		"created_at":               t.CreatedAt.UTC().Format(time.RFC3339),
		"archived_at":              timePtrRFC3339(t.ArchivedAt),
	}
}

func toVersionResponse(v *domain.TemplateVersion) map[string]any {
	if v == nil {
		return nil
	}

	return map[string]any{
		"id":                    v.ID,
		"template_id":           v.TemplateID,
		"version_number":        v.VersionNumber,
		"revision_number":       v.RevisionNumber,
		"status":                v.Status,
		"docx_storage_key":      v.DocxStorageKey,
		"content_hash":          v.ContentHash,
		"metadata_schema":       v.MetadataSchema,
		"placeholder_schema":    v.PlaceholderSchema,
		"author_id":             v.AuthorID,
		"pending_reviewer_role": v.PendingReviewerRole,
		"pending_approver_role": v.PendingApproverRole,
		"reviewer_id":           v.ReviewerID,
		"approver_id":           v.ApproverID,
		"submitted_at":          timePtrRFC3339(v.SubmittedAt),
		"reviewed_at":           timePtrRFC3339(v.ReviewedAt),
		"approved_at":           timePtrRFC3339(v.ApprovedAt),
		"published_at":          timePtrRFC3339(v.PublishedAt),
		"obsoleted_at":          timePtrRFC3339(v.ObsoletedAt),
		"lock_version":          v.LockVersion,
		"created_at":            v.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func timePtrRFC3339(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
