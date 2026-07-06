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
	"metaldocs/internal/modules/documents/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// SubmitService handles document submission for approval.
type SubmitService struct {
	repo    infrastructure.ApprovalRepository
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
	ReasonForChange string         // structured 21 CFR Part 11 change reason (F6.3); distinct from RevisionTitle
	ReasonCategory  string         // optional enum classifying ReasonForChange (F6.3); "" = unset
	ContentFormData map[string]any // raw form data for hashing
	RevisionVersion int            // OCC version from caller
	IdempotencyKey  string         // client Idempotency-Key header, threaded from the handler
}

// SubmitResult is returned on successful submission.
type SubmitResult struct {
	InstanceID string // UUID of created approval_instance
}

// SubmitRevisionForReview creates a new approval instance for the document revision.
// Returns infrastructure.ErrDuplicateSubmission (unwrapped) when a concurrent submission
// with the same idempotency key already exists so callers can check via errors.Is.
func (s *SubmitService) SubmitRevisionForReview(ctx context.Context, runner db.TxRunner, req SubmitRequest) (SubmitResult, error) {
	// Step 1: validate payload — no float64 values.
	if err := ValidateEventPayload(req.ContentFormData); err != nil {
		return SubmitResult{}, err
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
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

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

		// Step 4b: derive the governed documents.revision_number in-tx (T8b).
		// This value is NEVER trusted from the client — SubmitRequest has no
		// RevisionNumber field. Reading it here (tenant-scoped, via the caller's
		// tx for GUC/RLS correctness) is what makes the REV>=1
		// reason_for_change/revision_title gates and the content-hash binding
		// fire correctly on live traffic instead of silently defaulting to 0.
		revisionNumber, err := s.repo.LoadGovernedRevisionNumber(ctx, tx, req.TenantID, req.DocumentID)
		if err != nil {
			return fmt.Errorf("submit: load governed revision number: %w", err)
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

		revisionTitle, err := normalizeGovernedRevisionTitle(revisionNumber, req.RevisionTitle)
		if err != nil {
			return err
		}

		reasonForChange, reasonCategory, err := normalizeReasonForChange(revisionNumber, req.ReasonForChange, req.ReasonCategory)
		if err != nil {
			return err
		}

		// Step 6b: compute the content hash bound to the SAME derived governed
		// revision_number (HS-2: this changes the hash for REV>=1 submits versus
		// the prior always-0 behavior — see submit_service.go package docs /
		// T8b handoff notes for the governed-semantics consequence).
		contentHash, err := ComputeContentHash(ContentHashInput{
			TenantID:       req.TenantID,
			DocumentID:     req.DocumentID,
			RevisionNumber: revisionNumber,
			FormData:       req.ContentFormData,
		})
		if err != nil {
			return fmt.Errorf("submit: content hash: %w", err)
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
			if errors.Is(err, infrastructure.ErrDuplicateSubmission) {
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

		// Step 8b: transition document draft → under_review. Friendly first-line
		// legality check (M4/F4.1) mirrors the DB trigger; the OCC WHERE below
		// remains the atomic CAS + optimistic-lock enforcement.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusDraft, docsdomain.DocStatusUnderReview); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status            = 'under_review',
			       revision_title    = $4,
			       reason_for_change = $5,
			       reason_category   = $6,
			       revision_version  = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'draft'
			   AND revision_version = $3`,
			req.DocumentID, req.TenantID, req.RevisionVersion, revisionTitle,
			nullableString(reasonForChange), nullableString(reasonCategory),
		)
		if err != nil {
			return fmt.Errorf("submit: transition document to under_review: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("submit: rows affected: %w", err)
		}
		if n == 0 {
			return infrastructure.ErrStaleRevision
		}

		// Step 9: emit governance event. reason_for_change/reason_category ride
		// along in the SAME in-tx payload (F6.3 §5.2) — no second write, no
		// separate audit call; empty/unset fields are omitted rather than
		// serialized as "".
		payloadMap := map[string]any{
			"instance_id":  instanceID,
			"route_id":     req.RouteID,
			"content_hash": contentHash,
		}
		if reasonForChange != "" {
			payloadMap["reason_for_change"] = reasonForChange
		}
		if reasonCategory != "" {
			payloadMap["reason_category"] = reasonCategory
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

// validReasonCategories mirrors the DB CHECK ck_documents_reason_category
// (db/migrations/0274_document_review_and_reason.sql). The app-level check is
// the friendly first line; the DB CHECK remains the enforced last line.
var validReasonCategories = map[string]bool{
	"content":         true,
	"corrective":      true,
	"regulatory":      true,
	"periodic_review": true,
	"administrative":  true,
}

// ErrReasonForChangeRequired is returned when a REV>=1 submission omits the
// structured reason_for_change (F6.3 contract §5.3). Not required at REV 0
// (initial creation), mirroring normalizeGovernedRevisionTitle's REV-0 default.
var ErrReasonForChangeRequired = errors.New("reasonForChange is required for revision >= 1")

// ErrReasonCategoryInvalid is returned when a non-empty reason_category is not
// one of the fixed enum values (content, corrective, regulatory,
// periodic_review, administrative). Friendly first-line check mirroring the DB
// CHECK ck_documents_reason_category.
var ErrReasonCategoryInvalid = errors.New("reasonCategory must be one of: content, corrective, regulatory, periodic_review, administrative")

// normalizeReasonForChange trims and validates reason_for_change/reason_category
// (F6.3 contract §5.2/§5.3): required for REV>=1 (friendly-first-line before the
// DB CHECK/422 mapping), optional at REV 0; reason_category, when present, must
// be in the fixed enum.
func normalizeReasonForChange(revisionNumber int, reason, category string) (string, string, error) {
	trimmedReason := strings.TrimSpace(reason)
	trimmedCategory := strings.TrimSpace(category)

	if trimmedReason == "" && revisionNumber != 0 {
		return "", "", ErrReasonForChangeRequired
	}
	if trimmedCategory != "" && !validReasonCategories[trimmedCategory] {
		return "", "", ErrReasonCategoryInvalid
	}
	return trimmedReason, trimmedCategory, nil
}

// nullableString maps an empty string to a driver NULL bind (any(nil)) so an
// unset optional field persists as SQL NULL rather than the literal empty
// string, matching the migration's nullable columns (expand-only, F6.3 §5.1).
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
