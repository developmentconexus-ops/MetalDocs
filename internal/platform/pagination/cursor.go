// Package pagination provides the canonical opaque keyset-cursor primitive
// shared by list endpoints (api-contract-hardening Phase F · FD-2). A cursor
// encodes the last row's (primary-sort-value, id) tuple; the repository filters
// `(sort_col, id) < (cursor_sort, cursor_id)` under `ORDER BY sort_col DESC,
// id DESC` and fetches limit+1 rows to derive has_more. This mirrors the audit
// module's proven keyset shape; audit predates this package and keeps its own
// helper (its cursor is occurred_at-specific) — adopting this is a later tidy.
package pagination

import (
	"encoding/base64"
	"errors"
	"strings"
)

// ErrInvalidCursor signals a malformed opaque cursor (bad base64 or shape).
var ErrInvalidCursor = errors.New("invalid cursor")

// DefaultLimit / MaxLimit bound the page size (design-system: max 100).
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// ClampLimit applies the design-system bounds: <=0 → DefaultLimit, >MaxLimit →
// MaxLimit.
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// EncodeCursor builds an opaque forward cursor from a keyset tuple: a primary
// sort value already formatted as a stable string (e.g. an RFC3339Nano
// timestamp) plus the row id tiebreaker.
func EncodeCursor(sortValue, id string) string {
	return base64.StdEncoding.EncodeToString([]byte(sortValue + "|" + id))
}

// DecodeCursor splits an opaque cursor back into (sortValue, id). A blank cursor
// is the "first page" and returns ("", "", nil) so callers can branch on the
// empty sortValue without a sentinel error.
func DecodeCursor(cursor string) (sortValue, id string, err error) {
	c := strings.TrimSpace(cursor)
	if c == "" {
		return "", "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return "", "", ErrInvalidCursor
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidCursor
	}
	return parts[0], parts[1], nil
}
