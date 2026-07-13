//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 2 (M4 ActorSelector, B2 DB expand)
// real-Postgres coverage for public.approval_route_stage_selectors
// (migration 0303) and the Go read/write paths that keep it in sync with
// approval_route_stages during the flat-column/selector coexistence window
// (docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md §3).
//
//  1. Round-trip: RouteAdminService.Create persists a legacy stage
//     (RequiredRole/AreaCode only, no explicit Selectors) — insertRouteStages
//     writes the stage's EffectiveSelectors() (a synthesized role_in_fixed_area
//     selector) via insertRouteStageSelectors, and
//     postgresApprovalRepository.LoadRoute reads it back into
//     domain.Stage.Selectors. Proves write+read of selector rows through the
//     real Go paths, not just the migration's own backfill.
//  2. CHECK rejection: a raw malformed insert per selector kind is rejected by
//     approval_route_stage_selectors_fields_consistent.
//  3. Backfill coverage: every approval_route_stages row (seeded before this
//     slice's Go writer existed, via the raw-SQL idiom other tests already
//     use) has >=1 selector row after the 0303 migration's backfill.
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
	taxonomy := testdb.NewTaxonomy(t, database, testdb.WithTenant(tenantID))
	actor := testdb.NewUser(t, database, testdb.WithTenant(tenantID), testdb.WithRole("system_admin"))

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	idCtx := platformtenant.WithActorID(platformtenant.WithTenantID(ctx, tenantID), actor.ID)
	runner := db.NewTxRunner(database)

	stages := []domain.Stage{
		{
			Order:              1,
			Name:               "quality",
			RequiredRole:       "qa_reviewer",
			RequiredCapability: "document.signoff",
			AreaCode:           "area-1",
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
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

	route := testdb.NewApprovalRoute(t, database)
	stageID := seedBareRouteStage(t, database, route.ID)

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

// TestRouteStageSelectors_BackfillCoverage_RealDB proves migration 0303's
// backfill statement: a bare approval_route_stages row (the pre-slice-2 shape,
// no Go selector writer involved) gains exactly one role_in_fixed_area selector
// carrying its required_role/area_code and the owning route's tenant_id, and
// the WHERE NOT EXISTS guard makes a re-run a no-op (idempotent).
//
// testdb clones an already-migrated template with zero tenant rows, so the
// one-time backfill in 0303 has nothing to act on at clone time — a row seeded
// now postdates it. This test therefore executes the migration's exact backfill
// INSERT against the seeded row, which is what validates the SQL that WILL run
// against real pre-existing rows on a production deploy.
func TestRouteStageSelectors_BackfillCoverage_RealDB(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	route := testdb.NewApprovalRoute(t, database)
	stageID := seedBareRouteStage(t, database, route.ID)

	// The verbatim backfill statement from db/migrations/0303 (kept in sync).
	const backfill = `
		INSERT INTO public.approval_route_stage_selectors
			(tenant_id, route_stage_id, selector_order, kind, role, area_code)
		SELECT r.tenant_id, s.id, 1, 'role_in_fixed_area', s.required_role, s.area_code
		  FROM public.approval_route_stages s
		  JOIN public.approval_routes r ON r.id = s.route_id
		 WHERE s.id = $1::uuid
		   AND NOT EXISTS (
			SELECT 1 FROM public.approval_route_stage_selectors x
			 WHERE x.route_stage_id = s.id
		)`

	if _, err := database.ExecContext(ctx, backfill, stageID); err != nil {
		t.Fatalf("backfill insert: %v", err)
	}

	assertBackfilledSelector := func() {
		t.Helper()
		var count int
		if err := database.QueryRowContext(ctx,
			`SELECT count(*) FROM public.approval_route_stage_selectors WHERE route_stage_id = $1::uuid`,
			stageID,
		).Scan(&count); err != nil {
			t.Fatalf("count selectors: %v", err)
		}
		if count != 1 {
			t.Fatalf("selector count for stage %s = %d, want exactly 1", stageID, count)
		}
		var kind, role, areaCode string
		if err := database.QueryRowContext(ctx,
			`SELECT kind, role, area_code FROM public.approval_route_stage_selectors WHERE route_stage_id = $1::uuid AND selector_order = 1`,
			stageID,
		).Scan(&kind, &role, &areaCode); err != nil {
			t.Fatalf("read backfilled selector: %v", err)
		}
		if kind != "role_in_fixed_area" || role != "author" || areaCode != "area-1" {
			t.Fatalf("backfilled selector = {kind:%s role:%s area:%s}, want {role_in_fixed_area author area-1}", kind, role, areaCode)
		}
	}

	assertBackfilledSelector()

	// Idempotency: re-running the backfill (the WHERE NOT EXISTS guard) is a
	// no-op — still exactly one selector, unchanged.
	if _, err := database.ExecContext(ctx, backfill, stageID); err != nil {
		t.Fatalf("backfill re-run: %v", err)
	}
	assertBackfilledSelector()
}

// seedBareRouteStage inserts a public.approval_route_stages row directly
// (mirrors route_versioning_test.go's seedRouteStage — no tripwire on this
// table), simulating a stage written before this slice's Go selector writer
// existed. The 0303 migration's backfill (already applied to the testdb
// template) is what gives this row its selector row, not any code this slice
// adds — that is exactly the backfill-coverage property under test.
func seedBareRouteStage(t *testing.T, database *sql.DB, routeID string) string {
	t.Helper()
	var stageID string
	if err := database.QueryRowContext(context.Background(), `
		INSERT INTO public.approval_route_stages
		  (route_id, stage_order, name, required_role, required_capability,
		   area_code, quorum, on_eligibility_drift)
		VALUES ($1::uuid, 1, 'Stage 1', 'author', 'document.review', 'area-1',
		        'any_1_of', 'reduce_quorum')
		RETURNING id`,
		routeID,
	).Scan(&stageID); err != nil {
		t.Fatalf("seed approval_route_stages row: %v", err)
	}
	return stageID
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
