package httpdelivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	searchapp "metaldocs/internal/modules/search/application"
	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/tenant"
)

type handlerStubReader struct {
	called       bool
	listTenantID string
}

func (r *handlerStubReader) ListDocuments(_ context.Context, tenantID string, _ int) ([]searchdomain.Document, error) {
	r.called = true
	r.listTenantID = tenantID
	return nil, nil
}

func (r *handlerStubReader) ListAccessPolicies(_ context.Context, _, _ string) ([]searchdomain.AccessPolicy, error) {
	return nil, nil
}

func TestHandleSearchDocumentsRequiresAuthentication(t *testing.T) {
	reader := &handlerStubReader{}
	h := NewHandler(searchapp.NewService(reader))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents", nil)
	rec := httptest.NewRecorder()

	h.handleSearchDocuments(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if reader.called {
		t.Fatal("reader was called for unauthenticated request")
	}
}

func TestHandleSearchDocumentsRequiresTenantContext(t *testing.T) {
	reader := &handlerStubReader{}
	h := NewHandler(searchapp.NewService(reader))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents", nil)
	ctx := iamdomain.WithAuthContext(req.Context(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.handleSearchDocuments(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if reader.called {
		t.Fatal("reader was called without tenant context")
	}
}

func TestHandleSearchDocumentsPassesTenantContext(t *testing.T) {
	reader := &handlerStubReader{}
	h := NewHandler(searchapp.NewService(reader))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents", nil)
	ctx := iamdomain.WithAuthContext(req.Context(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})
	ctx = tenant.WithTenantID(ctx, "tenant-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.handleSearchDocuments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if reader.listTenantID != "tenant-1" {
		t.Fatalf("tenant id = %q, want tenant-1", reader.listTenantID)
	}
}
