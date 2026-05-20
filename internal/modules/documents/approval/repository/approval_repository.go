package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
)

// SignoffInsertResult returned by InsertSignoff.
type SignoffInsertResult struct {
	ID        string
	WasReplay bool // true if ON CONFLICT detected existing matching signoff
}

var ErrScheduledSupersedeConflict = errors.New("approval: scheduled supersede target no longer matches the current published head")

// ScheduledPublishRow remains as a persistence-shape helper for tests and
// shared state assertions, but it is no longer fetched by a legacy scanner path.
type ScheduledPublishRow struct {
	DocumentID           string
	TenantID             string
	ControlledDocumentID string
	SupersededDocumentID sql.NullString
	EffectiveFrom        time.Time
	RevisionVersion      int
	ScheduleGeneration   int64
}

// ApprovalRepository defines all persistence operations for the approval subsystem.
// All mutating methods take *sql.Tx — callers own tx lifecycle (Phase 5 services).
type ApprovalRepository interface {
	InsertInstance(ctx context.Context, tx *sql.Tx, inst domain.Instance) error
	InsertStageInstances(ctx context.Context, tx *sql.Tx, stages []domain.StageInstance) error
	InsertSignoff(ctx context.Context, tx *sql.Tx, s domain.Signoff) (SignoffInsertResult, error)
	LoadSignoffByActor(ctx context.Context, tx *sql.Tx, tenantID, instanceID, actorUserID string) (*domain.Signoff, error)
	LoadInstance(ctx context.Context, tx *sql.Tx, tenantID, id string) (*domain.Instance, error)
	LoadActiveInstanceByDocument(ctx context.Context, tx *sql.Tx, tenantID, docID string) (*domain.Instance, error)
	ValidateScheduledSupersedeTarget(ctx context.Context, tx *sql.Tx, tenantID, documentID, supersededDocumentID string) error
	LoadCurrentPublishedHeadForDocument(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error)
	LoadCurrentPublishedHead(ctx context.Context, tx *sql.Tx, tenantID, controlledDocumentID string) (string, error)
	MarkSuperseded(ctx context.Context, tx *sql.Tx, tenantID, documentID string) error
	UpdateStageStatus(ctx context.Context, tx *sql.Tx, tenantID, stageID string, newStatus, expectedOldStatus domain.StageStatus) error
	UpdateInstanceStatus(ctx context.Context, tx *sql.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error
}
