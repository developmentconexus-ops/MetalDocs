package application

import (
	"context"
	"errors"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/search/domain"
)

type stubReader struct {
	called       bool
	listTenantID string
	listLimit    int
	docs         []domain.Document
}

func (r *stubReader) ListDocuments(_ context.Context, query domain.Query, limit int) ([]domain.Document, error) {
	r.called = true
	r.listTenantID = query.TenantID
	r.listLimit = limit
	return r.docs, nil
}

func (r *stubReader) ListAccessPolicies(_ context.Context, _, _ string) ([]domain.AccessPolicy, error) {
	return nil, nil
}

func TestSearchDocumentsRequiresTenantID(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)

	_, err := svc.SearchDocuments(context.Background(), domain.Query{})
	if !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("error = %v, want ErrTenantRequired", err)
	}
	if reader.called {
		t.Fatal("reader was called without tenant id")
	}
}

func TestSearchDocumentsPassesTenantIDToReader(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)
	ctx := iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})

	_, err := svc.SearchDocuments(ctx, domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if reader.listTenantID != "tenant-1" {
		t.Fatalf("tenant id = %q, want tenant-1", reader.listTenantID)
	}
	if reader.listLimit != defaultLimit {
		t.Fatalf("limit = %d, want %d", reader.listLimit, defaultLimit)
	}
}

func TestSearchDocumentsPassesCappedLimitToReader(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)
	ctx := iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})

	_, err := svc.SearchDocuments(ctx, domain.Query{TenantID: "tenant-1", Limit: maxLimit + 1})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if reader.listLimit != maxLimit {
		t.Fatalf("limit = %d, want %d", reader.listLimit, maxLimit)
	}
}

func TestSearchDocumentsWithoutAuthContextDeniesResults(t *testing.T) {
	reader := &stubReader{docs: []domain.Document{{
		ID:        "doc-1",
		Title:     "Manual",
		CreatedAt: time.Now(),
	}}}
	svc := NewService(reader)

	got, err := svc.SearchDocuments(context.Background(), domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("results = %d, want 0", len(got))
	}
}
