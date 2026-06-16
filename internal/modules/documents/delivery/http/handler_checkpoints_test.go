package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/domain"
)

var checkpointDeclaredKeys = map[string]struct{}{
	"id":          {},
	"document_id": {},
	"revision_id": {},
	"version_num": {},
	"label":       {},
	"created_at":  {},
	"created_by":  {},
}

var checkpointPascalKeys = []string{
	"ID", "DocumentID", "RevisionID", "VersionNum", "Label", "CreatedAt", "CreatedBy",
}

func assertCheckpointShape(t *testing.T, raw map[string]json.RawMessage) {
	t.Helper()
	for k := range raw {
		if _, ok := checkpointDeclaredKeys[k]; !ok {
			t.Fatalf("unexpected key %q in checkpoint response (extra/PascalCase): keys=%v", k, mapKeys(raw))
		}
	}
	for k := range checkpointDeclaredKeys {
		if _, ok := raw[k]; !ok {
			t.Fatalf("missing required key %q in checkpoint response: keys=%v", k, mapKeys(raw))
		}
	}
	for _, pk := range checkpointPascalKeys {
		if _, ok := raw[pk]; ok {
			t.Fatalf("forbidden PascalCase key %q in checkpoint response", pk)
		}
	}
	var versionNum int
	if err := json.Unmarshal(raw["version_num"], &versionNum); err != nil {
		t.Fatalf("version_num must decode to integer: %v (raw=%s)", err, raw["version_num"])
	}
}

func mapKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestListCheckpoints_TypedResponseShape(t *testing.T) {
	t.Run("Items", func(t *testing.T) {
		svc := &fakeSvc{
			listCheckpointsItems: []domain.Checkpoint{{
				ID:         "11111111-1111-4111-8111-111111111111",
				DocumentID: "22222222-2222-4222-8222-222222222222",
				RevisionID: "33333333-3333-4333-8333-333333333333",
				VersionNum: 7,
				Label:      "before publish",
				CreatedAt:  time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC),
				CreatedBy:  "user-7",
			}},
		}
		mux := newMux(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/22222222-2222-4222-8222-222222222222/checkpoints", nil)
		req.SetPathValue("id", "22222222-2222-4222-8222-222222222222")
		withAuthHeaders(req, "editor")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
		}

		var body []map[string]json.RawMessage
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
		}
		if len(body) != 1 {
			t.Fatalf("len(body) = %d, want 1", len(body))
		}
		assertCheckpointShape(t, body[0])
	})

	t.Run("Empty", func(t *testing.T) {
		svc := &fakeSvc{listCheckpointsItems: []domain.Checkpoint{}}
		mux := newMux(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/22222222-2222-4222-8222-222222222222/checkpoints", nil)
		req.SetPathValue("id", "22222222-2222-4222-8222-222222222222")
		withAuthHeaders(req, "editor")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		trimmed := bytes.TrimSpace(rr.Body.Bytes())
		if string(trimmed) != "[]" {
			t.Fatalf("empty list must serialize as [], got %q", rr.Body.String())
		}
	})
}

func TestCreateCheckpoint_TypedResponseShape(t *testing.T) {
	svc := &fakeSvc{
		createCheckpointResult: &domain.Checkpoint{
			ID:         "11111111-1111-4111-8111-111111111111",
			DocumentID: "22222222-2222-4222-8222-222222222222",
			RevisionID: "33333333-3333-4333-8333-333333333333",
			VersionNum: 3,
			Label:      "v1",
			CreatedAt:  time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC),
			CreatedBy:  "user-7",
		},
	}
	mux := newMux(t, svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents/22222222-2222-4222-8222-222222222222/checkpoints",
		bytes.NewReader([]byte(`{"label":"v1"}`)),
	)
	req.SetPathValue("id", "22222222-2222-4222-8222-222222222222")
	withAuthHeaders(req, "editor")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
	}
	assertCheckpointShape(t, body)
}
