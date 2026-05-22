// export_test.go — test-only helpers that expose package internals.
// This file is compiled only during `go test`.
package ratelimit

// NewConfigUnchecked builds a Config bypassing quota validation.
// Use only in tests that need to exercise the middleware belt for invalid quotas.
func NewConfigUnchecked(q map[RouteKey]int) Config {
	return Config{quotas: q}
}
