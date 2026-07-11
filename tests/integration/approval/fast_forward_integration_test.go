//go:build integration
// +build integration

// Package approval_test — R5 (unit 2.3 G3) fast-forward ("Aprovar já")
// end-to-end coverage against a real Postgres instance (testdb per-test
// clone). Exercises application.FastForwardService, which composes
// ReviewVerdictService + DecisionService in one transaction: reuses
// seedReviewThenApprovalFixture (review_verdict_integration_test.go) — a
// review-kind active stage (any_1_of) followed by a pending approval-kind
// stage (all_of) whose eligible pool is caller-configurable.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	docsapp "metaldocs/internal/modules/documents/application"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// fakePinInvoker is a no-op application.PinInvoker, mirroring the identical
// fake in internal/modules/documents/approval/application/decision_service_freeze_test.go.
// The signoff leg's async-freeze seam (ADR 0015) requires a PinInvoker to be
// wired before RecordSignoff's instance-approved branch runs; a no-op is
// sufficient here since these tests assert approval's OWN frozen_content_hash
// pin (F5/F6, set by executeFreeze during the verdict leg), not the
// documents-module template/placeholder freeze this seam exists for.
type fakePinInvoker struct{}

func (fakePinInvoker) Pin(_ context.Context, _ db.Tx, _, _ string, _ docsapp.ApproverContext) error {
	return nil
}

// validFastForwardContentHash matches testdb.NewApprovalInstance's hardcoded
// content_hash_at_submit = repeat('a', 64), which executeFreeze pins as
// frozen_content_hash on the review->approval stage transition (F5/F6).
// Mirrors application/decision_service_test.go's validContentHash constant.
const validFastForwardContentHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// fastForwardSvc builds a fresh *application.FastForwardService sharing the
// same underlying database as fx — mirrors TestCancelInstance_ReasonPersists'
// precedent of constructing a second *application.Services against the same
// *sql.DB to reach a sibling service the fixture's own svc field doesn't
// expose.
func fastForwardSvc(database *sql.DB) *application.FastForwardService {
	repo := infrastructure.NewPostgresApprovalRepository(database, iampostgres.NewUserDisplayNameRepository(database))
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)
	services.Decision.WithPinInvoker(fakePinInvoker{})
	return services.FastForward
}

func reviewVerdictRowCount(t *testing.T, database *sql.DB, instanceID string) int {
	t.Helper()
	var n int
	if err := database.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.approval_review_verdicts WHERE approval_instance_id = $1::uuid`, instanceID,
	).Scan(&n); err != nil {
		t.Fatalf("count approval_review_verdicts: %v", err)
	}
	return n
}

func signoffRowCount(t *testing.T, database *sql.DB, instanceID string) int {
	t.Helper()
	var n int
	if err := database.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.approval_signoffs WHERE approval_instance_id = $1::uuid`, instanceID,
	).Scan(&n); err != nil {
		t.Fatalf("count approval_signoffs: %v", err)
	}
	return n
}

func governanceEventCount(t *testing.T, database *sql.DB, instanceID string) int {
	t.Helper()
	var n int
	if err := database.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.governance_events WHERE resource_id = $1`, instanceID,
	).Scan(&n); err != nil {
		t.Fatalf("count governance_events: %v", err)
	}
	return n
}

func frozenContentHashFF(t *testing.T, database *sql.DB, instanceID string) sql.NullString {
	t.Helper()
	var hash sql.NullString
	if err := database.QueryRowContext(context.Background(),
		`SELECT frozen_content_hash FROM public.approval_instances WHERE id = $1::uuid`, instanceID,
	).Scan(&hash); err != nil {
		t.Fatalf("query frozen_content_hash: %v", err)
	}
	return hash
}

// TestFastForward_HappyPath_ReviewerAlsoApprover: reviewer is in the review
// stage's any_1_of pool AND the approval stage's all_of pool. RecordFastForward
// records exactly one ready verdict + one approve signoff in one transaction,
// freezes the content hash, approves the instance, and transitions the
// document to approved.
func TestFastForward_HappyPath_ReviewerAlsoApprover(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+fx.reviewerID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("widen approval stage eligible pool: %v", err)
	}

	beforeVerdicts := reviewVerdictRowCount(t, database, fx.instanceID)
	beforeSignoffs := signoffRowCount(t, database, fx.instanceID)
	beforeEvents := governanceEventCount(t, database, fx.instanceID)

	svc := fastForwardSvc(database)
	result, err := svc.RecordFastForward(fx.ctxFor(fx.reviewerID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.reviewStageID,
		ActorUserID:      fx.reviewerID,
		Comment:          "fast forward",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      validFastForwardContentHash,
	})
	if err != nil {
		t.Fatalf("RecordFastForward: %v", err)
	}
	if !result.Verdict.StageCompleted {
		t.Errorf("Verdict.StageCompleted = false; want true")
	}
	if !result.Signoff.InstanceApproved {
		t.Errorf("Signoff.InstanceApproved = false; want true")
	}

	if got := reviewVerdictRowCount(t, database, fx.instanceID); got != beforeVerdicts+1 {
		t.Errorf("approval_review_verdicts rows = %d; want %d (exactly 1 new row)", got, beforeVerdicts+1)
	}
	if got := signoffRowCount(t, database, fx.instanceID); got != beforeSignoffs+1 {
		t.Errorf("approval_signoffs rows = %d; want %d (exactly 1 new row)", got, beforeSignoffs+1)
	}
	if got := governanceEventCount(t, database, fx.instanceID); got != beforeEvents+2 {
		t.Errorf("governance_events rows = %d; want %d (verdict + signoff events)", got, beforeEvents+2)
	}

	hash := frozenContentHashFF(t, database, fx.instanceID)
	if !hash.Valid || hash.String != validFastForwardContentHash {
		t.Errorf("frozen_content_hash = %v; want %q", hash, validFastForwardContentHash)
	}
	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceApproved) {
		t.Errorf("instance status = %q; want %q", got, domain.InstanceApproved)
	}
	if got := documentStatus(t, database, fx.documentID); got != "approved" {
		t.Errorf("document status = %q; want approved", got)
	}
}

// TestFastForward_NotEligible_ReviewerNotInApprovalPool: the reviewer
// completes the review stage but is NOT in the now-active approval stage's
// eligible pool — RecordFastForward fails with ErrFastForwardNotEligible and
// rolls back BOTH legs atomically (zero new rows in either ledger table).
func TestFastForward_NotEligible_ReviewerNotInApprovalPool(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	otherApprover := testdb.NewUser(t, database, testdb.WithTenant(fx.tenantID), testdb.WithDisplayName("Other Approver"))
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+otherApprover.ID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("set approval stage eligible pool: %v", err)
	}

	beforeVerdicts := reviewVerdictRowCount(t, database, fx.instanceID)
	beforeSignoffs := signoffRowCount(t, database, fx.instanceID)

	svc := fastForwardSvc(database)
	_, err := svc.RecordFastForward(fx.ctxFor(fx.reviewerID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.reviewStageID,
		ActorUserID:      fx.reviewerID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      validFastForwardContentHash,
	})
	if !errors.Is(err, domain.ErrFastForwardNotEligible) {
		t.Fatalf("err = %v; want domain.ErrFastForwardNotEligible", err)
	}

	if got := reviewVerdictRowCount(t, database, fx.instanceID); got != beforeVerdicts {
		t.Errorf("approval_review_verdicts rows = %d; want %d (atomic rollback, zero new rows)", got, beforeVerdicts)
	}
	if got := signoffRowCount(t, database, fx.instanceID); got != beforeSignoffs {
		t.Errorf("approval_signoffs rows = %d; want %d (atomic rollback, zero new rows)", got, beforeSignoffs)
	}
}

// TestFastForward_StageNotCompleted_AllOfQuorumOnlyOneActs seeds the REVIEW
// stage itself with an all_of quorum across two reviewers so a single actor's
// ready verdict records but does not complete the stage — RecordFastForward
// fails with ErrFastForwardStageNotCompleted before ever attempting the
// signoff leg, and rolls back the verdict write too (zero new rows in either
// table; the review verdict row is part of the same failed transaction).
func TestFastForward_StageNotCompleted_AllOfQuorumOnlyOneActs(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	secondReviewer := testdb.NewUser(t, database, testdb.WithTenant(fx.tenantID), testdb.WithDisplayName("Second Reviewer"))

	testdb.SeedWithCaps(t, database, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,$3,$4,now(),NULL,$1)`, secondReviewer.ID, fx.tenantID, "qa", "approver")
		return err
	})

	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET quorum_snapshot = 'all_of', eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+fx.authorID+`","`+fx.reviewerID+`","`+secondReviewer.ID+`"]`, fx.reviewStageID,
	); err != nil {
		t.Fatalf("widen review stage to all_of quorum: %v", err)
	}

	beforeVerdicts := reviewVerdictRowCount(t, database, fx.instanceID)
	beforeSignoffs := signoffRowCount(t, database, fx.instanceID)

	svc := fastForwardSvc(database)
	_, err := svc.RecordFastForward(fx.ctxFor(fx.reviewerID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.reviewStageID,
		ActorUserID:      fx.reviewerID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      validFastForwardContentHash,
	})
	if !errors.Is(err, domain.ErrFastForwardStageNotCompleted) {
		t.Fatalf("err = %v; want domain.ErrFastForwardStageNotCompleted", err)
	}

	if got := reviewVerdictRowCount(t, database, fx.instanceID); got != beforeVerdicts {
		t.Errorf("approval_review_verdicts rows = %d; want %d (atomic rollback, zero new rows)", got, beforeVerdicts)
	}
	if got := signoffRowCount(t, database, fx.instanceID); got != beforeSignoffs {
		t.Errorf("approval_signoffs rows = %d; want %d (atomic rollback, zero new rows)", got, beforeSignoffs)
	}
}

// TestFastForward_G2Regression_ApprovalKindActiveStageRejected: fast-forward
// targeting an approval-kind ACTIVE stage (StageInstanceID = fx.stageID from
// seedReviewVerdictFixture with StageKindApproval) is rejected by the same
// G2 rule ReviewVerdict enforces directly — ErrVerdictReadyOnApprovalStage —
// before any eligibility/SoD work, propagated unchanged through the
// fast-forward composition (recordVerdictInTx is the identical tx-scoped core
// RecordVerdict calls).
func TestFastForward_G2Regression_ApprovalKindActiveStageRejected(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindApproval)

	svc := fastForwardSvc(database)
	_, err := svc.RecordFastForward(fx.ctxFor(fx.reviewerID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.stageID,
		ActorUserID:      fx.reviewerID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      validFastForwardContentHash,
	})
	if !errors.Is(err, domain.ErrVerdictReadyOnApprovalStage) {
		t.Fatalf("err = %v; want domain.ErrVerdictReadyOnApprovalStage", err)
	}

	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceInProgress) {
		t.Errorf("instance status = %q; want in_progress (no mutation on validation failure)", got)
	}
}

// TestFastForward_SoDBlocksAuthorSelfFastForward: the document's author
// cannot fast-forward their own submission — domain.CheckSoD blocks the
// verdict leg with ErrAuthorCannotSign before the signoff leg ever runs.
func TestFastForward_SoDBlocksAuthorSelfFastForward(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+fx.authorID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("widen approval stage eligible pool: %v", err)
	}

	beforeVerdicts := reviewVerdictRowCount(t, database, fx.instanceID)
	beforeSignoffs := signoffRowCount(t, database, fx.instanceID)

	svc := fastForwardSvc(database)
	_, err := svc.RecordFastForward(fx.ctxFor(fx.authorID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.reviewStageID,
		ActorUserID:      fx.authorID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      validFastForwardContentHash,
	})
	if !errors.Is(err, domain.ErrAuthorCannotSign) {
		t.Fatalf("err = %v; want domain.ErrAuthorCannotSign", err)
	}

	if got := reviewVerdictRowCount(t, database, fx.instanceID); got != beforeVerdicts {
		t.Errorf("approval_review_verdicts rows = %d; want %d (atomic rollback, zero new rows)", got, beforeVerdicts)
	}
	if got := signoffRowCount(t, database, fx.instanceID); got != beforeSignoffs {
		t.Errorf("approval_signoffs rows = %d; want %d (atomic rollback, zero new rows)", got, beforeSignoffs)
	}
}

// TestFastForward_ContentHashMismatch_SignoffLegErrorPropagates: the review
// leg completes fine, but a wrong ContentHash makes the signoff leg's
// LoadFrozenContentHash comparison fail with ErrContentHashMismatch — the
// whole transaction (including the already-recorded ready verdict) rolls
// back, so zero new rows land in either ledger table.
func TestFastForward_ContentHashMismatch_SignoffLegErrorPropagates(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewThenApprovalFixture(t, database, nil)
	if _, err := database.ExecContext(context.Background(),
		`UPDATE public.approval_stage_instances SET eligible_actor_ids = $1::jsonb WHERE id = $2::uuid`,
		`["`+fx.reviewerID+`"]`, fx.approvalStageID,
	); err != nil {
		t.Fatalf("widen approval stage eligible pool: %v", err)
	}

	beforeVerdicts := reviewVerdictRowCount(t, database, fx.instanceID)
	beforeSignoffs := signoffRowCount(t, database, fx.instanceID)

	svc := fastForwardSvc(database)
	_, err := svc.RecordFastForward(fx.ctxFor(fx.reviewerID), fx.runner, application.FastForwardRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.reviewStageID,
		ActorUserID:      fx.reviewerID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentHash:      "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	})
	if !errors.Is(err, application.ErrContentHashMismatch) {
		t.Fatalf("err = %v; want application.ErrContentHashMismatch", err)
	}

	if got := reviewVerdictRowCount(t, database, fx.instanceID); got != beforeVerdicts {
		t.Errorf("approval_review_verdicts rows = %d; want %d (atomic rollback, zero new rows)", got, beforeVerdicts)
	}
	if got := signoffRowCount(t, database, fx.instanceID); got != beforeSignoffs {
		t.Errorf("approval_signoffs rows = %d; want %d (atomic rollback, zero new rows)", got, beforeSignoffs)
	}
}
