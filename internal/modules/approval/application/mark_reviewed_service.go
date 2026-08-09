package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ErrDocumentNotFound is returned when the target document does not exist for
// the tenant (or is invisible under RLS) — mapped to 404, never a silent no-op.
var ErrDocumentNotFound = errors.New("approval: document not found")

// ErrDocumentNotPublished is returned when mark-reviewed targets a document
// whose status is not "published". Mark-reviewed is a review-date update, not
// a status transition (M6 F6.2 spec interview #4) — a non-published document
// is rejected as a friendly first-line precondition failure, not a silent
// no-op mistaken for success.
var ErrDocumentNotPublished = errors.New("approval: mark-reviewed requires the document to be currently published")

// ErrMarkReviewedStaleRevision is returned when the OCC CAS affects 0 rows
// because the caller's ExpectedRevisionVersion no longer matches the current
// row (lost-update guard, mirrors repository.ErrStaleRevision for siblings).
var ErrMarkReviewedStaleRevision = errors.New("approval: mark-reviewed stale revision (If-Match mismatch)")

// ErrReviewDueBeforeEffective is returned when the caller-supplied
// review_due_at would precede the document's effective_from — the friendly
// first-line mirror of ck_documents_review_due_sane (0274).
var ErrReviewDueBeforeEffective = errors.New("approval: review_due_at must not be before effective_from")

// ErrEffectiveToNotAfterEffectiveFrom is returned when the caller-supplied
// effective_to would not be strictly after effective_from — the friendly
// first-line mirror of ck_documents_effective_window (0274).
var ErrEffectiveToNotAfterEffectiveFrom = errors.New("approval: effective_to must be strictly after effective_from")

// MarkReviewedRequest carries all inputs for MarkReviewedService.MarkReviewed.
type MarkReviewedRequest struct {
	TenantID                string
	DocumentID              string
	ReviewDueAt             time.Time
	EffectiveTo             *time.Time // optional; nil = leave effective_to unchanged
	ExpectedRevisionVersion int        // OCC precondition; the concrete revision the caller expects
	ReviewedBy              string     // actor user_id triggering the review
}

// MarkReviewedResult is returned on a successful mark-reviewed.
type MarkReviewedResult struct {
	DocumentID      string
	NewStatus       string // always "published" — mark-reviewed never changes status
	RevisionVersion int    // the new revision_version, for the caller's ETag
}

// MarkReviewedService implements the F6.2 mark-reviewed workflow: records a
// periodic-review completion on a live published revision by setting
// last_reviewed_at + the next review_due_at (and optionally effective_to).
//
// This is NOT a status transition — the document stays "published"; there is
// no 11th status and no published->published call through
// docsdomain.CanTransitionDocumentStatus (that fn has no published->published
// branch by design, per the M6 spec interview #4). The only M4-adjacent
// concern here is a precondition: the document must currently be published,
// asserted directly rather than routed through the transition fn.
type MarkReviewedService struct {
	emitter EventEmitter
	clock   Clock
}

// NewMarkReviewedService constructs a MarkReviewedService.
func NewMarkReviewedService(emitter EventEmitter, clock Clock) *MarkReviewedService {
	return &MarkReviewedService{emitter: emitter, clock: clock}
}

// MarkReviewed runs the mark-reviewed workflow in a single transaction:
// authz.Require(document.review, "tenant") gates the write (document.review
// is ScopeTenant per ADR 0069 — the "tenant" sentinel, never a resolved area
// code); a friendly first-line precondition confirms the document is
// currently published; the OCC CAS UPDATE sets
// last_reviewed_at/review_due_at/effective_to and bumps revision_version so
// the ETag advances (mirrors SchedulePublish, the sibling metadata-bearing
// write on this same table, which also bumps revision_version); a
// "document_reviewed" governance event is emitted in the same tx (outbox
// invariant). Callers must seed authz identity (authz.SeedTxIdentity) before
// runner.Do, matching the composition root's per-request seed used by the
// sibling approval services.
func (s *MarkReviewedService) MarkReviewed(ctx context.Context, runner db.TxRunner, req MarkReviewedRequest) (MarkReviewedResult, error) {
	if req.ReviewDueAt.IsZero() {
		return MarkReviewedResult{}, errors.New("approval: mark-reviewed requires review_due_at")
	}

	var result MarkReviewedResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		res, err := s.markReviewedTx(ctx, tx, req)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return MarkReviewedResult{}, err
	}
	return result, nil
}

// markReviewedTx is the tx-scoped core of MarkReviewed: authz gate,
// precondition checks against the current row, the OCC CAS update, and the
// governance-event emission, all under the same tx/identity.
func (s *MarkReviewedService) markReviewedTx(ctx context.Context, tx *sql.Tx, req MarkReviewedRequest) (MarkReviewedResult, error) {
	ctx = authz.WithCapCache(ctx)

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentReview), "tenant"); err != nil {
		return MarkReviewedResult{}, err
	}

	// Load the current row (existence + status + effective_from) under the
	// same tx/identity as the authz check, before any write.
	status, effectiveFrom, currentRevisionVersion, err := s.loadDocumentForReview(ctx, tx, req)
	if err != nil {
		return MarkReviewedResult{}, err
	}

	// Friendly first-line precondition: mark-reviewed is not a status
	// transition (no 11th state; docsdomain.CanTransitionDocumentStatus is
	// intentionally NOT called here — it has no published->published
	// branch). Just assert the current status is published.
	if status != string(docsdomain.DocStatusPublished) {
		return MarkReviewedResult{}, ErrDocumentNotPublished
	}

	// Friendly first-line mirrors of the 0274 DB CHECKs
	// (ck_documents_review_due_sane, ck_documents_effective_window) — the
	// DB remains the enforced last line.
	if effectiveFrom.Valid {
		if req.ReviewDueAt.Before(effectiveFrom.Time) {
			return MarkReviewedResult{}, ErrReviewDueBeforeEffective
		}
		if req.EffectiveTo != nil && !req.EffectiveTo.After(effectiveFrom.Time) {
			return MarkReviewedResult{}, ErrEffectiveToNotAfterEffectiveFrom
		}
	}

	// OCC precondition is used verbatim. There is no "<=0 means skip" escape
	// hatch: the If-Match wildcard no longer parses (see parseIfMatchMin),
	// and substituting the just-read current version would turn the CAS into
	// a no-op guard — the caller would be told the precondition held against
	// a revision it never saw.
	now := s.clock.Now()
	if err := s.applyMarkReviewedUpdate(ctx, tx, req, now); err != nil {
		return MarkReviewedResult{}, err
	}

	if err := s.emitDocumentReviewedEvent(ctx, tx, req, now); err != nil {
		return MarkReviewedResult{}, err
	}

	return MarkReviewedResult{
		DocumentID:      req.DocumentID,
		NewStatus:       string(docsdomain.DocStatusPublished),
		RevisionVersion: currentRevisionVersion + 1,
	}, nil
}

// loadDocumentForReview reads the document row's status/effective_from/
// revision_version, mapping a missing row to ErrDocumentNotFound.
func (s *MarkReviewedService) loadDocumentForReview(ctx context.Context, tx *sql.Tx, req MarkReviewedRequest) (status string, effectiveFrom sql.NullTime, revisionVersion int, err error) {
	err = tx.QueryRowContext(ctx, `
		SELECT status, effective_from, revision_version
		  FROM documents
		 WHERE id = $1 AND tenant_id = $2`,
		req.DocumentID, req.TenantID,
	).Scan(&status, &effectiveFrom, &revisionVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.NullTime{}, 0, ErrDocumentNotFound
	}
	if err != nil {
		return "", sql.NullTime{}, 0, fmt.Errorf("markReviewed: load document: %w", err)
	}
	return status, effectiveFrom, revisionVersion, nil
}

// applyMarkReviewedUpdate runs the OCC CAS UPDATE recording the review.
func (s *MarkReviewedService) applyMarkReviewedUpdate(ctx context.Context, tx *sql.Tx, req MarkReviewedRequest, now time.Time) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET last_reviewed_at = $1,
		       review_due_at    = $2,
		       effective_to     = COALESCE($3, effective_to),
		       revision_version = revision_version + 1
		 WHERE id               = $4
		   AND tenant_id        = $5
		   AND status           = 'published'
		   AND revision_version = $6`,
		now, req.ReviewDueAt.UTC(), nullableTime(req.EffectiveTo), req.DocumentID, req.TenantID, req.ExpectedRevisionVersion,
	)
	if err != nil {
		return fmt.Errorf("markReviewed: update document: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("markReviewed: rows affected: %w", err)
	}
	if affected == 0 {
		return ErrMarkReviewedStaleRevision
	}
	return nil
}

// emitDocumentReviewedEvent emits the "document_reviewed" governance event
// in the same tx as the update (outbox invariant). Nil emitter is a
// test-only no-op.
func (s *MarkReviewedService) emitDocumentReviewedEvent(ctx context.Context, tx *sql.Tx, req MarkReviewedRequest, now time.Time) error {
	payloadMap := map[string]any{
		"review_due_at": req.ReviewDueAt.UTC().Format(time.RFC3339),
	}
	if req.EffectiveTo != nil {
		payloadMap["effective_to"] = req.EffectiveTo.UTC().Format(time.RFC3339)
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return fmt.Errorf("markReviewed: marshal event payload: %w", err)
	}
	event := GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    EventTypeDocumentReviewed,
		ActorUserID:  req.ReviewedBy,
		ResourceType: "document",
		ResourceID:   req.DocumentID,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}
	if s.emitter == nil {
		return nil
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		return fmt.Errorf("markReviewed: emit event: %w", err)
	}
	return nil
}

// nullableTime maps a nil *time.Time to a driver NULL bind so COALESCE($3,
// effective_to) leaves the column unchanged when the caller did not supply
// effective_to.
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC()
}
