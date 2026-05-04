package resolvers

import "testing"

func TestRegisterBuiltins_RegistersAllV1Resolvers(t *testing.T) {
	r := NewRegistry()
	RegisterBuiltins(r)

	known := r.Known()
	if len(known) != 8 {
		t.Fatalf("expected 8 resolvers, got %d", len(known))
	}

	expectedVersions := map[string]int{
		"doc_code":           1,
		"doc_title":          1,
		"revision_number":    1,
		"effective_date":     1,
		"controlled_by_area": 2,
		"author":             1,
		"approvers":          1,
		"approval_date":      1,
	}
	for key, wantVer := range expectedVersions {
		if known[key] != wantVer {
			t.Fatalf("expected resolver %s to be version %d, got %d", key, wantVer, known[key])
		}
	}
}
