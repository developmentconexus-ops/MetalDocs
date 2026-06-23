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

// MarkAllRead marks every one of the caller's unread (PENDING/SENT) notifications
// READ in a single statement and returns the number of rows transitioned.
// Self-scoped and idempotent: the predicate requires rows to belong to
// recipientUserID in tenantID; a second call affects 0 rows (all already READ)
// and returns 0 — no error, no existence leak. Mirrors MarkRead's predicate at
// collection scope (no id filter).
//
// Authz: same as MarkRead — tier-1 CapNotificationRead route guard plus this
// self-scope predicate; the repository method holds no authz.Require.
func (r *NotificationsRepository) MarkAllRead(ctx context.Context, tenantID, recipientUserID string) (int, error) {
	const q = `
		UPDATE metaldocs.notifications
		   SET status = 'READ', read_at = now()
		 WHERE tenant_id = $1::uuid
		   AND recipient_user_id = $2
		   AND status IN ('PENDING', 'SENT')
	`
	res, err := r.db.ExecContext(ctx, q, tenantID, recipientUserID)
	if err != nil {
		return 0, fmt.Errorf("notifications.MarkAllRead: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("notifications.MarkAllRead rows: %w", err)
	}
	return int(n), nil
}
