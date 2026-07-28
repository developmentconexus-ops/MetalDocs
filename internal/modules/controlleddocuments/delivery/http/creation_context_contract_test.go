package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	"metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// creationContextRequest builds an authenticated GET for the creation-context
// endpoint. Tier-1 (permissions.go) already gated the route on
// controlled_documents.create before the handler runs.
func creationContextRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/creation-context", nil)
	ctx := tenant.WithTenantID(req.Context(), "test-tenant")
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{"author"})
	return req.WithContext(ctx)
}

// TestAtomicCreate_ApprovalRouteMissing_Returns409 is the D2 gate's wire
// contract: creation into a route-less profile must emit the SAME
// problem+json code the submit path emits (state.approval_route_missing), so
// the client has one code to handle for one condition.
func TestAtomicCreate_ApprovalRouteMissing_Returns409(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{
		createErr: controlleddocumentsdomain.ErrApprovalRouteMissing,
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(`{
		"document_name":"Policy v1",
		"visibility":{"scope":"company","area_codes":[],"user_ids":[]},
		"profile_code":"DC",
		"process_area_code":"RH",
		"title":"Policy",
		"owner_user_id":"user-1"
	}`))
	ctx := tenant.WithTenantID(req.Context(), "test-tenant")
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	var problemBody struct {
		Status int    `json:"status"`
		Code   string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &problemBody); err != nil {
		t.Fatalf("decode problem body: %v (body=%s)", err, rec.Body.String())
	}
	if problemBody.Code != "state.approval_route_missing" {
		t.Fatalf("code = %q, want state.approval_route_missing", problemBody.Code)
	}
	if problemBody.Status != http.StatusConflict {
		t.Fatalf("problem.status = %d, want 409", problemBody.Status)
	}
}

// TestGetControlledDocumentCreationContext_ReturnsFlatBody pins the ADR 0035
// flat-body shape and the has_active_route annotation the create form needs to
// disable a profile that would otherwise 409.
func TestGetControlledDocumentCreationContext_ReturnsFlatBody(t *testing.T) {
	spy := &spyControlledDocumentService{creationContextResult: &application.CreationContext{
		Profiles: []application.CreationContextProfile{
			{Code: "po", Name: "Procedimento", HasActiveRoute: true},
			{Code: "it", Name: "Instrucao", HasActiveRoute: false},
		},
		Areas: []application.CreationContextArea{
			{Code: "quality", Name: "Qualidade"},
		},
	}}
	handler := &Handler{svc: spy}
	rec := httptest.NewRecorder()

	handler.GetControlledDocumentCreationContext(rec, creationContextRequest())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotCreationContextTenantID != "test-tenant" {
		t.Errorf("tenant forwarded = %q, want test-tenant", spy.gotCreationContextTenantID)
	}

	var body struct {
		Profiles []struct {
			Code           string `json:"code"`
			Name           string `json:"name"`
			HasActiveRoute *bool  `json:"has_active_route"`
		} `json:"profiles"`
		Areas []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"areas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	if len(body.Profiles) != 2 || len(body.Areas) != 1 {
		t.Fatalf("body = %s, want 2 profiles and 1 area", rec.Body.String())
	}
	for _, p := range body.Profiles {
		if p.HasActiveRoute == nil {
			t.Fatalf("profile %q missing required has_active_route", p.Code)
		}
	}
	if body.Profiles[0].Code != "po" || !*body.Profiles[0].HasActiveRoute {
		t.Errorf("profiles[0] = %+v, want po/true", body.Profiles[0])
	}
	if body.Profiles[1].Code != "it" || *body.Profiles[1].HasActiveRoute {
		t.Errorf("profiles[1] = %+v, want it/false", body.Profiles[1])
	}
	if body.Areas[0].Code != "quality" || body.Areas[0].Name != "Qualidade" {
		t.Errorf("areas[0] = %+v, want quality/Qualidade", body.Areas[0])
	}
}

// TestGetControlledDocumentCreationContext_EmptyCollectionsAreArrays keeps the
// contract's required arrays non-null when the caller may create nowhere —
// a null would break typed clients that iterate unconditionally.
func TestGetControlledDocumentCreationContext_EmptyCollectionsAreArrays(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{
		creationContextResult: &application.CreationContext{},
	}}
	rec := httptest.NewRecorder()

	handler.GetControlledDocumentCreationContext(rec, creationContextRequest())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	for _, field := range []string{"profiles", "areas"} {
		val, ok := raw[field]
		if !ok {
			t.Fatalf("required field %q absent from %s", field, rec.Body.String())
		}
		if string(val) != "[]" {
			t.Errorf("%s = %s, want []", field, string(val))
		}
	}
}

// TestGetControlledDocumentCreationContext_Unconfigured_Returns500 proves the
// composition-root fail-closed path surfaces as a server error, never as an
// empty (silently permissive-looking) 200.
func TestGetControlledDocumentCreationContext_Unconfigured_Returns500(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{
		creationContextErr: application.ErrCreationContextUnconfigured,
	}}
	rec := httptest.NewRecorder()

	handler.GetControlledDocumentCreationContext(rec, creationContextRequest())

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body = %s", rec.Code, rec.Body.String())
	}
	var problemBody struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &problemBody); err != nil {
		t.Fatalf("decode problem body: %v (body=%s)", err, rec.Body.String())
	}
	if problemBody.Code != "creation_context.unconfigured" {
		t.Fatalf("code = %q, want creation_context.unconfigured", problemBody.Code)
	}
}

// TestGetControlledDocumentCreationContext_MissingActor_Returns401 keeps the
// endpoint useless without an authenticated principal: its body is a
// per-caller authorization view.
func TestGetControlledDocumentCreationContext_MissingActor_Returns401(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{
		creationContextErr: application.ErrActorMissing,
	}}
	rec := httptest.NewRecorder()

	handler.GetControlledDocumentCreationContext(rec, creationContextRequest())

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", rec.Code, rec.Body.String())
	}
}
