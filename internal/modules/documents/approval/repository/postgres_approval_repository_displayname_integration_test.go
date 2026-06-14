//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/internal/modules/documents/approval/repository"
)

// TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema is the F1.3 (AC5) runtime
// proof. It connects DIRECTLY to the live, already-migrated dev database (not a
// freshly-built throwaway), seeds one metaldocs.iam_users row, and exercises the
// real off-tx signoff display-name read. Proves, against the live schema + the
// live RLS NULL-permissive policy (migration 0237) with no GUC set and the
// explicit tenant_id = $N::uuid predicate, that the contained read on the pool
// (r.db) — outside any lock-holding signoff transaction (H-PRE-1) — returns the
// actor's real display_name and is empty-on-missing (AC3).
func TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	if dsn == "" {
		t.Skip("METALDOCS_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	const tenantID = "f13a0000-0000-4000-8000-0000000000aa"
	const userID = "approver-displayname-int-f13"
	const displayName = "Alice Approver"

	// Cleanup any prior run, and clean up after this run.
	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id = $1::uuid`, tenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug)
		 VALUES ($1::uuid, 'F1.3 Int Tenant', 'f13-int')
		 ON CONFLICT (id) DO NOTHING`,
		tenantID,
	); err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, TRUE, $3::uuid, now(), now())`,
		userID, displayName, tenantID,
	); err != nil {
		t.Fatalf("insert iam_users: %v", err)
	}

	repo := repository.NewPostgresApprovalRepository(db)

	// AC3 + AC5: the off-tx read returns the actor's real display_name.
	got, err := repo.LoadActorDisplayName(ctx, tenantID, userID)
	if err != nil {
		t.Fatalf("LoadActorDisplayName: %v", err)
	}
	if got != displayName {
		t.Fatalf("display_name = %q, want %q", got, displayName)
	}
	t.Logf("AC5 live read-back: LoadActorDisplayName(%s, %s) = %q", tenantID, userID, got)

	// AC3: empty-on-missing (no row) returns "" and a nil error.
	missing, err := repo.LoadActorDisplayName(ctx, tenantID, "no-such-user-f13")
	if err != nil {
		t.Fatalf("LoadActorDisplayName(missing): %v", err)
	}
	if missing != "" {
		t.Fatalf("missing display_name = %q, want empty", missing)
	}
	t.Logf("AC3 empty-on-missing: LoadActorDisplayName(missing) = %q (nil err)", missing)
}
