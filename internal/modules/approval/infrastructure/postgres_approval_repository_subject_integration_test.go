//go:build integration
// +build integration

package infrastructure

// P3.S2a (M3 kernel extraction, ADR 0082) — real-DB proof that the two
// document-only shortcuts in the repository are closed:
//
//  1. InsertInstance now HARD-REQUIRES a valid domain.Subject (no silent
//     document fallback for a zero-value Subject).
//  2. Route/Instance read-hydration derives Subject from the REAL
//     subject_kind/subject_key columns, not from the legacy
//     profile_code/document_id columns — proven by round-tripping a
//     template-subject row whose subject_key differs from the legacy
//     document_id/profile_code value.
//
// Before the P3.S2a fix, TestLoadInstance_TemplateSubject_RoundTrips_RealDB
// and TestLoadRoute_TemplateSubject_RoundTrips_RealDB were RED: the read path
// derived Subject as {document, <document_id>} / {document, <profile_code>}
// regardless of what was actually persisted in subject_kind/subject_key.

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// insertInstanceTx begins a fresh tx, seeds tenant/actor identity plus the
// document.submit capability the approval_instances INSERT tripwire (migration
// 0259) requires, and returns the tx for the caller to use with
// repo.InsertInstance and then Commit/Rollback.
func insertInstanceTx(t *testing.T, db *sql.DB, tenantID, actorID string) *sql.Tx {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	// Assert BOTH submit caps: the approval_instances INSERT tripwire is now
	// subject-discriminated (ADR 0083, migration 0299) — a document row requires
	// document.submit, a template row requires template.submit. This shared
	// helper seeds instances of either subject kind, so it asserts both; the
	// trigger's per-row CASE picks the one it needs. (Asserting an unused cap is
	// harmless — the trigger checks membership, not exclusivity.)
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`,
		`[{"cap":"document.submit"},{"cap":"template.submit"}]`,
	); err != nil {
		t.Fatalf("assert submit caps: %v", err)
	}
	return tx
}

// TestLoadInstance_TemplateSubject_RoundTrips_RealDB proves (b): every
// read-hydration path (LoadInstance, LoadInstancesByIDs) surfaces the real
// persisted Subject (subject_kind='template', subject_key=<template key>)
// rather than deriving {document, document_id} from the legacy column.
//
// The fixture is inserted via raw SQL with a NULL document_id — the legal
// post-migration-0297 shape for a template row (0297's projection CHECK now
// forbids a template row from carrying a document_id). It is NOT inserted
// through repo.InsertInstance because that method still writes inst.DocumentID
// verbatim and cannot yet emit a NULL document_id; the repository's
// template-write cutover is P3.S2b-1. This test's mandate is the READ side,
// which is fully exercised here against a real template row.
func TestLoadInstance_TemplateSubject_RoundTrips_RealDB(t *testing.T) {
	// P3.S2b-1: repo READ path is now NULL-document_id tolerant (LEFT JOIN +
	// nullable document_id/revision scans in LoadInstance/LoadInstancesByIDs).

	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, dbc)
	doc := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cd), testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(doc.TenantID))

	templateKey := "tmpl-" + uuid.NewString()
	if templateKey == doc.ID {
		t.Fatal("templateKey collided with document.ID — test setup broken")
	}
	instID := uuid.NewString()

	// Raw insert of a legal template row (NULL document_id). insertInstanceTx
	// seeds tenant/actor identity + the document.submit cap the approval_instances
	// INSERT tripwire requires.
	tx := insertInstanceTx(t, dbc, doc.TenantID, doc.Owner)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_instances
		   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
		    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		    subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, NULL, $3::uuid, 1, 'in_progress', $4, now(), repeat('a', 64), $5, 'template', $6)`,
		instID, doc.TenantID, route.ID, doc.Owner, "idem-"+uuid.NewString(), templateKey,
	); err != nil {
		tx.Rollback()
		t.Fatalf("raw insert template instance: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Sanity: the persisted row is a NULL-document_id template subject.
	var documentID sql.NullString
	var subjectKind, subjectKey string
	if err := dbc.QueryRowContext(ctx,
		`SELECT document_id, subject_kind, subject_key FROM public.approval_instances WHERE id = $1::uuid`,
		instID,
	).Scan(&documentID, &subjectKind, &subjectKey); err != nil {
		t.Fatalf("query persisted subject: %v", err)
	}
	if documentID.Valid {
		t.Fatalf("document_id = %q, want NULL for a template row", documentID.String)
	}
	if subjectKind != "template" || subjectKey != templateKey {
		t.Fatalf("persisted subject = (%q,%q), want (template,%q)", subjectKind, subjectKey, templateKey)
	}

	readTx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin read tx: %v", err)
	}
	defer readTx.Rollback()

	// LoadInstance must hydrate Subject from subject_kind/subject_key, not
	// derive {document, document_id} from the legacy column.
	got, err := repo.LoadInstance(ctx, readTx, doc.TenantID, instID)
	if err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if got.Subject.Kind != domain.SubjectKindTemplate || got.Subject.Key != templateKey {
		t.Fatalf("LoadInstance Subject = %+v, want {template %s} — hydration is still deriving from document_id", got.Subject, templateKey)
	}

	// LoadInstancesByIDs (batch path) must agree.
	batch, err := repo.LoadInstancesByIDs(ctx, readTx, doc.TenantID, []string{instID})
	if err != nil {
		t.Fatalf("LoadInstancesByIDs: %v", err)
	}
	if len(batch) != 1 {
		t.Fatalf("LoadInstancesByIDs len = %d, want 1", len(batch))
	}
	if batch[0].Subject.Kind != domain.SubjectKindTemplate || batch[0].Subject.Key != templateKey {
		t.Fatalf("LoadInstancesByIDs Subject = %+v, want {template %s} — hydration is still deriving from document_id", batch[0].Subject, templateKey)
	}
}

// TestInsertInstance_ZeroSubject_ReturnsError_RealDB proves (a): InsertInstance
// no longer silently defaults a zero-value Subject to {document, DocumentID}.
// A caller that forgot to set Subject must get a hard error, not a mis-tagged
// row.
func TestInsertInstance_ZeroSubject_ReturnsError_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, dbc)
	doc := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cd), testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(doc.TenantID))

	inst := domain.Instance{
		ID:         uuid.NewString(),
		TenantID:   doc.TenantID,
		DocumentID: doc.ID,
		// Subject intentionally left zero-value (Kind == "").
		RouteID:              route.ID,
		RouteVersionSnapshot: 1,
		Status:               domain.InstanceInProgress,
		SubmittedBy:          doc.Owner,
		SubmittedAt:          time.Now(),
		ContentHashAtSubmit:  strings.Repeat("a", 64),
		IdempotencyKey:       "idem-" + uuid.NewString(),
	}

	tx := insertInstanceTx(t, dbc, doc.TenantID, doc.Owner)
	defer tx.Rollback()

	err := repo.InsertInstance(ctx, tx, inst)
	if err == nil {
		t.Fatal("InsertInstance with zero-value Subject: want error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidSubjectKind) {
		t.Fatalf("InsertInstance err = %v, want wrapping domain.ErrInvalidSubjectKind", err)
	}

	// No row should have been written.
	var count int
	if err := dbc.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.approval_instances WHERE id = $1::uuid`,
		inst.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("row count for rejected insert = %d, want 0", count)
	}
}

// TestLoadRoute_TemplateSubject_RoundTrips_RealDB proves (b) for routes:
// LoadRoute hydrates Subject from the real subject_kind/subject_key columns,
// not by deriving the kind from the legacy profile_code column. Under ADR 0086
// a template ROUTE is profile-keyed (profile_code = subject_key, enforced by
// approval_routes_template_subject_key_check, migration 0315), so the
// discriminator that must survive the round trip is subject_kind: a document
// route and a template route on the SAME profile are distinguishable only by it.
func TestLoadRoute_TemplateSubject_RoundTrips_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	route := testdb.NewApprovalRoute(t, dbc,
		testdb.WithSubjectKind(string(domain.SubjectKindTemplate)))

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadRoute(ctx, tx, route.TenantID, route.ID)
	if err != nil {
		t.Fatalf("LoadRoute: %v", err)
	}
	if got.Subject.Kind != domain.SubjectKindTemplate || got.Subject.Key != route.ProfileCode {
		t.Fatalf("LoadRoute Subject = %+v, want {template %s} — hydration is still deriving the kind from profile_code", got.Subject, route.ProfileCode)
	}
}

// TestLoadActiveRouteIDBySubject_TemplateSubject_RealDB proves the
// subject-generic route-selection method (P3.S2b-2, ADR 0082) selects a
// template-subject route by (subject_kind, subject_key) — something the legacy
// profile_code-only LoadActiveRouteIDByProfile can never do, because under ADR
// 0086 both kinds share the profile key and only subject_kind separates them —
// while the document path stays selectable both via the generic method AND via
// the now-delegating LoadActiveRouteIDByProfile (delegation-equivalence axis).
//
// The two routes are deliberately seeded on the SAME profile: that is the case
// the rebuilt approval_routes_active_profile_uq index must permit, and the case
// where a kind-blind selector would return the wrong route.
func TestLoadActiveRouteIDBySubject_TemplateSubject_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	baseRepo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	// LoadActiveRouteIDByProfile / LoadActiveRouteIDBySubject are declared on
	// application.SubmitDefaultsResolver, not on the broader ApprovalRepository
	// interface NewPostgresApprovalRepository returns (same asserted-interface
	// pattern as application.newSubmitServiceWithResolver). Assert against a
	// minimal local interface to reach both methods on the concrete repo.
	repo, ok := interface{}(baseRepo).(interface {
		LoadActiveRouteIDByProfile(ctx context.Context, tx db.Tx, tenantID, profileCode string) (string, error)
		LoadActiveRouteIDBySubject(ctx context.Context, tx db.Tx, tenantID, subjectKind, subjectKey string) (string, error)
	})
	if !ok {
		t.Fatalf("postgres approval repository does not satisfy the route-selection interface; production wiring would silently break")
	}

	// Document-subject route via the ordinary factory (active by default).
	docRoute := testdb.NewApprovalRoute(t, dbc)

	// Template-subject route on the SAME profile as the document route.
	tmplRoute := testdb.NewApprovalRoute(t, dbc,
		testdb.WithTenant(docRoute.TenantID),
		testdb.WithProfile(docRoute.ProfileCode),
		testdb.WithSubjectKind(string(domain.SubjectKindTemplate)))
	if tmplRoute.ID == docRoute.ID {
		t.Fatal("factory returned the same route for both kinds — test setup broken")
	}
	templateKey := tmplRoute.ProfileCode

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Template subject: only reachable via the new generic method.
	gotTmplID, err := repo.LoadActiveRouteIDBySubject(ctx, tx, docRoute.TenantID, string(domain.SubjectKindTemplate), templateKey)
	if err != nil {
		t.Fatalf("LoadActiveRouteIDBySubject(template): %v", err)
	}
	if gotTmplID != tmplRoute.ID {
		t.Fatalf("LoadActiveRouteIDBySubject(template) = %q, want %q", gotTmplID, tmplRoute.ID)
	}

	// Document subject: both the generic method and the delegating legacy
	// method must find the same route id.
	gotDocViaSubject, err := repo.LoadActiveRouteIDBySubject(ctx, tx, docRoute.TenantID, string(domain.SubjectKindDocument), docRoute.ProfileCode)
	if err != nil {
		t.Fatalf("LoadActiveRouteIDBySubject(document): %v", err)
	}
	gotDocViaProfile, err := repo.LoadActiveRouteIDByProfile(ctx, tx, docRoute.TenantID, docRoute.ProfileCode)
	if err != nil {
		t.Fatalf("LoadActiveRouteIDByProfile: %v", err)
	}
	if gotDocViaSubject != docRoute.ID || gotDocViaProfile != docRoute.ID || gotDocViaSubject != gotDocViaProfile {
		t.Fatalf("document route ids diverge: viaSubject=%q viaProfile=%q want=%q", gotDocViaSubject, gotDocViaProfile, docRoute.ID)
	}
}

// TestLoadInstance_DocumentSubject_StaysByteEqual_RealDB is the byte-equal
// control for the document path: an instance seeded through the ordinary
// testdb factory (which omits subject_kind/subject_key and relies on the
// 0296 compat trigger to backfill {document, document_id}) must still
// hydrate to exactly that Subject after the P3.S2a read-hydration change —
// proving the real-column read is equivalent to the old derived-from-
// document_id read for every existing document row.
func TestLoadInstance_DocumentSubject_StaysByteEqual_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, dbc)
	doc := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cd), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, dbc, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadInstance(ctx, tx, doc.TenantID, inst.ID)
	if err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if got.Subject.Kind != domain.SubjectKindDocument || got.Subject.Key != doc.ID {
		t.Fatalf("Subject = %+v, want {document %s} (byte-equal to the legacy derived value)", got.Subject, doc.ID)
	}
}
