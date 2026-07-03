package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"metaldocs/internal/modules/documents/approval/repository"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// SchedulerService processes River-delivered scheduled publish jobs.
type SchedulerService struct {
	repo    repository.ApprovalRepository
	emitter EventEmitter
	clock   Clock
}

const updateScheduledDocSQL = `
UPDATE documents
   SET status = 'published',
       effective_from = NULL,
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
	EffectiveFrom        sql.NullTime
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
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("scheduler: bypass authz for doc %s: %w", input.DocumentID, err)
		}

		state, err := s.loadScheduledDocumentState(ctx, tx, input.TenantID, input.DocumentID)
		if errors.Is(err, sql.ErrNoRows) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "document_not_found")
			return nil
		}
		if err != nil {
			return fmt.Errorf("scheduler: load scheduled state for doc %s: %w", input.DocumentID, err)
		}
		if !scheduledJobMatchesState(state, input) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "stale_job")
			return nil
		}
		if s.clock.Now().UTC().Before(state.EffectiveFrom.Time.UTC()) {
			slog.InfoContext(ctx, "scheduler skipping publish job", "document_id", input.DocumentID, "reason", "pre_effective_date")
			return nil
		}

		return s.publishScheduledDocumentTx(ctx, tx, scheduledPublishState{
			DocumentID:           state.DocumentID,
			TenantID:             state.TenantID,
			ControlledDocumentID: state.ControlledDocumentID,
			SupersededDocumentID: state.SupersededDocumentID,
			EffectiveFrom:        state.EffectiveFrom.Time,
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
	EffectiveFrom        time.Time
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
		currentPublishedID, err := s.repo.LoadCurrentPublishedHead(ctx, tx, row.TenantID, row.ControlledDocumentID)
		if err != nil {
			return fmt.Errorf("scheduler: load current published head for doc %s: %w", row.DocumentID, err)
		}
		if currentPublishedID != row.SupersededDocumentID.String {
			return repository.ErrScheduledSupersedeConflict
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
		"effective_date":   row.EffectiveFrom.Format("2006-01-02T15:04:05Z"),
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

func (s *SchedulerService) loadScheduledDocumentState(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (scheduledDocumentState, error) {
	var state scheduledDocumentState
	err := tx.QueryRowContext(ctx, `
		SELECT id, tenant_id, status, controlled_document_id, superseded_document_id,
		       effective_from, revision_version, schedule_generation
		  FROM documents
		 WHERE tenant_id = $1
		   AND id = $2
		 FOR UPDATE`,
		tenantID, documentID,
	).Scan(
		&state.DocumentID,
		&state.TenantID,
		&state.Status,
		&state.ControlledDocumentID,
		&state.SupersededDocumentID,
		&state.EffectiveFrom,
		&state.RevisionVersion,
		&state.ScheduleGeneration,
	)
	return state, err
}

func scheduledJobMatchesState(state scheduledDocumentState, input ScheduledPublishJobInput) bool {
	if state.Status != string(docsdomain.DocStatusScheduled) {
		return false
	}
	if !state.EffectiveFrom.Valid {
		return false
	}
	if state.ScheduleGeneration != input.ScheduleGeneration {
		return false
	}
	if state.RevisionVersion != input.ExpectedRevisionVersion {
		return false
	}
	return state.EffectiveFrom.Time.UTC().Equal(input.ScheduledEffectiveAt.UTC())
}
