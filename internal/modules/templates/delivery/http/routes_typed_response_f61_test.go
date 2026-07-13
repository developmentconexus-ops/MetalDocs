package http_test

// Feature F6.1 — typed-shape contract tests for the templates lifecycle 200 responses.
// Each subtest decodes the handler body with json.Decoder + DisallowUnknownFields() into
// the codegen-generated *Response type. Asserts the typed field round-trips non-zero —
// locking the handlers to the OpenAPI declaration so future drift fails at this gate.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/domain"
)

func decodeStrict[T any](t *testing.T, body []byte) T {
	t.Helper()
	var out T
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode strict: %v\nbody: %s", err, string(body))
	}
	return out
}

func TestLifecycle_TypedResponseShape(t *testing.T) {
	const (
		tplID    = "11111111-1111-1111-1111-111111111111"
		tenantID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	)

	t.Run("archive", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/archive", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ArchiveTemplateResponse](t, rr.Body.Bytes())
		if out.Data.Template.ArchivedAt == nil || out.Data.Template.ArchivedAt.IsZero() {
			t.Fatal("template.archived_at not set")
		}
	})
}
