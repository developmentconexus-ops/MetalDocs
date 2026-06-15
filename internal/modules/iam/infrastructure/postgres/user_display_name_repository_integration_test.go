//go:build integration

package postgres_test

import (
	"context"
	"testing"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
)

// TestUserDisplayNameRepository_DisplayName_Live exercises DisplayName against a
// live Postgres instance. Seeds one metaldocs.iam_users row, asserts present→value
// and absent→"", and verifies tenant-scoping (same user_id in another tenant is
// not returned).
func TestUserDisplayNameRepository_DisplayName_Live(t *testing.T) {
	db := openLiveIAMDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := iampg.NewUserDisplayNameRepository(db)

	const tenantID = "f41a0000-0000-4000-8000-0000000000aa"
	const otherTenantID = "f41a0000-0000-4000-8000-0000000000bb"
	const userID = "dn-reader-int-f41a"
	const displayName = "F4.1 Test Actor"

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id IN ($1::uuid, $2::uuid)`, tenantID, otherTenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	for _, tid := range []string{tenantID, otherTenantID} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, 'F4.1 Tenant '||$1, 'f41-'||$1)
			 ON CONFLICT (id) DO NOTHING`,
			tid,
		); err != nil {
			t.Fatalf("insert tenant %s: %v", tid, err)
		}
	}

	// Seed user only in tenantID — not in otherTenantID.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, TRUE, $3::uuid, now(), now())`,
		userID, displayName, tenantID,
	); err != nil {
		t.Fatalf("insert iam_users: %v", err)
	}

	t.Run("present_returns_value", func(t *testing.T) {
		got, err := repo.DisplayName(ctx, tenantID, userID)
		if err != nil {
			t.Fatalf("DisplayName: %v", err)
		}
		if got != displayName {
			t.Fatalf("DisplayName = %q, want %q", got, displayName)
		}
	})

	t.Run("absent_returns_empty", func(t *testing.T) {
		got, err := repo.DisplayName(ctx, tenantID, "no-such-user-f41")
		if err != nil {
			t.Fatalf("DisplayName(missing): %v", err)
		}
		if got != "" {
			t.Fatalf("DisplayName(missing) = %q, want empty", got)
		}
	})

	t.Run("tenant_scoped_other_tenant_returns_empty", func(t *testing.T) {
		got, err := repo.DisplayName(ctx, otherTenantID, userID)
		if err != nil {
			t.Fatalf("DisplayName(other tenant): %v", err)
		}
		if got != "" {
			t.Fatalf("DisplayName(other tenant) = %q, want empty (tenant-scoped)", got)
		}
	})
}

// TestUserDisplayNameRepository_DisplayNames_Live exercises DisplayNames batch
// against a live Postgres instance. Asserts present users are returned, absent
// users are omitted, and empty display_name rows are omitted.
func TestUserDisplayNameRepository_DisplayNames_Live(t *testing.T) {
	db := openLiveIAMDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := iampg.NewUserDisplayNameRepository(db)

	const tenantID = "f41b0000-0000-4000-8000-0000000000cc"
	const userA = "dn-batch-a-f41b"
	const userB = "dn-batch-b-f41b"
	const userC = "dn-batch-c-f41b" // empty display_name — should be omitted

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id IN ($1, $2, $3)`, userA, userB, userC)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id = $1::uuid`, tenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug) VALUES ($1::uuid, 'F4.1B Batch Tenant', 'f41b-batch') ON CONFLICT (id) DO NOTHING`,
		tenantID,
	); err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	seeds := []struct {
		userID      string
		displayName string
	}{
		{userA, "Actor A"},
		{userB, "Actor B"},
		{userC, ""}, // empty string display_name → should be omitted from batch result
	}
	for _, s := range seeds {
		// Use empty string (not NULL) — the schema enforces NOT NULL on display_name.
		// The port query filters WHERE display_name <> '' so empty-string rows are omitted.
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
			 VALUES ($1, $2, TRUE, $3::uuid, now(), now())`,
			s.userID, s.displayName, tenantID,
		); err != nil {
			t.Fatalf("insert iam_users %s: %v", s.userID, err)
		}
	}

	t.Run("present_with_name_returned", func(t *testing.T) {
		got, err := repo.DisplayNames(ctx, tenantID, []string{userA, userB, userC, "no-such-user-f41b"})
		if err != nil {
			t.Fatalf("DisplayNames: %v", err)
		}
		if got[userA] != "Actor A" {
			t.Fatalf("DisplayNames[%s] = %q, want %q", userA, got[userA], "Actor A")
		}
		if got[userB] != "Actor B" {
			t.Fatalf("DisplayNames[%s] = %q, want %q", userB, got[userB], "Actor B")
		}
		if _, ok := got[userC]; ok {
			t.Fatalf("DisplayNames[%s] = %q, want omitted (empty display_name)", userC, got[userC])
		}
		if _, ok := got["no-such-user-f41b"]; ok {
			t.Fatal("DisplayNames: absent user must be omitted")
		}
		if len(got) != 2 {
			t.Fatalf("len(DisplayNames) = %d, want 2", len(got))
		}
	})

	t.Run("empty_input_returns_empty_map", func(t *testing.T) {
		got, err := repo.DisplayNames(ctx, tenantID, []string{})
		if err != nil {
			t.Fatalf("DisplayNames(empty): %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("DisplayNames(empty) = %v, want empty map", got)
		}
	})
}
