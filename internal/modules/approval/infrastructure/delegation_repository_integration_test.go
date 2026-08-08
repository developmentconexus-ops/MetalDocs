//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// F9 (spec.md, ADR 0077) — approval_delegations repository proof against real
// Postgres. Covers: insert + fresh in-tx active-window load, boundary
// windows (not-yet-active / expired / active), real-time revocation (hard
// delete, immediately reflected by a fresh LoadActiveDelegationsFor call —
// never cached), and the ownership-gated delete (delegator-only vs
// oversee-widened).

func newTestDelegationRow(t *testing.T, tenantID, delegatorID, delegateID string, startsAt, endsAt time.Time) domain.Delegation {
	t.Helper()
	d, err := domain.NewDelegation(domain.DelegationParams{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		DelegatorID: delegatorID,
		DelegateID:  delegateID,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Reason:      "integration test delegation",
		CreatedBy:   delegatorID,
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("newTestDelegationRow: domain.NewDelegation: %v", err)
	}
	return *d
}

func TestApprovalDelegations_InsertAndLoadActive_WindowBoundaries(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	now := time.Now().UTC()

	active := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	future := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(2*time.Hour), now.Add(3*time.Hour))
	expired := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-3*time.Hour), now.Add(-2*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	for _, d := range []domain.Delegation{active, future, expired} {
		if err := repo.InsertDelegation(ctx, tx, d); err != nil {
			t.Fatalf("InsertDelegation(%s): %v", d.ID(), err)
		}
	}

	got, err := repo.LoadActiveDelegationsFor(ctx, tx, tenant.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("LoadActiveDelegationsFor at now = %d delegations, want exactly 1 (only the active window); got %#v", len(got), got)
	}
	if got[0].ID() != active.ID() {
		t.Fatalf("LoadActiveDelegationsFor returned delegation %s, want the active-window one %s", got[0].ID(), active.ID())
	}
	if got[0].DelegatorID() != delegator.ID {
		t.Fatalf("returned delegation delegator_id = %s, want %s", got[0].DelegatorID(), delegator.ID)
	}
}

func TestApprovalDelegations_OverlappingWindows_BothUsable(t *testing.T) {
	// F9/ADR 0077 §Interview #7: overlapping delegation windows for the same
	// delegator->delegate pair are explicitly allowed (union semantics, no
	// uniqueness constraint) — both rows must be visible when active.
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	now := time.Now().UTC()
	first := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(2*time.Hour))
	second := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-30*time.Minute), now.Add(3*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.InsertDelegation(ctx, tx, first); err != nil {
		t.Fatalf("InsertDelegation(first): %v", err)
	}
	if err := repo.InsertDelegation(ctx, tx, second); err != nil {
		t.Fatalf("InsertDelegation(second): %v", err)
	}

	got, err := repo.LoadActiveDelegationsFor(ctx, tx, tenant.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("LoadActiveDelegationsFor with two overlapping active windows = %d rows, want 2", len(got))
	}
}

func TestApprovalDelegations_Revoke_RealTimeEnforcement(t *testing.T) {
	// F9/ADR 0077 §5: revocation is a hard DELETE; a fresh in-tx
	// LoadActiveDelegationsFor call immediately stops seeing the row — there
	// is no cache and no soft revoked_at flag to go stale.
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	now := time.Now().UTC()
	d := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.InsertDelegation(ctx, tx, d); err != nil {
		t.Fatalf("InsertDelegation: %v", err)
	}

	before, err := repo.LoadActiveDelegationsFor(ctx, tx, tenant.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor (before revoke): %v", err)
	}
	if len(before) != 1 {
		t.Fatalf("before revoke: LoadActiveDelegationsFor = %d rows, want 1", len(before))
	}

	deleted, err := repo.DeleteDelegation(ctx, tx, tenant.ID, d.ID(), delegator.ID, false)
	if err != nil {
		t.Fatalf("DeleteDelegation: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteDelegation by the delegator = false, want true")
	}

	after, err := repo.LoadActiveDelegationsFor(ctx, tx, tenant.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor (after revoke): %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("after revoke: LoadActiveDelegationsFor = %d rows, want 0 (real-time revocation)", len(after))
	}
}

func TestApprovalDelegations_Delete_NonOwnerWithoutOversee_Denied(t *testing.T) {
	// Ownership gate: a caller who is neither the delegator nor
	// oversee-flagged cannot delete another user's delegation grant. The
	// repository enforces this atomically in the DELETE's WHERE clause
	// (delegator_id = $3 OR $4::boolean) — this proves the false branch.
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	stranger := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	now := time.Now().UTC()
	d := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.InsertDelegation(ctx, tx, d); err != nil {
		t.Fatalf("InsertDelegation: %v", err)
	}

	deleted, err := repo.DeleteDelegation(ctx, tx, tenant.ID, d.ID(), stranger.ID, false)
	if err != nil {
		t.Fatalf("DeleteDelegation (non-owner, no oversee): %v", err)
	}
	if deleted {
		t.Fatal("DeleteDelegation by a non-owner without oversee = true, want false")
	}

	// The row must still be there and still active.
	got, err := repo.LoadActiveDelegationsFor(ctx, tx, tenant.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("after denied delete attempt: LoadActiveDelegationsFor = %d rows, want 1 (unaffected)", len(got))
	}
}

func TestApprovalDelegations_Delete_OverseeCanRevokeAnothersGrant(t *testing.T) {
	// Oversee-widened scope: a caller flagged callerIsOversee=true may delete
	// a delegation they did not create (the tenant-grade CapApprovalOversee
	// capability check itself happens one layer up in DelegationService; this
	// proves the repository's WHERE clause honors the flag once granted).
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	overseer := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	now := time.Now().UTC()
	d := newTestDelegationRow(t, tenant.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.InsertDelegation(ctx, tx, d); err != nil {
		t.Fatalf("InsertDelegation: %v", err)
	}

	deleted, err := repo.DeleteDelegation(ctx, tx, tenant.ID, d.ID(), overseer.ID, true)
	if err != nil {
		t.Fatalf("DeleteDelegation (oversee): %v", err)
	}
	if !deleted {
		t.Fatal("DeleteDelegation by an oversee-flagged caller = false, want true")
	}
}

func TestApprovalDelegations_TenantIsolation(t *testing.T) {
	// RLS FORCE + tenant_isolation policy on approval_delegations: a
	// delegation seeded under tenant A must never surface in a
	// LoadActiveDelegationsFor call scoped to tenant B, even for the same
	// delegate user_id string (defense in depth alongside the app-level
	// tenant_id predicate already in the query).
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	delegator := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID))
	delegate := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID))

	now := time.Now().UTC()
	d := newTestDelegationRow(t, tenantA.ID, delegator.ID, delegate.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.InsertDelegation(ctx, tx, d); err != nil {
		t.Fatalf("InsertDelegation: %v", err)
	}

	gotB, err := repo.LoadActiveDelegationsFor(ctx, tx, tenantB.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor (tenant B): %v", err)
	}
	if len(gotB) != 0 {
		t.Fatalf("LoadActiveDelegationsFor scoped to tenant B = %d rows, want 0 (tenant A's delegation must not leak)", len(gotB))
	}

	gotA, err := repo.LoadActiveDelegationsFor(ctx, tx, tenantA.ID, delegate.ID, now)
	if err != nil {
		t.Fatalf("LoadActiveDelegationsFor (tenant A): %v", err)
	}
	if len(gotA) != 1 {
		t.Fatalf("LoadActiveDelegationsFor scoped to tenant A = %d rows, want 1", len(gotA))
	}
}
