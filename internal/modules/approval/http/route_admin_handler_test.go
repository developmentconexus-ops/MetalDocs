package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			req.Header.Set("Idempotency-Key", "idem-1")

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
			req.Header.Set("Idempotency-Key", "idem-1")

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
			req.Header.Set("Idempotency-Key", "idem-1")

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
	req.Header.Set("Idempotency-Key", "idem-1")

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
	req.Header.Set("Idempotency-Key", "idem-1")

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
// The fixture uses a TEMPLATE subject, not a document one: per rail R1,
// profile_code is the backward-compat alias for (document, profile_code), so
// a document-subject route's subject_key must equal its profile_code
// (application.ErrDocumentSubjectKeyMismatch — M3 P3.S2b-2 regression fix).
// A document+divergent-key body is no longer a legal shape to prove
// pass-through with; a template subject's key has no such relationship to
// profile_code, so it still exercises the same decode/pass-through path
// without encoding an illegal state.
//
// The fixture omits profile_code entirely (S1, F18 repair): under the
// conditional contract rule (contracts.CreateRouteRequest.Validate, ADR 0082 /
// migration 0297), a template subject MUST NOT carry a profile_code — a body
// combining subject_kind=template with a non-empty profile_code is now a
// validation-error 400, not a legal pass-through fixture (previously this
// test asserted 201 through a mock that never reaches the DB check —
// false-green; see TestCreateRoute_TemplateSubjectRejectsProfileCode below
// for the negative case at the handler layer).
func TestCreateRoute_SubjectFieldsPassedThrough(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-123"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"name":"Ops Route","subject_kind":"template","subject_key":"tmpl-custom-1","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if svc.createReq.SubjectKind != "template" || svc.createReq.SubjectKey != "tmpl-custom-1" {
		t.Fatalf("subject fields not passed through: kind=%q key=%q", svc.createReq.SubjectKind, svc.createReq.SubjectKey)
	}
}

// TestCreateRoute_TemplateSubjectRejectsProfileCode pins the S1 (F18) contract
// fix at the handler layer: a create-route body carrying both
// subject_kind=template and a non-empty profile_code must be rejected as a
// 400 validation error before ever reaching the service (contracts layer
// Validate), never a 201 (the former false-green behavior asserted through a
// mock service that never touched the DB check).
func TestCreateRoute_TemplateSubjectRejectsProfileCode(t *testing.T) {
	svc := &fakeRouteAdminService{
		createResult: application.CreateRouteResult{RouteID: "route-should-not-be-created"},
	}
	h := &Handler{routeAdmin: svc}
	mux := routeAdminTestMux(h)

	body := `{"profile_code":"ops","name":"Ops Route","subject_kind":"template","subject_key":"tmpl-custom-1","stages":[{"order":1,"name":"Review","required_capability":"document.signoff","quorum":"any_1_of","drift_policy":"reduce_quorum","selectors":[{"kind":"role_in_fixed_area","role":"approver","area_code":"ops"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-2")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	if svc.createReq.SubjectKind != "" {
		t.Fatalf("service must not be invoked when validation fails, got createReq=%+v", svc.createReq)
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
			req.Header.Set("Idempotency-Key", "idem-3")

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
	req.Header.Set("Idempotency-Key", "idem-1")

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
			req.Header.Set("Idempotency-Key", "idem-1")
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
			req.Header.Set("Idempotency-Key", "idem-1")
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
			req.Header.Set("Idempotency-Key", "idem-1")
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
			if svc.deactivateReq.IdempotencyKey != "idem-1" {
				t.Fatalf("idempotency_key = %q, want %q", svc.deactivateReq.IdempotencyKey, "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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
	req.Header.Set("Idempotency-Key", "idem-1")
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

// TestListRoutes_ProfileCodeNullForTemplateRoutes pins F18 completion S6
// (hub ruling, contract-lock extended to RouteSummary): a template route has
// no profile by DB constraint (approval_routes_template_subject_projection_check,
// ADR 0082) and the list response must represent that truthfully as JSON
// null — not the "" sentinel. This decodes the raw response body into
// map[string]any (not the typed contracts.ListRouteItem) because a typed
// *string decode cannot distinguish an absent/null key from "": the
// assertion must inspect the wire bytes themselves. A sibling document route
// in the same response is asserted to keep serializing as a non-empty
// string, unaffected.
func TestListRoutes_ProfileCodeNullForTemplateRoutes(t *testing.T) {
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
				ID: "r2", Name: "Tmpl", TenantID: "tenant-1", ProfileCode: "",
				Active: true, Version: 1, Total: 2,
				SubjectKind: "template",
				SubjectKey:  "tmpl-1",
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
	if string(tmplCode) != "null" {
		t.Fatalf("template route profile_code raw = %s, want null (not the \"\" sentinel)", tmplCode)
	}
}
