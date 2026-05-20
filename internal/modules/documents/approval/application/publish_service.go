package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// PublishService handles transitioning an approved document to published state.
type PublishService struct {
	repo                     repository.ApprovalRepository
	emitter                  EventEmitter
	clock                    Clock
	scheduledPublishEnqueuer ScheduledPublishEnqueuer
}

// ErrInstanceNotApproved is returned when PublishApproved is called on an
// instance whose status is not "approved".
var ErrInstanceNotApproved = errors.New("approval: instance is not in approved state")

// PublishRequest carries all inputs for PublishApproved.
type PublishRequest struct {
	TenantID    string
	InstanceID  string
	PublishedBy string // user_id triggering publish
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
func (s *PublishService) PublishApproved(ctx context.Context, db *sql.DB, req PublishRequest) (PublishResult, error) {
	// Step 1: begin transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return PublishResult{}, fmt.Errorf("publishApproved: begin tx: %w", err)
	}

	// Step 2: load the approval instance.
	instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return PublishResult{}, repository.ErrNoActiveInstance
		}
		return PublishResult{}, fmt.Errorf("publishApproved: load instance: %w", err)
	}
	if instance == nil {
		_ = tx.Rollback()
		return PublishResult{}, repository.ErrNoActiveInstance
	}

	// Verify instance is in approved state.
	if instance.Status != domain.InstanceApproved {
		_ = tx.Rollback()
		return PublishResult{}, ErrInstanceNotApproved
	}

	if err := setAuthzGUC(ctx, tx, req.TenantID, req.PublishedBy); err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: %w", err)
	}

	areaCode, err := loadDocumentAreaCode(ctx, tx, req.TenantID, instance.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: load document area: %w", err)
	}
	if err := authz.Require(ctx, tx, "doc.publish", areaCode); err != nil {
		_ = tx.Rollback()
		return PublishResult{}, err
	}
	// public.documents is protected by the shared documents tripwire, so the
	// approval publish transaction must assert document.edit before updating it.
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
		_ = tx.Rollback()
		return PublishResult{}, err
	}

	// Step 3: transition the document from "approved" to "published".
	// Status check is the concurrency guard — atomic UPDATE catches any racing transition.
	result, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET status           = 'published',
		       revision_version = revision_version + 1
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'approved'`,
		instance.DocumentID, req.TenantID,
	)
	if err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: update document state: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: rows affected: %w", err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return PublishResult{}, repository.ErrStaleRevision
	}

	// Step 4: emit "document_published" governance event.
	now := s.clock.Now()
	payloadMap := map[string]any{
		"instance_id":      req.InstanceID,
		"revision_version": instance.RevisionVersion,
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: marshal event payload: %w", err)
	}
	event := GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    "document_published",
		ActorUserID:  req.PublishedBy,
		ResourceType: "document",
		ResourceID:   instance.DocumentID,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		_ = tx.Rollback()
		return PublishResult{}, fmt.Errorf("publishApproved: emit event: %w", err)
	}

	// Step 5: commit.
	if err := tx.Commit(); err != nil {
		return PublishResult{}, fmt.Errorf("publishApproved: commit: %w", err)
	}

	return PublishResult{DocumentID: instance.DocumentID, NewStatus: "published"}, nil
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
}

// SchedulePublishResult is returned on successful scheduling.
type SchedulePublishResult struct {
	DocumentID         string
	EffectiveDate      time.Time
	ScheduleGeneration int64
}

func (s *PublishService) WithScheduledPublishEnqueuer(enqueuer ScheduledPublishEnqueuer) *PublishService {
	s.scheduledPublishEnqueuer = enqueuer
	return s
}

// SchedulePublish transitions an approved document to "scheduled" state with a
// future effective date. It guards against past dates, performs an OCC UPDATE
// on the documents table (approved → scheduled), emits a "publish_scheduled"
// governance event, and commits.
func (s *PublishService) SchedulePublish(ctx context.Context, db *sql.DB, req SchedulePublishRequest) (SchedulePublishResult, error) {
	// Step 1: guard — effective_date must be strictly in the future.
	if !req.EffectiveDate.After(s.clock.Now()) {
		return SchedulePublishResult{}, ErrEffectiveDateInPast
	}

	// Step 2: begin transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: begin tx: %w", err)
	}

	// Step 3: load the approval instance.
	instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return SchedulePublishResult{}, repository.ErrNoActiveInstance
		}
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: load instance: %w", err)
	}
	if instance == nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, repository.ErrNoActiveInstance
	}
	if req.ExpectedRevision > 0 && req.ExpectedRevision != instance.RevisionVersion {
		_ = tx.Rollback()
		return SchedulePublishResult{}, repository.ErrStaleRevision
	}

	// Verify instance is in approved state.
	if instance.Status != domain.InstanceApproved {
		_ = tx.Rollback()
		return SchedulePublishResult{}, ErrInstanceNotApproved
	}

	if err := setAuthzGUC(ctx, tx, req.TenantID, req.ScheduledBy); err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: %w", err)
	}

	areaCode, err := loadDocumentAreaCode(ctx, tx, req.TenantID, instance.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: load document area: %w", err)
	}
	if err := authz.Require(ctx, tx, "doc.publish", areaCode); err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, err
	}
	// Scheduling also mutates public.documents and must satisfy the shared
	// document-edit tripwire in the same transaction.
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, err
	}
	supersededDocumentID, err := s.repo.LoadCurrentPublishedHeadForDocument(ctx, tx, req.TenantID, instance.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: load current published head: %w", err)
	}
	if req.SupersededDocumentID != "" {
		if req.SupersededDocumentID == instance.DocumentID {
			_ = tx.Rollback()
			return SchedulePublishResult{}, repository.ErrInvalidScheduledSupersedeTarget
		}
		if supersededDocumentID == "" || req.SupersededDocumentID != supersededDocumentID {
			_ = tx.Rollback()
			return SchedulePublishResult{}, repository.ErrInvalidScheduledSupersedeTarget
		}
	}
	if supersededDocumentID == instance.DocumentID {
		_ = tx.Rollback()
		return SchedulePublishResult{}, repository.ErrInvalidScheduledSupersedeTarget
	}

	// Step 4: OCC transition the document from "approved" to "scheduled".
	var scheduleGeneration int64
	err = tx.QueryRowContext(ctx, `
		UPDATE documents
		   SET status           = 'scheduled',
		       effective_from   = $1,
		       superseded_document_id = NULLIF($2, '')::uuid,
		       revision_version = revision_version + 1,
		       schedule_generation = schedule_generation + 1
		 WHERE id               = $3
		   AND tenant_id        = $4
		   AND status           = 'approved'
		   AND revision_version = $5
		RETURNING schedule_generation`,
		req.EffectiveDate.UTC(), supersededDocumentID, instance.DocumentID, req.TenantID, instance.RevisionVersion,
	).Scan(&scheduleGeneration)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return SchedulePublishResult{}, repository.ErrStaleRevision
	}
	if err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: update document state: %w", err)
	}

	if s.scheduledPublishEnqueuer != nil {
		if err := s.scheduledPublishEnqueuer.EnqueueScheduledPublishTx(ctx, tx, ScheduledPublishJobInput{
			TenantID:                req.TenantID,
			DocumentID:              instance.DocumentID,
			ExpectedRevisionVersion: instance.RevisionVersion + 1,
			ScheduledEffectiveAt:    req.EffectiveDate.UTC(),
			ScheduleGeneration:      scheduleGeneration,
		}); err != nil {
			_ = tx.Rollback()
			return SchedulePublishResult{}, fmt.Errorf("schedulePublish: enqueue scheduled publish job: %w", err)
		}
	}

	// Step 5: emit "publish_scheduled" governance event.
	now := s.clock.Now()
	payloadMap := map[string]any{
		"effective_date": req.EffectiveDate.UTC().Format(time.RFC3339),
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: marshal event payload: %w", err)
	}
	event := GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    "publish_scheduled",
		ActorUserID:  req.ScheduledBy,
		ResourceType: "document",
		ResourceID:   instance.DocumentID,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		_ = tx.Rollback()
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: emit event: %w", err)
	}

	// Step 6: commit.
	if err := tx.Commit(); err != nil {
		return SchedulePublishResult{}, fmt.Errorf("schedulePublish: commit: %w", err)
	}

	return SchedulePublishResult{
		DocumentID:         instance.DocumentID,
		EffectiveDate:      req.EffectiveDate,
		ScheduleGeneration: scheduleGeneration,
	}, nil
}
