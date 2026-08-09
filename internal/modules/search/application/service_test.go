package application

import (
	"context"
	"errors"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/pagination"
)

type stubReader struct {
	called     bool
	lastQuery  domain.Query
	lastLimit  int
	lastOffset int
	docs       []domain.Document
	listErr    error
}

func (r *stubReader) ListDocuments(_ context.Context, query domain.Query, limit, offset int) ([]domain.Document, error) {
	r.called = true
	r.lastQuery = query
	r.lastLimit = limit
	r.lastOffset = offset
	return r.docs, r.listErr
}

func authedCtx(userID string) context.Context {
	return iamdomain.WithAuthContext(context.Background(), userID, []iamdomain.Role{iamdomain.RoleViewer})
}

func TestSearchDocumentsRequiresTenantID(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)

	_, err := svc.SearchDocuments(authedCtx("user-1"), domain.Query{})
	if !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("error = %v, want ErrTenantRequired", err)
	}
	if reader.called {
		t.Fatal("reader was called without tenant id")
	}
}

// TestSearchDocumentsForwardsActorAndTenantToReader guards REQ-SEARCH-1: the
// application layer makes no independent authz decision of its own — it
// forwards the actor and tenant straight to the reader, which enforces
// visibility in SQL alongside every other read path. See ADR 0095.
func TestSearchDocumentsForwardsActorAndTenantToReader(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)

	_, err := svc.SearchDocuments(authedCtx("user-1"), domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if !reader.called {
		t.Fatal("reader was not called")
	}
	if reader.lastQuery.TenantID != "tenant-1" {
		t.Fatalf("tenant id = %q, want tenant-1", reader.lastQuery.TenantID)
	}
	if reader.lastQuery.ActorUserID != "user-1" {
		t.Fatalf("actor = %q, want user-1 forwarded for data-layer visibility", reader.lastQuery.ActorUserID)
	}
	if reader.lastLimit != pagination.DefaultLimit {
		t.Fatalf("limit = %d, want %d", reader.lastLimit, pagination.DefaultLimit)
	}
	if reader.lastOffset != 0 {
		t.Fatalf("offset = %d, want 0 (visibility filtered in SQL, no paging loop)", reader.lastOffset)
	}
}

func TestSearchDocumentsCapsLimit(t *testing.T) {
	reader := &stubReader{}
	svc := NewService(reader)

	_, err := svc.SearchDocuments(authedCtx("user-1"), domain.Query{TenantID: "tenant-1", Limit: pagination.MaxLimit + 1})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if reader.lastLimit != pagination.MaxLimit {
		t.Fatalf("limit = %d, want %d", reader.lastLimit, pagination.MaxLimit)
	}
}

func TestSearchDocumentsReturnsReaderResults(t *testing.T) {
	reader := &stubReader{docs: []domain.Document{
		{ID: "doc-1", Title: "Granted", CreatedAt: time.Unix(10, 0)},
	}}
	svc := NewService(reader)

	got, err := svc.SearchDocuments(authedCtx("user-1"), domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 1 || got[0].ID != "doc-1" {
		t.Fatalf("results = %#v, want the single visible doc the reader returned", got)
	}
}

// TestSearchDocumentsWithoutAuthContextDeniesResults: an unauthenticated caller
// must see nothing, and the reader must not be consulted (no actor to scope by).
func TestSearchDocumentsWithoutAuthContextDeniesResults(t *testing.T) {
	reader := &stubReader{docs: []domain.Document{{ID: "doc-1", Title: "Manual", CreatedAt: time.Now()}}}
	svc := NewService(reader)

	got, err := svc.SearchDocuments(context.Background(), domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("results = %d, want 0 for unauthenticated caller", len(got))
	}
	if reader.called {
		t.Fatal("reader must not be called without an authenticated actor")
	}
}
