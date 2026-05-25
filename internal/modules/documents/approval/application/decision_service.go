package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

var ErrApprovalBlockedByUnresolvedComments = errors.New("approval: unresolved comments block approval")

type FreezeInvoker interface {
	Freeze(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, approver docapp.ApproverContext) error
}

type PDFDispatchInvoker interface {
	Dispatch(ctx context.Context, tenantID, revisionID string) error
}

// PDFOutboxEnqueuer enqueues a PDF dispatch inside the approval transaction.
type PDFOutboxEnqueuer interface {
	Enqueue(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, contentHash []byte) error
}

// DecisionService handles approver approve/reject decisions.
type DecisionService struct {
	repo          repository.ApprovalRepository
	emitter       EventEmitter
	clock         Clock
	freezeInvoker FreezeInvoker
	// deprecated: post-commit best-effort dispatcher; replaced by pdfOutbox.
	pdfDispatcher PDFDispatchInvoker
	pdfOutbox     PDFOutboxEnqueuer
}

func NewDecisionService(
	repo repository.ApprovalRepository,
	emitter EventEmitter,
	clock Clock,
	freezeInvoker FreezeInvoker,
	pdfDispatcher PDFDispatchInvoker,
) *DecisionService {
	return &DecisionService{
		repo:          repo,
		emitter:       emitter,
		clock:         clock,
		freezeInvoker: freezeInvoker,
		pdfDispatcher: pdfDispatcher,
	}
}

// WithPDFOutbox sets the transactional outbox enqueuer, replacing the post-commit dispatcher.
func (s *DecisionService) WithPDFOutbox(enqueuer PDFOutboxEnqueuer) *DecisionService {
	s.pdfOutbox = enqueuer
	return s
}

// SignoffRequest carries all inputs for RecordSignoff.
type SignoffRequest struct {
	TenantID                string
	InstanceID              string
	StageInstanceID         string
	ActorUserID             string
	Decision                string // "approve" or "reject"
	Comment                 string
	SignatureMethod         string
	SignaturePayload        map[string]any
	ContentFormData         map[string]any // current document content for hash
	ExpectedRevisionVersion int
	Capabilities            []string
}

// SignoffResult is returned by RecordSignoff.
type SignoffResult struct {
	StageCompleted   bool
	InstanceApproved bool // true when all stages complete
	InstanceRejected bool // true when a reject decision collapses instance
}

// RecordSignoff records an approve or reject decision for the given stage instance.
// Approve path only; reject path shares this method and is gated by req.Decision.
func (s *DecisionService) RecordSignoff(ctx context.Context, db *sql.DB, req SignoffRequest) (SignoffResult, error) {
	// Step 1: validate signature payload — no float64 values.
	if err := ValidateEventPayload(req.SignaturePayload); err != nil {
		return SignoffResult{}, err
	}

	// Step 2: begin transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return SignoffResult{}, fmt.Errorf("recordSignoff: begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, req.TenantID, req.ActorUserID); err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: %w", err)
	}

	// Step 4: load approval instance; child stage rows locked FOR UPDATE inside LoadInstance (J1).
	instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return SignoffResult{}, fmt.Errorf("recordSignoff: %w", repository.ErrNoActiveInstance)
		}
		return SignoffResult{}, fmt.Errorf("recordSignoff: load instance: %w", err)
	}
	if instance == nil {
		_ = tx.Rollback()
		return SignoffResult{}, repository.ErrNoActiveInstance
	}
	if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
		_ = tx.Rollback()
		return SignoffResult{}, repository.ErrStaleRevision
	}

	// Reject if instance is already terminal.
	if instance.Status != domain.InstanceInProgress {
		_ = tx.Rollback()
		return SignoffResult{}, repository.ErrInstanceCompleted
	}

	areaCode, err := loadDocumentAreaCode(ctx, tx, req.TenantID, instance.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: load document area: %w", err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSignoff), areaCode); err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, err
	}

	formData, err := loadCurrentDocumentFormData(ctx, tx, req.TenantID, instance.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: load document form data: %w", err)
	}
	contentHash, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.InstanceID, // keyed on instance for signoff hashing
		RevisionNumber: 0,              // signoff hash does not embed revision
		FormData:       formData,
	})
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: content hash: %w", err)
	}

	// Step 5: identify active stage.
	activeStage := instance.Active()
	if activeStage == nil {
		_ = tx.Rollback()
		return SignoffResult{}, domain.ErrNoActiveStage
	}
	// Ensure the requested StageInstanceID matches the active stage.
	if req.StageInstanceID == "" || activeStage.ID != req.StageInstanceID {
		_ = tx.Rollback()
		return SignoffResult{}, repository.ErrStageNotActive
	}

	// Step 5b: eligibility check — actor must be in the eligible_actor_ids snapshot (J1).
	if err := domain.CheckEligibility(req.ActorUserID, activeStage.EligibleActorIDs); err != nil {
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    "signoff.rejected",
			ActorUserID:  req.ActorUserID,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			Reason:       "not_eligible",
			OccurredAt:   s.clock.Now(),
		}
		_ = tx.Rollback()
		if emitErr := s.emitEligibilityRejection(ctx, db, req.TenantID, req.ActorUserID, event); emitErr != nil {
			return SignoffResult{}, fmt.Errorf("recordSignoff: emit eligibility rejection: %w", emitErr)
		}
		return SignoffResult{}, err
	}

	// Step 6: SoD check — author cannot sign, actor cannot sign twice in same instance.
	priorSignoffs, err := s.loadPriorSignoffs(ctx, tx, req.TenantID, req.InstanceID, activeStage.ID)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: load prior signoffs: %w", err)
	}
	if err := domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, priorSignoffs); err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, err
	}

	// Step 7: build the domain Signoff value object.
	sigPayload, err := marshalSignaturePayload(req.SignaturePayload)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: marshal signature payload: %w", err)
	}

	var actorDisplayName sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT display_name
		  FROM metaldocs.iam_users
		 WHERE user_id = $1
		   AND tenant_id = $2::uuid`,
		req.ActorUserID, req.TenantID,
	).Scan(&actorDisplayName); err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: lookup actor display name: %w", err)
	}

	now := s.clock.Now()
	signoff, err := domain.NewSignoff(domain.SignoffParams{
		ID:                       uuid.New().String(),
		ApprovalInstanceID:       req.InstanceID,
		StageInstanceID:          activeStage.ID,
		ActorUserID:              req.ActorUserID,
		ActorTenantID:            req.TenantID,
		Decision:                 domain.Decision(req.Decision),
		Comment:                  req.Comment,
		SignedAt:                 now,
		SignatureMethod:          req.SignatureMethod,
		SignaturePayload:         sigPayload,
		ContentHash:              contentHash,
		ActorDisplayNameSnapshot: actorDisplayName.String,
	})
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: build signoff: %w", err)
	}

	// Step 8: persist the signoff, handling idempotent replay.
	insertResult, err := s.repo.InsertSignoff(ctx, tx, *signoff)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, repository.ErrActorAlreadySigned) {
			return SignoffResult{}, err
		}
		return SignoffResult{}, fmt.Errorf("recordSignoff: insert signoff: %w", err)
	}
	if insertResult.WasReplay {
		// Idempotent replay: commit and return neutral result (stage not advanced again).
		if err := tx.Commit(); err != nil {
			return SignoffResult{}, fmt.Errorf("recordSignoff: commit replay: %w", err)
		}
		return SignoffResult{}, nil
	}

	// Step 9: collect all signoffs for the active stage to evaluate quorum.
	allStageSignoffs, err := s.loadStageSignoffs(ctx, tx, req.TenantID, activeStage.ID)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: load stage signoffs: %w", err)
	}

	// Step 10: evaluate quorum.
	approvals, rejections := splitSignoffs(allStageSignoffs)
	currentEligible := activeStage.EligibleActorIDs
	if activeStage.OnEligibilityDriftSnapshot != domain.DriftKeepSnapshot {
		currentEligible, err = resolveEligibleActors(ctx, tx, req.TenantID, activeStage.AreaCodeSnapshot, activeStage.RequiredRoleSnapshot)
		if err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: resolve current eligible actors: %w", err)
		}
	}
	drift := domain.ApplyEligibilityDrift(*activeStage, currentEligible)
	effectiveDenominator := drift.EffectiveDenominator
	skipLegacyQuorumFallback := false
	if skipLegacyQuorumFallback && effectiveDenominator == 0 {
		// Fallback: treat every eligible actor as in scope.
		effectiveDenominator = len(activeStage.EligibleActorIDs)
	}
	if skipLegacyQuorumFallback && effectiveDenominator == 0 {
		_ = tx.Rollback()
		return SignoffResult{}, domain.ErrEmptyEligiblePool
	}
	outcome := drift.ForcedOutcome
	if outcome == domain.QuorumPending {
		outcome = domain.EvaluateQuorum(*activeStage, approvals, rejections, effectiveDenominator)
		if outcome == domain.QuorumError {
			_ = tx.Rollback()
			return SignoffResult{}, domain.ErrEmptyEligiblePool
		}
	}

	var result SignoffResult
	var shouldDispatchPDF bool
	var pdfTenantID string
	var pdfRevisionID string

	switch outcome {
	case domain.QuorumApprovedStage:
		// Step 11a: mark stage completed.
		if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageCompleted, domain.StageActive); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: complete stage: %w", err)
		}
		result.StageCompleted = true

		// Advance the in-memory instance to determine next step.
		if err := instance.AdvanceStage(); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: advance stage: %w", err)
		}

		if instance.Status == domain.InstanceApproved {
			blocked, err := hasUnresolvedComments(ctx, tx, req.TenantID, instance.DocumentID)
			if err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: check unresolved comments: %w", err)
			}
			if blocked {
				_ = tx.Rollback()
				return SignoffResult{}, ErrApprovalBlockedByUnresolvedComments
			}

			// All stages done — complete instance.
			if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
				domain.InstanceApproved, domain.InstanceInProgress, &now); err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: complete instance: %w", err)
			}
			if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, err
			}
			if s.freezeInvoker == nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: freezeInvoker not configured")
			}
			if err := s.freezeInvoker.Freeze(ctx, tx, req.TenantID, instance.DocumentID, docapp.ApproverContext{
				UserID:       req.ActorUserID,
				Capabilities: req.Capabilities,
			}); err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: freeze: %w", err)
			}
			// Transition document under_review → approved.
			res, err := tx.ExecContext(ctx, `
				UPDATE documents
				   SET status           = 'approved',
				       revision_version = revision_version + 1
				 WHERE id        = $1
				   AND tenant_id = $2
				   AND status    = 'under_review'`,
				instance.DocumentID, req.TenantID,
			)
			if err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: approve document: %w", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				_ = tx.Rollback()
				return SignoffResult{}, fmt.Errorf("recordSignoff: approve document rows affected: %w", err)
			}
			if rows == 0 {
				_ = tx.Rollback()
				return SignoffResult{}, repository.ErrStaleRevision
			}
			result.InstanceApproved = true
			shouldDispatchPDF = true
			pdfTenantID = req.TenantID
			pdfRevisionID = instance.DocumentID
		} else {
			// Activate the next stage that AdvanceStage marked active.
			nextStage := instance.Active()
			if nextStage != nil {
				if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, nextStage.ID, domain.StageActive, domain.StagePending); err != nil {
					_ = tx.Rollback()
					return SignoffResult{}, fmt.Errorf("recordSignoff: activate next stage: %w", err)
				}
			}
		}

	case domain.QuorumRejectedStage:
		// Reject path — mark stage and instance rejected.
		if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageRejectedHere, domain.StageActive); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: reject stage: %w", err)
		}
		if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
			domain.InstanceRejected, domain.InstanceInProgress, &now); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: reject instance: %w", err)
		}
		result.InstanceRejected = true

		// SET LOCAL cancel GUC authorises under_review -> draft transition in trigger.
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.cancel_in_progress', $1, true)`,
			instance.ID,
		); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: set cancel GUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, err
		}

		// Transition document under_review -> draft so the author can edit and resubmit.
		res, err := tx.ExecContext(ctx, `
        UPDATE documents
           SET status           = 'draft',
               revision_version = revision_version + 1
         WHERE id        = $1
           AND tenant_id = $2
           AND status    = 'under_review'`,
			instance.DocumentID, req.TenantID,
		)
		if err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: reject document: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: reject document rows affected: %w", err)
		}
		if rows == 0 {
			_ = tx.Rollback()
			return SignoffResult{}, repository.ErrStaleRevision
		}

	default:
		// QuorumPending — no stage transition needed.
	}

	// Step 12: emit governance event.
	payloadMap := map[string]any{
		"instance_id":       req.InstanceID,
		"stage_instance_id": activeStage.ID,
		"decision":          req.Decision,
		"content_hash":      contentHash,
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: marshal event payload: %w", err)
	}
	event := GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    "signoff_recorded",
		ActorUserID:  req.ActorUserID,
		ResourceType: "approval_instance",
		ResourceID:   req.InstanceID,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		_ = tx.Rollback()
		return SignoffResult{}, fmt.Errorf("recordSignoff: emit event: %w", err)
	}

	// Step 13: enqueue PDF dispatch inside tx (transactional outbox).
	if shouldDispatchPDF && s.pdfOutbox != nil {
		if err := s.pdfOutbox.Enqueue(ctx, tx, pdfTenantID, pdfRevisionID, []byte(contentHash)); err != nil {
			_ = tx.Rollback()
			return SignoffResult{}, fmt.Errorf("recordSignoff: enqueue pdf outbox: %w", err)
		}
	}

	// Step 14: commit.
	if err := tx.Commit(); err != nil {
		return SignoffResult{}, fmt.Errorf("recordSignoff: commit: %w", err)
	}
	// deprecated: post-commit best-effort dispatch (replaced by pdfOutbox transactional enqueue above).
	// Left in place for callers that have not yet wired pdfOutbox.
	if shouldDispatchPDF && s.pdfOutbox == nil && s.pdfDispatcher != nil {
		// Client fails fast; retry owned by PDFOutboxWorker (wiki/decisions/0009-pdf-dispatch-outbox.md).
		dispatchCtx, dispatchCancel := context.WithTimeout(ctx, 45*time.Second)
		defer dispatchCancel()
		if err := s.pdfDispatcher.Dispatch(dispatchCtx, pdfTenantID, pdfRevisionID); err != nil {
			slog.ErrorContext(dispatchCtx, "pdf dispatch failed", "tenantID", pdfTenantID, "revisionID", pdfRevisionID, "err", err)
		}
	}
	return result, nil
}

func (s *DecisionService) emitEligibilityRejection(ctx context.Context, db *sql.DB, tenantID, actorID string, event GovernanceEvent) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		return fmt.Errorf("emit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// loadPriorSignoffs fetches all signoffs for the instance EXCEPT the active stage,
// used for SoD checking (actor must not have signed in any prior stage).
func (s *DecisionService) loadPriorSignoffs(ctx context.Context, tx *sql.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id,
		       actor_user_id, actor_tenant_id, decision,
		       comment, signed_at, signature_method, signature_payload, content_hash
		FROM approval_signoffs
		WHERE approval_instance_id = $1
		  AND stage_instance_id != $2
		  AND actor_tenant_id = $3
		ORDER BY signed_at ASC`,
		instanceID, activeStageID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSignoffs(rows)
}

// loadStageSignoffs fetches all signoffs for a single stage instance.
func (s *DecisionService) loadStageSignoffs(ctx context.Context, tx *sql.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id,
		       s.actor_user_id, s.actor_tenant_id, s.decision,
		       s.comment, s.signed_at, s.signature_method, s.signature_payload, s.content_hash
		  FROM approval_signoffs s
		  JOIN approval_stage_instances asi
		    ON asi.id = s.stage_instance_id
		  JOIN approval_instances ai
		    ON ai.id = asi.approval_instance_id
		   AND ai.id = s.approval_instance_id
		 WHERE s.stage_instance_id = $1
		   AND ai.tenant_id = $2::uuid
		   AND s.actor_tenant_id = ai.tenant_id
		 ORDER BY s.signed_at ASC`,
		stageInstanceID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSignoffs(rows)
}

// scanSignoffs reads rows into domain.Signoff slice.
func scanSignoffs(rows *sql.Rows) ([]domain.Signoff, error) {
	var signoffs []domain.Signoff
	for rows.Next() {
		var (
			id                 string
			approvalInstanceID string
			stageInstanceID    string
			actorUserID        string
			actorTenantID      string
			decision           string
			comment            string
			signedAt           time.Time
			signatureMethod    string
			signaturePayload   []byte
			contentHash        string
		)
		if err := rows.Scan(
			&id, &approvalInstanceID, &stageInstanceID,
			&actorUserID, &actorTenantID, &decision,
			&comment, &signedAt, &signatureMethod, &signaturePayload, &contentHash,
		); err != nil {
			return nil, err
		}
		s, err := domain.NewSignoff(domain.SignoffParams{
			ID:                 id,
			ApprovalInstanceID: approvalInstanceID,
			StageInstanceID:    stageInstanceID,
			ActorUserID:        actorUserID,
			ActorTenantID:      actorTenantID,
			Decision:           domain.Decision(decision),
			Comment:            comment,
			SignedAt:           signedAt,
			SignatureMethod:    signatureMethod,
			SignaturePayload:   json.RawMessage(signaturePayload),
			ContentHash:        contentHash,
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s: %w", id, err)
		}
		signoffs = append(signoffs, *s)
	}
	return signoffs, rows.Err()
}

// splitSignoffs partitions a slice of Signoff into approvals and rejections.
func splitSignoffs(all []domain.Signoff) (approvals, rejections []domain.Signoff) {
	for _, s := range all {
		switch s.Decision() {
		case domain.DecisionApprove:
			approvals = append(approvals, s)
		case domain.DecisionReject:
			rejections = append(rejections, s)
		}
	}
	return
}

// marshalSignaturePayload converts the map to json.RawMessage.
// Returns an empty JSON object for a nil/empty map.
func marshalSignaturePayload(payload map[string]any) (json.RawMessage, error) {
	if len(payload) == 0 {
		return json.RawMessage("{}"), nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func hasUnresolvedComments(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (bool, error) {
	var unresolvedCount int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL`,
		tenantID, documentID,
	).Scan(&unresolvedCount)
	if err != nil {
		return false, err
	}
	return unresolvedCount > 0, nil
}

func loadCurrentDocumentFormData(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (map[string]any, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(form_data_json, '{}'::jsonb)
		  FROM documents
		 WHERE id = $1
		   AND tenant_id = $2
		 FOR UPDATE`,
		documentID, tenantID,
	).Scan(&raw)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var formData map[string]any
	if err := json.Unmarshal(raw, &formData); err != nil {
		return nil, fmt.Errorf("unmarshal form_data_json: %w", err)
	}
	if formData == nil {
		return map[string]any{}, nil
	}
	return formData, nil
}
