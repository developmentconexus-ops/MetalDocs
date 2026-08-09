package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// RouteReadinessReaderPG satisfies the approval-owned published port.
var _ domain.RouteReadinessReader = (*RouteReadinessReaderPG)(nil)

// RouteReadinessReaderPG is the Postgres-backed
// approvaldomain.RouteReadinessReader. It is the single home for cross-module
// reads of approval_routes readiness — the table token lives here, never in a
// consumer module (ADR-0039 D3(b)). The predicate is the same one route
// resolution uses (postgres_approval_repository.go LoadActiveRouteIDBySubject):
// (tenant_id, subject_kind, subject_key, active = true).
//
// The struct holds a pool only for the off-tx bulk read; HasActiveRoute is
// executor-driven and ignores the pool, so the in-tx creation gate runs on the
// caller's own *sql.Tx. Both SELECTs are plain and non-recording (HS-PRE-1 safe).
type RouteReadinessReaderPG struct {
	pool db.DB
}

// NewRouteReadinessReaderPG constructs the route-readiness reader bound to pool
// for its bulk (off-tx) read. pool must be non-nil.
func NewRouteReadinessReaderPG(pool db.DB) *RouteReadinessReaderPG {
	if pool == nil {
		panic("approval/infrastructure: NewRouteReadinessReaderPG requires a non-nil pool")
	}
	return &RouteReadinessReaderPG{pool: pool}
}

// ActiveRouteSubjectKeys returns the set of subject_key values with at least one
// active route for (tenantID, subjectKind). Runs off-tx on the bound pool.
func (r *RouteReadinessReaderPG) ActiveRouteSubjectKeys(ctx context.Context, tenantID, subjectKind string) (map[string]struct{}, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT DISTINCT subject_key
		  FROM approval_routes
		 WHERE tenant_id    = $1
		   AND subject_kind = $2
		   AND active       = true`,
		tenantID, subjectKind,
	)
	if err != nil {
		return nil, fmt.Errorf("approval: list active route subject keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]struct{})
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("approval: scan active route subject key: %w", err)
		}
		out[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("approval: iterate active route subject keys: %w", err)
	}
	return out, nil
}

// HasActiveRoute reports whether an active route exists for
// (tenantID, subjectKind, subjectKey), running on the caller-supplied executor
// so an in-tx gate observes the same snapshot as the write it guards.
func (r *RouteReadinessReaderPG) HasActiveRoute(ctx context.Context, exec db.DB, tenantID, subjectKind, subjectKey string) (bool, error) {
	var one int
	err := exec.QueryRowContext(ctx, `
		SELECT 1
		  FROM approval_routes
		 WHERE tenant_id    = $1
		   AND subject_kind = $2
		   AND subject_key  = $3
		   AND active       = true
		 LIMIT 1`,
		tenantID, subjectKind, subjectKey,
	).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("approval: check active route: %w", err)
	}
	return true, nil
}
