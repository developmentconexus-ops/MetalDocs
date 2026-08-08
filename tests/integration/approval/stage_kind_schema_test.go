//go:build integration
// +build integration

// Package approval_test — F1 (milestone 2b, approval-kernel-backend)
// schema-expand validation. Confirms migration 0286 applies cleanly on top
// of 0285, that pre-existing rows read the DEFAULT stage_kind='approval',
// and that the DB CHECK constraints reject unknown stage_kind values on both
// approval_route_stages and approval_stage_instances.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/tests/integration/testdb"
)

// TestStageKindSchemaExpand_Default asserts a factory-seeded route stage (via
// the canonical testdb factory) reads the migration 0286 DEFAULT
// stage_kind='approval' with no explicit value supplied — proving the
// migration applied cleanly on top of 0285 and existing rows behave
// unchanged.
func TestStageKindSchemaExpand_Default(t *testing.T) {
	db, _ := testdb.Open(t)

	route := testdb.NewApprovalRoute(t, db,
		testdb.WithGovernanceClass("controlado"),
		// Governed class => the route may not commit stageless (ADR 0087 /
		// migration 0316), so its stage is seeded in the route's own tx.
		testdb.WithStageSeeder(func(tx *sql.Tx, routeID string) error {
			return insertRouteStageTx(tx, routeID, "Stage 1")
		}),
	)

	var stageKind string
	if err := db.QueryRowContext(context.Background(),
		`SELECT stage_kind FROM public.approval_route_stages WHERE route_id = $1::uuid LIMIT 1`,
		route.ID,
	).Scan(&stageKind); err != nil {
		// The stage seeder above always commits a stage row with the route, and
		// it supplies no stage_kind — so this read exercises the column DEFAULT.
		t.Fatalf("query approval_route_stages: %v", err)
	}

	if stageKind != "approval" {
		t.Fatalf("approval_route_stages.stage_kind DEFAULT = %q, want %q", stageKind, "approval")
	}

	// approval_stage_instances: seed via a minimal instance the same way
	// (no stage_kind supplied) and assert the DEFAULT there too.
	instance := testdb.NewApprovalInstance(t, db, testdb.WithRoute(route))

	var instStageID string
	var instStageKind string
	if err := db.QueryRowContext(context.Background(), `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status)
		VALUES ($1::uuid, 1, 'Stage 1', 'author', 'document.review', 'area-1',
		        'any_1_of', 'reduce_quorum', '[]'::jsonb, 'active')
		RETURNING id::text, stage_kind`,
		instance.ID,
	).Scan(&instStageID, &instStageKind); err != nil {
		t.Fatalf("seed approval_stage_instances row: %v", err)
	}
	if instStageKind != "approval" {
		t.Fatalf("approval_stage_instances.stage_kind DEFAULT = %q, want %q", instStageKind, "approval")
	}
	_ = instStageID
}

// TestStageKindSchemaExpand_RejectsUnknownValue asserts the CHECK constraint
// added by migration 0286 rejects any stage_kind value outside
// ('review','approval') on both approval_route_stages and
// approval_stage_instances, surfacing the named constraint in the error.
func TestStageKindSchemaExpand_RejectsUnknownValue(t *testing.T) {
	db, _ := testdb.Open(t)

	t.Run("approval_route_stages", func(t *testing.T) {
		route := testdb.NewApprovalRoute(t, db,
			testdb.WithGovernanceClass("controlado"),
			// Governed class => the route may not commit stageless (ADR 0087 /
			// migration 0316), so its stage is seeded in the route's own tx.
			testdb.WithStageSeeder(func(tx *sql.Tx, routeID string) error {
				return insertRouteStageTx(tx, routeID, "Stage 1")
			}),
		)

		_, err := db.ExecContext(context.Background(), `
			INSERT INTO public.approval_route_stages
			  (route_id, stage_order, name, required_capability,
			   quorum, on_eligibility_drift, stage_kind)
			VALUES ($1::uuid, 2, 'Stage 2', 'document.review',
			        'any_1_of', 'reduce_quorum', 'signature')`,
			route.ID,
		)
		assertCheckViolation(t, err, "approval_route_stages_stage_kind_check")
	})

	t.Run("approval_stage_instances", func(t *testing.T) {
		route := testdb.NewApprovalRoute(t, db,
			testdb.WithGovernanceClass("controlado"),
			// Governed class => the route may not commit stageless (ADR 0087 /
			// migration 0316), so its stage is seeded in the route's own tx.
			testdb.WithStageSeeder(func(tx *sql.Tx, routeID string) error {
				return insertRouteStageTx(tx, routeID, "Stage 1")
			}),
		)
		instance := testdb.NewApprovalInstance(t, db, testdb.WithRoute(route))

		_, err := db.ExecContext(context.Background(), `
			INSERT INTO public.approval_stage_instances
			  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
			   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
			   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
			VALUES ($1::uuid, 1, 'Stage 1', 'author', 'document.review', 'area-1',
			        'any_1_of', 'reduce_quorum', '[]'::jsonb, 'active', 'signature')`,
			instance.ID,
		)
		assertCheckViolation(t, err, "approval_stage_instances_stage_kind_check")
	})

	t.Run("update_existing_row_to_invalid_value", func(t *testing.T) {
		route := testdb.NewApprovalRoute(t, db,
			testdb.WithGovernanceClass("controlado"),
			// Governed class => the route may not commit stageless (ADR 0087 /
			// migration 0316), so its stage is seeded in the route's own tx.
			testdb.WithStageSeeder(func(tx *sql.Tx, routeID string) error {
				return insertRouteStageTx(tx, routeID, "Stage 1")
			}),
		)
		var stageID string
		if err := db.QueryRowContext(context.Background(), `
			INSERT INTO public.approval_route_stages
			  (route_id, stage_order, name, required_capability,
			   quorum, on_eligibility_drift)
			VALUES ($1::uuid, 2, 'Stage 2', 'document.review',
			        'any_1_of', 'reduce_quorum')
			RETURNING id::text`,
			route.ID,
		).Scan(&stageID); err != nil {
			t.Fatalf("seed approval_route_stages row: %v", err)
		}

		_, err := db.ExecContext(context.Background(),
			`UPDATE public.approval_route_stages SET stage_kind = 'signature' WHERE id = $1::uuid`,
			stageID,
		)
		assertCheckViolation(t, err, "approval_route_stages_stage_kind_check")
	})
}

// assertCheckViolation fails the test unless err is a Postgres check_violation
// (SQLSTATE 23514) naming the given constraint.
func assertCheckViolation(t *testing.T, err error, constraintName string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected check_violation for constraint %s, got nil error", constraintName)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23514" {
			t.Fatalf("expected SQLSTATE 23514 (check_violation), got %s: %v", pgErr.Code, err)
		}
		if pgErr.ConstraintName != constraintName {
			t.Fatalf("expected constraint %q, got %q (err: %v)", constraintName, pgErr.ConstraintName, err)
		}
		return
	}
	// Fallback string match in case the driver doesn't populate ConstraintName.
	if !strings.Contains(err.Error(), constraintName) {
		t.Fatalf("expected error mentioning constraint %q, got: %v", constraintName, err)
	}
}
