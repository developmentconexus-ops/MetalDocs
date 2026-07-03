package http

import (
	"encoding/json"
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

// TestTemplateDTO_NullableFields_PresentAndNull pins the governance contract
// behind the openapi.yaml `required` list for TemplateDTO: doc_type_code,
// published_version_id, published_version_number, current_revision_number,
// and archived_at must always be present in the JSON body — as `null` when
// the template has no published version / doc type / archival — never
// omitted.
//
// /templates list consumers gate template clonability on published_version_id
// being present and comparable against `null` (the frontend does a strict
// `!== null` selectability check). Before api.gen.go was regenerated off the
// updated openapi.yaml `required` list, these fields carried `omitempty` and
// were silently dropped from the wire body for never-published templates,
// defeating that check (bug fixed 2026-07-03). Regression is only possible
// by re-editing openapi.yaml's `required` list for TemplateDTO and
// regenerating api.gen.go — this test pins the generated struct's JSON shape
// directly so such a regression fails compilation-adjacent CI, not just a
// manual QA pass.
func TestTemplateDTO_NullableFields_PresentAndNull(t *testing.T) {
	tpl := &domain.Template{
		ID:                     "11111111-1111-1111-1111-111111111111",
		TenantID:               "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Key:                    "invoice-template",
		Name:                   "Invoice Template",
		CreatedBy:              "user-123",
		LatestVersion:          1,
		CreatedAt:              time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		DocTypeCode:            "",
		PublishedVersionID:     nil,
		PublishedVersionNumber: nil,
		CurrentRevisionNumber:  nil,
		ArchivedAt:             nil,
	}

	dto, err := toAPITemplateDTO(tpl, "")
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

	nullableKeys := []string{
		"published_version_id",
		"published_version_number",
		"current_revision_number",
		"doc_type_code",
		"archived_at",
	}
	for _, k := range nullableKeys {
		v, ok := raw[k]
		if !ok {
			t.Fatalf("missing required key %q in TemplateDTO body (omitempty regression) — keys=%v", k, mapKeysT(raw))
		}
		if string(v) != "null" {
			t.Fatalf("expected key %q to encode as null for a never-published template, got %s", k, string(v))
		}
	}
}

// TestTemplateDTO_NullableFields_PresentWhenSet is the positive sibling of
// TestTemplateDTO_NullableFields_PresentAndNull: once a template has a
// published version, published_version_id must encode as the quoted UUID
// string, not null.
func TestTemplateDTO_NullableFields_PresentWhenSet(t *testing.T) {
	publishedVersionID := "22222222-2222-4222-8222-222222222222"
	publishedVersionNumber := 1
	currentRevisionNumber := 0

	tpl := &domain.Template{
		ID:                     "11111111-1111-1111-1111-111111111111",
		TenantID:               "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Key:                    "invoice-template",
		Name:                   "Invoice Template",
		CreatedBy:              "user-123",
		LatestVersion:          1,
		CreatedAt:              time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		PublishedVersionID:     &publishedVersionID,
		PublishedVersionNumber: &publishedVersionNumber,
		CurrentRevisionNumber:  &currentRevisionNumber,
	}

	dto, err := toAPITemplateDTO(tpl, "")
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

	v, ok := raw["published_version_id"]
	if !ok {
		t.Fatalf("missing required key %q in TemplateDTO body — keys=%v", "published_version_id", mapKeysT(raw))
	}
	want := `"` + publishedVersionID + `"`
	if string(v) != want {
		t.Fatalf("expected published_version_id to encode as %s, got %s", want, string(v))
	}
}
