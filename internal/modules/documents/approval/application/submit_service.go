package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// SubmitService handles document submission for approval.
type SubmitService struct {
	repo    repository.ApprovalRepository
	emitter EventEmitter
	clock   Clock
	cdRead  controlleddocumentsdomain.CDFieldReader
}

// SubmitRequest carries all inputs for SubmitRevisionForReview.
type SubmitRequest struct {
	TenantID        string
	DocumentID      string // UUID as string
	RouteID         string // UUID as string
	SubmittedBy     string // user_id
	RevisionTitle   string
	ContentFormData map[string]any // raw form data for hashing
	RevisionVersion int            // OCC version from caller
	RevisionNumber  int            // governed documents.revision_number
	IdempotencyKey  string         // client Idempotency-Key header, threaded from the handler
}

// SubmitResult is returned on successful submission.
type SubmitResult struct {
	InstanceID string // UUID of created approval_instance
}

// SubmitRevisionForReview creates a new approval instance for the document revision.
// Returns repository.ErrDuplicateSubmission (unwrapped) when a concurrent submission
// with the same idempotency key already exists so callers can check via errors.Is.
func (s *SubmitService) SubmitRevisionForReview(ctx context.Context, runner db.TxRunner, req SubmitRequest) (SubmitResult, error) {
	// Step 1: validate payload — no float64 values.
	if err := ValidateEventPayload(req.ContentFormData); err != nil {
		return SubmitResult{}, err
	}

	// Step 2: compute content hash.
	contentHash, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.DocumentID,
		RevisionNumber: req.RevisionNumber,
		FormData:       req.ContentFormData,
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("submit: content hash: %w", err)
	}

	// Step 3: require the caller-supplied idempotency key. It is the client's
	// Idempotency-Key header, threaded through the handler, and becomes the
	// approval_instances.idempotency_key value so the
	// UNIQUE(document_id, idempotency_key) constraint is a real DB-enforced
	// replay backstop behind the HTTP idempotency middleware (F-D4). A
	// server-clock-derived key is never used: a retry seconds later would
	// otherwise mint a fresh key, bypass the constraint, and create a duplicate
	// approval instance.
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		return SubmitResult{}, ErrIdempotencyKeyRequired
	}

	// Step 4: run the submission as one unit of work; the runner owns
	// begin/commit/rollback.
	var instanceID string
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.SeedTxIdentity(ctx, tx, req.TenantID, req.SubmittedBy); err != nil {
			return fmt.Errorf("submit: %w", err)
		}

		// document.submit is area-grade: pass the resolved area as-is ("" fail-closes,
		// denying non-system actors). ADR 0022 Phase 11 (F7): shared LoadDocumentAreaCode.
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, req.DocumentID)
		if err != nil {
			return fmt.Errorf("submit: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode); err != nil {
			return err
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}

		// Step 5: load route with stages.
		route, err := s.repo.LoadRoute(ctx, tx, req.TenantID, req.RouteID)
		if err != nil {
			return fmt.Errorf("submit: load route: %w", err)
		}

		// Step 6: validate route structural invariants.
		if err := route.Validate(); err != nil {
			return fmt.Errorf("submit: invalid route: %w", err)
		}

		revisionTitle, err := normalizeGovernedRevisionTitle(req.RevisionNumber, req.RevisionTitle)
		if err != nil {
			return err
		}

		// Step 7: create the approval instance.
		instanceID = uuid.New().String()
		now := s.clock.Now()

		inst := domain.Instance{
			ID:                   instanceID,
			TenantID:             req.TenantID,
			DocumentID:           req.DocumentID,
			RouteID:              req.RouteID,
			RouteVersionSnapshot: route.Version,
			Status:               domain.InstanceInProgress,
			SubmittedBy:          req.SubmittedBy,
			SubmittedAt:          now,
			ContentHashAtSubmit:  contentHash,
			IdempotencyKey:       idempotencyKey,
			RevisionVersion:      req.RevisionVersion,
		}

		if err := s.repo.InsertInstance(ctx, tx, inst); err != nil {
			// Pass through duplicate submission sentinel unwrapped.
			if errors.Is(err, repository.ErrDuplicateSubmission) {
				return err
			}
			return fmt.Errorf("submit: %w", err)
		}

		// Step 8: create stage instances.
		// First stage is active; all others start pending.
		stageInstances := make([]domain.StageInstance, len(route.Stages))
		for i, stage := range route.Stages {
			status := domain.StagePending
			var openedAt *time.Time
			if i == 0 {
				status = domain.StageActive
				openedAt = &now
			}
			eligibleIDs, err := s.repo.ResolveEligibleActors(ctx, tx, req.TenantID, stage.AreaCode, stage.RequiredRole)
			if err != nil {
				return fmt.Errorf("submit: resolve eligible actors for stage %d: %w", stage.Order, err)
			}
			stageInstances[i] = domain.StageInstance{
				ID:                         uuid.New().String(),
				ApprovalInstanceID:         instanceID,
				StageOrder:                 stage.Order,
				NameSnapshot:               stage.Name,
				RequiredRoleSnapshot:       stage.RequiredRole,
				RequiredCapabilitySnapshot: stage.RequiredCapability,
				AreaCodeSnapshot:           stage.AreaCode,
				QuorumSnapshot:             stage.Quorum,
				QuorumMSnapshot:            stage.QuorumM,
				OnEligibilityDriftSnapshot: stage.OnEligibilityDrift,
				EligibleActorIDs:           eligibleIDs,
				Status:                     status,
				OpenedAt:                   openedAt,
			}
		}

		if err := s.repo.InsertStageInstances(ctx, tx, stageInstances); err != nil {
			return fmt.Errorf("submit: %w", err)
		}

		// Step 8b: transition document draft → under_review.
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'under_review',
			       revision_title   = $4,
			       revision_version = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'draft'
			   AND revision_version = $3`,
			req.DocumentID, req.TenantID, req.RevisionVersion, revisionTitle,
		)
		if err != nil {
			return fmt.Errorf("submit: transition document to under_review: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("submit: rows affected: %w", err)
		}
		if n == 0 {
			return repository.ErrStaleRevision
		}

		// Step 9: emit governance event.
		payloadMap := map[string]any{
			"instance_id":  instanceID,
			"route_id":     req.RouteID,
			"content_hash": contentHash,
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("submit: marshal event payload: %w", err)
		}

		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    "approval_submitted",
			ActorUserID:  req.SubmittedBy,
			ResourceType: "document",
			ResourceID:   req.DocumentID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("submit: emit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return SubmitResult{}, err
	}

	// Step 5: return result.
	return SubmitResult{InstanceID: instanceID}, nil
}

const defaultInitialRevisionTitle = "Criacao do documento"

// ErrRevisionTitleRequired is returned when a revision title is required but was not supplied.
var ErrRevisionTitleRequired = errors.New("revisionTitle is required")

// ErrIdempotencyKeyRequired is returned when SubmitRevisionForReview is called
// without a client idempotency key. Both HTTP entrypoints (the submit handler and
// finalizeDocument) guarantee a non-empty key, so this is a fail-closed
// defense-in-depth guard at the service boundary.
var ErrIdempotencyKeyRequired = errors.New("submit: idempotency key is required")

func normalizeGovernedRevisionTitle(revisionNumber int, title string) (string, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed != "" {
		return trimmed, nil
	}
	if revisionNumber == 0 {
		return defaultInitialRevisionTitle, nil
	}
	return "", ErrRevisionTitleRequired
}
