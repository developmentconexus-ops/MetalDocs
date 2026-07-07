//go:build integration
// +build integration

// Package approval_test — F2 (milestone 2b, approval-kernel-backend) route
// versioning + supersession validation. Confirms migration 0287 applies
// cleanly, the column-scoped enforce_route_immutable exception allows an
// active/superseded_at-only UPDATE on an in-use route while still blocking
// any definition-column UPDATE, and the supersede sequence (new versioned
// row + old row marked superseded) leaves in-flight instances resolving the
// old (now-superseded) row unchanged.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/tests/integration/testdb"
)

// TestRouteVersioning_NotInUse_InPlaceUpdateSucceeds asserts that a route with
// no referencing approval_instances can still be mutated in place (name +
// version bump on the same row) — the unchanged cheap path (W1).
func TestRouteVersioning_NotInUse_InPlaceUpdateSucceeds(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db)

	var newVersion int
	err := db.QueryRowContext(context.Background(), `
		UPDATE public.approval_routes
		   SET name = 'Renamed Route', version = version + 1
		 WHERE id = $1::uuid
		RETURNING version`,
		route.ID,
	).Scan(&newVersion)
	if err != nil {
		t.Fatalf("in-place update of not-in-use route: %v", err)
	}
	if newVersion != 2 {
		t.Fatalf("version = %d, want 2", newVersion)
	}

	var active bool
	var supersededAt sql.NullTime
	if err := db.QueryRowContext(context.Background(),
		`SELECT active, superseded_at FROM public.approval_routes WHERE id = $1::uuid`,
		route.ID,
	).Scan(&active, &supersededAt); err != nil {
		t.Fatalf("read route after update: %v", err)
	}
	if !active {
		t.Fatalf("route should remain active after in-place update")
	}
	if supersededAt.Valid {
		t.Fatalf("superseded_at should stay NULL after in-place update, got %v", supersededAt.Time)
	}
}

// TestRouteVersioning_InUse_DefinitionUpdateBlocked_ThenSupersede exercises the
// full W1 in-use path: a definition-column UPDATE (name/version) is blocked by
// the column-scoped enforce_route_immutable trigger (P0001), an active/
// superseded_at-only UPDATE on the same row succeeds, a new versioned row is
// created, and the old row's own stage rows are left untouched (still
// resolvable by any in-flight instance pointing at it).
func TestRouteVersioning_InUse_DefinitionUpdateBlocked_ThenSupersede(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db)
	seedRouteStage(t, db, route.ID, "Stage 1")

	// Referencing approval_instances row makes the route "in use".
	testdb.NewApprovalInstance(t, db, testdb.WithRoute(route))

	// 1. Direct SQL UPDATE of a definition column (name) on the in-use row
	// must be rejected with the P0001 tripwire.
	_, err := db.ExecContext(context.Background(),
		`UPDATE public.approval_routes SET name = 'Should Not Apply' WHERE id = $1::uuid`,
		route.ID,
	)
	assertRouteInUseException(t, err)

	// 2. Direct SQL UPDATE of ONLY active/superseded_at on the same in-use
	// row must be permitted by the column-scoped trigger exception.
	newRouteID := uuid.NewString()
	_, err = db.ExecContext(context.Background(), `
		UPDATE public.approval_routes
		   SET active = FALSE, superseded_at = now()
		 WHERE id = $1::uuid`,
		route.ID,
	)
	if err != nil {
		t.Fatalf("active/superseded_at-only update on in-use row must succeed: %v", err)
	}

	// 3. Insert the superseding row (new id, version+1, active=true, same
	// tenant/profile) — the shape route_admin_service.go's supersede branch
	// produces.
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO public.approval_routes (id, tenant_id, profile_code, name, version, active, created_by)
		SELECT $1::uuid, tenant_id, profile_code, 'Renamed Route', version + 1, TRUE, created_by
		  FROM public.approval_routes
		 WHERE id = $2::uuid`,
		newRouteID, route.ID,
	); err != nil {
		t.Fatalf("insert superseding route row: %v", err)
	}
	seedRouteStage(t, db, newRouteID, "Stage 1 v2")

	// Old row: now inactive, superseded_at set, definition columns untouched.
	var oldActive bool
	var oldSupersededAt sql.NullTime
	var oldName string
	if err := db.QueryRowContext(context.Background(),
		`SELECT active, superseded_at, name FROM public.approval_routes WHERE id = $1::uuid`,
		route.ID,
	).Scan(&oldActive, &oldSupersededAt, &oldName); err != nil {
		t.Fatalf("read old route row: %v", err)
	}
	if oldActive {
		t.Fatalf("old route row should be inactive after supersede")
	}
	if !oldSupersededAt.Valid {
		t.Fatalf("old route row superseded_at should be set")
	}
	if oldName != "Route" {
		t.Fatalf("old route name = %q, want unchanged default %q (blocked UPDATE must not apply)", oldName, "Route")
	}

	// New row: distinct id, version bumped, active.
	var newActive bool
	var newVersion int
	if err := db.QueryRowContext(context.Background(),
		`SELECT active, version FROM public.approval_routes WHERE id = $1::uuid`,
		newRouteID,
	).Scan(&newActive, &newVersion); err != nil {
		t.Fatalf("read new route row: %v", err)
	}
	if !newActive {
		t.Fatalf("new route row should be active")
	}
	if newVersion != 2 {
		t.Fatalf("new route row version = %d, want 2", newVersion)
	}
	if newRouteID == route.ID {
		t.Fatalf("new route id must differ from the superseded path id")
	}

	// In-flight instance still resolves stages from the OLD (superseded) row —
	// its stage rows were never touched by the supersede sequence.
	var oldStageName string
	if err := db.QueryRowContext(context.Background(),
		`SELECT name FROM public.approval_route_stages WHERE route_id = $1::uuid LIMIT 1`,
		route.ID,
	).Scan(&oldStageName); err != nil {
		t.Fatalf("read old route's stages: %v", err)
	}
	if oldStageName != "Stage 1" {
		t.Fatalf("old route stage name = %q, want unchanged %q", oldStageName, "Stage 1")
	}
}

// TestRouteVersioning_InUse_VersionColumnUpdateBlocked asserts the trigger's
// column-scoped exception also rejects a bare version-only UPDATE on an
// in-use row (only active/superseded_at are exempted).
func TestRouteVersioning_InUse_VersionColumnUpdateBlocked(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db)
	testdb.NewApprovalInstance(t, db, testdb.WithRoute(route))

	_, err := db.ExecContext(context.Background(),
		`UPDATE public.approval_routes SET version = version + 1 WHERE id = $1::uuid`,
		route.ID,
	)
	assertRouteInUseException(t, err)
}

// TestRouteVersioning_InactiveRoute_DefinitionUpdateBlocked asserts an
// already-inactive (e.g. previously superseded or deactivated) row also
// rejects definition-column updates, per the trigger's "NOT OLD.active"
// branch.
func TestRouteVersioning_InactiveRoute_DefinitionUpdateBlocked(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db)
	if _, err := db.ExecContext(context.Background(),
		`UPDATE public.approval_routes SET active = FALSE, superseded_at = now() WHERE id = $1::uuid`,
		route.ID,
	); err != nil {
		t.Fatalf("seed inactive route: %v", err)
	}

	_, err := db.ExecContext(context.Background(),
		`UPDATE public.approval_routes SET name = 'Should Not Apply' WHERE id = $1::uuid`,
		route.ID,
	)
	assertRouteInUseException(t, err)
}

func seedRouteStage(t *testing.T, database *sql.DB, routeID, name string) {
	t.Helper()
	if _, err := database.ExecContext(context.Background(), `
		INSERT INTO public.approval_route_stages
		  (route_id, stage_order, name, required_role, required_capability,
		   area_code, quorum, on_eligibility_drift)
		VALUES ($1::uuid, 1, $2, 'author', 'document.review', 'area-1',
		        'any_1_of', 'reduce_quorum')`,
		routeID, name,
	); err != nil {
		t.Fatalf("seed approval_route_stages row: %v", err)
	}
}

// assertRouteInUseException fails the test unless err is the P0001
// "ErrRouteInUse:"-prefixed exception raised by enforce_route_immutable().
func assertRouteInUseException(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected ErrRouteInUse P0001 exception, got nil error")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "P0001" {
			t.Fatalf("expected SQLSTATE P0001, got %s: %v", pgErr.Code, err)
		}
		if !strings.HasPrefix(pgErr.Message, "ErrRouteInUse") {
			t.Fatalf("expected message prefix %q, got %q", "ErrRouteInUse", pgErr.Message)
		}
		return
	}
	if !strings.Contains(err.Error(), "ErrRouteInUse") {
		t.Fatalf("expected error mentioning ErrRouteInUse, got: %v", err)
	}
}
