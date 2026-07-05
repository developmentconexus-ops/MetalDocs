//go:build integration
// +build integration

package application

// M6 F6.2 T5 — real-DB integration proof for MarkReviewedService.MarkReviewed
// (validation-contract.md §3/§4.4, f6.2-periodic-review/spec.md interview #2/#4).
//
// Proves:
//  1. TestMarkReviewed_RequiresDocumentReviewCapability — an actor WITHOUT the
//     document.review grant is denied by authz.Require (a friendly tier-2
//     ErrCapDenied), not a DB tripwire P0001 leak; the same actor WITH the
//     grant (role qms_admin, which curated role_capabilities maps to
//     document.review, db/reference-data/0001_product_reference_data.sql:187)
//     succeeds.
//  2. TestMarkReviewed_SetsReviewDates — happy path: last_reviewed_at ~ now,
//     review_due_at = request value, effective_to set when provided; status
//     stays "published"; revision_version advances by 1.
//  3. TestMarkReviewed_OCCConflict — stale ExpectedRevisionVersion -> 0 rows
//     affected -> ErrMarkReviewedStaleRevision, no write.
//  4. TestMarkReviewed_RejectsNonPublished — a draft document -> friendly
//     ErrDocumentNotPublished, not a silent no-op mistaken for success.
//  5. TestMarkReviewed_TenantIsolation — a cross-tenant document id is
//     invisible (RLS) -> ErrDocumentNotFound, no cross-tenant write.
//
// authz.Require needs real Postgres (role_capabilities JOIN
// user_process_areas, system_admin bypass check) — sqlmock cannot exercise
// this; every test here runs against the real testdb harness (ADR 0034).

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// markReviewedCtxWithIdentity builds the ctx TxRunner.Do's chokepoint
// auto-seed expects: tenant + actor identity, so seedTxIdentityFromContext
// seeds the metaldocs.tenant_id/actor_id GUCs authz.Require reads (F3.1/M3
// contract) — mirrors submitCtxWithIdentity (submit_service_reason_integration_test.go).
func markReviewedCtxWithIdentity(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, actorID)
	return ctx
}

// grantDocumentReview seeds a user_process_areas membership under role
// qms_admin (curated role_capabilities maps qms_admin -> document.review,
// db/reference-data/0001_product_reference_data.sql:187) so actorID holds
// document.review tenant-wide. document.review is ScopeTenant: the "tenant"
// sentinel short-circuits authz.Require's area predicate
// ($2 = 'tenant' OR upa.area_code = $2), so any governed area_code works.
func grantDocumentReview(t *testing.T, dbc *sql.DB, tenantID, actorID, areaCode string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', now() - interval '1 hour')`,
			actorID, tenantID, areaCode,
		)
		return err
	})
}

func newMarkReviewedTestService() *MarkReviewedService {
	return NewMarkReviewedService(NewSQLEmitter(), RealClock{})
}

// TestMarkReviewed_RequiresDocumentReviewCapability proves the tier-2
// authz.Require gate: the SAME request denied without the document.review
// grant, then succeeds once granted. This is the friendly tier-2 denial
// (authz.ErrCapDenied), never a DB tripwire P0001 — the tripwire arm itself
// is pinned separately (tests/integration/documents/tripwire_documents_test.go
// TestTripwire_DocumentsUpdate_DocumentReviewArm /
// TestTripwire_DocumentsUpdate_NoCapAssertedIsRejected).
func TestMarkReviewed_RequiresDocumentReviewCapability(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	owner := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithTaxonomy(tax), testdb.WithOwner(owner.ID))
	// Actor with NO role/grant at all (not system_admin, no user_process_areas row).
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID),
		testdb.WithControlledDoc(cd),
		testdb.WithOwner(owner.ID),
		testdb.WithStatus("published"),
		testdb.WithEffectiveFrom(effectiveFrom),
	)

	svc := newMarkReviewedTestService()
	runner := db.NewTxRunner(dbc)
	req := MarkReviewedRequest{
		TenantID:    tnt.ID,
		DocumentID:  doc.ID,
		ReviewDueAt: now.Add(365 * 24 * time.Hour),
		ReviewedBy:  actor.ID,
	}

	callCtx := markReviewedCtxWithIdentity(tnt.ID, actor.ID)

	// (a) withheld -> denied.
	_, err := svc.MarkReviewed(callCtx, runner, req)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("MarkReviewed without document.review grant = %v, want authz.ErrCapDenied", err)
	}

	// (b) granted -> succeeds.
	grantDocumentReview(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)

	res, err := svc.MarkReviewed(callCtx, runner, req)
	if err != nil {
		t.Fatalf("MarkReviewed with document.review grant = %v, want success", err)
	}
	if res.NewStatus != "published" {
		t.Fatalf("NewStatus = %q, want published", res.NewStatus)
	}
}

// TestMarkReviewed_SetsReviewDates is the happy-path proof: last_reviewed_at
// is set to ~now, review_due_at to the request value, effective_to when
// provided; status stays published; revision_version advances by 1.
func TestMarkReviewed_SetsReviewDates(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(actor.ID),
		testdb.WithStatus("published"),
		testdb.WithEffectiveFrom(effectiveFrom),
	)

	svc := newMarkReviewedTestService()
	runner := db.NewTxRunner(dbc)

	nextReviewDue := now.Add(365 * 24 * time.Hour)
	nextEffectiveTo := now.Add(730 * 24 * time.Hour)

	callCtx := markReviewedCtxWithIdentity(tnt.ID, actor.ID)
	res, err := svc.MarkReviewed(callCtx, runner, MarkReviewedRequest{
		TenantID:                tnt.ID,
		DocumentID:              doc.ID,
		ReviewDueAt:             nextReviewDue,
		EffectiveTo:             &nextEffectiveTo,
		ExpectedRevisionVersion: doc.RevisionVersion,
		ReviewedBy:              actor.ID,
	})
	if err != nil {
		t.Fatalf("MarkReviewed: unexpected error: %v", err)
	}
	wantRevVersion := doc.RevisionVersion + 1
	if res.RevisionVersion != wantRevVersion {
		t.Fatalf("RevisionVersion = %d, want %d", res.RevisionVersion, wantRevVersion)
	}
	if res.NewStatus != "published" {
		t.Fatalf("NewStatus = %q, want published (mark-reviewed is not a status transition)", res.NewStatus)
	}

	var gotStatus string
	var gotRevVersion int
	var gotLastReviewedAt, gotReviewDueAt, gotEffectiveTo sql.NullTime
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status, revision_version, last_reviewed_at, review_due_at, effective_to
		   FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		doc.ID, tnt.ID,
	).Scan(&gotStatus, &gotRevVersion, &gotLastReviewedAt, &gotReviewDueAt, &gotEffectiveTo); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotStatus != "published" {
		t.Fatalf("documents.status = %q, want published (no 11th status, no transition)", gotStatus)
	}
	if gotRevVersion != wantRevVersion {
		t.Fatalf("documents.revision_version = %d, want %d", gotRevVersion, wantRevVersion)
	}
	if !gotLastReviewedAt.Valid {
		t.Fatal("documents.last_reviewed_at not set")
	}
	if gotLastReviewedAt.Time.Before(now.Add(-time.Minute)) || gotLastReviewedAt.Time.After(time.Now().UTC().Add(time.Minute)) {
		t.Fatalf("documents.last_reviewed_at = %v, want ~now", gotLastReviewedAt.Time)
	}
	if !gotReviewDueAt.Valid || !gotReviewDueAt.Time.Equal(nextReviewDue) {
		t.Fatalf("documents.review_due_at = %v, want %v", gotReviewDueAt.Time, nextReviewDue)
	}
	if !gotEffectiveTo.Valid || !gotEffectiveTo.Time.Equal(nextEffectiveTo) {
		t.Fatalf("documents.effective_to = %v, want %v", gotEffectiveTo.Time, nextEffectiveTo)
	}
}

// TestMarkReviewed_OCCConflict proves a stale ExpectedRevisionVersion (a
// stale If-Match at the HTTP layer) yields 0 rows affected ->
// ErrMarkReviewedStaleRevision, with no write to the row.
func TestMarkReviewed_OCCConflict(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	now := time.Now().UTC().Truncate(time.Second)
	effectiveFrom := now.Add(-30 * 24 * time.Hour)
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(actor.ID),
		testdb.WithStatus("published"),
		testdb.WithEffectiveFrom(effectiveFrom),
	)

	svc := newMarkReviewedTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := markReviewedCtxWithIdentity(tnt.ID, actor.ID)

	staleVersion := doc.RevisionVersion + 999
	_, err := svc.MarkReviewed(callCtx, runner, MarkReviewedRequest{
		TenantID:                tnt.ID,
		DocumentID:              doc.ID,
		ReviewDueAt:             now.Add(365 * 24 * time.Hour),
		ExpectedRevisionVersion: staleVersion,
		ReviewedBy:              actor.ID,
	})
	if !errors.Is(err, ErrMarkReviewedStaleRevision) {
		t.Fatalf("MarkReviewed with stale revision = %v, want ErrMarkReviewedStaleRevision", err)
	}

	var gotRevVersion int
	var gotLastReviewedAt sql.NullTime
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT revision_version, last_reviewed_at FROM public.documents WHERE id = $1::uuid`,
		doc.ID,
	).Scan(&gotRevVersion, &gotLastReviewedAt); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotRevVersion != doc.RevisionVersion {
		t.Fatalf("documents.revision_version = %d after OCC conflict, want unchanged %d", gotRevVersion, doc.RevisionVersion)
	}
	if gotLastReviewedAt.Valid {
		t.Fatal("documents.last_reviewed_at was written despite the OCC conflict — no write must occur")
	}
}

// TestMarkReviewed_RejectsNonPublished proves a non-published document (draft)
// is rejected with the friendly ErrDocumentNotPublished precondition, not a
// silent no-op mistaken for success and not routed through
// docsdomain.CanTransitionDocumentStatus (mark-reviewed is not a transition).
func TestMarkReviewed_RejectsNonPublished(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(actor.ID),
		testdb.WithStatus("draft"),
	)

	svc := newMarkReviewedTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := markReviewedCtxWithIdentity(tnt.ID, actor.ID)

	_, err := svc.MarkReviewed(callCtx, runner, MarkReviewedRequest{
		TenantID:    tnt.ID,
		DocumentID:  doc.ID,
		ReviewDueAt: time.Now().UTC().Add(365 * 24 * time.Hour),
		ReviewedBy:  actor.ID,
	})
	if !errors.Is(err, ErrDocumentNotPublished) {
		t.Fatalf("MarkReviewed on draft document = %v, want ErrDocumentNotPublished", err)
	}
}

// TestMarkReviewed_TenantIsolation proves a cross-tenant document id is
// invisible (RLS FORCE policy on public.documents) -> ErrDocumentNotFound, no
// cross-tenant write.
func TestMarkReviewed_TenantIsolation(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, dbc)
	tntB := testdb.NewTenant(t, dbc)
	ownerB := testdb.NewUser(t, dbc, testdb.WithTenant(tntB.ID))
	actorA := testdb.NewUser(t, dbc, testdb.WithTenant(tntA.ID), testdb.WithRole("system_admin"))

	now := time.Now().UTC().Truncate(time.Second)
	docB := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tntB.ID),
		testdb.WithOwner(ownerB.ID),
		testdb.WithStatus("published"),
		testdb.WithEffectiveFrom(now.Add(-30*24*time.Hour)),
	)

	svc := newMarkReviewedTestService()
	runner := db.NewTxRunner(dbc)
	// Identity is tenant A, but the target document belongs to tenant B.
	callCtx := markReviewedCtxWithIdentity(tntA.ID, actorA.ID)

	_, err := svc.MarkReviewed(callCtx, runner, MarkReviewedRequest{
		TenantID:    tntA.ID,
		DocumentID:  docB.ID,
		ReviewDueAt: now.Add(365 * 24 * time.Hour),
		ReviewedBy:  actorA.ID,
	})
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("MarkReviewed on cross-tenant document id = %v, want ErrDocumentNotFound", err)
	}

	var gotStatus string
	var gotLastReviewedAt sql.NullTime
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status, last_reviewed_at FROM public.documents WHERE id = $1::uuid`,
		docB.ID,
	).Scan(&gotStatus, &gotLastReviewedAt); err != nil {
		t.Fatalf("read back tenant-B document row: %v", err)
	}
	if gotLastReviewedAt.Valid {
		t.Fatal("tenant-B document was written under tenant-A identity — cross-tenant write leaked")
	}
}
