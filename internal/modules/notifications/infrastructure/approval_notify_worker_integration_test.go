//go:build integration
// +build integration

package notificationsinfra_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// countNotificationsAsCI counts a tenant's notification rows through the
// non-owner metaldocs_ci role, inside a tx seeded with SET LOCAL
// metaldocs.tenant_id — the same isolation-proof shape used by
// tests/integration/security/rls_truth_test.go. metaldocs_app (the owner
// handle every other helper in this package writes through) is
// superuser+BYPASSRLS in dev, so a tenancy assertion made through it is a
// false green; that is exactly how the M6 surfacer tenancy bug survived
// review. Reading through metaldocs_ci makes RLS genuinely filter.
func countNotificationsAsCI(t *testing.T, ci *sql.DB, tenantID string) int {
	t.Helper()
	ctx := context.Background()

	tx, err := ci.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("countNotificationsAsCI: begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SET LOCAL metaldocs.tenant_id = '`+tenantID+`'`); err != nil {
		t.Fatalf("countNotificationsAsCI: SET LOCAL tenant_id: %v", err)
	}

	var n int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.notifications WHERE tenant_id = $1::uuid`,
		tenantID,
	).Scan(&n); err != nil {
		t.Fatalf("countNotificationsAsCI: count: %v", err)
	}
	return n
}

// Run the DB-mutating side through the owner handle (db, from testdb.Open) —
// worker.Work itself opens its own tx via authz.SeedTxTenant, so this is the
// worker's normal production shape. Isolation is then PROVED by reading back
// through countNotificationsAsCI (metaldocs_ci), not through the owner
// connection worker.Work wrote with.
func TestApprovalNotifyWorker_InsertsOnePerRecipient_TenantScoped(t *testing.T) {
	db, dbName := testdb.Open(t)
	ci := testdb.OpenAsCIRole(t, dbName)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	userA1 := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID))
	userA2 := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID))
	_ = testdb.NewUser(t, db, testdb.WithTenant(tenantB.ID))

	w := notificationsinfra.NewApprovalNotifyWorker(db)
	args := approvaldomain.ApprovalNotificationArgs{
		EventID:          uuid.NewString(),
		TenantID:         tenantA.ID,
		EventType:        approvaldomain.EventTypeApprovalPending,
		SubjectKind:      "template",
		SubjectID:        uuid.NewString(),
		RecipientUserIDs: []string{userA1.ID, userA2.ID},
	}

	job := &river.Job[approvaldomain.ApprovalNotificationArgs]{Args: args}
	if err := w.Work(context.Background(), job); err != nil {
		t.Fatalf("Work: %v", err)
	}

	if got := countNotificationsAsCI(t, ci, tenantA.ID); got != 2 {
		t.Fatalf("tenant A notifications = %d, want 2", got)
	}
	if got := countNotificationsAsCI(t, ci, tenantB.ID); got != 0 {
		t.Fatalf("tenant B notifications = %d, want 0 — the event is single-tenant by construction", got)
	}
}

// Redelivery is a no-op: the same EventID lands on the partial unique index
// (recipient_user_id, source_event_id) and DOES NOTHING.
func TestApprovalNotifyWorker_SameEventTwice_InsertsOnce(t *testing.T) {
	db, dbName := testdb.Open(t)
	ci := testdb.OpenAsCIRole(t, dbName)

	tenant := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	w := notificationsinfra.NewApprovalNotifyWorker(db)
	args := approvaldomain.ApprovalNotificationArgs{
		EventID:          uuid.NewString(),
		TenantID:         tenant.ID,
		EventType:        approvaldomain.EventTypeApprovalPending,
		SubjectKind:      "document",
		SubjectID:        uuid.NewString(),
		RecipientUserIDs: []string{user.ID},
	}

	for i := 0; i < 2; i++ {
		job := &river.Job[approvaldomain.ApprovalNotificationArgs]{Args: args}
		if err := w.Work(context.Background(), job); err != nil {
			t.Fatalf("Work %d: %v", i, err)
		}
	}

	if got := countNotificationsAsCI(t, ci, tenant.ID); got != 1 {
		t.Fatalf("notifications = %d, want 1 — redelivery must be idempotent", got)
	}
}

// An empty recipient list is NOT an error: a livre route (ADR 0087) has zero
// stages and therefore no eligible actors.
func TestApprovalNotifyWorker_NoRecipients_IsNoOp(t *testing.T) {
	db, dbName := testdb.Open(t)
	ci := testdb.OpenAsCIRole(t, dbName)
	tenant := testdb.NewTenant(t, db)

	w := notificationsinfra.NewApprovalNotifyWorker(db)
	job := &river.Job[approvaldomain.ApprovalNotificationArgs]{Args: approvaldomain.ApprovalNotificationArgs{
		EventID:   uuid.NewString(),
		TenantID:  tenant.ID,
		EventType: approvaldomain.EventTypeApprovalPending,
	}}
	if err := w.Work(context.Background(), job); err != nil {
		t.Fatalf("empty recipients must be a no-op, got error: %v", err)
	}
	if got := countNotificationsAsCI(t, ci, tenant.ID); got != 0 {
		t.Fatalf("notifications = %d, want 0", got)
	}
}

// An unrecognised event type is an ERROR, never a silent drop. River retries
// and then dead-letters, so the failure leaves a trace.
func TestApprovalNotifyWorker_UnknownEventType_Errors(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	w := notificationsinfra.NewApprovalNotifyWorker(db)
	job := &river.Job[approvaldomain.ApprovalNotificationArgs]{Args: approvaldomain.ApprovalNotificationArgs{
		EventID:          uuid.NewString(),
		TenantID:         tenant.ID,
		EventType:        "approval.something_nobody_registered",
		RecipientUserIDs: []string{user.ID},
	}}
	if err := w.Work(context.Background(), job); err == nil {
		t.Fatal("err = nil, want an error — a dropped event type must never be silent")
	}
}
