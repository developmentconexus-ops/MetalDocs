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
	"metaldocs/internal/modules/iam/authz"
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

func (s *SchedulerService) RunScheduledPublishJob(ctx context.Context, db *sql.DB, input ScheduledPublishJobInput) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("scheduler: begin publish tx for doc %s: %w", input.DocumentID, err)
	}
	defer tx.Rollback()
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

	if _, err := s.publishScheduledDocumentTx(ctx, tx, scheduledPublishState{
		DocumentID:           state.DocumentID,
		TenantID:             state.TenantID,
		ControlledDocumentID: state.ControlledDocumentID,
		SupersededDocumentID: state.SupersededDocumentID,
		EffectiveFrom:        state.EffectiveFrom.Time,
		RevisionVersion:      state.RevisionVersion,
		ScheduleGeneration:   state.ScheduleGeneration,
	}); err != nil {
		return err
	}

	return nil
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

func (s *SchedulerService) publishScheduledDocumentTx(ctx context.Context, tx *sql.Tx, row scheduledPublishState) (bool, error) {
	if row.SupersededDocumentID.Valid {
		currentPublishedID, err := s.repo.LoadCurrentPublishedHead(ctx, tx, row.TenantID, row.ControlledDocumentID)
		if err != nil {
			_ = tx.Rollback()
			return false, fmt.Errorf("scheduler: load current published head for doc %s: %w", row.DocumentID, err)
		}
		if currentPublishedID != row.SupersededDocumentID.String {
			_ = tx.Rollback()
			return false, repository.ErrScheduledSupersedeConflict
		}
		if err := s.repo.MarkSuperseded(ctx, tx, row.TenantID, row.SupersededDocumentID.String); err != nil {
			_ = tx.Rollback()
			return false, fmt.Errorf("scheduler: supersede prior head for doc %s: %w", row.DocumentID, err)
		}
	}

	res, err := tx.ExecContext(ctx, updateScheduledDocSQL,
		row.DocumentID, row.TenantID, row.RevisionVersion,
	)
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("scheduler: update document %s: %w", row.DocumentID, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("scheduler: rows affected for doc %s: %w", row.DocumentID, err)
	}

	if affected == 0 {
		// Another runner already won the scheduled row. Roll back so any
		// supersede work done earlier in this transaction is not persisted.
		_ = tx.Rollback()
		return false, nil
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
		return false, fmt.Errorf("scheduler: marshal event payload for doc %s: %w", row.DocumentID, err)
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
		_ = tx.Rollback()
		return false, fmt.Errorf("scheduler: emit event for doc %s: %w", row.DocumentID, err)
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("scheduler: commit publish tx for doc %s: %w", row.DocumentID, err)
	}

	return true, nil
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
	if state.Status != "scheduled" {
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
