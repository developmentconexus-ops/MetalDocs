package repository

import (
	"strings"
	"testing"
)

// TestBuildDocumentFilter_EscapesLikeWildcards guards A2: user input in the
// `q` filter must be escaped for LIKE so wildcard metacharacters (%, _, \)
// cannot be injected (CWE-943), and the surrounding clause must carry the
// matching ESCAPE so Postgres interprets the backslash escapes.
func TestBuildDocumentFilter_EscapesLikeWildcards(t *testing.T) {
	where, args := buildDocumentFilter("tenant-1", ListOptions{Q: `a%b_c\d`})

	if !strings.Contains(where, "name ILIKE") || !strings.Contains(where, `ESCAPE '\'`) {
		t.Fatalf("expected escaped ILIKE clause with ESCAPE, got: %q", where)
	}

	last, ok := args[len(args)-1].(string)
	if !ok {
		t.Fatalf("expected string q arg, got %T", args[len(args)-1])
	}
	const want = `%a\%b\_c\\d%`
	if last != want {
		t.Fatalf("q arg not escaped:\n got %q\nwant %q", last, want)
	}
}
