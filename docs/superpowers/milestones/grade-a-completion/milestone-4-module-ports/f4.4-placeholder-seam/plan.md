# Feature F4.4 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — Read-pass
Read `documents/repository/repository.go:112`, `documents/application/service.go:29`,
all other `templatesdomain` references in `documents/`, and `templates/domain/schemas.go:61`
to determine legitimate-vs-leak.

### T2 — Record BOUNDARY DECISION
Write the decision in `spec.md` with: the rule applied, why it's not an H-G site, why a port
would be over-engineering, and how the auditor can confirm it.

### T3 — Confirm build + tests
No code change. Verify `go build ./...` and `go test ./internal/modules/documents/...` still pass.

## Files touched

| File | Change |
|------|--------|
| `docs/.../f4.4-placeholder-seam/spec.md` | T2: BOUNDARY DECISION |
| *(none in production code)* | Legitimate dependency; no seam change needed |
