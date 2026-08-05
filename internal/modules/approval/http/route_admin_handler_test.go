package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

type fakeRouteAdminService struct {
	createResult application.CreateRouteResult
	createErr    error
	createReq    application.CreateRouteInput

	updateResult application.UpdateRouteResult
	updateErr    error
	updateReq    application.UpdateRouteInput

	deactivateResult application.DeactivateRouteResult
	deactivateErr    error
	deactivateReq    application.DeactivateRouteInput

	listResult application.ListRoutesResult
	listErr    error
	listTenant string
	listActor  string
}

func (f *fakeRouteAdminService) Create(_ context.Context, _ db.TxRunner, in application.CreateRouteInput) (application.CreateRouteResult, error) {
	f.createReq = in
	if f.createErr != nil {
		return application.CreateRouteResult{}, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeRouteAdminService) Update(_ context.Context, _ db.TxRunner, in application.UpdateRouteInput) (application.UpdateRouteResult, error) {
	f.updateReq = in
	if f.updateErr != nil {
		return application.UpdateRouteResult{}, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeRouteAdminService) Deactivate(_ context.Context, _ db.TxRunner, in application.DeactivateRouteInput) (application.DeactivateRouteResult, error) {
	f.deactivateReq = in
	if f.deactivateErr != nil {
		return application.DeactivateRouteResult{}, f.deactivateErr
	}
	return f.deactivateResult, nil
}

func (f *fakeRouteAdminService) List(_ context.Context, _ db.TxRunner, tenantID, actorID string) (application.ListRoutesResult, error) {
	f.listTenant = tenantID
	f.listActor = actorID
	if f.listErr != nil {
		return application.ListRoutesResult{}, f.listErr
	}
	return f.listResult, nil
}

func routeAdminTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/approval/routes", h.CreateRouteHandler)
	mux.HandleFunc("PUT /api/v1/approval/routes/{id}", h.UpdateRouteHandler)
	mux.HandleFunc("DELETE /api/v1/approval/routes/{id}", h.DeactivateRouteHandler)
	mux.HandleFunc("GET /api/v1/approval/routes", h.ListRoutesHandler)
	return mux
}

func TestCreateRoute_HappyPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "created"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeRouteAdminService{
				createResult: application.CreateRouteResult{RouteID: "route-123"},
			}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
			}

			var out contracts.RouteResponse
			if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.RouteID != "route-123" {
				t.Fatalf("route_id = %q, want %q", out.RouteID, "route-123")
			}

			if svc.createReq.TenantID != "tenant-1" {
				t.Fatalf("tenant_id = %q, want %q", svc.createReq.TenantID, "tenant-1")
			}
			if svc.createReq.ActorUserID != "actor-1" {
				t.Fatalf("actor_user_id = %q, want %q", svc.createReq.ActorUserID, "actor-1")
			}
			if svc.createReq.ProfileCode != "ops" {
				t.Fatalf("profile_code = %q, want %q", svc.createReq.ProfileCode, "ops")
			}
			if svc.createReq.Name != "Ops Route" {
				t.Fatalf("name = %q, want %q", svc.createReq.Name, "Ops Route")
			}
			if len(svc.createReq.Stages) != 1 {
				t.Fatalf("stages len = %d, want 1", len(svc.createReq.Stages))
			}

			stage := svc.createReq.Stages[0]
			flatRole, flatArea := domain.FlatRoleArea(stage.Selectors)
			if stage.Order != 1 || stage.Name != "Review" || flatRole != "approver" || stage.RequiredCapability != "document.signoff" || flatArea != "ops" {
				t.Fatalf("unexpected stage mapping: %+v", stage)
			}
			if stage.Quorum != domain.QuorumPolicy("any_1_of") {
				t.Fatalf("stage quorum = %q, want %q", stage.Quorum, domain.QuorumPolicy("any_1_of"))
			}
			if stage.OnEligibilityDrift != domain.DriftPolicy("reduce_quorum") {
				t.Fatalf("stage drift policy = %q, want %q", stage.OnEligibilityDrift, domain.DriftPolicy("reduce_quorum"))
			}
		})
	}
}

func TestCreateRoute_CapDenied(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "capability denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeRouteAdminService{
				createErr: authz.ErrCapDenied{Capability: "route.admin", AreaCode: "tenant", ActorID: "actor-1"},
			}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
			req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
			}
		})
	}
}

func TestCreateRoute_DuplicateProfile(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "duplicate profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeRouteAdminService{createErr: infrastructure.ErrDuplicateRouteProfile}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
			req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
			}
		})
	}
}

func TestCreateRoute_ProfileUnknown(t *testing.T) {
	// A profile_code with no matching document profile trips the FK; the
	// service translates it to ErrRouteProfileUnknown, which must surface as a
	// 422 validation error — never an opaque 500.
	svc := &fakeRouteAdminService{createErr: application.ErrRouteProfileUnknown}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"qa-preview","name":"QA Preview","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	var prob struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&prob); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if prob.Code != "validation.profile_unknown" {
		t.Fatalf("code = %q, want %q", prob.Code, "validation.profile_unknown")
	}
}

// TestCreateRoute_SubjectFieldsOmitted_NoSubjectKindPassed verifies the
// backward-compat default path: a CreateRouteRequest with no subject_kind /
// subject_key fields decodes to a CreateRouteInput with both empty, leaving
// the service's default-to-document logic to apply.
func TestCreateRoute_SubjectFieldsOmitted_NoSubjectKindPassed(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-123"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if svc.createReq.SubjectKind != "" || svc.createReq.SubjectKey != "" {
		t.Fatalf("expected empty subject fields on omission; got kind=%q key=%q", svc.createReq.SubjectKind, svc.createReq.SubjectKey)
	}
}

// TestCreateRoute_SubjectFieldsPassedThrough verifies an explicit
// subject_kind/subject_key on the wire is decoded and passed to the service.
//
// The fixture uses a TEMPLATE subject with subject_key == profile_code: since
// ADR 0086 BOTH kinds are profile-keyed, so a divergent key of either kind is
// an illegal state (ErrDocumentSubjectKeyMismatch /
// ErrTemplateSubjectKeyMismatch) and cannot be used to prove pass-through.
// profile_code is now mandatory here too — see
// TestCreateRoute_TemplateSubjectRequiresProfileCode for the negative case.
func TestCreateRoute_SubjectFieldsPassedThrough(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-123"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","subject_kind":"template","subject_key":"ops","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if svc.createReq.SubjectKind != "template" || svc.createReq.SubjectKey != "ops" {
		t.Fatalf("subject fields not passed through: kind=%q key=%q", svc.createReq.SubjectKind, svc.createReq.SubjectKey)
	}
	if svc.createReq.ProfileCode != "ops" {
		t.Fatalf("profile_code = %q, want %q for a template route", svc.createReq.ProfileCode, "ops")
	}
}

// TestCreateRoute_ResponseProfileCode keeps QR-A finding C's wire discipline
// under ADR 0086: the create response's profile_code is read from the
// persisted row (CreateRouteResult.ProfileCode, F-E4-2), never echoed from the
// request, and it is a non-empty string for BOTH subject kinds now that
// template routes are profile-keyed. Decoded via raw json.RawMessage so the
// null-vs-missing-vs-empty-string distinction is checked on the actual wire
// bytes. The pointer stays nullable so a data defect would surface honestly as
// JSON null rather than collapse to "".
func TestCreateRoute_ResponseProfileCode(t *testing.T) {
	t.Run("template create emits the profile code", func(t *testing.T) {
		opsCode := "ops"
		svc := &fakeRouteAdminService{
			createResult: application.CreateRouteResult{RouteID: "route-tmpl-1", ProfileCode: &opsCode},
		}
		h := &Handler{routeAdmin: svc}
		mux := routeAdminTestMux(h)

		body := `{"profile_code":"ops","name":"Template Route","subject_kind":"template","subject_key":"ops","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
		req.Header.Set("Idempotency-Key", "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusCreated, rr.Body.String())
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
			t.Fatalf("decode raw response: %v", err)
		}
		code, ok := raw["profile_code"]
		if !ok {
			t.Fatalf("response missing profile_code key entirely: %s", rr.Body.String())
		}
		if string(code) != `"ops"` {
			t.Fatalf("template create profile_code raw = %s, want %q", code, `"ops"`)
		}
	})

	t.Run("document create emits the profile code unchanged", func(t *testing.T) {
		opsCode := "ops"
		svc := &fakeRouteAdminService{
			createResult: application.CreateRouteResult{RouteID: "route-doc-1", ProfileCode: &opsCode},
		}
		h := &Handler{routeAdmin: svc}
		mux := routeAdminTestMux(h)

		body := `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
		req.Header.Set("Idempotency-Key", "55555555-5555-4555-8555-555555555555")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusCreated, rr.Body.String())
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
			t.Fatalf("decode raw response: %v", err)
		}
		code, ok := raw["profile_code"]
		if !ok {
			t.Fatalf("response missing profile_code key entirely: %s", rr.Body.String())
		}
		if string(code) != `"ops"` {
			t.Fatalf("document create profile_code raw = %s, want %q", code, `"ops"`)
		}
	})
}

// TestCreateRoute_ReturnsPersistedProjection is the F-E4-2 regression pin: the
// 201 body must carry the real persisted route projection — the same shape the
// read side returns — not the former two-field stub that emitted
// name:"", version:0, active:false, in_use:false, stages:null, created_at:""
// for a fully-persisted row. Both subject kinds go through the same handler,
// so both are asserted.
func TestCreateRoute_ReturnsPersistedProjection(t *testing.T) {
	createdAt := time.Date(2026, 7, 29, 10, 30, 0, 0, time.UTC)
	quorumM := 2
	persistedStages := []domain.Stage{{
		Order:              1,
		Name:               "Review",
		RequiredCapability: "document.signoff",
		Quorum:             domain.QuorumPolicy("m_of_n"),
		QuorumM:            &quorumM,
		OnEligibilityDrift: domain.DriftPolicy("reduce_quorum"),
		Kind:               domain.StageKind("approval"),
		Selectors: []domain.ActorSelector{
			{Kind: domain.SelectorRoleInFixedArea, Role: "approver", AreaCode: "ops"},
		},
	}}

	opsCode := "ops"
	cases := map[string]struct {
		body            string
		result          application.CreateRouteResult
		wantProfileCode string // raw JSON bytes
	}{
		"document subject": {
			body: `{"profile_code":"ops","name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"m_of_n","quorum_m":2,"drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`,
			result: application.CreateRouteResult{
				RouteID: "route-doc-9", ProfileCode: &opsCode, Name: "Ops Route",
				Version: 1, Active: true, InUse: false, CreatedAt: createdAt, Stages: persistedStages,
			},
			wantProfileCode: `"ops"`,
		},
		"template subject": {
			body: `{"profile_code":"ops","name":"Template Route","subject_kind":"template","subject_key":"ops","stages":[{"order":1,"name":"Review","required_capability":"template.approve","quorum":"m_of_n","quorum_m":2,"drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`,
			result: application.CreateRouteResult{
				RouteID: "route-tmpl-9", ProfileCode: &opsCode, Name: "Template Route",
				Version: 1, Active: true, InUse: false, CreatedAt: createdAt, Stages: persistedStages,
			},
			wantProfileCode: `"ops"`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &fakeRouteAdminService{createResult: tc.result}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "77777777-7777-4777-8777-777777777777")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusCreated, rr.Body.String())
			}

			var out contracts.RouteResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.RouteID != tc.result.RouteID {
				t.Errorf("route_id = %q, want %q", out.RouteID, tc.result.RouteID)
			}
			if out.Name != tc.result.Name {
				t.Errorf("name = %q, want %q (was \"\" before F-E4-2)", out.Name, tc.result.Name)
			}
			if out.Version != tc.result.Version {
				t.Errorf("version = %d, want %d (was 0 before F-E4-2)", out.Version, tc.result.Version)
			}
			if !out.Active {
				t.Errorf("active = false, want true (was false before F-E4-2)")
			}
			if out.InUse {
				t.Errorf("in_use = true, want false (a just-created route has no instances)")
			}
			if out.CreatedAt != createdAt.Format(time.RFC3339) {
				t.Errorf("created_at = %q, want %q (was \"\" before F-E4-2)", out.CreatedAt, createdAt.Format(time.RFC3339))
			}
			if len(out.Stages) != 1 {
				t.Fatalf("stages len = %d, want 1 (was null before F-E4-2)", len(out.Stages))
			}
			stage := out.Stages[0]
			if stage.Order != 1 || stage.Name != "Review" || stage.RequiredCapability != "document.signoff" {
				t.Errorf("stage projection = %+v, want the persisted stage", stage)
			}
			if stage.Quorum != contracts.QuorumKind("m_of_n") || stage.QuorumM == nil || *stage.QuorumM != 2 {
				t.Errorf("stage quorum = %q/%v, want m_of_n/2", stage.Quorum, stage.QuorumM)
			}
			if len(stage.Selectors) != 1 || stage.Selectors[0].Role != "approver" {
				t.Errorf("stage selectors = %+v, want the persisted role_in_fixed_area selector", stage.Selectors)
			}
			// new_version is update-only per the spec ("Omitted for
			// create/deactivate") and must stay absent on create.
			if out.NewVersion != nil {
				t.Errorf("new_version = %v, want nil on create", out.NewVersion)
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
				t.Fatalf("decode raw: %v", err)
			}
			if got := string(raw["profile_code"]); got != tc.wantProfileCode {
				t.Errorf("profile_code raw = %s, want %s", got, tc.wantProfileCode)
			}
		})
	}
}

// TestCreateRoute_TemplateSubjectRequiresProfileCode is the ADR 0086
// inversion of the retired ..._TemplateSubjectRejectsProfileCode case: a
// create-route body with subject_kind=template and NO profile_code must be
// rejected as a 400 validation error before ever reaching the service. The
// old rule (template must not carry profile_code) is gone outright.
func TestCreateRoute_TemplateSubjectRequiresProfileCode(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-should-not-be-created"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"name":"Ops Route","subject_kind":"template","subject_key":"ops","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	if svc.createReq.SubjectKind != "" {
		t.Fatalf("service must not be invoked when validation fails, got createReq=%+v", svc.createReq)
	}
}

// TestCreateRoute_TemplateSubjectKeyMismatch_Returns422 is the template twin
// of the document mismatch mapping: application.ErrTemplateSubjectKeyMismatch
// (a template route whose subject_key diverges from profile_code — the
// pre-ADR-0086 template-instance keying) must surface as 422 problem+json,
// not a generic 500.
func TestCreateRoute_TemplateSubjectKeyMismatch_Returns422(t *testing.T) {
	svc := &fakeRouteAdminService{createErr: application.ErrTemplateSubjectKeyMismatch}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","subject_kind":"template","subject_key":"tmpl-custom-1","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusUnprocessableEntity, rr.Body.String())
	}
}

// TestCreateRoute_DocumentSubjectRequiresProfileCode pins the other direction
// of the S1 (F18) conditional contract at the handler layer: a create-route
// body with document (or absent) subject_kind and no profile_code must be
// rejected as a 400 validation error before reaching the service — the
// profile_code requirement moved from unconditional to conditional, and this
// guards against it silently becoming optional for document routes.
func TestCreateRoute_DocumentSubjectRequiresProfileCode(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-should-not-be-created"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	for name, body := range map[string]string{
		"absent kind":            `{"name":"Ops Route","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`,
		"explicit document kind": `{"name":"Ops Route","subject_kind":"document","subject_key":"ops","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "44444444-4444-4444-8444-444444444444")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusBadRequest, rr.Body.String())
			}
			if svc.createReq.Name != "" {
				t.Fatalf("service must not be invoked when validation fails, got createReq=%+v", svc.createReq)
			}
		})
	}
}

// TestCreateRoute_DocumentSubjectKeyMismatch_Returns422 verifies the reviewer
// finding from M3 P3.S2b-2: application.ErrDocumentSubjectKeyMismatch (a
// document-kind route created with subject_key != profile_code) must map to
// a 422 problem+json response, not fall through to a generic 500.
func TestCreateRoute_DocumentSubjectKeyMismatch_Returns422(t *testing.T) {
	svc := &fakeRouteAdminService{createErr: application.ErrDocumentSubjectKeyMismatch}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","subject_kind":"document","subject_key":"other-profile","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	var prob struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&prob); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if prob.Code != "validation.document_subject_key_mismatch" {
		t.Fatalf("code = %q, want %q", prob.Code, "validation.document_subject_key_mismatch")
	}
}

func TestListRoutes_CanonicalStageNames(t *testing.T) {
	// The list response must serialise stages with canonical field names
	// (`name`, `quorum`) — not the legacy `label`/`quorum_kind`.
	svc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{Routes: []infrastructure.Route{
			{
				ID: "r1", Name: "Ops", TenantID: "tenant-1", ProfileCode: "ops",
				Active: true, Version: 3, Total: 1,
				Stages: []infrastructure.RouteStage{
					{Order: 1, Name: "Review", RequiredCapability: "document.signoff", Quorum: "any_1_of", DriftPolicy: "reduce_quorum"},
				},
			},
		}},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	raw := rr.Body.String()
	for _, want := range []string{`"name":"Review"`, `"quorum":"any_1_of"`, `"order":1`, `"version":3`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("response missing %q; body = %s", want, raw)
		}
	}
	for _, banned := range []string{`"label"`, `"quorum_kind"`} {
		if strings.Contains(raw, banned) {
			t.Fatalf("response contains legacy field %q; body = %s", banned, raw)
		}
	}
}

func TestUpdateRoute_HappyPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "updated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := 2
			svc := &fakeRouteAdminService{
				updateResult: application.UpdateRouteResult{RouteID: "route-1", NewVersion: 5},
			}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			body := `{"name":"Ops Route v2","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"m_of_n","quorum_m":2,"drift_policy":"keep_snapshot","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
			req := httptest.NewRequest(http.MethodPut, "/api/v1/approval/routes/route-1", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
			req.Header.Set("If-Match", "\"v4\"")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}

			var out contracts.RouteResponse
			if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.RouteID != "route-1" {
				t.Fatalf("route_id = %q, want %q", out.RouteID, "route-1")
			}
			if out.NewVersion == nil || *out.NewVersion != 5 {
				t.Fatalf("new_version = %v, want %d", out.NewVersion, 5)
			}

			if svc.updateReq.TenantID != "tenant-1" || svc.updateReq.RouteID != "route-1" || svc.updateReq.ActorUserID != "actor-1" {
				t.Fatalf("unexpected request mapped to service: %+v", svc.updateReq)
			}
			if svc.updateReq.ExpectedVersion != 4 {
				t.Fatalf("expected version = %d, want 4", svc.updateReq.ExpectedVersion)
			}
			if svc.updateReq.Name != "Ops Route v2" {
				t.Fatalf("name = %q, want %q", svc.updateReq.Name, "Ops Route v2")
			}
			if len(svc.updateReq.Stages) != 1 {
				t.Fatalf("stages len = %d, want 1", len(svc.updateReq.Stages))
			}
			if svc.updateReq.Stages[0].Quorum != domain.QuorumPolicy("m_of_n") {
				t.Fatalf("stage quorum = %q, want %q", svc.updateReq.Stages[0].Quorum, domain.QuorumPolicy("m_of_n"))
			}
			if svc.updateReq.Stages[0].QuorumM == nil || *svc.updateReq.Stages[0].QuorumM != m {
				t.Fatalf("stage quorum_m = %#v, want %d", svc.updateReq.Stages[0].QuorumM, m)
			}
		})
	}
}

func TestUpdateRoute_RouteInUse(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "route in use"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeRouteAdminService{updateErr: infrastructure.ErrRouteInUse}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			body := `{"name":"Ops Route v2","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"all_of","drift_policy":"fail_stage","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
			req := httptest.NewRequest(http.MethodPut, "/api/v1/approval/routes/route-1", strings.NewReader(body))
			req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
			req.Header.Set("If-Match", "\"v4\"")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
			}
		})
	}
}

func TestDeactivateRoute_HappyPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "deactivated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeRouteAdminService{
				deactivateResult: application.DeactivateRouteResult{RouteID: "route-1"},
			}
			h := &Handler{routeAdmin: svc}
			mux := routeAdminTestMux(h)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(`{"reason":"end of lifecycle"}`))
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
			req.Header.Set("If-Match", "\"v3\"")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}

			var out contracts.RouteResponse
			if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.RouteID != "route-1" {
				t.Fatalf("route_id = %q, want %q", out.RouteID, "route-1")
			}
			if svc.deactivateReq.TenantID != "tenant-1" || svc.deactivateReq.RouteID != "route-1" || svc.deactivateReq.ActorUserID != "actor-1" {
				t.Fatalf("unexpected request mapped to service: %+v", svc.deactivateReq)
			}
			if svc.deactivateReq.ExpectedVersion != 3 {
				t.Fatalf("expected version = %d, want 3", svc.deactivateReq.ExpectedVersion)
			}
			if svc.deactivateReq.Reason != "end of lifecycle" {
				t.Fatalf("reason = %q, want %q", svc.deactivateReq.Reason, "end of lifecycle")
			}
			if svc.deactivateReq.IdempotencyKey != "11111111-1111-4111-8111-111111111111" {
				t.Fatalf("idempotency_key = %q, want %q", svc.deactivateReq.IdempotencyKey, "11111111-1111-4111-8111-111111111111")
			}
		})
	}
}

func TestDeactivateRoute_RejectsEmptyReason(t *testing.T) {
	svc := &fakeRouteAdminService{}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(`{"reason":"   "}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "\"v3\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeactivateRoute_BodyTooLargeReturns413(t *testing.T) {
	svc := &fakeRouteAdminService{}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	huge := strings.Repeat("a", 70*1024)
	body := `{"reason":"` + huge + `"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "\"v3\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestDeactivateRoute_WrongContentTypeReturns415(t *testing.T) {
	svc := &fakeRouteAdminService{}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(`reason=foo`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "\"v3\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnsupportedMediaType)
	}
}

func TestDeactivateRoute_EmptyBodyReturns400(t *testing.T) {
	svc := &fakeRouteAdminService{}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(""))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "\"v3\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeactivateRoute_UnknownFieldReturns400(t *testing.T) {
	svc := &fakeRouteAdminService{}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", strings.NewReader(`{"reason":"x","bogus":1}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "\"v3\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestListRoutes_PassesTenantAndActor(t *testing.T) {
	svc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-7"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-7", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if svc.listTenant != "tenant-7" || svc.listActor != "actor-7" {
		t.Fatalf("List args tenant=%q actor=%q; want tenant-7 / actor-7", svc.listTenant, svc.listActor)
	}
	var out contracts.ListRoutesResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Total != 0 || len(out.Routes) != 0 {
		t.Fatalf("got %+v; want empty envelope", out)
	}
}

func TestUpdateRoute_RequiresIfMatch(t *testing.T) {
	h := &Handler{routeAdmin: &fakeRouteAdminService{}}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/approval/routes/route-1", strings.NewReader(`{"name":"Ops Route v2","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"all_of","drift_policy":"keep_snapshot","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPreconditionRequired)
	}
}

func TestDeactivateRoute_RequiresIfMatch(t *testing.T) {
	h := &Handler{routeAdmin: &fakeRouteAdminService{}}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/approval/routes/route-1", nil)
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPreconditionRequired)
	}
}

func TestUpdateRoute_RejectsIfMatchV0(t *testing.T) {
	h := &Handler{routeAdmin: &fakeRouteAdminService{}}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/approval/routes/route-1", strings.NewReader(`{"name":"Ops Route v2","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"all_of","drift_policy":"keep_snapshot","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "v0")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestListRoutes_TotalReflectsRepoCount(t *testing.T) {
	svc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{Routes: []infrastructure.Route{
			{ID: "r1", Total: 42},
		}},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-7"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-7", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var out contracts.ListRoutesResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Total != 42 {
		t.Fatalf("total = %d; want 42", out.Total)
	}
}

// TestListRoutes_ExposesSubjectFields verifies P2.S3's read-model delta: the
// list response includes subject_kind/subject_key for a route, mirroring the
// document/profile_code default for an existing document route.
func TestListRoutes_ExposesSubjectFields(t *testing.T) {
	svc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{Routes: []infrastructure.Route{
			{
				ID: "r1", Name: "Ops", TenantID: "tenant-1", ProfileCode: "ops",
				Active: true, Version: 3, Total: 1,
				SubjectKind: "document",
				SubjectKey:  "ops",
				Stages: []infrastructure.RouteStage{
					{Order: 1, Name: "Review", RequiredCapability: "document.signoff", Quorum: "any_1_of", DriftPolicy: "reduce_quorum"},
				},
			},
		}},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var out contracts.ListRoutesResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Routes) != 1 {
		t.Fatalf("routes len = %d, want 1", len(out.Routes))
	}
	if out.Routes[0].SubjectKind != "document" || out.Routes[0].SubjectKey != "ops" {
		t.Fatalf("subject fields = kind=%q key=%q; want kind=document key=ops", out.Routes[0].SubjectKind, out.Routes[0].SubjectKey)
	}
}

// TestListRoutes_ProfileCodeForTemplateRoutes is the ADR 0086 inversion of the
// retired ..._ProfileCodeNullForTemplateRoutes case: a template route IS
// profile-keyed (approval_routes_template_subject_key_check, migration 0315),
// so the list response must carry its profile_code as a non-empty string —
// the same shape a document route already has, and with the same
// subject_key == profile_code identity. This still decodes the raw response
// body (not the typed contracts.ListRouteItem) because a typed *string decode
// cannot distinguish an absent/null key from "": the assertion must inspect
// the wire bytes themselves.
func TestListRoutes_ProfileCodeForTemplateRoutes(t *testing.T) {
	svc := &fakeRouteAdminService{
		listResult: application.ListRoutesResult{Routes: []infrastructure.Route{
			{
				ID: "r1", Name: "Ops", TenantID: "tenant-1", ProfileCode: "ops",
				Active: true, Version: 3, Total: 2,
				SubjectKind: "document",
				SubjectKey:  "ops",
				Stages: []infrastructure.RouteStage{
					{Order: 1, Name: "Review", RequiredCapability: "document.signoff", Quorum: "any_1_of", DriftPolicy: "reduce_quorum"},
				},
			},
			{
				ID: "r2", Name: "Tmpl", TenantID: "tenant-1", ProfileCode: "sop",
				Active: true, Version: 1, Total: 2,
				SubjectKind: "template",
				SubjectKey:  "sop",
				Stages: []infrastructure.RouteStage{
					{Order: 1, Name: "Review", RequiredCapability: "template.approve", Quorum: "any_1_of", DriftPolicy: "reduce_quorum"},
				},
			},
		}},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/routes", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var raw struct {
		Routes []map[string]json.RawMessage `json:"routes"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	if len(raw.Routes) != 2 {
		t.Fatalf("routes len = %d, want 2", len(raw.Routes))
	}

	docCode, ok := raw.Routes[0]["profile_code"]
	if !ok {
		t.Fatalf("document route missing profile_code key entirely")
	}
	if string(docCode) != `"ops"` {
		t.Fatalf("document route profile_code raw = %s, want %q", docCode, `"ops"`)
	}

	tmplCode, ok := raw.Routes[1]["profile_code"]
	if !ok {
		t.Fatalf("template route missing profile_code key entirely")
	}
	if string(tmplCode) != `"sop"` {
		t.Fatalf("template route profile_code raw = %s, want %q (profile-keyed since ADR 0086)", tmplCode, `"sop"`)
	}
}

// due_in_days survives the wire round-trip. Nil means NO deadline — never zero,
// never a default (the no-fallback principle): the SLA engine leaves due_at NULL
// and the surfacer skips the stage entirely.
func TestMapStageRequests_CarriesDueInDays(t *testing.T) {
	days := 30
	in := []contracts.StageRequest{
		{Order: 1, Name: "Qualidade", RequiredCapability: "document.signoff", Quorum: "all", DriftPolicy: "recompute", DueInDays: &days},
		{Order: 2, Name: "Direção", RequiredCapability: "document.signoff", Quorum: "all", DriftPolicy: "recompute"},
	}

	got := mapStageRequests(in)

	if got[0].DueInDays == nil || *got[0].DueInDays != 30 {
		t.Fatalf("stage 1 DueInDays = %v, want 30", got[0].DueInDays)
	}
	if got[1].DueInDays != nil {
		t.Fatalf("stage 2 DueInDays = %v, want nil (absent means no deadline)", got[1].DueInDays)
	}
}

func TestMapStagesToResponse_CarriesDueInDays(t *testing.T) {
	days := 30
	out := mapStagesToResponse([]domain.Stage{
		{Order: 1, Name: "Qualidade", RequiredCapability: "document.signoff", DueInDays: &days},
		{Order: 2, Name: "Direção", RequiredCapability: "document.signoff"},
	})

	if out[0].DueInDays == nil || *out[0].DueInDays != 30 {
		t.Fatalf("stage 1 DueInDays = %v, want 30", out[0].DueInDays)
	}
	if out[1].DueInDays != nil {
		t.Fatalf("stage 2 DueInDays = %v, want nil", out[1].DueInDays)
	}
}

// mapListRoute is the mapper feeding ListStageItem, whose wire field
// (StageSummary.due_in_days) is nullable-and-required — a reviewer finding on
// Task 2 flagged that this carry-through had no direct test. Mirrors
// TestMapStagesToResponse_CarriesDueInDays for the list-route path.
func TestMapListRoute_CarriesDueInDays(t *testing.T) {
	days := 30
	route := infrastructure.Route{
		ID:          "route-1",
		Name:        "Qualidade",
		TenantID:    "tenant-1",
		ProfileCode: "ops",
		SubjectKind: "document",
		SubjectKey:  "ops",
		Version:     1,
		Stages: []infrastructure.RouteStage{
			{Order: 1, Name: "Qualidade", RequiredCapability: "document.signoff", Quorum: "all_of", DriftPolicy: "reduce_quorum", DueInDays: &days},
			{Order: 2, Name: "Direção", RequiredCapability: "document.signoff", Quorum: "all_of", DriftPolicy: "reduce_quorum"},
		},
	}

	out := mapListRoute(route)

	if out.Stages[0].DueInDays == nil || *out.Stages[0].DueInDays != 30 {
		t.Fatalf("stage 1 DueInDays = %v, want 30", out.Stages[0].DueInDays)
	}
	if out.Stages[1].DueInDays != nil {
		t.Fatalf("stage 2 DueInDays = %v, want nil (absent means no deadline)", out.Stages[1].DueInDays)
	}
}
