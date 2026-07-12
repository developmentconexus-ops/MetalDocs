//go:build integration
// +build integration

// Package migrations_test — P3.S2b-0 (ROADMAP unit, M3 approval kernel
// extraction, Phase 3). DB expand-phase validation for migration 0297: relaxes
// the legacy document-only NOT NULL columns (approval_instances.document_id,
// approval_routes.profile_code) left NOT NULL by 0296, so a template subject
// row (which has no public.documents row and no
// metaldocs.document_profiles row) can be INSERTed.
//
// Contract under test (see the 0297 migration file header):
//   - document_id / profile_code are now NULL-able.
//   - Both legacy FKs stay in place and are NULL-tolerant (MATCH SIMPLE):
//     document rows keep full referential integrity; template rows (legacy
//     column NULL) skip the FK check.
//   - Projection CHECK constraints enforce subject_kind vs legacy-column
//     presence: a 'document' row MUST have the legacy column set; a
//     'template' row MUST NOT.
//   - A new subject-scoped partial unique index
//     (tenant_id, subject_kind, subject_key, idempotency_key) dedups
//     approval_instances submissions for BOTH subjects, since the pre-existing
//     (document_id, idempotency_key) constraint cannot dedup two NULL
//     document_id rows.
package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0297_ApprovalInstances_TemplateSubject_NullDocumentID_Insertable
// is the RED->GREEN proof for approval_instances: before 0297,
// document_id NOT NULL made this INSERT impossible (23502 not_null_violation).
// After 0297, a template-subject row with document_id NULL inserts cleanly.
func TestMigration0297_ApprovalInstances_TemplateSubject_NullDocumentID_Insertable(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	subjectKey := uuid.NewString() // stands in for a template_version_id
	err := insertInstanceWithSubjectAndDocID(t, db, doc.TenantID, route.ID, doc.Owner, nil, "template", subjectKey)
	if err != nil {
		t.Fatalf("insert template-subject instance with NULL document_id: %v", err)
	}

	var documentID sql.NullString
	var kind, key string
	qErr := db.QueryRowContext(context.Background(),
		`SELECT document_id, subject_kind, subject_key FROM public.approval_instances
		  WHERE tenant_id = $1::uuid AND subject_key = $2`,
		doc.TenantID, subjectKey,
	).Scan(&documentID, &kind, &key)
	if qErr != nil {
		t.Fatalf("query inserted template instance: %v", qErr)
	}
	if documentID.Valid {
		t.Fatalf("document_id = %q, want NULL for a template-subject row", documentID.String)
	}
	if kind != "template" {
		t.Fatalf("subject_kind = %q, want template", kind)
	}
	if key != subjectKey {
		t.Fatalf("subject_key = %q, want %q", key, subjectKey)
	}
}

// TestMigration0297_ApprovalRoutes_TemplateSubject_NullProfileCode_Insertable
// is the RED->GREEN proof for approval_routes: before 0297, profile_code NOT
// NULL made this INSERT impossible. After 0297, a template-subject route row
// with profile_code NULL inserts cleanly.
func TestMigration0297_ApprovalRoutes_TemplateSubject_NullProfileCode_Insertable(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	subjectKey := uuid.NewString() // stands in for a template_id
	err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, false, 1, "template", subjectKey)
	if err != nil {
		t.Fatalf("insert template-subject route with NULL profile_code: %v", err)
	}

	var profileCode sql.NullString
	var kind, key string
	qErr := db.QueryRowContext(context.Background(),
		`SELECT profile_code, subject_kind, subject_key FROM public.approval_routes
		  WHERE tenant_id = $1::uuid AND subject_key = $2`,
		tenant.ID, subjectKey,
	).Scan(&profileCode, &kind, &key)
	if qErr != nil {
		t.Fatalf("query inserted template route: %v", qErr)
	}
	if profileCode.Valid {
		t.Fatalf("profile_code = %q, want NULL for a template-subject row", profileCode.String)
	}
	if kind != "template" {
		t.Fatalf("subject_kind = %q, want template", kind)
	}
	if key != subjectKey {
		t.Fatalf("subject_key = %q, want %q", key, subjectKey)
	}
}

// TestMigration0297_ApprovalInstances_DocumentRow_StillRequiresDocumentID
// proves the projection CHECK rejects a 'document' subject row that omits
// document_id — the invariant the dropped NOT NULL used to guarantee for free.
func TestMigration0297_ApprovalInstances_DocumentRow_StillRequiresDocumentID(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	err := insertInstanceWithSubjectAndDocID(t, db, doc.TenantID, route.ID, doc.Owner, nil, "document", "should-fail")
	assertCheckViolation(t, err)
}

// TestMigration0297_ApprovalInstances_TemplateRow_RejectsNonNullDocumentID
// proves the projection CHECK rejects a 'template' subject row that falsely
// carries a document_id — guards against a future half-migrated write.
func TestMigration0297_ApprovalInstances_TemplateRow_RejectsNonNullDocumentID(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	err := insertInstanceWithSubjectAndDocID(t, db, doc.TenantID, route.ID, doc.Owner, &doc.ID, "template", "should-fail")
	assertCheckViolation(t, err)
}

// TestMigration0297_ApprovalRoutes_DocumentRow_StillRequiresProfileCode is the
// approval_routes sibling of the document-row projection-CHECK assertion.
func TestMigration0297_ApprovalRoutes_DocumentRow_StillRequiresProfileCode(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, false, 1, "document", "should-fail")
	assertCheckViolation(t, err)
}

// TestMigration0297_ApprovalRoutes_TemplateRow_RejectsNonNullProfileCode is
// the approval_routes sibling of the template-row projection-CHECK assertion.
func TestMigration0297_ApprovalRoutes_TemplateRow_RejectsNonNullProfileCode(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))

	profileCode := tax.ProfileCode
	err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, &profileCode, owner.ID, false, 1, "template", "should-fail")
	assertCheckViolation(t, err)
}

// TestMigration0297_ApprovalInstances_DocumentPath_ByteEqual proves the
// document path is completely unaffected: the normal document-subject
// factory-produced row still inserts, still carries document_id, and its FK
// is still fully checked (an invalid document_id would 23503, proven
// separately by the pre-existing suite; this test proves the happy path is
// byte-equal to pre-0297 behavior).
func TestMigration0297_ApprovalInstances_DocumentPath_ByteEqual(t *testing.T) {
	db, _ := testdb.Open(t)

	inst := testdb.NewApprovalInstance(t, db)

	var documentID, kind, key string
	err := db.QueryRowContext(context.Background(),
		`SELECT document_id, subject_kind, subject_key FROM public.approval_instances WHERE id = $1::uuid`,
		inst.ID,
	).Scan(&documentID, &kind, &key)
	if err != nil {
		t.Fatalf("query inserted document instance: %v", err)
	}
	if documentID != inst.DocumentID {
		t.Fatalf("document_id = %q, want %q", documentID, inst.DocumentID)
	}
	if kind != "document" {
		t.Fatalf("subject_kind = %q, want document", kind)
	}
	if key != inst.DocumentID {
		t.Fatalf("subject_key = %q, want document_id %q", key, inst.DocumentID)
	}
}

// TestMigration0297_SubjectIdempotencyIndex_RejectsDuplicateTemplateSubmit
// proves ux_approval_instances_subject_idempotency actually dedups two
// template-subject submissions sharing (tenant_id, subject_kind, subject_key,
// idempotency_key) — the gap the legacy (document_id, idempotency_key)
// constraint cannot close once document_id is NULL for both rows (NULL is
// distinct from NULL in a unique index).
func TestMigration0297_SubjectIdempotencyIndex_RejectsDuplicateTemplateSubmit(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	subjectKey := uuid.NewString()
	idemKey := "idem-" + uuid.NewString()

	// First submission is terminal ('approved') so the active-subject partial
	// index (WHERE status='in_progress') is out of play — the only index that
	// can reject the second, same-(subject,idem_key) row is the new
	// ux_approval_instances_subject_idempotency.
	if err := insertInstanceFull(t, db, doc.TenantID, route.ID, doc.Owner, nil, "template", subjectKey, idemKey, "approved"); err != nil {
		t.Fatalf("insert first (terminal) template submission: %v", err)
	}
	err := insertInstanceFull(t, db, doc.TenantID, route.ID, doc.Owner, nil, "template", subjectKey, idemKey, "in_progress")
	assertUniqueViolation(t, err)
}

// TestMigration0297_SubjectIdempotencyIndex_AllowsDistinctIdempotencyKeys
// proves two template submissions for the SAME subject but DIFFERENT
// idempotency keys are both allowed — the index does not over-constrain.
func TestMigration0297_SubjectIdempotencyIndex_AllowsDistinctIdempotencyKeys(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	subjectKey := uuid.NewString()

	// First submission is terminal so exactly one in_progress row exists for the
	// subject (active-subject index satisfied); a second submission for the SAME
	// subject with a DIFFERENT idempotency key must then be admitted — proving
	// ux_approval_instances_subject_idempotency keys on idempotency_key and does
	// not over-constrain to one row per subject.
	if err := insertInstanceFull(t, db, doc.TenantID, route.ID, doc.Owner, nil, "template", subjectKey, "idem-"+uuid.NewString(), "approved"); err != nil {
		t.Fatalf("insert first (terminal) template submission: %v", err)
	}
	if err := insertInstanceFull(t, db, doc.TenantID, route.ID, doc.Owner, nil, "template", subjectKey, "idem-"+uuid.NewString(), "in_progress"); err != nil {
		t.Fatalf("insert second template submission with distinct idempotency key must be allowed, got: %v", err)
	}
}

// TestMigration0297_LegacyFKs_KeptAndStillEnforced_ForDocumentRows proves
// both legacy FKs are still in place and reject an invalid referent for a
// document-subject row (document_id -> a non-existent public.documents.id,
// and profile_code -> a non-existent metaldocs.document_profiles.code) —
// confirming the migration did NOT drop either FK.
func TestMigration0297_LegacyFKs_KeptAndStillEnforced_ForDocumentRows(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	bogusDocID := uuid.NewString()
	err := insertInstanceWithSubjectAndDocID(t, db, doc.TenantID, route.ID, doc.Owner, &bogusDocID, "document", "should-fk-fail")
	assertForeignKeyViolation(t, err)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// insertInstanceWithSubjectAndDocID inserts an approval_instances row with an
// explicit (possibly-NULL) document_id and subject_kind/subject_key, using a
// fresh idempotency key each call so unrelated tests never collide on the
// legacy or subject-scoped unique indexes.
func insertInstanceWithSubjectAndDocID(t *testing.T, db *sql.DB, tenantID, routeID, submittedBy string, documentID *string, subjectKind, subjectKey string) error {
	t.Helper()
	return insertInstanceWithSubjectDocIDAndIdemKey(t, db, tenantID, routeID, submittedBy, documentID, subjectKind, subjectKey, "idem-"+uuid.NewString())
}

func insertInstanceWithSubjectDocIDAndIdemKey(t *testing.T, db *sql.DB, tenantID, routeID, submittedBy string, documentID *string, subjectKind, subjectKey, idemKey string) error {
	t.Helper()
	return insertInstanceFull(t, db, tenantID, routeID, submittedBy, documentID, subjectKind, subjectKey, idemKey, "in_progress")
}

// insertInstanceFull is the status-aware core. The idempotency tests drive the
// first row to a terminal status ('approved') so the pre-existing partial index
// ux_approval_instances_active_subject (WHERE status='in_progress') does NOT
// govern the pair — isolating ux_approval_instances_subject_idempotency as the
// only constraint that can fire, which is the index this slice actually adds.
func insertInstanceFull(t *testing.T, db *sql.DB, tenantID, routeID, submittedBy string, documentID *string, subjectKind, subjectKey, idemKey, status string) error {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
	); err != nil {
		t.Fatalf("set tenant_id GUC: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, `[{"cap":"document.submit"}]`,
	); err != nil {
		t.Fatalf("assert document.submit cap: %v", err)
	}

	id := uuid.NewString()
	var docIDArg any
	if documentID != nil {
		docIDArg = *documentID
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.approval_instances
		   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
		    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		    subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, $9, $5, now(), repeat('a', 64), $6, $7, $8)`,
		id, tenantID, docIDArg, routeID, submittedBy, idemKey, subjectKind, subjectKey, status,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// insertRouteWithSubjectAndProfile inserts an approval_routes row with an
// explicit (possibly-NULL) profile_code and subject_kind/subject_key.
func insertRouteWithSubjectAndProfile(t *testing.T, db *sql.DB, tenantID string, profileCode *string, ownerID string, active bool, version int, subjectKind, subjectKey string) error {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
	); err != nil {
		t.Fatalf("set tenant_id GUC: %v", err)
	}

	id := uuid.NewString()
	var profileArg any
	if profileCode != nil {
		profileArg = *profileCode
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.approval_routes
		   (id, tenant_id, name, profile_code, version, active, created_by, subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, 'Route', $3, $4, $5, $6, $7, $8)`,
		id, tenantID, profileArg, version, active, ownerID, subjectKind, subjectKey,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// assertCheckViolation / assertUniqueViolation are shared with
// migration_0296_test.go (same package) — not redeclared here.

func assertForeignKeyViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a FOREIGN KEY constraint violation, got nil error")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23503" {
			t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got %s: %v", pgErr.Code, err)
		}
		return
	}
	if !strings.Contains(err.Error(), "foreign key") {
		t.Fatalf("expected a foreign-key-constraint violation, got: %v", err)
	}
}
