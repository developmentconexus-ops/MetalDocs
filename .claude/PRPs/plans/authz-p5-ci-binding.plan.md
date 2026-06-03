# Plan: ADR 0022 Phase 5 — Bind the Authz Capability Model in CI

## Summary
Turn every silent authz-model drift surfaced in ADR 0022 Finding 3 into a red build. Add deterministic, Windows-runnable parity checks (api-lint binary rules + Go tests) binding the Go capability registry ↔ route table ↔ DB seed ↔ OpenAPI annotations ↔ wiki, reconcile the stale OpenAPI access-policy enum (document-as-separate-vocabulary, not prune), and verify/report the two RLS-hardening criteria (transaction-local GUCs; native-RLS-vs-trigger-tripwire).

## User Story
As a MetalDocs maintainer, I want capability-model drift (a route cap, seed grant, annotation, or wiki claim that diverges from the typed Go registry) to fail CI, so that the authz model stays one coherent system instead of rotting silently between five sources of truth.

## Problem → Solution
Five sources of truth (Go registry, route table, DB seed, OpenAPI annotations, wiki) agree today only by discipline; Phase 1–4 fixed the known drift but nothing prevents new drift → CI mechanically enforces agreement so the next `membership.grant`-style typo is a red build, not latent rot.

## Metadata
- **Complexity**: Large
- **Source PRD**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 5)
- **PRD Phase**: Phase 5 — Bind the sources in CI
- **Estimated Files**: ~10 (3 new lint/test files, 4 edits, 2 docs, 1 ADR)

---

## UX Design
Internal change — no user-facing UX transformation. The "user" is the CI pipeline and the next engineer who introduces drift.

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `scripts/api-lint/main.go` | all | Binary entry: `<spec> [<modulesRoot>]`, prints `file:line: RULE: msg`, exits 1 if any violation |
| P0 | `scripts/api-lint/code_rules.go` | 1-101, 294-377 | `RunCodeRules` walk pattern, `indexModuleFuncs` AST walk, `tripwire-pairing` — mirror for new AST/file rules |
| P0 | `scripts/api-lint/spec_rules.go` | 13-56, 124-139, 158-175 | `Violation` struct, `RunSpecRules` op walk, `checkAuthz` state-transition annotation rule, `mapGet`/`scalarValue` yaml helpers |
| P0 | `internal/modules/iam/domain/model.go` | 50-139 | Registry: `Capability` consts, `validCapabilities`, exported `AllCapabilities()`/`IsValidCapability()` |
| P0 | `internal/modules/iam/domain/capability_scope.go` | all | `capabilityScopes`, exported `IsAreaGrade()`/`ScopeOf()` — the area-grade classification to bind |
| P0 | `apps/api/cmd/metaldocs-api/permissions.go` | 24-55, 85-259 | `routeRule` struct + `routeRules` table (cap per route) — item-1 Go test reads this |
| P0 | `apps/api/cmd/metaldocs-api/permissions_test.go` | 487-565 | `TestEveryRouteCapInRegistry`, `seededCaps()` SQL regex, `TestEveryCapSeededOrDeferred` — mirror + generalize |
| P1 | `internal/modules/iam/authz/context.go` | 47-69 | `SeedTxIdentity` — transaction-local `set_config(...,true)` (item 6 verify + regression test) |
| P1 | `internal/modules/iam/authz/authz.go` | 51-118, 152-190 | tier-2 `Require`, `asserted_caps` GUC, `bypass_authz`; system_admin bypass |
| P1 | `db/reference-data/0001_product_reference_data.sql` | 18-22 | `role_capabilities` INSERT grammar the seed-parity rule parses |
| P1 | `db/baseline/0001_current_schema.sql` | 490-560, 3875-4050 | `enforce_capability_asserted()` trigger + `trg_require_cap_asserted` on sensitive tables — the actual isolation model (item 7) |
| P1 | `api/openapi/v1/openapi.yaml` | 5898-5932 | `AccessPolicyItem`/`AccessPolicyWriteItem` capability enums (item 5) |
| P2 | `wiki/concepts/authz-tiers.md` | all | Item-4 target doc; contains illustrative `doc.create` + deferred `route.view` (must NOT be flagged) |
| P2 | `wiki/references/local-dev-credentials.md` | role tables | Item-4 target; Phase-1-fixed `membership.grant`→`membership.manage` is the drift class to lock |
| P2 | `wiki/references/authz-industry-evidence.md` | §3 | GUC/RLS footgun source for items 6-7 |

## External Documentation
No external research needed — feature uses established internal patterns (api-lint AST/yaml walks, Go table tests, sqlmock). RLS footgun references already captured in `wiki/references/authz-industry-evidence.md` §3.

---

## Patterns to Mirror

### VIOLATION_EMIT
```go
// SOURCE: scripts/api-lint/code_rules.go:321-328
out = append(out, Violation{
	File:    path,
	Line:    fset.Position(fn.Pos()).Line,
	Rule:    "tripwire-pairing",
	Message: fmt.Sprintf("mutating SQL in %s without authz.Require call", fn.Name.Name),
})
```

### AST_WALK_MODULES
```go
// SOURCE: scripts/api-lint/code_rules.go:77-100  (filepath.WalkDir, skip dirs + _test.go, parser.ParseFile)
if !strings.HasSuffix(strings.ToLower(path), ".go") || strings.HasSuffix(strings.ToLower(path), "_test.go") {
	return nil
}
file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
```

### SEED_SQL_PARSE
```go
// SOURCE: apps/api/cmd/metaldocs-api/permissions_test.go:524-534
re := regexp.MustCompile(`(?i)INSERT INTO\s+metaldocs\.role_capabilities[^V]*VALUES\s*\(\s*'[^']*'\s*,\s*'([^']+)'`)
matches := re.FindAllStringSubmatch(string(raw), -1)
// m[1] = capability value
```

### REGISTRY_BIND_TEST
```go
// SOURCE: apps/api/cmd/metaldocs-api/permissions_test.go:491-503
for i, r := range routeRules {
	if r.capability == "" { continue }
	if !iamdomain.IsValidCapability(r.capability) {
		t.Errorf("routeRules[%d]: capability %q is not in the registry ...", i, r.capability)
	}
}
```

### YAML_OP_WALK
```go
// SOURCE: scripts/api-lint/spec_rules.go:41-54
for i := 0; i+1 < len(paths.Content); i += 2 {
	pathKey, pathVal := paths.Content[i], paths.Content[i+1]
	for j := 0; j+1 < len(pathVal.Content); j += 2 {
		mKey, op := pathVal.Content[j], pathVal.Content[j+1]
		opID := scalarValue(mapGet(op, "operationId"))
	}
}
```

### SQLMOCK_GUC_TEST
```go
// SOURCE: internal/modules/taxonomy/infrastructure/authz_guc_test.go:220-223
mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.tenant_id', $1, true)")).
	WithArgs(tenantID).WillReturnResult(...)
```

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `scripts/api-lint/registry_rules.go` | CREATE | New binary rules: `no-inline-capability` (item 2), `seed-registry-parity` (item 3), `wiki-capability-parity` (item 4). Imports `iam/domain`. |
| `scripts/api-lint/registry_rules_test.go` | CREATE | Unit tests for the new rules incl. injected-drift cases (prove they bite without manual revert dance). |
| `scripts/api-lint/code_rules.go` | UPDATE | Call new code-side rules from `RunCodeRules`; thread `modulesRoot` to SQL/wiki paths. |
| `apps/api/cmd/metaldocs-api/permissions_authz_scope_test.go` | CREATE | Item 1: every area-grade routeRule write op ↔ spec `x-authz-area`/`x-authz-skip-area`. |
| `apps/api/cmd/metaldocs-api/permissions_test.go` | UPDATE | Add `TestEveryCapInRegistryFromRoutes`-style two-way seed parity note pointer (keep existing tests). |
| `internal/modules/iam/authz/context_test.go` | CREATE or UPDATE | Item 6 regression: `SeedTxIdentity` emits `set_config(...,true)` (transaction-local) for both GUCs. |
| `wiki/concepts/authz-tiers.md` | UPDATE | Item 4: add `cap:` marker convention doc + mark real enforcement-claim caps. |
| `wiki/references/local-dev-credentials.md` | UPDATE | Item 4: mark role→capability claims with `cap:` so the parity rule binds them. |
| `api/openapi/v1/openapi.yaml` | UPDATE | Item 5: add `description`/comment on `AccessPolicyItem`/`AccessPolicyWriteItem` capability enums noting separate document-ACL vocabulary. NO shape change. |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | Mark Phase 5 complete; reclassify Finding-3 OpenAPI row; add items 6-7 RLS findings to amendment. |
| `wiki/references/authz-industry-evidence.md` | UPDATE (optional) | Cross-link the trigger-tripwire-vs-native-RLS conclusion (item 7). |

## NOT Building
- **Native Postgres RLS** (`ENABLE`/`FORCE ROW LEVEL SECURITY` + policies). The codebase has zero native RLS; isolation is a `BEFORE INSERT/UPDATE/DELETE` trigger (`enforce_capability_asserted`) reading the transaction-local `asserted_caps` GUC. Adding native RLS is a storage-architecture redesign → CLAUDE.md hard-stop. Item 7 = **report only**.
- **Pruning the 3 `AccessPolicyItem` caps** (`document.upload_attachment`, `document.change_workflow`, `document.manage_permissions`). They are NOT confirmable-unused: backed by a live DB CHECK constraint (`document_access_policies`, baseline:1074 + migration 0011) and FE generated types. Pruning = cross-layer contract-shape change on a dead feature. Item 5 = **document as separate vocabulary**.
- **Activating `authz-call-present` for tx-layer enforcement** (ADR Phase 5 bullet 2). The rewrite to recognize DB-derived area sources is a separate, large lint-engine change; the parity bindings here deliver the "make drift a red build" goal without it. Flag as residual Phase 5 work / defer. *(Decision point — see Risks.)*
- **Route-consumption parity** for the `route.manage` orphan grant (ADR 0018 deferral). Out of scope; unrelated to registry/seed/wiki parity.
- Changing any OpenAPI **shape** (paths, params, schemas, enum membership).

---

## Step-by-Step Tasks

### Task 1: Item 6 — verify + lock transaction-local identity GUCs
- **ACTION**: Verified by reading: ALL `set_config('metaldocs.*', …)` call sites pass the 3rd arg `true` (transaction-local). `SeedTxIdentity` (context.go:60-61), `asserted_caps` (authz.go:188), `bypass_authz`, and every module `authz_guc.go`. Grep for session-level (`set_config(` without `, true)`) returns NONE. R6 SATISFIED.
- **IMPLEMENT**: Add/confirm a regression unit test on `SeedTxIdentity` asserting both `set_config('metaldocs.tenant_id',$1,true)` and `set_config('metaldocs.actor_id',$2,true)` are emitted (single combined statement). Use sqlmock; mirror SQLMOCK_GUC_TEST. NOTE the real statement is a single `SELECT set_config(...), set_config(...)` (context.go:59-62) — match the combined form, not two separate ExpectExec.
- **MIRROR**: SQLMOCK_GUC_TEST
- **VALIDATE**: `go test ./internal/modules/iam/authz/... -run SeedTxIdentity -count=1`

### Task 2: Item 2 — `no-inline-capability` lint rule
- **ACTION**: Add a code-side rule banning `Capability("literal")` / `iamdomain.Capability("literal")` / `domain.Capability("literal")` string-conversion expressions outside `model.go` and `_test.go`.
- **IMPLEMENT**: In `registry_rules.go`, `checkNoInlineCapability(modulesRoot, fset) []Violation`: WalkDir like AST_WALK_MODULES; skip `_test.go` and `domain/model.go`; `ast.Inspect` for `*ast.CallExpr` whose `Fun` is `Ident{Name:"Capability"}` OR `SelectorExpr{Sel:"Capability"}` with exactly one arg that is a `*ast.BasicLit{Kind:STRING}`. Emit `no-inline-capability` violation. (Const decls `Capability = "x"` in model.go are assignments, not conversions — won't match; safe even if model.go weren't skipped, but skip for clarity.)
- **MIRROR**: AST_WALK_MODULES, VIOLATION_EMIT
- **IMPORTS**: `go/ast`, `go/token`, `path/filepath`, `strings`
- **GOTCHA**: Don't flag `MustCapability("x")` (different func name) or `Capability(var)` (non-literal). Only `Capability(<stringlit>)`.
- **VALIDATE**: clean tree → 0 hits. Inject `iamdomain.Capability("doc.bogus")` into permissions.go → rule fires.

### Task 3: Item 3 — `seed-registry-parity` lint rule (in binary)
- **ACTION**: Bind DB seed ↔ Go registry both directions in the api-lint binary (not just the existing unit test).
- **IMPLEMENT**: In `registry_rules.go`, `checkSeedRegistryParity(modulesRoot) []Violation`: read `<modulesRoot>/db/reference-data/0001_product_reference_data.sql`; parse caps with SEED_SQL_PARSE regex. (a) every seeded cap MUST be `iamdomain.IsValidCapability` else `seed-registry-parity: seeded cap %q not in registry`; (b) every `iamdomain.AllCapabilities()` cap MUST be seeded OR in a documented `deferredCaps` set (currently empty, mirror TestEveryCapSeededOrDeferred) else `seed-registry-parity: registry cap %q seeded to no role`. File/Line = seed file + matched line (compute line by counting `\n` up to match index) or line 0.
- **MIRROR**: SEED_SQL_PARSE, REGISTRY_BIND_TEST, VIOLATION_EMIT
- **IMPORTS**: `metaldocs/internal/modules/iam/domain`, `os`, `regexp`, `sort`, `strings`
- **GOTCHA**: `AllCapabilities()` map order is random — sort before emit for deterministic output (tests compare stable order). Path join must be OS-agnostic (`filepath.Join`) for Windows.
- **VALIDATE**: clean → 0. Inject a seed row `('viewer','doc.bogus',…)` → (a) fires. Comment out a registry cap's only seed row → (b) fires.

### Task 4: Item 4 — `wiki-capability-parity` lint rule (marker convention)
- **ACTION**: Bind capability names *claimed as enforced* in wiki authz docs ↔ registry, WITHOUT false-positiving on prose/filenames/GUCs/deferred caps.
- **IMPLEMENT**: Convention: an enforcement-claim capability reference is written `` `cap:<name>` `` (e.g. `` `cap:membership.manage` ``). In `registry_rules.go`, `checkWikiCapabilityParity(modulesRoot) []Violation`: scan a fixed list of authz docs (`wiki/concepts/authz-tiers.md`, `wiki/references/local-dev-credentials.md`, `wiki/modules/iam.md`); regex `` `cap:([a-z][a-z0-9_]*\.[a-z0-9_.]+)` ``; every captured `<name>` MUST be `iamdomain.IsValidCapability` else `wiki-capability-parity: wiki cap %q (file:line) not in registry`. Compute line by newline count. Unmarked dotted tokens (illustrative `doc.create`, deferred `route.view`, `permissions.go`, `metaldocs.actor_id`) are ignored by design.
- **MIRROR**: VIOLATION_EMIT
- **IMPORTS**: `os`, `regexp`, `metaldocs/internal/modules/iam/domain`
- **GOTCHA**: Retrofit the marker ONLY onto genuine enforcement claims (Task 7). Do NOT mark illustrative/deferred mentions. The rule is meaningless until ≥1 marker exists, so Task 7 must land with this.
- **VALIDATE**: after Task 7 markers exist → 0. Inject `` `cap:membership.grant` `` into local-dev-credentials.md → fires.

### Task 5: Wire new code-side rules into the binary
- **ACTION**: Call the three new rules from `RunCodeRules` so `go run ./scripts/api-lint <spec> .` runs them.
- **IMPLEMENT**: In `code_rules.go` `RunCodeRules`, after `tripwire`, append `checkNoInlineCapability(modulesRoot, fset)`, `checkSeedRegistryParity(modulesRoot)`, `checkWikiCapabilityParity(modulesRoot)`. They only run when `modulesRoot != ""` (already gated). Keep determinism: rules append in fixed order.
- **MIRROR**: code_rules.go:61-66
- **GOTCHA**: Binary currently exits 1 on the 455 pre-existing dormant violations; new rules are clean-state-zero, so they don't change the baseline count. "Biting" is demonstrated by the specific violation line appearing + count 455→456 on injected drift. Document this in the gate evidence.
- **VALIDATE**: `go run ./scripts/api-lint api/openapi/v1/openapi.yaml .` → still 455 (zero new). `go build ./scripts/api-lint`.

### Task 6: Item 1 — area-grade annotation parity (Go test, routeRules-aware)
- **ACTION**: Every area-grade capability's routed *write* operation must carry `x-authz-area` or `x-authz-skip-area` in the spec.
- **IMPLEMENT**: `permissions_authz_scope_test.go` (package main, has `routeRules`): build write-verb routeRules where `iamdomain.IsAreaGrade(r.capability)`. Load `api/openapi/v1/openapi.yaml` (locate via `runtime.Caller` → repo root, mirror `seededCaps` path math) with `gopkg.in/yaml.v3`. For each spec operation collect `(method, specPath)` + presence of `x-authz-area`/`x-authz-skip-area`. Match a routeRule to spec ops: strip `/api/v1` server prefix; treat spec `{param}` segments as wildcards; apply the routeRule's `pathExact`/`pathPrefix`/`pathSuffix`/`contains`/`notSuffix` semantics (reuse the matching logic). Assert ≥1 spec op matches AND every matched mutating op is annotated; else `t.Errorf`. Skip non-mutating GET rules.
- **MIRROR**: REGISTRY_BIND_TEST, YAML_OP_WALK, `seededCaps` path resolution (permissions_test.go:512-519)
- **IMPORTS**: `gopkg.in/yaml.v3`, `iamdomain`, `runtime`, `path/filepath`, `os`, `net/http`
- **GOTCHA**: Area-grade caps today: document.{create,edit,submit,signoff}, doc.{publish,obsolete,supersede}, controlled_documents.{create,obsolete,supersede}, membership.manage. Several map to routes whose spec ops are already `x-authz-skip-area`-annotated (Phase 2/3). `document.create`/`controlled_documents.*` create routes may NOT yet be annotated (Phase-2-noted runtime gap) — if the test flags them, that is a TRUE finding: either annotate the spec op `x-authz-skip-area` with a documented reason (matching Phase 2 precedent) or scope the test to *annotated-or-skip-required* with an explicit allow-set. Decide per finding; do not silently exclude.
- **VALIDATE**: `go test ./apps/api/cmd/metaldocs-api/ -run AuthzScope -count=1`. Remove an `x-authz-skip-area` from `grantAreaMembership` → red.

### Task 7: Item 4 doc — marker convention + retrofit
- **ACTION**: Document the `cap:` marker convention and mark genuine enforcement-claim caps in the scanned docs.
- **IMPLEMENT**: In `authz-tiers.md` add a short "Capability references" note: enforcement claims use `` `cap:<name>` ``; bound to the registry by `scripts/api-lint` `wiki-capability-parity`. Mark the real caps in role→capability claim tables (local-dev-credentials.md credentials/role rows; iam.md capability matrix references). LEAVE illustrative `doc.create` (authz-tiers.md:21) and deferred `route.view` (authz-tiers.md:71) UNMARKED. Bump `Last verified:` stamps.
- **GOTCHA**: This touches wiki — keep edits surgical (markers + one note paragraph). Full wiki refresh is Phase 6.
- **VALIDATE**: `wiki-capability-parity` finds ≥1 marker and 0 violations.

### Task 8: Item 5 — reconcile OpenAPI access-policy enum (document, not prune)
- **ACTION**: Annotate the two enums as a separate document-ACL vocabulary; reclassify the ADR Finding-3 row.
- **IMPLEMENT**: Add a `description:` to `AccessPolicyItem.capability` and `AccessPolicyWriteItem.capability` (openapi.yaml:5912,5927) noting: "Document-level ACL action codes for the (currently backend-unimplemented) per-document access-policy feature; backed by the `document_access_policies` CHECK constraint. Deliberately disjoint from the IAM tier-1/tier-2 capability registry (`validCapabilities`) — NOT IAM capabilities." Adding `description` to an existing schema is metadata, NOT a shape change (no path/param/required/enum-membership change). In ADR Finding-3, change the "Stale OpenAPI" row from drift → "intentional separate vocabulary, documented".
- **GOTCHA**: Do NOT touch the `enum:` membership. Redocly must stay valid.
- **VALIDATE**: `npx @redocly/cli lint api/openapi/v1/openapi.yaml` → valid. `go test ./apps/api/... ./internal/modules/iam/...` unaffected.

### Task 9: Item 7 — RLS hardening report (no migration)
- **ACTION**: Document the findings in the ADR amendment.
- **IMPLEMENT**: Add to ADR 0022 amendment / a Phase 5 close-out note: (i) **GUC transaction-locality (R6): PASS** — all `set_config` use `,true`; `SeedTxIdentity` + `asserted_caps` confirmed; regression test added (Task 1). (ii) **Native RLS (R7): N/A by design** — the schema uses NO Postgres `ROW LEVEL SECURITY`; isolation is the `enforce_capability_asserted` `BEFORE INSERT/UPDATE/DELETE` trigger (baseline:3875-4050) reading transaction-local `asserted_caps`. Triggers are NOT subject to the table-owner/superuser RLS-bypass footgun (only `session_replication_role=replica` or explicit `DISABLE TRIGGER` skips them, both superuser-only). Residual guard: the app DB role MUST NOT be superuser and MUST NOT set `session_replication_role` — a deployment/role-grant concern, not a schema migration. `FORCE ROW LEVEL SECURITY` is inapplicable (no RLS policies to force); adding native RLS is an out-of-scope storage redesign (hard-stop). No migration required.
- **VALIDATE**: ADR reads coherently; claims match grep evidence captured in this plan.

### Task 10: Close-out — gates + ADR Phase 5 complete
- **ACTION**: Run all gates, demonstrate each new guard biting, mark Phase 5 complete, open PR.
- **IMPLEMENT**: Run the four gates. For each new guard (Tasks 2,3,4,6) inject drift, capture the red output, revert. Update ADR status line + Phase 5 section with evidence. Update memory `authz-coherence-program.md`.
- **VALIDATE**: see Validation Commands.

---

## Testing Strategy

### Unit Tests
| Test | Input | Expected | Edge? |
|---|---|---|---|
| `registry_rules_test`: inline-cap clean | repo go tree | 0 violations | — |
| `registry_rules_test`: inline-cap drift | fixture w/ `Capability("x")` | 1 violation | ✓ |
| `registry_rules_test`: seed parity clean | real seed + registry | 0 | — |
| `registry_rules_test`: seed parity, seeded∉registry | fixture seed w/ bogus cap | 1 | ✓ |
| `registry_rules_test`: seed parity, registry∉seed | registry cap minus seed | 1 | ✓ |
| `registry_rules_test`: wiki clean | docs w/ markers | 0 | — |
| `registry_rules_test`: wiki drift | `cap:membership.grant` | 1 | ✓ |
| `permissions_authz_scope_test` | routeRules + spec | 0 (all area-grade write ops annotated) | — |
| `context_test`: SeedTxIdentity | sqlmock | combined `set_config(...,true)` ×2 | ✓ |

### Edge Cases Checklist
- [ ] Windows path joins (`filepath.Join`, no hardcoded `/`)
- [ ] `AllCapabilities()` random map order → sorted output
- [ ] Spec `{param}` wildcard vs routeRule prefix/suffix matching
- [ ] Marker regex does NOT match unmarked prose (`doc.create`, `route.view`)
- [ ] Deferred caps set (empty today) honored by seed-parity (b)
- [ ] api-lint still emits the 455 pre-existing violations unchanged (zero new in clean state)

---

## Validation Commands

### Static Analysis
```powershell
go build ./...
go build ./scripts/api-lint
```
EXPECT: builds clean.

### Lint (binary)
```powershell
go run ./scripts/api-lint api/openapi/v1/openapi.yaml .
```
EXPECT: `455 violation(s)` — unchanged vs Phase-4 baseline (new rules clean-state-zero).

### Unit Tests
```powershell
go test ./scripts/api-lint/... ./apps/api/cmd/metaldocs-api/... ./internal/modules/iam/... -count=1
```
EXPECT: all pass, incl. new parity tests.

### OpenAPI
```powershell
npx '@redocly/cli' lint api/openapi/v1/openapi.yaml
```
EXPECT: valid (no shape change).

### Full touched-slice
```powershell
go test ./... -count=1
```
EXPECT: no NEW failures vs base (pre-existing `documents/repository` pagination Scan + `TenantIsolation` env failures are known-unrelated per Phase 2/3 close-out — confirm identical on base).

### Demonstrate guards biting (then revert each)
```powershell
# item 2
# inject iamdomain.Capability("doc.bogus") in permissions.go → expect no-inline-capability line
# item 3
# inject seed row ('viewer','doc.bogus',...) → expect seed-registry-parity line; +1 count
# item 4
# inject `cap:membership.grant` in local-dev-credentials.md → expect wiki-capability-parity line
# item 1
# delete x-authz-skip-area from grantAreaMembership op → expect AuthzScope test red
```
EXPECT: each injection produces the named failure; revert restores green/455.

### Manual Validation
- [ ] Grep confirms zero session-level `set_config` (item 6)
- [ ] Grep confirms zero native `ROW LEVEL SECURITY` (item 7)
- [ ] ADR Phase 5 marked complete with evidence

---

## Acceptance Criteria
- [ ] Items 1-5 each enforced by a check that is proven to fail on injected drift then reverted
- [ ] Item 3 wired into the api-lint binary (not just the unit test)
- [ ] Item 6 verified + regression-locked; Item 7 reported (no migration)
- [ ] api-lint clean-state count unchanged (455); all gates green
- [ ] NO OpenAPI shape change (redocly valid; enum membership untouched)
- [ ] ADR 0022 Phase 5 marked complete; PR opened

## Completion Checklist
- [ ] New rules mirror existing api-lint Violation/AST/yaml patterns
- [ ] Deterministic output (sorted), Windows path-safe
- [ ] Wiki edits surgical (markers + one note); stamps bumped
- [ ] Memory updated
- [ ] Evidence recorded (commands + biting demos)

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Item-1 routeRule↔spec path matching misses an op or false-matches | Med | Med | Reuse routeRule.matches semantics with `{param}` wildcard; assert ≥1 match per area-grade write rule so a non-match is itself a red test |
| `document.create`/`controlled_documents.*` create ops not yet annotated → Task 6 fails | Med | Low | True finding; annotate `x-authz-skip-area` w/ documented reason (Phase 2 precedent) — bounded |
| api-lint importing `iam/domain` introduces a cycle | Low | Med | domain is a leaf pkg (no imports back into scripts); verified `go build ./scripts/api-lint` |
| 455-violation baseline masks new-rule exit signal | Med | Low | New rules clean-state-zero; bite shown via specific line + count delta + dedicated unit tests (0-tolerance) |
| Marker retrofit reads as wiki scope-creep | Low | Low | Minimal markers + one note para; full refresh deferred to Phase 6 |
| `authz-call-present` activation deferred again | High | Low | Explicitly out of scope here; parity bindings meet the "red build" goal; record as residual Phase 5 item (Risks/Notes + ADR) |

## Notes
- **Isolation model (item 7) is trigger-based, not native RLS.** `enforce_capability_asserted()` fires `BEFORE INSERT/UPDATE/DELETE` on sensitive tables (document_families, document_process_areas, document_profiles, iam_user_roles, cd_sequence_counters, controlled_documents, documents, templates_template(_version), user_process_areas, approval_instances/signoffs) and raises `ErrCapabilityNotAsserted` unless the transaction-local `asserted_caps` GUC carries the required cap. This is the ADR's "stricter fail-closed variant" of session-GUC RLS.
- **Decision surfaced**: `authz-call-present` tx-layer activation (ADR Phase 5 bullet 2) is deferred — it is a separate lint-engine rewrite (call-graph / `source: derived` mode) and is not required to make registry/seed/wiki/annotation drift a red build. If the user wants it in this PR, it becomes a materially larger change.
- **Item 5 decision**: document-as-separate-vocabulary chosen over prune because the 3 caps are referenced by a live DB CHECK + FE types (not unused), and narrowing the enum is a forbidden shape change spanning DB+FE.
