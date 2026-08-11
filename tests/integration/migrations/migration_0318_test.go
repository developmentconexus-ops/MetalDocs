//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestMigration0318_SchemaLanded proves the two new relations and the
// prerequisite unique keys exist after bootstrap, and that the ledger row was
// written.
func TestMigration0318_SchemaLanded(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	if got := regclass(t, ctx, db, "metaldocs.capability_bindings"); !got.Valid {
		t.Fatal("metaldocs.capability_bindings does not exist after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.roles"); !got.Valid {
		t.Fatal("metaldocs.roles does not exist after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.iam_users_tenant_user_uk"); !got.Valid {
		t.Fatal("iam_users_tenant_user_uk (promoted unique constraint) is missing after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.iam_groups_tenant_id_id_uk"); !got.Valid {
		t.Fatal("iam_groups_tenant_id_id_uk is missing after bootstrap")
	}

	if n := ledgerCount(t, ctx, db, "0318"); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0318', got %d", n)
	}

	var roleCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.roles`).Scan(&roleCount); err != nil {
		t.Fatalf("count metaldocs.roles: %v", err)
	}
	if roleCount != 8 {
		t.Fatalf("expected 8 hand-seeded roles, got %d", roleCount)
	}
}

// TestMigration0318_BackfillPreservesHistory proves the backfill is a
// straight row-count carryover from each source relation (provenance
// preserved, nothing flattened) on a fresh bootstrap with only reference-data
// seeded (no dev-seed rows), so the expected counts are the reference-data
// grant rows themselves.
func TestMigration0318_BackfillPreservesHistory(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	for _, tc := range []struct {
		source string
		countQ string
	}{
		{"iam_user_roles", `SELECT count(*) FROM metaldocs.iam_user_roles`},
		{"user_process_areas", `SELECT count(*) FROM public.user_process_areas`},
		{"iam_group_roles", `SELECT count(*) FROM metaldocs.iam_group_roles`},
	} {
		var srcCount, boundCount int
		if err := db.QueryRowContext(ctx, tc.countQ).Scan(&srcCount); err != nil {
			t.Fatalf("count %s: %v", tc.source, err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.capability_bindings WHERE source_relation = $1`,
			tc.source).Scan(&boundCount); err != nil {
			t.Fatalf("count capability_bindings backfilled from %s: %v", tc.source, err)
		}
		if boundCount != srcCount {
			t.Fatalf("source_relation=%s: capability_bindings has %d rows, source table has %d — backfill lost or fabricated rows", tc.source, boundCount, srcCount)
		}
	}
}

// TestMigration0318_RejectsDanglingSubject proves a capability_bindings row
// cannot name a subject that does not exist — RED: before this migration's
// FK, a dangling subject_user_id/subject_group_id was representable and
// silently unenforced by anything but application code.
func TestMigration0318_RejectsDanglingSubject(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	t.Run("dangling subject_user_id", func(t *testing.T) {
		err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
			TenantID:      tenant.DevTenantID,
			SubjectKind:   "user",
			SubjectUserID: sql.NullString{String: "no-such-user-" + uuid.NewString(), Valid: true},
			RoleCode:      "viewer",
			ScopeKind:     "tenant",
		})
		if err == nil {
			t.Fatal("insert with a subject_user_id naming a nonexistent iam_users row succeeded — FK is not enforced")
		}
		if !hasSQLState(err, "23503") {
			t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
		}
	})

	t.Run("dangling subject_group_id", func(t *testing.T) {
		err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
			TenantID:       tenant.DevTenantID,
			SubjectKind:    "group",
			SubjectGroupID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
			RoleCode:       "viewer",
			ScopeKind:      "tenant",
		})
		if err == nil {
			t.Fatal("insert with a subject_group_id naming a nonexistent iam_groups row succeeded — FK is not enforced")
		}
		if !hasSQLState(err, "23503") {
			t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
		}
	})
}

// TestMigration0318_RejectsCrossTenantScope proves a binding cannot name an
// area (scope_ref) that belongs to a DIFFERENT tenant than the binding's own
// tenant_id — the composite FK (tenant_id, scope_ref) -> document_process_areas
// makes a cross-tenant scope reference structurally impossible, matching the
// "cross-tenant URL -> 404, not a leaked row" invariant one level down at the
// schema line.
func TestMigration0318_RejectsCrossTenantScope(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	otherTenant := seedTenant(t, ctx, db)
	otherTenantUser := seedUser(t, ctx, db, otherTenant)
	areaInOtherTenant := seedArea(t, ctx, db, otherTenant)

	// A valid user in DevTenantID's own tenant, so the failure is attributable
	// to the scope FK, not the subject FK.
	devTenantUser := seedUser(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: devTenantUser, Valid: true},
		RoleCode:      "signer",
		ScopeKind:     "area",
		ScopeRef:      sql.NullString{String: areaInOtherTenant, Valid: true},
	})
	if err == nil {
		t.Fatal("insert binding DevTenantID to an area owned by a different tenant succeeded — composite (tenant_id, scope_ref) FK is not enforced")
	}
	if !hasSQLState(err, "23503") {
		t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
	}

	_ = otherTenantUser // seeded for symmetry / future extension; unused directly here
}

// TestMigration0318_RejectsDuplicateActiveBinding proves at most one ACTIVE
// (effective_to IS NULL) binding can exist per (tenant, subject, role,
// scope) — the partial unique index ux_capability_bindings_active_identity.
// A second, already-revoked binding for the same identity must NOT collide
// (history is preserved, not deduplicated away).
func TestMigration0318_RejectsDuplicateActiveBinding(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)

	row := capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: userID, Valid: true},
		RoleCode:      "editor",
		ScopeKind:     "tenant",
	}

	if err := insertCapabilityBinding(t, ctx, db, row); err != nil {
		t.Fatalf("first (active) binding insert failed unexpectedly: %v", err)
	}

	err := insertCapabilityBinding(t, ctx, db, row)
	if err == nil {
		t.Fatal("second ACTIVE binding with the identical (tenant,subject,role,scope) identity succeeded — duplicate active grant is representable")
	}
	if !hasSQLState(err, "23505") {
		t.Fatalf("expected SQLSTATE 23505 (unique_violation) for duplicate active identity, got: %v", err)
	}

	// A REVOKED binding for the exact same identity must NOT collide — history
	// is preserved, not deduplicated. Revoke the first row, then insert a
	// second active one for the same identity: this must succeed. The UPDATE
	// itself is tripwire-guarded (arm #21, same as INSERT), so it needs its
	// own capability assertion and tenant GUC, not just a bare ExecContext.
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("revoke first binding: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenant.DevTenantID,
		); err != nil {
			t.Fatalf("revoke first binding: set tenant GUC: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE metaldocs.capability_bindings
			   SET effective_to = now(), revoked_by = 'test-harness'
			 WHERE tenant_id = $1 AND subject_user_id = $2 AND role_code = $3
			   AND scope_kind = 'tenant' AND effective_to IS NULL`,
			tenant.DevTenantID, userID, row.RoleCode); err != nil {
			t.Fatalf("revoke first binding: update: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("revoke first binding: commit: %v", err)
		}
	}()

	if err := insertCapabilityBinding(t, ctx, db, row); err != nil {
		t.Fatalf("re-granting the same identity after revocation should succeed (history, not dedup), got: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM metaldocs.capability_bindings
		 WHERE tenant_id = $1 AND subject_user_id = $2 AND role_code = $3 AND scope_kind = 'tenant'`,
		tenant.DevTenantID, userID, row.RoleCode).Scan(&n); err != nil {
		t.Fatalf("count history rows: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 history rows (1 revoked + 1 active) for the identity, got %d — history was flattened", n)
	}
}

// TestMigration0318_RejectsSubjectShapeMismatch and
// TestMigration0318_RejectsScopeShapeMismatch prove the discriminated-union
// CHECKs: a 'user' subject cannot also carry a group id (or vice versa), and
// a 'tenant' scope cannot carry an area ref (or vice versa).
func TestMigration0318_RejectsSubjectShapeMismatch(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:       tenant.DevTenantID,
		SubjectKind:    "user",
		SubjectUserID:  sql.NullString{String: userID, Valid: true},
		SubjectGroupID: uuid.NullUUID{UUID: uuid.New(), Valid: true}, // both set — illegal
		RoleCode:       "viewer",
		ScopeKind:      "tenant",
	})
	if err == nil {
		t.Fatal("insert with both subject_user_id and subject_group_id set succeeded — subject-shape CHECK is not enforced")
	}
	if !hasSQLState(err, "23514") {
		t.Fatalf("expected SQLSTATE 23514 (check_violation), got: %v", err)
	}
}

func TestMigration0318_RejectsScopeShapeMismatch(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)
	area := seedArea(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: userID, Valid: true},
		RoleCode:      "viewer",
		ScopeKind:     "tenant",
		ScopeRef:      sql.NullString{String: area, Valid: true}, // scope_kind='tenant' but scope_ref set — illegal
	})
	if err == nil {
		t.Fatal("insert with scope_kind='tenant' and a non-NULL scope_ref succeeded — scope-shape CHECK is not enforced")
	}
	if !hasSQLState(err, "23514") {
		t.Fatalf("expected SQLSTATE 23514 (check_violation), got: %v", err)
	}
}

// ── fixtures ────────────────────────────────────────────────────────────

type capabilityBindingRow struct {
	TenantID       string
	SubjectKind    string
	SubjectUserID  sql.NullString
	SubjectGroupID uuid.NullUUID
	RoleCode       string
	ScopeKind      string
	ScopeRef       sql.NullString
}

// insertCapabilityBinding attempts the write with the same capability
// (user.manage — arm #21, match-one with membership.manage) and tenant GUC
// context the application itself would hold, so a failure is attributable to
// the constraint under test, not to the tripwire or RLS.
func insertCapabilityBinding(t *testing.T, ctx context.Context, db *sql.DB, row capabilityBindingRow) error {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, row.TenantID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO metaldocs.capability_bindings
			(tenant_id, subject_kind, subject_user_id, subject_group_id, role_code, scope_kind, scope_ref)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		row.TenantID, row.SubjectKind, row.SubjectUserID, row.SubjectGroupID, row.RoleCode, row.ScopeKind, row.ScopeRef,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// seedTenant inserts a fresh tenant row (asserting tenant.onboard, the arm
// this table's trigger requires) and returns its id.
func seedTenant(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()
	id := uuid.NewString()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"tenant.onboard"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug) VALUES ($1::uuid, $2, $2)`,
		id, "fixture-"+id[:8]); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed tenant: %v", err)
	}
	return id
}

// seedUser inserts a fresh iam_users row in the given tenant (asserting
// user.manage, arm #15) and returns its user_id.
func seedUser(t *testing.T, ctx context.Context, db *sql.DB, tenantID string) string {
	t.Helper()
	userID := "fixture-user-" + uuid.NewString()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id) VALUES ($1, $1, $2::uuid)`,
		userID, tenantID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed user: %v", err)
	}
	return userID
}

// seedArea inserts a fresh document_process_areas row in the given tenant
// (asserting taxonomy.manage, arm #11) and returns its code.
func seedArea(t *testing.T, ctx context.Context, db *sql.DB, tenantID string) string {
	t.Helper()
	code := "fx" + uuid.NewString()[:8]
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"taxonomy.manage"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.document_process_areas (code, name, tenant_id) VALUES ($1, $1, $2::uuid)`,
		code, tenantID); err != nil {
		t.Fatalf("seed area: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed area: %v", err)
	}
	return code
}
