package http

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/domain"
)

// toAPIVersionDTO maps a domain.TemplateVersion to the OpenAPI-generated VersionDTO type.
// F1.2 / ADR 0035 — flat typed wire shape, no envelope. Mirrors the F1.1 toAPICheckpoint
// pattern. The mapper round-trips MetadataSchema / PlaceholderSchema through JSON so the
// wire shape stays driven by the domain JSON tags (single source of truth).
func toAPIVersionDTO(v *domain.TemplateVersion) (templatesapi.VersionDTO, error) {
	if v == nil {
		return templatesapi.VersionDTO{}, fmt.Errorf("toAPIVersionDTO: nil version")
	}
	id, err := uuid.Parse(v.ID)
	if err != nil {
		return templatesapi.VersionDTO{}, fmt.Errorf("version id %q: %w", v.ID, err)
	}
	tplID, err := uuid.Parse(v.TemplateID)
	if err != nil {
		return templatesapi.VersionDTO{}, fmt.Errorf("version template_id %q: %w", v.TemplateID, err)
	}

	metaPtr, err := metadataSchemaToMap(v.MetadataSchema)
	if err != nil {
		return templatesapi.VersionDTO{}, fmt.Errorf("version metadata_schema: %w", err)
	}
	placeholdersPtr, err := placeholdersToSlice(v.PlaceholderSchema)
	if err != nil {
		return templatesapi.VersionDTO{}, fmt.Errorf("version placeholder_schema: %w", err)
	}

	revisionNumber := int32(v.RevisionNumber)
	lockVersion := int32(v.LockVersion)

	dto := templatesapi.VersionDTO{
		Id:                id,
		TemplateId:        tplID,
		VersionNumber:     v.VersionNumber,
		RevisionNumber:    &revisionNumber,
		Status:            templatesapi.VersionDTOStatus(v.Status),
		DocxStorageKey:    strPtrIfNonEmpty(v.DocxStorageKey),
		ContentHash:       strPtrIfNonEmpty(v.ContentHash),
		MetadataSchema:    metaPtr,
		PlaceholderSchema: placeholdersPtr,
		AuthorId:          v.AuthorID,
		LockVersion:       &lockVersion,
		CreatedAt:         v.CreatedAt.UTC(),
	}

	if v.PendingReviewerRole != nil {
		dto.PendingReviewerRole = v.PendingReviewerRole
	}
	if v.PendingApproverRole != "" {
		s := v.PendingApproverRole
		dto.PendingApproverRole = &s
	}
	dto.ReviewerId = v.ReviewerID
	dto.ApproverId = v.ApproverID
	dto.SubmittedAt = utcPtr(v.SubmittedAt)
	dto.ReviewedAt = utcPtr(v.ReviewedAt)
	dto.ApprovedAt = utcPtr(v.ApprovedAt)
	dto.PublishedAt = utcPtr(v.PublishedAt)
	dto.ObsoletedAt = utcPtr(v.ObsoletedAt)

	return dto, nil
}

func strPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func metadataSchemaToMap(m domain.MetadataSchema) (*map[string]interface{}, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func placeholdersToSlice(ps []domain.Placeholder) (*[]map[string]interface{}, error) {
	if ps == nil {
		empty := []map[string]interface{}{}
		return &empty, nil
	}
	raw, err := json.Marshal(ps)
	if err != nil {
		return nil, err
	}
	out := []map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
