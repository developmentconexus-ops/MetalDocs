// tenant_lifecycle_handler_test.go — M7 F7.3 Task E's named unit-test
// deliverable: "enqueue error mapping (fakes)". Exercises
// TenantHandler.handleExportTenant/handleEraseTenant's status-code mapping
// via writeLifecycleError, using a fake TenantLifecycler — no DB, no River.
package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// codeStr unwraps a problem.Code to a plain string for comparison against a
// JSON-decoded (map[string]any) response body field.
func codeStr(c problem.Code) string { return c.String() }

// fakeTenantLifecycler is a test double for TenantLifecycler letting each
// test case control the exact error RequestExport/RequestErase return.
type fakeTenantLifecycler struct {
	jobID string
	err   error
}

func (f *fakeTenantLifecycler) RequestExport(ctx context.Context, tenantID, actorID, actorTenantID string) (string, error) {
	return f.jobID, f.err
}

func (f *fakeTenantLifecycler) RequestErase(ctx context.Context, tenantID, actorID, actorTenantID string) (string, error) {
	return f.jobID, f.err
}

func newLifecycleRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/tenants/11111111-1111-1111-1111-111111111111/export", nil)
	ctx := tenant.WithTenantID(req.Context(), "actor-tenant")
	return req.WithContext(ctx)
}

func TestHandleExportTenant_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		lcErr      error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "tenant not found maps to 404",
			lcErr:      iamdomain.ErrTenantNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   codeStr(problem.CodeNotFoundResource),
		},
		{
			name:       "already erased maps to 409",
			lcErr:      iamdomain.ErrTenantAlreadyErased,
			wantStatus: http.StatusConflict,
			wantCode:   codeStr(problem.CodeConflictGeneric),
		},
		{
			name:       "validation error maps to 400",
			lcErr:      iamdomain.ErrTenantOnboardValidation,
			wantStatus: http.StatusBadRequest,
			wantCode:   codeStr(problem.CodeRequestInvalid),
		},
		{
			name:       "capability denied maps to 403",
			lcErr:      authz.ErrCapDenied{Capability: "tenant.export", AreaCode: "tenant", ActorID: "u1"},
			wantStatus: http.StatusForbidden,
			wantCode:   codeStr(problem.CodePermissionCapabilityDenied),
		},
		{
			name:       "wrapped capability denied still maps to 403",
			lcErr:      errWrap(authz.ErrCapDenied{Capability: "tenant.export", AreaCode: "tenant", ActorID: "u1"}),
			wantStatus: http.StatusForbidden,
			wantCode:   codeStr(problem.CodePermissionCapabilityDenied),
		},
		{
			name:       "unknown error maps to 500",
			lcErr:      errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   codeStr(problem.CodeInternalUnknown),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewTenantHandler(nil).WithLifecycle(&fakeTenantLifecycler{err: tc.lcErr})
			rec := httptest.NewRecorder()

			handler.handleExportTenant(rec, newLifecycleRequest(t), "11111111-1111-1111-1111-111111111111")

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response body not valid JSON: %v (%s)", err, rec.Body.String())
			}
			if got := body["code"]; got != tc.wantCode {
				t.Errorf("problem code = %v, want %v", got, tc.wantCode)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json (RFC 9457)", ct)
			}
		})
	}
}

func TestHandleEraseTenant_SuccessReturns202WithJobID(t *testing.T) {
	handler := NewTenantHandler(nil).WithLifecycle(&fakeTenantLifecycler{jobID: "22222222-2222-2222-2222-222222222222"})
	rec := httptest.NewRecorder()

	handler.handleEraseTenant(rec, newLifecycleRequest(t), "11111111-1111-1111-1111-111111111111")

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (body: %s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body not valid JSON: %v", err)
	}
	if body["job_id"] != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("job_id = %v, want 22222222-2222-2222-2222-222222222222", body["job_id"])
	}
}

func TestHandleExportTenant_NoLifecycleWiredReturns501(t *testing.T) {
	handler := NewTenantHandler(nil) // lifecycle left nil (WithLifecycle never called)
	rec := httptest.NewRecorder()

	handler.handleExportTenant(rec, newLifecycleRequest(t), "11111111-1111-1111-1111-111111111111")

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501 when lifecycle service is not configured", rec.Code)
	}
}

// errWrap wraps err so errors.Is/errors.As must unwrap to find it, exercising
// writeLifecycleError's use of errors.As (not a bare type assertion).
func errWrap(err error) error {
	return &wrappedErr{inner: err}
}

type wrappedErr struct{ inner error }

func (w *wrappedErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrappedErr) Unwrap() error { return w.inner }
