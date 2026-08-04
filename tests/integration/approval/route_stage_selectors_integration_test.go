//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 2 (M4 ActorSelector, B2 DB expand)
// real-Postgres coverage for public.approval_route_stage_selectors
// (migration 0303) and the Go read/write paths that keep it in sync.
// Selectors is the sole source of truth for a stage's actor pool post-slice-6b
// (migration 0305 dropped approval_route_stages.required_role/area_code); the
// flat<->selector synthesis now lives at the HTTP boundary
// (route_admin_handler.go), not in domain.Stage or RouteAdminService.
//
//  1. Round-trip: RouteAdminService.Create persists a stage with an explicit
//     role_in_fixed_area selector — insertRouteStages writes it via
//     insertRouteStageSelectors, and postgresApprovalRepository.LoadRoute reads
//     it back into domain.Stage.Selectors. Proves write+read of selector rows
//     through the real Go paths.
//  2. CHECK rejection: a raw malformed insert per selector kind is rejected by
//     approval_route_stage_selectors_fields_consistent.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestRouteStageSelectors_RoundTrip_RealDB proves the create-route write path
// (RouteAdminService.Create -> insertRouteStages -> insertRouteStageSelectors)
// and the LoadRoute read path (postgresApprovalRepository.LoadRoute ->
// loadRouteStageSelectors) round-trip a legacy stage's synthesized
// role_in_fixed_area selector.
func TestRouteStageSelectors_RoundTrip_RealDB(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	tenantID := testdb.NewTenant(t, database).ID
	taxonomy := testdb.NewTaxonomy(t, database, testdb.WithTenant(tenantID), testdb.WithGovernanceClass("controlado"))
	actor := testdb.NewUser(t, database, testdb.WithTenant(tenantID), testdb.WithRole("system_admin"))

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	idCtx := platformtenant.WithActorID(platformtenant.WithTenantID(ctx, tenantID), actor.ID)
	runner := db.NewTxRunner(database)

	stages := []domain.Stage{
		{
			Order:              1,
			Name:               "quality",
			RequiredCapability: "document.signoff",
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorRoleInFixedArea, Role: "qa_reviewer", AreaCode: "area-1"},
			},
		},
	}

	result, err := services.RouteAdmin.Create(idCtx, runner, application.CreateRouteInput{
		TenantID:    tenantID,
		ProfileCode: taxonomy.ProfileCode,
		Name:        "Selector Round Trip Route",
		ActorUserID: actor.ID,
		Stages:      stages,
	})
	if err != nil {
		t.Fatalf("RouteAdmin.Create: %v", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	route, err := repo.LoadRoute(ctx, tx, tenantID, result.RouteID)
	if err != nil {
		t.Fatalf("LoadRoute: %v", err)
	}
	if len(route.Stages) != 1 {
		t.Fatalf("route.Stages = %d stages, want 1", len(route.Stages))
	}
	got := route.Stages[0].Selectors
	want := []domain.ActorSelector{{Kind: domain.SelectorRoleInFixedArea, Role: "qa_reviewer", AreaCode: "area-1"}}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("route.Stages[0].Selectors = %+v, want %+v", got, want)
	}
}

// TestRouteStageSelectors_CheckRejection_RealDB proves the per-kind CHECK
// (approval_route_stage_selectors_fields_consistent, migration 0303) rejects
// a malformed selector row for each of the four kinds.
func TestRouteStageSelectors_CheckRejection_RealDB(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	// Route + stage in ONE tx: a controlado route may not commit stageless
	// (ADR 0087 / migration 0316).
	var stageID string
	route := testdb.NewApprovalRoute(t, database,
		testdb.WithGovernanceClass("controlado"),
		testdb.WithStageSeeder(func(tx *sql.Tx, routeID string) error {
			return tx.QueryRowContext(ctx, `
				INSERT INTO public.approval_route_stages
				  (route_id, stage_order, name, required_capability,
				   quorum, on_eligibility_drift)
				VALUES ($1::uuid, 1, 'Stage 1', 'document.review',
				        'any_1_of', 'reduce_quorum')
				RETURNING id`,
				routeID,
			).Scan(&stageID)
		}),
	)

	cases := []struct {
		name     string
		kind     string
		userID   sql.NullString
		role     sql.NullString
		areaCode sql.NullString
	}{
		{
			name:     "named_user with role set",
			kind:     "named_user",
			userID:   sql.NullString{String: "user-1", Valid: true},
			role:     sql.NullString{String: "qa_reviewer", Valid: true},
			areaCode: sql.NullString{},
		},
		{
			name:     "role_in_fixed_area missing area_code",
			kind:     "role_in_fixed_area",
			userID:   sql.NullString{},
			role:     sql.NullString{String: "qa_reviewer", Valid: true},
			areaCode: sql.NullString{},
		},
		{
			name:     "role_in_document_area with area_code set",
			kind:     "role_in_document_area",
			userID:   sql.NullString{},
			role:     sql.NullString{String: "qa_reviewer", Valid: true},
			areaCode: sql.NullString{String: "area-1", Valid: true},
		},
		{
			name:     "submit_choice missing role",
			kind:     "submit_choice",
			userID:   sql.NullString{},
			role:     sql.NullString{},
			areaCode: sql.NullString{String: "area-1", Valid: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := database.ExecContext(ctx, `
				INSERT INTO public.approval_route_stage_selectors
					(tenant_id, route_stage_id, selector_order, kind, user_id, role, area_code)
				VALUES ($1::uuid, $2::uuid, 99, $3, $4, $5, $6)`,
				route.TenantID, stageID, tc.kind, tc.userID, tc.role, tc.areaCode,
			)
			assertSelectorCheckViolation(t, err)
		})
	}
}

// assertSelectorCheckViolation fails the test unless err is a Postgres CHECK
// violation (SQLSTATE 23514) on approval_route_stage_selectors_fields_consistent.
func assertSelectorCheckViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a CHECK violation, got nil error")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23514" {
			t.Fatalf("expected SQLSTATE 23514 (check_violation), got %s: %v", pgErr.Code, err)
		}
		if !strings.Contains(pgErr.ConstraintName, "fields_consistent") {
			t.Fatalf("expected constraint name containing %q, got %q", "fields_consistent", pgErr.ConstraintName)
		}
		return
	}
	if !strings.Contains(err.Error(), "fields_consistent") {
		t.Fatalf("expected error mentioning the fields_consistent CHECK, got: %v", err)
	}
}
