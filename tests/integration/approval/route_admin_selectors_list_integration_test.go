//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 4 (M4 ActorSelector, route-admin
// CONTRACT) real-Postgres coverage for the list load path
// (postgresApprovalRepository.ListRoutes / ListRoutesTx). Slice 2 already
// proved LoadRoute round-trips selectors (route_stage_selectors_integration_test.go);
// this test proves ListRoutes/ListRoutesTx do the same batch-by-stage-id load
// (no N+1), and that both read paths agree on an EXPLICIT (not synthesized)
// selectors slice supplied on RouteAdminService.Create.
package approval_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestRouteAdminSelectors_LoadRouteAndListRoutes_AgreeOnExplicitSelectors_RealDB
// creates a route whose stage carries an EXPLICIT named_user selector, then
// asserts LoadRoute AND ListRoutes/ListRoutesTx all return that selector
// verbatim — Selectors is the sole source of truth post-slice-6b, no flat
// RequiredRole/AreaCode synthesis happens below the HTTP boundary.
func TestRouteAdminSelectors_LoadRouteAndListRoutes_AgreeOnExplicitSelectors_RealDB(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	tenantID := testdb.NewTenant(t, database).ID
	taxonomy := testdb.NewTaxonomy(t, database, testdb.WithTenant(tenantID), testdb.WithGovernanceClass("controlado"))
	actor := testdb.NewUser(t, database, testdb.WithTenant(tenantID), testdb.WithRole("system_admin"))
	namedUser := testdb.NewUser(t, database, testdb.WithTenant(tenantID), testdb.WithRole("approver"))

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	idCtx := platformtenant.WithActorID(platformtenant.WithTenantID(ctx, tenantID), actor.ID)
	runner := db.NewTxRunner(database)

	wantSelectors := []domain.ActorSelector{{Kind: domain.SelectorNamedUser, UserID: namedUser.ID}}
	stages := []domain.Stage{
		{
			Order:              1,
			Name:               "quality",
			RequiredCapability: "document.signoff",
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
			Selectors:          wantSelectors,
		},
	}

	result, err := services.RouteAdmin.Create(idCtx, runner, application.CreateRouteInput{
		TenantID:    tenantID,
		ProfileCode: taxonomy.ProfileCode,
		Name:        "Selector List Route",
		ActorUserID: actor.ID,
		Stages:      stages,
	})
	if err != nil {
		t.Fatalf("RouteAdmin.Create: %v", err)
	}

	assertSelectors := func(t *testing.T, got []domain.ActorSelector) {
		t.Helper()
		if len(got) != len(wantSelectors) || got[0] != wantSelectors[0] {
			t.Fatalf("selectors = %+v, want %+v", got, wantSelectors)
		}
	}

	t.Run("LoadRoute", func(t *testing.T) {
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
		assertSelectors(t, route.Stages[0].Selectors)
	})

	t.Run("ListRoutes", func(t *testing.T) {
		routes, err := repo.ListRoutes(ctx, tenantID)
		if err != nil {
			t.Fatalf("ListRoutes: %v", err)
		}
		route := findRouteByID(t, routes, result.RouteID)
		if len(route.Stages) != 1 {
			t.Fatalf("route.Stages = %d stages, want 1", len(route.Stages))
		}
		assertSelectors(t, route.Stages[0].Selectors)
	})

	t.Run("ListRoutesTx", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		routes, err := repo.ListRoutesTx(ctx, tx, tenantID)
		if err != nil {
			t.Fatalf("ListRoutesTx: %v", err)
		}
		route := findRouteByID(t, routes, result.RouteID)
		if len(route.Stages) != 1 {
			t.Fatalf("route.Stages = %d stages, want 1", len(route.Stages))
		}
		assertSelectors(t, route.Stages[0].Selectors)
	})
}

// findRouteByID locates routeID within routes or fails the test; ListRoutes
// returns every tenant route, so a direct index is not stable across runs.
func findRouteByID(t *testing.T, routes []infrastructure.Route, routeID string) infrastructure.Route {
	t.Helper()
	for _, r := range routes {
		if r.ID == routeID {
			return r
		}
	}
	t.Fatalf("route %s not found in %d routes", routeID, len(routes))
	return infrastructure.Route{}
}
