package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

var ErrApprovalBlockedByUnresolvedComments = errors.New("approval: unresolved comments block approval")

// ErrReauthNotConfigured is returned (fail-closed) when a password_reauth
// sign-off reaches RecordSignoff without a signature verifier wired. It must
// never be possible to record a password_reauth sign-off without verification.
var ErrReauthNotConfigured = errors.New("approval: signature verifier not configured")

// signatureMethodPasswordReauth is the only signature method that carries an
// e-signature re-authentication control (21 CFR Part 11). It is set server-side
// by the sign-off handlers, never by the client.
const signatureMethodPasswordReauth = "password_reauth"

type FreezeInvoker interface {
	Freeze(ctx context.Context, tx db.Tx, tenantID, revisionID string, approver docapp.ApproverContext) error
}

// PinInvoker is the async-freeze replacement for FreezeInvoker (ADR 0015).
// Pin performs validation, resolves computed placeholders, writes values_hash +
// frozen_at, and enqueues a materialize_dispatch_outbox row — all inside tx.
// No network calls to docx-renderer.
type PinInvoker interface {
	Pin(ctx context.Context, tx db.Tx, tenantID, revisionID string, approver docapp.ApproverContext) error
}

// PDFOutboxEnqueuer enqueues a PDF dispatch inside the approval transaction.
type PDFOutboxEnqueuer interface {
	Enqueue(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte) error
}

// DecisionService handles approver approve/reject decisions.
type DecisionService struct {
	repo          repository.ApprovalRepository
	emitter       EventEmitter
	clock         Clock
	freezeInvoker FreezeInvoker
	pinInvoker    PinInvoker
	pdfOutbox     PDFOutboxEnqueuer
	// sigRegistry verifies the e-signature credential before a sign-off is
	// recorded. nil only in tests that exercise non-reauth methods.
	sigRegistry *signature.Registry
}

func NewDecisionService(
	repo repository.ApprovalRepository,
	emitter EventEmitter,
	clock Clock,
	freezeInvoker FreezeInvoker,
) *DecisionService {
	return &DecisionService{
		repo:          repo,
		emitter:       emitter,
		clock:         clock,
		freezeInvoker: freezeInvoker,
	}
}

// WithPDFOutbox sets the transactional outbox enqueuer, replacing the post-commit dispatcher.
func (s *DecisionService) WithPDFOutbox(enqueuer PDFOutboxEnqueuer) *DecisionService {
	s.pdfOutbox = enqueuer
	return s
}

// WithPinInvoker enables the async-freeze path (ADR 0015). When set, Pin is
// called instead of Freeze during signoff, eliminating the in-tx network call.
func (s *DecisionService) WithPinInvoker(invoker PinInvoker) *DecisionService {
	s.pinInvoker = invoker
	return s
}

// WithSignatureRegistry wires the e-signature verifier used to re-authenticate
// the acting user before a password_reauth sign-off is recorded.
func (s *DecisionService) WithSignatureRegistry(registry *signature.Registry) *DecisionService {
	s.sigRegistry = registry
	return s
}

// resolveSignaturePayload re-authenticates the acting user and returns the
// signature payload to persist. For password_reauth it verifies password_token
// against the stored bcrypt hash via the signature Registry and returns the
// verified attestation (no secret), so the raw credential is never stored. For
// any other (legacy/test) method it marshals the supplied payload unchanged.
func (s *DecisionService) resolveSignaturePayload(ctx context.Context, req SignoffRequest) (json.RawMessage, error) {
	if req.SignatureMethod != signatureMethodPasswordReauth {
		return marshalSignaturePayload(req.SignaturePayload)
	}
	if s.sigRegistry == nil {
		return nil, ErrReauthNotConfigured
	}
	provider, err := s.sigRegistry.Get(req.SignatureMethod)
	if err != nil {
		return nil, err
	}
	token, _ := req.SignaturePayload["password_token"].(string)
	result, err := provider.Sign(ctx, signature.SignRequest{
		ActorUserID:   req.ActorUserID,
		ActorTenantID: req.TenantID,
		Credentials:   map[string]string{"password": token},
	})
	if err != nil {
		return nil, err
	}
	return result.Payload, nil
}

// SignoffRequest carries all inputs for RecordSignoff.
type SignoffRequest struct {
	TenantID                string
	InstanceID              string
	StageInstanceID         string
	ActorUserID             string
	Decision                domain.Decision
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
func (s *DecisionService) RecordSignoff(ctx context.Context, runner db.TxRunner, req SignoffRequest) (SignoffResult, error) {
	ctx, span := otel.Tracer("metaldocs/documents/approval").Start(ctx, "signoff.record",
		oteltrace.WithAttributes(attribute.String("signoff.verdict", string(req.Decision))),
	)
	defer span.End()

	// Step 1: validate signature payload — no float64 values.
	if err := ValidateEventPayload(req.SignaturePayload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return SignoffResult{}, err
	}

	var result SignoffResult
	var eligibilityEvent *GovernanceEvent
	// H-PRE-1: resolve the actor display-name snapshot OFF the signoff transaction.
	// This is a cross-module read of metaldocs.iam_users; running it inside the
	// lock-holding signoff tx (advisory lock + FOR UPDATE stage rows) on a fresh
	// connection risks deadlock. req.TenantID/req.ActorUserID are server-derived and
	// available pre-flight. Contained on ApprovalRepository (not a shared port — M4/F4.1).
	actorDisplayName, err := s.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)
	if err != nil {
		err = fmt.Errorf("recordSignoff: lookup actor display name: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return SignoffResult{}, err
	}
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.SeedTxIdentity(ctx, tx, req.TenantID, req.ActorUserID); err != nil {
			return fmt.Errorf("recordSignoff: %w", err)
		}

		// Step 4: load approval instance; child stage rows locked FOR UPDATE inside LoadInstance (J1).
		instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("recordSignoff: %w", repository.ErrNoActiveInstance)
			}
			return fmt.Errorf("recordSignoff: load instance: %w", err)
		}
		if instance == nil {
			return repository.ErrNoActiveInstance
		}
		if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
			return repository.ErrStaleRevision
		}

		// Reject if instance is already terminal.
		if instance.Status != domain.InstanceInProgress {
			return repository.ErrInstanceCompleted
		}

		// document.signoff is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, req.TenantID, instance.DocumentID)
		if err != nil {
			return fmt.Errorf("recordSignoff: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSignoff), areaCode); err != nil {
			return err
		}

		// 21 CFR Part 11 e-signature: re-authenticate the acting user BEFORE the
		// sign-off is recorded. resolveSignaturePayload verifies password_token
		// against the stored bcrypt hash and returns the attestation that gets
		// persisted (raw credential never stored). A bad/blank token fails here and
		// no sign-off is recorded.
		sigPayload, err := s.resolveSignaturePayload(ctx, req)
		if err != nil {
			return err
		}

		// Content pin: client echoes back the content hash from the active-document
		// endpoint to confirm the instance content has not drifted since they loaded it.
		// Resolve the same value here (documents.content_hash_at_submit, falling back
		// to the latest document_revisions.content_hash) so client and server agree on
		// the canonicalization. The approval_instance's content_hash_at_submit is not
		// the right source — submit canonicalizes over the client-provided hash, which
		// is irreproducible at signoff time.
		contentHash, err := s.repo.LoadActiveDocumentContentHash(ctx, tx, req.TenantID, instance.DocumentID)
		if err != nil {
			if errors.Is(err, repository.ErrNoActiveContentHash) {
				return ErrContentHashMismatch
			}
			return fmt.Errorf("recordSignoff: load active document content hash: %w", err)
		}
		// Content pin is mandatory: an unauthenticated or programmatic caller must not
		// be able to skip the check by omitting `_content_hash`. The HTTP boundary
		// already enforces a 64-hex hash, so this is a defense-in-depth guard.
		clientHash, ok := clientContentHash(req.ContentFormData)
		if !ok {
			return ErrContentHashMismatch
		}
		if clientHash != contentHash {
			return ErrContentHashMismatch
		}

		// Step 5: identify active stage.
		activeStage := instance.Active()
		if activeStage == nil {
			return domain.ErrNoActiveStage
		}
		// Ensure the requested StageInstanceID matches the active stage.
		if req.StageInstanceID == "" || activeStage.ID != req.StageInstanceID {
			return repository.ErrStageNotActive
		}

		// Step 5b: eligibility check — actor must be in the eligible_actor_ids snapshot (J1).
		if err := domain.CheckEligibility(req.ActorUserID, activeStage.EligibleActorIDs); err != nil {
			event := GovernanceEvent{
				TenantID:     req.TenantID,
				EventType:    EventTypeSignoffRejected,
				ActorUserID:  req.ActorUserID,
				ResourceType: "approval_instance",
				ResourceID:   req.InstanceID,
				Reason:       "not_eligible",
				OccurredAt:   s.clock.Now(),
			}
			eligibilityEvent = &event
			return err
		}

		// Step 6: SoD check — author cannot sign, actor cannot sign twice in same instance.
		priorSignoffs, err := s.repo.LoadPriorSignoffs(ctx, tx, req.TenantID, req.InstanceID, activeStage.ID)
		if err != nil {
			return fmt.Errorf("recordSignoff: load prior signoffs: %w", err)
		}
		if err := domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, priorSignoffs); err != nil {
			return err
		}

		// Step 7: build the domain Signoff value object. sigPayload was resolved
		// (and the actor re-authenticated) above.
		now := s.clock.Now()
		signoff, err := domain.NewSignoff(domain.SignoffParams{
			ID:                       uuid.New().String(),
			ApprovalInstanceID:       req.InstanceID,
			StageInstanceID:          activeStage.ID,
			ActorUserID:              req.ActorUserID,
			ActorTenantID:            req.TenantID,
			Decision:                 req.Decision,
			Comment:                  req.Comment,
			SignedAt:                 now,
			SignatureMethod:          req.SignatureMethod,
			SignaturePayload:         sigPayload,
			ContentHash:              contentHash,
			ActorDisplayNameSnapshot: actorDisplayName,
		})
		if err != nil {
			return fmt.Errorf("recordSignoff: build signoff: %w", err)
		}

		// Step 8: persist the signoff, handling idempotent replay.
		insertResult, err := s.repo.InsertSignoff(ctx, tx, *signoff)
		if err != nil {
			if errors.Is(err, repository.ErrActorAlreadySigned) {
				return err
			}
			return fmt.Errorf("recordSignoff: insert signoff: %w", err)
		}
		if insertResult.WasReplay {
			// Idempotent replay: commit and return neutral result (stage not advanced again).
			result = SignoffResult{}
			return nil
		}

		// Step 9: collect all signoffs for the active stage to evaluate quorum.
		allStageSignoffs, err := s.repo.LoadStageSignoffs(ctx, tx, req.TenantID, activeStage.ID)
		if err != nil {
			return fmt.Errorf("recordSignoff: load stage signoffs: %w", err)
		}

		// Step 10: evaluate quorum.
		approvals, rejections := splitSignoffs(allStageSignoffs)
		currentEligible := activeStage.EligibleActorIDs
		if activeStage.OnEligibilityDriftSnapshot != domain.DriftKeepSnapshot {
			currentEligible, err = s.repo.ResolveEligibleActors(ctx, tx, req.TenantID, activeStage.AreaCodeSnapshot, activeStage.RequiredRoleSnapshot)
			if err != nil {
				return fmt.Errorf("recordSignoff: resolve current eligible actors: %w", err)
			}
		}
		drift := domain.ApplyEligibilityDrift(*activeStage, currentEligible)
		effectiveDenominator := drift.EffectiveDenominator
		outcome := drift.ForcedOutcome
		if outcome == domain.QuorumPending {
			outcome = domain.EvaluateQuorum(*activeStage, approvals, rejections, effectiveDenominator)
			if outcome == domain.QuorumError {
				return domain.ErrEmptyEligiblePool
			}
		}

		var shouldDispatchPDF bool
		var pdfTenantID string
		var pdfRevisionID string

		switch outcome {
		case domain.QuorumApprovedStage:
			// Step 11a: mark stage completed.
			if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageCompleted, domain.StageActive); err != nil {
				return fmt.Errorf("recordSignoff: complete stage: %w", err)
			}
			result.StageCompleted = true

			// Advance the in-memory instance to determine next step.
			if err := instance.AdvanceStage(); err != nil {
				return fmt.Errorf("recordSignoff: advance stage: %w", err)
			}

			if instance.Status == domain.InstanceApproved {
				blocked, err := s.repo.HasUnresolvedComments(ctx, tx, req.TenantID, instance.DocumentID)
				if err != nil {
					return fmt.Errorf("recordSignoff: check unresolved comments: %w", err)
				}
				if blocked {
					return ErrApprovalBlockedByUnresolvedComments
				}

				// All stages done — complete instance.
				if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
					domain.InstanceApproved, domain.InstanceInProgress, &now); err != nil {
					return fmt.Errorf("recordSignoff: complete instance: %w", err)
				}
				if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
					return err
				}
				if s.pinInvoker != nil {
					if err := s.pinInvoker.Pin(ctx, tx, req.TenantID, instance.DocumentID, docapp.ApproverContext{
						UserID:       req.ActorUserID,
						Capabilities: req.Capabilities,
					}); err != nil {
						return fmt.Errorf("recordSignoff: pin: %w", err)
					}
				} else {
					if s.freezeInvoker == nil {
						return fmt.Errorf("recordSignoff: neither pinInvoker nor freezeInvoker configured")
					}
					if err := s.freezeInvoker.Freeze(ctx, tx, req.TenantID, instance.DocumentID, docapp.ApproverContext{
						UserID:       req.ActorUserID,
						Capabilities: req.Capabilities,
					}); err != nil {
						return fmt.Errorf("recordSignoff: freeze: %w", err)
					}
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
					return fmt.Errorf("recordSignoff: approve document: %w", err)
				}
				rows, err := res.RowsAffected()
				if err != nil {
					return fmt.Errorf("recordSignoff: approve document rows affected: %w", err)
				}
				if rows == 0 {
					return repository.ErrStaleRevision
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
						return fmt.Errorf("recordSignoff: activate next stage: %w", err)
					}
				}
			}

		case domain.QuorumRejectedStage:
			// Reject path — mark stage and instance rejected.
			if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageRejectedHere, domain.StageActive); err != nil {
				return fmt.Errorf("recordSignoff: reject stage: %w", err)
			}
			if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
				domain.InstanceRejected, domain.InstanceInProgress, &now); err != nil {
				return fmt.Errorf("recordSignoff: reject instance: %w", err)
			}
			result.InstanceRejected = true

			// SET LOCAL cancel GUC authorises under_review -> draft transition in trigger.
			if _, err := tx.ExecContext(ctx,
				`SELECT set_config('metaldocs.cancel_in_progress', $1, true)`,
				instance.ID,
			); err != nil {
				return fmt.Errorf("recordSignoff: set cancel GUC: %w", err)
			}
			if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
				return err
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
				return fmt.Errorf("recordSignoff: reject document: %w", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return fmt.Errorf("recordSignoff: reject document rows affected: %w", err)
			}
			if rows == 0 {
				return repository.ErrStaleRevision
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
			return fmt.Errorf("recordSignoff: marshal event payload: %w", err)
		}
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeSignoffRecorded,
			ActorUserID:  req.ActorUserID,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("recordSignoff: emit event: %w", err)
		}

		// Step 13: enqueue PDF dispatch inside tx (transactional outbox).
		// Skipped when pinInvoker is active: PDF dispatch is enqueued by MaterializeJobRunner
		// after the fanout call succeeds (ADR 0015).
		if shouldDispatchPDF && s.pdfOutbox != nil && s.pinInvoker == nil {
			if err := s.pdfOutbox.Enqueue(ctx, tx, pdfTenantID, pdfRevisionID, []byte(contentHash)); err != nil {
				return fmt.Errorf("recordSignoff: enqueue pdf outbox: %w", err)
			}
		}

		return nil
	})
	if eligibilityEvent != nil {
		if emitErr := s.emitEligibilityRejection(ctx, runner, req.TenantID, req.ActorUserID, *eligibilityEvent); emitErr != nil {
			wrappedErr := fmt.Errorf("recordSignoff: emit eligibility rejection: %w", emitErr)
			span.RecordError(wrappedErr)
			span.SetStatus(codes.Error, wrappedErr.Error())
			return SignoffResult{}, wrappedErr
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return SignoffResult{}, err
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return SignoffResult{}, err
	}
	return result, nil
}

func (s *DecisionService) emitEligibilityRejection(ctx context.Context, runner db.TxRunner, tenantID, actorID string, event GovernanceEvent) error {
	return runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
			return err
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}
		return nil
	})
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

func clientContentHash(formData map[string]any) (string, bool) {
	if len(formData) == 0 {
		return "", false
	}
	raw, ok := formData["_content_hash"]
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}
