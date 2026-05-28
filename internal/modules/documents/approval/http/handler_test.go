package approvalhttp

import "testing"

func TestParseIfMatch(t *testing.T) {
	t.Run("wildcard", func(t *testing.T) {
		version, err := parseIfMatch("*")
		if err != nil {
			t.Fatalf("parseIfMatch(*): %v", err)
		}
		if version != 0 {
			t.Fatalf("version = %d, want 0", version)
		}
	})

	t.Run("positive version", func(t *testing.T) {
		version, err := parseIfMatch(`"v7"`)
		if err != nil {
			t.Fatalf("parseIfMatch(v7): %v", err)
		}
		if version != 7 {
			t.Fatalf("version = %d, want 7", version)
		}
	})

	t.Run("zero version rejected", func(t *testing.T) {
		version, err := parseIfMatch(`"v0"`)
		if err == nil {
			t.Fatalf("expected malformed If-Match error, got version %d", version)
		}
		if err != ErrIfMatchMalformed {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchMalformed)
		}
	})
}
