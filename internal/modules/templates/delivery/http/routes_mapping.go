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

// toAPIVersionRef maps a domain.VersionRef to the generated wire value object
// (ADR 0065). Version pointers travel as one nested object, never parallel
// scalars.
func toAPIVersionRef(r domain.VersionRef) (templatesapi.TemplateVersionRef, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return templatesapi.TemplateVersionRef{}, fmt.Errorf("version ref id %q: %w", r.ID, err)
	}
	return templatesapi.TemplateVersionRef{
		Id:             id,
		Number:         r.Number,
		RevisionNumber: int32(r.RevisionNumber),
		Status:         templatesapi.TemplateVersionRefStatus(r.Status),
	}, nil
}

// toAPITemplateDTO maps a domain.TemplateRead read model to the
// OpenAPI-generated TemplateDTO type. ADR 0065 — version pointers are nested
// value objects: latest_version is a required ref; published_version is a
// required-and-nullable ref (nil → marshals "published_version": null, the
// 9f86828b guarantee). SystemOwned is not present in TemplateDTO (not a public
// API field).
//
// displayName is the resolved display name for t.CreatedBy (FE-08). Pass ""
// when unresolved/unavailable (e.g. iamdomain.NoopUserDisplayNameReader{} or a
// miss in the batch lookup) — CreatedByDisplayName is left nil and callers
// fall back to CreatedBy (the raw user id), matching the
// UserDisplayNameReader port's documented miss behavior.
func toAPITemplateDTO(t *domain.TemplateRead, displayName string) (templatesapi.TemplateDTO, error) {
	if t == nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("toAPITemplateDTO: nil template")
	}
	id, err := uuid.Parse(t.ID)
	if err != nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("template id %q: %w", t.ID, err)
	}
	tenantID, err := uuid.Parse(t.TenantID)
	if err != nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("template tenant_id %q: %w", t.TenantID, err)
	}

	latest, err := toAPIVersionRef(t.Latest)
	if err != nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("template latest_version: %w", err)
	}
	dto := templatesapi.TemplateDTO{
		Id:            id,
		TenantId:      tenantID,
		Key:           t.Key,
		Name:          t.Name,
		LatestVersion: latest,
		CreatedBy:     t.CreatedBy,
		CreatedAt:     t.CreatedAt.UTC(),
	}
	if !t.UpdatedAt.IsZero() {
		u := t.UpdatedAt.UTC()
		dto.UpdatedAt = &u
	}
	if displayName != "" {
		dto.CreatedByDisplayName = &displayName
	}

	if t.Description != "" {
		dto.Description = &t.Description
	}
	if t.DocTypeCode != "" {
		dto.DocTypeCode = &t.DocTypeCode
	}
	if t.Published != nil {
		pub, err := toAPIVersionRef(*t.Published)
		if err != nil {
			return templatesapi.TemplateDTO{}, fmt.Errorf("template published_version: %w", err)
		}
		dto.PublishedVersion = &pub
	}
	// t.Published == nil → dto.PublishedVersion stays nil → marshals
	// "published_version": null (present-and-null; no omitempty on the field).
	if t.ArchivedAt != nil {
		u := t.ArchivedAt.UTC()
		dto.ArchivedAt = &u
	}
	return dto, nil
}
