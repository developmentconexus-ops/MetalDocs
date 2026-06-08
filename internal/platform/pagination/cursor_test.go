package pagination

import "testing"

func TestEncodeDecodeCursor_RoundTrips(t *testing.T) {
	enc := EncodeCursor("2026-06-08T11:00:00.123456789Z", "doc-1")
	sv, id, err := DecodeCursor(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if sv != "2026-06-08T11:00:00.123456789Z" || id != "doc-1" {
		t.Fatalf("round-trip mismatch: sv=%q id=%q", sv, id)
	}
}

func TestDecodeCursor_BlankIsFirstPage(t *testing.T) {
	sv, id, err := DecodeCursor("  ")
	if err != nil || sv != "" || id != "" {
		t.Fatalf("blank cursor should be first page, got sv=%q id=%q err=%v", sv, id, err)
	}
}

func TestDecodeCursor_Malformed(t *testing.T) {
	for _, c := range []string{"!!!notbase64", EncodeCursor("only", "")[:4]} {
		if _, _, err := DecodeCursor(c); err != ErrInvalidCursor {
			t.Fatalf("expected ErrInvalidCursor for %q, got %v", c, err)
		}
	}
	// base64 of a value with no "|" separator.
	bad := EncodeCursor("", "") // "|" with empty parts → invalid
	if _, _, err := DecodeCursor(bad); err != ErrInvalidCursor {
		t.Fatalf("expected ErrInvalidCursor for empty-parts cursor, got %v", err)
	}
}

func TestClampLimit(t *testing.T) {
	cases := map[int]int{0: DefaultLimit, -5: DefaultLimit, 10: 10, 100: 100, 250: MaxLimit}
	for in, want := range cases {
		if got := ClampLimit(in); got != want {
			t.Fatalf("ClampLimit(%d)=%d want %d", in, got, want)
		}
	}
}
