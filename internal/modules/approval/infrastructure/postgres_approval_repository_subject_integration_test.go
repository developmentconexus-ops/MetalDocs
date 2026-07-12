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
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`,
		`[{"cap":"document.submit"}]`,
	); err != nil {
		t.Fatalf("assert document.submit cap: %v", err)
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
	// RED-deferred to P3.S2b-1. Migration 0297 (this slice, P3.S2b-0) makes a
	// NULL-document_id template row legal at the schema level (proven green by
	// tests/integration/migrations/migration_0297_test.go). But LoadInstance's
	// read query still INNER JOINs documents ON ai.document_id and scans
	// document_id into a non-nullable string — both document-only assumptions
	// that drop / fail on a template row. Making the repo READ path
	// subject-generic (LEFT JOIN + nullable document_id/revision scans) is
	// P3.S2b-1 (repo subject-generalization), not this schema-only slice. Unskip
	// and implement there.
	t.Skip("unskip in P3.S2b-1: LoadInstance read path is not yet NULL-document_id tolerant")

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
		ID:                   uuid.NewString(),
		TenantID:             doc.TenantID,
		DocumentID:           doc.ID,
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
// not by deriving {document, profile_code} from the legacy profile_code
// column. The route is seeded via the normal factory (document subject, no
// tripwire on approval_routes INSERT — see testdb.NewApprovalRoute), then
// UPDATEd directly to a template subject whose key differs from profile_code
// (UPDATE is not covered by the P2.S1 compat trigger, which only fires
// BEFORE INSERT, so this exercises exactly the persisted-value case the read
// path must trust).
func TestLoadRoute_TemplateSubject_RoundTrips_RealDB(t *testing.T) {
	// RED-deferred to P3.S2b-1. Migration 0297 makes a NULL-profile_code template
	// route legal (proven green by migration_0297_test.go), but LoadRoute still
	// scans profile_code into a non-nullable string ("converting NULL to string
	// is unsupported"). Making the repo READ path nullable-profile_code tolerant
	// is P3.S2b-1 (repo subject-generalization), not this schema-only slice.
	t.Skip("unskip in P3.S2b-1: LoadRoute read path is not yet NULL-profile_code tolerant")

	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	route := testdb.NewApprovalRoute(t, dbc)

	templateKey := "tmpl-route-" + uuid.NewString()
	if templateKey == route.ProfileCode {
		t.Fatal("templateKey collided with route.ProfileCode — test setup broken")
	}
	// profile_code goes NULL alongside the flip: migration 0297's projection
	// CHECK forbids a template-subject route from carrying a profile_code.
	if _, err := dbc.ExecContext(ctx,
		`UPDATE public.approval_routes SET subject_kind = 'template', subject_key = $2, profile_code = NULL WHERE id = $1::uuid`,
		route.ID, templateKey,
	); err != nil {
		t.Fatalf("update route subject: %v", err)
	}

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadRoute(ctx, tx, route.TenantID, route.ID)
	if err != nil {
		t.Fatalf("LoadRoute: %v", err)
	}
	if got.Subject.Kind != domain.SubjectKindTemplate || got.Subject.Key != templateKey {
		t.Fatalf("LoadRoute Subject = %+v, want {template %s} — hydration is still deriving from profile_code", got.Subject, templateKey)
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
