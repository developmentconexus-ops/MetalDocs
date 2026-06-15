//go:build integration

package authz_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

func TestRequire_SystemAdmin_Bypasses(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	// Mint a fresh tenant UUID — no hardcoded literals.
	tenant := testdb.NewTenant(t, db)

	// Seed the system_admin actor: iam_users row (no tripwire) + iam_user_roles
	// row (user.manage tripwire asserted tx-locally by SeedSystemAdmin).
	const actorID = "admin-local"
	testdb.SeedSystemAdmin(t, db, tenant.ID, actorID, "Administrator")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Set the transaction-local authz-context GUCs (actor_id + tenant_id) that
	// authz.Require reads via MustActorID / MustTenantID. These are NOT the
	// cap-tripwire metaldocs.asserted_caps GUC.
	if err := authz.SeedTxIdentity(ctx, tx, tenant.ID, actorID); err != nil {
		t.Fatalf("seed tx identity: %v", err)
	}

	// system_admin should bypass even a non-existent capability.
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "does.not.exist", "any-area"); err != nil {
		t.Fatalf("system_admin bypass failed: %v", err)
	}
}
