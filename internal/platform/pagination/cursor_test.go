package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
)

func equalSort(a, b []SortField) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Field != b[i].Field || a[i].Direction != b[i].Direction {
			return false
		}
	}
	return true
}

func equalAnchor(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		ab, errA := json.Marshal(av)
		bb, errB := json.Marshal(bv)
		if errA != nil || errB != nil {
			return false
		}
		if string(ab) != string(bb) {
			return false
		}
	}
	return true
}

func equalCursor(a, b Cursor) bool {
	if a.V != b.V || a.FilterHash != b.FilterHash {
		return false
	}
	if !equalSort(a.Sort, b.Sort) {
		return false
	}
	if !equalAnchor(a.Anchor, b.Anchor) {
		return false
	}
	return true
}

func TestCursor(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		in := Cursor{
			V: CursorVersion,
			Sort: []SortField{
				{Field: "updated_at", Direction: "desc"},
				{Field: "id", Direction: "asc"},
			},
			Anchor: map[string]any{
				"updated_at": "2026-05-10T14:30:00Z",
				"id":         "550e8400-abc",
			},
			FilterHash: "abc123",
		}

		enc, err := Encode(in)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}
		out, err := Decode(enc)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if !equalCursor(in, out) {
			t.Fatalf("decoded cursor mismatch: in=%+v out=%+v", in, out)
		}
	})

	t.Run("tampered base64", func(t *testing.T) {
		enc, err := Encode(Cursor{V: CursorVersion, Sort: []SortField{{Field: "id", Direction: "asc"}}, Anchor: map[string]any{"id": "1"}, FilterHash: "h"})
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}
		r := []rune(enc)
		r[len(r)/2] = '!'
		_, err = Decode(string(r))
		if err == nil {
			t.Fatal("expected error for tampered base64")
		}
	})

	t.Run("bad inner JSON", func(t *testing.T) {
		s := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		_, err := Decode(s)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong version", func(t *testing.T) {
		enc, err := Encode(Cursor{V: 2, Sort: []SortField{{Field: "id", Direction: "asc"}}, Anchor: map[string]any{"id": "1"}, FilterHash: "h"})
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}
		_, err = Decode(enc)
		if !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("expected ErrInvalidCursor, got: %v", err)
		}
	})

	t.Run("ValidateMatch happy path", func(t *testing.T) {
		c := Cursor{V: CursorVersion, Sort: []SortField{{Field: "updated_at", Direction: "desc"}, {Field: "id", Direction: "asc"}}, FilterHash: "h1"}
		err := ValidateMatch(c, []SortField{{Field: "updated_at", Direction: "desc"}, {Field: "id", Direction: "asc"}}, "h1")
		if err != nil {
			t.Fatalf("expected nil, got: %v", err)
		}
	})

	t.Run("ValidateMatch sort-field mismatch", func(t *testing.T) {
		c := Cursor{V: CursorVersion, Sort: []SortField{{Field: "updated_at", Direction: "desc"}}, FilterHash: "h1"}
		err := ValidateMatch(c, []SortField{{Field: "created_at", Direction: "desc"}}, "h1")
		if !errors.Is(err, ErrCursorMismatch) {
			t.Fatalf("expected ErrCursorMismatch, got: %v", err)
		}
	})

	t.Run("ValidateMatch sort-direction mismatch", func(t *testing.T) {
		c := Cursor{V: CursorVersion, Sort: []SortField{{Field: "updated_at", Direction: "desc"}}, FilterHash: "h1"}
		err := ValidateMatch(c, []SortField{{Field: "updated_at", Direction: "asc"}}, "h1")
		if !errors.Is(err, ErrCursorMismatch) {
			t.Fatalf("expected ErrCursorMismatch, got: %v", err)
		}
	})

	t.Run("ValidateMatch filter-hash mismatch", func(t *testing.T) {
		c := Cursor{V: CursorVersion, Sort: []SortField{{Field: "updated_at", Direction: "desc"}}, FilterHash: "h1"}
		err := ValidateMatch(c, []SortField{{Field: "updated_at", Direction: "desc"}}, "h2")
		if !errors.Is(err, ErrCursorMismatch) {
			t.Fatalf("expected ErrCursorMismatch, got: %v", err)
		}
	})

	t.Run("HashFilters deterministic", func(t *testing.T) {
		p1 := url.Values{}
		p1.Add("status", "draft")
		p1.Add("q", "golang")
		p1.Add("status", "published")

		p2 := url.Values{}
		p2.Add("status", "published")
		p2.Add("status", "draft")
		p2.Add("q", "golang")

		h1 := HashFilters(p1, []string{"q", "status"})
		h2 := HashFilters(p2, []string{"q", "status"})
		if h1 != h2 {
			t.Fatalf("expected equal hashes, got %s vs %s", h1, h2)
		}
	})

	t.Run("HashFilters ignores disallowed keys", func(t *testing.T) {
		withExtra := url.Values{"status": {"draft"}, "ignored": {"x"}}
		withoutExtra := url.Values{"status": {"draft"}}
		h1 := HashFilters(withExtra, []string{"status"})
		h2 := HashFilters(withoutExtra, []string{"status"})
		if h1 != h2 {
			t.Fatalf("expected equal hashes, got %s vs %s", h1, h2)
		}
	})

	t.Run("HashFilters multi-value stable", func(t *testing.T) {
		p1 := url.Values{"status": {"draft", "published"}}
		p2 := url.Values{"status": {"published", "draft"}}
		h1 := HashFilters(p1, []string{"status"})
		h2 := HashFilters(p2, []string{"status"})
		if h1 != h2 {
			t.Fatalf("expected equal hashes, got %s vs %s", h1, h2)
		}
	})

	t.Run("ClampLimit cases", func(t *testing.T) {
		cases := []struct {
			name      string
			requested int
			want      int
		}{
			{name: "zero", requested: 0, want: 20},
			{name: "negative", requested: -1, want: 20},
			{name: "within", requested: 50, want: 50},
			{name: "over", requested: 200, want: 100},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := ClampLimit(tc.requested)
				if got != tc.want {
					t.Fatalf("ClampLimit(%d)=%d, want %d", tc.requested, got, tc.want)
				}
			})
		}
	})

	t.Run("AppendIDTieBreaker", func(t *testing.T) {
		cases := []struct {
			name string
			in   []SortField
			want []SortField
		}{
			{name: "empty", in: []SortField{}, want: []SortField{{Field: "id", Direction: "asc"}}},
			{name: "append to non-id tail", in: []SortField{{Field: "updated_at", Direction: "desc"}}, want: []SortField{{Field: "updated_at", Direction: "desc"}, {Field: "id", Direction: "asc"}}},
			{name: "already has id tail", in: []SortField{{Field: "updated_at", Direction: "desc"}, {Field: "id", Direction: "asc"}}, want: []SortField{{Field: "updated_at", Direction: "desc"}, {Field: "id", Direction: "asc"}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := AppendIDTieBreaker(tc.in)
				if !equalSort(got, tc.want) {
					t.Fatalf("got=%+v want=%+v", got, tc.want)
				}
			})
		}
	})
}