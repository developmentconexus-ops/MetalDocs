//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/internal/modules/approval/infrastructure"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

// TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema is the F1.3 (AC5) runtime
// proof. It connects to a reset-safe leased clone from the canonical testdb
// factory (ADR 0034), NOT a raw sql.Open against the shared dev database — the
// prior body read METALDOCS_DATABASE_URL directly and skipped-as-green when
// unset, and its comment claimed this probe was "intentionally not part of the
// canonical testdb suite"; that claim was the defect (cluster-4 framework
// bypass), not a deliberate exemption. It seeds one metaldocs.iam_users row and
// exercises the real off-tx signoff display-name read. Proves, against the
// migrated schema and the explicit tenant_id = $N::uuid predicate, that the
// contained read on the pool (r.db) — outside any lock-holding signoff
// transaction (H-PRE-1) — returns the actor's real display_name and is
// empty-on-missing (AC3). The off-tx property (no lock-holding transaction
// wraps this read) is the invariant under test and is unchanged by this
// migration — testdb.Open just supplies the *sql.DB the repository reads from.
//
// Caveat: the leased-clone role is superuser+BYPASSRLS (same posture as the dev
// DB), so RLS (migration 0237) is inert here too — this test does NOT prove RLS
// enforcement, only the off-tx read path against a live, migrated schema.
func TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	const tenantID = "f13a0000-0000-4000-8000-0000000000aa"
	const userID = "approver-displayname-int-f13"
	const displayName = "Alice Approver"

	// Cleanup any prior run, and clean up after this run. metaldocs.iam_users and
	// metaldocs.tenants both carry trg_require_cap_asserted (SEC-05 / migration
	// 0259: user.manage / tenant.onboard), so the deletes must assert those caps
	// tx-locally via the sanctioned testdb helper — a plain db.ExecContext would
	// trip P0001 on a leased clone.
	cleanup := func() {
		testdb.SeedWithCaps(t, db, `[{"cap":"user.manage"},{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id = $1::uuid`, tenantID)
			return nil
		})
	}
	cleanup()
	t.Cleanup(cleanup)

	testdb.SeedWithCaps(t, db, `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, 'F1.3 Int Tenant', 'f13-int')
			 ON CONFLICT (id) DO NOTHING`,
			tenantID,
		)
		return err
	})

	testdb.SeedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
			 VALUES ($1, $2, TRUE, $3::uuid, now(), now())`,
			userID, displayName, tenantID,
		)
		return err
	})

	// M4/F4.1: inject the real iam-owned display-name port so LoadActorDisplayName
	// delegates to it rather than issuing raw iam_users SQL.
	repo := infrastructure.NewPostgresApprovalRepository(db, iampg.NewUserDisplayNameRepository(db))

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
