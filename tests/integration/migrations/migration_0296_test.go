//go:build integration
// +build integration

// Package migrations_test — P2.S1 (ROADMAP unit 3.1, M3 approval kernel
// extraction). DB expand-phase validation for migration 0296: adds
// subject_kind/subject_key to public.approval_routes and
// public.approval_instances, generalizing the approval kernel's keying from
// document-specific (profile_code / document_id) to a subject abstraction
// that will also cover templates.
//
// Contract under test (see docs/superpowers ROADMAP unit 3.1 / ADR 0082):
//   - subject_kind is an enum ('document' | 'template'), NOT NULL, CHECK-enforced.
//   - subject_key is NOT NULL text.
//   - approval_routes.profile_code stays NOT NULL as a backward-compat alias
//     (document, profile_code); legacy INSERTs that omit subject_kind/subject_key
//     get them auto-populated ('document', profile_code) by a BEFORE INSERT
//     compatibility trigger, so P2.S2's Go-code cutover is not required for this
//     column to go NOT NULL.
//   - approval_instances.document_id stays NOT NULL this phase; legacy INSERTs
//     get subject_kind/subject_key auto-populated ('document', document_id::text).
//   - New tenant-scoped subject-keyed unique/partial-unique/lookup indexes are
//     added ALONGSIDE the existing document-keyed ones (expand-only; no drops).
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

// TestMigration0296_ApprovalRoutes_LegacyInsertBackfillsDocumentSubject proves
// an INSERT that never mentions subject_kind/subject_key (the shape every
// existing Go call site and test fixture uses, unmodified in this slice)
// still succeeds and lands 'document' + profile_code — the compatibility
// trigger, not app code, does the backfill.
func TestMigration0296_ApprovalRoutes_LegacyInsertBackfillsDocumentSubject(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db)

	var kind, key string
	err := db.QueryRowContext(context.Background(),
		`SELECT subject_kind, subject_key FROM public.approval_routes WHERE id = $1::uuid`,
		route.ID,
	).Scan(&kind, &key)
	if err != nil {
		t.Fatalf("query subject_kind/subject_key: %v", err)
	}
	if kind != "document" {
		t.Fatalf("subject_kind = %q, want document", kind)
	}
	if key != route.ProfileCode {
		t.Fatalf("subject_key = %q, want profile_code %q", key, route.ProfileCode)
	}
}

// TestMigration0296_ApprovalInstances_LegacyInsertBackfillsDocumentSubject is
// the approval_instances sibling: NewApprovalInstance uses the same
// unmodified INSERT shape as internal/modules/approval/infrastructure's
// production repository.
func TestMigration0296_ApprovalInstances_LegacyInsertBackfillsDocumentSubject(t *testing.T) {
	db, _ := testdb.Open(t)

	inst := testdb.NewApprovalInstance(t, db)

	var kind, key string
	err := db.QueryRowContext(context.Background(),
		`SELECT subject_kind, subject_key FROM public.approval_instances WHERE id = $1::uuid`,
		inst.ID,
	).Scan(&kind, &key)
	if err != nil {
		t.Fatalf("query subject_kind/subject_key: %v", err)
	}
	if kind != "document" {
		t.Fatalf("subject_kind = %q, want document", kind)
	}
	if key != inst.DocumentID {
		t.Fatalf("subject_key = %q, want document_id %q", key, inst.DocumentID)
	}
}

// TestMigration0296_ApprovalRoutes_CheckRejectsInvalidSubjectKind asserts the
// CHECK constraint is a real last line: an explicit out-of-enum subject_kind
// is rejected even though profile_code/subject_key are otherwise valid.
func TestMigration0296_ApprovalRoutes_CheckRejectsInvalidSubjectKind(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))

	err := insertRouteWithSubject(t, db, tenant.ID, tax.ProfileCode, owner.ID, false, 1, "invalid", "whatever")
	assertCheckViolation(t, err)
}

// TestMigration0296_ApprovalInstances_CheckRejectsInvalidSubjectKind is the
// approval_instances sibling of the CHECK-rejection assertion.
func TestMigration0296_ApprovalInstances_CheckRejectsInvalidSubjectKind(t *testing.T) {
	db, _ := testdb.Open(t)

	doc := testdb.NewDocument(t, db, testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(doc.TenantID))

	err := insertInstanceWithSubject(t, db, doc, route, "invalid", "whatever")
	assertCheckViolation(t, err)
}

// NOTE: template-subject INSERTABILITY (route + instance) is proven by
// migration_0297_test.go's *_TemplateSubject_Null{ProfileCode,DocumentID}_Insertable.
// The 0296-era versions of those tests inserted template rows with the legacy
// profile_code/document_id columns still POPULATED — the only shape possible
// while those columns were NOT NULL (this phase). Migration 0297 relaxes that
// NOT NULL and adds a projection CHECK requiring template rows to leave the
// legacy column NULL, which is the correct final invariant; the transitional
// 0296 shape is superseded, so those two tests moved to migration_0297_test.go
// rather than being kept here asserting a now-illegal row shape.

// TestMigration0296_NewSubjectIndexesExist confirms the additive tenant-scoped
// subject-keyed indexes exist ALONGSIDE the pre-existing document-keyed ones
// (expand-only: nothing dropped).
func TestMigration0296_NewSubjectIndexesExist(t *testing.T) {
	db, _ := testdb.Open(t)

	wantIndexes := []struct {
		table string
		index string
	}{
		{"approval_routes", "ux_approval_routes_tenant_subject"},
		{"approval_routes", "approval_routes_active_profile_uq"},  // kept (0287 route-versioning)
		{"approval_routes", "approval_routes_profile_version_uq"}, // kept (0287 route-versioning)
		{"approval_instances", "ux_approval_instances_active_subject"},
		{"approval_instances", "ix_approval_instances_tenant_subject"},
		{"approval_instances", "approval_instances_document_id_idempotency_key_key"}, // kept
		{"approval_instances", "ux_approval_instances_active_document_id"},           // kept
		{"approval_instances", "ix_approval_instances_tenant_document_id"},           // kept
	}

	for _, w := range wantIndexes {
		var found bool
		err := db.QueryRowContext(context.Background(),
			`SELECT EXISTS (
				SELECT 1 FROM pg_indexes
				 WHERE schemaname = 'public' AND tablename = $1 AND indexname = $2
			)`, w.table, w.index,
		).Scan(&found)
		if err != nil {
			t.Fatalf("query pg_indexes for %s.%s: %v", w.table, w.index, err)
		}
		if !found {
			t.Fatalf("expected index %s on %s to exist", w.index, w.table)
		}
	}
}

// TestMigration0296_ApprovalRoutes_UniqueSubjectIndexRejectsDuplicate proves
// ux_approval_routes_tenant_subject actually enforces uniqueness on
// (tenant_id, subject_kind, subject_key). The index is PARTIAL (WHERE active)
// mirroring approval_routes_active_profile_uq (0287 route versioning), so the
// duplicate must be exercised with active=true rows — two inactive rows
// sharing a subject_key is legitimate (superseded route history) and must NOT
// collide.
func TestMigration0296_ApprovalRoutes_UniqueSubjectIndexRejectsDuplicate(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	// Template routes carry NULL profile_code (0297 projection CHECK), so the
	// only unique index that can fire on two active same-subject rows is
	// ux_approval_routes_tenant_subject — profile-keyed indexes never collide on
	// NULL. (insertRouteWithSubjectAndProfile is defined in migration_0297_test.go,
	// same package.)
	sharedKey := "tpl:" + uuid.NewString()
	if err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, true, 1, "template", sharedKey); err != nil {
		t.Fatalf("insert first template route: %v", err)
	}
	err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, true, 2, "template", sharedKey)
	assertUniqueViolation(t, err)
}

// TestMigration0296_ApprovalRoutes_InactiveDuplicateSubjectAllowed proves the
// index does NOT block two historical (active=false) rows sharing the same
// subject_key — the exact shape route supersession produces (0287), and the
// regression this partial-index design fixes:
// TestRouteVersioning_InUse_DefinitionUpdateBlocked_ThenSupersede failed
// 23505 against an earlier non-partial draft of this index.
func TestMigration0296_ApprovalRoutes_InactiveDuplicateSubjectAllowed(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	// NULL profile_code (template rows) — see UniqueSubjectIndexRejectsDuplicate.
	sharedKey := "tpl:" + uuid.NewString()
	if err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, false, 1, "template", sharedKey); err != nil {
		t.Fatalf("insert first inactive route: %v", err)
	}
	if err := insertRouteWithSubjectAndProfile(t, db, tenant.ID, nil, owner.ID, false, 2, "template", sharedKey); err != nil {
		t.Fatalf("insert second inactive route with same subject_key must be allowed (partial index WHERE active), got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// insertRouteWithSubject inserts a route at the given version. Callers that
// insert two rows sharing the same (tenant, profileCode) — e.g. to exercise
// subject_key collisions — must pass distinct versions, since
// approval_routes_profile_version_uq (tenant_id, profile_code, version) is a
// pre-existing (0287) full-unique constraint independent of subject_key.
func insertRouteWithSubject(t *testing.T, db *sql.DB, tenantID, profileCode, ownerID string, active bool, version int, subjectKind, subjectKey string) error {
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
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.approval_routes
		   (id, tenant_id, name, profile_code, version, active, created_by, subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, 'Route', $3, $4, $5, $6, $7, $8)`,
		id, tenantID, profileCode, version, active, ownerID, subjectKind, subjectKey,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func insertInstanceWithSubject(t *testing.T, db *sql.DB, doc testdb.Document, route testdb.ApprovalRoute, subjectKind, subjectKey string) error {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.tenant_id', $1, true)`, doc.TenantID,
	); err != nil {
		t.Fatalf("set tenant_id GUC: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, `[{"cap":"document.submit"}]`,
	); err != nil {
		t.Fatalf("assert document.submit cap: %v", err)
	}

	id := uuid.NewString()
	idemKey := "idem-" + uuid.NewString()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.approval_instances
		   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
		    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		    subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'in_progress', $5, now(), repeat('a', 64), $6, $7, $8)`,
		id, doc.TenantID, doc.ID, route.ID, doc.Owner, idemKey, subjectKind, subjectKey,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func assertCheckViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a CHECK constraint violation, got nil error")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23514" {
			t.Fatalf("expected SQLSTATE 23514 (check_violation), got %s: %v", pgErr.Code, err)
		}
		return
	}
	if !strings.Contains(err.Error(), "check") {
		t.Fatalf("expected a check-constraint violation, got: %v", err)
	}
}

func assertUniqueViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a UNIQUE constraint violation, got nil error")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23505" {
			t.Fatalf("expected SQLSTATE 23505 (unique_violation), got %s: %v", pgErr.Code, err)
		}
		return
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected a unique-constraint violation, got: %v", err)
	}
}
