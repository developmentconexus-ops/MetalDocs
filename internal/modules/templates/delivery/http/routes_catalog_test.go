package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)

func TestPlaceholderCatalog_Returns7Entries(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)
	req := httptest.NewRequest("GET", "/api/v1/templates/placeholder-catalog", nil)
	withHeaders(req)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		Items []struct {
			Key         string `json:"key"`
			Label       string `json:"label"`
			Description string `json:"description"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 7 {
		t.Fatalf("items len = %d, want 7", len(body.Items))
	}
	wantKeys := []string{"doc_code", "doc_title", "revision_number", "author", "effective_date", "approvers", "controlled_by_area"}
	for i, k := range wantKeys {
		if body.Items[i].Key != k {
			t.Errorf("items[%d].Key = %q, want %q", i, body.Items[i].Key, k)
		}
		if body.Items[i].Label == "" {
			t.Errorf("items[%d].Label empty", i)
		}
	}
}

func TestPlaceholderCatalog_RequiresTemplateViewAuthz(t *testing.T) {
	repo := newFakeRepo()
	var gotAction string
	mux := newMux(t, func(_ *http.Request, _, _, action string) error {
		gotAction = action
		return domain.ErrForbidden
	}, repo)
	req := httptest.NewRequest("GET", "/api/v1/templates/placeholder-catalog", nil)
	withHeaders(req)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 body=%s", rec.Code, rec.Body.String())
	}
	if gotAction != "template.view" {
		t.Fatalf("authz action = %q, want template.view", gotAction)
	}
}
