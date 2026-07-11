//go:build integration

package migrations_test

import (
	"context"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0153_TenantMismatchRejected verifies that inserting a
// document_placeholder_values row with a tenant_id that differs from the
// owning revision's tenant_id is rejected by the trigger.
//
// revision_id targets public.document_revisions(id) (see seedDocumentRevision
// in migration_0152_test.go, same package) — document_revisions carries no
// tenant_id column itself; tenancy is derived through documents, which is why
// the mismatch must be seeded against a real revision, not a bare document id.
func TestMigration0153_TenantMismatchRejected(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tenantA := testdb.DeterministicID(t, "tenant-a")
	tenantB := testdb.DeterministicID(t, "tenant-b")
	docID := testdb.DeterministicID(t, "doc")
	userID := testdb.DeterministicID(t, "user")

	testdb.SeedUser(t, ctx, db, schema, userID, "Tenant Mismatch User")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantA, userID)
	revID := seedDocumentRevision(t, ctx, db, schema, docID, tenantA, userID)

	_, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "document_placeholder_values")+`
		 (tenant_id, revision_id, placeholder_id, source)
		 VALUES ($1::uuid, $2::uuid, 'ph_title', 'user')`,
		tenantB, revID,
	)
	if err == nil {
		t.Fatal("expected tenant mismatch error from trigger, got nil")
	}
	if !strings.Contains(err.Error(), "tenant mismatch") &&
		!strings.Contains(err.Error(), "check_violation") &&
		!strings.Contains(err.Error(), "23514") {
		t.Fatalf("unexpected error (wanted tenant mismatch): %v", err)
	}
}

// TestMigration0153_TenantMatchAccepted verifies that inserting with the
// correct tenant_id succeeds.
func TestMigration0153_TenantMatchAccepted(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tenantA := testdb.DeterministicID(t, "tenant-a")
	docID := testdb.DeterministicID(t, "doc")
	userID := testdb.DeterministicID(t, "user")

	testdb.SeedUser(t, ctx, db, schema, userID, "Tenant Match User")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantA, userID)
	revID := seedDocumentRevision(t, ctx, db, schema, docID, tenantA, userID)

	_, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "document_placeholder_values")+`
		 (tenant_id, revision_id, placeholder_id, source)
		 VALUES ($1::uuid, $2::uuid, 'ph_body', 'user')`,
		tenantA, revID,
	)
	if err != nil {
		t.Fatalf("insert with matching tenant should succeed: %v", err)
	}
}

// TestMigration0153_ZoneContentTenantMismatchRejected was DELETED: it targeted
// public.document_editable_zone_content, which archive/migrations/0157_drop_editable_zones.sql
// dropped (confirmed absent from db/baseline/0001_current_schema.sql — no
// CREATE TABLE for it anywhere in the baseline or db/migrations; a live probe
// against an isolated test DB returned SQLSTATE 42P01 relation does not exist).
// The table and its tenant-consistency trigger are retired; no replacement
// coverage is owed since the feature itself (editable zones) was removed.
