package infrastructure

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"metaldocs/internal/modules/templates/domain"
)

type rowScanner interface {
	Scan(dest ...any) error
}

// scanTemplateRead scans a template row plus its projected latest/published
// version references into the domain.TemplateRead read model (ADR 0065). The
// aggregate no longer carries revision-number projections; the refs do.
func scanTemplateRead(row rowScanner) (*domain.TemplateRead, error) {
	var (
		t                      domain.TemplateRead
		latestVersionID        sql.NullString
		latestRevisionNumber   sql.NullInt32
		latestStatus           sql.NullString
		publishedVersionID     sql.NullString
		publishedVersionNumber sql.NullInt32
		publishedRevisionNum   sql.NullInt32
		publishedStatus        sql.NullString
		archivedAt             sql.NullTime
	)
	if err := row.Scan(
		&t.ID, &t.TenantID, &t.DocTypeCode, &t.Key, &t.Name, &t.Description,
		&t.LatestVersion, &latestVersionID, &latestRevisionNumber, &latestStatus,
		&publishedVersionID, &publishedVersionNumber, &publishedRevisionNum, &publishedStatus,
		&t.CreatedBy, &t.SystemOwned, &t.CreatedAt, &t.UpdatedAt, &archivedAt,
	); err != nil {
		return nil, err
	}
	// Latest ref: the lv JOIN can only miss on data corruption (latest_version
	// always points at an existing row); degrade to zero-values rather than
	// error, matching the previous NullInt32 behavior.
	t.Latest = domain.VersionRef{
		ID:     latestVersionID.String,
		Number: t.LatestVersion,
		Status: domain.VersionStatus(latestStatus.String),
	}
	if latestRevisionNumber.Valid {
		t.Latest.RevisionNumber = int(latestRevisionNumber.Int32)
	}
	if publishedVersionID.Valid {
		t.PublishedVersionID = &publishedVersionID.String
		t.Published = &domain.VersionRef{
			ID:             publishedVersionID.String,
			Number:         int(publishedVersionNumber.Int32),
			RevisionNumber: int(publishedRevisionNum.Int32),
			Status:         domain.VersionStatus(publishedStatus.String),
		}
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

func unmarshalAuditDetails(raw []byte, details *map[string]any) error {
	if err := json.Unmarshal(raw, details); err != nil {
		return fmt.Errorf("templates repository unmarshal audit details: %w", err)
	}
	return nil
}
