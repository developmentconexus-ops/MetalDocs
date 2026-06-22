# Feature F2.1c — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.1c-distribution-contract`  ·  **Closed:** 2026-06-22
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

- OpenAPI distribution tag — 3 denominator-only read endpoints (`420bd0ad`)
  - `GET /documents/{documentId}/distribution` → `DistributionSummaryResponse { total_targets }`
  - `GET /documents/{documentId}/distribution/recipients` → `DistributionRecipientsResponse` (keyset cursor)
  - `GET /documents/{documentId}/distribution/coverage` → `DistributionAreaCoverage[]`
  - `format:uuid` added to path params for parity with siblings (`01c696f2`)
- Go server types generated via oapi-codegen into `internal/modules/distribution/api/api.gen.go` (`4aa921fa`)
- `CapDistributionRead = "distribution.read"` registered: const + `validCapabilities` in `model.go`, `ScopeTenant` in `capability_scope.go`, `deferredCaps` entry in `registry_rules.go`, `capabilityDescriptions` entry in `catalog.go` — all four anchors complete (`c542b66a`, `5f44b06d`)
- FE types regenerated via openapi-typescript into `frontend/apps/web/src/lib/api-types/index.d.ts` (`37ccd46d`)
- ADR-0042 recorded at `wiki/decisions/0042-distribution-module-and-cap.md` (`6a71cdbe`)
- ADR index Last-verified stamp refreshed (`d666499e`)
- Catalog description fix that unblocked initial Task-6 run (`5f44b06d`)

**Producer matches consumer contract** in `spec.md`: denominator-only, no numerator fields, tenant-scoped cap, new greenfield module — confirmed by all gates below.

Full F2.1c commit list (chronological):
```
420bd0ad feat(M2/F2.1c): OpenAPI distribution tag — 3 denominator-only read endpoints
01c696f2 fix(M2/F2.1c): add format:uuid to distribution id path params (sibling parity)
4aa921fa feat(M2/F2.1c): generate distribution Go server types (oapi-codegen)
c542b66a feat(M2/F2.1c): register CapDistributionRead (tenant-scope, deferred)
37ccd46d feat(M2/F2.1c): regenerate FE distribution types (openapi-typescript)
6a71cdbe docs(adr): ADR-0042 distribution module + CapDistributionRead (M2/F2.1c)
d666499e docs(adr): refresh decisions index Last-verified stamp (M2/F2.1c)
5f44b06d fix(M2/F2.1c): add CapDistributionRead catalog description (complete cap registration)
```
All 8 commits present and verified with `git log --oneline -12`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| A — api-lint strict | `./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` / exit 0 | real |
| B — No numerator field in distribution schemas | `grep -nE "read\|acknowledg\|overdue\|pending\|deadline\|timeline\|reminder" api/openapi/v1/openapi.yaml \| grep -iE "distribution\|total_targets\|area_grant\|DistributionRecipient"` | exit 1 (0 matches) | real |
| C — FE types present + denominator-only | `grep -n "DistributionSummaryResponse\|DistributionRecipient\|DistributionAreaCoverage" frontend/apps/web/src/lib/api-types/index.d.ts \| head` | 10 hits; types present; no read/ack fields | real |
| D — Go server types file exists | `ls internal/modules/distribution/api/api.gen.go` | file present | real |
| D — Go build clean | `go build ./...` | exit 0 | real |
| E — Cap const + validCapabilities | `grep -n "CapDistributionRead" internal/modules/iam/domain/model.go` | line 122 (const), line 163 (validCapabilities) — 2 hits | real |
| E — Cap scope = ScopeTenant | `grep -n "CapDistributionRead" internal/modules/iam/domain/capability_scope.go` | line 69: `CapDistributionRead: ScopeTenant` | real |
| E — Cap catalog description | `grep -n "CapDistributionRead" internal/modules/iam/domain/catalog.go` | line 134: `"Visualizar cobertura de distribuição de documentos controlados"` | real |
| E — Cap deferredCaps entry | `grep -n "distribution.read" scripts/api-lint/registry_rules.go` | line 37 (comment + deferredCaps) | real |
| F — ADR-0042 present | `ls wiki/decisions/0042-distribution-module-and-cap.md` | file present | real |
| G — go vet | `go vet ./...` | exit 0 | real |
| G — IAM domain tests (fresh) | `go test ./internal/modules/iam/domain/... -count=1` | `ok metaldocs/internal/modules/iam/domain 1.042s` / exit 0 | real |
| G — TestEveryCapSeededOrDeferred | `go test ./apps/api/cmd/metaldocs-api/... -run TestEveryCapSeededOrDeferred -count=1` | `ok metaldocs/apps/api/cmd/metaldocs-api 6.569s` / exit 0 | real |
| G — api-lint package tests | `go test ./scripts/api-lint/... -count=1` | `ok metaldocs/scripts/api-lint 6.677s` / exit 0 | real |
| G — Full test suite | `go test ./... -count=1` | `fulltest exit: 0` — all packages ok, zero failures | real |
| G — cilint | `go run ./tools/cilint/...` | exit 0 | real |
| G — publish_service.go diff (HS-2) | `git diff --stat origin/main -- internal/modules/documents/approval/application/publish_service.go` | empty (no change) | real |
| G — search module diff (HS-2) | `git diff --stat origin/main -- internal/modules/search` | empty (no change) | real |
| G — db/migrations scope check | `git diff --stat origin/main -- db/migrations` | 2 files (0245, 0246) — F2.1a/F2.1b migrations only (commits f357fb15, 33c82cfc); `git log --oneline origin/main..HEAD -- db/migrations` shows only those two prior-feature commits; no F2.1c commit touched db/migrations | real |

## Acceptance vs spec Validation Gate

Restating the 6 criteria from `spec.md §Validation Gate`:

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `api-lint -strict` parses the new `distribution` paths with **0** violations | yes | Gate A: `0 violation(s)` / exit 0 |
| New response schemas contain **no** numerator field | yes | Gate B: grep exit 1 (0 matches) |
| Generated **FE types** exist and are denominator-only | yes | Gate C: 10 hits in index.d.ts; `DistributionSummaryResponse`, `DistributionRecipient`, `DistributionAreaCoverage` present; no `read`/`acknowledged`/`overdue` fields |
| Generated **Go server types** exist (`go build ./...` clean) | yes | Gate D: `api.gen.go` present; `go build ./...` exit 0 |
| Cap registered + scoped + deferred (COMPLETE — all four anchors) | yes | Gate E: model.go const (line 122) + validCapabilities (line 163); capability_scope.go ScopeTenant (line 69); catalog.go description (line 134); registry_rules.go deferredCaps (line 37). TestCapabilityRegistrySize + TestEveryCapabilityClassified + TestEveryCapSeededOrDeferred all pass (IAM domain suite exit 0; perm suite exit 0) |
| Durable decisions recorded — ADR-0042 present + linked | yes | Gate F: `wiki/decisions/0042-distribution-module-and-cap.md` present |
| Regression (Gate G) — go vet, in-scope suites, full suite, HS-2 scope guards | yes | Gate G: all green; HS-2 diffs empty; db/migrations breach ruled out (prior-feature only) |

## Review disposition

- **Spec-compliance review:** PASS. All 6 spec Validation Gate criteria met with real evidence. Contract is denominator-only (no numerator leak), cap is tenant-scoped and correctly deferred, ADR-0042 recorded. Non-goals confirmed: no migration, no handler, no pre-grant, no numerator field, no `role` field.
- **Code-quality review:** PASS. `go vet`, `cilint`, full `go test ./...` all exit 0. IAM domain tests (including `TestCapabilityCatalogShape`, `TestCapabilityRegistrySize`, `TestEveryCapabilityClassified`) and `TestEveryCapSeededOrDeferred` pass from clean state (no cache). No adjacent code touched.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Handler implementation (SQL, read-source wiring, TDD) | Explicit non-goal — belongs to F2.2 | F2.2 spec + plan are the trigger; operator owns scheduling |
| FE query hooks + screen integration | Explicit non-goal — belongs to F2.3 | F2.3 spec + plan are the trigger |
| Operator role-grant of `CapDistributionRead` | `deferredCaps` is intentional; op wires role grants separately | Operator action at deploy; no agent pre-grant |

---

## Execution note (honest)

**Initial Task-6 BLOCK and resolution.** The first Task-6 run (same session, prior invocation) was
blocked because `CapDistributionRead` was present in `model.go`, `capability_scope.go`, and
`registry_rules.go` but **missing from `catalog.go`'s `capabilityDescriptions` map** — the fourth
required anchor for complete cap registration. This caused `TestCapabilityCatalogShape` to fail.
The defect was fixed in commit `5f44b06d` (`fix(M2/F2.1c): add CapDistributionRead catalog
description (complete cap registration)`). This evidence run confirms all four anchors are present
and all in-scope gates are green.
