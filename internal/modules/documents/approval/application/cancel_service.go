package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// CancelService cancels an in-progress approval instance and reverts the
// document back to draft status.
type CancelService struct {
	repo    repository.ApprovalRepository
	emitter EventEmitter
	clock   Clock
}

// ErrReasonRequired is returned when CancelInput.Reason is empty.
var ErrReasonRequired = errors.New("cancel: reason must not be empty")

// CancelInput carries all inputs for CancelService.CancelInstance.
type CancelInput struct {
	TenantID                string
	InstanceID              string
	ExpectedRevisionVersion int // OCC guard on the document
	ActorUserID             string
	Reason                  string
}

// CancelResult is returned on a successful cancellation.
type CancelResult struct {
	DocumentID string
}

// CancelInstance cancels an in-progress approval instance, transitions all
// active/pending stages to cancelled, and reverts the document to draft.
// Requires the document.edit capability for the document's area (ADR 0022 P10).
func (s *CancelService) CancelInstance(ctx context.Context, runner db.TxRunner, in CancelInput) (CancelResult, error) {
	return s.cancelInstance(ctx, runner, in, false)
}

// SystemCancelInstance cancels an in-progress approval instance for internal
// scheduler/watchdog callers using system authz bypass.
func (s *CancelService) SystemCancelInstance(ctx context.Context, runner db.TxRunner, in CancelInput) (CancelResult, error) {
	return s.cancelInstance(ctx, runner, in, true)
}

func (s *CancelService) cancelInstance(ctx context.Context, runner db.TxRunner, in CancelInput, system bool) (CancelResult, error) {
	// Guard: reason required.
	if in.Reason == "" {
		return CancelResult{}, ErrReasonRequired
	}

	var docID string
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if system {
			if err := authz.BypassSystem(ctx, tx); err != nil {
				return fmt.Errorf("cancel: system authz bypass: %w", err)
			}
			// System path (stuck-instance-watchdog auto-cancel) runs on a
			// carrier-less background ctx, so the TxRunner chokepoint auto-seed
			// (F3.1) is a no-op here. Seed the claimed instance's tenant
			// explicitly so FORCE RLS still backstops this per-instance action
			// (validation-contract §2.4 stuck-watchdog: "per-instance action …
			// seeded per its tenant"). The sync CancelInstance path is seeded by
			// the chokepoint from its request identity and needs no manual seed.
			if err := authz.SeedTxIdentity(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
				return fmt.Errorf("cancel: seed system identity: %w", err)
			}
		}

		// Load instance to get document ID and verify not already terminal.
		inst, err := s.repo.LoadInstance(ctx, tx, in.TenantID, in.InstanceID)
		if err != nil {
			return fmt.Errorf("cancel: load instance: %w", err)
		}
		if inst == nil {
			return repository.ErrNoActiveInstance
		}
		if inst.Status != domain.InstanceInProgress {
			return repository.ErrInstanceCompleted
		}

		docID = inst.DocumentID

		// Fetch document area_code for authz check. FOR UPDATE locks the document row
		// to prevent concurrent area_code changes between authz decision and status update.
		var areaCode sql.NullString
		if err := tx.QueryRowContext(ctx,
			`SELECT process_area_code_snapshot FROM documents WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
			docID, in.TenantID,
		).Scan(&areaCode); err != nil {
			return fmt.Errorf("cancel: fetch area_code: %w", err)
		}

		// Authz gate: require document.edit capability (area-grade). Cancelling an
		// in-progress workflow reverts the document to draft — an area-scoped edit.
		// ADR 0022 Phase 10 (F2): the redundant workflow.instance.cancel cap was
		// merged into the canonical CapDocumentEdit (identical grant set).
		// areaCode.String is "" when process_area_code_snapshot IS NULL — "" is
		// intentionally fail-closed: authz.Require denies non-system actors for an
		// area-grade cap (ADR 0022 Phase 8, matches docapp.LoadDocumentAreaCode). Do NOT
		// COALESCE(..., 'tenant') here — that would silently re-open the area filter.
		if !system {
			if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode.String); err != nil {
				return err
			}
		}

		// SET LOCAL cancel GUC — authorises under_review→draft transition in trigger.
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.cancel_in_progress', $1, true)`,
			in.InstanceID,
		); err != nil {
			return fmt.Errorf("cancel: set cancel GUC: %w", err)
		}

		// Cancel approval instance.
		now := s.clock.Now()
		if err := s.repo.UpdateInstanceStatus(ctx, tx, in.TenantID, in.InstanceID,
			domain.InstanceCancelled, domain.InstanceInProgress, &now); err != nil {
			return fmt.Errorf("cancel: update instance status: %w", err)
		}

		// Cancel all active and pending stage instances.
		if _, err := tx.ExecContext(ctx, `
			UPDATE approval_stage_instances asi
			   SET status = 'cancelled'
			  FROM approval_instances ai
			 WHERE asi.approval_instance_id = ai.id
			   AND asi.approval_instance_id = $1
			   AND ai.tenant_id = $2
			   AND asi.status IN ('active','pending')`,
			in.InstanceID, in.TenantID,
		); err != nil {
			return fmt.Errorf("cancel: cancel stages: %w", err)
		}

		// Revert document to draft (trigger enforces under_review→draft only with GUC set).
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'draft',
			       revision_version = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'under_review'
			   AND revision_version = $3`,
			docID, in.TenantID, in.ExpectedRevisionVersion,
		)
		if err != nil {
			return fmt.Errorf("cancel: revert doc to draft: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("cancel: rows affected: %w", err)
		}
		if rows == 0 {
			return repository.ErrStaleRevision
		}

		// Emit governance event.
		payload, err := json.Marshal(map[string]any{
			"instance_id": in.InstanceID,
			"reason":      in.Reason,
		})
		if err != nil {
			return fmt.Errorf("cancel: marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeApprovalInstanceCancelled,
			ActorUserID:  in.ActorUserID,
			ResourceType: "document",
			ResourceID:   docID,
			Reason:       in.Reason,
			PayloadJSON:  payload,
			OccurredAt:   now,
		}); err != nil {
			return fmt.Errorf("cancel: emit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return CancelResult{}, err
	}
	return CancelResult{DocumentID: docID}, nil
}

// newCancelService constructs a CancelService.
func newCancelService(repo repository.ApprovalRepository, emitter EventEmitter, clock Clock) *CancelService {
	return &CancelService{repo: repo, emitter: emitter, clock: clock}
}
