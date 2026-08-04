//go:build integration
// +build integration

package approvalhttp

// ADR 0087 (P0-1) — HTTP-level proof that a ZERO-STAGE route is creatable
// through the API for a livre profile, and only for a livre profile.
//
// The contract used to carry `minItems: 1` on CreateRouteRequest.stages and
// the hand-written contract layer rejected len==0 outright, which made a livre
// route un-creatable through the API at all — the config-first onboarding flow
// would have had no way to configure the one class that needs no approval.
// Stage COUNT is a policy question, so the wire shape must permit zero and the
// policy layers (application RoutePolicy, and authoritatively the DB
// route-shape trigger) decide.
//
// Drives the real handler over the real RouteAdminService and a real DB, with
// the real approval->taxonomy policy reader wired, so the 422 problem codes
// asserted here are the ones a client actually receives.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
)

func postZeroStageRoute(t *testing.T, governanceClass string) (int, string) {
	t.Helper()

	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass(governanceClass))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil).
		WithProfilePolicyReader(approvalrepo.NewProfilePolicyReader(taxonomyinfra.NewProfileRepository(dbc)))

	h := &Handler{routeAdmin: services.RouteAdmin, runner: db.NewTxRunner(dbc)}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"` + tax.ProfileCode + `","name":"Livre Route","stages":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.NewString())
	ctx := tenant.WithTenantID(context.Background(), tnt.ID)
	// TxRunner's chokepoint auto-seed reads the PLATFORM actor id to seed the
	// metaldocs.actor_id GUC authz.Require needs; the iam auth context is what
	// the handler itself reads. Both are set in production by the middleware
	// chain, so both are set here.
	ctx = tenant.WithActorID(ctx, actor.ID)
	ctx = iamdomain.WithAuthContext(ctx, actor.ID, []iamdomain.Role{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr.Code, rr.Body.String()
}

// TestCreateRoute_ZeroStages_LivreProfile_Created is the positive case: the
// configured no-approval route.
func TestCreateRoute_ZeroStages_LivreProfile_Created(t *testing.T) {
	code, body := postZeroStageRoute(t, "livre")
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (a livre route MUST be creatable through the API; body: %s)", code, body)
	}
}

// TestCreateRoute_ZeroStages_GovernedProfiles_Rejected asserts the same body is
// a 422 for BOTH governed classes, each with its own problem code — the wire
// permits zero stages, the policy layer does not.
func TestCreateRoute_ZeroStages_GovernedProfiles_Rejected(t *testing.T) {
	cases := []struct {
		class    string
		wantCode string
	}{
		{class: "controlado", wantCode: "validation.approval_stage_required"},
		{class: "simples", wantCode: "validation.route_stage_required"},
	}
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			status, body := postZeroStageRoute(t, tc.class)
			if status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422 (body: %s)", status, body)
			}
			var problem struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal([]byte(body), &problem); err != nil {
				t.Fatalf("decode problem+json: %v (body: %s)", err, body)
			}
			if problem.Code != tc.wantCode {
				t.Fatalf("problem code = %q, want %q (body: %s)", problem.Code, tc.wantCode, body)
			}
		})
	}
}
