package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Ensure UserDisplayNameRepository satisfies iamdomain.UserDisplayNameReader.
var _ iamdomain.UserDisplayNameReader = (*UserDisplayNameRepository)(nil)

// UserDisplayNameRepository implements iamdomain.UserDisplayNameReader. It reads
// metaldocs.iam_users.display_name on the connection pool (never on a caller's
// transaction), so cross-module consumers never hold an iam_users read inside a
// lock-holding transaction (H-PRE-1, M4/F4.1).
type UserDisplayNameRepository struct {
	db *sql.DB
}

// NewUserDisplayNameRepository constructs a pool-backed UserDisplayNameReader.
func NewUserDisplayNameRepository(db *sql.DB) *UserDisplayNameRepository {
	return &UserDisplayNameRepository{db: db}
}

// DisplayName returns metaldocs.iam_users.display_name for (tenantID, userID).
// Returns ("", nil) when no row is found in that tenant.
func (r *UserDisplayNameRepository) DisplayName(ctx context.Context, tenantID, userID string) (string, error) {
	var displayName sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT display_name
		   FROM metaldocs.iam_users
		  WHERE user_id   = $1
		    AND tenant_id = $2::uuid`,
		userID, tenantID,
	).Scan(&displayName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("iam: DisplayName(%s, %s): %w", tenantID, userID, err)
	}
	return displayName.String, nil
}

// DisplayNames returns userID -> display_name for those users in the tenant whose
// display_name is non-null and non-empty. Absent users and those with a null/empty
// display_name are omitted; the caller applies its own fallback.
func (r *UserDisplayNameRepository) DisplayNames(ctx context.Context, tenantID string, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, display_name
		   FROM metaldocs.iam_users
		  WHERE tenant_id  = $1::uuid
		    AND user_id    = ANY($2)
		    AND display_name IS NOT NULL
		    AND display_name <> ''`,
		tenantID, pq.Array(userIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("iam: DisplayNames(%s): %w", tenantID, err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]string, len(userIDs))
	for rows.Next() {
		var uid, name string
		if err := rows.Scan(&uid, &name); err != nil {
			return nil, fmt.Errorf("iam: DisplayNames scan: %w", err)
		}
		out[uid] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: DisplayNames rows: %w", err)
	}
	return out, nil
}
