//go:build integration
// +build integration

package notificationsinfra_test

import (
	"context"
	"testing"

	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/tests/integration/testdb"
)

func TestNotifications(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	repo := notificationsinfra.NewNotificationsRepository(db)

	t.Run("self_scope_isolation", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		userA := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		userB := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(userA.ID))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(userA.ID))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(userB.ID))

		page, err := repo.List(ctx, ten.ID, userA.ID, "", "", 100)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(page.Items) != 2 {
			t.Fatalf("want 2 items for A, got %d", len(page.Items))
		}
		for _, it := range page.Items {
			if it.RecipientUserID != userA.ID {
				t.Errorf("self-scope leak: got row for %s, want only %s", it.RecipientUserID, userA.ID)
			}
		}
	})

	t.Run("status_filter", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		u := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("PENDING"))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("READ"))

		page, err := repo.List(ctx, ten.ID, u.ID, "READ", "", 100)
		if err != nil {
			t.Fatalf("List(status=READ): %v", err)
		}
		if len(page.Items) != 1 || page.Items[0].Status != "READ" {
			t.Fatalf("want 1 READ row, got %d", len(page.Items))
		}
	})

	t.Run("cursor_stability", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		u := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		for i := 0; i < 25; i++ {
			testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID))
		}
		seen := map[string]bool{}
		cursor := ""
		pages := 0
		for {
			page, err := repo.List(ctx, ten.ID, u.ID, "", cursor, 10)
			if err != nil {
				t.Fatalf("List page: %v", err)
			}
			pages++
			for _, it := range page.Items {
				if seen[it.ID] {
					t.Fatalf("duplicate row across pages: %s", it.ID)
				}
				seen[it.ID] = true
			}
			if !page.HasMore {
				break
			}
			cursor = page.NextCursor
			if pages > 10 {
				t.Fatalf("pagination did not terminate")
			}
		}
		if len(seen) != 25 {
			t.Fatalf("want 25 distinct rows across pages, got %d", len(seen))
		}
	})

	t.Run("unread_count_accuracy", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		u := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		other := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("PENDING"))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("SENT"))
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("READ"))
		// Another user's PENDING must not count toward u's unread.
		testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(other.ID), testdb.WithStatus("PENDING"))

		n, err := repo.UnreadCount(ctx, ten.ID, u.ID)
		if err != nil {
			t.Fatalf("UnreadCount: %v", err)
		}
		if n != 2 { // PENDING + SENT, not READ, not other user's
			t.Fatalf("want unread=2, got %d", n)
		}
	})

	t.Run("mark_read_flips_and_idempotent", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		u := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		notif := testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(u.ID), testdb.WithStatus("PENDING"))

		if err := repo.MarkRead(ctx, ten.ID, notif.ID, u.ID); err != nil {
			t.Fatalf("MarkRead: %v", err)
		}
		page, err := repo.List(ctx, ten.ID, u.ID, "READ", "", 10)
		if err != nil {
			t.Fatalf("List after MarkRead: %v", err)
		}
		if len(page.Items) != 1 || page.Items[0].ReadAt == nil {
			t.Fatalf("want 1 READ row with read_at set, got %d", len(page.Items))
		}
		// Idempotent: second call is a no-op, still no error.
		if err := repo.MarkRead(ctx, ten.ID, notif.ID, u.ID); err != nil {
			t.Fatalf("MarkRead idempotent re-run: %v", err)
		}
		n, _ := repo.UnreadCount(ctx, ten.ID, u.ID)
		if n != 0 {
			t.Fatalf("want unread=0 after mark-read, got %d", n)
		}
	})

	t.Run("mark_read_wrong_owner_noop", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		owner := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		attacker := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		notif := testdb.NewNotification(t, db, testdb.WithTenant(ten.ID), testdb.WithRecipient(owner.ID), testdb.WithStatus("PENDING"))

		// Attacker tries to mark the owner's row read → silent no-op, no error.
		if err := repo.MarkRead(ctx, ten.ID, notif.ID, attacker.ID); err != nil {
			t.Fatalf("MarkRead wrong-owner returned error, want silent no-op: %v", err)
		}
		// The owner's row is unchanged (still unread).
		n, _ := repo.UnreadCount(ctx, ten.ID, owner.ID)
		if n != 1 {
			t.Fatalf("want owner's row still unread (1), got %d", n)
		}
	})
}
