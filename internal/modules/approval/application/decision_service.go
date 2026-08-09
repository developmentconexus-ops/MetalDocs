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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/approval/infrastructure/signature"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
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
	Pin(ctx context.Context, tx db.Tx, tenantID, revisionID string, approver docapp.ApproverContext, releaseGenerationID string) error
}

// TemplateCompletionWriter is the approval-owned completion port for a
// template-subject instance reaching a terminal outcome (M3
// P3.S2b-3b-iii-b). Mirrors the PinInvoker / TemplateVersionReader
// injection pattern: approval never imports templates infrastructure or
// writes templates_template_version directly — this is the ONLY seam a
// terminal approve/reject decision crosses the module boundary through. The
// production adapter lives in templates/infrastructure
// (ApprovalCompletionWriter) and runs its UPDATE on the CALLER's
// transaction (tx) so the version transition commits atomically with the
// signoff write. Fail-closed: a template instance reaching a terminal
// outcome with this port unset is a wiring error surfaced as an explicit
// error, never a silent skip of the version transition (mirrors the
// pinInvoker == nil guard on the document path).
// approverUserID (F-E4-4) is the deciding actor whose sign-off carried the
// instance to terminal approval — SignoffRequest.ActorUserID, the same
// server-derived identity the approval_signoffs ledger row records. It is
// stamped onto templates_template_version.approver_id so the version row
// carries approver attribution alongside approved_at, mirroring
// TerminalApprovalInput.FinalApproverID on the document arm. Never optional
// and never substituted: an empty value is a wiring fault, not a value to
// invent (no-fallback principle).
type TemplateCompletionWriter interface {
	MarkTemplateVersionApproved(ctx context.Context, tx db.Tx, tenantID, templateVersionID, approverUserID string) error
	MarkTemplateVersionRejected(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error
}

// DecisionService handles approver approve/reject decisions.
type DecisionService struct {
	repo    infrastructure.ApprovalRepository
	emitter EventEmitter
	clock   Clock
	// sigRegistry verifies the e-signature credential before a sign-off is
	// recorded. nil only in tests that exercise non-reauth methods.
	sigRegistry       *signature.Registry
	cdRead            controlleddocumentsdomain.CDFieldReader
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
	// templateCompletion is the M3 P3.S2b-3b-iii-b completion port: on a
	// terminal decision for a subject_kind='template' instance, drives the
	// templates_template_version status transition in place of the
	// document-only `UPDATE documents` path. nil in tests that never
	// exercise a template-subject instance.
	templateCompletion TemplateCompletionWriter
	// templateVersionReader is the same approval-owned port
	// TemplateSubmitService uses (#24), reused here to read a template
	// version's real content_hash in place of the document-only freeze pin.
	templateVersionReader TemplateVersionReader
	// releaseRecorder is the ADR 0085 terminal-approval seam (approval fact +
	// async-freeze pin + coordinator evaluation, all on the caller's tx).
	// Mandatory on the document terminal-approval path.
	releaseRecorder TerminalApprovalReleaseRecorder
}

// NewDecisionService builds the approve/reject decision service. The
// terminal-approval seam is wired separately via WithReleaseRecorder (ADR 0015
// pin + ADR 0085 release facts); RecordSignoff requires it to be set before the
// approval-quorum path runs.
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

// WithCDFieldReader wires the controlleddocuments read-port used to resolve a
// document's controlled-document area in the area-grade authz checks (M2/F2.1).
func (s *DecisionService) WithCDFieldReader(r controlleddocumentsdomain.CDFieldReader) *DecisionService {
	s.cdRead = r
	return s
}

// WithReleaseRecorder wires the ADR 0085 terminal-approval seam, which
// subsumes the ADR 0015 async-freeze path: during signoff the recorder stamps
// the approval fact, pins the frozen pointer in-tx and enqueues the
// coordinator evaluation; the heavy materialize/PDF work is dispatched via the
// outbox afterward, eliminating any in-tx network call.
func (s *DecisionService) WithReleaseRecorder(recorder TerminalApprovalReleaseRecorder) *DecisionService {
	s.releaseRecorder = recorder
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

// WithTemplateCompletionWriter wires the M3 P3.S2b-3b-iii-b completion port
// used to transition templates_template_version on a terminal
// template-subject decision. Call after NewServices/NewDecisionService. A
// nil writer fails closed (an explicit error from recordSignoffInTx) rather
// than silently skipping the version transition.
func (s *DecisionService) WithTemplateCompletionWriter(writer TemplateCompletionWriter) *DecisionService {
	s.templateCompletion = writer
	return s
}

// WithTemplateVersionReader wires the approval-owned TemplateVersionReader
// port (shared with TemplateSubmitService, #24) so a template-subject
// signoff can read the version's real content_hash in place of the
// document-only freeze pin (M3 P3.S2b-3b-iii-b).
func (s *DecisionService) WithTemplateVersionReader(reader TemplateVersionReader) *DecisionService {
	s.templateVersionReader = reader
	return s
}

// Ready reports whether every port required for full runtime function is
// wired: templateVersionReader, templateCompletion, pinInvoker, sigRegistry,
// and cdRead. Their absence causes a 500 (nil pointer / not found) or a silent
// skip of a state transition (e.g. the templates_template_version write).
// lifecycleEnqueuer is excluded — it is best-effort
// and nil-tolerant (see WithLifecycleEnqueuer), guarded at every call site.
// A composition root should call Ready before serving traffic so a wiring
// regression (e.g. rebuilding Decision instead of mutating it in place) fails
// fast at boot instead of surfacing later as a runtime error.
func (s *DecisionService) Ready() error {
	var missing []string
	if s.templateVersionReader == nil {
		missing = append(missing, "templateVersionReader")
	}
	if s.templateCompletion == nil {
		missing = append(missing, "templateCompletion")
	}
	if s.releaseRecorder == nil {
		missing = append(missing, "releaseRecorder")
	}
	if s.sigRegistry == nil {
		missing = append(missing, "sigRegistry")
	}
	if s.cdRead == nil {
		missing = append(missing, "cdRead")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("decision service missing required ports: %s", strings.Join(missing, ", "))
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
	// SignoffID is the approval_signoffs row id persisted by this call. On the
	// DB-level idempotent replay branch (ON CONFLICT) it carries the ORIGINAL
	// row id, so a retrying caller always sees the same identifier (F-QA4-7).
	SignoffID        string
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

	var result SignoffResult
	var eligibilityEvent *GovernanceEvent
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		var txErr error
		result, eligibilityEvent, txErr = s.recordSignoffInTx(ctx, tx, req, actorDisplayName)
		return txErr
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

// recordSignoffInTx is the tx-scoped core of RecordSignoff, extracted so a
// fast-forward flow (unit 2.3 G3) can run it inside a transaction it shares
// with ReviewVerdictService.recordVerdictInTx. It does not own commit/rollback
// and must not call runner itself. actorDisplayName and sigPayload resolution
// stay off-tx in RecordSignoff exactly as before (signature payload
// resolution — resolveSignaturePayload — runs in-tx today and is kept in-tx
// here). The eligibility-rejection event (when non-nil) must be emitted by the
// caller AFTER the tx this ran in has closed — mirrors RecordSignoff's
// original post-tx emitEligibilityRejection call.
func (s *DecisionService) recordSignoffInTx(ctx context.Context, tx *sql.Tx, req SignoffRequest, actorDisplayName string) (SignoffResult, *GovernanceEvent, error) {
	ctx = authz.WithCapCache(ctx)

	// Step 4: load approval instance, verify not terminal/stale, resolve
	// area + authz. Child stage rows locked FOR UPDATE inside LoadInstance (J1).
	instance, areaCode, err := s.loadSignoffInstance(ctx, tx, req)
	if err != nil {
		return SignoffResult{}, nil, err
	}

	// 21 CFR Part 11 e-signature: re-authenticate the acting user BEFORE the
	// sign-off is recorded. resolveSignaturePayload verifies password_token
	// against the stored bcrypt hash and returns the attestation that gets
	// persisted (raw credential never stored). A bad/blank token fails here and
	// no sign-off is recorded.
	sigPayload, err := s.resolveSignaturePayload(ctx, req)
	if err != nil {
		return SignoffResult{}, nil, err
	}

	// Content pin (document vs template — M3 P3.S2b-3b-iii-b).
	contentHash, err := s.resolveSignoffContentHash(ctx, tx, req, instance)
	if err != nil {
		return SignoffResult{}, nil, err
	}

	// Step 5: identify active stage.
	activeStage := instance.Active()
	if activeStage == nil {
		return SignoffResult{}, nil, domain.ErrNoActiveStage
	}
	// Ensure the requested StageInstanceID matches the active stage.
	if req.StageInstanceID == "" || activeStage.ID != req.StageInstanceID {
		return SignoffResult{}, nil, infrastructure.ErrStageNotActive
	}

	// Step 5b/6: eligibility (incl. delegation widening) + SoD.
	onBehalfOf, eligibilityRejection, err := s.checkSignoffEligibility(ctx, tx, req, instance, activeStage)
	if err != nil {
		return SignoffResult{}, eligibilityRejection, err
	}

	// Step 7: build the domain Signoff value object. sigPayload was resolved
	// (and the actor re-authenticated) above.
	//
	// SignatureMeaning (F7, W8, 21 CFR 11.50(a)(3)): the signed record must
	// state what the signature means. Derived deterministically from the
	// decision the actor already submitted — never client-writable, never
	// left to NewSignoff's empty-field default (which would silently stamp
	// every reject signoff as "approval").
	signatureMeaning := "approval"
	if req.Decision == domain.DecisionReject {
		signatureMeaning = "rejection"
	}
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
		SignatureMeaning:         signatureMeaning,
		OnBehalfOfUserID:         onBehalfOf,
	})
	if err != nil {
		return SignoffResult{}, nil, fmt.Errorf("recordSignoff: build signoff: %w", err)
	}

	// Step 8: persist the signoff, handling idempotent replay.
	var result SignoffResult
	insertResult, err := s.repo.InsertSignoff(ctx, tx, *signoff)
	if err != nil {
		if errors.Is(err, infrastructure.ErrActorAlreadySigned) {
			return SignoffResult{}, nil, err
		}
		return SignoffResult{}, nil, fmt.Errorf("recordSignoff: insert signoff: %w", err)
	}
	// F-QA4-7: carry the persisted approval_signoffs id out of the tx. On the
	// ON CONFLICT branch InsertSignoff returns the ORIGINAL row's id, so a
	// retrying caller sees the same identifier instead of an empty string.
	result.SignoffID = insertResult.ID
	if insertResult.WasReplay {
		// Idempotent replay: commit and return the otherwise-neutral result
		// (stage not advanced again), which at this point carries only SignoffID.
		return result, nil, nil
	}

	// Steps 9-11: evaluate quorum (incl. any eligibility-drift policy) and
	// apply the resulting stage/instance transition.
	if err := s.applySignoffQuorum(ctx, tx, req, instance, activeStage, areaCode, now, &result); err != nil {
		return SignoffResult{}, nil, err
	}

	// Step 12+: emit governance event and (terminal document rejection only)
	// the F3.3 domain lifecycle event.
	//
	// PDF dispatch note (F-QA2-2 / QR-C): the document-approve path always Pins
	// via the async-freeze seam (ADR 0015) — MaterializeJobRunner is the sole pdf
	// producer, enqueuing PDF dispatch (with the renderer-produced
	// final_docx_s3_key) after the fanout call succeeds. The old synchronous
	// in-tx pdf-dispatch block that used to live here was structurally dead
	// (it required pinInvoker == nil, but the approve branch above hard-requires
	// pinInvoker != nil) and was removed rather than defensively threaded.
	if err := s.emitSignoffOutcome(ctx, tx, req, instance, activeStage, contentHash, onBehalfOf, result, now); err != nil {
		return SignoffResult{}, nil, err
	}

	return result, nil, nil
}

// loadSignoffInstance loads the approval instance for a signoff, verifies it
// is neither stale (OCC) nor already terminal, and asserts the
// subject-conditional signoff capability (document.signoff vs
// template.approve — M3 P3.S2b-3b-iii-b, ADR 0083) against the resolved area
// code. Returns the loaded instance and that area code for reuse by the
// terminal-approval path.
func (s *DecisionService) loadSignoffInstance(ctx context.Context, tx *sql.Tx, req SignoffRequest) (*domain.Instance, string, error) {
	instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("recordSignoff: %w", infrastructure.ErrNoActiveInstance)
		}
		return nil, "", fmt.Errorf("recordSignoff: load instance: %w", err)
	}
	if instance == nil {
		return nil, "", infrastructure.ErrNoActiveInstance
	}
	if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
		return nil, "", infrastructure.ErrStaleRevision
	}
	// Reject if instance is already terminal.
	if instance.Status != domain.InstanceInProgress {
		return nil, "", infrastructure.ErrInstanceCompleted
	}

	// document.signoff is area-grade: pass the resolved area as-is ("" fail-closes).
	// Subject-generic resolver (M3 P3.S2b-3a): instance.Subject is hydrated from
	// the real columns (P3.S2a), so a document instance resolves identically to
	// the prior hardcoded LoadDocumentAreaCode(instance.DocumentID) call.
	areaCode, err := resolveSubjectAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.Subject)
	if err != nil {
		return nil, "", fmt.Errorf("recordSignoff: load document area: %w", err)
	}
	// Subject-conditional assertion (M3 P3.S2b-3b-iii-b, ADR 0083): the
	// approval_signoffs tripwire (migration 0300) discriminates by the
	// parent instance's subject_kind — a template-subject signoff requires
	// template.approve, a document-subject signoff requires document.signoff.
	// resolveSubjectAreaCode already returns "tenant" (area-blind) for a
	// template subject (subject_area.go), which is api-lint clean for
	// CapTemplateApprove (ScopeTenant).
	signoffCap := string(iamdomain.CapDocumentSignoff)
	if instance.Subject.Kind == domain.SubjectKindTemplate {
		signoffCap = string(iamdomain.CapTemplateApprove)
	}
	if err := authz.Require(ctx, tx, signoffCap, areaCode); err != nil {
		return nil, "", err
	}
	return instance, areaCode, nil
}

// resolveSignoffContentHash resolves the content hash to pin the signoff
// against. A template never freezes (#24's template submit deliberately
// omits freeze), so a template-subject instance has no frozen_content_hash to
// pin against and no client-echo to verify — its content_hash is instead read
// straight off templates_template_version.content_hash through the
// approval-owned TemplateVersionReader port (#24), the version's real content
// identity, locked from author edits once under_review — never a direct table
// read. A document-subject instance instead requires the caller to echo back
// the content hash from the active-document endpoint to confirm the instance
// content has not drifted since they loaded it.
//
// No-fallback (F6, spec §11): the ONLY authoritative source for a document is
// the instance's frozen_content_hash, pinned once at the freeze boundary
// (F5). By the time an approval-kind stage is active and signoff is possible,
// the instance must already be frozen — a NULL pin here is an impossible
// state, not a legitimate "not yet computed" case, so it fails closed via
// ErrNoActiveContentHash rather than falling back to any document-table or
// revision-history hash.
func (s *DecisionService) resolveSignoffContentHash(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance) (string, error) {
	if instance.Subject.Kind == domain.SubjectKindTemplate {
		if s.templateVersionReader == nil {
			return "", fmt.Errorf("recordSignoff: template version reader not configured")
		}
		hash, ok, err := s.templateVersionReader.LoadTemplateVersionContentHash(ctx, tx, req.TenantID, instance.Subject.Key)
		if err != nil {
			return "", fmt.Errorf("recordSignoff: load template version content hash: %w", err)
		}
		if !ok {
			return "", ErrContentHashMismatch
		}
		return hash, nil
	}

	contentHash, err := s.repo.LoadFrozenContentHash(ctx, tx, req.TenantID, instance.ID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrNoActiveContentHash) {
			return "", ErrContentHashMismatch
		}
		return "", fmt.Errorf("recordSignoff: load frozen content hash: %w", err)
	}
	// Content pin is mandatory: an unauthenticated or programmatic caller must not
	// be able to skip the check by omitting `_content_hash`. The HTTP boundary
	// already enforces a 64-hex hash, so this is a defense-in-depth guard.
	clientHash, ok := clientContentHash(req.ContentFormData)
	if !ok {
		return "", ErrContentHashMismatch
	}
	if clientHash != contentHash {
		return "", ErrContentHashMismatch
	}
	return contentHash, nil
}

// checkSignoffEligibility verifies the acting user (or a delegator they act
// on behalf of) is eligible to sign the active stage and satisfies SoD. It
// returns a non-nil *GovernanceEvent only on the not-eligible rejection path
// — the caller must emit it AFTER the tx this ran in has closed, mirroring
// RecordSignoff's original post-tx emitEligibilityRejection call.
func (s *DecisionService) checkSignoffEligibility(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance, activeStage *domain.StageInstance) (string, *GovernanceEvent, error) {
	// Eligibility check — actor must be in the eligible_actor_ids snapshot
	// (J1), widened by any active delegation (F9/ADR 0077):
	// domain.ResolveEligibleIdentity tries the direct membership check first
	// (unchanged fast path) and only falls back to the actor's active
	// delegations — loaded fresh, in-tx, at this exact moment — on failure.
	// It calls the SAME domain.CheckEligibility either way; this is not a
	// second, parallel eligibility rule.
	delegations, err := s.repo.LoadActiveDelegationsFor(ctx, tx, req.TenantID, req.ActorUserID, s.clock.Now())
	if err != nil {
		return "", nil, fmt.Errorf("recordSignoff: load active delegations: %w", err)
	}
	onBehalfOf, err := domain.ResolveEligibleIdentity(req.ActorUserID, activeStage.EligibleActorIDs, delegations)
	if err != nil {
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeSignoffRejected,
			ActorUserID:  req.ActorUserID,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			Reason:       "not_eligible",
			OccurredAt:   s.clock.Now(),
		}
		return "", &event, err
	}

	// SoD check — author cannot sign, actor cannot sign twice in same
	// instance, and (F9/ADR 0077) a delegate cannot act on behalf of a
	// delegator who is the author — same shared predicate, widened input.
	priorSignoffs, err := s.repo.LoadPriorSignoffs(ctx, tx, req.TenantID, req.InstanceID, activeStage.ID)
	if err != nil {
		return "", nil, fmt.Errorf("recordSignoff: load prior signoffs: %w", err)
	}
	if err := domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, onBehalfOf, priorSignoffs); err != nil {
		return "", nil, err
	}
	return onBehalfOf, nil, nil
}

// applySignoffQuorum collects all signoffs recorded so far for the active
// stage, evaluates quorum (applying any eligibility-drift policy), and
// applies the resulting stage/instance transition in place on result.
func (s *DecisionService) applySignoffQuorum(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance, activeStage *domain.StageInstance, areaCode string, now time.Time, result *SignoffResult) error {
	// Step 9: collect all signoffs for the active stage to evaluate quorum.
	allStageSignoffs, err := s.repo.LoadStageSignoffs(ctx, tx, req.TenantID, activeStage.ID)
	if err != nil {
		return fmt.Errorf("recordSignoff: load stage signoffs: %w", err)
	}

	// Step 10: evaluate quorum.
	approvals, rejections := splitSignoffs(allStageSignoffs)
	currentEligible := activeStage.EligibleActorIDs
	if activeStage.OnEligibilityDriftSnapshot != domain.DriftKeepSnapshot {
		currentEligible, err = resolveCurrentEligibleForDrift(ctx, tx, s.repo, s.cdRead, req.TenantID, *activeStage, instance.Subject)
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

	switch outcome {
	case domain.QuorumApprovedStage:
		// Step 11a: mark stage completed.
		return s.applyApprovedStageOutcome(ctx, tx, req, instance, activeStage, areaCode, now, result)
	case domain.QuorumRejectedStage:
		// Reject path — mark stage and instance rejected.
		return s.applyRejectedStageOutcome(ctx, tx, req, instance, activeStage, areaCode, now, result)
	default:
		// QuorumPending — no stage transition needed.
		return nil
	}
}

// applyApprovedStageOutcome marks the active stage completed and advances
// the in-memory instance. When that completes every stage it drives the
// document or template terminal-approval path; otherwise it activates the
// next stage AdvanceStage marked active.
func (s *DecisionService) applyApprovedStageOutcome(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance, activeStage *domain.StageInstance, areaCode string, now time.Time, result *SignoffResult) error {
	if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageCompleted, domain.StageActive); err != nil {
		return fmt.Errorf("recordSignoff: complete stage: %w", err)
	}
	result.StageCompleted = true

	// Advance the in-memory instance to determine next step.
	if err := instance.AdvanceStage(); err != nil {
		return fmt.Errorf("recordSignoff: advance stage: %w", err)
	}

	if instance.Status != domain.InstanceApproved {
		// Activate the next stage that AdvanceStage marked active.
		if nextStage := instance.Active(); nextStage != nil {
			if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, nextStage.ID, domain.StageActive, domain.StagePending); err != nil {
				return fmt.Errorf("recordSignoff: activate next stage: %w", err)
			}
		}
		return nil
	}

	// Note (F5, W10): the unresolved-comments gate that used to run here was
	// removed. By construction (plan.md "no new call site" finding) an
	// approval-kind stage only ever activates after freeze has already fired
	// — from ReviewVerdictService.RecordVerdict's stage-advance path
	// (review->approval transitions) or from
	// SubmitService.SubmitRevisionForReview (approval-only routes). Freeze's
	// own instance-scoped comment check (ErrFreezeBlockedByUnresolvedComments)
	// is now the sole gate for this concern; decision_service.go never needs
	// its own copy.
	//
	// All stages done — complete instance. The DOCUMENT arm defers the status
	// write to the shared terminal-approval path below (it owns the CAS); the
	// TEMPLATE arm still writes it here.
	if instance.Subject.Kind == domain.SubjectKindTemplate {
		// Shared template terminal-approval path: instance CAS +
		// templates_template_version under_review -> approved in this same tx
		// (M3 P3.S2b-3b-iii-b). No document-table transition and no
		// async-freeze Pin — templates never freeze. F-E4-4: req.ActorUserID
		// is the deciding approver (the same identity the signoff ledger row
		// and the document arm's FinalApproverID carry), stamped onto the
		// version's approver_id so approved_at never lands without
		// attribution. The ADR 0087 template auto-approve route reaches
		// terminal approval through this very helper.
		if err := completeTemplateTerminalApproval(ctx, tx, templateTerminalApprovalPorts{
			repo:               s.repo,
			templateCompletion: s.templateCompletion,
			serviceName:        "decision service",
		}, templateTerminalApprovalInput{
			TenantID:          req.TenantID,
			InstanceID:        req.InstanceID,
			TemplateVersionID: instance.Subject.Key,
			ApproverID:        req.ActorUserID,
			FromStatus:        domain.InstanceInProgress,
			Now:               now,
		}); err != nil {
			return err
		}
		result.InstanceApproved = true
		return nil
	}

	// Shared terminal-approval path (F-QA4-14 / ADR 0085): identical instance
	// CAS, document.edit assertion, approval fact + pin + coordinator
	// evaluation, documents transition and lifecycle event as the
	// review-verdict and ADR 0087 auto-approve routes.
	if err := completeDocumentTerminalApproval(ctx, tx, documentTerminalApprovalPorts{
		repo:              s.repo,
		releaseRecorder:   s.releaseRecorder,
		lifecycleEnqueuer: s.lifecycleEnqueuer,
		serviceName:       "decision service",
	}, documentTerminalApprovalInput{
		TenantID:             req.TenantID,
		InstanceID:           req.InstanceID,
		DocumentID:           instance.DocumentID,
		AreaCode:             areaCode,
		RevisionVersion:      instance.RevisionVersion,
		FrozenContentHash:    derefString(instance.FrozenContentHash),
		FinalApproverID:      req.ActorUserID,
		SubmittedBy:          instance.SubmittedBy,
		ApproverCapabilities: req.Capabilities,
		FromStatus:           domain.InstanceInProgress,
		Now:                  now,
	}); err != nil {
		return err
	}
	result.InstanceApproved = true
	return nil
}

// applyRejectedStageOutcome marks the active stage and instance rejected,
// then reverts the subject (template version or document) out of
// under_review so the author can revise and resubmit.
func (s *DecisionService) applyRejectedStageOutcome(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance, activeStage *domain.StageInstance, areaCode string, now time.Time, result *SignoffResult) error {
	if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageRejectedHere, domain.StageActive); err != nil {
		return fmt.Errorf("recordSignoff: reject stage: %w", err)
	}
	if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
		domain.InstanceRejected, domain.InstanceInProgress, &now); err != nil {
		return fmt.Errorf("recordSignoff: reject instance: %w", err)
	}
	result.InstanceRejected = true

	if instance.Subject.Kind == domain.SubjectKindTemplate {
		// M3 P3.S2b-3b-iii-b: no cancel GUC, no CapDocumentEdit assert (both
		// authorize the documents-table trigger arc only) — the completion
		// port drives templates_template_version under_review -> draft
		// atomically in this same tx.
		if s.templateCompletion == nil {
			return fmt.Errorf("recordSignoff: template completion writer not configured")
		}
		if err := s.templateCompletion.MarkTemplateVersionRejected(ctx, tx, req.TenantID, instance.Subject.Key); err != nil {
			return fmt.Errorf("recordSignoff: mark template version rejected: %w", err)
		}
		return nil
	}

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
	return nil
}

// emitSignoffOutcome emits the governance event recording this signoff and,
// for a terminal document rejection only, the F3.3 domain lifecycle event.
// DOCUMENT-ONLY (M3 P3.S2b-3b-iii-b): the lifecycle event types are
// documents-domain events; a template-subject instance has no document to
// notify about, so that block is skipped entirely for it. The APPROVED leg
// lives in completeDocumentTerminalApproval (the shared terminal-approval
// path) so all three routes to terminal approval emit it identically; only
// the REJECTED leg is emitted here.
func (s *DecisionService) emitSignoffOutcome(ctx context.Context, tx *sql.Tx, req SignoffRequest, instance *domain.Instance, activeStage *domain.StageInstance, contentHash, onBehalfOf string, result SignoffResult, now time.Time) error {
	payloadMap := map[string]any{
		"instance_id":       req.InstanceID,
		"stage_instance_id": activeStage.ID,
		"decision":          req.Decision,
		"content_hash":      contentHash,
		"on_behalf_of":      onBehalfOf,
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

	// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Author events — terminal
	// transitions only.
	if s.lifecycleEnqueuer != nil && instance.Subject.Kind != domain.SubjectKindTemplate && result.InstanceRejected {
		largs := docsdomain.LifecycleEventArgs{
			EventID:      uuid.NewString(),
			TenantID:     req.TenantID,
			EventType:    docsdomain.EventTypeDocumentRejected,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			SubmittedBy:  instance.SubmittedBy,
			OccurredAt:   now,
		}
		if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
			return fmt.Errorf("recordSignoff: enqueue lifecycle event: %w", err)
		}
	}
	return nil
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
