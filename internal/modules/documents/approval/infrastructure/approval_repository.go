// Package infrastructure defines ApprovalRepository, the persistence port for the
// approval subsystem, plus its Postgres implementation and the pgError→domain-error
// mapping (MapPgError). All mutating methods take a caller-owned db.Tx; the
// repository never opens its own transaction or commits.
package infrastructure

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

// VerdictInsertResult returned by InsertVerdict. Mirrors SignoffInsertResult
// (F4): idempotent replay detection via the same stage+actor unique-key shape.
type VerdictInsertResult struct {
	ID        string
	WasReplay bool // true if ON CONFLICT detected an existing matching verdict
}

// ErrScheduledSupersedeConflict is returned when a scheduled-publish cutover's
// recorded supersede target no longer matches the document currently published,
// meaning another write raced ahead of the schedule; the job should no-op.
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
	Kind               string
	DueInDays          *int
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
	// LoadGovernedRevisionNumber reads the document's governed
	// documents.revision_number inside the caller's transaction (tenant-scoped,
	// GUC/RLS-correct). Added for T8b: SubmitRevisionForReview must derive this
	// value from the row itself rather than trust a client-supplied
	// SubmitRequest.RevisionNumber, which defaults to 0 on live traffic and
	// silently defeats the REV>=1 reason_for_change/revision_title gates.
	LoadGovernedRevisionNumber(ctx context.Context, tx db.Tx, tenantID, documentID string) (int, error)
	ListRoutes(ctx context.Context, tenantID string) ([]Route, error)
	ListRoutesTx(ctx context.Context, tx db.Tx, tenantID string) ([]Route, error)
	MarkSuperseded(ctx context.Context, tx db.Tx, tenantID, documentID string) error
	UpdateStageStatus(ctx context.Context, tx db.Tx, tenantID, stageID string, newStatus, expectedOldStatus domain.StageStatus) error
	UpdateInstanceStatus(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error
	// UpdateInstanceStatusWithReason is UpdateInstanceStatus plus a
	// cancel_reason write (F4): used by CancelInstance (reason from the
	// caller) and the request_changes verdict path (reason = the verdict
	// comment) — both transition an in_progress instance away and want the
	// human-readable reason persisted on approval_instances.cancel_reason
	// rather than only reaching the governance event.
	UpdateInstanceStatusWithReason(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time, reason string) error

	// InsertVerdict inserts a review-stage runtime verdict (F4), idempotent-replay
	// aware exactly like InsertSignoff: ON CONFLICT on (stage_instance_id,
	// actor_user_id) — matching fields → WasReplay=true, mismatching → error.
	InsertVerdict(ctx context.Context, tx db.Tx, v domain.ReviewVerdict) (VerdictInsertResult, error)
	// LoadStageVerdicts fetches all verdicts recorded for a single stage
	// instance, used to evaluate quorum for the `ready` path exactly like
	// LoadStageSignoffs.
	LoadStageVerdicts(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.ReviewVerdict, error)

	// Read helpers relocated from application layer (H-5.1).
	// All run inside the caller's transaction so the atomic boundary is preserved.
	LoadPriorSignoffs(ctx context.Context, tx db.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error)
	LoadStageSignoffs(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error)
	HasUnresolvedComments(ctx context.Context, tx db.Tx, tenantID, documentID string) (bool, error)
	// HasUnresolvedInstanceComments is HasUnresolvedComments scoped to comments
	// created at or after `since` (the instance's SubmittedAt) — F5, spec.md
	// §2.2 "Comment-resolution scope": the freeze gate must not block on stale
	// historical comments from a prior revision. HasUnresolvedComments itself
	// stays document-wide and unmodified (kept for now — bounded defer,
	// spec.md F5 Interview #8 — no remaining production call site after F5,
	// but not deleted in this feature).
	HasUnresolvedInstanceComments(ctx context.Context, tx db.Tx, tenantID, documentID string, since time.Time) (bool, error)
	// PinFrozenHash CAS-writes approval_instances.frozen_content_hash the first
	// time an instance reaches the freeze boundary (F5, spec.md §2.2). The
	// write only succeeds when the column is currently NULL: won=true means
	// this call performed the freeze; won=false means the instance was
	// already frozen (idempotent no-op, never an error — freeze idempotency
	// is a first-class outcome). The enclosing stage-transition tx already
	// holds the instance row FOR UPDATE (via LoadInstance), so this CAS is
	// defense-in-depth, not the actual race-resolution mechanism — see F5
	// spec.md's Consumer contract disambiguation note.
	PinFrozenHash(ctx context.Context, tx db.Tx, tenantID, instanceID, hash string) (won bool, err error)
	// LoadFrozenContentHash returns the instance's frozen_content_hash (F1/F5's
	// freeze-boundary pin) by instance id. No-fallback (F6, spec §11): returns
	// ErrNoActiveContentHash when the instance row is missing or its
	// frozen_content_hash is NULL — NEVER substitutes a head document_revisions
	// hash or any other value. By the time any signoff/publish call site reads
	// this, the instance must already be frozen (F5: freeze always precedes
	// activation of an approval-kind stage); a NULL pin at that point indicates
	// an impossible state that must fail closed, not a legitimate "not yet
	// computed" case to paper over. The application layer maps
	// ErrNoActiveContentHash to ErrContentHashMismatch.
	LoadFrozenContentHash(ctx context.Context, tx db.Tx, tenantID, instanceID string) (string, error)
	ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error)
	// LoadActorDisplayName returns metaldocs.iam_users.display_name for (tenantID,
	// userID), or "" when the user row is absent. It runs OFF the caller's
	// transaction (on the pool) so it never executes inside the signoff
	// advisory-lock atomic tx (H-PRE-1). Tenant scope is the explicit tenant_id
	// predicate; the metaldocs.tenant_id RLS GUC is unset on the pool connection,
	// which the NULL-permissive tenant_isolation policy (migration 0237) allows.
	LoadActorDisplayName(ctx context.Context, tenantID, userID string) (string, error)
	LoadRoute(ctx context.Context, tx db.Tx, tenantID, routeID string) (domain.Route, error)
}
