# Feature F3.2 — Notifications Backend (read surface)

> **Milestone:** 3 — Notifications (full-stack; surface + document-lifecycle emitters)  ·  **Folder:** `f3.2-notifications-backend`
> **Status:** Planning

This is the feature's **execution plan** — the "how" that `milestone.md` deliberately
left out. Produced with `superpowers:writing-plans`.

## Source

- Milestone spec row (F3.2): *"Implement the **read surface** to the Grade-A bar: forward-only migration
  creating the notifications table owned by the `notifications` module (per-recipient rows, `status`
  read-state, `read_at`, tenant-scoped, RLS consistent with the 0237 pattern); repository + handlers in
  `internal/modules/notifications/` for list (cursor pagination per the existing `CursorPage` convention,
  keyset on `(created_at, id)`), unread-count, and mark-read. Reads/writes gated by `CapNotificationRead`
  **self-scoped**. **No emitter in this feature** (F3.3 owns production)."*
  Acceptance: self-scope A≠B, unread-count accuracy, mark-read idempotent, cursor stability,
  `api-lint -strict`=0, 6 CI guards green, `go build`/`vet`/`test ./...` green, publish_service.go diff empty.
- Feature contract: `./spec.md` (approved 2026-06-22).
- Governing ADR: `wiki/decisions/0043-notifications-module-and-lifecycle-bundle.md`.

---

## Plan

# Notifications Backend (F3.2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Grade-A read surface for the `notifications` module — one migration-owned table plus list / unread-count / mark-read endpoints, self-scoped by `CapNotificationRead` + a SQL `recipient_user_id` predicate — implementing the F3.1 generated `StrictServerInterface`.

**Architecture:** New module `internal/modules/notifications/` in the four-layer shape of the distribution module (M2 precedent): `api/` (F3.1, untouched), `domain/` (DTOs), `infrastructure/` (PG repo, own-table reads only), `delivery/http/` (strict handler + routes). Tier-1 route guard enforces the cap; every query additionally filters `recipient_user_id = caller`. Mark-read is a self-scoped idempotent UPDATE whose authz lives one layer up (tripwire-allowlisted, the established pattern for dumb repo writes).

**Tech Stack:** Go 1.22, `database/sql` + Postgres, `oapi-codegen` strict server (generated in F3.1), `internal/platform/pagination` keyset cursor, `internal/platform/problem` error envelope, `tests/integration/testdb` factory framework (live-PG integration tests, `//go:build integration`).

**Files touched (map):**
- Create: `db/migrations/0247_notifications_table.sql`
- Modify: `tests/integration/testdb/factory.go` (add `Notification` struct + `NewNotification` builder)
- Create: `internal/modules/notifications/domain/types.go`
- Create: `internal/modules/notifications/infrastructure/notifications_repository.go`
- Create: `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go`
- Modify: `scripts/api-lint/tripwire-allowlist.txt` (allowlist `MarkRead`)
- Modify: `tools/cilint/internal/analyzers/hgcrossmodule.go` (register table ownership)
- Create: `internal/modules/notifications/delivery/http/handler.go`
- Create: `internal/modules/notifications/delivery/http/routes.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go` (wire repo→handler→routes)
- Modify: `apps/api/cmd/metaldocs-api/permissions.go` (3 tier-1 route rules)
- Modify: `apps/api/cmd/metaldocs-api/permissions_test.go` (route resolver cases)

**Ordering rationale:** migration + factory builder are prerequisites for any integration test (the tests seed rows into the real table). Then domain types, then the repo built method-by-method TDD (failing live-PG test first). The tripwire allowlist lands in the same task as `MarkRead` so the build never goes red. Ownership manifest, handler, wiring, and permissions follow once the repo is green.

---

### Task 1: Migration — `metaldocs.notifications` table

**Files:**
- Create: `db/migrations/0247_notifications_table.sql`

- [ ] **Step 1: Write the migration**

Create `db/migrations/0247_notifications_table.sql`:

```sql
-- 0247_notifications_table.sql
-- Notifications module (M3/F3.2): per-recipient inbox table owned by the
-- notifications module. Read surface = list / unread-count / mark-read.
-- ADR-0043. source_event_id + the partial unique index are the F3.3 projector
-- idempotency key, shipped now so F3.3 needs no ALTER (operator decision
-- 2026-06-22). RLS uses the 0237 NULL-permissive tenant_isolation pattern
-- verbatim. Forward-only, idempotent.

BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.notifications (
    id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         uuid        NOT NULL,
    recipient_user_id text        NOT NULL,
    event_type        text        NOT NULL,
    resource_type     text        NOT NULL,
    resource_id       text        NOT NULL,
    title             text        NOT NULL,
    message           text        NOT NULL,
    status            text        NOT NULL DEFAULT 'PENDING'
                                  CHECK (status IN ('PENDING', 'SENT', 'READ')),
    created_at        timestamptz NOT NULL DEFAULT now(),
    read_at           timestamptz,
    source_event_id   uuid
);

-- Keyset list index: (tenant_id, recipient_user_id) equality + (created_at DESC, id DESC) order.
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created
    ON metaldocs.notifications (tenant_id, recipient_user_id, created_at DESC, id DESC);

-- Projector idempotency (F3.3): at most one row per (recipient, source event).
-- Partial so non-projector rows (source_event_id IS NULL, e.g. test fixtures) are unconstrained.
CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_recipient_event
    ON metaldocs.notifications (recipient_user_id, source_event_id)
    WHERE source_event_id IS NOT NULL;

-- RLS: 0237 pattern verbatim (ENABLE + FORCE + one NULL-permissive tenant_isolation policy).
ALTER TABLE metaldocs.notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.notifications FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.notifications;
CREATE POLICY tenant_isolation ON metaldocs.notifications
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

COMMIT;
```

- [ ] **Step 2: Apply the migration and verify schema**

Run: `.\scripts\check-system-runnable.ps1`
Expected: migrates clean, system runnable. Then confirm the table:

Run (psql against the dev DB): `\d metaldocs.notifications`
Expected: 12 columns, `status` CHECK constraint, both indexes (`idx_notifications_recipient_created`, `uq_notifications_recipient_event`), RLS enabled + forced, `tenant_isolation` policy present.

- [ ] **Step 3: Verify migration sequence is gapless**

Run: `go test ./tools/cilint/... -run TestMigrationGapless` (or the invariants guard equivalent)
Expected: PASS — 0247 follows 0246 with no gap.

- [ ] **Step 4: Commit**

```bash
git add db/migrations/0247_notifications_table.sql
git commit -m "feat(M3/F3.2): notifications table migration (0247) + RLS + projector idempotency index"
```

---

### Task 2: testdb factory — `NewNotification` builder

The integration tests seed notification rows directly (no projector exists yet). The `notifications` table carries no tripwire trigger, so the builder inserts with the plain `exec` helper (like `NewTenant`/`NewUser`), not `seedWithCaps`.

**Files:**
- Modify: `tests/integration/testdb/factory.go`

- [ ] **Step 1: Add the `Notification` fixture struct**

In `tests/integration/testdb/factory.go`, after the `ApprovalRoute` struct (near line 80), add:

```go
type Notification struct {
	ID              string
	TenantID        string
	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string
	Status          string
}
```

- [ ] **Step 2: Add `Recipient`/`EventType`/`ResourceType`/`ResourceID` option fields**

In the `Spec` struct add the fields (near the other string fields, ~line 110):

```go
	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string
```

And the option setters (with the other `WithX` funcs, ~line 137):

```go
func WithRecipient(userID string) Opt   { return func(s *Spec) { s.RecipientUserID = userID } }
func WithEventType(t string) Opt         { return func(s *Spec) { s.EventType = t } }
func WithResourceType(t string) Opt      { return func(s *Spec) { s.ResourceType = t } }
func WithResourceID(id string) Opt       { return func(s *Spec) { s.ResourceID = id } }
```

- [ ] **Step 3: Add the `NewNotification` builder**

After `NewApprovalInstance` (end of the builders section), add:

```go
// NewNotification seeds a metaldocs.notifications row (no tripwire). Auto-wires a
// tenant and a recipient user when not supplied. status defaults to 'PENDING'.
// source_event_id is left NULL (the partial unique index does not constrain NULLs),
// so multiple fixture rows for one recipient never collide.
func NewNotification(t *testing.T, db *sql.DB, opts ...Opt) Notification {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	recipient := s.RecipientUserID
	if recipient == "" {
		recipient = NewUser(t, db, WithTenant(tenantID)).ID
	}
	id := uuid.NewString()
	eventType := s.EventType
	if eventType == "" {
		eventType = "document_published"
	}
	resourceType := s.ResourceType
	if resourceType == "" {
		resourceType = "document"
	}
	resourceID := s.ResourceID
	if resourceID == "" {
		resourceID = uuid.NewString()
	}
	status := s.Status
	if status == "" {
		status = "PENDING"
	}

	exec(t, db,
		`INSERT INTO metaldocs.notifications
		   (id, tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, status)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)`,
		id, tenantID, recipient, eventType, resourceType, resourceID,
		"Novo documento controlado para leitura", "Um documento foi publicado.", status,
	)

	return Notification{
		ID: id, TenantID: tenantID, RecipientUserID: recipient,
		EventType: eventType, ResourceType: resourceType, ResourceID: resourceID, Status: status,
	}
}
```

- [ ] **Step 4: Verify the factory still builds under the integration tag**

Run: `go build -tags integration ./tests/integration/testdb/...`
Expected: exit 0.

- [ ] **Step 5: Commit**

```bash
git add tests/integration/testdb/factory.go
git commit -m "test(M3/F3.2): testdb NewNotification builder + Notification fixture"
```

---

### Task 3: Domain DTOs

**Files:**
- Create: `internal/modules/notifications/domain/types.go`

- [ ] **Step 1: Write the domain types**

Create `internal/modules/notifications/domain/types.go`:

```go
// Package notificationsdomain defines the shared data-transfer types for the
// notifications module. Placing them here (rather than in infra) breaks the
// cross-layer import: the delivery layer depends on domain types without
// importing the infrastructure package — the distribution module's precedent.
package notificationsdomain

import "time"

// NotificationRow is the flat representation of one per-recipient notification,
// the stored shape behind the generated api.Notification.
type NotificationRow struct {
	ID              string
	TenantID        string
	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string
	Title           string
	Message         string
	Status          string // "PENDING" | "SENT" | "READ"
	CreatedAt       time.Time
	ReadAt          *time.Time
}

// NotificationsPage is the result of a paginated List call.
type NotificationsPage struct {
	Items      []NotificationRow
	HasMore    bool
	NextCursor string
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/modules/notifications/domain/...`
Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/modules/notifications/domain/types.go
git commit -m "feat(M3/F3.2): notifications domain DTOs"
```

---

### Task 4: Repository — `List` (failing test first)

**Files:**
- Create: `internal/modules/notifications/infrastructure/notifications_repository.go`
- Create: `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go`

- [ ] **Step 1: Write the failing integration test for List self-scope + status filter + cursor**

Create `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go`:

```go
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
}
```

- [ ] **Step 2: Run the test — verify it fails to compile (repo does not exist)**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run TestNotifications`
Expected: FAIL — `undefined: notificationsinfra.NewNotificationsRepository`.

- [ ] **Step 3: Write the repository with `List` implemented**

Create `internal/modules/notifications/infrastructure/notifications_repository.go`:

```go
// Package notificationsinfra provides the PostgreSQL-backed repository for the
// notifications module. It reads and writes ONLY metaldocs.notifications (its
// own table — no cross-module reads, ADR-0039 compliant). All reads are
// non-transactional, pool-backed, with an explicit tenant_id + recipient_user_id
// predicate from the caller's auth context (self-scope enforced in SQL, not only
// by the CapNotificationRead route guard). No tx, no advisory locks (H-PRE-1).
package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	notificationsdomain "metaldocs/internal/modules/notifications/domain"
	"metaldocs/internal/platform/pagination"
)

// NotificationsRepository reads/writes the notifications table.
type NotificationsRepository struct {
	db *sql.DB
}

// NewNotificationsRepository builds a ready repository. db is required.
func NewNotificationsRepository(db *sql.DB) *NotificationsRepository {
	if db == nil {
		panic("notificationsinfra: db is required")
	}
	return &NotificationsRepository{db: db}
}

// List returns a self-scoped page of the caller's notifications, newest first.
// statusFilter == "" means no filter; otherwise it must be PENDING/SENT/READ.
// Keyset cursor is on (created_at DESC, id DESC) — a stable total order. An empty
// cursor starts the first page; a malformed cursor returns pagination.ErrInvalidCursor.
func (r *NotificationsRepository) List(ctx context.Context, tenantID, recipientUserID, statusFilter, cursor string, limit int) (notificationsdomain.NotificationsPage, error) {
	limit = pagination.ClampLimit(limit)

	var (
		cursorTS string // decoded cursor: created_at RFC3339Nano
		cursorID string // decoded cursor: id tiebreaker
	)
	if cursor != "" {
		var err error
		cursorTS, cursorID, err = pagination.DecodeCursor(cursor)
		if err != nil {
			return notificationsdomain.NotificationsPage{}, pagination.ErrInvalidCursor
		}
	}

	// statusArg is a *string so the predicate ($N::text IS NULL OR status = $N)
	// applies the filter only when set.
	var statusArg *string
	if statusFilter != "" {
		statusArg = &statusFilter
	}

	var (
		rows *sql.Rows
		err  error
	)
	if cursorID == "" {
		// First page: no keyset filter.
		const q = `
			SELECT id, recipient_user_id, event_type, resource_type, resource_id,
			       title, message, status, created_at, read_at
			  FROM metaldocs.notifications
			 WHERE tenant_id = $1::uuid
			   AND recipient_user_id = $2
			   AND ($3::text IS NULL OR status = $3)
			 ORDER BY created_at DESC, id DESC
			 LIMIT $4
		`
		rows, err = r.db.QueryContext(ctx, q, tenantID, recipientUserID, statusArg, limit+1)
	} else {
		// Keyset: rows strictly older than the cursor tuple, NULL-free total order.
		const q = `
			SELECT id, recipient_user_id, event_type, resource_type, resource_id,
			       title, message, status, created_at, read_at
			  FROM metaldocs.notifications
			 WHERE tenant_id = $1::uuid
			   AND recipient_user_id = $2
			   AND ($3::text IS NULL OR status = $3)
			   AND (created_at < $4::timestamptz
			        OR (created_at = $4::timestamptz AND id < $5::uuid))
			 ORDER BY created_at DESC, id DESC
			 LIMIT $6
		`
		rows, err = r.db.QueryContext(ctx, q, tenantID, recipientUserID, statusArg, cursorTS, cursorID, limit+1)
	}
	if err != nil {
		return notificationsdomain.NotificationsPage{}, fmt.Errorf("notifications.List query: %w", err)
	}
	defer rows.Close()

	var items []notificationsdomain.NotificationRow
	for rows.Next() {
		var (
			row    notificationsdomain.NotificationRow
			readAt sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.RecipientUserID, &row.EventType, &row.ResourceType, &row.ResourceID,
			&row.Title, &row.Message, &row.Status, &row.CreatedAt, &readAt,
		); err != nil {
			return notificationsdomain.NotificationsPage{}, fmt.Errorf("notifications.List scan: %w", err)
		}
		row.TenantID = tenantID
		if readAt.Valid {
			v := readAt.Time
			row.ReadAt = &v
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return notificationsdomain.NotificationsPage{}, fmt.Errorf("notifications.List rows: %w", err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = pagination.EncodeCursor(last.CreatedAt.UTC().Format(time.RFC3339Nano), last.ID)
	}

	return notificationsdomain.NotificationsPage{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}
```

- [ ] **Step 4: Run the List tests — verify they pass**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run 'TestNotifications/self_scope_isolation|TestNotifications/status_filter|TestNotifications/cursor_stability' -v`
Expected: PASS (3 subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/modules/notifications/infrastructure/
git commit -m "feat(M3/F3.2): notifications repository List (self-scoped keyset, status filter)"
```

---

### Task 5: Repository — `UnreadCount` (failing test first)

**Files:**
- Modify: `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go`
- Modify: `internal/modules/notifications/infrastructure/notifications_repository.go`

- [ ] **Step 1: Add the failing UnreadCount test**

Add this subtest inside `TestNotifications` in the integration test file:

```go
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
```

- [ ] **Step 2: Run it — verify failure (method undefined)**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run TestNotifications/unread_count_accuracy`
Expected: FAIL — `repo.UnreadCount undefined`.

- [ ] **Step 3: Implement UnreadCount**

Append to `notifications_repository.go`:

```go
// UnreadCount returns the count of the caller's unread (PENDING or SENT)
// notifications. Self-scoped in SQL. Non-transactional pool read.
func (r *NotificationsRepository) UnreadCount(ctx context.Context, tenantID, recipientUserID string) (int, error) {
	const q = `
		SELECT COUNT(*)
		  FROM metaldocs.notifications
		 WHERE tenant_id = $1::uuid
		   AND recipient_user_id = $2
		   AND status IN ('PENDING', 'SENT')
	`
	var n int
	if err := r.db.QueryRowContext(ctx, q, tenantID, recipientUserID).Scan(&n); err != nil {
		return 0, fmt.Errorf("notifications.UnreadCount: %w", err)
	}
	return n, nil
}
```

- [ ] **Step 4: Run it — verify pass**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run TestNotifications/unread_count_accuracy -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/notifications/infrastructure/
git commit -m "feat(M3/F3.2): notifications repository UnreadCount (self-scoped PENDING+SENT)"
```

---

### Task 6: Repository — `MarkRead` (failing test first) + tripwire allow-list

`MarkRead` runs mutating SQL in a `*repository.go` file. The `tripwire-pairing` api-lint rule flags any such function without an `authz.Require` in its body — so the allow-list entry MUST land in this same task, or `api-lint -strict` goes red. Authz is enforced one layer up (tier-1 `CapNotificationRead`) plus the self-scope predicate, exactly like the 13 existing allow-listed repo writes.

**Files:**
- Modify: `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go`
- Modify: `internal/modules/notifications/infrastructure/notifications_repository.go`
- Modify: `scripts/api-lint/tripwire-allowlist.txt`

- [ ] **Step 1: Add the failing MarkRead tests (flip + idempotent + wrong-owner no-op)**

Add these subtests inside `TestNotifications`:

```go
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
```

- [ ] **Step 2: Run them — verify failure (method undefined)**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run 'TestNotifications/mark_read'`
Expected: FAIL — `repo.MarkRead undefined`.

- [ ] **Step 3: Implement MarkRead**

Append to `notifications_repository.go`:

```go
// MarkRead marks one of the caller's notifications READ. Self-scoped and
// idempotent: the predicate requires the row to belong to recipientUserID in
// tenantID; 0 rows affected (wrong owner, missing id, or already READ) is a
// silent no-op (no error, no existence leak). status != 'READ' keeps read_at
// stable on re-runs.
//
// Authz: tier-2 is enforced one layer up — the tier-1 CapNotificationRead route
// guard plus this self-scope predicate. The repository method holds no
// authz.Require (single-file tripwire-pairing scan allow-lists it accordingly).
func (r *NotificationsRepository) MarkRead(ctx context.Context, tenantID, notificationID, recipientUserID string) error {
	const q = `
		UPDATE metaldocs.notifications
		   SET status = 'READ', read_at = now()
		 WHERE id = $1::uuid
		   AND tenant_id = $2::uuid
		   AND recipient_user_id = $3
		   AND status != 'READ'
	`
	if _, err := r.db.ExecContext(ctx, q, notificationID, tenantID, recipientUserID); err != nil {
		return fmt.Errorf("notifications.MarkRead: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Add the tripwire allow-list entry**

In `scripts/api-lint/tripwire-allowlist.txt`, append after the last entry (after the documents block, before/after the Z-10 comment is fine — keep it with the other entries):

```
# M3/F3.2: notifications MarkRead is a self-scoped idempotent UPDATE. Tier-2 authz
# is enforced one layer up — the tier-1 CapNotificationRead route guard
# (permissions.go) plus the recipient_user_id self-scope predicate in the SQL.
# The repository method intentionally holds no authz.Require (dumb repo write,
# same pattern as the documents/approval/CD entries above).
internal/modules/notifications/infrastructure/notifications_repository.go|MarkRead
```

- [ ] **Step 5: Run the MarkRead tests — verify pass**

Run: `go test -tags integration ./internal/modules/notifications/infrastructure/... -run 'TestNotifications/mark_read' -v`
Expected: PASS (both subtests).

- [ ] **Step 6: Verify api-lint strict stays green (tripwire pairing satisfied)**

Run: `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .`
Expected: `0 violation(s)` — no new tripwire-pairing finding, no stale allow-list entry.

- [ ] **Step 7: Commit**

```bash
git add internal/modules/notifications/infrastructure/ scripts/api-lint/tripwire-allowlist.txt
git commit -m "feat(M3/F3.2): notifications repository MarkRead (self-scoped idempotent) + tripwire allow-list"
```

---

### Task 7: Register table ownership in the hgcrossmodule manifest

**Files:**
- Modify: `tools/cilint/internal/analyzers/hgcrossmodule.go`

- [ ] **Step 1: Add the ownership entry**

In `tools/cilint/internal/analyzers/hgcrossmodule.go`, in the `hgOwnerByTable` map, after the `// jobs` block (after `"job_leases": "jobs",`), add:

```go
	// notifications
	"notifications": "notifications",
```

- [ ] **Step 2: Verify cilint builds and is green**

Run: `go test ./tools/cilint/...`
Expected: PASS — the new own-table entry introduces no cross-module finding (notifications reads only its own table; `owner == reader` is skipped).

- [ ] **Step 3: Commit**

```bash
git add tools/cilint/internal/analyzers/hgcrossmodule.go
git commit -m "chore(M3/F3.2): register notifications table ownership in hgcrossmodule manifest"
```

---

### Task 8: Delivery — handler + routes

**Files:**
- Create: `internal/modules/notifications/delivery/http/handler.go`
- Create: `internal/modules/notifications/delivery/http/routes.go`

- [ ] **Step 1: Write the handler**

Create `internal/modules/notifications/delivery/http/handler.go`:

```go
// Package notificationshttp implements the HTTP delivery layer for the
// notifications module. It satisfies the generated notificationsapi.StrictServerInterface
// using the repository for data access. Auth (401/403) is produced upstream by
// tier-1 middleware (CapNotificationRead); this handler only sees authenticated,
// authorized requests. Self-scope (a caller sees only their own rows) is enforced
// by passing the authenticated user id into every repository call, where the SQL
// predicate filters on it.
package notificationshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	notificationsapi "metaldocs/internal/modules/notifications/api"
	notificationsdomain "metaldocs/internal/modules/notifications/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// Repository is the minimal surface the handler needs from the infrastructure layer.
type Repository interface {
	List(ctx context.Context, tenantID, recipientUserID, statusFilter, cursor string, limit int) (notificationsdomain.NotificationsPage, error)
	UnreadCount(ctx context.Context, tenantID, recipientUserID string) (int, error)
	MarkRead(ctx context.Context, tenantID, notificationID, recipientUserID string) error
}

// Handler implements notificationsapi.StrictServerInterface.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler. repo must not be nil.
func NewHandler(repo Repository) *Handler {
	if repo == nil {
		panic("notificationshttp: repo is required")
	}
	return &Handler{repo: repo}
}

// ListNotifications handles GET /notifications. Self-scoped, newest first.
func (h *Handler) ListNotifications(
	ctx context.Context,
	req notificationsapi.ListNotificationsRequestObject,
) (notificationsapi.ListNotificationsResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.ListNotifications500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	statusFilter := ""
	if req.Params.Status != nil {
		statusFilter = string(*req.Params.Status)
	}
	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	limit := pagination.DefaultLimit
	if req.Params.Limit != nil && *req.Params.Limit > 0 {
		limit = *req.Params.Limit
	}
	limit = pagination.ClampLimit(limit)

	page, err := h.repo.List(ctx, tenantID, userID, statusFilter, cursor, limit)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return notificationsapi.ListNotifications400ApplicationProblemPlusJSONResponse{
				BadRequestApplicationProblemPlusJSONResponse: notificationsapi.BadRequestApplicationProblemPlusJSONResponse(
					toProblem(problem.New(http.StatusBadRequest, problem.CodeInvalidCursor, "Invalid cursor")),
				),
			}, nil
		}
		slog.Error("notifications.ListNotifications", "error", err)
		return notificationsapi.ListNotifications500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	items := make([]notificationsapi.Notification, len(page.Items))
	for i, n := range page.Items {
		items[i] = toAPINotification(n)
	}

	var nextCursor *string
	if page.HasMore && page.NextCursor != "" {
		nc := page.NextCursor
		nextCursor = &nc
	}

	return notificationsapi.ListNotifications200JSONResponse{
		Items: items,
		Page: notificationsapi.CursorPage{
			HasMore:    page.HasMore,
			NextCursor: nextCursor,
		},
	}, nil
}

// GetNotificationsUnreadCount handles GET /notifications/unread-count.
func (h *Handler) GetNotificationsUnreadCount(
	ctx context.Context,
	req notificationsapi.GetNotificationsUnreadCountRequestObject,
) (notificationsapi.GetNotificationsUnreadCountResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.GetNotificationsUnreadCount500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	count, err := h.repo.UnreadCount(ctx, tenantID, userID)
	if err != nil {
		slog.Error("notifications.GetNotificationsUnreadCount", "error", err)
		return notificationsapi.GetNotificationsUnreadCount500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return notificationsapi.GetNotificationsUnreadCount200JSONResponse{
		Count: count,
	}, nil
}

// MarkNotificationRead handles POST /notifications/{id}/read. Idempotent, self-scoped.
func (h *Handler) MarkNotificationRead(
	ctx context.Context,
	req notificationsapi.MarkNotificationReadRequestObject,
) (notificationsapi.MarkNotificationReadResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.MarkNotificationRead404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: notificationsapi.NotFoundApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusNotFound, problem.CodeNotFound, "Notification not found")),
			),
		}, nil
	}

	if err := h.repo.MarkRead(ctx, tenantID, req.Id.String(), userID); err != nil {
		slog.Error("notifications.MarkNotificationRead", "error", err)
		return notificationsapi.MarkNotificationRead500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return notificationsapi.MarkNotificationRead204Response{}, nil
}

// toAPINotification maps a stored row to the generated wire shape.
func toAPINotification(n notificationsdomain.NotificationRow) notificationsapi.Notification {
	id, _ := uuidParse(n.ID)
	return notificationsapi.Notification{
		Id:              id,
		RecipientUserId: n.RecipientUserID,
		EventType:       n.EventType,
		ResourceType:    n.ResourceType,
		ResourceId:      n.ResourceID,
		Title:           n.Title,
		Message:         n.Message,
		Status:          notificationsapi.NotificationStatus(n.Status),
		CreatedAt:       n.CreatedAt,
		ReadAt:          n.ReadAt,
	}
}

// extractTenantAndUser reads the authenticated tenant + user from context. Returns
// ok=false when auth context is missing. 401/403 are handled upstream by tier-1
// middleware before the handler is invoked.
func extractTenantAndUser(ctx context.Context) (tenantID, userID string, ok bool) {
	userID, ok = authn.UserIDFromContext(ctx)
	if !ok {
		return "", "", false
	}
	tid, err := tenant.FromContext(ctx)
	if err != nil {
		return "", "", false
	}
	return tid, userID, true
}

// toProblem maps a *problem.Problem to the generated notificationsapi.Problem shape.
func toProblem(p *problem.Problem) notificationsapi.Problem {
	var detail *string
	if p.Detail != "" {
		d := p.Detail
		detail = &d
	}
	var instance *string
	if p.Instance != "" {
		inst := p.Instance
		instance = &inst
	}
	return notificationsapi.Problem{
		Status:   p.Status,
		Title:    p.Title,
		Code:     string(p.Code),
		Detail:   detail,
		Instance: instance,
	}
}
```

> Note on `uuidParse`: the generated `Notification.Id` is `openapi_types.UUID` (a `github.com/google/uuid` UUID alias). Add the import + a tiny helper rather than importing the package twice with different names. See Step 2.

- [ ] **Step 2: Add the uuid import + helper used by `toAPINotification`**

In `handler.go`, add to the import block:

```go
	openapi_types "github.com/oapi-codegen/runtime/types"
```

and replace the `uuidParse` call site by inlining the parse — change `toAPINotification` to:

```go
func toAPINotification(n notificationsdomain.NotificationRow) notificationsapi.Notification {
	var id openapi_types.UUID
	_ = id.UnmarshalText([]byte(n.ID))
	return notificationsapi.Notification{
		Id:              id,
		RecipientUserId: n.RecipientUserID,
		EventType:       n.EventType,
		ResourceType:    n.ResourceType,
		ResourceId:      n.ResourceID,
		Title:           n.Title,
		Message:         n.Message,
		Status:          notificationsapi.NotificationStatus(n.Status),
		CreatedAt:       n.CreatedAt,
		ReadAt:          n.ReadAt,
	}
}
```

(Delete the earlier `id, _ := uuidParse(n.ID)` version and the note — there is no `uuidParse` function. `openapi_types.UUID` is `uuid.UUID`, whose `UnmarshalText` parses the canonical string. The DB id is always a valid uuid, so the error is impossible; ignoring it is safe and mirrors how the generated code round-trips ids.)

- [ ] **Step 3: Write the routes file**

Create `internal/modules/notifications/delivery/http/routes.go`:

```go
package notificationshttp

import (
	"net/http"

	notificationsapi "metaldocs/internal/modules/notifications/api"
)

// RegisterRoutes wires the notifications endpoints onto mux via the generated
// strict handler. All endpoints share the /api/v1 base URL prefix.
func RegisterRoutes(h *Handler, mux *http.ServeMux) {
	strict := notificationsapi.NewStrictHandler(h, nil)
	notificationsapi.HandlerWithOptions(strict, notificationsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
	})
}
```

- [ ] **Step 4: Verify the handler satisfies the strict interface + compiles**

Run: `go build ./internal/modules/notifications/...`
Expected: exit 0 (the handler structurally satisfies `notificationsapi.StrictServerInterface`).

- [ ] **Step 5: Commit**

```bash
git add internal/modules/notifications/delivery/
git commit -m "feat(M3/F3.2): notifications delivery handler + routes (strict interface, self-scope)"
```

---

### Task 9: Wire the module in main.go

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`

- [ ] **Step 1: Add the imports**

In `apps/api/cmd/metaldocs-api/main.go`, in the import block alongside `distributionhttp` / `distributioninfra` (~line 46), add:

```go
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
```

- [ ] **Step 2: Construct + register after the distribution block**

In `main.go`, immediately after the distribution wiring (after line 542 `distributionhttp.RegisterRoutes(distributionHandler, mux)`), add:

```go
	// M3/F3.2: notifications module — read surface (list / unread-count / mark-read).
	// Self-scoped by CapNotificationRead (tier-1) + recipient_user_id SQL predicate.
	notificationsRepo := notificationsinfra.NewNotificationsRepository(deps.SQLDB)
	notificationsHandler := notificationshttp.NewHandler(notificationsRepo)
	notificationshttp.RegisterRoutes(notificationsHandler, mux)
```

- [ ] **Step 3: Verify it builds**

Run: `go build ./...`
Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
git add apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(M3/F3.2): wire notifications module into the API"
```

---

### Task 10: Permissions — tier-1 route guards + resolver tests

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/permissions.go`
- Modify: `apps/api/cmd/metaldocs-api/permissions_test.go`

- [ ] **Step 1: Add the three route rules**

In `apps/api/cmd/metaldocs-api/permissions.go`, in the route-rule table, add a notifications block (place it near the audit rules, ~line 238, or any spot in the table — the resolver matches most-specific-first within a path; these three are mutually exclusive so ordering among them is by specificity):

```go
	// Notifications — self-scoped read surface (M3/F3.2). CapNotificationRead at tier-1
	// (mirrors CapAuditRead). Most specific first: unread-count (exact) and /read (suffix)
	// precede the bare collection GET.
	{method: http.MethodGet, pathExact: "/api/v1/notifications/unread-count", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodPost, pathPrefix: "/api/v1/notifications/", pathSuffix: "/read", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
	{method: http.MethodGet, pathExact: "/api/v1/notifications", capability: iamdomain.CapNotificationRead, visibility: iamdelivery.VisibilityPermissionGuarded},
```

- [ ] **Step 2: Add resolver test cases**

In `apps/api/cmd/metaldocs-api/permissions_test.go`, find the table of `TestPermissionResolver` cases and add three entries (match the struct shape used by the existing cases — `{method, path, wantCap, wantVisibility}`):

```go
	{http.MethodGet, "/api/v1/notifications", iamdomain.CapNotificationRead, iamdelivery.VisibilityPermissionGuarded},
	{http.MethodGet, "/api/v1/notifications/unread-count", iamdomain.CapNotificationRead, iamdelivery.VisibilityPermissionGuarded},
	{http.MethodPost, "/api/v1/notifications/3f2504e0-4f89-11d3-9a0c-0305e82c3301/read", iamdomain.CapNotificationRead, iamdelivery.VisibilityPermissionGuarded},
```

> If the local case struct uses named fields or a different signature, mirror the existing rows exactly — read two neighboring cases first and copy their shape. The three assertions are: bare GET, unread-count GET, and a `{id}/read` POST all resolve to `CapNotificationRead` + `VisibilityPermissionGuarded`.

- [ ] **Step 3: Run the permission tests — verify pass**

Run: `go test ./apps/api/cmd/metaldocs-api/ -run 'TestPermissionResolver|TestRouteCoverage|TestEveryRouteCapInRegistry|TestEveryCapSeededOrDeferred' -v`
Expected: PASS — the new routes resolve correctly, no fallthrough in `TestRouteCoverage`, `CapNotificationRead` is a registered const (F3.1) and is in the deferred set (F3.1), so `TestEveryCapSeededOrDeferred` stays green.

- [ ] **Step 4: Commit**

```bash
git add apps/api/cmd/metaldocs-api/permissions.go apps/api/cmd/metaldocs-api/permissions_test.go
git commit -m "feat(M3/F3.2): tier-1 CapNotificationRead guards for notifications routes + resolver tests"
```

---

### Task 11: Full gate — green across the board

**Files:** none (verification only).

- [ ] **Step 1: Build + vet + unit tests**

Run: `go build ./...` then `go vet ./...` then `go test ./...`
Expected: all exit 0 / PASS.

- [ ] **Step 2: Integration suite for notifications**

Run: `go test -tags integration ./internal/modules/notifications/... -v`
Expected: PASS — all `TestNotifications` subtests (self_scope_isolation, status_filter, cursor_stability, unread_count_accuracy, mark_read_flips_and_idempotent, mark_read_wrong_owner_noop).

- [ ] **Step 3: api-lint strict = 0**

Run: `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .`
Expected: `0 violation(s)`.

- [ ] **Step 4: cilint guards (incl. hgcrossmodule)**

Run: `go test ./tools/cilint/...`
Expected: PASS — no cross-module finding, ownership manifest consistent.

- [ ] **Step 5: Publish path untouched**

Run: `git diff --quiet -- internal/modules/documents/approval/application/publish_service.go && echo CLEAN`
Expected: `CLEAN` (empty diff).

- [ ] **Step 6: System runnable**

Run: `.\scripts\check-system-runnable.ps1`
Expected: green — migrations applied, API boots, notifications routes mounted.

- [ ] **Step 7: Final commit (if any verification-driven fixups were needed)**

```bash
git add -A
git commit -m "test(M3/F3.2): full gate green — notifications read surface verified"
```

---

## Execution notes

Decisions made during `superpowers:subagent-driven-development` go here (model choices,
deviations from plan with rationale, questions answered). The durable record is `evidence.md`.
