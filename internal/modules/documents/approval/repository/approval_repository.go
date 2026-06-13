package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/platform/db"
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

// Route is the repository projection for approval route administration lists.
type Route struct {
	ID          string
	Name        string
	TenantID    string
	ProfileCode string
	Active      bool
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Stages      []RouteStage
	Total       int
}

// RouteStage is the repository projection for an approval route stage.
type RouteStage struct {
	Order              int
	Name               string
	RequiredRole       string
	RequiredCapability string
	AreaCode           string
	Quorum             string
	QuorumM            *int
	DriftPolicy        string
}

// ApprovalRepository defines all persistence operations for the approval subsystem.
// All mutating methods take db.Tx — callers own tx lifecycle (Phase 5 services).
type ApprovalRepository interface {
	InsertInstance(ctx context.Context, tx db.Tx, inst domain.Instance) error
	InsertStageInstances(ctx context.Context, tx db.Tx, stages []domain.StageInstance) error
	InsertSignoff(ctx context.Context, tx db.Tx, s domain.Signoff) (SignoffInsertResult, error)
	LoadSignoffByActor(ctx context.Context, tx db.Tx, tenantID, instanceID, actorUserID string) (*domain.Signoff, error)
	LoadInstance(ctx context.Context, tx db.Tx, tenantID, id string) (*domain.Instance, error)
	// LoadInstancesByIDs batch-loads multiple approval instances in a single
	// query set. Order of the returned slice matches ids. Missing IDs (tenant
	// mismatch or not found) are silently omitted.
	LoadInstancesByIDs(ctx context.Context, tx db.Tx, tenantID string, ids []string) ([]domain.Instance, error)
	LoadActiveInstanceByDocument(ctx context.Context, tx db.Tx, tenantID, docID string) (*domain.Instance, error)
	ValidateScheduledSupersedeTarget(ctx context.Context, tx db.Tx, tenantID, documentID, supersededDocumentID string) error
	LoadCurrentPublishedHeadForDocument(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error)
	LoadCurrentPublishedHead(ctx context.Context, tx db.Tx, tenantID, controlledDocumentID string) (string, error)
	GetDocumentRevisionVersion(ctx context.Context, tx db.Tx, documentID, tenantID string) (int, error)
	ListRoutes(ctx context.Context, tenantID string) ([]Route, error)
	ListRoutesTx(ctx context.Context, tx db.Tx, tenantID string) ([]Route, error)
	MarkSuperseded(ctx context.Context, tx db.Tx, tenantID, documentID string) error
	UpdateStageStatus(ctx context.Context, tx db.Tx, tenantID, stageID string, newStatus, expectedOldStatus domain.StageStatus) error
	UpdateInstanceStatus(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error

	// Read helpers relocated from application layer (H-5.1).
	// All run inside the caller's transaction so the atomic boundary is preserved.
	LoadPriorSignoffs(ctx context.Context, tx db.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error)
	LoadStageSignoffs(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error)
	HasUnresolvedComments(ctx context.Context, tx db.Tx, tenantID, documentID string) (bool, error)
	// LoadActiveDocumentContentHash returns ErrNoActiveContentHash when the
	// document is missing or has no content hash. The application layer maps
	// this to ErrContentHashMismatch.
	LoadActiveDocumentContentHash(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error)
	ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error)
	LoadRoute(ctx context.Context, tx db.Tx, tenantID, routeID string) (domain.Route, error)
}
