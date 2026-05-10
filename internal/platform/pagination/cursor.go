package pagination

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
)

type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"dir"`
}

type Cursor struct {
	V          int            `json:"v"`
	Sort       []SortField    `json:"sort"`
	Anchor     map[string]any `json:"anchor"`
	FilterHash string         `json:"filter_hash"`
}

const (
	DefaultLimit  = 20
	MaxLimit      = 100
	CursorVersion = 1
)

var (
	ErrInvalidCursor = errors.New("invalid cursor")
	ErrCursorMismatch = errors.New("cursor mismatch")
)

func Encode(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Decode(s string) (Cursor, error) {
	if s == "" {
		return Cursor{}, ErrInvalidCursor
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}

	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	if c.V != CursorVersion {
		return Cursor{}, fmt.Errorf("%w: bad version %d", ErrInvalidCursor, c.V)
	}
	for i, sf := range c.Sort {
		if sf.Direction != "asc" && sf.Direction != "desc" {
			return Cursor{}, fmt.Errorf("%w: sort[%d].dir %q must be asc or desc", ErrInvalidCursor, i, sf.Direction)
		}
		if sf.Field == "" {
			return Cursor{}, fmt.Errorf("%w: sort[%d].field empty", ErrInvalidCursor, i)
		}
	}

	return c, nil
}

func ValidateMatch(c Cursor, currentSort []SortField, currentFilterHash string) error {
	if len(c.Sort) != len(currentSort) {
		return ErrCursorMismatch
	}

	for i := range c.Sort {
		if c.Sort[i].Field != currentSort[i].Field || c.Sort[i].Direction != currentSort[i].Direction {
			return ErrCursorMismatch
		}
	}

	if c.FilterHash != currentFilterHash {
		return ErrCursorMismatch
	}

	return nil
}

// HashFilters builds a deterministic SHA-256 hex over allowed filter params.
// Keys + values are URL-escaped before concatenation so that values containing
// reserved delimiters (`;`, `,`, `=`) cannot collide between distinct inputs.
func HashFilters(params url.Values, allowed []string) string {
	keys := make([]string, 0, len(allowed))
	for _, k := range allowed {
		if _, ok := params[k]; ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	canonical := ""
	for i, k := range keys {
		vals := append([]string(nil), params[k]...)
		sort.Strings(vals)

		if i > 0 {
			canonical += ";"
		}
		canonical += url.QueryEscape(k) + "="
		for j, v := range vals {
			if j > 0 {
				canonical += ","
			}
			canonical += url.QueryEscape(v)
		}
	}

	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func ClampLimit(requested int) int {
	if requested <= 0 {
		return DefaultLimit
	}
	if requested > MaxLimit {
		return MaxLimit
	}
	return requested
}

// AppendIDTieBreaker appends `{id, asc}` if the tail is not already that exact
// pair. A tail of `{id, desc}` would still break stable forward pagination, so
// both Field AND Direction must match before we skip the append.
func AppendIDTieBreaker(sortFields []SortField) []SortField {
	if n := len(sortFields); n > 0 {
		tail := sortFields[n-1]
		if tail.Field == "id" && tail.Direction == "asc" {
			return sortFields
		}
	}
	return append(sortFields, SortField{Field: "id", Direction: "asc"})
}