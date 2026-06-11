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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents?q=manual&document_type=po&document_profile=qa-doc&document_family=sop&process_area=quality&owner_id=user-42&department=qa&status=ARCHIVED&expiry_before=2026-12-01T00:00:00Z&expiry_after=2026-01-01T00:00:00Z&limit=7&subject=deviation&classification=INTERNAL&tag=tag-a", nil)
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
	// businessUnit (camelCase, undocumented) was removed in Wave 1.5 (F-13b).
	if reader.lastQuery.BusinessUnit != "" {
		t.Fatalf("query.BusinessUnit = %q, want empty (parameter removed)", reader.lastQuery.BusinessUnit)
	}
	if reader.lastQuery.Classification != searchdomain.ClassificationInternal {
		t.Fatalf("query.Classification = %q, want INTERNAL", reader.lastQuery.Classification)
	}
	if reader.lastQuery.Tag != "tag-a" {
		t.Fatalf("query.Tag = %q, want tag-a", reader.lastQuery.Tag)
	}
}

func TestHandleSearchDocumentsMethodNotAllowedIsProblemJSON(t *testing.T) {
	reader := &handlerStubReader{}
	h := NewHandler(searchapp.NewService(reader))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/documents", nil)
	rec := httptest.NewRecorder()

	h.handleSearchDocuments(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", allow)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("code = %q, want METHOD_NOT_ALLOWED", body.Code)
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
	// F-13a (Wave 1.5): the legacy fields are gone from the wire format even
	// when the domain model carries values — the spec never declared them and
	// the SQL reader never populates them in production.
	for _, removed := range []string{"subject", "business_unit", "classification", "tags"} {
		if _, present := item[removed]; present {
			t.Fatalf("removed legacy field %q still present in response: %#v", removed, item)
		}
	}
	if got := item["document_code"]; got != "PO-QA-012" {
		t.Fatalf("document_code = %#v, want PO-QA-012", got)
	}
	if got := item["process_area"]; got != "quality" {
		t.Fatalf("process_area = %#v, want quality", got)
	}
	if got := item["status"]; got != "ARCHIVED" {
		t.Fatalf("status = %#v, want ARCHIVED", got)
	}
}
