package application_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
)

func TestService_ListDocumentsPaginated_AdminBypassesUserFilter(t *testing.T) {
	repo := &fakeRepo{
		listPaginatedReturn: []*domain.Document{{ID: "doc_1"}, {ID: "doc_2"}},
		countReturn:         2,
	}
	svc := application.New(repo, nil, nil, nil, nil)

	items, total, err := svc.ListDocumentsPaginated(context.Background(), "tenant_1", "", application.ListOptions{CreatedBy: "client_value", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated() error = %v", err)
	}
	if len(items) != 2 || total != 2 {
		t.Fatalf("unexpected result len=%d total=%d", len(items), total)
	}
	if repo.lastOpts.CreatedBy != "client_value" {
		t.Fatalf("admin call should not force CreatedBy; got %q", repo.lastOpts.CreatedBy)
	}
}

func TestService_ListDocumentsPaginated_FillerForcesUserScope(t *testing.T) {
	repo := &fakeRepo{
		listPaginatedReturn: []*domain.Document{{ID: "doc_1"}},
		countReturn:         1,
	}
	svc := application.New(repo, nil, nil, nil, nil)

	_, _, err := svc.ListDocumentsPaginated(context.Background(), "tenant_1", "u1", application.ListOptions{CreatedBy: "ignored", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated() error = %v", err)
	}
	if repo.lastOpts.CreatedBy != "u1" {
		t.Fatalf("expected forced CreatedBy u1, got %q", repo.lastOpts.CreatedBy)
	}
}

func TestService_Stats_ReturnsMaps(t *testing.T) {
	repo := &fakeRepo{
		statsByStatusReturn: map[string]int64{"draft": 5},
		statsByAreaReturn:   map[string]int64{"RH": 3},
	}
	svc := application.New(repo, nil, nil, nil, nil)

	stats, err := svc.DocumentStats(context.Background(), "tenant_1", "", application.ListOptions{})
	if err != nil {
		t.Fatalf("DocumentStats() error = %v", err)
	}
	if stats.ByStatus["draft"] != 5 {
		t.Fatalf("expected byStatus[draft]=5, got %d", stats.ByStatus["draft"])
	}
	if stats.ByArea["RH"] != 3 {
		t.Fatalf("expected byArea[RH]=3, got %d", stats.ByArea["RH"])
	}
}
