# F1.3 evidence — createTemplate declared-fields-only

> **Feature:** F1.3 — declared-fields-only (Milestone 1 / Contract integrity)
> **Closed:** 2026-06-15
> **Scope:** A3 (H-D class) — `createTemplate` dropped undeclared top-level `id`/`version_id`; handler now emits typed `CreateTemplateResponse` with `TemplateDTO`/`VersionDTO` (ADR 0035 pattern).

---

## 1. Changes

**Backend**
- `internal/modules/templates/delivery/http/routes_generated.go` — `CreateTemplate` replaces `map[string]any` literal with typed `templatesapi.CreateTemplateResponse`; calls `toAPITemplateDTO` + `toAPIVersionDTO` pre-flight; 500 on mapper error.
- `internal/modules/templates/delivery/http/routes_mapping.go` — `toAPITemplateDTO(t *domain.Template) (templatesapi.TemplateDTO, error)` added; mirrors `toAPIVersionDTO` pattern.
- `internal/modules/templates/delivery/http/routes_typed_response_test.go` — `TestCreateTemplate_TypedResponseShape` added (V1/V2 TDD test).
- `internal/modules/templates/delivery/http/routes_create_test.go` — `TestCreateTemplate_Happy` updated with F1.3 top-level absence assertions; `withHeaders` migrated from `"tenant-a"` to UUID tenant; sqlmock mock rows updated to match.
- `internal/modules/templates/delivery/http/routes_autosave_test.go` — template seeds migrated from `TenantID: "tenant-a"` to UUID.
- `internal/modules/templates/delivery/http/routes_contract_test.go` — same migration.
- `internal/modules/templates/delivery/http/routes_lifecycle_test.go` — same migration.
- `internal/modules/templates/delivery/http/routes_query_test.go` — same migration + authz assertion updated.

**Docs**
- `f1.3-declared-fields-only/spec.md` (new)
- `f1.3-declared-fields-only/plan.md` (new)
- `f1.3-declared-fields-only/evidence.md` (this file)

**Scope guard (HS-6):** `toVersionResponse` and `toTemplateResponse` retained — 5+ callers each in lifecycle/query/schema routes (F1.4 territory). Not touched.

---

## 2. Validation gates — outcomes

| Gate | Result | Evidence |
|------|--------|----------|
| V1 — `id`/`version_id` absent at top level | **GREEN** | `TestCreateTemplate_TypedResponseShape` PASS. `TestCreateTemplate_Happy` top-level absence assertions PASS. |
| V2 — typed `data.template` / `data.version` shape | **GREEN** | `TestCreateTemplate_TypedResponseShape` strict-decodes `data.template.id`, `data.template.tenant_id`, `data.template.key`, `data.version.version_number` — PASS. |
| V3 — `TestCreateTemplate_Happy` still passes + absence assertion added | **GREEN** | `go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate` → all 4 PASS. |
| V4 — H-D grep clean | **GREEN** | `grep -n '"id"\|"version_id"' routes_generated.go` → line 92 only (`SetPathValue("id", id)` — path routing, not JSON). Zero JSON field leaks. |
| V5 — build + vet + full suite | **GREEN** | `go build ./...` exit 0; `go vet ./internal/modules/templates/...` exit 0; `go test ./...` 0 FAIL (all packages `ok`). |

---

## 3. TDD proof

- **RED** (pre-impl): `TestCreateTemplate_TypedResponseShape` failed — old `map[string]any` literal emitted `"id"` and `"version_id"` at top level. Test asserted their absence → FAIL.
- **GREEN** (post-impl): handler replaced with typed response, both undeclared keys dropped. Test PASS.
- **Regression**: `TestCreateTemplate_Happy`, `TestCreateTemplate_KeyConflict`, `TestCreateTemplate_RejectUnknownField` all PASS.

---

## 4. Test fixture repair note

`withHeaders` and all `domain.Template.TenantID` seeds in `routes_*_test.go` were `"tenant-a"` (non-UUID string). `toAPITemplateDTO` calls `uuid.Parse(t.TenantID)` — non-UUID tenant → 500. Migrated all 32 occurrences across 6 test files to `"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`. Pre-existing inconsistency (real tenants are always UUIDs in production); F1.3 is the natural trigger to fix it. No production code changed.

---

## 5. Run logs

```
$ go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate -v
=== RUN   TestCreateTemplate_Happy
--- PASS: TestCreateTemplate_Happy (0.00s)
=== RUN   TestCreateTemplate_RejectUnknownField
--- PASS: TestCreateTemplate_RejectUnknownField (0.00s)
=== RUN   TestCreateTemplate_KeyConflict
--- PASS: TestCreateTemplate_KeyConflict (0.00s)
=== RUN   TestCreateTemplate_TypedResponseShape
--- PASS: TestCreateTemplate_TypedResponseShape (0.00s)
PASS
ok  metaldocs/internal/modules/templates/delivery/http 2.970s

$ go build ./...
(exit 0)

$ go vet ./internal/modules/templates/...
(exit 0)

$ go test ./...
(all packages ok — 0 FAIL)
```

---

## 6. Bounded defers

- **D-1 — Zero-timestamp in test fixtures.** `toAPITemplateDTO` calls `t.CreatedAt.UTC()` unconditionally. Fake repo doesn't set `CreatedAt`, so tests emit `"0001-01-01T00:00:00Z"` on the wire. No crash; tests don't assert `created_at`. Same pattern as `toAPIVersionDTO` (pre-existing). In production, `CreatedAt` is always populated from DB. Fix: enforce `CreatedAt` in fake repo OR add a zero-time guard. Defer to a test-quality follow-up.
- **D-2 — `TestCreateTemplate_TypedResponseShape` does not exhaustively check `data.template` declared-key set.** Absence assertions in `TestCreateTemplate_Happy` cover `visibility`, `areas`, `specific_areas` specifically but no full declared-key iteration. Narrow gap; does not affect contract correctness (TemplateDTO governs the wire shape by construction). Defer to F1.4 or a dedicated shape-test hardening task.
- **D-3 — `toVersionResponse`/`toTemplateResponse` retirement.** F1.4 territory (HS-6 scope guard — recorded in spec.md non-goals).

## 7. Closure

F1.3 closed at the Definition-of-Done bar declared in spec.md (V1–V5 GREEN). A3 defect resolved. No deferred items block Milestone 1.
