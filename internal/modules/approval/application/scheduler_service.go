package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// SchedulerService processes River-delivered scheduled publish jobs.
type SchedulerService struct {
	repo    infrastructure.ApprovalRepository
	emitter EventEmitter
	clock   Clock
}

// updateScheduledDocSQL publishes a scheduled document.
//
// ADR 0085 / F-QA4-13: effective_from is the ACTUAL release instant, stamped
// here — it used to be cleared to NULL, which silently excluded every
// scheduled-then-published document from the periodic-review surfacer. The
// PLAN (planned_effective_from) is preserved for attribution.
const updateScheduledDocSQL = `
UPDATE documents
   SET status = 'published',
       effective_from = now(),
       revision_version = revision_version + 1
 WHERE id = $1
   AND tenant_id = $2
   AND status = 'scheduled'
   AND revision_version = $3`

type scheduledDocumentState struct {
	DocumentID           string
	TenantID             string
	Status               string
	ControlledDocumentID string
	SupersededDocumentID sql.NullString
	PlannedEffectiveFrom sql.NullTime
	RevisionVersion      int
	ScheduleGeneration   int64
}

// errScheduledPublishNoOp signals that the scheduled documents row was already
// claimed by another runner (the OCC UPDATE affected zero rows). The runner
// rolls the transaction back so any partial supersede work is discarded, and
// RunScheduledPublishJob maps it to a successful no-op.
var errScheduledPublishNoOp = errors.New("scheduler: scheduled row already published")

// RunScheduledPublishJob performs the deferred publish (and supersede, when
// applicable) for one scheduled-publish job. It is idempotent: a stale or
// already-claimed job (OCC UPDATE affects zero rows) is treated as a
// successful no-op rather than an error, so River-triggered retries are safe.
func (s *SchedulerService) RunScheduledPublishJob(ctx context.Context, runner db.TxRunner, input ScheduledPublishJobInput) error {
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		// M3 F3.2 (validation-contract.md §2.2 site 3) — seed the tenant-only
		// RLS backstop GUC before the FOR UPDATE lock below (H-PRE-1: no
		// authz.Require is added to this locked path; BypassSystem remains
		// the separate write-tripwire gate).
		if err := authz.SeedTxTenant(ctx, tx, input.TenantID); err != nil {
			return fmt.Errorf("scheduler: seed tenant for doc %s: %w", input.DocumentID, err)
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("scheduler: bypass authz for doc %s: %w", input.DocumentID, err)
		}

		// ADR 0085 lock-order parity: this read is LOCK-FREE and exists only to
		// DISCOVER the lock set. The supersede target is knowable only after
		// reading the source row, so locking the source here would put it ahead
		// of a lower-sorting target and reintroduce exactly the AB-BA deadlock
		// the sorted order exists to prevent (release_coordinator.go
		// lockAndRedecide).
		state, err := s.loadScheduledDocumentState(ctx, tx, input.TenantID, input.DocumentID)
		if errors.Is(err, sql.ErrNoRows) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "document_not_found")
			return nil
		}
		if err != nil {
			return fmt.Errorf("scheduler: load scheduled state for doc %s: %w", input.DocumentID, err)
		}

		if err := lockDocumentRowsInIDOrder(ctx, tx, input.TenantID,
			state.DocumentID, state.SupersededDocumentID.String); err != nil {
			return fmt.Errorf("scheduler: %w", err)
		}
		// Values may have moved between the unlocked read and the lock, so every
		// gate below runs on the state read UNDER it. A supersede target that
		// changed in that window also bumped revision_version (every writer of
		// documents.superseded_document_id does), so scheduledJobMatchesState
		// rejects the job as stale rather than acting on a row we did not lock.
		state, err = s.loadScheduledDocumentState(ctx, tx, input.TenantID, input.DocumentID)
		if errors.Is(err, sql.ErrNoRows) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "document_not_found")
			return nil
		}
		if err != nil {
			return fmt.Errorf("scheduler: re-read scheduled state for doc %s: %w", input.DocumentID, err)
		}
		if !scheduledJobMatchesState(state, input) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "stale_job")
			return nil
		}
		if s.clock.Now().UTC().Before(state.PlannedEffectiveFrom.Time.UTC()) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "pre_effective_date")
			return nil
		}

		return s.publishScheduledDocumentTx(ctx, tx, scheduledPublishState{
			DocumentID:           state.DocumentID,
			TenantID:             state.TenantID,
			ControlledDocumentID: state.ControlledDocumentID,
			SupersededDocumentID: state.SupersededDocumentID,
			PlannedEffectiveFrom: state.PlannedEffectiveFrom.Time,
			RevisionVersion:      state.RevisionVersion,
			ScheduleGeneration:   state.ScheduleGeneration,
		})
	})
	if errors.Is(err, errScheduledPublishNoOp) {
		// Another runner already published the scheduled row; the partial
		// supersede work (if any) was rolled back. Treat as a successful no-op.
		return nil
	}
	return err
}

type scheduledPublishState struct {
	DocumentID           string
	TenantID             string
	ControlledDocumentID string
	SupersededDocumentID sql.NullString
	PlannedEffectiveFrom time.Time
	RevisionVersion      int
	ScheduleGeneration   int64
}

// publishScheduledDocumentTx performs the supersede + publish + emit work inside
// the caller's transaction. It does NOT own commit/rollback — the TxRunner does.
// It returns errScheduledPublishNoOp when the OCC UPDATE affects zero rows so the
// runner rolls back any partial supersede work; RunScheduledPublishJob maps that
// sentinel to a successful no-op.
func (s *SchedulerService) publishScheduledDocumentTx(ctx context.Context, tx *sql.Tx, row scheduledPublishState) error {
	if row.SupersededDocumentID.Valid {
		// NoLock: RunScheduledPublishJob already holds this row's lock, taken in
		// sorted id order with the source. The locking sibling would take it
		// here instead — after the source read — which is the out-of-order
		// acquisition ADR 0085's single lock order forbids.
		currentPublishedID, err := s.repo.LoadCurrentPublishedHeadNoLock(ctx, tx, row.TenantID, row.ControlledDocumentID)
		if err != nil {
			return fmt.Errorf("scheduler: load current published head for doc %s: %w", row.DocumentID, err)
		}
		if currentPublishedID != row.SupersededDocumentID.String {
			return infrastructure.ErrScheduledSupersedeConflict
		}
		if err := s.repo.MarkSuperseded(ctx, tx, row.TenantID, row.SupersededDocumentID.String); err != nil {
			return fmt.Errorf("scheduler: supersede prior head for doc %s: %w", row.DocumentID, err)
		}
	}

	res, err := tx.ExecContext(ctx, updateScheduledDocSQL,
		row.DocumentID, row.TenantID, row.RevisionVersion,
	)
	if err != nil {
		return fmt.Errorf("scheduler: update document %s: %w", row.DocumentID, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("scheduler: rows affected for doc %s: %w", row.DocumentID, err)
	}

	if affected == 0 {
		// Another runner already won the scheduled row. Return the no-op sentinel
		// so the runner rolls back any supersede work done earlier in this
		// transaction instead of persisting it.
		return errScheduledPublishNoOp
	}

	payloadMap := map[string]any{
		"revision_version": row.RevisionVersion + 1,
		"effective_date":   row.PlannedEffectiveFrom.Format("2006-01-02T15:04:05Z"),
	}
	if row.SupersededDocumentID.Valid {
		payloadMap["superseded_document_id"] = row.SupersededDocumentID.String
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return fmt.Errorf("scheduler: marshal event payload for doc %s: %w", row.DocumentID, err)
	}

	ev := GovernanceEvent{
		TenantID:     row.TenantID,
		EventType:    EventTypeDocumentPublished,
		ActorUserID:  "scheduler",
		ResourceType: "document",
		ResourceID:   row.DocumentID,
		Reason:       "scheduled publish",
		PayloadJSON:  json.RawMessage(payload),
		OccurredAt:   s.clock.Now(),
	}

	if err = s.emitter.Emit(ctx, tx, ev); err != nil {
		return fmt.Errorf("scheduler: emit event for doc %s: %w", row.DocumentID, err)
	}

	return nil
}

// loadScheduledDocumentState reads the scheduled row WITHOUT locking it. The
// lock on this row is taken by lockDocumentRowsInIDOrder together with the
// supersede target, in sorted id order; this helper is called once before that
// (to discover the target) and once after (to read the authoritative state).
func (s *SchedulerService) loadScheduledDocumentState(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (scheduledDocumentState, error) {
	var state scheduledDocumentState
	err := tx.QueryRowContext(ctx, `
		SELECT id, tenant_id, status, controlled_document_id, superseded_document_id,
		       planned_effective_from, revision_version, schedule_generation
		  FROM documents
		 WHERE tenant_id = $1
		   AND id = $2`,
		tenantID, documentID,
	).Scan(
		&state.DocumentID,
		&state.TenantID,
		&state.Status,
		&state.ControlledDocumentID,
		&state.SupersededDocumentID,
		&state.PlannedEffectiveFrom,
		&state.RevisionVersion,
		&state.ScheduleGeneration,
	)
	return state, err
}

func scheduledJobMatchesState(state scheduledDocumentState, input ScheduledPublishJobInput) bool {
	// Friendly first-line legality check (M4/F4.1) mirrors the DB trigger; the
	// OCC UPDATE in publishScheduledDocumentTx remains the atomic CAS +
	// optimistic-lock enforcement. state.Status is loaded fresh under the row
	// lock, so this is the authoritative current status for this job run.
	if docsdomain.CanTransitionDocumentStatus(docsdomain.DocumentStatus(state.Status), docsdomain.DocStatusPublished) != nil {
		return false
	}
	if !state.PlannedEffectiveFrom.Valid {
		return false
	}
	if state.ScheduleGeneration != input.ScheduleGeneration {
		return false
	}
	if state.RevisionVersion != input.ExpectedRevisionVersion {
		return false
	}
	return state.PlannedEffectiveFrom.Time.UTC().Equal(input.ScheduledEffectiveAt.UTC())
}
