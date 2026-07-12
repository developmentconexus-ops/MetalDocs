package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// PublishService handles transitioning an approved document to published state.
type PublishService struct {
	repo                     infrastructure.ApprovalRepository
	emitter                  EventEmitter
	clock                    Clock
	scheduledPublishEnqueuer ScheduledPublishEnqueuer
	cdRead                   controlleddocumentsdomain.CDFieldReader
	lifecycleEnqueuer        docsdomain.LifecycleEventEnqueuer
}

// ErrInstanceNotApproved is returned when PublishApproved is called on an
// instance whose status is not "approved".
var ErrInstanceNotApproved = errors.New("approval: instance is not in approved state")

// PublishRequest carries all inputs for PublishApproved.
type PublishRequest struct {
	TenantID                string
	InstanceID              string
	PublishedBy             string // user_id triggering publish
	ExpectedRevisionVersion int
}

// PublishResult is returned on successful publish.
type PublishResult struct {
	DocumentID string
	NewStatus  string // "published"
}

// PublishApproved transitions an approved document to published state.
// It verifies the approval instance is in "approved" status, performs an OCC
// UPDATE on the documents table (approved → published), emits a
// "document_published" governance event, and commits.
func (s *PublishService) PublishApproved(ctx context.Context, runner db.TxRunner, req PublishRequest) (PublishResult, error) {
	var result PublishResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// Step 2: load the approval instance.
		instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return fmt.Errorf("publishApproved: load instance: %w", err)
		}
		if instance == nil {
			return infrastructure.ErrNoActiveInstance
		}
		if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
			return infrastructure.ErrStaleRevision
		}

		// Verify instance is in approved state.
		if instance.Status != domain.InstanceApproved {
			return ErrInstanceNotApproved
		}

		// No-fallback (F6, spec §11): publish is the final link in the canonical
		// hash chain (pin at freeze → echo at signoff → verify at publish). An
		// approved instance must already carry a frozen_content_hash pinned at
		// the freeze boundary (F5) — this is structurally guaranteed by F4/F5's
		// stage-kind gating, but the whole point of the no-fallback principle is
		// to make an impossible state loudly fail rather than silently proceed.
		// This is an additive fail-closed guard, not a client-echo comparison —
		// publish has no client-supplied hash to verify against.
		if instance.FrozenContentHash == nil {
			return ErrContentHashMismatch
		}

		// document.publish is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.DocumentID)
		if err != nil {
			return fmt.Errorf("publishApproved: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentPublish), areaCode); err != nil {
			return err
		}
		// public.documents is protected by the shared documents tripwire, so the
		// approval publish transaction must assert document.edit before updating it.
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}

		// Step 3: transition the document from "approved" to "published". Friendly
		// first-line legality check (M4/F4.1) mirrors the DB trigger; the OCC
		// WHERE below remains the atomic CAS + optimistic-lock enforcement.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusApproved, docsdomain.DocStatusPublished); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'published',
			       revision_version = revision_version + 1
			 WHERE id        = $1
			   AND tenant_id = $2
			   AND status    = 'approved'
			   AND revision_version = $3`,
			instance.DocumentID, req.TenantID, instance.RevisionVersion,
		)
		if err != nil {
			return fmt.Errorf("publishApproved: update document state: %w", err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("publishApproved: rows affected: %w", err)
		}
		if affected == 0 {
			return infrastructure.ErrStaleRevision
		}

		// Step 4: emit "document_published" governance event.
		now := s.clock.Now()
		payloadMap := map[string]any{
			"instance_id":      req.InstanceID,
			"revision_version": instance.RevisionVersion,
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("publishApproved: marshal event payload: %w", err)
		}
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeDocumentPublished,
			ActorUserID:  req.PublishedBy,
			ResourceType: "document",
			ResourceID:   instance.DocumentID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("publishApproved: emit event: %w", err)
		}

		// Additive in-tx domain-event enqueue (ADR-0044; F3.3). After audit emit.
		if s.lifecycleEnqueuer != nil {
			cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, instance.DocumentID)
			if err != nil {
				return fmt.Errorf("publishApproved: load cd id for lifecycle event: %w", err)
			}
			largs := docsdomain.LifecycleEventArgs{
				EventID:              uuid.NewString(),
				TenantID:             req.TenantID,
				EventType:            docsdomain.EventTypeDocumentPublished,
				ResourceType:         "document",
				ResourceID:           instance.DocumentID,
				ControlledDocumentID: cdID,
				OccurredAt:           now,
			}
			if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
				return fmt.Errorf("publishApproved: enqueue lifecycle event: %w", err)
			}
		}

		result = PublishResult{DocumentID: instance.DocumentID, NewStatus: string(docsdomain.DocStatusPublished)}
		return nil
	})
	if err != nil {
		return PublishResult{}, err
	}
	return result, nil
}

// ErrEffectiveDateInPast is returned when SchedulePublish is called with an
// effective_date that is not strictly in the future.
var ErrEffectiveDateInPast = errors.New("approval: effective_date must be in the future")

// SchedulePublishRequest carries all inputs for SchedulePublish.
type SchedulePublishRequest struct {
	TenantID             string
	InstanceID           string
	ExpectedRevision     int       // optional client precondition; 0 means "not provided"
	EffectiveDate        time.Time // must be strictly after clock.Now()
	ScheduledBy          string
	SupersededDocumentID string
	// EffectiveTo is the optional expiry date (M6 F6.2 contract §0.1/§2):
	// nil leaves the column unwritten (NULL).
	EffectiveTo *time.Time
	// ReviewDueAt is the optional initial periodic-review due date set at
	// schedule/publish time (M6 F6.2 contract §0.1/§2): nil leaves the
	// column unwritten (NULL).
	ReviewDueAt *time.Time
}

// SchedulePublishResult is returned on successful scheduling.
type SchedulePublishResult struct {
	DocumentID         string
	EffectiveDate      time.Time
	ScheduleGeneration int64
}

// WithScheduledPublishEnqueuer wires the job enqueuer used to schedule the
// deferred publish for a future effective date.
func (s *PublishService) WithScheduledPublishEnqueuer(enqueuer ScheduledPublishEnqueuer) *PublishService {
	s.scheduledPublishEnqueuer = enqueuer
	return s
}

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer used to publish
// a lifecycle event once the document is published.
func (s *PublishService) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *PublishService {
	s.lifecycleEnqueuer = e
	return s
}

// SchedulePublish transitions an approved document to "scheduled" state with a
// future effective date. It guards against past dates, performs an OCC UPDATE
// on the documents table (approved → scheduled), emits a "publish_scheduled"
// governance event, and commits.
func (s *PublishService) SchedulePublish(ctx context.Context, runner db.TxRunner, req SchedulePublishRequest) (SchedulePublishResult, error) {
	// Step 1: guard — effective_date must be strictly in the future.
	if !req.EffectiveDate.After(s.clock.Now()) {
		return SchedulePublishResult{}, ErrEffectiveDateInPast
	}

	var result SchedulePublishResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)


		// Step 3: load the approval instance.
		instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return fmt.Errorf("schedulePublish: load instance: %w", err)
		}
		if instance == nil {
			return infrastructure.ErrNoActiveInstance
		}
		if req.ExpectedRevision > 0 && req.ExpectedRevision != instance.RevisionVersion {
			return infrastructure.ErrStaleRevision
		}

		// Verify instance is in approved state.
		if instance.Status != domain.InstanceApproved {
			return ErrInstanceNotApproved
		}

		// No-fallback (F6, spec §11): scheduling is also a publish-family
		// transition off the same approved instance; same fail-closed pin guard
		// as PublishApproved (see its comment for full rationale).
		if instance.FrozenContentHash == nil {
			return ErrContentHashMismatch
		}

		// document.publish is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.DocumentID)
		if err != nil {
			return fmt.Errorf("schedulePublish: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentPublish), areaCode); err != nil {
			return err
		}
		// Scheduling also mutates public.documents and must satisfy the shared
		// document-edit tripwire in the same transaction.
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}
		supersededDocumentID, err := s.repo.LoadCurrentPublishedHeadForDocument(ctx, tx, req.TenantID, instance.DocumentID)
		if err != nil {
			return fmt.Errorf("schedulePublish: load current published head: %w", err)
		}
		if req.SupersededDocumentID != "" {
			if req.SupersededDocumentID == instance.DocumentID {
				return infrastructure.ErrInvalidScheduledSupersedeTarget
			}
			if supersededDocumentID == "" || req.SupersededDocumentID != supersededDocumentID {
				return infrastructure.ErrInvalidScheduledSupersedeTarget
			}
		}
		if supersededDocumentID == instance.DocumentID {
			return infrastructure.ErrInvalidScheduledSupersedeTarget
		}

		// Step 4: OCC transition the document from "approved" to "scheduled".
		// Friendly first-line legality check (M4/F4.1) mirrors the DB trigger; the
		// OCC WHERE below remains the atomic CAS + optimistic-lock enforcement.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusApproved, docsdomain.DocStatusScheduled); err != nil {
			return err
		}
		var scheduleGeneration int64
		err = tx.QueryRowContext(ctx, `
			UPDATE documents
			   SET status           = 'scheduled',
			       effective_from   = $1,
			       superseded_document_id = NULLIF($2, '')::uuid,
			       effective_to     = $6,
			       review_due_at    = $7,
			       revision_version = revision_version + 1,
			       schedule_generation = schedule_generation + 1
			 WHERE id               = $3
			   AND tenant_id        = $4
			   AND status           = 'approved'
			   AND revision_version = $5
			RETURNING schedule_generation`,
			req.EffectiveDate.UTC(), supersededDocumentID, instance.DocumentID, req.TenantID, instance.RevisionVersion,
			nullableTime(req.EffectiveTo), nullableTime(req.ReviewDueAt),
		).Scan(&scheduleGeneration)
		if errors.Is(err, sql.ErrNoRows) {
			return infrastructure.ErrStaleRevision
		}
		if err != nil {
			return fmt.Errorf("schedulePublish: update document state: %w", err)
		}

		if s.scheduledPublishEnqueuer != nil {
			if err := s.scheduledPublishEnqueuer.EnqueueScheduledPublishTx(ctx, tx, ScheduledPublishJobInput{
				TenantID:                req.TenantID,
				DocumentID:              instance.DocumentID,
				ExpectedRevisionVersion: instance.RevisionVersion + 1,
				ScheduledEffectiveAt:    req.EffectiveDate.UTC(),
				ScheduleGeneration:      scheduleGeneration,
			}); err != nil {
				return fmt.Errorf("schedulePublish: enqueue scheduled publish job: %w", err)
			}
		}

		// Step 5: emit "publish_scheduled" governance event.
		now := s.clock.Now()
		payloadMap := map[string]any{
			"effective_date": req.EffectiveDate.UTC().Format(time.RFC3339),
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("schedulePublish: marshal event payload: %w", err)
		}
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypePublishScheduled,
			ActorUserID:  req.ScheduledBy,
			ResourceType: "document",
			ResourceID:   instance.DocumentID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("schedulePublish: emit event: %w", err)
		}

		result = SchedulePublishResult{
			DocumentID:         instance.DocumentID,
			EffectiveDate:      req.EffectiveDate,
			ScheduleGeneration: scheduleGeneration,
		}
		return nil
	})
	if err != nil {
		return SchedulePublishResult{}, err
	}
	return result, nil
}
