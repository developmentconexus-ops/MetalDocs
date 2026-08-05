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

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// TemplateVersionReader is the approval-owned, subject-generic port onto the
// templates bounded context (M3 P3.S2b-3b-ii). approval never imports
// templates infrastructure or reads templates_template_version directly —
// this narrow interface (mirroring the PinInvoker injection pattern in
// decision_service.go) is the ONLY surface a template-version status check
// crosses the module boundary through. The production adapter lives in
// templates/infrastructure and is wired at the composition root.
type TemplateVersionReader interface {
	// LoadTemplateVersionStatus returns the status of templateVersionID,
	// scoped to tenantID, read inside the caller's transaction. ok is false
	// when the version does not exist for this tenant (no row, or a
	// cross-tenant id) — the same not-found shape LoadControlledDocumentID
	// uses elsewhere in this package.
	LoadTemplateVersionStatus(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (status string, ok bool, err error)

	// LoadTemplateVersionContentHash returns
	// templates_template_version.content_hash for templateVersionID, scoped
	// to tenantID, read inside the caller's transaction (M3
	// P3.S2b-3b-iii-b). DecisionService reads through this in place of the
	// document-only freeze pin (LoadFrozenContentHash): a template version's
	// content is locked (no author edits) once it is under_review, so its
	// stored content_hash is the same immutable content identity a frozen
	// pin captures for a document — not a fallback, the authoritative value.
	// ok is false when no row matches (absent id, cross-tenant id, or an
	// empty/never-set hash).
	LoadTemplateVersionContentHash(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (contentHash string, ok bool, err error)

	// LoadTemplateInboxMeta (Unit 4.2) batch-resolves display metadata for the
	// worklist/inbox: for each of versionIDs, the owning template's id and
	// name, scoped to tenantID, read inside the caller's transaction. A
	// version id absent from versionIDs, not found, or belonging to another
	// tenant is simply OMITTED from the returned map — not an error — mirroring
	// the not-found shape of the other two methods on this port (no-fallback
	// principle: a missing row never resolves to a substitute value).
	LoadTemplateInboxMeta(ctx context.Context, tx db.Tx, tenantID string, versionIDs []string) (map[string]domain.TemplateInboxMeta, error)

	// LoadTemplateDocTypeCode returns templates_template.doc_type_code for
	// templateID, scoped to tenantID, read inside the caller's transaction
	// (ADR 0086). It is the template ROUTE's subject key: routes are keyed by
	// configuration (the profile), never by the template instance. Fails closed
	// — an absent/cross-tenant template, or one whose doc_type_code is empty,
	// returns an error rather than a substitute value, so a mis-configured
	// template can never silently fall through to some other profile's route.
	LoadTemplateDocTypeCode(ctx context.Context, tx db.Tx, tenantID, templateID string) (docTypeCode string, err error)
}

// TemplateVersionSubmitWriter is the approval-owned port through which the
// kernel submit path locks a template version's own status column
// (draft -> under_review) in the submit tx (M3 P3.S2b-3b-iii-b, hub Option
// (a)). Mirrors the TemplateCompletionWriter / PinInvoker seam: approval
// never UPDATEs templates_template_version directly — this narrow interface
// is the ONLY write surface, satisfied by templates/infrastructure's
// ApprovalCompletionWriter and wired at the composition root. The write runs
// INSIDE the submit tx so the version-status lock is atomic with the
// approval-instance creation (a version can never be kernel-submitted yet
// still author-editable).
type TemplateVersionSubmitWriter interface {
	// MarkTemplateVersionUnderReview flips the version draft -> under_review.
	// Implementations MUST carry the committed-content precondition INSIDE the
	// CAS predicate (not as a preceding read): under Read Committed a
	// concurrent edit can clear content_hash between this service's fast-path
	// check and the UPDATE, and only an atomic CAS turns that race into a
	// clean 0-row miss instead of a raw
	// chk_template_version_content_hash_non_draft violation. A 0-row CAS
	// explained by a still-draft row with no committed hash MUST return
	// ErrTemplateVersionNoContent; every other 0-row cause keeps the
	// implementation's own stale/conflict error.
	MarkTemplateVersionUnderReview(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error
}

// ErrTemplateVersionNotFound is returned when the submitted template version
// does not exist for the tenant.
var ErrTemplateVersionNotFound = errors.New("approval: template version not found")

// ErrTemplateVersionNotDraft is returned when the submitted template version
// is not in the draft status required to enter approval (mirrors the
// templates module's own draft-guard semantics, domain.VersionStatusDraft).
var ErrTemplateVersionNotDraft = errors.New("approval: template version is not draft")

// ErrTemplateVersionNoContent is returned when the submitted template version
// carries no committed content hash (F-E4-1). templates_template_version's
// chk_template_version_content_hash_non_draft CHECK constraint forbids any
// non-draft row without a 64-char content_hash, so a submit that flipped the
// version to under_review with an empty hash surfaced as a raw 23514 →
// INTERNAL_ERROR 500 with the constraint name leaked into the problem title.
//
// Two lines produce it, both mapping to the same sentinel: the friendly
// fast-path read below (rejects the submit before the transition is attempted)
// and the submit-lock CAS itself, whose WHERE clause carries the same
// precondition so a concurrent edit that clears the hash mid-tx loses the CAS
// instead of tripping the CHECK. The DB constraint remains the authoritative
// backstop behind both.
var ErrTemplateVersionNoContent = domain.ErrTemplateVersionNoContent

// TemplateSubmitService creates approval instances for template versions
// (subject_kind='template'). It is a THIN, subject-generic subset of
// SubmitService.SubmitRevisionForReview: it reuses the same kernel
// stage-seeding primitives (ResolveEligibleActors → StageInstance →
// InsertStageInstances) but deliberately OMITS every document-only concern —
// content-hash computation, freeze, revision-title/reason-for-change, the
// documents status transition, and CD-link/profile-policy reads. Those do
// not apply to a template version.
type TemplateSubmitService struct {
	repo          infrastructure.ApprovalRepository
	emitter       EventEmitter
	clock         Clock
	versionReader TemplateVersionReader
	// notifier enqueues the approval.pending notification (Task 6) inside the
	// same submit transaction as the audit GovernanceEvent above — the
	// transactional-outbox discipline: state write and the promise to notify
	// commit together. Defaults to domain.NoopApprovalNotificationEnqueuer
	// (fail-closed: enqueues nothing, reports success) when nil is passed to
	// NewTemplateSubmitService.
	notifier domain.ApprovalNotificationEnqueuer
	// versionWriter locks the version's own status column draft->under_review
	// in the submit tx (M3 P3.S2b-3b-iii-b, hub Option (a)). nil fails closed
	// (an explicit error) rather than leaving the concurrent-edit hole open.
	versionWriter TemplateVersionSubmitWriter
	// templateCompletion drives templates_template_version under_review ->
	// approved when an ADR 0087 zero-stage (livre) route auto-approves the
	// version in the submit tx. Wired by Services.WithTemplateCompletionWriter,
	// the same adapter DecisionService uses; nil fails closed (an explicit
	// error) rather than leaving the version stranded in under_review behind an
	// already-approved instance.
	templateCompletion TemplateCompletionWriter
	// routeResolver reaches LoadActiveRouteIDBySubject, declared on
	// SubmitDefaultsResolver (not the broader ApprovalRepository interface —
	// same ISP split SubmitService.resolver uses). Populated via type
	// assertion against repo in NewTemplateSubmitService; nil for a repo that
	// does not implement it (fails closed to ErrNoActiveApprovalRoute-shaped
	// errors surfacing as a route-load failure).
	routeResolver SubmitDefaultsResolver
}

// NewTemplateSubmitService constructs a TemplateSubmitService. versionReader
// may be nil in unit tests that do not exercise the draft-status guard; a nil
// reader fails closed (ErrTemplateVersionNotFound) in production use. notifier
// may be nil, in which case it defaults to domain.NoopApprovalNotificationEnqueuer
// (Task 6) — production composition roots must wire a real enqueuer or submit
// silently tells nobody.
func NewTemplateSubmitService(repo infrastructure.ApprovalRepository, emitter EventEmitter, clock Clock, versionReader TemplateVersionReader, notifier domain.ApprovalNotificationEnqueuer) *TemplateSubmitService {
	resolver, _ := repo.(SubmitDefaultsResolver)
	if notifier == nil {
		notifier = domain.NoopApprovalNotificationEnqueuer{}
	}
	return &TemplateSubmitService{repo: repo, emitter: emitter, clock: clock, versionReader: versionReader, routeResolver: resolver, notifier: notifier}
}

// notifierOrNoop returns s.notifier, defaulting to the fail-closed Noop
// implementation. Guards against a raw struct literal (e.g. a test fixture
// built without NewTemplateSubmitService, or Services.WithApprovalNotificationEnqueuer
// passed an explicit nil) leaving the field unset.
func (s *TemplateSubmitService) notifierOrNoop() domain.ApprovalNotificationEnqueuer {
	if s.notifier == nil {
		return domain.NoopApprovalNotificationEnqueuer{}
	}
	return s.notifier
}

// TemplateSubmitRequest carries all inputs for SubmitTemplateVersionForReview.
type TemplateSubmitRequest struct {
	TenantID          string
	TemplateID        string // owning template; its doc_type_code is the ROUTE.subject_key (ADR 0086)
	TemplateVersionID string // artifact under approval (INSTANCE.subject_key)
	SubmittedBy       string // user_id
	IdempotencyKey    string // client Idempotency-Key header, threaded from the handler
	// ChosenActors mirrors SubmitRequest.ChosenActors (M4, unit 3.2, slice 5):
	// per-stage caller-chosen actors for stages governed by a submit_choice
	// actor selector.
	ChosenActors []StageChosenActors
}

// TemplateSubmitResult is returned on successful submission.
type TemplateSubmitResult struct {
	InstanceID string // UUID of created approval_instance
}

// SubmitTemplateVersionForReview creates a new approval instance for the
// template version. Returns infrastructure.ErrDuplicateSubmission (unwrapped)
// when a concurrent submission with the same idempotency key already exists,
// domain.ErrEmptyEligiblePool when any stage resolves to zero eligible
// actors, and ErrTemplateVersionNotDraft when the version is not draft.
func (s *TemplateSubmitService) SubmitTemplateVersionForReview(ctx context.Context, runner db.TxRunner, req TemplateSubmitRequest) (TemplateSubmitResult, error) {
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		return TemplateSubmitResult{}, ErrIdempotencyKeyRequired
	}

	var instanceID string
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// template.submit is the ratified capability for a template-subject
		// submit (ADR 0083). It is ScopeTenant, so the "tenant" area
		// sentinel is the correct area-blind enforcement — templates carry
		// no process area to resolve (P3.S2b-3a) — and api-lint's
		// area-scope-binding rule is satisfied (ScopeTenant caps are
		// enforced area-blind by design, unlike area-grade CapDocumentSubmit).
		// The DB tripwire on approval_instances is now subject-discriminated
		// (ADR 0083, migration 0299): the template arm accepts template.submit,
		// the document arm accepts document.submit, fail-closed ELSE.
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
			return err
		}

		// Precondition: the version must be draft before it can enter
		// approval. Read through the approval-owned port only — never a
		// direct SQL read against templates_template_version.
		if s.versionReader == nil {
			return ErrTemplateVersionNotFound
		}
		status, ok, err := s.versionReader.LoadTemplateVersionStatus(ctx, tx, req.TenantID, req.TemplateVersionID)
		if err != nil {
			return fmt.Errorf("template submit: load template version status: %w", err)
		}
		if !ok {
			return ErrTemplateVersionNotFound
		}
		if status != "draft" {
			return ErrTemplateVersionNotDraft
		}

		// Precondition (F-E4-1): the version must carry committed content
		// before it can enter approval. This read is the FRIENDLY FAST PATH
		// only — it is not the enforcement point. Read Committed lets a
		// concurrent edit clear content_hash after this read and before the
		// flip below, so the authoritative precondition lives in the
		// submit-lock CAS predicate itself (see TemplateVersionSubmitWriter),
		// with the DB's chk_template_version_content_hash_non_draft CHECK as
		// the backstop behind that. ok is false for an absent row, a
		// cross-tenant id, or an empty/never-set hash; the row existence case
		// is already resolved above, so !ok here means "no committed content".
		if _, ok, err := s.versionReader.LoadTemplateVersionContentHash(ctx, tx, req.TenantID, req.TemplateVersionID); err != nil {
			return fmt.Errorf("template submit: load template version content hash: %w", err)
		} else if !ok {
			return ErrTemplateVersionNoContent
		}

		// Submit-lock (M3 P3.S2b-3b-iii-b, hub Option (a)): flip the version's
		// own status column draft -> under_review inside this same tx,
		// atomically with the approval-instance creation below. This closes
		// the concurrent-edit hole — the templates module's edit/upload gates
		// permit writes only while status='draft', so a kernel-submitted
		// version becomes immutable to the author for the approval's duration.
		// Fail-closed on a nil writer or a 0-row CAS (draft precondition
		// lost). template.submit is already asserted above in this tx, so the
		// tripwire GUC authorizes the UPDATE without a second assertion.
		if s.versionWriter == nil {
			return fmt.Errorf("template submit: version submit-writer not configured")
		}
		if err := s.versionWriter.MarkTemplateVersionUnderReview(ctx, tx, req.TenantID, req.TemplateVersionID); err != nil {
			// The CAS lost to a concurrent edit that cleared the content hash:
			// same condition the fast path above reports, so it surfaces as the
			// same sentinel (409 UPLOAD_MISSING) rather than an opaque wrap.
			if errors.Is(err, ErrTemplateVersionNoContent) {
				return ErrTemplateVersionNoContent
			}
			return fmt.Errorf("template submit: lock version under_review: %w", err)
		}

		// Resolve the active route by the template's PROFILE (doc_type_code) —
		// ROUTE.subject_key = doc_type_code (ADR 0086: a route is
		// configuration, one per profile), distinct from INSTANCE.subject_key =
		// template_version_id (the two-level keying). The doc_type_code is read
		// in-tx through the approval-owned port and fails closed; PreviewRoute
		// resolves it through the SAME method so preview and submit can never
		// disagree about which route governs a template.
		if s.routeResolver == nil {
			return fmt.Errorf("template submit: route resolver not configured")
		}
		docTypeCode, err := s.versionReader.LoadTemplateDocTypeCode(ctx, tx, req.TenantID, req.TemplateID)
		if err != nil {
			return fmt.Errorf("template submit: resolve template doc_type_code: %w", err)
		}
		routeID, err := s.routeResolver.LoadActiveRouteIDBySubject(ctx, tx, req.TenantID, string(domain.SubjectKindTemplate), docTypeCode)
		if err != nil {
			if errors.Is(err, infrastructure.ErrNoActiveApprovalRoute) {
				return infrastructure.ErrNoActiveApprovalRoute
			}
			return fmt.Errorf("template submit: resolve active route: %w", err)
		}

		route, err := s.repo.LoadRoute(ctx, tx, req.TenantID, routeID)
		if err != nil {
			return fmt.Errorf("template submit: load route: %w", err)
		}
		if err := route.Validate(""); err != nil {
			return fmt.Errorf("template submit: invalid route: %w", err)
		}

		instanceID = uuid.New().String()
		now := s.clock.Now()

		inst := domain.Instance{
			ID:                   instanceID,
			TenantID:             req.TenantID,
			Subject:              domain.NewTemplateSubject(req.TemplateVersionID),
			RouteID:              routeID,
			RouteVersionSnapshot: route.Version,
			Status:               domain.InstanceInProgress,
			SubmittedBy:          req.SubmittedBy,
			SubmittedAt:          now,
			IdempotencyKey:       idempotencyKey,
		}

		if err := s.repo.InsertInstance(ctx, tx, inst); err != nil {
			if errors.Is(err, infrastructure.ErrDuplicateSubmission) {
				return err
			}
			return fmt.Errorf("template submit: %w", err)
		}

		// submit_choice pre-pass (M4, unit 3.2, slice 5): mirrors
		// SubmitService.SubmitRevisionForReview's Step 7b. Fail-closed before
		// any stage instance is created.
		submitChoiceStages := stagesWithSubmitChoice(route.Stages)
		if err := validateChosenActorsTargets(req.ChosenActors, submitChoiceStages); err != nil {
			return err
		}

		stageInstances := make([]domain.StageInstance, len(route.Stages))
		for i, stage := range route.Stages {
			status := domain.StagePending
			var openedAt *time.Time
			if i == 0 {
				status = domain.StageActive
				openedAt = &now
			}
			flatRole, flatArea := domain.FlatRoleArea(stage.Selectors)
			eligibleIDs, err := s.repo.ResolveEligibleActorsForSelectors(ctx, tx, req.TenantID, stage.Selectors, flatArea)
			if err != nil {
				return fmt.Errorf("template submit: resolve eligible actors for stage %d: %w", stage.Order, err)
			}
			var chosenIDs []string
			if sel, ok := submitChoiceStages[stage.Order]; ok {
				chosenIDs, err = resolveSubmitChoiceEligibleIDs(ctx, tx, s.repo, req.TenantID, sel, req.ChosenActors, stage.Order)
				if err != nil {
					return err
				}
				eligibleIDs = unionUserIDs(eligibleIDs, chosenIDs)
			}
			if len(eligibleIDs) == 0 {
				return domain.ErrEmptyEligiblePool
			}
			stageInstances[i] = domain.StageInstance{
				ID:                         uuid.New().String(),
				ApprovalInstanceID:         instanceID,
				StageOrder:                 stage.Order,
				NameSnapshot:               stage.Name,
				RequiredRoleSnapshot:       flatRole,
				RequiredCapabilitySnapshot: stage.RequiredCapability,
				AreaCodeSnapshot:           flatArea,
				QuorumSnapshot:             stage.Quorum,
				QuorumMSnapshot:            stage.QuorumM,
				OnEligibilityDriftSnapshot: stage.OnEligibilityDrift,
				Kind:                       stage.Kind,
				EligibleActorIDs:           eligibleIDs,
				Status:                     status,
				OpenedAt:                   openedAt,
				DueInDaysSnapshot:          stage.DueInDays,
				SelectorsSnapshot:          materializeSelectorsSnapshot(stage.Selectors, chosenIDs),
			}
		}

		if err := s.repo.InsertStageInstances(ctx, tx, stageInstances); err != nil {
			return fmt.Errorf("template submit: %w", err)
		}

		// Governance event payload — no content_hash (document-only concern
		// omitted per rails).
		payloadBytes, err := json.Marshal(map[string]any{
			"instance_id": instanceID,
			"route_id":    routeID,
		})
		if err != nil {
			return fmt.Errorf("template submit: marshal event payload: %w", err)
		}

		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    "approval_submitted",
			ActorUserID:  req.SubmittedBy,
			ResourceType: "template",
			ResourceID:   req.TemplateVersionID,
			PayloadJSON:  json.RawMessage(payloadBytes),
			OccurredAt:   now,
		}
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("template submit: emit event: %w", err)
		}

		// The audit event above is a governance record; this is the
		// notification. They are deliberately separate paths with separate
		// consumers — conflating them is what left submit silent.
		//
		// Recipients are the ACTIVE stage's eligible actors, resolved above and
		// already snapshotted onto the stage instance. A livre route has no
		// active stage, so recipients come out empty and nothing is enqueued —
		// the degenerate case needs no branch of its own.
		if recipients := activeStageEligibleIDs(stageInstances); len(recipients) > 0 {
			if err := s.notifierOrNoop().EnqueueApprovalNotificationTx(ctx, tx, domain.ApprovalNotificationArgs{
				EventID:          uuid.New().String(),
				TenantID:         req.TenantID,
				EventType:        domain.EventTypeApprovalPending,
				SubjectKind:      "template",
				SubjectID:        req.TemplateVersionID,
				RecipientUserIDs: recipients,
				ActorUserID:      req.SubmittedBy,
				OccurredAt:       now,
			}); err != nil {
				// NOT swallowed. If the notification cannot be promised, the
				// submission is not confirmed: a half-transaction that reports
				// success and tells nobody is the defect this feature exists to
				// remove.
				return fmt.Errorf("template submit: enqueue approval notification: %w", err)
			}
		}

		// ADR 0087 §5: a route bound to a livre profile carries ZERO stages, so
		// the instance created above has nothing left to satisfy and reaches
		// terminal approval in THIS transaction — for BOTH subjects. Without
		// this, a livre template submit left an in_progress instance behind an
		// under_review version that nothing could ever advance.
		if len(route.Stages) == 0 {
			if err := s.autoApprove(ctx, tx, req, &inst, route, now); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return TemplateSubmitResult{}, err
	}

	return TemplateSubmitResult{InstanceID: instanceID}, nil
}

// autoApprove completes a stageless template instance inside the submit
// transaction (ADR 0087 §5), mirroring SubmitService.autoApprove for the
// document subject. It runs AFTER the version's draft→under_review lock and the
// approval_submitted event, so the trail reads submitted-then-approved and the
// version is in the under_review state the shared template terminal-approval
// path transitions out of.
//
// The domain decides that a stageless instance may complete
// (domain.Instance.AutoApprove — fail-closed on a staged or non-in_progress
// instance); the application layer performs the persistence and the side
// effects, through the same helper the signoff route uses.
func (s *TemplateSubmitService) autoApprove(
	ctx context.Context,
	tx *sql.Tx,
	req TemplateSubmitRequest,
	inst *domain.Instance,
	route domain.Route,
	now time.Time,
) error {
	if err := inst.AutoApprove(now); err != nil {
		return fmt.Errorf("template submit: auto-approve: %w", err)
	}

	// Record WHY this version is approved with an empty signoff set, naming the
	// route + route version that authorized it (ADR 0087 §5 — governance_events
	// is the binding record, same store and vocabulary as approval_submitted).
	// Subject info rides in the payload and in the resource fields.
	payloadBytes, err := json.Marshal(map[string]any{
		"instance_id":   inst.ID,
		"subject_kind":  string(domain.SubjectKindTemplate),
		"subject_key":   inst.Subject.Key,
		"template_id":   req.TemplateID,
		"route_id":      route.ID,
		"route_version": route.Version,
		"stage_count":   0,
		"reason":        "zero_stage_route",
	})
	if err != nil {
		return fmt.Errorf("template submit: auto-approve: marshal event payload: %w", err)
	}
	if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    EventTypeApprovalAutoApproved,
		ActorUserID:  req.SubmittedBy,
		ResourceType: "approval_instance",
		ResourceID:   inst.ID,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}); err != nil {
		return fmt.Errorf("template submit: auto-approve: emit event: %w", err)
	}

	// ApproverID is the submitter: on a zero-stage route the submission IS the
	// approving act, so approver attribution stays honest rather than empty.
	return completeTemplateTerminalApproval(ctx, tx, templateTerminalApprovalPorts{
		repo:               s.repo,
		templateCompletion: s.templateCompletion,
		serviceName:        "template submit service",
	}, templateTerminalApprovalInput{
		TenantID:          req.TenantID,
		InstanceID:        inst.ID,
		TemplateVersionID: inst.Subject.Key,
		ApproverID:        req.SubmittedBy,
		FromStatus:        domain.InstanceInProgress,
		Now:               now,
	})
}
