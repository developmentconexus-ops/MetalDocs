package httpdelivery

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// jsonTopKeys marshals v and returns its top-level JSON object keys, sorted. It
// is the F7.3 wire-contract probe: search is pre-codegen, so the hand-rolled
// envelope struct must serialize to exactly the OpenAPI-declared key set.
func jsonTopKeys(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// TestSearchDocumentsResponse_WireContract locks searchDocumentsResponse to the
// OpenAPI SearchDocumentsResponse schema (required [items]) and asserts the
// envelope wraps the item list under the unchanged "items" key.
func TestSearchDocumentsResponse_WireContract(t *testing.T) {
	resp := searchDocumentsResponse{
		Items: []SearchDocumentResponse{
			{DocumentID: "d-1", Title: "Spec", Status: "active"},
		},
	}
	if got, want := jsonTopKeys(t, resp), "items"; got != want {
		t.Fatalf("response keys = %q, want %q (OpenAPI SearchDocumentsResponse)", got, want)
	}

	raw, _ := json.Marshal(resp)
	var probe struct {
		Items []struct {
			DocumentID string `json:"document_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(probe.Items) != 1 || probe.Items[0].DocumentID != "d-1" {
		t.Fatalf("items round-trip wrong: %+v", probe.Items)
	}
}

// TestSearchDocumentsResponse_EmptyIsArrayNotNull guards the byte-identical
// parity contract: an empty result must marshal to {"items":[]}, never
// {"items":null}. The handler always passes a non-nil make()-d slice.
func TestSearchDocumentsResponse_EmptyIsArrayNotNull(t *testing.T) {
	resp := searchDocumentsResponse{Items: make([]SearchDocumentResponse, 0)}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(raw); got != `{"items":[]}` {
		t.Fatalf("empty envelope = %s, want {\"items\":[]}", got)
	}
}
