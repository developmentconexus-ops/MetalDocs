package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"metaldocs/internal/modules/templates/domain"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTemplate(row rowScanner) (*domain.Template, error) {
	var (
		t                       domain.Template
		publishedVersionID      sql.NullString
		publishedVersionNumber  sql.NullInt32
		currentRevisionNumber   sql.NullInt32
		archivedAt              sql.NullTime
	)
	if err := row.Scan(
		&t.ID, &t.TenantID, &t.DocTypeCode, &t.Key, &t.Name, &t.Description,
		&t.LatestVersion, &publishedVersionID, &publishedVersionNumber, &currentRevisionNumber,
		&t.CreatedBy, &t.SystemOwned, &t.CreatedAt, &archivedAt,
	); err != nil {
		return nil, err
	}
	if publishedVersionID.Valid {
		t.PublishedVersionID = &publishedVersionID.String
	}
	if publishedVersionNumber.Valid {
		n := int(publishedVersionNumber.Int32)
		t.PublishedVersionNumber = &n
	}
	if currentRevisionNumber.Valid {
		n := int(currentRevisionNumber.Int32)
		t.CurrentRevisionNumber = &n
	}
	if archivedAt.Valid {
		t.ArchivedAt = &archivedAt.Time
	}
	return &t, nil
}

func scanTemplateVersion(row rowScanner) (*domain.TemplateVersion, error) {
	var (
		v                   domain.TemplateVersion
		status              string
		metadataJSON        []byte
		placeholderJSON     []byte
		pendingReviewerRole sql.NullString
		reviewerID          sql.NullString
		approverID          sql.NullString
		submittedAt         sql.NullTime
		reviewedAt          sql.NullTime
		approvedAt          sql.NullTime
		publishedAt         sql.NullTime
		obsoletedAt         sql.NullTime
	)
	if err := row.Scan(
		&v.ID, &v.TemplateID, &v.VersionNumber, &v.RevisionNumber, &status, &v.DocxStorageKey, &v.ContentHash,
		&metadataJSON, &placeholderJSON, &v.AuthorID,
		&pendingReviewerRole, &v.PendingApproverRole, &reviewerID, &approverID,
		&submittedAt, &reviewedAt, &approvedAt, &publishedAt, &obsoletedAt, &v.LockVersion, &v.CreatedAt,
	); err != nil {
		return nil, err
	}
	v.Status = domain.VersionStatus(status)
	if pendingReviewerRole.Valid {
		v.PendingReviewerRole = &pendingReviewerRole.String
	}
	if reviewerID.Valid {
		v.ReviewerID = &reviewerID.String
	}
	if approverID.Valid {
		v.ApproverID = &approverID.String
	}
	if submittedAt.Valid {
		v.SubmittedAt = &submittedAt.Time
	}
	if reviewedAt.Valid {
		v.ReviewedAt = &reviewedAt.Time
	}
	if approvedAt.Valid {
		v.ApprovedAt = &approvedAt.Time
	}
	if publishedAt.Valid {
		v.PublishedAt = &publishedAt.Time
	}
	if obsoletedAt.Valid {
		v.ObsoletedAt = &obsoletedAt.Time
	}
	if err := json.Unmarshal(metadataJSON, &v.MetadataSchema); err != nil {
		return nil, fmt.Errorf("templates repository scan version metadata schema: %w", err)
	}
	if err := json.Unmarshal(placeholderJSON, &v.PlaceholderSchema); err != nil {
		return nil, fmt.Errorf("templates repository scan version placeholder schema: %w", err)
	}
	return &v, nil
}

func marshalVersionSchemas(v *domain.TemplateVersion) (metadataJSON, placeholderJSON []byte, err error) {
	metadataJSON, err = json.Marshal(v.MetadataSchema)
	if err != nil {
		return nil, nil, fmt.Errorf("templates repository marshal version metadata schema: %w", err)
	}
	placeholderJSON, err = json.Marshal(v.PlaceholderSchema)
	if err != nil {
		return nil, nil, fmt.Errorf("templates repository marshal version placeholder schema: %w", err)
	}
	return metadataJSON, placeholderJSON, nil
}

func marshalAuditDetails(details map[string]any) ([]byte, error) {
	if details == nil {
		return []byte("{}"), nil
	}
	raw, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("templates repository marshal audit details: %w", err)
	}
	return raw, nil
}

func unmarshalAuditDetails(raw []byte, details *map[string]any) error {
	if err := json.Unmarshal(raw, details); err != nil {
		return fmt.Errorf("templates repository unmarshal audit details: %w", err)
	}
	return nil
}
