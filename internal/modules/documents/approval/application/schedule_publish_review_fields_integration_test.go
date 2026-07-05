//go:build integration
// +build integration

package application

// M6 F6.2 — real-DB integration proof that effective_to + the initial
// review_due_at are wired onto the schedule-publish path (validation-contract.md
// §0.1/§2, spec.md interview #3): "effective_to (expiry) + initial
// review_due_at are set on the same schedule/publish request — new optional
// request fields on the existing schedule/publish op". Mirrors
// publish_no_autoversion_integration_test.go's real-DB setup for
// PublishService.SchedulePublish.

import (
	"database/sql"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// TestSchedulePublish_PersistsEffectiveToAndReviewDueAt proves the
// approved->scheduled OCC UPDATE (publish_service.go SchedulePublish) now
// also writes effective_to and review_due_at when the caller supplies them,
// alongside the pre-existing effective_from write.
func TestSchedulePublish_PersistsEffectiveToAndReviewDueAt(t *testing.T) {
	database, _ := testdb.Open(t)
	database.SetMaxOpenConns(4)

	tnt := testdb.NewTenant(t, database)
	user := testdb.NewUser(t, database, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	// SchedulePublish issues in-tx authz.Require (document.publish/edit); TxRunner's
	// chokepoint seeds the tenant_id + actor_id GUCs authz reads ONLY when BOTH are
	// on ctx (runner.go seedTxIdentityFromContext) — the production authn middleware
	// sets them, so mirror it here (same helper the F6.3 submit proof uses).
	ctx := submitCtxWithIdentity(tnt.ID, user.ID)

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(user.ID),
		testdb.WithStatus("approved"),
	)

	inst := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithStatus("approved"),
	)

	repo := repository.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	emitter := NewSQLEmitter()
	clock := RealClock{}
	cdRead := controlleddocumentsdomain.NoopCDFieldReader{}

	svcs := NewServices(repo, emitter, clock, cdRead)
	runner := db.NewTxRunner(database)

	effectiveFrom := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	effectiveTo := effectiveFrom.Add(365 * 24 * time.Hour)
	reviewDueAt := effectiveFrom.Add(180 * 24 * time.Hour)

	res, err := svcs.Publish.SchedulePublish(ctx, runner, SchedulePublishRequest{
		TenantID:      tnt.ID,
		InstanceID:    inst.ID,
		EffectiveDate: effectiveFrom,
		ScheduledBy:   user.ID,
		EffectiveTo:   &effectiveTo,
		ReviewDueAt:   &reviewDueAt,
	})
	if err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}
	if res.DocumentID != doc.ID {
		t.Fatalf("DocumentID = %q, want %q", res.DocumentID, doc.ID)
	}

	var gotStatus string
	var gotEffectiveFrom, gotEffectiveTo, gotReviewDueAt sql.NullTime
	if err := database.QueryRowContext(ctx,
		`SELECT status, effective_from, effective_to, review_due_at
		   FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ID, tnt.ID,
	).Scan(&gotStatus, &gotEffectiveFrom, &gotEffectiveTo, &gotReviewDueAt); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotStatus != "scheduled" {
		t.Fatalf("documents.status = %q, want scheduled", gotStatus)
	}
	if !gotEffectiveFrom.Valid || !gotEffectiveFrom.Time.Equal(effectiveFrom) {
		t.Fatalf("documents.effective_from = %v, want %v", gotEffectiveFrom.Time, effectiveFrom)
	}
	if !gotEffectiveTo.Valid || !gotEffectiveTo.Time.Equal(effectiveTo) {
		t.Fatalf("documents.effective_to = %v, want %v (F6.2: previously unwritten, now wired)", gotEffectiveTo.Time, effectiveTo)
	}
	if !gotReviewDueAt.Valid || !gotReviewDueAt.Time.Equal(reviewDueAt) {
		t.Fatalf("documents.review_due_at = %v, want %v (F6.2: initial review-due set at schedule time)", gotReviewDueAt.Time, reviewDueAt)
	}
}

// TestSchedulePublish_EffectiveToAndReviewDueAtOptional proves the pre-existing
// behavior is preserved when the new fields are omitted (nil): both columns
// stay NULL, no regression to the F4/F4.5 schedule-publish path.
func TestSchedulePublish_EffectiveToAndReviewDueAtOptional(t *testing.T) {
	database, _ := testdb.Open(t)
	database.SetMaxOpenConns(4)

	tnt := testdb.NewTenant(t, database)
	user := testdb.NewUser(t, database, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	// See sibling test: seed identity on ctx so TxRunner seeds the actor_id GUC
	// authz.Require needs (production authn middleware does this).
	ctx := submitCtxWithIdentity(tnt.ID, user.ID)

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(user.ID),
		testdb.WithStatus("approved"),
	)

	inst := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithStatus("approved"),
	)

	repo := repository.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	svcs := NewServices(repo, NewSQLEmitter(), RealClock{}, controlleddocumentsdomain.NoopCDFieldReader{})
	runner := db.NewTxRunner(database)

	effectiveFrom := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	if _, err := svcs.Publish.SchedulePublish(ctx, runner, SchedulePublishRequest{
		TenantID:      tnt.ID,
		InstanceID:    inst.ID,
		EffectiveDate: effectiveFrom,
		ScheduledBy:   user.ID,
	}); err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}

	var gotEffectiveTo, gotReviewDueAt sql.NullTime
	if err := database.QueryRowContext(ctx,
		`SELECT effective_to, review_due_at FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ID, tnt.ID,
	).Scan(&gotEffectiveTo, &gotReviewDueAt); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotEffectiveTo.Valid {
		t.Fatalf("documents.effective_to = %v, want NULL when not supplied", gotEffectiveTo.Time)
	}
	if gotReviewDueAt.Valid {
		t.Fatalf("documents.review_due_at = %v, want NULL when not supplied", gotReviewDueAt.Time)
	}
}
