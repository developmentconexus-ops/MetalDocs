package http

import (
	"encoding/json"
	"sort"
	"testing"
	"time"

	"metaldocs/internal/modules/templates/domain"
)

func mapKeysT(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// removedFlatKeys are the four coupled version/revision scalars that ADR 0065
// deleted from TemplateDTO. None of them may ever reappear on the wire.
var removedFlatKeys = []string{
	"latest_revision_number",
	"published_version_id",
	"published_version_number",
	"current_revision_number",
}

func marshalTemplateDTO(t *testing.T, read *domain.TemplateRead) map[string]json.RawMessage {
	t.Helper()
	dto, err := toAPITemplateDTO(read, "")
	if err != nil {
		t.Fatalf("toAPITemplateDTO: %v", err)
	}
	encoded, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal TemplateDTO: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode encoded body: %v (raw=%s)", err, string(encoded))
	}
	return raw
}

func assertRemovedKeysAbsent(t *testing.T, raw map[string]json.RawMessage) {
	t.Helper()
	for _, k := range removedFlatKeys {
		if _, ok := raw[k]; ok {
			t.Fatalf("ADR 0065: removed flat key %q reappeared on the wire — keys=%v", k, mapKeysT(raw))
		}
	}
}

func baseRead() *domain.TemplateRead {
	return &domain.TemplateRead{
		Template: domain.Template{
			ID:            "11111111-1111-1111-1111-111111111111",
			TenantID:      "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			Key:           "invoice-template",
			Name:          "Invoice Template",
			CreatedBy:     "user-123",
			LatestVersion: 1,
			CreatedAt:     time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		},
		Latest: domain.VersionRef{
			ID:             "22222222-2222-4222-8222-222222222222",
			Number:         1,
			RevisionNumber: 0,
			Status:         domain.VersionStatusUnderReview,
		},
	}
}

// TestTemplateDTO_PublishedVersionRef_PresentAndNull — ADR 0065 pin (P1+P2+P3).
// published_version is required-and-nullable: for a never-published template
// the key MUST be present with literal null (the 9f86828b guarantee carries
// forward), the nested latest_version ref carries exactly {id, number,
// revision_number, status}, and none of the four removed flat scalars reappear.
func TestTemplateDTO_PublishedVersionRef_PresentAndNull(t *testing.T) {
	raw := marshalTemplateDTO(t, baseRead())

	// P1 — published_version present-and-null.
	pv, ok := raw["published_version"]
	if !ok {
		t.Fatalf("P1: missing required key published_version (present-and-null contract) — keys=%v", mapKeysT(raw))
	}
	if string(pv) != "null" {
		t.Fatalf("P1: expected published_version to encode as null for a never-published template, got %s", string(pv))
	}

	// P2 — latest_version is a nested ref with exactly {id, number, revision_number, status}.
	lv, ok := raw["latest_version"]
	if !ok {
		t.Fatalf("P2: missing required key latest_version — keys=%v", mapKeysT(raw))
	}
	var ref map[string]json.RawMessage
	if err := json.Unmarshal(lv, &ref); err != nil {
		t.Fatalf("P2: latest_version is not an object: %v (raw=%s)", err, string(lv))
	}
	gotKeys := mapKeysT(ref)
	sort.Strings(gotKeys)
	wantKeys := []string{"id", "number", "revision_number", "status"}
	if len(gotKeys) != len(wantKeys) {
		t.Fatalf("P2: latest_version key-set = %v, want exactly %v", gotKeys, wantKeys)
	}
	for i, k := range wantKeys {
		if gotKeys[i] != k {
			t.Fatalf("P2: latest_version key-set = %v, want exactly %v", gotKeys, wantKeys)
		}
	}
	if string(ref["status"]) != `"under_review"` {
		t.Fatalf("P2: latest_version.status = %s, want \"under_review\"", string(ref["status"]))
	}

	// P3 — removed flat keys absent.
	assertRemovedKeysAbsent(t, raw)
}

// TestTemplateDTO_PublishedVersionRef_PresentWhenSet — ADR 0065 pin (P4). Once
// a template has a published version, published_version encodes as the full
// nested ref object with all four keys and status "published".
func TestTemplateDTO_PublishedVersionRef_PresentWhenSet(t *testing.T) {
	read := baseRead()
	read.Published = &domain.VersionRef{
		ID:             "33333333-3333-4333-8333-333333333333",
		Number:         1,
		RevisionNumber: 0,
		Status:         domain.VersionStatusPublished,
	}

	raw := marshalTemplateDTO(t, read)

	pv, ok := raw["published_version"]
	if !ok {
		t.Fatalf("P4: missing required key published_version — keys=%v", mapKeysT(raw))
	}
	var ref map[string]json.RawMessage
	if err := json.Unmarshal(pv, &ref); err != nil {
		t.Fatalf("P4: published_version is not an object: %v (raw=%s)", err, string(pv))
	}
	if string(ref["id"]) != `"33333333-3333-4333-8333-333333333333"` {
		t.Fatalf("P4: published_version.id = %s", string(ref["id"]))
	}
	if string(ref["status"]) != `"published"` {
		t.Fatalf("P4: published_version.status = %s, want \"published\"", string(ref["status"]))
	}
	for _, k := range []string{"id", "number", "revision_number", "status"} {
		if _, ok := ref[k]; !ok {
			t.Fatalf("P4: published_version missing key %q — keys=%v", k, mapKeysT(ref))
		}
	}

	assertRemovedKeysAbsent(t, raw)
}
