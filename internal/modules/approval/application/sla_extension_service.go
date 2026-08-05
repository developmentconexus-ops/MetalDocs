package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ErrSLAExtensionNotForward is returned when the requested due_at is not
// strictly after the stage's current due_at. Shortening a deadline is a
// different act with different governance consequences and is out of scope
// for this route — it stays a rejection, not a silent no-op.
var ErrSLAExtensionNotForward = errors.New("approval: sla extension must move the deadline forward")

// ErrSLAExtensionNoActiveStage is returned when the instance has no active
// stage, or the active stage has no due_at configured — either way there is
// no deadline to extend. A stage with no SLA is not silently given one here;
// that would invent a deadline the route never configured (no-fallback
// principle).
var ErrSLAExtensionNoActiveStage = errors.New("approval: instance has no active stage with a deadline")

// ErrSLAExtensionReasonRequired is returned for a blank reason. Extending a
// deadline in a regulated system is a governance event, and a blank reason is
// no reason.
var ErrSLAExtensionReasonRequired = errors.New("approval: sla extension requires a reason")

// SLAExtensionService postpones the deadline (due_at) of an approval
// instance's active stage.
//
// The route's due_in_days is the STANDARD; this is the recorded EXCEPTION. It
// exists so that giving one subject longer does not require creating a second
// route — which would multiply route rows to express a one-off.
type SLAExtensionService struct {
	emitter EventEmitter
	clock   Clock
}

// NewSLAExtensionService constructs an SLAExtensionService.
func NewSLAExtensionService(emitter EventEmitter, clock Clock) *SLAExtensionService {
	return &SLAExtensionService{emitter: emitter, clock: clock}
}

// ExtendRequest is the use-case input for Extend.
type ExtendRequest struct {
	TenantID   string
	InstanceID string
	ActorID    string
	NewDueAt   time.Time
	Reason     string
}

// Extend moves the active stage's due_at forward.
//
// The read and the write happen in ONE transaction against the same row,
// which is what makes the forward-only rule correct under concurrency — no
// If-Match is needed, and adding one would be contract surface solving a
// problem the transaction already solves.
//
// sla_surfaced_at is deliberately NOT touched. The SLA surfacer's eligibility
// predicate is (sla_surfaced_at IS NULL OR sla_surfaced_at < due_at)
// (infrastructure/sla_overdue_reader.go), so moving due_at forward re-arms
// the reminder by construction: the stale sla_surfaced_at timestamp is now
// less than the new, later due_at. Clearing the marker instead would only
// erase the evidence that the first reminder was actually sent.
func (s *SLAExtensionService) Extend(ctx context.Context, runner db.TxRunner, req ExtendRequest) error {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return ErrSLAExtensionReasonRequired
	}

	return runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		var stageID string
		var currentDue sql.NullTime
		var areaCode sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT asi.id, asi.due_at, asi.area_code_snapshot
			  FROM approval_stage_instances asi
			  JOIN approval_instances ai
			    ON ai.id = asi.approval_instance_id
			 WHERE ai.id = $1
			   AND ai.tenant_id = $2
			   AND asi.status = 'active'
			 FOR UPDATE OF asi`,
			req.InstanceID, req.TenantID,
		).Scan(&stageID, &currentDue, &areaCode)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSLAExtensionNoActiveStage
		}
		if err != nil {
			return fmt.Errorf("extend sla: load active stage: %w", err)
		}

		// Area-grade tier-2 check, in-tx, on the stage's own area snapshot.
		// areaCode.String is "" when area_code_snapshot IS NULL — "" is
		// intentionally fail-closed: authz.Require denies non-system actors
		// for an area-grade cap unless areaCode is the literal "tenant" (ADR
		// 0022 Phase 8), so an instance with no area matches no area row and
		// is DENIED, not judged at tenant grade. Do NOT
		// COALESCE(areaCode.String, "tenant") here — that would silently
		// re-open the area filter, matching CancelService's identical guard
		// (cancel_service.go).
		if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalSLAExtend), areaCode.String); err != nil {
			return err
		}

		// A stage with no deadline configured has nothing to extend — setting
		// one here would invent a deadline the route never configured
		// (no-fallback principle).
		if !currentDue.Valid {
			return ErrSLAExtensionNoActiveStage
		}
		if !req.NewDueAt.After(currentDue.Time) {
			return ErrSLAExtensionNotForward
		}

		// No tenant_id predicate here: approval_stage_instances carries no
		// tenant_id column of its own (tenancy is inherited through its parent
		// approval_instances row, matching every other write in this package —
		// see UpdateStageStatus). The row is already tenant-scoped and locked by
		// the SELECT ... FOR UPDATE OF asi above, in the same transaction.
		if _, err := tx.ExecContext(ctx,
			`UPDATE approval_stage_instances SET due_at = $1 WHERE id = $2`,
			req.NewDueAt.UTC(), stageID,
		); err != nil {
			return fmt.Errorf("extend sla: update due_at: %w", err)
		}

		// The audit event IS the extension history. No parallel history table
		// is created — that would be exactly the redundancy the design
		// forbids.
		payload, err := json.Marshal(map[string]any{
			"instance_id":       req.InstanceID,
			"stage_instance_id": stageID,
			"previous_due_at":   currentDue.Time.UTC(),
			"new_due_at":        req.NewDueAt.UTC(),
			"reason":            reason,
		})
		if err != nil {
			return fmt.Errorf("extend sla: marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeApprovalSLAExtended,
			ActorUserID:  req.ActorID,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			Reason:       reason,
			PayloadJSON:  json.RawMessage(payload),
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("extend sla: emit event: %w", err)
		}
		return nil
	})
}
