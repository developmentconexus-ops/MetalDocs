package domain_test

import (
	"testing"

	"metaldocs/internal/modules/render/domain"
)

func TestComputedCatalog_HasAllEightTokens(t *testing.T) {
	got := domain.ComputedCatalog()
	if len(got) != 8 {
		t.Fatalf("ComputedCatalog len = %d, want 8", len(got))
	}
	wantKeys := []string{
		"doc_code", "doc_title", "revision_number", "author",
		"effective_date", "approvers", "controlled_by_area", "approval_date",
	}
	byKey := make(map[string]domain.ComputedToken, len(got))
	for _, e := range got {
		if e.Key == "" || e.Label == "" || e.Description == "" {
			t.Errorf("token %q has empty key/label/description", e.Key)
		}
		byKey[e.Key] = e
	}
	for _, k := range wantKeys {
		if _, ok := byKey[k]; !ok {
			t.Errorf("missing computed token %q", k)
		}
	}
}

func TestComputedCatalog_ApprovalDateIsAuthorVisible(t *testing.T) {
	for _, e := range domain.ComputedCatalog() {
		if e.Key == "approval_date" && !e.AuthorVisible {
			t.Fatal("approval_date must be AuthorVisible (operator decision 2026-06-29)")
		}
	}
}

func TestComputedCatalog_AllCurrentlyAuthorVisible(t *testing.T) {
	for _, e := range domain.ComputedCatalog() {
		if !e.AuthorVisible {
			t.Errorf("token %q unexpectedly not AuthorVisible", e.Key)
		}
	}
}
