//go:build integration
// +build integration

// Package approval_test — F5 (milestone 2b, approval-kernel-backend) freeze
// boundary. Exercises the freeze executor end-to-end against a real Postgres
// instance (testdb per-test clone), covering the spec.md Validation Gate
// scenarios:
//   - freeze fires and persists approval_instances.frozen_content_hash when
//     the last review-kind stage completes into an approval-kind stage
//     (ReviewVerdictService.RecordVerdict's ready path)
//   - freeze fires at submit time for approval-only routes (no review stage)
//   - a pre-instance comment (created before submitted_at) does NOT block
//     freeze — proves the instance-scoped comment check, not the old
//     document-wide HasUnresolvedComments
//   - a during-instance comment (created after submitted_at, unresolved) DOES
//     block freeze
//   - two sequential freeze triggers on the same instance are idempotent
//     (hash unchanged, no error) — re-entry via RecordVerdict's own repeated
//     stage-advance path is not reachable from the service API twice, so this
//     is asserted directly against the executeFreeze-exercising service call
//     plus a manual repeat PinFrozenHash call through the repository.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// insertUnresolvedComment inserts a minimal document_comments row at the
// given createdAt timestamp, unresolved (resolved_at NULL). Mirrors the
// module's raw-SQL FK-chain seeding pattern (no factory helper exists for
// this table yet).
func insertUnresolvedComment(t *testing.T, database *sql.DB, tenantID, documentID, authorID string, createdAt time.Time) {
	t.Helper()
	if _, err := database.ExecContext(context.Background(), `
		INSERT INTO public.document_comments
		  (tenant_id, document_id, library_comment_id, author_id, author_display, content_json, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, 1, $3, 'Test Author', '{"text":"needs a look"}'::jsonb, $4, $4)`,
		tenantID, documentID, authorID, createdAt,
	); err != nil {
		t.Fatalf("seed document_comments: %v", err)
	}
}

func frozenContentHash(t *testing.T, database *sql.DB, instanceID string) sql.NullString {
	t.Helper()
	var hash sql.NullString
	if err := database.QueryRowContext(context.Background(),
		`SELECT frozen_content_hash FROM public.approval_instances WHERE id = $1::uuid`, instanceID,
	).Scan(&hash); err != nil {
		t.Fatalf("query frozen_content_hash: %v", err)
	}
	return hash
}

// TestFreeze_ReviewToApprovalTransition_PinsHash: a single reviewer's `ready`
// verdict completing the ONLY review-kind stage, with no approval-kind stage
// following (instance goes straight to InstanceApproved), crosses the freeze
// boundary and pins frozen_content_hash.
func TestFreeze_ReviewToApprovalTransition_PinsHash(t *testing.T) {
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
	if !result.InstanceApproved {
		t.Fatalf("expected instance approved (single review stage, no approval stage), got %+v", result)
	}

	hash := frozenContentHash(t, database, fx.instanceID)
	if !hash.Valid || hash.String == "" {
		t.Fatalf("expected frozen_content_hash pinned after freeze, got %+v", hash)
	}
}

// TestFreeze_PreInstanceComment_DoesNotBlock: a comment created BEFORE the
// instance's submitted_at (stale, from a prior revision's review) must not
// block freeze — proves the instance-scoped check
// (HasUnresolvedInstanceComments) is used, not the old document-wide
// HasUnresolvedComments which would have blocked forever.
func TestFreeze_PreInstanceComment_DoesNotBlock(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	// Seed a comment created well before the instance existed (i.e. before
	// its submitted_at).
	staleTime := time.Now().Add(-24 * time.Hour)
	insertUnresolvedComment(t, database, fx.tenantID, fx.documentID, fx.authorID, staleTime)

	result, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict(ready) should not be blocked by a pre-instance comment: %v", err)
	}
	if !result.InstanceApproved {
		t.Fatalf("expected instance approved, got %+v", result)
	}

	hash := frozenContentHash(t, database, fx.instanceID)
	if !hash.Valid || hash.String == "" {
		t.Fatalf("expected frozen_content_hash pinned despite the pre-instance comment, got %+v", hash)
	}
}

// TestFreeze_DuringInstanceComment_Blocks: a comment created AFTER the
// instance's submitted_at, still unresolved, DOES block freeze —
// ErrFreezeBlockedByUnresolvedComments propagates and no hash is pinned.
func TestFreeze_DuringInstanceComment_Blocks(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReviewVerdictFixture(t, database, domain.StageKindReview)

	// Comment created after the fixture's instance (seeded moments ago) —
	// well within its review window.
	insertUnresolvedComment(t, database, fx.tenantID, fx.documentID, fx.reviewerID, time.Now())

	_, err := fx.svc.RecordVerdict(fx.ctxFor(fx.reviewerID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.reviewerID,
		Verdict:         domain.VerdictReady,
	})
	if !errors.Is(err, application.ErrFreezeBlockedByUnresolvedComments) {
		t.Fatalf("err = %v; want application.ErrFreezeBlockedByUnresolvedComments", err)
	}

	// The whole stage-transition tx (including UpdateStageStatus/AdvanceStage
	// persistence) must have rolled back: instance stays in_progress.
	if got := instanceStatus(t, database, fx.instanceID); got != string(domain.InstanceInProgress) {
		t.Errorf("instance status = %q; want in_progress (freeze block rolls back the whole tx)", got)
	}

	hash := frozenContentHash(t, database, fx.instanceID)
	if hash.Valid {
		t.Fatalf("expected frozen_content_hash to remain NULL when freeze is blocked, got %q", hash.String)
	}
}

// TestFreeze_SubmitApprovalOnlyRoute_PinsHashAtSubmit: a route with only
// approval-kind stages (no review stage) freezes at submit time — the sole
// freeze point for such routes, since there is no review->approval
// transition for ReviewVerdictService to hook into.
func TestFreeze_SubmitApprovalOnlyRoute_PinsHashAtSubmit(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	// system_admin bypasses the document.submit capability check (ADR 0022
	// Phase 11 F8), mirroring submit_service_reason_integration_test.go's
	// pattern — the GUC-seeded actor still needs SOME identity/authz path,
	// and this test isn't exercising capability-grant plumbing.
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"), testdb.WithRole("system_admin"))
	approver := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Approver"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("draft"),
		testdb.WithSubmitReadySnapshots(),
	)
	// public.user_process_areas.area_code carries an FK to
	// metaldocs.document_process_areas(tenant_id, code), and that table's
	// area_code_format CHECK requires a lowercase code — seed the 'qa' area
	// row for this tenant before anything references it below.
	testdb.SeedWithCaps(t, database, `[{"cap":"taxonomy.manage"},{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ('qa', $1::uuid, 'qa') ON CONFLICT (tenant_id, code) DO NOTHING`,
			tenant.ID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET process_area_code_snapshot = 'qa' WHERE id = $1::uuid`,
			doc.ID,
		)
		return err
	})

	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))
	// Seed a single approval-kind stage on the route (stage_kind defaults to
	// 'approval' per migration 0286; route.Stages is loaded by LoadRoute).
	if _, err := database.ExecContext(ctx, `
		INSERT INTO public.approval_route_stages
		  (route_id, stage_order, name, required_role, required_capability,
		   area_code, quorum, on_eligibility_drift, stage_kind)
		VALUES ($1::uuid, 1, 'Approval Stage', 'approver', 'document.signoff', 'qa',
		        'any_1_of', 'keep_snapshot', 'approval')`,
		route.ID,
	); err != nil {
		t.Fatalf("seed approval_route_stages: %v", err)
	}

	// ResolveEligibleActors (called by SubmitRevisionForReview) reads
	// metaldocs.v_active_user_areas, backed by public.user_process_areas —
	// the approver must have an active area-membership row whose `role`
	// matches the stage's required_role ('approver') in area 'qa'.
	testdb.SeedWithCaps(t, database, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, 'qa', 'approver', now() - interval '1 hour')`,
			approver.ID, tenant.ID,
		)
		return err
	})

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	// authz.Require needs the metaldocs.tenant_id / actor_id GUCs seeded on
	// the tx; TxRunner.seedTxIdentityFromContext pulls them from ctx.
	authzCtx := testdb.AuthzCtx(tenant.ID, author.ID)
	result, err := services.Submit.SubmitRevisionForReview(authzCtx, runner, application.SubmitRequest{
		TenantID:        tenant.ID,
		DocumentID:      doc.ID,
		RouteID:         route.ID,
		SubmittedBy:     author.ID,
		RevisionTitle:   "Initial",
		ContentFormData: map[string]any{"title": "Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "freeze-submit-approval-only-1",
	})
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: %v", err)
	}

	hash := frozenContentHash(t, database, result.InstanceID)
	if !hash.Valid || hash.String == "" {
		t.Fatalf("expected frozen_content_hash pinned at submit for an approval-only route, got %+v", hash)
	}
}

// TestFreeze_Idempotent_ReentryIsNoop: calling PinFrozenHash a second time on
// an already-frozen instance is a no-op (won=false) and the pinned hash is
// unchanged — proves the CAS `WHERE frozen_content_hash IS NULL` guard.
func TestFreeze_Idempotent_ReentryIsNoop(t *testing.T) {
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
	if !result.InstanceApproved {
		t.Fatalf("expected instance approved, got %+v", result)
	}

	firstHash := frozenContentHash(t, database, fx.instanceID)
	if !firstHash.Valid {
		t.Fatalf("expected frozen_content_hash pinned after first freeze")
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	reentryCtx := fx.ctxFor(fx.reviewerID)
	err = fx.runner.Do(reentryCtx, func(tx *sql.Tx) error {
		won, perr := repo.PinFrozenHash(reentryCtx, tx, fx.tenantID, fx.instanceID, "a-different-hash-should-never-win")
		if perr != nil {
			return perr
		}
		if won {
			t.Fatalf("expected won=false on re-entry against an already-frozen instance")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("re-entry PinFrozenHash call: %v", err)
	}

	secondHash := frozenContentHash(t, database, fx.instanceID)
	if secondHash.String != firstHash.String {
		t.Fatalf("hash changed after re-entry: first=%q second=%q", firstHash.String, secondHash.String)
	}
}

