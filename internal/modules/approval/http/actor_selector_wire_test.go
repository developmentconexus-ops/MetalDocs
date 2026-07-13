package approvalhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// TestCreateRoute_SelectorRoundTrip pins the M4 ActorSelector contract wiring
// (unit 3.2 slice 7a, wire contract extermination) end-to-end at the wire
// boundary: a stage carrying only an explicit named_user selector (the flat
// required_role/area_code wire fields no longer exist at all) decodes through
// CreateRouteHandler into a domain.Stage.Selectors slice, and the same
// selector shape read back through ListRoutesHandler's wire mapping
// round-trips byte-for-byte.
func TestCreateRoute_SelectorRoundTrip(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-123"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"named_user","user_id":"user-42"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if len(svc.createReq.Stages) != 1 {
		t.Fatalf("stages len = %d, want 1", len(svc.createReq.Stages))
	}
	gotSelectors := svc.createReq.Stages[0].Selectors
	if len(gotSelectors) != 1 {
		t.Fatalf("selectors len = %d, want 1", len(gotSelectors))
	}
	want := domain.ActorSelector{Kind: domain.SelectorNamedUser, UserID: "user-42"}
	if gotSelectors[0] != want {
		t.Fatalf("decoded selector = %+v, want %+v", gotSelectors[0], want)
	}

	// Now feed the same shape (as the DB read-side would project it) back
	// through ListRoutesHandler and assert the wire selector round-trips.
	listSvc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{Routes: []infrastructure.Route{
			{
				ID: "route-123", Name: "Ops Route", TenantID: "tenant-1", ProfileCode: "ops",
				Active: true, Version: 1, Total: 1,
				Stages: []infrastructure.RouteStage{
					{
						Order: 1, Name: "Review", RequiredCapability: "document.signoff",
						Quorum: "any_1_of", DriftPolicy: "reduce_quorum",
						Selectors: []domain.ActorSelector{{Kind: domain.SelectorNamedUser, UserID: "user-42"}},
					},
				},
			},
		}},
	}
	listH := &Handler{routeAdmin: listSvc}
	listMux := routeAdminTestMux(listH)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	listReq = listReq.WithContext(tenant.WithTenantID(listReq.Context(), "tenant-1"))
	listReq = listReq.WithContext(iamdomain.WithAuthContext(listReq.Context(), "actor-1", []iamdomain.Role{}))

	listRR := httptest.NewRecorder()
	listMux.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %s", listRR.Code, listRR.Body.String())
	}
	var out contracts.ListRoutesResponse
	if err := json.NewDecoder(listRR.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Routes) != 1 || len(out.Routes[0].Stages) != 1 {
		t.Fatalf("unexpected list shape: %+v", out)
	}
	gotWire := out.Routes[0].Stages[0].Selectors
	if len(gotWire) != 1 {
		t.Fatalf("wire selectors len = %d, want 1", len(gotWire))
	}
	wantWire := contracts.ActorSelector{Kind: contracts.SelectorKindNamedUser, UserID: "user-42"}
	if gotWire[0] != wantWire {
		t.Fatalf("wire selector = %+v, want %+v", gotWire[0], wantWire)
	}
}
