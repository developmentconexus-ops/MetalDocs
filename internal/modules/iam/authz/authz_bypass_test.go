//go:build integration

package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

func TestRequire_SystemAdmin_Bypasses(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	actorID := uuid.NewString()
	testdb.SeedSystemAdmin(t, db, tnt.ID, actorID, "Admin")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, actorID); err != nil {
		t.Fatalf("seed tx identity: %v", err)
	}

	// system_admin should bypass even a non-existent capability
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "does.not.exist", "any-area"); err != nil {
		t.Fatalf("system_admin bypass failed: %v", err)
	}
}
