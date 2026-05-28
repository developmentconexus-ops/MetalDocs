package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	searchapp "metaldocs/internal/modules/search/application"
	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/tenant"
)

type handlerStubReader struct {
	called       bool
	listTenantID string
	lastQuery    searchdomain.Query
	documents    []searchdomain.Document
}

func (r *handlerStubReader) ListDocuments(_ context.Context, query searchdomain.Query, _, _ int) ([]searchdomain.Document, error) {
	r.called = true
	r.listTenantID = query.TenantID
	r.lastQuery = query
	return append([]searchdomain.Document(nil), r.documents...), nil
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

func TestHandleSearchDocumentsMapsAdvertisedFilters(t *testing.T) {
	reader := &handlerStubReader{}
	h := NewHandler(searchapp.NewService(reader))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents?q=manual&documentType=po&documentProfile=qa-doc&documentFamily=sop&processArea=quality&ownerId=user-42&department=qa&status=ARCHIVED&expiryBefore=2026-12-01T00:00:00Z&expiryAfter=2026-01-01T00:00:00Z&limit=7&subject=deviation&businessUnit=ops&classification=INTERNAL&tag=tag-a", nil)
	ctx := iamdomain.WithAuthContext(req.Context(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})
	ctx = tenant.WithTenantID(ctx, "tenant-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.handleSearchDocuments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if reader.lastQuery.Text != "manual" {
		t.Fatalf("query.Text = %q, want manual", reader.lastQuery.Text)
	}
	if reader.lastQuery.DocumentType != "po" {
		t.Fatalf("query.DocumentType = %q, want po", reader.lastQuery.DocumentType)
	}
	if reader.lastQuery.DocumentProfile != "qa-doc" {
		t.Fatalf("query.DocumentProfile = %q, want qa-doc", reader.lastQuery.DocumentProfile)
	}
	if reader.lastQuery.DocumentFamily != "sop" {
		t.Fatalf("query.DocumentFamily = %q, want sop", reader.lastQuery.DocumentFamily)
	}
	if reader.lastQuery.ProcessArea != "quality" {
		t.Fatalf("query.ProcessArea = %q, want quality", reader.lastQuery.ProcessArea)
	}
	if reader.lastQuery.OwnerID != "user-42" {
		t.Fatalf("query.OwnerID = %q, want user-42", reader.lastQuery.OwnerID)
	}
	if reader.lastQuery.Department != "qa" {
		t.Fatalf("query.Department = %q, want qa", reader.lastQuery.Department)
	}
	if reader.lastQuery.Status != searchdomain.Status("ARCHIVED") {
		t.Fatalf("query.Status = %q, want ARCHIVED", reader.lastQuery.Status)
	}
	if reader.lastQuery.Limit != 7 {
		t.Fatalf("query.Limit = %d, want 7", reader.lastQuery.Limit)
	}
	if reader.lastQuery.ExpiryBefore == nil || reader.lastQuery.ExpiryBefore.UTC().Format(time.RFC3339) != "2026-12-01T00:00:00Z" {
		t.Fatalf("query.ExpiryBefore = %v, want 2026-12-01T00:00:00Z", reader.lastQuery.ExpiryBefore)
	}
	if reader.lastQuery.ExpiryAfter == nil || reader.lastQuery.ExpiryAfter.UTC().Format(time.RFC3339) != "2026-01-01T00:00:00Z" {
		t.Fatalf("query.ExpiryAfter = %v, want 2026-01-01T00:00:00Z", reader.lastQuery.ExpiryAfter)
	}
	if reader.lastQuery.Subject != "deviation" {
		t.Fatalf("query.Subject = %q, want deviation", reader.lastQuery.Subject)
	}
	if reader.lastQuery.BusinessUnit != "ops" {
		t.Fatalf("query.BusinessUnit = %q, want ops", reader.lastQuery.BusinessUnit)
	}
	if reader.lastQuery.Classification != searchdomain.ClassificationInternal {
		t.Fatalf("query.Classification = %q, want INTERNAL", reader.lastQuery.Classification)
	}
	if reader.lastQuery.Tag != "tag-a" {
		t.Fatalf("query.Tag = %q, want tag-a", reader.lastQuery.Tag)
	}
}

func TestHandleSearchDocumentsResponseIncludesAdvertisedFields(t *testing.T) {
	now := time.Date(2026, 5, 26, 15, 4, 5, 0, time.UTC)
	reader := &handlerStubReader{
		documents: []searchdomain.Document{
			{
				ID:               "doc-1",
				Title:            "Document title",
				DocumentType:     "po",
				DocumentProfile:  "po",
				DocumentFamily:   "sop",
				DocumentSequence: 12,
				DocumentCode:     "PO-QA-012",
				ProcessArea:      "quality",
				Subject:          "legacy-subject",
				OwnerID:          "user-1",
				BusinessUnit:     "legacy-bu",
				Department:       "qa",
				Classification:   searchdomain.ClassificationInternal,
				Status:           searchdomain.StatusArchived,
				Tags:             []string{"legacy-tag"},
				CreatedAt:        now,
			},
		},
	}
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
	var payload map[string][]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	items := payload["items"]
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	item := items[0]
	if got := item["subject"]; got != "legacy-subject" {
		t.Fatalf("subject = %#v, want legacy-subject", got)
	}
	if got := item["businessUnit"]; got != "legacy-bu" {
		t.Fatalf("businessUnit = %#v, want legacy-bu", got)
	}
	if got := item["classification"]; got != "INTERNAL" {
		t.Fatalf("classification = %#v, want INTERNAL", got)
	}
	tags, ok := item["tags"].([]any)
	if !ok {
		t.Fatalf("tags type = %T, want []any", item["tags"])
	}
	if len(tags) != 1 || tags[0] != "legacy-tag" {
		t.Fatalf("tags = %#v, want [legacy-tag]", tags)
	}
}
