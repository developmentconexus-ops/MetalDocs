package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// SubmitService handles document submission for approval.
type SubmitService struct {
	repo         infrastructure.ApprovalRepository
	emitter      EventEmitter
	clock        Clock
	cdRead       controlleddocumentsdomain.CDFieldReader
	resolver     SubmitDefaultsResolver // in-tx route/hash resolution when the client omits them (ADR 0073); nil in unit tests that always supply route_id
	policyReader ProfilePolicyReader    // G1: resolves the profile route-signature policy off-tx; nil ⇒ friendly check skipped (DB trigger is the last line)
}

// WithPolicyReader returns a copy of the service wired with the G1
// profile-policy reader. Passing nil leaves the friendly route-shape check
// disabled (the DB deferrable trigger remains the authoritative last line).
func (s *SubmitService) WithPolicyReader(reader ProfilePolicyReader) *SubmitService {
	if s == nil {
		return nil
	}
	cp := *s
	cp.policyReader = reader
	return &cp
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
	// ChosenActors carries the caller-chosen actors per stage_order, for
	// stages governed by a submit_choice actor selector (M4, unit 3.2, slice
	// 5; spec §5, §4). A submit_choice stage contributes no pool from the
	// route alone — the chosen users are validated server-side against the
	// selector's role x area_code constraint, then unioned into that stage's
	// eligible pool. Ignored for stage_orders that carry no submit_choice
	// selector — see resolveSubmitChoiceEligibleIDs for the fail-closed
	// contract on mismatches.
	ChosenActors []StageChosenActors

	// Publication plan (ADR 0085). Persisted onto the documents row in the SAME
	// transaction as the draft→under_review transition, so the release
	// coordinator reads a plan that is durable the instant the revision enters
	// governance. All four are optional; a nil/empty plan means "release
	// immediately once every readiness fact holds".
	//
	// PlannedEffectiveFrom is the PLAN. documents.effective_from remains the
	// ACTUAL release instant and is written only by the release transaction —
	// this path must never touch it.
	PlannedEffectiveFrom *time.Time
	EffectiveTo          *time.Time
	ReviewDueAt          *time.Time
	// SupersedeTargetID is a CROSS-document supersede target (a published
	// document of a DIFFERENT controlled document). Naming one requires
	// document.supersede on the TARGET's area, checked in this transaction.
	// "" = none; same-controlled-document supersession is implicit and is
	// discovered by the coordinator, never named here.
	SupersedeTargetID string
}

// StageChosenActors carries the caller-supplied chosen_actors for one stage
// (M4, unit 3.2, slice 5). Shared between SubmitRequest and
// TemplateSubmitRequest — both submit paths resolve submit_choice stages
// identically.
type StageChosenActors struct {
	StageOrder int
	UserIDs    []string
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
		areaCode, err := resolveSubjectAreaCode(ctx, tx, s.cdRead, req.TenantID, domain.NewDocumentSubject(req.DocumentID))
		if err != nil {
			return fmt.Errorf("submit: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode); err != nil {
			return err
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return err
		}

		// Step 4a-bis (ADR 0085): a publication plan that names a CROSS-document
		// supersede target requires document.supersede on the TARGET's area —
		// not the submitting document's. Authority over retiring a document
		// belongs to that document's area, and this is the one moment the
		// caller is present to be asked; the release coordinator later acts as
		// a system principal with no human authority to check.
		//
		// H-PRE-1: this authz-recording read runs here, at the top of the tx,
		// BEFORE any FOR UPDATE. The target validation below it is likewise a
		// plain non-locking SELECT.
		if supersedeTargetID := strings.TrimSpace(req.SupersedeTargetID); supersedeTargetID != "" {
			if supersedeTargetID == req.DocumentID {
				return infrastructure.ErrInvalidSupersedeTarget
			}
			targetAreaCode, err := resolveSubjectAreaCode(ctx, tx, s.cdRead, req.TenantID, domain.NewDocumentSubject(supersedeTargetID))
			if err != nil {
				return fmt.Errorf("submit: load supersede target area: %w", err)
			}
			if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSupersede), targetAreaCode); err != nil {
				return err
			}
			if err := s.repo.ValidateCrossDocumentSupersedeTarget(ctx, tx, req.TenantID, req.DocumentID, supersedeTargetID); err != nil {
				return err
			}
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

		// Step 4c: resolve the approval route in-tx when the client omits route_id
		// (ADR 0073 §2). An author cannot supply route_id (route listing needs an
		// admin read the author lacks); the server resolves the single active route
		// for the document's controlled-document profile inside the SAME
		// transaction, closing the wrapper-era off-tx TOCTOU. Explicit client
		// route_id keeps its exact prior semantics. Resolution runs only when the
		// resolver is wired (production + F1 integration); unit tests that always
		// pass an explicit route_id leave it nil and are unaffected.
		routeID := req.RouteID
		if strings.TrimSpace(routeID) == "" {
			resolvedRouteID, rerr := s.resolveActiveRoute(ctx, tx, req.TenantID, req.DocumentID)
			if rerr != nil {
				return rerr
			}
			routeID = resolvedRouteID
		}

		// Step 5: load route with stages.
		route, err := s.repo.LoadRoute(ctx, tx, req.TenantID, routeID)
		if err != nil {
			return fmt.Errorf("submit: load route: %w", err)
		}

		// Step 6: validate route structural invariants. The G1 per-profile
		// route-signature policy is intentionally NOT re-read here (operator-
		// ratified 2026-07-10, ADR 0073-era constraints; see ADR on per-profile
		// signature policy). The invariant "every ACTIVE route conforms to its
		// profile's current governance class" is held by construction at the
		// write boundaries: route-admin Validate(policy) off-tx + the DEFERRABLE
		// both-directions DB trigger (route writes AND reclassification). Submit
		// binds only active routes, so the bound route already conforms; livre
		// profiles have no active route and fail at resolution. Re-reading the
		// policy here cannot satisfy ADR 0073 (route/profile resolved in-tx) and
		// H-PRE-1 (no recording CapTaxonomyView read inside this lock-holding tx)
		// simultaneously — in-tx enforcement is structurally wrong, not merely
		// harder. So Validate runs structural-only (empty Kind fail-closes to
		// approval).
		if err := route.Validate(""); err != nil {
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
		// Step 6b-pre: resolve the head content hash in-tx when the client omits it
		// (ADR 0073 §2). The HTTP submit boundary conveys the client hash through the
		// reserved `_content_hash` form-data key; an empty value there means "author
		// did not supply one" — bind the head autosaved revision's stored hash,
		// exactly as the finalize wrapper did (prereqs.ContentHash → `_content_hash`),
		// so content_hash_at_submit semantics are unchanged (HS-2). No `_content_hash`
		// key at all (service-level callers passing real form data) → no resolution.
		formData, err := s.resolveContentFormData(ctx, tx, req.TenantID, req.DocumentID, req.ContentFormData)
		if err != nil {
			return err
		}

		contentHash, err := ComputeContentHash(ContentHashInput{
			TenantID:       req.TenantID,
			DocumentID:     req.DocumentID,
			RevisionNumber: revisionNumber,
			FormData:       formData,
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
			Subject:              domain.NewDocumentSubject(req.DocumentID),
			RouteID:              routeID,
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

		// Step 7b: submit_choice pre-pass (M4, unit 3.2, slice 5). Fail-closed:
		// any ChosenActors entry naming a stage_order with no submit_choice
		// selector is rejected before any stage instance is created — no
		// silent accept of an out-of-scope choice (no-fallback principle).
		submitChoiceStages := stagesWithSubmitChoice(route.Stages)
		if err := validateChosenActorsTargets(req.ChosenActors, submitChoiceStages); err != nil {
			return err
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
			eligibleIDs, err := s.repo.ResolveEligibleActorsForSelectors(ctx, tx, req.TenantID, stage.Selectors, areaCode)
			if err != nil {
				return fmt.Errorf("submit: resolve eligible actors for stage %d: %w", stage.Order, err)
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
				// W6: a stage pool that resolves to zero eligible actors can never be
				// signed off — fail closed at submit rather than creating a stuck
				// instance. Shared sentinel with decision_service.go's quorum-eval gap.
				return domain.ErrEmptyEligiblePool
			}
			flatRole, flatArea := domain.FlatRoleArea(stage.Selectors)
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
				// DueInDaysSnapshot (F8, spec.md §4/W4): captured for EVERY
				// stage at submit time, not just the first — a stage that
				// activates later (via AdvanceStage) still needs its own SLA
				// config pinned to what the route looked like at submit,
				// consistent with every other *Snapshot field on this struct
				// (F2 route-version pinning: in-flight instances stay pinned
				// to the version they started on).
				DueInDaysSnapshot: stage.DueInDays,
				SelectorsSnapshot: materializeSelectorsSnapshot(stage.Selectors, chosenIDs),
			}
		}

		if err := s.repo.InsertStageInstances(ctx, tx, stageInstances); err != nil {
			return fmt.Errorf("submit: %w", err)
		}

		// Freeze boundary (F5, design spec.md §2.2): approval-only routes (no
		// review-kind stage anywhere) have no review→approval transition for
		// ReviewVerdictService to hook into, so this is the ONLY point content
		// becomes immutable for them — at submit time, using the hash already
		// computed above. Routes containing a review stage freeze later, from
		// ReviewVerdictService.RecordVerdict's stage-advance path.
		hasReviewStage := false
		for _, stage := range route.Stages {
			if stage.Kind == domain.StageKindReview {
				hasReviewStage = true
				break
			}
		}
		if !hasReviewStage {
			if err := executeFreeze(ctx, tx, s.repo, req.TenantID, &inst); err != nil {
				return fmt.Errorf("submit: freeze: %w", err)
			}
		}

		// Step 8b: transition document draft → under_review and persist the ADR
		// 0085 publication plan in the SAME write. Friendly first-line legality
		// check (M4/F4.1) mirrors the DB trigger; the OCC WHERE below remains
		// the atomic CAS + optimistic-lock enforcement.
		//
		// The plan columns are set unconditionally, not COALESCEd: the plan is
		// declared per submission, so a resubmission that omits it clears the
		// previous submission's plan rather than silently inheriting it (a
		// stale inherited effective date would hold a release nobody asked to
		// hold). effective_from is NOT touched here — it is the actual release
		// instant, owned solely by the release transaction.
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusDraft, docsdomain.DocStatusUnderReview); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status                 = 'under_review',
			       revision_title         = $4,
			       reason_for_change      = $5,
			       reason_category        = $6,
			       planned_effective_from = $7,
			       effective_to           = $8,
			       review_due_at          = $9,
			       superseded_document_id = $10,
			       revision_version       = revision_version + 1
			 WHERE id               = $1
			   AND tenant_id        = $2
			   AND status           = 'draft'
			   AND revision_version = $3`,
			req.DocumentID, req.TenantID, req.RevisionVersion, revisionTitle,
			nullableString(reasonForChange), nullableString(reasonCategory),
			nullableTime(req.PlannedEffectiveFrom), nullableTime(req.EffectiveTo),
			nullableTime(req.ReviewDueAt), nullableString(strings.TrimSpace(req.SupersedeTargetID)),
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
			"route_id":     routeID,
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

// resolveActiveRoute resolves the single active approval route for the document's
// controlled-document profile entirely inside the caller's transaction (ADR 0073
// §2). It surfaces the same domain sentinels the finalize wrapper returned:
// docsdomain.ErrProfileNotConfigured when the document has no linked profile,
// docsdomain.ErrApprovalRouteMissing when no active route exists for the profile.
// All reads are plain non-recording SELECTs (HS-PRE-1).
func (s *SubmitService) resolveActiveRoute(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) {
	if s.resolver == nil {
		// route_id omitted but no resolver wired: production always wires one, so
		// this means a misconfigured composition root. Fail closed rather than
		// LoadRoute("") — surface the missing-route sentinel.
		return "", docsdomain.ErrApprovalRouteMissing
	}
	cdID, ok, err := s.resolver.LoadControlledDocumentID(ctx, tx, tenantID, documentID)
	if err != nil {
		return "", fmt.Errorf("submit: resolve controlled document: %w", err)
	}
	var profileCode string
	if ok && cdID != "" {
		// CDFieldReader is the only cross-module read surface (ADR 0072); *sql.Tx
		// satisfies db.DB so the resolution shares this transaction's GUC/RLS scope.
		profileCode, err = s.cdRead.ProfileCode(ctx, tx, tenantID, cdID)
		if err != nil {
			return "", fmt.Errorf("submit: resolve profile code: %w", err)
		}
	}
	if strings.TrimSpace(profileCode) == "" {
		return "", docsdomain.ErrProfileNotConfigured
	}
	routeID, err := s.resolver.LoadActiveRouteIDByProfile(ctx, tx, tenantID, profileCode)
	if err != nil {
		if errors.Is(err, infrastructure.ErrNoActiveApprovalRoute) {
			return "", docsdomain.ErrApprovalRouteMissing
		}
		return "", fmt.Errorf("submit: resolve active route: %w", err)
	}
	return routeID, nil
}

// resolveContentFormData binds the head content hash in-tx when the client omits
// it (ADR 0073 §2). The HTTP submit boundary conveys the client hash through the
// reserved `_content_hash` key; a present-but-empty value means "author supplied
// none" — substitute the head autosaved revision's stored hash. No `_content_hash`
// key (service-level callers passing real form data) or a non-empty value → the
// map is returned unchanged. The caller's map is never mutated.
func (s *SubmitService) resolveContentFormData(ctx context.Context, tx *sql.Tx, tenantID, documentID string, in map[string]any) (map[string]any, error) {
	raw, present := in["_content_hash"]
	if !present || s.resolver == nil {
		return in, nil
	}
	if current, _ := raw.(string); strings.TrimSpace(current) != "" {
		return in, nil
	}
	headHash, err := s.resolver.LoadHeadContentHash(ctx, tx, tenantID, documentID)
	if err != nil {
		return nil, fmt.Errorf("submit: resolve head content hash: %w", err)
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	out["_content_hash"] = headHash
	return out, nil
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

// ---------------------------------------------------------------------------
// submit_choice resolution (M4, unit 3.2, slice 5; spec §5, §4). Shared by
// SubmitService.SubmitRevisionForReview and
// TemplateSubmitService.SubmitTemplateVersionForReview — both submit paths
// resolve submit_choice stages identically.
// ---------------------------------------------------------------------------

// stagesWithSubmitChoice returns, for every stage in stages that carries a
// submit_choice selector, a map of stage.Order to that selector. A stage with
// more than one submit_choice selector is not a contract this feature
// supports; the first one found wins (route construction is not expected to
// produce that shape).
func stagesWithSubmitChoice(stages []domain.Stage) map[int]domain.ActorSelector {
	out := make(map[int]domain.ActorSelector)
	for _, stage := range stages {
		for _, sel := range stage.Selectors {
			if sel.Kind == domain.SelectorSubmitChoice {
				out[stage.Order] = sel
				break
			}
		}
	}
	return out
}

// validateChosenActorsTargets fails closed with
// domain.ErrSubmitChoiceConstraintViolated when any entry in chosen names a
// stage_order that has no submit_choice selector in submitChoiceStages — the
// no-fallback principle applied to an out-of-scope choice (never a silent
// accept).
func validateChosenActorsTargets(chosen []StageChosenActors, submitChoiceStages map[int]domain.ActorSelector) error {
	for _, c := range chosen {
		if _, ok := submitChoiceStages[c.StageOrder]; !ok {
			return domain.ErrSubmitChoiceConstraintViolated
		}
	}
	return nil
}

// findChosenActors returns the ChosenActors entry targeting stageOrder, if
// any.
func findChosenActors(chosen []StageChosenActors, stageOrder int) (StageChosenActors, bool) {
	for _, c := range chosen {
		if c.StageOrder == stageOrder {
			return c, true
		}
	}
	return StageChosenActors{}, false
}

// resolveSubmitChoiceEligibleIDs applies the submit_choice stage's
// fail-closed contract for stageOrder: chosen must carry a non-empty entry
// for it (domain.ErrSubmitChoiceRequired otherwise), and every chosen
// user_id must satisfy selector's role x area_code constraint
// (domain.ErrSubmitChoiceConstraintViolated otherwise, via
// ApprovalRepository.ValidateSubmitChoiceActors). Returns the validated ids
// for the caller to union into the stage's route-derived eligible pool.
func resolveSubmitChoiceEligibleIDs(ctx context.Context, tx *sql.Tx, repo infrastructure.ApprovalRepository, tenantID string, selector domain.ActorSelector, chosen []StageChosenActors, stageOrder int) ([]string, error) {
	entry, ok := findChosenActors(chosen, stageOrder)
	if !ok || len(entry.UserIDs) == 0 {
		return nil, domain.ErrSubmitChoiceRequired
	}
	return repo.ValidateSubmitChoiceActors(ctx, tx, tenantID, selector, entry.UserIDs)
}

// unionUserIDs is the pure union+dedup+sort of one or more user_id pools,
// mirroring infrastructure.unionDedupSort's semantics (deterministic
// ascending order, never nil) for the application layer's post-resolve
// submit_choice union — infrastructure's helper is unexported and DB-package
// scoped, so this is a small, intentional, side-effect-free duplicate rather
// than a cross-package reach.
func unionUserIDs(pools ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, pool := range pools {
		for _, id := range pool {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				out = append(out, id)
			}
		}
	}
	sort.Strings(out)
	return out
}

// materializeSelectorsSnapshot builds the per-stage selectors snapshot frozen
// onto the StageInstance at submit (M4, unit 3.2 slice 6 / Option C′). Every
// submit_choice selector is materialized into one concrete named_user selector
// per validated chosen id (chosenIDs) — post-submit the instance carries only
// named_user (exempt) + role_in_fixed_area/role_in_document_area (pool), so the
// drift path never needs the caller's choice again. All other selectors pass
// through unchanged.
func materializeSelectorsSnapshot(selectors []domain.ActorSelector, chosenIDs []string) []domain.ActorSelector {
	out := make([]domain.ActorSelector, 0, len(selectors))
	for _, sel := range selectors {
		if sel.Kind == domain.SelectorSubmitChoice {
			for _, id := range chosenIDs {
				out = append(out, domain.ActorSelector{Kind: domain.SelectorNamedUser, UserID: id})
			}
			continue
		}
		out = append(out, sel)
	}
	return out
}
