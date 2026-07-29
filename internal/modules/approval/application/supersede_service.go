package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// SupersedeService marks a published document as superseded by a newer revision.
type SupersedeService struct {
	repo              infrastructure.ApprovalRepository
	emitter           EventEmitter
	clock             Clock
	cdRead            controlleddocumentsdomain.CDFieldReader
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
}

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer used to publish
// a lifecycle event once the document is superseded.
func (s *SupersedeService) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *SupersedeService {
	s.lifecycleEnqueuer = e
	return s
}

// SupersedeRequest carries all inputs for PublishSuperseding.
type SupersedeRequest struct {
	TenantID           string
	NewDocumentID      string // the document being published (becomes "published")
	PriorDocumentID    string // the previous published document (becomes "superseded")
	SupersededBy       string // user_id
	NewRevisionVersion int    // OCC for new doc
	// PriorRevisionVersion is IGNORED by PublishSuperseding: the prior
	// document's OCC value is re-read from the row under the lock (see below),
	// because any caller-supplied value is pre-lock and therefore possibly
	// stale. No production caller populates it (the HTTP handler leaves it zero
	// — internal/modules/approval/http/supersede_handler.go:47); it is on the
	// ADR 0085 stage B retirement inventory and must never be reintroduced as a
	// CAS input.
	PriorRevisionVersion int
}

// SupersedeResult is returned on successful publish-and-supersede.
type SupersedeResult struct {
	NewDocumentStatus   string // "published"
	PriorDocumentStatus string // "superseded"
}

// PublishSuperseding atomically transitions a new document from "approved" to
// "published" and the prior document from "published" to "superseded".
// Both OCC guards must pass; otherwise the transaction is rolled back and
// infrastructure.ErrStaleRevision is returned.
func (s *SupersedeService) PublishSuperseding(ctx context.Context, runner db.TxRunner, req SupersedeRequest) (SupersedeResult, error) {
	var result SupersedeResult
	// Fail closed: the prior document's OCC value is re-read from the row under
	// the lock (step 2b below) and there is no honest substitute for it — the
	// request field is pre-lock and possibly stale. Without a repository the
	// CAS predicate cannot be established, so refuse rather than fall back
	// (no-fallback principle).
	if s.repo == nil {
		return SupersedeResult{}, fmt.Errorf("publishSuperseding: repository not wired")
	}
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// document.supersede is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, req.NewDocumentID)
		if err != nil {
			return fmt.Errorf("publishSuperseding: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSupersede), areaCode); err != nil {
			return err
		}
		// Supersede mutates public.documents twice in the same transaction, so it
		// must also satisfy the shared documents tripwire.
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}

		// Step 2: take the row locks. ADR 0085 global lock-order doctrine —
		// documents rows are locked individually, in sorted id order, by the one
		// shared helper (release_coordinator.go), so this path can never deadlock
		// against the coordinator or SchedulePublish. Locks come AFTER the
		// authz.Require calls above: H-PRE-1 forbids an authz-recording read
		// inside a lock-holding transaction.
		if err := lockDocumentRowsInIDOrder(ctx, tx, req.TenantID, req.PriorDocumentID, req.NewDocumentID); err != nil {
			return fmt.Errorf("publishSuperseding: %w", err)
		}

		// Step 2b: re-read the prior row's revision_version AFTER the lock so the
		// CAS predicate below uses the under-lock value. This is unconditional —
		// req.PriorRevisionVersion is never used as a CAS input, because a
		// pre-lock value would defeat the very guarantee the lock provides.
		priorRevisionVersion, err := s.repo.GetDocumentRevisionVersion(ctx, tx, req.PriorDocumentID, req.TenantID)
		if err != nil {
			return fmt.Errorf("publishSuperseding: load prior revision version: %w", err)
		}

		// Step 3: OCC UPDATE for the PRIOR document (published → superseded).
		// This runs BEFORE the new document is published: ux_documents_published_head
		// (migration 0310) allows at most one published row per
		// (tenant_id, controlled_document_id), so demoting the incumbent head first
		// is what keeps the invariant from being transiently violated inside the
		// transaction for a same-controlled-document supersession. Friendly
		// first-line legality check (M4/F4.1) mirrors the DB trigger; the OCC
		// WHERE below remains the atomic CAS + optimistic-lock enforcement.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusPublished, docsdomain.DocStatusSuperseded); err != nil {
			return err
		}
		priorResult, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'superseded',
			       revision_version = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'published'
			   AND revision_version = $3`,
			req.PriorDocumentID, req.TenantID, priorRevisionVersion,
		)
		if err != nil {
			return fmt.Errorf("publishSuperseding: update prior document: %w",
				infrastructure.MapPgError(err, infrastructure.MapHints{}))
		}
		priorAffected, err := priorResult.RowsAffected()
		if err != nil {
			return fmt.Errorf("publishSuperseding: rows affected (prior): %w", err)
		}
		if priorAffected == 0 {
			return infrastructure.ErrStaleRevision
		}

		// Step 4: OCC UPDATE for the NEW document (approved → published), now that
		// the head slot is free. Friendly first-line legality check (M4/F4.1)
		// mirrors the DB trigger; the OCC WHERE below remains the atomic CAS +
		// optimistic-lock enforcement.
		//
		// ADR 0085 / F-QA4-13: a published document MUST carry an actual
		// effective_from. This legacy supersede-publish path is retired in ADR
		// 0085 stage B (retirement inventory); until then it stamps the actual
		// instant so ck_documents_published_effective_from (migration 0310) holds.
		// A residual 23505 on ux_documents_published_head (a third document took
		// the head) is mapped to ErrStaleRevision — a 409, not a naked 500.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusApproved, docsdomain.DocStatusPublished); err != nil {
			return err
		}
		newResult, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'published',
			       effective_from   = COALESCE(effective_from, now()),
			       revision_version = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'approved'
			   AND revision_version = $3`,
			req.NewDocumentID, req.TenantID, req.NewRevisionVersion,
		)
		if err != nil {
			return fmt.Errorf("publishSuperseding: update new document: %w",
				infrastructure.MapPgError(err, infrastructure.MapHints{}))
		}
		newAffected, err := newResult.RowsAffected()
		if err != nil {
			return fmt.Errorf("publishSuperseding: rows affected (new): %w", err)
		}
		if newAffected == 0 {
			return infrastructure.ErrStaleRevision
		}

		// Step 5: emit "document_superseded" governance event.
		now := s.clock.Now()
		payloadMap := map[string]any{
			"new_document_id":   req.NewDocumentID,
			"prior_document_id": req.PriorDocumentID,
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("publishSuperseding: marshal event payload: %w", err)
		}
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeDocumentSuperseded,
			ActorUserID:  req.SupersededBy,
			ResourceType: "document",
			ResourceID:   req.NewDocumentID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("publishSuperseding: emit event: %w", err)
		}

		// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Reader event for new doc's CD.
		if s.lifecycleEnqueuer != nil {
			cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, req.NewDocumentID)
			if err != nil {
				return fmt.Errorf("publishSuperseding: load cd id for lifecycle event: %w", err)
			}
			largs := docsdomain.LifecycleEventArgs{
				EventID:              uuid.NewString(),
				TenantID:             req.TenantID,
				EventType:            docsdomain.EventTypeDocumentSuperseded,
				ResourceType:         "document",
				ResourceID:           req.NewDocumentID,
				ControlledDocumentID: cdID,
				OccurredAt:           now,
			}
			if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
				return fmt.Errorf("publishSuperseding: enqueue lifecycle event: %w", err)
			}
		}

		result = SupersedeResult{
			NewDocumentStatus:   string(docsdomain.DocStatusPublished),
			PriorDocumentStatus: string(docsdomain.DocStatusSuperseded),
		}
		return nil
	})
	if err != nil {
		return SupersedeResult{}, err
	}
	return result, nil
}
