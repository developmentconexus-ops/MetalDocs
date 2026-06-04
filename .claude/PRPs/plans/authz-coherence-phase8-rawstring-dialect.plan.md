# Plan: ADR 0022 Phase 8 — Close the Raw-String Capability Dialect

## Summary
Four authorization "dialects" exist; three are CI-bound by the SSOT lint. The fourth — raw `string` literal capability arguments to `authz.Require` — is structurally invisible to the lint (`no-inline-capability` bans only the `Capability("…")` conversion, not a bare `"doc.publish"` arg). This phase closes that hole: a new lint rule makes any raw-string cap arg a red build, the 13 surviving raw-string sites are converted to typed registry consts, the 5 phantom (unregistered) caps are registered/classified/seeded, and two fail-open area derivations are corrected to fail-closed.

## User Story
As a MetalDocs security maintainer, I want every tier-2 capability check to reference a typed, registered, seeded capability const enforced by CI, so that a phantom or raw-string cap can never silently rot into a dead-grant or fail-open check again.

## Problem → Solution
- **Current**: 13 `authz.Require` sites pass raw strings; 4 reference real registry caps by string, 5 reference phantom caps absent from the registry (enforced only by the `system_admin` bypass → dead-grant for everyone else); 2 area loaders fail OPEN (empty area → `"tenant"`, area filter silently off).
- **Desired**: lint forbids raw-string cap args; all 13 sites use typed consts; 5 phantoms registered + classified + seeded to the correct roles; area-grade sites fail CLOSED on empty area.

## Metadata
- **Complexity**: Large (cross-module: iam domain, documents, approval, templates, api-lint, db, apps/api)
- **Source PRD**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 8 amendment, lines 193-213)
- **PRD Phase**: Phase 8
- **Estimated Files**: ~16 (4 registry/catalog, 8 call sites, 1 lint + test, 1 migration, 1 seed, 2-3 test goldens, ADR)

---

## UX Design
Internal change — no user-facing UX transformation. Net runtime effect: phantom-gated ops (cancel / edit-draft / reconstruct), previously reachable only by `system_admin` (bypass), become reachable by the roles that already hold the parallel tier-1 cap. Empty-area edge cases flip from fail-open to fail-closed.

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `scripts/api-lint/registry_rules.go` | 93-202 | `checkAuthzAreaScopeBinding` AST walk + `capConstName`; the new rule reuses this machinery |
| P0 | `internal/modules/iam/domain/model.go` | 50-139 | Capability const block + `validCapabilities`; add 4 new consts |
| P0 | `internal/modules/iam/domain/capability_scope.go` | 36-81 | `capabilityScopes` classification map; add 4 entries |
| P0 | `internal/modules/iam/domain/catalog.go` | 104-134 | `capabilityDescriptions`; `CapabilityCatalog()` panics if a const lacks a description |
| P0 | `internal/modules/iam/authz/authz.go` | 76-122 | `Require` SQL — `($2 = 'tenant' OR upa.area_code = $2)`; empty `$2` denies non-system (fail-closed mechanism) |
| P0 | `db/migrations/0225_authz_p2_document_lifecycle_grants.sql` | all | Migration template to mirror (multi-row VALUES, ON CONFLICT, schema_migrations row) |
| P0 | `db/reference-data/0001_product_reference_data.sql` | 17-98 | Single-row seed format required by `seedCapRE`; add new rows here too |
| P1 | `apps/api/cmd/metaldocs-api/permissions.go` | 85-259 | tier-1 routeRules — authority evidence for each cap's scope classification |
| P1 | `internal/modules/documents/approval/application/submit_service.go` | 317-336 | approval `loadDocumentAreaCode` (shared by 6 callers — mind tenant-grade `read_service`) |
| P1 | `internal/modules/documents/application/fillin_authz.go` | 16-53 | documents `loadDocumentAreaCode` (shared by reconstruct) |
| P2 | `wiki/decisions/0018-approval-route-lifecycle.md` | 87-103 | §6 defers tier-1 route cap split — route.admin tier-1 follow-up stays cross-referenced, NOT redesigned here |
| P2 | `scripts/api-lint/registry_rules.go` | 30-36 | `deferredCaps` (empty) — new caps seeded so none deferred |

## External Documentation
No external research needed — feature uses established internal patterns (Go AST lint, capability registry, single-row SQL seed).

---

## Patterns to Mirror

### AST_LINT_RULE
```go
// SOURCE: scripts/api-lint/registry_rules.go:138-175
ast.Inspect(file, func(n ast.Node) bool {
    call, ok := n.(*ast.CallExpr)
    if !ok || len(call.Args) < 4 { return true }
    sel, ok := call.Fun.(*ast.SelectorExpr)
    if !ok || sel.Sel.Name != "Require" { return true }
    pkg, ok := sel.X.(*ast.Ident)
    if !ok || pkg.Name != "authz" { return true }
    // 3rd arg (index 2) is the capability.
    capName, ok := capConstName(call.Args[2]) // unresolvable variable/literal → skip in area rule
    ...
})
```
New rule: same shape, but flag when `call.Args[2]` is `*ast.BasicLit` with `Kind == token.STRING`.

### TYPED_CONST + MAP
```go
// SOURCE: internal/modules/iam/domain/model.go:58, 92  +  capability_scope.go:42  +  catalog.go:110
CapDocumentPublish   Capability = "doc.publish"   // model.go const block
CapDocumentPublish: {},                            // validCapabilities
CapDocumentPublish: ScopeArea,                     // capabilityScopes
CapDocumentPublish: "Publicar documento",          // capabilityDescriptions (pt-BR)
```

### TYPED CALL SITE
```go
// SOURCE: publish_service.go:94  (already typed for document.edit)
if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
```

### SEED — single row (reference-data, parser-bound)
```sql
-- SOURCE: db/reference-data/0001_product_reference_data.sql:35
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
```

### SEED — migration (multi-row, existing DBs)
```sql
-- SOURCE: db/migrations/0225_authz_p2_document_lifecycle_grants.sql:27-39
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('area_admin', 'doc.obsolete', 'Make a document obsolete'),
    ...
ON CONFLICT (role, capability) DO NOTHING;
INSERT INTO public.schema_migrations (version, description) VALUES ('0225', '...') ON CONFLICT (version) DO NOTHING;
```

### FAIL-CLOSED area derivation (Phase 7 precedent)
```go
// Phase 7: empty area → pass "" → Require SQL ($2='tenant' OR area_code=$2) denies non-system.
// NOT a fallback to "tenant".
```

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `scripts/api-lint/registry_rules.go` | UPDATE | Add `checkNoRawStringCapability` rule + wire into `RunRegistryRules` |
| `scripts/api-lint/registry_rules_test.go` | UPDATE | bite / green / skip-variable / skip-tenant-literal-area tests |
| `internal/modules/iam/domain/model.go` | UPDATE | 4 new consts (EditDraft, Reconstruct, WorkflowInstanceCancel, ViewPublished) + validCapabilities |
| `internal/modules/iam/domain/capability_scope.go` | UPDATE | 4 new scope classifications |
| `internal/modules/iam/domain/catalog.go` | UPDATE | 4 new pt-BR descriptions |
| `internal/modules/documents/approval/application/publish_service.go` | UPDATE | `"doc.publish"`×2 → `string(iamdomain.CapDocumentPublish)` |
| `internal/modules/documents/approval/application/obsolete_service.go` | UPDATE | `"doc.obsolete"` → `string(iamdomain.CapDocumentObsolete)` |
| `internal/modules/templates/application/lifecycle.go` | UPDATE | `"template.archive"` → `string(iamdomain.CapTemplateArchive)` (tenant-grade, keep `"tenant"`) |
| `internal/modules/documents/approval/application/route_admin_service.go` | UPDATE | `"route.admin"`×4 → `string(iamdomain.CapRouteManage)` |
| `internal/modules/documents/approval/http/route_admin_handler.go` | UPDATE | Remove dead `const CapManageRoutes = "route.admin"` |
| `internal/modules/documents/approval/application/cancel_service.go` | UPDATE | `"workflow.instance.cancel"` → `string(iamdomain.CapWorkflowInstanceCancel)` |
| `internal/modules/documents/application/fillin_authz.go` | UPDATE | `"doc.edit_draft"` → const; helper fail-closed (`"tenant"`→`""`) |
| `internal/modules/documents/application/reconstruct_service.go` | UPDATE | `"doc.reconstruct"` → const (benefits from helper fail-closed) |
| `internal/modules/documents/application/view_service.go` | UPDATE | `"doc.view_published"` → const; tenant-grade → pass `"tenant"` |
| `internal/modules/documents/approval/application/submit_service.go` | UPDATE | `loadDocumentAreaCode` fail-closed (`"tenant"`→`""`) |
| `internal/modules/documents/approval/application/read_service.go` | UPDATE | explicit `""`→`"tenant"` coalesce (document.view tenant-grade — preserve behavior) |
| `db/migrations/0226_authz_p8_phantom_cap_grants.sql` | CREATE | Seed 4 new caps to existing DBs |
| `db/reference-data/0001_product_reference_data.sql` | UPDATE | Mirror new grants (single-row) for fresh bootstrap + lint parity |
| `apps/api/cmd/metaldocs-api/permissions_test.go` | UPDATE | golden row count (94→120); `TestEveryCapSeededOrDeferred` covers new caps |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | Phase 8 completion note |

## NOT Building
- **Tier-1 route cap split** (route.admin / CapRouteView / read-write split at permissions.go) — ADR 0018 §6 owns this and explicitly defers to the F-001 follow-up. Open follow-up (b) (approval-route tier-1 = `document.submit`) stays cross-referenced, NOT changed.
- **forceReleaseDocumentSession tier-1 alignment** — NOTE: original task scope item 5 references this; re-evaluate during implementation (see Task 12). Not a phantom/raw-string item.
- **No OpenAPI shape change** — descriptions + x-authz-* metadata only (none required here; no spec edits).
- **No reclassification of existing caps** — only the 4 NEW phantom caps get classified.

---

## Step-by-Step Tasks

### Task 1: New lint rule `no-rawstring-capability`
- **ACTION**: Add `checkNoRawStringCapability(modulesRoot, fset)` to `registry_rules.go`; call it in `RunRegistryRules`.
- **IMPLEMENT**: Walk non-test `.go` (skip `.git/.claude/node_modules/vendor`). For each `*ast.CallExpr` whose `Fun` is `authz.Require` (≥3 args) OR `authz.RequireAll`, inspect the capability arg (`Args[2]` for Require). If it is `*ast.BasicLit` with `Kind == token.STRING` → emit Violation rule `no-rawstring-capability`, message: `raw-string capability literal <lit> passed to authz.Require; reference a typed const from internal/modules/iam/domain (single source of truth, ADR 0022 Phase 8)`.
- **MIRROR**: AST_LINT_RULE pattern (registry_rules.go:138-175).
- **GOTCHA**: `lifecycle.go:555` passes a VARIABLE `cap` (`*ast.Ident`, not `BasicLit`) → must NOT flag. `string("x")` conversion is a `*ast.CallExpr`, not BasicLit → already covered by `no-inline-capability`; do not double-flag — only flag a bare BasicLit. For `RequireAll` the cap arg index differs (testdata `handler.go:17` passes a non-literal) — guard arg count; flag only BasicLit string args. Skip `model.go` exemption not needed (consts are not Require calls).
- **VALIDATE**: temporarily run `go run ./scripts/api-lint` BEFORE converting sites → expect **13** `no-rawstring-capability` violations at the exact known lines.

### Task 2: Register 4 new capability consts
- **ACTION**: Add to `model.go` const block + `validCapabilities`.
- **IMPLEMENT**:
  ```go
  CapDocumentEditDraft       Capability = "doc.edit_draft"
  CapDocumentReconstruct     Capability = "doc.reconstruct"
  CapDocumentViewPublished   Capability = "doc.view_published"
  CapWorkflowInstanceCancel  Capability = "workflow.instance.cancel"
  ```
- **MIRROR**: TYPED_CONST + MAP.
- **GOTCHA**: every const MUST be added to `validCapabilities` too or `IsValidCapability` returns false and seed-parity lint fails. `route.admin` is NOT added — it reconciles to existing `CapRouteManage`.
- **VALIDATE**: `go build ./internal/modules/iam/...`.

### Task 3: Classify scope
- **ACTION**: Add to `capabilityScopes` in `capability_scope.go`.
- **IMPLEMENT**: `CapDocumentEditDraft: ScopeArea`, `CapDocumentReconstruct: ScopeArea`, `CapWorkflowInstanceCancel: ScopeArea`, `CapDocumentViewPublished: ScopeTenant`.
- **GOTCHA**: `TestEveryCapabilityClassified` asserts the map covers the registry exactly — all 4 must be present or it fails. Rationale per cap documented inline (tier-1 evidence: edit_draft/reconstruct/cancel → `document.edit`; view_published → `document.view`, registry rule "all *.view reads tenant-grade").
- **VALIDATE**: `go test ./internal/modules/iam/domain/... -run TestEveryCapabilityClassified -count=1`.

### Task 4: Descriptions (pt-BR)
- **ACTION**: Add to `capabilityDescriptions` in `catalog.go`.
- **IMPLEMENT**: `CapDocumentEditDraft: "Editar rascunho de documento"`, `CapDocumentReconstruct: "Reconstruir documento a partir de revisão"`, `CapDocumentViewPublished: "Visualizar documento publicado"`, `CapWorkflowInstanceCancel: "Cancelar instância de workflow"`.
- **GOTCHA**: `CapabilityCatalog()` panics if any const lacks a description.
- **VALIDATE**: `go test ./internal/modules/iam/domain/... -count=1`.

### Task 5: Convert 4 registry-cap sites to typed consts
- **ACTION**: publish_service.go:88,234; obsolete_service.go:79; lifecycle.go:521.
- **IMPLEMENT**: `"doc.publish"`→`string(iamdomain.CapDocumentPublish)`, `"doc.obsolete"`→`string(iamdomain.CapDocumentObsolete)`, `"template.archive"`→`string(iamdomain.CapTemplateArchive)`.
- **GOTCHA**: publish/obsolete are area-grade and already pass real `areaCode` → `authz-area-scope-binding` rule satisfied. template.archive is tenant-grade → keep `"tenant"` (rule ignores tenant-grade). Confirm `iamdomain` already imported in each file (publish_service already uses `iamdomain.CapDocumentEdit`; obsolete/lifecycle — verify import).
- **VALIDATE**: `go build ./...`.

### Task 6: Reconcile route.admin → CapRouteManage
- **ACTION**: route_admin_service.go:166,297,461,552 → `string(iamdomain.CapRouteManage), "tenant"`. Remove dead `const CapManageRoutes = "route.admin"` (route_admin_handler.go:15, zero refs).
- **GOTCHA**: `CapRouteManage` is tenant-grade + already seeded (qms_admin, system_admin) → no seed change, no area-rule bite. Add `iamdomain` import to route_admin_service.go if absent. Tier-1 mapping unchanged (ADR 0018 §6 deferral).
- **VALIDATE**: `go build ./...`; grep confirms no `route.admin` / `CapManageRoutes` literal remains.

### Task 7: Convert 4 phantom area/tenant call sites
- **ACTION**: cancel_service.go:109 → `string(iamdomain.CapWorkflowInstanceCancel), areaCode.String`; fillin_authz.go:32 → `string(iamdomain.CapDocumentEditDraft), areaCode`; reconstruct_service.go:41 → `string(iamdomain.CapDocumentReconstruct), areaCode`; view_service.go:74 → `string(iamdomain.CapDocumentViewPublished), "tenant"`.
- **GOTCHA**: view_published is tenant-grade → pass `"tenant"` and drop the now-unused `area` local in view_service (compile error otherwise). Confirm `iamdomain` import in each file.
- **VALIDATE**: `go build ./...`.

### Task 8: Fail-closed — documents `loadDocumentAreaCode`
- **ACTION**: fillin_authz.go:35 helper — change both `return "tenant", nil` (no-rows at :45, empty at :50) to `return "", nil`. Update the doc comment (drop "area 'tenant' is used").
- **GOTCHA**: shared ONLY by fillin_authz (edit_draft, area-grade) + reconstruct_service (reconstruct, area-grade) — both want fail-closed. Empty `""` → `Require` denies non-system, `system_admin` still bypasses.
- **VALIDATE**: `go test ./internal/modules/documents/application/... -count=1`.

### Task 9: Fail-closed — approval `loadDocumentAreaCode` + read_service guard
- **ACTION**: submit_service.go:317 helper — `"tenant"`→`""` on no-rows (:328) and empty (:333). At read_service.go:98, add explicit `if areaCode == "" { areaCode = "tenant" }` BEFORE the `CapDocumentView` Require (tenant-grade — preserve current empty→tenant behavior).
- **GOTCHA**: 6 callers — decision(signoff)/publish×2/submit/supersede are area-grade → fail-closed correct; **read(document.view) is tenant-grade** → the coalesce keeps it behavior-identical. Do NOT skip the read_service guard or empty-area reads break for non-system viewers.
- **VALIDATE**: `go test ./internal/modules/documents/approval/... -count=1`.

### Task 10: Seed migration 0226
- **ACTION**: Create `db/migrations/0226_authz_p8_phantom_cap_grants.sql`.
- **IMPLEMENT**: multi-row `INSERT ... ON CONFLICT (role, capability) DO NOTHING` granting:
  - `doc.edit_draft`, `doc.reconstruct`, `workflow.instance.cancel` → approver, area_admin, author, editor, qms_admin, system_admin (mirror `document.edit` set).
  - `doc.view_published` → approver, area_admin, author, editor, qms_admin, signer, system_admin, viewer (mirror `document.view` set).
  - schema_migrations row `('0226', 'ADR 0022 Phase 8: seed phantom caps doc.edit_draft/doc.reconstruct/workflow.instance.cancel/doc.view_published')`.
- **MIRROR**: 0225 migration template.
- **GOTCHA**: `route.manage` already seeded — do NOT re-grant. Header comment must explain the document.edit/document.view mirror rationale ("derived from existing seed parallels, not guessed").
- **VALIDATE**: applies on fresh bootstrap (`scripts/start-api.ps1 -Build` against clean DB, or migration dry-run per metaldocs-database skill).

### Task 11: Mirror grants into reference-data + update goldens
- **ACTION**: Add 26 single-row INSERTs to `db/reference-data/0001_product_reference_data.sql` (matching Task 10 matrix). Update golden row count in `permissions_test.go` (94→120). Verify `TestEveryCapSeededOrDeferred` passes (all 4 new caps now seeded; `route.manage` already seeded).
- **MIRROR**: SEED single-row pattern.
- **GOTCHA**: `seedCapRE` only parses SINGLE-row `VALUES ('role','cap',...` — multi-row VALUES in reference-data would break the parser + the golden. Keep one INSERT per row. Count must be EXACT (26 new = 6+6+6+8).
- **VALIDATE**: `go run ./scripts/api-lint` → 0 `seed-registry-parity` violations; `go test ./apps/api/... -run 'TestSeedRowCount|TestEveryCapSeededOrDeferred' -count=1`.

### Task 12: forceReleaseDocumentSession tier-1 (scope item 5a)
- **ACTION**: Evaluate `permissions.go:175` (`/session/force-release` → `CapMembershipManage`) vs its tier-2 `document.edit`. Either align tier-1 to `CapDocumentEdit` (1-line, matches tier-2 AND the generic `/session/` row at :176) OR log as explicit defect with rationale.
- **GOTCHA**: This is NOT a raw-string/phantom item; it is a cross-tier coherence follow-up. Aligning is low-risk (the generic session row already uses CapDocumentEdit). If aligning changes a passing test that encodes membership.manage intent → surface, do not force. Decide during implementation; document the choice in the ADR note.
- **VALIDATE**: `go test ./apps/api/... -count=1`.

### Task 13: Lint green + full gates
- **ACTION**: Re-run all gates; the new rule must now report **0**.
- **VALIDATE**: see Validation Commands.

### Task 14: ADR Phase 8 completion note
- **ACTION**: Append a "Phase 8 — COMPLETE" subsection to `wiki/decisions/0022-authz-capability-coherence.md` (mirror the Phase 7 completion-note style): rule bite→green, per-cap rulings table, seed matrix, fail-closed fixes, gates, branch, open follow-ups (tier-1 route cap → ADR 0018 §6; forceRelease disposition).

---

## Testing Strategy

### Unit Tests
| Test | Input | Expected | Edge? |
|---|---|---|---|
| `no-rawstring-capability` bite | fixture `authz.Require(ctx,tx,"doc.x",a)` | 1 violation | — |
| green | `authz.Require(ctx,tx,string(Cap...),a)` | 0 | — |
| skip variable cap | `authz.Require(ctx,tx,cap,"tenant")` | 0 | ✓ (lifecycle.go:555) |
| skip inline-conv (owned by other rule) | `authz.Require(ctx,tx,Capability("x"),a)` | 0 from THIS rule | ✓ |
| RequireAll non-literal | `authz.RequireAll(x)` | 0 | ✓ |
| Classified map | new caps | `TestEveryCapabilityClassified` pass | — |
| Seed parity | reference-data | 0 violations, count=120 | — |

### Edge Cases Checklist
- [x] Empty/missing area → fail-closed (deny non-system) for area-grade caps
- [x] Empty area → still `"tenant"` for read_service (document.view tenant-grade)
- [x] Variable cap arg not flagged (lifecycle.go:555)
- [x] Permission denied path: phantom-seeded roles now pass tier-2
- [x] system_admin bypass still works (not explicitly required, but seeded for parity)

---

## Validation Commands

### Lint rule (bite → green)
```powershell
# BEFORE converting sites (Task 1 done, Tasks 5-7 not):
go run ./scripts/api-lint   # EXPECT 13 no-rawstring-capability violations
# AFTER all conversions:
go run ./scripts/api-lint   # EXPECT 0 no-rawstring-capability; total delta vs 455 baseline = 0
```

### Build
```powershell
go build ./...   # EXPECT clean
```

### Tests
```powershell
go test ./internal/modules/documents/... ./internal/modules/controlleddocuments/... ./internal/modules/templates/... ./internal/modules/iam/... ./apps/api/... ./scripts/api-lint/... -count=1
```
EXPECT pass. KNOWN pre-existing failure baseline `426cc00dc`: `TestListDocumentsPaginated` (Scan-arg drift, unrelated) — confirm identical on base before/after.

### OpenAPI
```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml   # EXPECT valid (no shape change)
```

### Migration (fresh bootstrap)
```powershell
.\scripts\start-api.ps1 -Build   # against clean DB → 0226 applies; or metaldocs-database dry-run
```

### Manual Validation
- [ ] `git grep -nE 'authz\.Require(All)?\(' -- '*.go' | grep -v _test | grep -E '"[a-z]'` → only the variable-cap `lifecycle.go:555` and typed `string(iamdomain...)` remain; ZERO bare string cap literals.
- [ ] `git grep -n 'route.admin\|CapManageRoutes'` → no Go literal (ADR text references allowed).
- [ ] Security review + code review on the diff.

---

## Acceptance Criteria
- [ ] New lint rule proven bite(13) → green(0)
- [ ] `go build ./...` clean
- [ ] Targeted suites pass (modulo known TestListDocumentsPaginated baseline)
- [ ] `go run ./scripts/api-lint` total = 455 (no net new violations at green)
- [ ] redocly lint valid
- [ ] 0226 applies on fresh bootstrap
- [ ] ADR Phase 8 note added

## Completion Checklist
- [ ] All 13 sites typed
- [ ] 4 phantoms registered/classified/described/seeded; route.admin reconciled to route.manage
- [ ] Fail-open → fail-closed (both helpers; read_service preserved)
- [ ] Goldens updated (120)
- [ ] No OpenAPI shape change
- [ ] Security + code review done

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Seeding phantoms broadens access (system_admin-only → document.edit/view holders) | High (intended) | Med | Verify approval/documents suites don't encode the phantom restriction as intended; surface any that do before merge |
| read_service empty-area regression if helper fail-closed without guard | Med | High | Explicit `""`→`"tenant"` coalesce at read_service (Task 9) |
| Golden count off-by-N | Med | Low | Compute 26 = 6+6+6+8 explicitly; rerun TestSeedRowCount |
| Multi-row VALUES in reference-data breaks seedCapRE | Low | High | Single-row only in reference-data; multi-row only in migration |
| route.admin→route.manage grants qms_admin route mgmt at tier-2 | High (intended, registry-faithful) | Low | route.manage already seeded to qms_admin in registry; tier-1 still gates; run approval route suite |
| forceRelease tier-1 change breaks a membership.manage test | Low | Med | Evaluate; align only if green, else log defect |

## Notes
- **Why register doc.view_published / doc.reconstruct instead of reusing document.view/edit**: ADR Phase 8 explicitly mandates registering all 5 phantom strings. Registering (not collapsing) is the faithful, minimal "close the dialect" move — each raw string becomes a typed registered cap with no call-site semantic rewrite.
- **Fail-closed mechanism**: `authz.Require` SQL `($2 = 'tenant' OR upa.area_code = $2)` — empty `$2` matches neither branch (`area_code` is never `''`) → not granted → denied; `system_admin` EXISTS bypass still short-circuits.
- **Tripwire pairing**: all conversions keep `authz.Require` before the mutating SQL (unchanged order).
