//go:build integration
// +build integration

// Package approval_test — F4 (milestone 2b, approval-kernel-backend)
// review-stage runtime verdicts. Exercises application.ReviewVerdictService
// end-to-end against a real Postgres instance (testdb per-test clone),
// covering the spec.md Validation Gate scenarios:
//   - ready verdict advances quorum (single-actor quorum -> stage completes ->
//     instance approved -> document approved)
//   - request_changes transitions instance -> changes_requested and
//     document -> draft
//   - comment required for request_changes (contract-level; unit-tested in
//     domain/review_verdict_test.go, re-asserted here at the domain boundary)
//   - wrong-stage-kind (approval-kind active stage) rejected
//   - SoD blocks author self-verdict
//   - cancel-with-reason persists to approval_instances.cancel_reason
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// ctxFor builds a context carrying the given actor's identity, tenant-scoped
// to this fixture, for a SUT call. Every RecordVerdict/CancelInstance/
// PinFrozenHash call in this package runs through a runner.Do (TxRunner),
// which seeds the metaldocs.tenant_id/actor_id GUCs from ctx — an unseeded
// context.Background() leaves them unset and authz.Require fails closed with
// "actor_id GUC not set on transaction". Each call site must state which
// actor it intends to act as (author vs reviewer) rather than collapsing to
// one shared ctx — the SoD test specifically depends on the author acting as
// both submitter and verdict-recorder.
func (fx reviewVerdictFixture) ctxFor(actorID string) context.Context {
	return testdb.AuthzCtx(fx.tenantID, actorID)
}

// reviewVerdictFixture bundles the seeded rows a RecordVerdict call needs.
type reviewVerdictFixture struct {
	tenantID   string
	authorID   string
	reviewerID string
	documentID string
	instanceID string
	stageID    string
	runner     db.TxRunner
	svc        *application.ReviewVerdictService
}

// seedReviewVerdictFixture builds: tenant, author + reviewer users, a
// document at status=under_review (process_area_code_snapshot set so
// LoadDocumentAreaCode resolves without a controlled-document read-port), an
// in_progress approval_instance, and a single active stage_instance of the
// given kind. Mirrors eligibility_test.go's raw-SQL FK-chain seeding (no
// factory builder yet exists for a review-kind stage_instance).
func seedReviewVerdictFixture(t *testing.T, database *sql.DB, stageKind domain.StageKind) reviewVerdictFixture {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"))
	reviewer := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Reviewer"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)

	// Stamp process_area_code_snapshot so LoadDocumentAreaCode resolves the
	// area without needing a controlled-document field-reader wired. The
	// documents UPDATE tripwire (migration 0275) requires an asserted
	// capability; SeedWithCaps asserts document.edit tx-locally on the same
	// connection so the guarded raw write is sanctioned.
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET process_area_code_snapshot = 'qa' WHERE id = $1::uuid`,
			doc.ID)
		return err
	})

	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))

	instance := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)

	// Grant chain (production authz.Require, authz.go:144): actor ->
	// metaldocs.user_process_areas(role) -> JOIN role_capabilities(role,
	// capability) for area 'qa' (matches area_code_snapshot below). Every
	// RecordVerdict/CancelInstance call in this file's tests runs through
	// authz.Require BEFORE eligibility/SoD, so both actors need a
	// capability-bearing role in area QA, not just eligible_actor_ids
	// membership:
	//   - reviewer -> role 'approver': carries approval.review (RecordVerdict's
	//     tier-2 gate) AND document.edit (the same call's downstream
	//     instance-approved / request_changes document-transition gate,
	//     review_verdict_service.go:267,342).
	//   - author -> role 'area_admin': carries document.edit (CancelInstance's
	//     gate, cancel_service.go:88) AND approval.review. The author needs
	//     approval.review too so TestReviewVerdict_SoDBlocksSelfVerdict reaches
	//     the domain.CheckSoD block (sod.go) instead of failing earlier on a
	//     plain capability denial — granting the role does NOT bypass SoD,
	//     which is a separate same-actor check keyed off instance.SubmittedBy,
	//     not off role/capability. No role carries BOTH 'author' and
	//     'approver' simultaneously (ux_user_process_areas_one_active allows
	//     only one active row per (user, tenant, area)), so 'area_admin' is
	//     used as the single role satisfying both cap needs for this actor
	//     across every test in this file that shares the fixture.
	testdb.SeedWithCaps(t, database, `[{"cap":"membership.manage"},{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		// user_process_areas.(tenant_id, area_code) FK-references
		// metaldocs.document_process_areas(tenant_id, code), whose code column
		// carries CHECK area_code_format (^[a-z][a-z0-9_-]{1,63}$ — lowercase
		// only). 'qa' is not the tenant's auto-generated taxonomy area
		// (NewTaxonomy/NewDocument mint a random per-test code); it must be
		// seeded explicitly (idempotent, ON CONFLICT DO NOTHING mirrors
		// NewTaxonomy's own seeding) or the grants below 23503-fail. Lowercase
		// 'qa' (not 'QA') to satisfy area_code_format — the documents/stage
		// snapshot columns above have no such CHECK but must still match this
		// value exactly for authz.Require's `upa.area_code = $2` join.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ('qa', $1::uuid, 'QA') ON CONFLICT (tenant_id, code) DO NOTHING`,
			tenant.ID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,$3,$4,now(),NULL,$1)`, author.ID, tenant.ID, "qa", "area_admin")
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,$3,$4,now(),NULL,$1)`, reviewer.ID, tenant.ID, "qa", "approver")
		return err
	})

	// eligible_actor_ids includes BOTH actors: the author must be a member of
	// the eligible pool for TestReviewVerdict_SoDBlocksSelfVerdict to reach
	// the SoD check (ResolveEligibleIdentity/eligibility.go runs BEFORE
	// domain.CheckSoD in review_verdict_service.go — an author absent from
	// this pool would fail with ErrActorNotEligible first, never reaching
	// ErrAuthorCannotSign). Widening the pool does not affect the other
	// tests: quorum_snapshot is any_1_of (quorum.go), which only counts cast
	// votes, never pool size.
	eligible := `["` + author.ID + `","` + reviewer.ID + `"]`
	var stageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 1, 'Review Stage', 'reviewer', 'approval.review', 'qa',
		        'any_1_of', 'keep_snapshot', $2::jsonb, 'active', $3)
		RETURNING id::text`,
		instance.ID, eligible, string(stageKind),
	).Scan(&stageID); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iampostgres.NewUserDisplayNameRepository(database))
	runner := db.NewTxRunner(database)
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)
	services.WithReleaseRecorder(fakeReleaseRecorder{})

	return reviewVerdictFixture{
		tenantID:   tenant.ID,
		authorID:   author.ID,
		reviewerID: reviewer.ID,
		documentID: doc.ID,
		instanceID: instance.ID,
		stageID:    stageID,
		runner:     runner,
		svc:        services.ReviewVerdict,
	}
}

// reviewThenApprovalFixture bundles the seeded rows for R5 (unit 2.3 G3)
// fast-forward eligibility tests: a review-kind stage (active, any_1_of)
// followed by a pending approval-kind stage whose eligible_actor_ids is
// caller-configurable, so the same reviewer can be seeded either in or out
// of the approval pool.
type reviewThenApprovalFixture struct {
	tenantID        string
	authorID        string
	reviewerID      string
	documentID      string
	instanceID      string
	reviewStageID   string
	approvalStageID string
	runner          db.TxRunner
	svc             *application.ReviewVerdictService
}

func (fx reviewThenApprovalFixture) ctxFor(actorID string) context.Context {
	return testdb.AuthzCtx(fx.tenantID, actorID)
}

// seedReviewThenApprovalFixture mirrors seedReviewVerdictFixture's seeding
// but inserts TWO stage_instances: stage 1 (review-kind, active, any_1_of,
// eligible = author+reviewer per the same SoD-pool-widening rationale) and
// stage 2 (approval-kind, pending, all_of, eligible = approvalEligible —
// the caller decides whether the reviewer is in this pool).
func seedReviewThenApprovalFixture(t *testing.T, database *sql.DB, approvalEligible []string) reviewThenApprovalFixture {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"))
	reviewer := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Reviewer"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)

	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET process_area_code_snapshot = 'qa' WHERE id = $1::uuid`,
			doc.ID)
		return err
	})

	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))

	instance := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)

	testdb.SeedWithCaps(t, database, `[{"cap":"membership.manage"},{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ('qa', $1::uuid, 'QA') ON CONFLICT (tenant_id, code) DO NOTHING`,
			tenant.ID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,$3,$4,now(),NULL,$1)`, author.ID, tenant.ID, "qa", "area_admin")
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,$3,$4,now(),NULL,$1)`, reviewer.ID, tenant.ID, "qa", "approver")
		return err
	})

	reviewEligible := `["` + author.ID + `","` + reviewer.ID + `"]`
	var reviewStageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 1, 'Review Stage', 'reviewer', 'approval.review', 'qa',
		        'any_1_of', 'keep_snapshot', $2::jsonb, 'active', 'review')
		RETURNING id::text`,
		instance.ID, reviewEligible,
	).Scan(&reviewStageID); err != nil {
		t.Fatalf("seed review stage_instance: %v", err)
	}

	approvalEligibleJSON := "["
	for i, id := range approvalEligible {
		if i > 0 {
			approvalEligibleJSON += ","
		}
		approvalEligibleJSON += `"` + id + `"`
	}
	approvalEligibleJSON += "]"
	var approvalStageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 2, 'Approval Stage', 'approver', 'document.signoff', 'qa',
		        'all_of', 'keep_snapshot', $2::jsonb, 'pending', 'approval')
		RETURNING id::text`,
		instance.ID, approvalEligibleJSON,
	).Scan(&approvalStageID); err != nil {
		t.Fatalf("seed approval stage_instance: %v", err)
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iampostgres.NewUserDisplayNameRepository(database))
	runner := db.NewTxRunner(database)
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)
	services.WithReleaseRecorder(fakeReleaseRecorder{})

	return reviewThenApprovalFixture{
		tenantID:        tenant.ID,
		authorID:        author.ID,
		reviewerID:      reviewer.ID,
		documentID:      doc.ID,
		instanceID:      instance.ID,
		reviewStageID:   reviewStageID,
		approvalStageID: approvalStageID,
		runner:          runner,
		svc:             services.ReviewVerdict,
	}
}

// TestReviewVerdict_FastForwardEligible_ReviewerAlsoInApprovalPool (R5, unit
// 2.3 G3): the reviewer's `ready` verdict completes the (only-reviewer,
// any_1_of) review stage; the reviewer is ALSO in the now-active approval
// stage's eligible pool, so RecordVerdict reports FastForwardEligible=true
// with NextStageID pointing at the approval stage.
func TestReviewVerdict_FastForwardEligible_ReviewerAlsoInApprovalPool(t *testing.T) {
	database, _ := testdb.Open(t)
	// Fixture seeds the reviewer first; use a placeholder pool then widen once
	// the reviewer's id is known — mirrors the two-step user-then-grant
	// pattern used elsewhere in this fixture file.
	fx := seedReviewThenApprovalFixture(t, database, nil)
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+fx.reviewerID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("widen approval stage eligible pool: %v", err)
	}

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.reviewStageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict(ready): %v", err)
	}
	if !result.StageCompleted {
		t.Errorf("StageCompleted = false; want true")
	}
	if result.InstanceApproved {
		t.Errorf("InstanceApproved = true; want false (a pending approval stage remains)")
	}
	if !result.FastForwardEligible {
		t.Errorf("FastForwardEligible = false; want true (reviewer is also eligible on the now-active approval stage)")
	}
	if result.NextStageID == nil || *result.NextStageID != fx.approvalStageID {
		t.Errorf("NextStageID = %v; want %q", result.NextStageID, fx.approvalStageID)
	}
}

// TestReviewVerdict_FastForwardIneligible_ReviewerNotInApprovalPool (R5,
// unit 2.3 G3): same setup, but the reviewer is NOT in the approval stage's
// eligible pool — RecordVerdict still completes the review stage but reports
// FastForwardEligible=false.
func TestReviewVerdict_FastForwardIneligible_ReviewerNotInApprovalPool(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	otherApprover := testdb.NewUser(t, database, testdb.WithTenant(fx.tenantID), testdb.WithDisplayName("Other Approver"))
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+otherApprover.ID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("set approval stage eligible pool: %v", err)
	}

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.reviewStageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict(ready): %v", err)
	}
	if !result.StageCompleted {
		t.Errorf("StageCompleted = false; want true")
	}
	if result.FastForwardEligible {
		t.Errorf("FastForwardEligible = true; want false (reviewer is not eligible on the now-active approval stage)")
	}
	if result.NextStageID != nil {
		t.Errorf("NextStageID = %v; want nil", result.NextStageID)
	}
}

func documentStatus(t *testing.T, database *sql.DB, docID string) string {
	t.Helper()
	var status string
	if err := database.QueryRowContext(context.Background(),
		`SELECT status FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&status); err != nil {
		t.Fatalf("query document status: %v", err)
	}
	return status
}

func instanceStatus(t *testing.T, database *sql.DB, instanceID string) string {
	t.Helper()
	var status string
	if err := database.QueryRowContext(context.Background(),
		`SELECT status FROM public.approval_instances WHERE id = $1::uuid`, instanceID,
	).Scan(&status); err != nil {
		t.Fatalf("query instance status: %v", err)
	}
	return status
}

// TestReviewVerdict_ReadyAdvancesQuorum: a single reviewer's `ready` verdict
// on a 1-of-1-quorum stage completes the (only) stage, approves the
// instance, and transitions the document under_review -> approved.
func TestReviewVerdict_ReadyAdvancesQuorum(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict(ready): %v", err)
	}
	if !result.StageCompleted {
		t.Errorf("StageCompleted = false; want true")
	}
	if !result.InstanceApproved {
		t.Errorf("InstanceApproved = false; want true (single-stage 1-of-1 quorum)")
	}

	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceApproved) {
		t.Errorf("instance status = %q; want %q", got, domain.InstanceApproved)
	}
	if got := documentStatus(t, database, fx.documentID); got != "approved" {
		t.Errorf("document status = %q; want approved", got)
	}
}

// TestReviewVerdict_RequestChangesCollapsesInstanceAndDocument: a single
// request_changes verdict immediately collapses the instance to
// changes_requested and reverts the document to draft — no quorum required.
func TestReviewVerdict_RequestChangesCollapsesInstanceAndDocument(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictRequestChanges,
		Comment:         "please fix section 3",
	})
	if err != nil {
		t.Fatalf("RecordVerdict(request_changes): %v", err)
	}
	if !result.ChangesRequested {
		t.Errorf("ChangesRequested = false; want true")
	}

	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceChangesRequested) {
		t.Errorf("instance status = %q; want %q", got, domain.InstanceChangesRequested)
	}
	if got := documentStatus(t, database, fx.documentID); got != "draft" {
		t.Errorf("document status = %q; want draft", got)
	}
}

// TestReviewVerdict_RequestChangesRequiresComment asserts the domain-level
// comment requirement (spec.md Validation Gate row) surfaces through the
// full service call when Comment is empty — RecordVerdict's domain.NewVerdict
// call returns ErrVerdictCommentRequired before any row is written.
func TestReviewVerdict_RequestChangesRequiresComment(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	_, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictRequestChanges,
		Comment:         "",
	})
	if !errors.Is(err, domain.ErrVerdictCommentRequired) {
		t.Fatalf("err = %v; want domain.ErrVerdictCommentRequired", err)
	}

	// No side effects: instance must remain in_progress.
	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceInProgress) {
		t.Errorf("instance status = %q; want in_progress (no mutation on validation failure)", got)
	}
}

// TestReviewVerdict_ReadyOnApprovalStageRejected asserts a `ready` verdict
// against an approval-kind active stage is rejected with
// ErrVerdictReadyOnApprovalStage before any eligibility/SoD work (R3/G2:
// approval stages accept only request_changes; `ready` would bypass the
// e-signature ceremony).
func TestReviewVerdict_ReadyOnApprovalStageRejected(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindApproval)

	_, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if !errors.Is(err, domain.ErrVerdictReadyOnApprovalStage) {
		t.Fatalf("err = %v; want domain.ErrVerdictReadyOnApprovalStage", err)
	}

	// No side effects: instance must remain in_progress.
	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceInProgress) {
		t.Errorf("instance status = %q; want in_progress (no mutation on validation failure)", got)
	}
}

// TestReviewVerdict_RequestChangesOnApprovalStageThawsToDraft asserts a
// request_changes verdict against an approval-kind active stage (R3/G2: the
// power to converse/return without signing) collapses the instance to
// changes_requested and reverts the document to draft — mirrors
// TestReviewVerdict_RequestChangesCollapsesInstanceAndDocument for a
// review-kind stage.
func TestReviewVerdict_RequestChangesOnApprovalStageThawsToDraft(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindApproval)

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictRequestChanges,
		Comment:         "return for edits",
	})
	if err != nil {
		t.Fatalf("RecordVerdict(request_changes): %v", err)
	}
	if !result.ChangesRequested {
		t.Errorf("ChangesRequested = false; want true")
	}

	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceChangesRequested) {
		t.Errorf("instance status = %q; want %q", got, domain.InstanceChangesRequested)
	}
	if got := documentStatus(t, database, fx.documentID); got != "draft" {
		t.Errorf("document status = %q; want draft", got)
	}
}

// TestReviewVerdict_SoDBlocksSelfVerdict asserts the document's submitter
// (author) cannot record a verdict on their own submission.
func TestReviewVerdict_SoDBlocksSelfVerdict(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	_, err := fx.svc.RecordVerdict(fx.ctxFor(fx.authorID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.authorID,
		Verdict:         domain.VerdictReady,
	})
	if !errors.Is(err, domain.ErrAuthorCannotSign) {
		t.Fatalf("err = %v; want domain.ErrAuthorCannotSign", err)
	}
}

// TestReviewVerdictSoDTrigger_RejectsAuthorSelfInsert asserts the F7
// DB-tripwire-last-line SoD trigger (migration 0290, trg_review_verdict_sod /
// enforce_approval_sod) rejects a raw INSERT into approval_review_verdicts
// where actor_user_id equals the document's author — the same rule
// domain.CheckSoD already enforces at the app layer (defense in depth: a
// direct-SQL bypass of RecordVerdict must still be blocked at the DB).
// Mirrors the existing (undocumented-in-Go, DB-only) enforce_signoff_sod
// trigger's behavior for approval_signoffs, now symmetric for review verdicts.
func TestReviewVerdictSoDTrigger_RejectsAuthorSelfInsert(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	_, err := database.ExecContext(context.Background(), `
		INSERT INTO public.approval_review_verdicts
		  (approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id, verdict, comment)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid, 'ready', NULL)`,
		fx.instanceID, fx.stageID, fx.authorID, fx.tenantID,
	)
	if err == nil {
		t.Fatal("direct SQL insert with actor_user_id = document author should be rejected by trg_review_verdict_sod; got no error")
	}
	if !strings.Contains(err.Error(), "SoD: author cannot sign own revision") {
		t.Errorf("err = %v; want SoD trigger rejection message", err)
	}
}

// TestCancelInstance_ReasonPersists exercises the F4 cancel-reason
// persistence gap fix: CancelInstance must now write the reason to
// approval_instances.cancel_reason (previously only reached the governance
// event, never the row itself).
func TestCancelInstance_ReasonPersists(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	repo := infrastructure.NewPostgresApprovalRepository(database, iampostgres.NewUserDisplayNameRepository(database))
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	const reason = "stakeholder withdrew (integration)"
	_, err := services.Cancel.CancelInstance(fx.ctxFor(fx.authorID), fx.runner, application.CancelInput{
		TenantID:    fx.tenantID,
		InstanceID:  fx.instanceID,
		ActorUserID: fx.authorID,
		Reason:      reason,
	})
	if err != nil {
		t.Fatalf("CancelInstance: %v", err)
	}

	var gotReason sql.NullString
	if err := database.QueryRowContext(context.Background(),
		`SELECT cancel_reason FROM public.approval_instances WHERE id = $1::uuid`, fx.instanceID,
	).Scan(&gotReason); err != nil {
		t.Fatalf("query cancel_reason: %v", err)
	}
	if !gotReason.Valid || gotReason.String != reason {
		t.Errorf("cancel_reason = %v; want %q", gotReason, reason)
	}
}
