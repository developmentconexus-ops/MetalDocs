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

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ErrApprovalBlockedByUnresolvedComments is returned when an approve decision is
// attempted while the document still has unresolved review comments.
var ErrApprovalBlockedByUnresolvedComments = errors.New("approval: unresolved comments block approval")

// ErrReauthNotConfigured is returned (fail-closed) when a password_reauth
// sign-off reaches RecordSignoff without a signature verifier wired. It must
// never be possible to record a password_reauth sign-off without verification.
var ErrReauthNotConfigured = errors.New("approval: signature verifier not configured")

// signatureMethodPasswordReauth is the only signature method that carries an
// e-signature re-authentication control (21 CFR Part 11). It is set server-side
// by the sign-off handlers, never by the client.
const signatureMethodPasswordReauth = "password_reauth"

// PinInvoker is the async-freeze seam (ADR 0015).
// Pin performs validation, resolves computed placeholders, writes values_hash +
// frozen_at, and enqueues a materialize_dispatch_outbox row — all inside tx.
// No network calls to docx-renderer.
type PinInvoker interface {
	Pin(ctx context.Context, tx db.Tx, tenantID, revisionID string, approver docapp.ApproverContext) error
}

// pdfDispatchEnqueuer is the minimal published interface for the staging
// pdf dispatch Enqueuer (render/fanout/dispatchjobs), owned here (the
// consumer) and satisfied by *dispatchjobs.Enqueuer. It inserts the paired
// (outbox row, River job) atomically inside tx (M5 F5.3 T3).
type pdfDispatchEnqueuer interface {
	EnqueuePDFTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte) error
}

// DecisionService handles approver approve/reject decisions.
type DecisionService struct {
	repo       infrastructure.ApprovalRepository
	emitter    EventEmitter
	clock      Clock
	pinInvoker  PinInvoker
	pdfDispatch pdfDispatchEnqueuer
	// sigRegistry verifies the e-signature credential before a sign-off is
	// recorded. nil only in tests that exercise non-reauth methods.
	sigRegistry       *signature.Registry
	cdRead            controlleddocumentsdomain.CDFieldReader
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
}

// NewDecisionService builds the approve/reject decision service. The async-freeze
// seam is wired separately via WithPinInvoker (ADR 0015); RecordSignoff requires
// a PinInvoker to be set before the approval-quorum path runs.
func NewDecisionService(
	repo infrastructure.ApprovalRepository,
	emitter EventEmitter,
	clock Clock,
) *DecisionService {
	return &DecisionService{
		repo:    repo,
		emitter: emitter,
		clock:   clock,
	}
}

// WithPDFOutbox sets the transactional staging dispatch Enqueuer, replacing
// the post-commit dispatcher. Takes the narrow pdfDispatchEnqueuer interface,
// satisfied by *dispatchjobs.Enqueuer (M5 F5.3 T3).
func (s *DecisionService) WithPDFOutbox(enqueuer pdfDispatchEnqueuer) *DecisionService {
	s.pdfDispatch = enqueuer
	return s
}

// WithCDFieldReader wires the controlleddocuments read-port used to resolve a
// document's controlled-document area in the area-grade authz checks (M2/F2.1).
func (s *DecisionService) WithCDFieldReader(r controlleddocumentsdomain.CDFieldReader) *DecisionService {
	s.cdRead = r
	return s
}

// WithPinInvoker wires the async-freeze path (ADR 0015): during signoff Pin
// records the frozen pointer in-tx and the heavy materialize/PDF work is
// dispatched via the outbox afterward, eliminating any in-tx network call.
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

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer used to publish
// lifecycle events after a decision completes a stage or instance transition.
func (s *DecisionService) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *DecisionService {
	s.lifecycleEnqueuer = e
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

		// Step 4: load approval instance; child stage rows locked FOR UPDATE inside LoadInstance (J1).
		instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("recordSignoff: %w", infrastructure.ErrNoActiveInstance)
			}
			return fmt.Errorf("recordSignoff: load instance: %w", err)
		}
		if instance == nil {
			return infrastructure.ErrNoActiveInstance
		}
		if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
			return infrastructure.ErrStaleRevision
		}

		// Reject if instance is already terminal.
		if instance.Status != domain.InstanceInProgress {
			return infrastructure.ErrInstanceCompleted
		}

		// document.signoff is area-grade: pass the resolved area as-is ("" fail-closes).
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.DocumentID)
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
		// No-fallback (F6, spec §11): the ONLY authoritative source is the instance's
		// frozen_content_hash, pinned once at the freeze boundary (F5). By the time an
		// approval-kind stage is active and signoff is possible, the instance must
		// already be frozen — a NULL pin here is an impossible state, not a legitimate
		// "not yet computed" case, so it fails closed via ErrNoActiveContentHash rather
		// than falling back to any document-table or revision-history hash.
		contentHash, err := s.repo.LoadFrozenContentHash(ctx, tx, req.TenantID, instance.ID)
		if err != nil {
			if errors.Is(err, infrastructure.ErrNoActiveContentHash) {
				return ErrContentHashMismatch
			}
			return fmt.Errorf("recordSignoff: load frozen content hash: %w", err)
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
			return infrastructure.ErrStageNotActive
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
			if errors.Is(err, infrastructure.ErrActorAlreadySigned) {
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
				// Note (F5, W10): the unresolved-comments gate that used to run
				// here was removed. By construction (plan.md "no new call site"
				// finding) an approval-kind stage only ever activates after freeze
				// has already fired — from ReviewVerdictService.RecordVerdict's
				// stage-advance path (review->approval transitions) or from
				// SubmitService.SubmitRevisionForReview (approval-only routes).
				// Freeze's own instance-scoped comment check
				// (ErrFreezeBlockedByUnresolvedComments) is now the sole gate for
				// this concern; decision_service.go never needs its own copy.

				// All stages done — complete instance.
				if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
					domain.InstanceApproved, domain.InstanceInProgress, &now); err != nil {
					return fmt.Errorf("recordSignoff: complete instance: %w", err)
				}
				if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
					return err
				}
				if s.pinInvoker == nil {
					return fmt.Errorf("recordSignoff: pinInvoker not configured")
				}
				if err := s.pinInvoker.Pin(ctx, tx, req.TenantID, instance.DocumentID, docapp.ApproverContext{
					UserID:       req.ActorUserID,
					Capabilities: req.Capabilities,
				}); err != nil {
					return fmt.Errorf("recordSignoff: pin: %w", err)
				}
				// Transition document under_review → approved. Friendly first-line
				// legality check (M4/F4.1) mirrors the DB trigger; the OCC WHERE
				// below remains the atomic CAS + optimistic-lock enforcement.
				if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusApproved); err != nil {
					return err
				}
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
					return infrastructure.ErrStaleRevision
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

			// Transition document under_review -> draft so the author can edit and
			// resubmit. Friendly first-line legality check (M4/F4.1) mirrors the DB
			// trigger; the OCC WHERE below remains the atomic CAS + optimistic-lock
			// enforcement (the DB trigger additionally gates this specific arc on the
			// metaldocs.cancel_in_progress GUC set above).
			if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusDraft); err != nil {
				return err
			}
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
				return infrastructure.ErrStaleRevision
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

		// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Author events — terminal transitions only.
		if s.lifecycleEnqueuer != nil {
			var lifecycleEventType string
			if result.InstanceApproved {
				lifecycleEventType = docsdomain.EventTypeDocumentApproved
			} else if result.InstanceRejected {
				lifecycleEventType = docsdomain.EventTypeDocumentRejected
			}
			if lifecycleEventType != "" {
				largs := docsdomain.LifecycleEventArgs{
					EventID:      uuid.NewString(),
					TenantID:     req.TenantID,
					EventType:    lifecycleEventType,
					ResourceType: "approval_instance",
					ResourceID:   req.InstanceID,
					SubmittedBy:  instance.SubmittedBy,
					OccurredAt:   now,
				}
				if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
					return fmt.Errorf("recordSignoff: enqueue lifecycle event: %w", err)
				}
			}
		}

		// Step 13: enqueue PDF dispatch inside tx (transactional outbox).
		// Skipped when pinInvoker is active: PDF dispatch is enqueued by MaterializeJobRunner
		// after the fanout call succeeds (ADR 0015).
		if shouldDispatchPDF && s.pdfDispatch != nil && s.pinInvoker == nil {
			if err := s.pdfDispatch.EnqueuePDFTx(ctx, tx, pdfTenantID, pdfRevisionID, []byte(contentHash)); err != nil {
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
