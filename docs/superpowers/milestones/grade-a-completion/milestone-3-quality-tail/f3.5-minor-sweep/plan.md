# Feature F3.5 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — Deprecate `New` constructor
**File:** `internal/modules/documents/application/service.go:127`
Add `// Deprecated: use NewService.` above `func New(...)`.

### T2 — SHA-1 → SHA-256 in `stableID`
**File:** `internal/modules/security/application/service.go`
- Replace `"crypto/sha1"` import with `"crypto/sha256"`
- Replace `sha1.New()` with `sha256.New()`

### Defers (recorded, not implemented)
- Hardcoded PT string `"Criacao do documento"` — defer to locale/i18n work
- `tenantIDFromRequest` cross-module extraction — defer to M4

## Files touched

| File | Change |
|------|--------|
| `internal/modules/documents/application/service.go` | T1: deprecation comment |
| `internal/modules/security/application/service.go` | T2: sha1 → sha256 |
