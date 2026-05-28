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
	listOffsets  []int
	docs         []domain.Document
	docsByOffset map[int][]domain.Document
	policies     map[string][]domain.AccessPolicy
	policyCalls  []string
}

func (r *stubReader) ListDocuments(_ context.Context, query domain.Query, limit, offset int) ([]domain.Document, error) {
	r.called = true
	r.listTenantID = query.TenantID
	r.listLimit = limit
	r.listOffsets = append(r.listOffsets, offset)
	if r.docsByOffset != nil {
		return r.docsByOffset[offset], nil
	}
	return r.docs, nil
}

func (r *stubReader) ListAccessPolicies(_ context.Context, scope, id string) ([]domain.AccessPolicy, error) {
	r.policyCalls = append(r.policyCalls, scope+":"+id)
	if r.policies != nil {
		return r.policies[scope+":"+id], nil
	}
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

func TestSearchDocumentsPagesUntilAuthorizedMatchesFillLimit(t *testing.T) {
	reader := &stubReader{
		docsByOffset: map[int][]domain.Document{
			0: {
				{ID: "doc-1", Title: "Draft 1", CreatedAt: time.Unix(30, 0)},
				{ID: "doc-2", Title: "Draft 2", CreatedAt: time.Unix(20, 0)},
			},
			2: {
				{ID: "doc-3", Title: "Allowed", CreatedAt: time.Unix(10, 0)},
			},
		},
		policies: map[string][]domain.AccessPolicy{
			"document:doc-1": {{
				SubjectType: domain.SubjectTypeUser,
				SubjectID:   "user-1",
				Capability:  searchCapabilityView,
				Effect:      domain.EffectDeny,
			}},
			"document:doc-2": {{
				SubjectType: domain.SubjectTypeUser,
				SubjectID:   "user-1",
				Capability:  searchCapabilityView,
				Effect:      domain.EffectDeny,
			}},
		},
	}
	svc := NewService(reader)
	ctx := iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})

	got, err := svc.SearchDocuments(ctx, domain.Query{TenantID: "tenant-1", Limit: 1})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 1 || got[0].ID != "doc-3" {
		t.Fatalf("results = %#v, want only doc-3", got)
	}
	if len(reader.listOffsets) != 2 || reader.listOffsets[0] != 0 || reader.listOffsets[1] != 2 {
		t.Fatalf("offsets = %#v, want [0 2]", reader.listOffsets)
	}
}

func TestSearchDocuments_DeniesByAreaPolicyUsingBusinessUnitAndDepartment(t *testing.T) {
	reader := &stubReader{
		docs: []domain.Document{{
			ID:           "doc-1",
			Title:        "Controlled Manual",
			BusinessUnit: "ops",
			Department:   "qa",
			CreatedAt:    time.Now(),
		}},
		policies: map[string][]domain.AccessPolicy{
			"area:ops:qa": {{
				SubjectType: domain.SubjectTypeUser,
				SubjectID:   "user-1",
				Capability:  searchCapabilityView,
				Effect:      domain.EffectDeny,
			}},
		},
	}
	svc := NewService(reader)
	ctx := iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})

	got, err := svc.SearchDocuments(ctx, domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("results = %#v, want area-policy deny", got)
	}
}

func TestSearchDocuments_DeniesByAreaPolicyUsingProcessAreaFallbackWhenBusinessUnitMissing(t *testing.T) {
	reader := &stubReader{
		docs: []domain.Document{{
			ID:          "doc-1",
			Title:       "Controlled Manual",
			ProcessArea: "ops",
			Department:  "qa",
			CreatedAt:   time.Now(),
		}},
		policies: map[string][]domain.AccessPolicy{
			"area:ops:qa": {{
				SubjectType: domain.SubjectTypeUser,
				SubjectID:   "user-1",
				Capability:  searchCapabilityView,
				Effect:      domain.EffectDeny,
			}},
		},
	}
	svc := NewService(reader)
	ctx := iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{iamdomain.RoleViewer})

	got, err := svc.SearchDocuments(ctx, domain.Query{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("SearchDocuments: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("results = %#v, want area-policy deny", got)
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
