package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	renderdomain "metaldocs/internal/modules/render/domain"
	"metaldocs/internal/modules/templates/domain"
)

func TestPlaceholderCatalog_MirrorsRenderDomainAuthorVisible(t *testing.T) {
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

	want := make(map[string]renderdomain.ComputedToken)
	for _, e := range renderdomain.ComputedCatalog() {
		if e.AuthorVisible {
			want[e.Key] = e
		}
	}
	if len(body.Items) != len(want) {
		t.Fatalf("items len = %d, want %d (render/domain author-visible set)", len(body.Items), len(want))
	}
	for _, it := range body.Items {
		w, ok := want[it.Key]
		if !ok {
			t.Errorf("endpoint returned %q not in render/domain author-visible set", it.Key)
			continue
		}
		if it.Label != w.Label || it.Description != w.Description {
			t.Errorf("token %q label/description mismatch with render/domain", it.Key)
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
