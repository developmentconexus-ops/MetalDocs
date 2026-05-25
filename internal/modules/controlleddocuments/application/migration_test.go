package application

import "testing"

func TestLegacyCodeFromDocumentID(t *testing.T) {
	t.Run("valid uuid-like id", func(t *testing.T) {
		got, err := legacyCodeFromDocumentID("2af9e3dd-48e2-4aef-8ecc-57f34eb35de5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "MIG-2AF9E3DD" {
			t.Fatalf("legacy code = %q, want MIG-2AF9E3DD", got)
		}
	})

	t.Run("short id returns error", func(t *testing.T) {
		_, err := legacyCodeFromDocumentID("abc")
		if err == nil {
			t.Fatalf("expected error for short id, got nil")
		}
	})
}
