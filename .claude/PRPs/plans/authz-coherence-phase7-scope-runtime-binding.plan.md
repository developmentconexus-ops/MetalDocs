# Plan: ADR 0022 Phase 7 — Bind Typed Capability Scope to Runtime

## Summary
Phase 1-5 declared a typed `ScopeArea`/`ScopeTenant` classification but 5 area-grade
capabilities still pass the literal `"tenant"` to tier-2 `authz.Require` — declared area,
enforced tenant-wide (the BOLA the ADR diagnosed, relocated into a typed-but-unbound map).
Phase 7 adds the missing **declaration↔enforcement binding**: a CI/AST guard that bans
`authz.Require(<areaGradeCap>, "tenant")`, enforces real area on all 5 caps, makes the
annotation test self-maintaining, and folds in the Phase 1-5 review MED/LOW fixes.

## User Story
As a tenant operator under ISO segregation, I want document/CD authoring authorized
against the resource's process area, so that an `editor`/`area_admin` in area QMS cannot
create or edit documents in area RH (cross-area BOLA closed), while `system_admin` stays
tenant-wide via the tier-2 bypass.

## Problem → Solution
5 area-grade caps (`document.create`, `document.edit`, `controlled_documents.create`,
`controlled_documents.obsolete`, `controlled_documents.supersede`) pass `"tenant"` →
area filter OFF. **Solution:** pass the real area (loaded/derived per call site), add an
AST guard so the gap cannot reappear, and bind the spec annotation test to `IsAreaGrade`.

## Metadata
- **Complexity**: Large
- **Source PRD**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 7 amendment, 2026-06-03)
- **PRD Phase**: Phase 7 — bind typed scope to runtime
- **Estimated Files**: ~16 (2 lint, 1 lint test, 2 documents, 2 CD, 1 authz, 1 iam repo, 1 iam service, 1 scope test, ADR, + test files)
- **Product ruling**: Option 1 (area-segregate all 5) — confirmed this session.

---

## UX Design
Internal authz change — no user-facing UX transformation. Behavior delta: a non-system
actor lacking a `user_process_areas` row (with an area-granting role) **in the target
document's/CD's area** now receives `ErrCapDenied` → 403 on create/edit/obsolete/supersede,
where previously any tenant-wide grant sufficed. `system_admin` unchanged (tier-2 bypass).

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `internal/modules/iam/authz/authz.go` | 51-118 | `Require` semantics (`$2='tenant'` skips filter); `BypassSystem` to unexport; system_admin EXISTS (66-82) to extract |
| P0 | `internal/modules/iam/domain/capability_scope.go` | all | `IsAreaGrade`, the 5 caps' classification (stays area-grade) |
| P0 | `internal/modules/iam/domain/model.go` | 54-110 | const→value mapping the AST guard must parse |
| P0 | `scripts/api-lint/registry_rules.go` | all | where the new `authz-area-scope-binding` rule lands; AST scan idiom |
| P0 | `scripts/api-lint/code_rules.go` | 21-76, 302-385 | `RunCodeRules` wiring; `parser.ParseFile`/`ast.Inspect` idiom; tripwire scan as a model |
| P0 | `internal/modules/documents/repository/repository.go` | 76-141, 250-290, 510-560, 620-665, 820-835, 960-970, 1410-1455 | `document.create` (93) + ~12 `document.edit` sites; each needs the area |
| P0 | `internal/modules/controlleddocuments/application/service.go` | 150-270, 440-560, 560-620 | CD create (211), preview (417), changeStatus (508, loads row FOR UPDATE), create-revision (607) |
| P0 | `internal/modules/controlleddocuments/infrastructure/repository.go` | 320-352 | CD repo Create (328) + CreateTx (348) |
| P0 | `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | 89-181 | `MembershipDirectoryScope` (SQL `now()` → param) + dup system_admin EXISTS (99-128) |
| P0 | `internal/modules/iam/application/area_membership_service.go` | 26-95 | `nowFn` (48,55); `DirectoryScope` caller (75) to thread `now` |
| P1 | `apps/api/cmd/metaldocs-api/permissions_authz_scope_test.go` | all | `areaEnforcedOps` to derive from `IsAreaGrade` |
| P1 | `apps/api/cmd/metaldocs-api/permissions.go` | route→cap table | map area-grade cap → operationId for self-maintaining test |
| P1 | `internal/modules/documents/approval/application/{cancel,cutover,scheduler}_service.go` | BypassSystem sites | background callers needing the new bridge marker |
| P2 | `internal/modules/iam/authz/authz_test.go`, `repository_autosave_tripwire_test.go` | BypassSystem usage | tests to update for the bridge |

## External Documentation
No external research needed — established internal patterns (Go AST lint, tier-2 authz,
context-marker capability). OWASP API1/API5 + CWE-269 rationale already in
`wiki/references/authz-industry-evidence.md`.

---

## Patterns to Mirror

### AST_LINT_SCAN
// SOURCE: scripts/api-lint/registry_rules.go:91-143 (checkNoInlineCapability)
Walk modulesRoot, skip `.git/.claude/node_modules/vendor` dirs and `_test.go`,
`parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)`, `ast.Inspect` for
`*ast.CallExpr`, emit `Violation{File,Line,Rule,Message}`.

### TIER2_AREA_CALL (the target shape)
// SOURCE: internal/modules/iam/infrastructure/postgres/user_area_repository.go:197
`if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), membership.AreaCode); err != nil {`
— the real-area pattern Phase 3 already uses; mirror for documents/CD.

### CD_ROW_LOAD_BEFORE_AUTHZ
// SOURCE: controlleddocuments/application/service.go:512-523 (FOR UPDATE select)
Add `process_area_code` to the existing `SELECT status ... FOR UPDATE`, reorder so the
row (status+area) loads, then `authz.Require(ctx, tx, cap, areaCode)`, then mutate.
(FOR UPDATE select is not mutating SQL → tripwire still satisfied: authz precedes the UPDATE.)

### CONTEXT_MARKER (for the bypass bridge)
// SOURCE: internal/modules/iam/authz/authz.go:21-39 (capCacheKey/WithCapCache)
Unexported key type + `context.WithValue` + typed getter — mirror for `bgBypassKey{}`.

### TEST_STRUCTURE (lint bite-then-green)
// SOURCE: ADR 0022 Phase 5 close-out — inject drift, assert violation, revert.
Go table tests in `scripts/api-lint/*_test.go`; real-DB tests in `tests/integration/`.

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `scripts/api-lint/registry_rules.go` | UPDATE | add `checkAuthzAreaScopeBinding` (the core control) + wire into `RunRegistryRules` |
| `scripts/api-lint/registry_rules_test.go` | UPDATE/CREATE | prove the guard bites on `Require(areaCap,"tenant")`, green otherwise |
| `internal/modules/iam/authz/authz.go` | UPDATE | unexport `BypassSystem`→bridge w/ ctx marker; extract `systemAdminExistsSQL` const |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | UPDATE | `MembershipDirectoryScope` takes `now time.Time` (bind `$`); reuse `systemAdminExistsSQL` |
| `internal/modules/iam/application/area_membership_service.go` | UPDATE | thread `s.nowFn()` into `DirectoryScope`→`MembershipDirectoryScope` |
| `internal/modules/documents/repository/repository.go` | UPDATE | `document.create` (93) → `d.ProcessAreaCodeSnapshot`; ~12 `document.edit` → loaded area; `BypassSystem`→bridge (687,1050) |
| `internal/modules/controlleddocuments/application/service.go` | UPDATE | create (211)/preview(417)/create-rev(607) → `cmd.ProcessAreaCode`/`cd.ProcessAreaCode`; changeStatus(508) → loaded area |
| `internal/modules/controlleddocuments/infrastructure/repository.go` | UPDATE | Create(328)/CreateTx(348) → `doc.ProcessAreaCode` |
| `internal/modules/documents/approval/application/{cancel,cutover,scheduler}_service.go` | UPDATE | mark ctx background before bridge call |
| `apps/api/cmd/metaldocs-api/permissions_authz_scope_test.go` | UPDATE | derive `areaEnforcedOps` from route-table ∩ `IsAreaGrade` |
| `scripts/api-lint/registry_rules.go` (godoc) | UPDATE | document `no-inline-capability` literal-only limitation |
| `apps/api/cmd/metaldocs-api/permissions_test.go` (or new golden) | UPDATE | golden seed `role_capabilities` row-count test |
| test files (documents, CD, iam, authz) | UPDATE | area-scoped assertions + bridge marker |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | Phase 7 completion note |

## NOT Building
- No OpenAPI **shape** change (paths/params/schemas) — only `x-authz-skip-area` metadata on document/CD ops if the self-maintaining test pulls them in.
- No new managed-areas table (YAGNI — `user_process_areas` already expresses it).
- No call-graph lint rewrite / `authz-call-present` tx-layer activation (still deferred residual Phase 5+).
- No reclassification of any cap (Option 1: all 5 stay `ScopeArea`).
- No change to `document.submit/signoff`, `doc.publish/obsolete/supersede`, `membership.manage` (already area-enforced).

---

## Step-by-Step Tasks

### Task 1: Core control — AST guard `authz-area-scope-binding`
- **ACTION**: Add `checkAuthzAreaScopeBinding(modulesRoot, fset)` to `scripts/api-lint/registry_rules.go`; call from `RunRegistryRules`.
- **IMPLEMENT**:
  1. Parse `internal/modules/iam/domain/model.go` once → `map[constName]Capability` from the `Cap... Capability = "..."` `ValueSpec`s.
  2. Walk non-test `.go` (skip the same dirs as `checkNoInlineCapability`). `ast.Inspect` for `*ast.CallExpr` where `Fun` is `*ast.SelectorExpr` `authz.Require`.
  3. Inspect the **4th arg** (areaCode, index 3): if `*ast.BasicLit` STRING with value `"tenant"`.
  4. Inspect the **3rd arg** (capability, index 2): unwrap `string(<expr>)` conversion; if `<expr>` is `iamdomain.CapXxx`/`domain.CapXxx`/`CapXxx`, resolve constName→value via the model.go map; if `iamdomain.IsAreaGrade(value)` → emit `Violation{Rule:"authz-area-scope-binding"}`.
  5. Variable/unresolvable cap args → skip (documented limitation; mirrors literal-only limit).
- **MIRROR**: AST_LINT_SCAN.
- **IMPORTS**: already in file (`go/ast`, `go/parser`, `go/token`, `iamdomain`).
- **GOTCHA**: `RequireAll` has a different arg shape — scope this rule to `Require` only. Build the const map from model.go, NOT reflection (consts aren't reflectable).
- **VALIDATE**: `go test ./scripts/api-lint/...` new test asserts a fixture `authz.Require(ctx,tx,string(iamdomain.CapDocumentEdit),"tenant")` yields 1 violation; `...,areaCode)` yields 0.

### Task 2: Prove the guard bites (pre-fix baseline)
- **ACTION**: Before any call-site fix, run `go run ./scripts/api-lint api/openapi/v1/openapi.yaml .` — capture the new violations against document.create/edit + CD create sites (the const+"tenant" gaps).
- **VALIDATE**: violation count > prior 455 baseline by exactly the area-grade-const+"tenant" sites. Record in ADR note.

### Task 3: Enforce area — CD changeStatus (obsolete/supersede)
- **ACTION**: In `changeStatus` (service.go:486), add `process_area_code` to the `FOR UPDATE` SELECT; move/keep `authz.Require` AFTER the row load, passing the loaded area.
- **IMPLEMENT**: `SELECT status, process_area_code FROM controlled_documents WHERE ... FOR UPDATE` → scan `currentStatus, areaCode`; check `ErrCDNotFound`/`ErrCDNotActive`; then `authz.Require(ctx, tx, cap, areaCode)`.
- **GOTCHA**: authz must precede `UpdateStatusTx` (the mutation) for the tripwire GUC. SELECT-FOR-UPDATE first is fine (read, not mutate). Remove the now-misplaced authz at line 508.
- **VALIDATE**: `go test ./internal/modules/controlleddocuments/...`.

### Task 4: Enforce area — CD create / preview / create-revision / repo
- **ACTION**: Replace `"tenant"` with the area at: service.go 211 (`cmd.ProcessAreaCode`), 417 (`areaCode` in scope), 607 (`cd.ProcessAreaCode`); repository.go 328/348 (`doc.ProcessAreaCode`).
- **GOTCHA**: `cmd.ProcessAreaCode` already validated active (service.go:155-161). `doc.ProcessAreaCode` is required (domain ctor:75). Fail-closed if empty.
- **VALIDATE**: `go test ./internal/modules/controlleddocuments/...`.

### Task 5: Enforce area — document.create
- **ACTION**: repository.go:93 + 139 (init pointers, same tx) → use `d.ProcessAreaCodeSnapshot`.
- **GOTCHA**: `ProcessAreaCodeSnapshot` is `*string` (nullable). Deref with fail-closed: empty/nil → pass `""` (denies non-system; surfaces area-less docs as a test signal — see Risks). Do NOT silently fall back to `"tenant"`.
- **VALIDATE**: `go test ./internal/modules/documents/...`.

### Task 6: Enforce area — document.edit (~12 sites)
- **ACTION**: For each `authz.Require(...CapDocumentEdit..., "tenant")`, load the document's `process_area_code_snapshot` within the tx before the authz call and pass it.
- **IMPLEMENT**: add `loadDocumentArea(ctx, tx, tenantID, docID) (string, error)` helper (`SELECT process_area_code_snapshot FROM documents WHERE tenant_id=$1 AND id=$2`); for session-keyed ops resolve docID from the session row first. Init-pointers site (139) reuses `d.ProcessAreaCodeSnapshot` (no extra query).
- **GOTCHA**: some edit fns take session id not doc id — derive doc id from `editor_sessions`. Keep authz before the mutating SQL. Empty area → fail-closed.
- **VALIDATE**: `go test ./internal/modules/documents/...` (autosave/session tripwire tests included).

### Task 7: Self-maintaining annotation test
- **ACTION**: In `permissions_authz_scope_test.go`, derive the area-enforced op set from `routeRules` ∩ `IsAreaGrade` instead of the hardcoded `areaEnforcedOps` slice.
- **IMPLEMENT**: iterate the route table (operationId→cap), keep ops whose cap `IsAreaGrade`; assert each carries `x-authz-area` or `x-authz-skip-area`+reason. Annotate any newly-pulled-in document/CD ops with `x-authz-skip-area: true` + reason "area DB-derived/payload-derived, tx-layer enforced (ADR 0022 Phase 7)".
- **GOTCHA**: document/CD area is NOT a uniform request field → `x-authz-skip-area`, not `x-authz-area`. No shape change (vendor extension only). Confirm redocly stays valid.
- **VALIDATE**: `go test ./apps/api/cmd/metaldocs-api/...`; `npx @redocly/cli lint`.

### Task 8: Review fix — unexport BypassSystem + background bridge (CWE-269)
- **ACTION**: Rename `BypassSystem`→unexported `setBypassGUC`; add exported `BypassSystem(ctx, tx)` that returns `ErrBypassNotBackground` unless ctx carries the background marker; add `WithBackgroundBypass(ctx) context.Context`.
- **IMPLEMENT**: `bgBypassKey{}` context key; the 5 background entrypoints (approval cancel/cutover/scheduler run methods; session-reaper cron calling `ExpireStaleSessions`/`DeleteExpiredPending`) call `WithBackgroundBypass` at their composition root. HTTP request contexts never set it → fail-closed.
- **MIRROR**: CONTEXT_MARKER.
- **GOTCHA**: set the marker at the background ROOT, not inside the repo method (circular). Update `authz_test.go` + `repository_autosave_tripwire_test.go` to mark ctx.
- **VALIDATE**: `go build ./...`; `go test ./internal/modules/iam/authz/... ./internal/modules/documents/...`.

### Task 9: Review fix — clock param on MembershipDirectoryScope
- **ACTION**: Add `now time.Time` to `MembershipDirectoryScope`; replace SQL `now()` (line 124) with a bound `$` param; feed `s.nowFn()` from `DirectoryScope`.
- **GOTCHA**: update the interface decl (service.go:30) + the test stub (area_membership_test.go:39).
- **VALIDATE**: `go test ./internal/modules/iam/...`.

### Task 10: Review fix — extract dup system_admin EXISTS SQL (DRY)
- **ACTION**: Define `const systemAdminExistsSQL` (the UNION ALL role/group EXISTS body) in `authz` package; reuse in `authz.go:66-82` and `user_area_repository.go:99-128` (tenant_wide branch).
- **GOTCHA**: param placeholders differ ($1/$2 vs positional in the bigger query) — extract the inner `SELECT 1 ... UNION ALL SELECT 1 ...` text with documented `$actor`/`$tenant` placeholder contract, or expose a small builder. Keep behavior byte-identical.
- **VALIDATE**: `go test ./internal/modules/iam/...`; integration directory + membership-area-scope tests still pass.

### Task 11: Review fix — godoc + golden seed-row-count test
- **ACTION**: Document `no-inline-capability` (and the new guard's) literal-only limitation in its godoc. Add a golden test asserting `0001_product_reference_data.sql` `role_capabilities` INSERT count equals the expected number (so a reformat/accidental drop fails loudly).
- **VALIDATE**: `go test ./scripts/api-lint/...` or `./apps/api/...` (wherever seed parsing lives).

### Task 12: Guard goes green + full gates
- **ACTION**: After Tasks 3-6, rerun the guard — expect 0 `authz-area-scope-binding` violations; total back to 455 baseline.
- **VALIDATE**: full gate suite (below).

### Task 13: ADR Phase 7 completion note + status bump
- **ACTION**: Append Phase 7 completion subsection (per-cap dispositions, guard evidence, review-fix dispositions, gates) to ADR 0022; bump the header status line + `Last verified`.

---

## Testing Strategy

### Unit / Integration
| Test | Input | Expected | Edge? |
|---|---|---|---|
| guard fixture | `Require(...,CapDocumentEdit,"tenant")` | 1 violation | core |
| guard fixture | `Require(...,CapDocumentEdit,areaCode)` | 0 violations | core |
| guard fixture | `Require(...,CapTemplateView,"tenant")` (tenant-grade) | 0 violations | edge |
| CD obsolete | actor area-granted in CD area | success | happy |
| CD obsolete | actor granted only in other area | `ErrCapDenied` | BOLA |
| doc.edit | editor in doc area | success | happy |
| doc.edit | editor in other area | `ErrCapDenied`/403 | BOLA |
| doc.create | area-less doc (nil snapshot) | fail-closed deny (non-system) | edge |
| system_admin | no per-area row, any area | success (bypass, R1) | inheritance |
| DirectoryScope | injected `now` | uses param not wall clock | clock |

### Edge Cases Checklist
- [ ] Empty/nil `ProcessAreaCodeSnapshot` → fail-closed (not "tenant")
- [ ] `system_admin` bypass preserved on every changed site (R1)
- [ ] Background bypass fail-closed when ctx unmarked
- [ ] Variable-cap call sites unaffected by guard (documented skip)

---

## Validation Commands

### Guard bite-then-green
```bash
go run ./scripts/api-lint api/openapi/v1/openapi.yaml .   # before fixes: new violations; after: back to 455
```
EXPECT: pre-fix > 455 (area-grade+"tenant" sites); post-fix == 455.

### Build + tests
```bash
go build ./...
go test ./internal/modules/documents/... ./internal/modules/controlleddocuments/... ./internal/modules/iam/... ./apps/api/... ./scripts/api-lint/... -count=1
```
EXPECT: all pass (note any pre-existing baseline failures explicitly).

### Spec
```bash
npx @redocly/cli lint api/openapi/v1/openapi.yaml
```
EXPECT: valid (no shape change).

### Integration (real DB)
```bash
go test -tags=integration ./tests/integration/... -run "Membership|Document|CD" -count=1
```
EXPECT: area-scoped allow/deny + R1 bypass pass.

---

## Acceptance Criteria
- [ ] `authz-area-scope-binding` guard proven to bite (pre-fix) then green (post-fix)
- [ ] All 5 caps pass real area; `system_admin` bypass intact
- [ ] `areaEnforcedOps` derived from `IsAreaGrade`
- [ ] BypassSystem fail-closed bridge; MembershipDirectoryScope clock param; dup SQL extracted; godoc + golden seed test
- [ ] `go build ./...`, full `go test`, api-lint==455, redocly valid
- [ ] No OpenAPI shape change; ADR Phase 7 note added

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Cross-area authoring in existing tests breaks | Med | Med | Run documents+CD suites; failures = real product signal, fix test seeding to actor's area |
| Area-less documents exist (nil snapshot) | Low-Med | Med | Fail-closed deny surfaces them; if legit, escalate as product question (don't fall back to "tenant") |
| changeStatus reorder breaks tripwire | Low | High | authz stays before the UPDATE; FOR-UPDATE select is read-only |
| Bypass marker missed at a background root | Low | High | Fail-closed → loud error in tests, not silent escalation |
| Dup-SQL extraction changes a query plan | Low | Low | Byte-identical text; integration tests assert behavior |

## Notes
- **HARD-STOP honored**: the document.create/edit ambiguity was surfaced; user ruled Option 1 (area-segregate all). Recorded.
- **Guard limitation** (documented): only literal `"tenant"` + resolvable area-grade **const** cap args are caught; variable-cap sites (e.g. `changeStatus`'s `cap` param) are skipped — acceptable since those are fixed to pass real area and carry no `"tenant"` literal post-fix.
- **Phase 6 (wiki sync) runs AFTER this**, per the ADR resequencing — do not assert "scope enforced" in wiki until Phase 7 merges.
