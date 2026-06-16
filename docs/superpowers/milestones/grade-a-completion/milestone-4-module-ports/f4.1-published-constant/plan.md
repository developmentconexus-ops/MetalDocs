# Feature F4.1 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — Replace hardcoded literal with owning-module constant
**File:** `internal/modules/documents/application/service.go:283`

Change:
```go
overrideStatus := "published"
```
to:
```go
overrideStatus := string(templatesdomain.VersionStatusPublished)
```

`templatesdomain` alias already imported at line 17 — no import change needed.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/documents/application/service.go` | T1: literal → constant |
