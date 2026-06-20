# Feature F7.5 — Evidence (honest H-D gate + FE codegen regen + final proof)

> **Status:** CLOSED 2026-06-20. Spec `./spec.md` (approved), plan `./plan.md`.

## Summary

Documented the honest two-part H-D gate in `wiki/architecture/api-contract.md` (§5b) so future
milestones measure response-literal completeness, not Grep A alone; regenerated the FE openapi-typescript
types for F7.4's 4 documents 200 schemas; refreshed the api-contract.md `Last verified` stamp to M7; and
ran the whole-repo terminal proof.

## Acceptance — every gate, real commands + output

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| 1 | Honest two-part gate documented | api-contract.md §5b "Response-body typing gate (H-D — honest two-part)" | DONE — Part A + Part B + non-response allowlist + anti-evasion note |
| 2 | Part A = 0 whole-repo | `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go \| wc -l` | **0** |
| 3 | Part B = allowlist only (0 response literals) | `grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` | **11 hits, all non-response** (see classification below) |
| 4 | FE codegen regen clean + scoped | `npm run gen:api`; `git diff --stat src/lib/api-types/index.d.ts` | +32/-4, limited to the 4 documents response types (`placeholder_schema: unknown[]`, `options: unknown[]`, `pdf_status: string`, …) |
| 5 | BE codegen fresh | `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...`; `git diff --exit-code api.gen.go` | CLEAN (no uncommitted diff) |
| 6 | Whole-repo build green | `GOFLAGS=-mod=mod go build ./...` | exit 0 |
| 7 | Whole-repo tests green | `GOFLAGS=-mod=mod go test -count=1 ./...` | exit 0, **0 FAIL** |
| 8 | Wiki stamp current | api-contract.md `Last verified: 2026-06-20 (M7 …)` | DONE |

## Part B classification (all 11 survivors are non-response — allowlisted)

| Site | Category | Allowlisted? |
|------|----------|--------------|
| `audit/.../handler.go:404` `payload := map[string]any{}` | JSON decode buffer feeding `AuditEventItem.Payload` domain-mirror field | ✓ |
| `auth/.../handler.go:98,109,127` | `recordAudit(...)` audit-emit payloads | ✓ |
| `auth/.../handler.go:204` | `recordAudit(... payload map[string]any)` param decl | ✓ |
| `controlleddocuments/.../routes.go:89,224` `formData := map[string]any(nil)` | command input | ✓ |
| `documents/.../handler.go:615` `ContentFormData: map[string]any{...}` | command input | ✓ |
| `iam/.../people_handler.go:321` | `recordAudit(...)` audit-emit payload | ✓ |
| `iam/.../people_handler.go:454` | `recordAudit(... payload map[string]any)` param decl | ✓ |
| `security/.../handler.go:54` `Evidence map[string]any` | domain-mirror struct field | ✓ |

**Zero response literals** — the H-D class is closed, and the gate that proves it is no longer blind to
the `writeFillInJSON` alias / multi-line construction (the M6 blindspot).

## Files changed

- `wiki/architecture/api-contract.md` — new §5b honest gate; `Last verified` → 2026-06-20 (M7).
- `frontend/apps/web/src/lib/api-types/index.d.ts` — regenerated (4 documents response types).

## Scope / HS discipline

- Documentation + FE type regen + proof only. No pipeline standup, no routing rewire (HS-2 held).
- No allowlisted non-response `map[string]any` converted (out of scope).
- Contract-first order honored: OpenAPI (F7.4) → BE `api.gen.go` (F7.4) → FE types (here).

## Defers

None.

## Milestone exit state (pre-validator)

All 5 features (F7.1–F7.5) closed with evidence. Whole-repo: Part A = 0, Part B = allowlist-only,
`go generate` no-diff, `go build` clean, `go test -count=1 ./...` 0 FAIL. Ready for Phase 4 — dispatch
the independent `milestone-validator`.
