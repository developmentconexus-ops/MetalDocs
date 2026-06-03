package presence

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository abstracts the storage operations the presence hub + bump
// middleware need. Production binding is *PostgresRepository.
type Repository interface {
	// BumpLastSeen sets iam_users.last_seen_at = at for the given user.
	// Implementations MUST be safe to call concurrently; the underlying
	// table is unique on (tenant_id, user_id). userID is the iam_users
	// primary key; tenant_id is fixed by that row.
	BumpLastSeen(ctx context.Context, userID string, at time.Time) error

	// Snapshot returns every active iam_users row for tenantID whose
	// last_seen_at >= now - OfflineAfter, classified into online / idle.
	// Results are sorted by last_seen_at DESC.
	Snapshot(ctx context.Context, tenantID string, now time.Time) ([]Item, error)

	// TenantsWithRecentPresence returns the distinct tenant_id values
	// observed in iam_users with last_seen_at >= since. The hub uses
	// this on cold start to discover which rooms it should pre-tick;
	// in steady-state, rooms are created on first connection.
	TenantsWithRecentPresence(ctx context.Context, since time.Time) ([]string, error)
}

// PostgresRepository is the production Repository wired against
// metaldocs.iam_users. The struct holds the *sql.DB injected at
// construction time; no shared mutable state.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository constructs a presence repository against the
// supplied database handle. Passing nil panics — presence has no
// in-memory fallback.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	if db == nil {
		panic("presence.NewPostgresRepository: db is nil")
	}
	return &PostgresRepository{db: db}
}

// BumpLastSeen updates iam_users.last_seen_at unconditionally. The
// middleware debounces (BumpDebounce) before calling here so the row
// is not hit on every authenticated request.
func (r *PostgresRepository) BumpLastSeen(ctx context.Context, userID string, at time.Time) error {
	const q = `UPDATE metaldocs.iam_users SET last_seen_at = $2 WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, q, userID, at.UTC()); err != nil {
		return fmt.Errorf("presence: bump last_seen_at: %w", err)
	}
	return nil
}

// Snapshot reads the current presence for one tenant. Filters out
// deactivated and inactive rows, and rows older than OfflineAfter.
func (r *PostgresRepository) Snapshot(ctx context.Context, tenantID string, now time.Time) ([]Item, error) {
	const q = `
SELECT user_id, display_name, last_seen_at
FROM metaldocs.iam_users
WHERE tenant_id = $1::uuid
  AND is_active = true
  AND deactivated_at IS NULL
  AND last_seen_at >= $2
ORDER BY last_seen_at DESC
`
	cutoff := now.Add(-OfflineAfter).UTC()
	rows, err := r.db.QueryContext(ctx, q, tenantID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("presence: snapshot query: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.UserID, &item.DisplayName, &item.LastSeenAt); err != nil {
			return nil, fmt.Errorf("presence: scan snapshot row: %w", err)
		}
		item.LastSeenAt = item.LastSeenAt.UTC()
		item.Status = ClassifyStatus(item.LastSeenAt, now)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("presence: iterate snapshot rows: %w", err)
	}
	return items, nil
}

// TenantsWithRecentPresence returns distinct tenant_ids with at least
// one iam_users row newer than since.
func (r *PostgresRepository) TenantsWithRecentPresence(ctx context.Context, since time.Time) ([]string, error) {
	const q = `
SELECT DISTINCT tenant_id::text
FROM metaldocs.iam_users
WHERE last_seen_at >= $1
`
	rows, err := r.db.QueryContext(ctx, q, since.UTC())
	if err != nil {
		return nil, fmt.Errorf("presence: tenants query: %w", err)
	}
	defer rows.Close()

	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("presence: scan tenant row: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("presence: iterate tenant rows: %w", err)
	}
	return out, nil
}
