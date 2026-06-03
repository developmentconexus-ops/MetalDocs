# Plan: Authz Coherence — Phase 1 (Registry Hygiene)

## Summary
Make the Go capability registry (`internal/modules/iam/domain/model.go`) the single source of truth: promote the three inline `Capability("…")` route strings to typed consts, resolve two dead caps, add CI tests binding `permissions.go` routes ↔ registry ↔ DB seed, and fix the wiki cap-name drift. **Zero runtime behavior change** — every capability keeps its exact string value; only the declaration site and CI guards change.

## User Story
As a MetalDocs backend engineer, I want every capability referenced through one typed registry with CI parity checks, so that a misspelled or unseeded capability fails the build instead of silently mis-gating a route.

## Problem → Solution
Capabilities are declared in 5 unsynced places; `permissions.go` uses raw strings (`doc.publish`, `doc.obsolete`, `template.archive`) that bypass `validCapabilities`; two caps are dead; wiki names a nonexistent cap (`membership.grant`). → One typed registry, no inline strings, CI-enforced route↔registry↔seed parity, corrected wiki.

## Metadata
- **Complexity**: Medium
- **Source PRD/ADR**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 1)
- **Estimated Files**: ~7
- **Branch**: off `qa/iam-area-membership` → `feat/authz-coherence-p1-registry`

---

## UX Design
Internal change — no user-facing UX transformation. No HTTP contract change. No OpenAPI change.

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `internal/modules/iam/domain/model.go` | 50-137 | Capability consts + `validCapabilities` + `IsValidCapability`/`AllCapabilities` — where new consts go |
| P0 | `apps/api/cmd/metaldocs-api/permissions.go` | 168,186-189 | The 3 inline `iamdomain.Capability("…")` strings to replace |
| P0 | `apps/api/cmd/metaldocs-api/permissions_test.go` | 49-52,78,345-405 | Test style to mirror; lines 49-52/78 assert inline strings and MUST switch to new consts (same string value) |
| P1 | `db/reference-data/0001_product_reference_data.sql` | 18-109 | `role_capabilities` seed — source for the seed-parity test (regex the INSERT lines) |
| P1 | `wiki/references/local-dev-credentials.md` | 73-74 | `membership.grant` drift → `membership.manage`; `doc.publish` listed for area_admin (it is NOT seeded — see GOTCHA) |
| P2 | `wiki/concepts/authz-tiers.md` | 50-58 | Tier-1 authoring rules the new tests reinforce |

## External Documentation
No external research needed — established internal patterns only.

---

## Patterns to Mirror

### CAPABILITY_CONST_DECLARATION
```go
// SOURCE: internal/modules/iam/domain/model.go:70-82
CapControlledDocumentSupersede Capability = "controlled_documents.supersede"
CapTaxonomyView                Capability = "taxonomy.view"
...
CapSessionManage               Capability = "session.manage"
```
New consts follow the same block + a matching entry in `validCapabilities` (lines 85-114).

### ROUTE_RULE_CAP_REFERENCE
```go
// SOURCE: apps/api/cmd/metaldocs-api/permissions.go:172,188
{method: http.MethodPost, pathExact: "/api/v1/documents", capability: iamdomain.CapDocumentCreate, visibility: iamdelivery.VisibilityPermissionGuarded},
{method: http.MethodPost, pathPrefix: "/api/v1/documents", pathSuffix: "/supersede", capability: iamdomain.CapDocumentSupersede, ...},
```
Inline-string rows (`iamdomain.Capability("doc.publish")`) become `iamdomain.CapDocumentPublish`.

### TABLE_DRIVEN_TEST
```go
// SOURCE: apps/api/cmd/metaldocs-api/permissions_test.go:352-405
func TestPermissionsTable_NoMethodlessWriteShadowing(t *testing.T) {
	t.Parallel()
	for i, r := range routeRules { ... t.Errorf(...) }
}
```
New parity tests live in the same file, same `package main`, iterate `routeRules` / `iamdomain` registry.

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `internal/modules/iam/domain/model.go` | UPDATE | Add `CapDocumentPublish`/`CapDocumentObsolete`/`CapTemplateArchive` consts + map entries; remove dead `CapWorkflowReview`/`CapWorkflowApprove` (const + map) |
| `apps/api/cmd/metaldocs-api/permissions.go` | UPDATE | Replace 3 inline `Capability("…")` with new consts (lines 168,186,187,189) |
| `apps/api/cmd/metaldocs-api/permissions_test.go` | UPDATE | Switch lines 49-52,78 to new consts; ADD `TestEveryRouteCapInRegistry`, `TestEveryCapSeededOrDeferred` |
| `wiki/references/local-dev-credentials.md` | UPDATE | `membership.grant`→`membership.manage`; correct/annotate `doc.publish` row |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | Mark Phase 1 status in-progress→complete on finish |
| `api/openapi/v1/openapi.yaml` | UPDATE (optional, only if trivial) | Prune the 3 stale governance enum caps IF isolated; else defer to Phase 5 |

## NOT Building
- Any tier-2 / area-scoping change (Phase 3).
- Any `x-authz-area` annotation (Phase 2).
- Removing the handler `RoleSystemAdmin` checks (Phase 3).
- Renaming `doc.supersede`→`document.supersede` (cosmetic; defer — it would touch seed + tripwire data and risks behavior drift).
- DB migration (no schema/seed value changes; seed already correct for the kept caps).

---

## Step-by-Step Tasks

### Task 1: Add typed consts for the 3 inline caps
- **ACTION**: In `model.go`, add to the const block + `validCapabilities` map.
- **IMPLEMENT**: `CapDocumentPublish Capability = "doc.publish"`, `CapDocumentObsolete Capability = "doc.obsolete"`, `CapTemplateArchive Capability = "template.archive"`. Add the 3 matching `: {}` lines to `validCapabilities`.
- **MIRROR**: CAPABILITY_CONST_DECLARATION.
- **GOTCHA**: Keep the exact string values (`doc.publish`, not `document.publish`) — these match the DB seed + tripwire data. Changing the string is out of scope.
- **VALIDATE**: `go build ./...`.

### Task 2: Resolve dead caps
- **ACTION**: Remove `CapWorkflowReview`/`CapWorkflowApprove` consts (model.go:59-60) and their `validCapabilities` entries (92-93).
- **GOTCHA**: First `grep -rn "CapWorkflowReview\|CapWorkflowApprove\|workflow.review\|workflow.approve" --include=*.go .` — if ANY non-test reference exists, STOP and keep them (seed+route instead). Audit says none exist.
- **VALIDATE**: `go build ./...`; grep returns only the deleted lines.

### Task 3: Replace inline strings in permissions.go
- **ACTION**: Lines 168,186,187,189 — swap `iamdomain.Capability("template.archive"|"doc.publish"|"doc.obsolete")` for the new consts.
- **MIRROR**: ROUTE_RULE_CAP_REFERENCE.
- **VALIDATE**: `go build ./...`; resolver output unchanged (Task 5 test).

### Task 4: Update existing resolver test to consts
- **ACTION**: `permissions_test.go` lines 49,50,52,78 — replace `iamdomain.Capability("doc.publish")` etc. with `iamdomain.CapDocumentPublish` etc. (line 51 `doc.supersede` stays as-is — not promoted this phase).
- **GOTCHA**: String values identical, so `wantCap` comparisons still pass.
- **VALIDATE**: `go test ./apps/api/...`.

### Task 5: Add route↔registry parity test
- **ACTION**: New test `TestEveryRouteCapInRegistry` in `permissions_test.go`.
- **IMPLEMENT**: iterate `routeRules`; for each non-empty `r.capability`, assert `iamdomain.IsValidCapability(r.capability)`; `t.Errorf` with the offending path on miss.
- **MIRROR**: TABLE_DRIVEN_TEST.
- **VALIDATE**: passes now; temporarily add a bogus inline cap row locally → fails → revert.

### Task 6: Add cap-seeded parity test
- **ACTION**: New test `TestEveryCapSeededOrDeferred` in `permissions_test.go`.
- **IMPLEMENT**: read `db/reference-data/0001_product_reference_data.sql` (relative path from test: `../../../../db/reference-data/0001_product_reference_data.sql` — verify with `filepath` + `os.ReadFile`; if path fragile, use `runtime.Caller` base). Regex-extract the 2nd VALUES arg from each `INSERT INTO metaldocs.role_capabilities (...) VALUES ('role','cap'...)`. Build a set. For each `iamdomain.AllCapabilities()`, assert it is in the seed set OR in a small `deferred` allow-list (`route.manage`, `doc.supersede`). `t.Errorf` unseeded caps.
- **GOTCHA**: `AllCapabilities()` ordering is map-random — sort before asserting for stable output. The seed uses fully-qualified `metaldocs.role_capabilities`; match case-insensitively.
- **VALIDATE**: `go test ./apps/api/...` green; removing a seed line locally makes it fail.

### Task 7: Fix wiki drift
- **ACTION**: `local-dev-credentials.md:73` `membership.grant`→`membership.manage`. Line 73 also lists `doc.publish` for `area_admin` — the seed does NOT grant `doc.publish` to area_admin (grep the seed to confirm); correct the row to reflect actual seeded caps or annotate as illustrative.
- **VALIDATE**: grep `membership.grant` across `wiki/` returns nothing; values match the seed.

### Task 8: Close-out
- **ACTION**: Mark ADR 0022 Phase 1 complete; run full gate suite; dispatch `wiki-curator` if any Key-files anchors shifted.
- **VALIDATE**: all Validation Commands below pass.

---

## Testing Strategy

### Unit Tests
| Test | Input | Expected | Edge Case? |
|---|---|---|---|
| `TestEveryRouteCapInRegistry` | all `routeRules` | every cap ∈ `validCapabilities` | bogus inline cap → fail |
| `TestEveryCapSeededOrDeferred` | `AllCapabilities()` vs seed | each seeded or deferred-allowlisted | unseeded cap → fail |
| `TestPermissionResolver` (existing) | route fixtures | unchanged caps after const swap | no regression |
| `TestPermissionsTable_NoMethodlessWriteShadowing` (existing) | `routeRules` | still green | no regression |

### Edge Cases Checklist
- [ ] Inline cap string reintroduced → `TestEveryRouteCapInRegistry` fails
- [ ] Cap added to registry but not seeded → `TestEveryCapSeededOrDeferred` fails (unless allow-listed)
- [ ] Dead-cap removal breaks no build reference
- [ ] Seed file path resolves from the test's working dir

---

## Validation Commands

### Static + Build
```powershell
$env:GOFLAGS = "-mod=mod"; go build ./...
```
EXPECT: zero errors.

### Unit Tests (touched packages)
```powershell
go test ./apps/api/... ./internal/modules/iam/... -count=1
```
EXPECT: all pass, including the 2 new tests.

### API lint (no spec shape change expected)
```powershell
go run ./scripts/api-lint
npx @redocly/cli lint api/openapi/v1/openapi.yaml
```
EXPECT: no new violations.

### Drift grep
```powershell
Select-String -Path wiki/**/*.md -Pattern "membership.grant"
```
EXPECT: no matches.

### Manual
- [ ] `git diff` shows only declaration-site + test + wiki changes; no tier-2/handler edits.

---

## Acceptance Criteria
- [ ] 3 inline caps promoted to consts; `permissions.go` has no `iamdomain.Capability("…")` literals except intentionally-deferred `doc.supersede` (kept as const already).
- [ ] Dead caps removed (or proven referenced and kept).
- [ ] 2 new parity tests pass and bite on injected drift.
- [ ] Wiki `membership.grant` drift fixed.
- [ ] Build + touched-package tests + api-lint green.
- [ ] No behavior change (resolver output identical).

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Seed-file path fragile in test | Med | Low | Use `runtime.Caller`-based base path; assert file exists first |
| Hidden reference to a dead cap | Low | Med | Grep gate in Task 2 before deletion |
| `doc.supersede` const/string mismatch tempts a rename | Low | Med | Explicitly OUT OF SCOPE; note in PR |

## Notes
This phase intentionally changes **only declaration + CI**, not enforcement — it builds the guardrails Phases 2–5 lean on. Re-enter `metaldocs-backend-api` skill before editing `permissions.go`. Next: `/prp-implement .claude/PRPs/plans/authz-coherence-phase1-registry-hygiene.plan.md`.
