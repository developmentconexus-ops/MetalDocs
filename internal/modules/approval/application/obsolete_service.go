package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/infrastructure"
	docapp "metaldocs/internal/modules/documents/application"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ObsoleteService marks a document as obsolete (end-of-life).
type ObsoleteService struct {
	repo              infrastructure.ApprovalRepository
	emitter           EventEmitter
	clock             Clock
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
}

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer used to publish
// a lifecycle event once the document transitions to obsolete.
func (s *ObsoleteService) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *ObsoleteService {
	s.lifecycleEnqueuer = e
	return s
}

// ErrInvalidObsoleteSource is returned when the document is not in a state
// that permits an → obsolete transition (must be "published" or "superseded").
var ErrInvalidObsoleteSource = errors.New("approval: document must be in 'published' or 'superseded' state to mark obsolete")

// MarkObsoleteRequest carries all inputs for ObsoleteService.MarkObsolete.
type MarkObsoleteRequest struct {
	TenantID        string
	DocumentID      string
	MarkedBy        string // user_id
	RevisionVersion int    // OCC guard
	Reason          string
}

// MarkObsoleteResult is returned on a successful obsolete transition.
type MarkObsoleteResult struct {
	PriorStatus string // the document's status before the transition
}

// MarkObsolete transitions a document from "published" or "superseded" to
// "obsolete" and cancels any in-progress approval instance for that document.
// All writes occur within a single transaction (outbox pattern).
func (s *ObsoleteService) MarkObsolete(ctx context.Context, runner db.TxRunner, req MarkObsoleteRequest) (MarkObsoleteResult, error) {
	var result MarkObsoleteResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// Step 2: fetch current status + revision_version under a row-level lock.
		var priorStatus string
		var currentRevision int
		var areaCode string
		err := tx.QueryRowContext(ctx, `
			SELECT status, revision_version, process_area_code_snapshot
			  FROM documents
			 WHERE id        = $1
			   AND tenant_id = $2
			 FOR UPDATE`,
			req.DocumentID, req.TenantID,
		).Scan(&priorStatus, &currentRevision, &areaCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return fmt.Errorf("markObsolete: load document: %w", err)
		}

		// Step 3: guard — only published or superseded may transition to obsolete.
		// Friendly first-line legality check (M4/F4.1) mirrors the DB trigger and
		// replaces the prior ad-hoc "!= X && != Y" lifecycle guard; the OCC WHERE
		// below remains the atomic CAS + optimistic-lock enforcement.
		if docsdomain.CanTransitionDocumentStatus(docsdomain.DocumentStatus(priorStatus), docsdomain.DocStatusObsolete) != nil {
			return ErrInvalidObsoleteSource
		}

		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentObsolete), areaCode); err != nil {
			return err
		}

		// Step 4: OCC UPDATE — atomically set status and bump revision_version.
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'obsolete',
			       revision_version = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = $3
			   AND revision_version = $4`,
			req.DocumentID, req.TenantID, priorStatus, req.RevisionVersion,
		)
		if err != nil {
			return fmt.Errorf("markObsolete: update document: %w", err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("markObsolete: rows affected: %w", err)
		}
		if affected == 0 {
			return infrastructure.ErrStaleRevision
		}

		// Step 5: cancel any in-progress approval instance (no error if none exist).
		_, err = tx.ExecContext(ctx, `
			UPDATE approval_instances
			   SET status       = 'cancelled',
			       completed_at = now()
			 WHERE document_id = $1
			   AND tenant_id = $2
			   AND status         = 'in_progress'`,
			req.DocumentID, req.TenantID,
		)
		if err != nil {
			return fmt.Errorf("markObsolete: cancel approval instance: %w", err)
		}

		// Step 6: emit "document_obsoleted" governance event.
		payloadMap := map[string]any{
			"reason":       req.Reason,
			"prior_status": priorStatus,
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("markObsolete: marshal event payload: %w", err)
		}
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    "document_obsoleted",
			ActorUserID:  req.MarkedBy,
			ResourceType: "document",
			ResourceID:   req.DocumentID,
			Reason:       req.Reason,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   s.clock.Now(),
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("markObsolete: emit event: %w", err)
		}

		// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Reader event.
		if s.lifecycleEnqueuer != nil {
			cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, req.DocumentID)
			if err != nil {
				return fmt.Errorf("markObsolete: load cd id for lifecycle event: %w", err)
			}
			obsoletedAt := s.clock.Now()
			largs := docsdomain.LifecycleEventArgs{
				EventID:              uuid.NewString(),
				TenantID:             req.TenantID,
				EventType:            docsdomain.EventTypeDocumentObsoleted,
				ResourceType:         "document",
				ResourceID:           req.DocumentID,
				ControlledDocumentID: cdID,
				OccurredAt:           obsoletedAt,
			}
			if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
				return fmt.Errorf("markObsolete: enqueue lifecycle event: %w", err)
			}
		}

		result = MarkObsoleteResult{PriorStatus: priorStatus}
		return nil
	})
	if err != nil {
		return MarkObsoleteResult{}, err
	}
	return result, nil
}
