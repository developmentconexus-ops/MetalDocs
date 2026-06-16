# Feature F3.4 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — Delete `template_keys.go` and `template_keys_test.go`

`git rm internal/platform/objectstore/template_keys.go internal/platform/objectstore/template_keys_test.go`

No callers to update — zero production refs confirmed pre-spec.

## Files touched

| File | Change |
|------|--------|
| `internal/platform/objectstore/template_keys.go` | deleted |
| `internal/platform/objectstore/template_keys_test.go` | deleted (only tested the deleted functions) |
